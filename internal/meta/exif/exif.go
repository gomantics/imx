package exif

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gomantics/imx/internal/container"
	"github.com/gomantics/imx/internal/meta"
)

// Parser implements meta.Parser for EXIF
type Parser struct{}

// New creates an EXIF parser
func New() *Parser {
	return &Parser{}
}

// Namespace returns the EXIF namespace
func (p *Parser) Namespace() meta.Namespace {
	return "exif"
}

// Parse extracts EXIF data from raw blocks
func (p *Parser) Parse(blocks []container.RawBlock) ([]meta.Directory, error) {
	var dirs []meta.Directory

	for _, block := range blocks {
		if block.Kind != container.MetaKindEXIF {
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
func (p *Parser) parseTIFF(data []byte) ([]meta.Directory, error) {
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
	magic := byteOrder.Uint16(data[2:4])
	if magic != 42 {
		return nil, fmt.Errorf("invalid TIFF magic number: %d", magic)
	}

	// Read offset to first IFD
	ifd0Offset := byteOrder.Uint32(data[4:8])

	var dirs []meta.Directory

	// Parse IFD0
	if ifd0Offset > 0 && int(ifd0Offset) < len(data) {
		ifd0, nextOffset, err := p.parseIFD(data, int(ifd0Offset), byteOrder, "IFD0")
		if err != nil {
			return nil, fmt.Errorf("parse IFD0: %w", err)
		}
		dirs = append(dirs, ifd0)

		// Check for EXIF sub-IFD pointer
		if exifOffset, ok := ifd0.Tags["Exif:ExifOffset"]; ok {
			if offset, ok := exifOffset.Value.(int); ok && offset > 0 && offset < len(data) {
				exifIFD, _, err := p.parseIFD(data, offset, byteOrder, "ExifIFD")
				if err == nil {
					dirs = append(dirs, exifIFD)
				}
			}
		}

		// Check for GPS sub-IFD pointer
		if gpsOffset, ok := ifd0.Tags["Exif:GPSInfo"]; ok {
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
func (p *Parser) parseIFD(data []byte, offset int, byteOrder binary.ByteOrder, name string) (meta.Directory, uint32, error) {
	if offset+2 > len(data) {
		return meta.Directory{}, 0, fmt.Errorf("IFD offset out of bounds")
	}

	// Read number of entries
	entryCount := byteOrder.Uint16(data[offset : offset+2])
	offset += 2

	dir := meta.Directory{
		Namespace: "exif",
		Name:      name,
		Tags:      make(map[meta.TagID]meta.Tag),
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
		nextOffset = byteOrder.Uint32(data[offset : offset+4])
	}

	return dir, nextOffset, nil
}

// parseEntry parses a single IFD entry (tag)
func (p *Parser) parseEntry(data []byte, offset int, byteOrder binary.ByteOrder, ifdName string) meta.Tag {
	tagID := byteOrder.Uint16(data[offset : offset+2])
	tagType := byteOrder.Uint16(data[offset+2 : offset+4])
	count := byteOrder.Uint32(data[offset+4 : offset+8])
	valueOffset := offset + 8 // Last 4 bytes contain value or offset

	tag := meta.Tag{
		Namespace: "exif",
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
	tag.ID = meta.TagID(fmt.Sprintf("Exif:%s", tagName))
	tag.Name = tagName

	// Parse value based on type
	value, typeName := p.parseValue(data, tagType, count, valueOffset, byteOrder)
	tag.Value = value
	tag.Type = typeName

	// Store raw bytes (4 bytes of value/offset)
	tag.Raw = make([]byte, 4)
	copy(tag.Raw, data[valueOffset:valueOffset+4])

	return tag
}

// parseValue parses a tag value based on its type
func (p *Parser) parseValue(data []byte, tagType uint16, count uint32, offset int, byteOrder binary.ByteOrder) (any, string) {
	// Type sizes in bytes
	typeSizes := map[uint16]int{
		1:  1, // BYTE
		2:  1, // ASCII
		3:  2, // SHORT
		4:  4, // LONG
		5:  8, // RATIONAL
		6:  1, // SBYTE
		7:  1, // UNDEFINED
		8:  2, // SSHORT
		9:  4, // SLONG
		10: 8, // SRATIONAL
	}

	typeSize, ok := typeSizes[tagType]
	if !ok {
		return nil, "unknown"
	}

	totalSize := int(count) * typeSize

	// If value fits in 4 bytes, it's stored directly in the offset field
	// Otherwise, the offset field points to the actual data
	var valueData []byte
	if totalSize <= 4 {
		valueData = data[offset : offset+4]
	} else {
		// Read offset to actual value
		valueOffset := int(byteOrder.Uint32(data[offset : offset+4]))
		if valueOffset+totalSize > len(data) {
			return nil, "invalid_offset"
		}
		valueData = data[valueOffset : valueOffset+totalSize]
	}

	switch tagType {
	case 1: // BYTE
		if count == 1 {
			return int(valueData[0]), "byte"
		}
		return valueData[:count], "bytes"

	case 2: // ASCII string
		// Remove trailing null bytes
		str := string(bytes.TrimRight(valueData[:count], "\x00"))
		return str, "string"

	case 3: // SHORT (uint16)
		if count == 1 {
			return int(byteOrder.Uint16(valueData)), "short"
		}
		vals := make([]int, count)
		for i := uint32(0); i < count; i++ {
			vals[i] = int(byteOrder.Uint16(valueData[i*2:]))
		}
		return vals, "shorts"

	case 4: // LONG (uint32)
		if count == 1 {
			return int(byteOrder.Uint32(valueData)), "long"
		}
		vals := make([]int, count)
		for i := uint32(0); i < count; i++ {
			vals[i] = int(byteOrder.Uint32(valueData[i*4:]))
		}
		return vals, "longs"

	case 5: // RATIONAL (num/denom as uint32)
		if count == 1 {
			num := byteOrder.Uint32(valueData[0:4])
			denom := byteOrder.Uint32(valueData[4:8])
			if denom == 0 {
				return 0.0, "rational"
			}
			return float64(num) / float64(denom), "rational"
		}
		vals := make([]float64, count)
		for i := uint32(0); i < count; i++ {
			num := byteOrder.Uint32(valueData[i*8:])
			denom := byteOrder.Uint32(valueData[i*8+4:])
			if denom == 0 {
				vals[i] = 0
			} else {
				vals[i] = float64(num) / float64(denom)
			}
		}
		return vals, "rationals"

	case 7: // UNDEFINED (raw bytes)
		return valueData[:count], "undefined"

	default:
		return valueData[:count], fmt.Sprintf("type_%d", tagType)
	}
}

