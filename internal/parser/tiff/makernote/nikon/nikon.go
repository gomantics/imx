// Package nikon implements Nikon MakerNote parsing.
//
// Nikon MakerNote has multiple variants:
//
// Type 1 (older cameras like D1, D100):
//   - Header: 'Nikon\x00\x01\x00' (8 bytes)
//   - IFD starts at offset 8
//   - Offsets are absolute (relative to EXIF TIFF header)
//   - Inherits byte order from parent TIFF
//
// Type 3 (modern cameras like D3, D700, D800, Z series):
//   - Header: 'Nikon\x00\x02\x00\x00' (10 bytes) + embedded TIFF header
//   - Embedded TIFF starts at offset 10 with own byte order marker
//   - IFD offset is read from embedded TIFF header
//   - Offsets are relative to MakerNote start (offset 10)
//
// Encrypted tags (0x0098 LensData, 0x00A8 FlashInfo) are skipped.
package nikon

import (
	"encoding/binary"
	"fmt"
	"io"

	imxbin "github.com/gomantics/imx/internal/binary"
	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/tiff/makernote"
)

// Handler implements makernote.Handler for Nikon cameras.
type Handler struct{}

// New creates a new Nikon MakerNote handler.
func New() *Handler {
	return &Handler{}
}

// Manufacturer returns "Nikon".
func (h *Handler) Manufacturer() string {
	return "Nikon"
}

// Detect checks if the data is a Nikon MakerNote.
// It tries Type 3 first (more common in modern cameras), then Type 1.
func (h *Handler) Detect(data []byte) (bool, *makernote.Config) {
	// Try Type 3 first (most common)
	if ok, cfg := makernote.DetectNikonType3(data); ok {
		return ok, cfg
	}
	// Try Type 1
	return makernote.DetectNikonType1(data)
}

// Parse extracts metadata from Nikon MakerNote.
func (h *Handler) Parse(r io.ReaderAt, makerNoteOffset, exifBase int64, cfg *makernote.Config) ([]parser.Tag, *parser.ParseError) {
	parseErr := parser.NewParseError()

	// Get byte order - may be nil for Type 1 (inherit from parent)
	order := cfg.ByteOrder
	if order == nil {
		// Try to detect from file header
		header := make([]byte, 2)
		_, err := r.ReadAt(header, 0)
		if err != nil {
			parseErr.Add(fmt.Errorf("failed to read TIFF header for byte order: %w", err))
			return nil, parseErr
		}
		if header[0] == 'I' && header[1] == 'I' {
			order = binary.LittleEndian
		} else {
			order = binary.BigEndian
		}
	}

	reader := imxbin.NewReader(r, order)

	// Calculate IFD start position based on variant
	var ifdOffset int64
	if cfg.Variant == "Type3" {
		// Type 3: IFD offset is stored in embedded TIFF header at offset 14-17
		// (MakerNote offset + 10 byte header + 4 byte for byte order + magic)
		tiffBase := makerNoteOffset + 10
		offsetVal, err := reader.ReadUint32(tiffBase + 4)
		if err != nil {
			parseErr.Add(fmt.Errorf("failed to read Type3 IFD offset: %w", err))
			return nil, parseErr
		}
		ifdOffset = tiffBase + int64(offsetVal)
	} else {
		// Type 1: IFD starts at offset 8
		ifdOffset = makerNoteOffset + cfg.IFDOffset
	}

	// Read number of IFD entries
	numEntries, err := reader.ReadUint16(ifdOffset)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read Nikon IFD entry count: %w", err))
		return nil, parseErr
	}

	// Sanity check entry count
	if numEntries == 0 || numEntries > 200 {
		parseErr.Add(fmt.Errorf("invalid Nikon IFD entry count: %d", numEntries))
		return nil, parseErr
	}

	tags := make([]parser.Tag, 0, numEntries)
	entryOffset := ifdOffset + 2 // Skip entry count

	for i := uint16(0); i < numEntries; i++ {
		tag, err := h.parseEntry(reader, entryOffset, makerNoteOffset, exifBase, cfg)
		if err != nil {
			parseErr.Add(fmt.Errorf("failed to parse Nikon tag at offset %d: %w", entryOffset, err))
		} else if tag != nil {
			tags = append(tags, *tag)
		}
		entryOffset += 12 // Each IFD entry is 12 bytes
	}

	return tags, parseErr.OrNil()
}

// parseEntry parses a single IFD entry.
func (h *Handler) parseEntry(r *imxbin.Reader, offset, makerNoteOffset, exifBase int64, cfg *makernote.Config) (*parser.Tag, error) {
	// Read tag ID
	tagID, err := r.ReadUint16(offset)
	if err != nil {
		return nil, err
	}

	// Skip encrypted tags (LensData, FlashInfo, etc.)
	if isEncryptedTag(tagID) {
		return nil, nil
	}

	// Read type
	typeVal, err := r.ReadUint16(offset + 2)
	if err != nil {
		return nil, err
	}

	// Read count
	count, err := r.ReadUint32(offset + 4)
	if err != nil {
		return nil, err
	}

	// Read value/offset
	valueOffset, err := r.ReadUint32(offset + 8)
	if err != nil {
		return nil, err
	}

	// Calculate data size
	typeSize := getTypeSize(typeVal)
	if typeSize == 0 {
		return nil, nil // Unknown type, skip
	}

	totalSize := int(count) * typeSize

	// Determine where to read the value from
	var value interface{}
	if totalSize <= 4 {
		// Value is inline
		value, err = h.readInlineValue(r, valueOffset, typeVal, count)
	} else {
		// Value is at offset
		dataOffset := h.resolveOffset(int64(valueOffset), makerNoteOffset, exifBase, cfg)
		value, err = h.readValue(r, dataOffset, typeVal, count)
	}

	if err != nil {
		return nil, err
	}

	tagName := h.TagName(tagID)
	if tagName == "" {
		tagName = fmt.Sprintf("0x%04X", tagID)
	}

	return &parser.Tag{
		ID:       parser.TagID(fmt.Sprintf("Nikon:0x%04X", tagID)),
		Name:     tagName,
		Value:    value,
		DataType: getTypeName(typeVal),
	}, nil
}

// isEncryptedTag returns true if the tag contains encrypted data.
func isEncryptedTag(tagID uint16) bool {
	switch tagID {
	case 0x0098: // LensData
		return true
	case 0x00A8: // FlashInfo
		return true
	default:
		return false
	}
}

// resolveOffset calculates the absolute file offset for a tag value.
func (h *Handler) resolveOffset(tagOffset int64, makerNoteOffset, exifBase int64, cfg *makernote.Config) int64 {
	switch cfg.OffsetBase {
	case makernote.OffsetAbsolute:
		return exifBase + tagOffset
	case makernote.OffsetRelativeToMakerNote:
		// Type 3: offsets are relative to the TIFF header within MakerNote (offset 10)
		return makerNoteOffset + 10 + tagOffset
	default:
		return tagOffset
	}
}

// readInlineValue reads a value stored inline in the value/offset field.
func (h *Handler) readInlineValue(r *imxbin.Reader, valueOffset uint32, typeVal uint16, count uint32) (interface{}, error) {
	buf := make([]byte, 4)
	r.PutUint32(buf, valueOffset)

	switch typeVal {
	case 1, 7: // BYTE, UNDEFINED
		if count == 1 {
			return buf[0], nil
		}
		return buf[:count], nil
	case 2: // ASCII
		if count > 0 {
			return string(buf[:count-1]), nil // Exclude null terminator
		}
		return "", nil
	case 3: // SHORT
		if count == 1 {
			return r.Uint16(buf[0:2]), nil
		}
		vals := make([]uint16, count)
		vals[0] = r.Uint16(buf[0:2])
		if count > 1 {
			vals[1] = r.Uint16(buf[2:4])
		}
		return vals, nil
	case 4: // LONG
		return valueOffset, nil
	case 8: // SSHORT
		if count == 1 {
			return int16(r.Uint16(buf[0:2])), nil
		}
		vals := make([]int16, count)
		vals[0] = int16(r.Uint16(buf[0:2]))
		if count > 1 {
			vals[1] = int16(r.Uint16(buf[2:4]))
		}
		return vals, nil
	case 9: // SLONG
		return int32(valueOffset), nil
	default:
		return buf[:4], nil
	}
}

// readValue reads a value from file at the given offset.
func (h *Handler) readValue(r *imxbin.Reader, offset int64, typeVal uint16, count uint32) (interface{}, error) {
	switch typeVal {
	case 1, 7: // BYTE, UNDEFINED
		data, err := r.ReadBytes(offset, int(count))
		if err != nil {
			return nil, err
		}
		if count == 1 {
			return data[0], nil
		}
		return data, nil
	case 2: // ASCII
		data, err := r.ReadBytes(offset, int(count))
		if err != nil {
			return nil, err
		}
		// Trim null terminator
		for len(data) > 0 && data[len(data)-1] == 0 {
			data = data[:len(data)-1]
		}
		return string(data), nil
	case 3: // SHORT
		vals := make([]uint16, count)
		for i := uint32(0); i < count; i++ {
			val, err := r.ReadUint16(offset + int64(i)*2)
			if err != nil {
				return nil, err
			}
			vals[i] = val
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil
	case 4: // LONG
		vals := make([]uint32, count)
		for i := uint32(0); i < count; i++ {
			val, err := r.ReadUint32(offset + int64(i)*4)
			if err != nil {
				return nil, err
			}
			vals[i] = val
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil
	case 5: // RATIONAL
		vals := make([]string, count)
		for i := uint32(0); i < count; i++ {
			num, err := r.ReadUint32(offset + int64(i)*8)
			if err != nil {
				return nil, err
			}
			denom, err := r.ReadUint32(offset + int64(i)*8 + 4)
			if err != nil {
				return nil, err
			}
			vals[i] = fmt.Sprintf("%d/%d", num, denom)
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil
	case 8: // SSHORT
		vals := make([]int16, count)
		for i := uint32(0); i < count; i++ {
			val, err := r.ReadInt16(offset + int64(i)*2)
			if err != nil {
				return nil, err
			}
			vals[i] = val
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil
	case 9: // SLONG
		vals := make([]int32, count)
		for i := uint32(0); i < count; i++ {
			val, err := r.ReadInt32(offset + int64(i)*4)
			if err != nil {
				return nil, err
			}
			vals[i] = val
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil
	case 10: // SRATIONAL
		vals := make([]string, count)
		for i := uint32(0); i < count; i++ {
			num, err := r.ReadInt32(offset + int64(i)*8)
			if err != nil {
				return nil, err
			}
			denom, err := r.ReadInt32(offset + int64(i)*8 + 4)
			if err != nil {
				return nil, err
			}
			vals[i] = fmt.Sprintf("%d/%d", num, denom)
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil
	default:
		data, err := r.ReadBytes(offset, int(count)*getTypeSize(typeVal))
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

// TagName returns the human-readable name for a Nikon tag.
func (h *Handler) TagName(tagID uint16) string {
	if name, ok := nikonTagNames[tagID]; ok {
		return name
	}
	return ""
}

// getTypeSize returns the size in bytes for a TIFF type.
func getTypeSize(typeVal uint16) int {
	switch typeVal {
	case 1, 2, 6, 7: // BYTE, ASCII, SBYTE, UNDEFINED
		return 1
	case 3, 8: // SHORT, SSHORT
		return 2
	case 4, 9, 11: // LONG, SLONG, FLOAT
		return 4
	case 5, 10, 12: // RATIONAL, SRATIONAL, DOUBLE
		return 8
	default:
		return 0
	}
}

// getTypeName returns the string name for a TIFF type.
func getTypeName(typeVal uint16) string {
	switch typeVal {
	case 1:
		return "BYTE"
	case 2:
		return "ASCII"
	case 3:
		return "SHORT"
	case 4:
		return "LONG"
	case 5:
		return "RATIONAL"
	case 6:
		return "SBYTE"
	case 7:
		return "UNDEFINED"
	case 8:
		return "SSHORT"
	case 9:
		return "SLONG"
	case 10:
		return "SRATIONAL"
	case 11:
		return "FLOAT"
	case 12:
		return "DOUBLE"
	default:
		return "UNKNOWN"
	}
}
