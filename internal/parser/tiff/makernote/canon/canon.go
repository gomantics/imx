// Package canon implements Canon MakerNote parsing.
//
// Canon MakerNote format:
//   - No header - IFD starts immediately at offset 0
//   - Byte order inherited from parent EXIF
//   - Offsets are absolute (relative to EXIF TIFF header)
//   - CR2 files may have non-zero next-IFD pointer (ignored)
//
// Key tags include CameraSettings arrays, SerialNumber, LensModel, and ModelID.
package canon

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/tiff/makernote"
)

// Handler implements makernote.Handler for Canon cameras.
type Handler struct{}

// New creates a new Canon MakerNote handler.
func New() *Handler {
	return &Handler{}
}

// Manufacturer returns the manufacturer name.
func (h *Handler) Manufacturer() string {
	return "Canon"
}

// Detect checks if the data is a Canon MakerNote.
// Canon has no header - detection is based on valid IFD structure.
// This should be called as a fallback after other manufacturers are ruled out.
func (h *Handler) Detect(data []byte) (bool, *makernote.Config) {
	ok, cfg := makernote.DetectCanon(data)
	if !ok {
		return false, nil
	}
	return true, cfg
}

// Parse extracts metadata from a Canon MakerNote.
func (h *Handler) Parse(r io.ReaderAt, makerNoteOffset, exifBase int64, cfg *makernote.Config) ([]parser.Tag, *parser.ParseError) {
	parseErr := parser.NewParseError()
	tags := make([]parser.Tag, 0)

	// Determine byte order - Canon inherits from parent
	order := cfg.ByteOrder
	if order == nil {
		// Default to little-endian if not specified (most common for Canon)
		order = binary.LittleEndian
	}

	// Read IFD at makerNoteOffset + cfg.IFDOffset (0 for Canon)
	ifdOffset := makerNoteOffset + cfg.IFDOffset

	// Read entry count
	entryCountBuf := make([]byte, 2)
	if _, err := r.ReadAt(entryCountBuf, ifdOffset); err != nil {
		parseErr.Add(fmt.Errorf("canon: failed to read IFD entry count: %w", err))
		return nil, parseErr
	}
	entryCount := order.Uint16(entryCountBuf)

	// Sanity check
	if entryCount == 0 || entryCount > 100 {
		parseErr.Add(fmt.Errorf("canon: invalid entry count: %d", entryCount))
		return nil, parseErr
	}

	// Read each IFD entry (12 bytes each)
	entryOffset := ifdOffset + 2
	entryBuf := make([]byte, 12)

	for i := uint16(0); i < entryCount; i++ {
		if _, err := r.ReadAt(entryBuf, entryOffset); err != nil {
			parseErr.Add(fmt.Errorf("canon: failed to read IFD entry %d: %w", i, err))
			entryOffset += 12
			continue
		}

		tagID := order.Uint16(entryBuf[0:2])
		tagType := order.Uint16(entryBuf[2:4])
		count := order.Uint32(entryBuf[4:8])
		valueOffset := order.Uint32(entryBuf[8:12])

		tag, err := h.parseTag(r, order, tagID, tagType, count, valueOffset, exifBase)
		if err != nil {
			parseErr.Add(fmt.Errorf("canon: tag 0x%04X: %w", tagID, err))
			entryOffset += 12
			continue
		}

		if tag != nil {
			tags = append(tags, *tag)
		}

		entryOffset += 12
	}

	return tags, parseErr.OrNil()
}

// parseTag parses a single Canon tag.
func (h *Handler) parseTag(r io.ReaderAt, order binary.ByteOrder, tagID, tagType uint16, count, valueOffset uint32, exifBase int64) (*parser.Tag, error) {
	tagName := GetTagName(tagID)
	if tagName == "" {
		tagName = fmt.Sprintf("0x%04X", tagID)
	}

	// Calculate data size
	typeSize := getTypeSize(tagType)
	if typeSize == 0 {
		return nil, fmt.Errorf("unknown type: %d", tagType)
	}

	totalSize := int(count) * typeSize
	if totalSize > 50*1024*1024 { // 50MB limit
		return nil, fmt.Errorf("data size exceeds limit: %d", totalSize)
	}

	// Read value
	var value any
	var err error

	if totalSize <= 4 {
		// Inline value
		value, err = h.readInlineValue(order, tagType, count, valueOffset)
	} else {
		// Value at offset (absolute, relative to EXIF TIFF header)
		dataOffset := exifBase + int64(valueOffset)
		value, err = h.readOffsetValue(r, order, tagType, count, dataOffset)
	}

	if err != nil {
		return nil, err
	}

	return &parser.Tag{
		ID:       parser.TagID(fmt.Sprintf("Canon:0x%04X", tagID)),
		Name:     tagName,
		Value:    value,
		DataType: getTypeName(tagType),
	}, nil
}

// readInlineValue reads a value stored inline in the IFD entry.
func (h *Handler) readInlineValue(order binary.ByteOrder, tagType uint16, count, valueOffset uint32) (any, error) {
	// Convert valueOffset to bytes for extraction
	buf := make([]byte, 4)
	order.PutUint32(buf, valueOffset)

	switch tagType {
	case 1, 7: // BYTE, UNDEFINED
		if count == 1 {
			return buf[0], nil
		}
		return buf[:count], nil

	case 2: // ASCII
		s := string(buf[:count])
		// Trim null terminator
		for i := 0; i < len(s); i++ {
			if s[i] == 0 {
				return s[:i], nil
			}
		}
		return s, nil

	case 3: // SHORT
		if count == 1 {
			return order.Uint16(buf[0:2]), nil
		}
		vals := make([]uint16, count)
		for i := uint32(0); i < count && i < 2; i++ {
			vals[i] = order.Uint16(buf[i*2 : i*2+2])
		}
		return vals, nil

	case 4: // LONG
		return valueOffset, nil

	case 6: // SBYTE
		if count == 1 {
			return int8(buf[0]), nil
		}
		vals := make([]int8, count)
		for i := uint32(0); i < count && i < 4; i++ {
			vals[i] = int8(buf[i])
		}
		return vals, nil

	case 8: // SSHORT
		if count == 1 {
			return int16(order.Uint16(buf[0:2])), nil
		}
		vals := make([]int16, count)
		for i := uint32(0); i < count && i < 2; i++ {
			vals[i] = int16(order.Uint16(buf[i*2 : i*2+2]))
		}
		return vals, nil

	case 9: // SLONG
		return int32(valueOffset), nil

	default:
		return buf[:4], nil
	}
}

// readOffsetValue reads a value stored at an offset in the file.
func (h *Handler) readOffsetValue(r io.ReaderAt, order binary.ByteOrder, tagType uint16, count uint32, dataOffset int64) (any, error) {
	typeSize := getTypeSize(tagType)
	totalSize := int(count) * typeSize

	data := make([]byte, totalSize)
	if _, err := r.ReadAt(data, dataOffset); err != nil {
		return nil, fmt.Errorf("failed to read data at offset %d: %w", dataOffset, err)
	}

	switch tagType {
	case 1, 7: // BYTE, UNDEFINED
		if count == 1 {
			return data[0], nil
		}
		return data, nil

	case 2: // ASCII
		// Trim null terminator
		for i := 0; i < len(data); i++ {
			if data[i] == 0 {
				return string(data[:i]), nil
			}
		}
		return string(data), nil

	case 3: // SHORT
		vals := make([]uint16, count)
		for i := uint32(0); i < count; i++ {
			vals[i] = order.Uint16(data[i*2 : i*2+2])
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil

	case 4: // LONG
		vals := make([]uint32, count)
		for i := uint32(0); i < count; i++ {
			vals[i] = order.Uint32(data[i*4 : i*4+4])
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil

	case 5: // RATIONAL
		vals := make([]string, count)
		for i := uint32(0); i < count; i++ {
			num := order.Uint32(data[i*8 : i*8+4])
			denom := order.Uint32(data[i*8+4 : i*8+8])
			vals[i] = fmt.Sprintf("%d/%d", num, denom)
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil

	case 6: // SBYTE
		vals := make([]int8, count)
		for i := uint32(0); i < count; i++ {
			vals[i] = int8(data[i])
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil

	case 8: // SSHORT
		vals := make([]int16, count)
		for i := uint32(0); i < count; i++ {
			vals[i] = int16(order.Uint16(data[i*2 : i*2+2]))
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil

	case 9: // SLONG
		vals := make([]int32, count)
		for i := uint32(0); i < count; i++ {
			vals[i] = int32(order.Uint32(data[i*4 : i*4+4]))
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil

	case 10: // SRATIONAL
		vals := make([]string, count)
		for i := uint32(0); i < count; i++ {
			num := int32(order.Uint32(data[i*8 : i*8+4]))
			denom := int32(order.Uint32(data[i*8+4 : i*8+8]))
			vals[i] = fmt.Sprintf("%d/%d", num, denom)
		}
		if count == 1 {
			return vals[0], nil
		}
		return vals, nil

	default:
		return data, nil
	}
}

// TagName returns the human-readable name for a Canon tag ID.
func (h *Handler) TagName(tagID uint16) string {
	return GetTagName(tagID)
}

// getTypeSize returns the size in bytes for a TIFF tag type.
func getTypeSize(tagType uint16) int {
	switch tagType {
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

// getTypeName returns the name for a TIFF tag type.
func getTypeName(tagType uint16) string {
	switch tagType {
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
