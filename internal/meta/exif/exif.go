package exif

import (
	"encoding/binary"
	"fmt"

	"github.com/gomantics/imx/internal/common"
)

// Parser implements meta.Parser for EXIF
type Parser struct{}

// New creates an EXIF parser
func New() *Parser {
	return &Parser{}
}

// Spec returns the EXIF metadata spec
func (p *Parser) Spec() common.Spec {
	return common.SpecEXIF
}

// Parse extracts EXIF data from raw blocks
func (p *Parser) Parse(blocks []common.RawBlock) ([]common.Directory, error) {
	var dirs []common.Directory

	for _, block := range blocks {
		if block.Spec != common.SpecEXIF {
			continue
		}

		// Parse TIFF structure
		blockDirs, err := p.parseTIFF(block.Payload)
		if err != nil {
			return nil, fmt.Errorf("parse TIFF: %w", err)
		}

		dirs = append(dirs, blockDirs...)
	}

	return dirs, nil
}

// parseTIFF parses a TIFF-formatted EXIF block
func (p *Parser) parseTIFF(data []byte) ([]common.Directory, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("TIFF header too short")
	}

	// Read byte order (first 2 bytes)
	var byteOrder binary.ByteOrder
	if data[0] == 'I' && data[1] == 'I' {
		byteOrder = binary.LittleEndian // Intel
	} else if data[0] == 'M' && data[1] == 'M' {
		byteOrder = binary.BigEndian // Motorola
	} else {
		return nil, fmt.Errorf("invalid TIFF byte order: %02X %02X", data[0], data[1])
	}

	// Verify TIFF magic number (should be 42)
	// Safe: we already checked len(data) >= 8 above
	magic, _ := common.ReadUint16(data, 2, byteOrder)
	if magic != 42 {
		return nil, fmt.Errorf("invalid TIFF magic number: %d", magic)
	}

	// Read offset to first IFD
	// Safe: we already checked len(data) >= 8 above
	ifd0Offset, _ := common.ReadUint32(data, 4, byteOrder)

	var dirs []common.Directory

	// Parse IFD0
	if ifd0Offset > 0 && int(ifd0Offset) < len(data) {
		ifd0, nextOffset, err := p.parseIFD(data, int(ifd0Offset), byteOrder, "IFD0")
		if err != nil {
			return nil, fmt.Errorf("parse IFD0: %w", err)
		}
		dirs = append(dirs, ifd0)

		// Check for EXIF sub-IFD pointer
		if exifOffset, ok := ifd0.Tags["EXIF:IFD0:ExifOffset"]; ok {
			if offset, ok := exifOffset.Value.(int); ok && offset > 0 && offset < len(data) {
				exifIFD, _, err := p.parseIFD(data, offset, byteOrder, "ExifIFD")
				if err == nil {
					dirs = append(dirs, exifIFD)
				}
			}
		}

		// Check for GPS sub-IFD pointer
		if gpsOffset, ok := ifd0.Tags["EXIF:IFD0:GPSInfo"]; ok {
			if offset, ok := gpsOffset.Value.(int); ok && offset > 0 && offset < len(data) {
				gpsIFD, _, err := p.parseIFD(data, offset, byteOrder, "GPS")
				if err == nil {
					dirs = append(dirs, gpsIFD)
				}
			}
		}

		// Parse IFD1 (thumbnail) if present
		if nextOffset > 0 && int(nextOffset) < len(data) {
			ifd1, _, err := p.parseIFD(data, int(nextOffset), byteOrder, "IFD1")
			if err == nil {
				dirs = append(dirs, ifd1)
			}
		}
	}

	return dirs, nil
}

// parseIFD parses a single IFD (Image File Directory)
func (p *Parser) parseIFD(data []byte, offset int, byteOrder binary.ByteOrder, name string) (common.Directory, uint32, error) {
	if offset+2 > len(data) {
		return common.Directory{}, 0, fmt.Errorf("IFD offset out of bounds")
	}

	// Read number of entries
	// Safe: we already checked offset+2 <= len(data) above
	entryCount, _ := common.ReadUint16(data, offset, byteOrder)
	offset += 2

	dir := common.Directory{
		Spec: common.SpecEXIF,
		Name: name,
		Tags: make(map[common.TagID]common.Tag, entryCount),
	}

	// Parse each entry (12 bytes each)
	for i := 0; i < int(entryCount); i++ {
		if offset+12 > len(data) {
			break
		}

		tag := p.parseEntry(data, offset, byteOrder, name)
		if tag.ID != "" {
			dir.Tags[tag.ID] = tag
		}

		offset += 12
	}

	// Read offset to next IFD
	var nextOffset uint32
	if offset+4 <= len(data) {
		val, err := common.ReadUint32(data, offset, byteOrder)
		if err == nil {
			nextOffset = val
		}
	}

	return dir, nextOffset, nil
}

// parseEntry parses a single IFD entry (tag)
func (p *Parser) parseEntry(data []byte, offset int, byteOrder binary.ByteOrder, ifdName string) common.Tag {
	tagID, _ := common.ReadUint16(data, offset, byteOrder)
	tagType, _ := common.ReadUint16(data, offset+2, byteOrder)
	count, _ := common.ReadUint32(data, offset+4, byteOrder)
	valueOffset := offset + 8 // Last 4 bytes contain value or offset

	tag := common.Tag{
		Spec: common.SpecEXIF,
	}

	// Get tag name and ID based on IFD
	var tagName string
	var ok bool

	if ifdName == "GPS" {
		// GPS tags have their own namespace because they conflict with main EXIF tags
		tagName, ok = gpsTags[tagID]
	} else {
		// All other tags (IFD0, ExifIFD, InteropIFD, IFD1) use the main tag map
		tagName, ok = knownTags[tagID]
	}

	if !ok {
		tagName = fmt.Sprintf("Tag%04X", tagID)
	}

	// Prefix all tags with IFD name for clarity and to avoid ambiguity
	// e.g., IFD0:XResolution vs IFD1:XResolution
	tagName = ifdName + ":" + tagName

	tag.ID = common.TagID("EXIF:" + tagName)
	tag.Name = tagName

	// Parse value based on type
	value, typeName := p.parseValue(data, tagType, count, valueOffset, byteOrder)
	tag.Value = value
	tag.DataType = typeName

	// Store raw bytes (4 bytes of value/offset)
	tag.Raw = make([]byte, 4)
	copy(tag.Raw, data[valueOffset:valueOffset+4])

	return tag
}

// parseValue parses a tag value based on its type
func (p *Parser) parseValue(data []byte, tagType uint16, count uint32, offset int, byteOrder binary.ByteOrder) (any, string) {
	// Get TIFF type size
	typeSize, ok := common.TIFFTypeSizes[tagType]
	if !ok {
		return nil, "unknown"
	}

	totalSize := int(count) * typeSize

	// If value fits in 4 bytes, it's stored directly in the offset field
	// Otherwise, the offset field points to the actual data
	var valueData []byte
	if totalSize <= 4 {
		// Safe: parseEntry is only called when offset+12 <= len(data)
		valueData, _ = common.SafeSlice(data, offset, 4)
	} else {
		// Read offset to actual value
		// Safe: parseEntry is only called when offset+12 <= len(data)
		valueOffsetVal, _ := common.ReadUint32(data, offset, byteOrder)

		// Validate offset to actual value data
		slice, err := common.SafeSlice(data, int(valueOffsetVal), totalSize)
		if err != nil {
			return nil, "invalid_offset"
		}
		valueData = slice
	}

	// Use TIFF type parser (guaranteed to exist since typeSize was found)
	parser := common.TIFFTypeParsers[tagType]
	return parser.Parse(valueData, count, byteOrder)
}
