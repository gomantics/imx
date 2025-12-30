package heic

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
)

// extractMetadata extracts all metadata using the HEIF index.
func (p *Parser) extractMetadata(r io.ReaderAt, index *HeifIndex) []parser.Directory {
	var dirs []parser.Directory

	// Find primary item
	primaryItem, exists := index.Items[index.PrimaryItemID]
	if !exists {
		return dirs
	}

	// Find metadata items that describe the primary item
	for _, item := range index.Items {
		if !describesPrimaryItem(item, index.PrimaryItemID) {
			continue
		}

		switch item.ItemType {
		case itemTypeExif:
			exifDirs := p.extractExif(r, item)
			dirs = append(dirs, exifDirs...)
		case itemTypeMime:
			xmpDirs := p.extractXMP(r, item)
			dirs = append(dirs, xmpDirs...)
		}
	}

	// Extract ICC from primary item's colr property
	iccDirs := p.extractICC(r, primaryItem)
	dirs = append(dirs, iccDirs...)

	return dirs
}

// describesPrimaryItem checks if an item references the primary item.
func describesPrimaryItem(item *HeifItem, primaryID uint32) bool {
	for _, refID := range item.References {
		if refID == primaryID {
			return true
		}
	}
	return false
}

// extractExif extracts EXIF metadata from an Exif item.
func (p *Parser) extractExif(r io.ReaderAt, item *HeifItem) []parser.Directory {
	data, err := readItemData(r, item)
	if err != nil || len(data) < 8 {
		return nil
	}

	// HEIF EXIF format:
	// - First 4 bytes: big-endian offset to TIFF header
	// - Followed by TIFF data at that offset
	tiffOffset := binary.BigEndian.Uint32(data[0:4])

	if tiffOffset < 4 || int(tiffOffset) >= len(data) {
		return nil
	}

	tiffData := data[tiffOffset:]

	// Scan for TIFF header (MM or II) within first bytes
	tiffStart := findTIFFHeader(tiffData)
	if tiffStart < 0 || tiffStart >= len(tiffData) {
		return nil
	}

	tiffData = tiffData[tiffStart:]
	if len(tiffData) < 8 {
		return nil
	}

	section := io.NewSectionReader(bytes.NewReader(tiffData), 0, int64(len(tiffData)))
	dirs, _ := p.tiff.Parse(section)

	return dirs
}

// findTIFFHeader scans for TIFF header signature.
func findTIFFHeader(data []byte) int {
	for i := 0; i+2 <= len(data) && i < maxTIFFScanOffset; i++ {
		if (data[i] == 'M' && data[i+1] == 'M') ||
			(data[i] == 'I' && data[i+1] == 'I') {
			return i
		}
	}
	return -1
}

// extractXMP extracts XMP metadata from a mime item.
func (p *Parser) extractXMP(r io.ReaderAt, item *HeifItem) []parser.Directory {
	data, err := readItemData(r, item)
	if err != nil || len(data) == 0 {
		return nil
	}

	if !isXMPData(data) {
		return nil
	}

	cleanData := removeNullBytes(data)
	reader := bytes.NewReader(cleanData)
	dirs, _ := p.xmp.Parse(reader)

	return dirs
}

// isXMPData checks if data contains XMP signatures.
func isXMPData(data []byte) bool {
	return bytes.Contains(data, xmpPacketSignature) ||
		bytes.Contains(data, xmpXmMetaSignature)
}

// extractICC extracts ICC profile from colr property.
func (p *Parser) extractICC(r io.ReaderAt, item *HeifItem) []parser.Directory {
	if item.ICCProperty == nil {
		return nil
	}

	colrBox := item.ICCProperty

	header := make([]byte, 4)
	if _, err := r.ReadAt(header[:4], colrBox.Payload); err != nil {
		return nil
	}

	colorType := string(header[:4])
	if colorType != colorTypeRICC && colorType != colorTypeProf {
		return nil
	}

	iccOffset := colrBox.Payload + 4
	iccSize := int64(colrBox.Size) - (colrBox.Payload - colrBox.Offset) - 4

	if iccSize <= 0 {
		return nil
	}

	iccData := make([]byte, iccSize)
	if _, err := r.ReadAt(iccData, iccOffset); err != nil {
		return nil
	}

	reader := bytes.NewReader(iccData)
	dirs, _ := p.icc.Parse(reader)

	return dirs
}

// readItemData reads all data for an item, assembling from extents.
func readItemData(r io.ReaderAt, item *HeifItem) ([]byte, error) {
	loc := item.Location

	var totalSize uint64
	for _, ext := range loc.Extents {
		totalSize += ext.Length
	}

	if totalSize == 0 {
		return nil, nil
	}

	data := make([]byte, totalSize)
	pos := uint64(0)

	for _, ext := range loc.Extents {
		fileOffset := int64(loc.BaseOffset + ext.Offset)

		n, err := r.ReadAt(data[pos:pos+ext.Length], fileOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to read extent at offset %d: %w", fileOffset, err)
		}
		if uint64(n) < ext.Length {
			return nil, fmt.Errorf("incomplete extent read: expected %d, got %d", ext.Length, n)
		}

		pos += ext.Length
	}

	return data, nil
}

// removeNullBytes removes null bytes in-place.
func removeNullBytes(data []byte) []byte {
	writeIdx := 0
	for i := 0; i < len(data); i++ {
		if data[i] != 0 {
			data[writeIdx] = data[i]
			writeIdx++
		}
	}
	return data[:writeIdx]
}
