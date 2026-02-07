// Package canon provides parsing for Canon MakerNote data.
//
// Canon MakerNote format:
//   - No header - IFD starts immediately at offset 0
//   - Byte order: inherited from parent EXIF
//   - Offsets: absolute (relative to EXIF TIFF header)
//   - Quirk: CR2 files may have non-zero next-IFD pointer (should be ignored)
package canon

import (
	"encoding/binary"
	"fmt"
	"io"

	imxbin "github.com/gomantics/imx/internal/binary"
	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/tiff/makernote"
)

// IFD entry constants
const (
	ifdEntrySize      = 12
	ifdEntryCountSize = 2
	inlineThreshold   = 4
)

// Handler implements makernote.Handler for Canon cameras.
type Handler struct {
	// parentByteOrder is set during detection for use in parsing
	parentByteOrder binary.ByteOrder
}

// New creates a new Canon MakerNote handler.
func New() *Handler {
	return &Handler{}
}

// Manufacturer returns the camera manufacturer.
func (h *Handler) Manufacturer() string {
	return "Canon"
}

// Detect checks if the data is a Canon MakerNote.
// Canon has no header, so detection is based on valid IFD entry count.
// This should be called as a fallback after other manufacturers are ruled out.
func (h *Handler) Detect(data []byte) (bool, *makernote.Config) {
	if len(data) < 14 { // Minimum: 2 bytes count + 1 entry (12 bytes)
		return false, nil
	}

	// Try little-endian first (most common)
	entryCount := binary.LittleEndian.Uint16(data[0:2])
	var byteOrder binary.ByteOrder = binary.LittleEndian

	// If little-endian count is invalid, try big-endian
	if entryCount < 1 || entryCount > 100 {
		entryCount = binary.BigEndian.Uint16(data[0:2])
		byteOrder = binary.BigEndian
	}

	// Validate entry count
	if entryCount < 1 || entryCount > 100 {
		return false, nil
	}

	// Validate that we have enough data for the claimed entries
	requiredSize := ifdEntryCountSize + int(entryCount)*ifdEntrySize
	if len(data) < requiredSize {
		return false, nil
	}

	// Additional validation: check first entry has valid tag type (1-12)
	tagType := byteOrder.Uint16(data[4:6]) // Type is at offset 2 within first entry
	if tagType < 1 || tagType > 12 {
		return false, nil
	}

	h.parentByteOrder = byteOrder

	return true, &makernote.Config{
		IFDOffset:  0,
		OffsetBase: makernote.OffsetAbsolute,
		ByteOrder:  byteOrder,
		HasNextIFD: false, // Ignore next-IFD pointer in Canon
		Variant:    "Standard",
	}
}

// Parse extracts Canon MakerNote tags.
func (h *Handler) Parse(r io.ReaderAt, makerNoteOffset, exifBase int64, cfg *makernote.Config) ([]parser.Tag, *parser.ParseError) {
	parseErr := parser.NewParseError()

	// Determine byte order
	byteOrder := cfg.ByteOrder
	if byteOrder == nil {
		byteOrder = h.parentByteOrder
	}
	if byteOrder == nil {
		byteOrder = binary.LittleEndian // Default fallback
	}

	reader := imxbin.NewReader(r, byteOrder)

	// Read entry count at MakerNote start
	ifdOffset := makerNoteOffset + cfg.IFDOffset
	entryCount, err := reader.ReadUint16(ifdOffset)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read Canon MakerNote entry count: %w", err))
		return nil, parseErr
	}

	if entryCount == 0 || entryCount > 100 {
		parseErr.Add(fmt.Errorf("invalid Canon MakerNote entry count: %d", entryCount))
		return nil, parseErr
	}

	tags := make([]parser.Tag, 0, entryCount)

	// Parse each IFD entry
	entryOffset := ifdOffset + ifdEntryCountSize
	for i := uint16(0); i < entryCount; i++ {
		tag, err := h.parseEntry(reader, entryOffset, exifBase)
		if err != nil {
			// Continue parsing other entries
			entryOffset += ifdEntrySize
			continue
		}
		if tag != nil {
			tags = append(tags, *tag)
		}
		entryOffset += ifdEntrySize
	}

	return tags, parseErr.OrNil()
}

// parseEntry parses a single IFD entry.
// For Canon, offsets are absolute (relative to EXIF TIFF header).
func (h *Handler) parseEntry(r *imxbin.Reader, entryOffset, exifBase int64) (*parser.Tag, error) {
	// Read entry fields
	tagID, err := r.ReadUint16(entryOffset)
	if err != nil {
		return nil, err
	}

	tagType, err := r.ReadUint16(entryOffset + 2)
	if err != nil {
		return nil, err
	}

	count, err := r.ReadUint32(entryOffset + 4)
	if err != nil {
		return nil, err
	}

	valueOffset, err := r.ReadUint32(entryOffset + 8)
	if err != nil {
		return nil, err
	}

	// Calculate data size
	typeSize := getTypeSize(tagType)
	if typeSize == 0 {
		return nil, fmt.Errorf("unknown type: %d", tagType)
	}

	totalSize := int64(count) * int64(typeSize)

	// Determine where to read data from
	var dataOffset int64
	if totalSize <= inlineThreshold {
		dataOffset = entryOffset + 8 // Inline in value field
	} else {
		// Canon uses absolute offsets (relative to EXIF TIFF header)
		dataOffset = exifBase + int64(valueOffset)
	}

	// Read and parse value
	value, err := h.readValue(r, tagType, count, dataOffset, valueOffset)
	if err != nil {
		return nil, err
	}

	return &parser.Tag{
		ID:       parser.TagID(fmt.Sprintf("Canon:0x%04X", tagID)),
		Name:     h.TagName(tagID),
		Value:    value,
		DataType: getTypeName(tagType),
	}, nil
}

// readValue reads a tag value based on its type.
func (h *Handler) readValue(r *imxbin.Reader, tagType uint16, count uint32, dataOffset int64, valueOffset uint32) (any, error) {
	switch tagType {
	case 1, 7: // BYTE, UNDEFINED
		return h.readBytes(r, count, dataOffset, valueOffset)
	case 2: // ASCII
		return h.readASCII(r, count, dataOffset, valueOffset)
	case 3: // SHORT
		return h.readShorts(r, count, dataOffset, valueOffset)
	case 4: // LONG
		return h.readLongs(r, count, dataOffset, valueOffset)
	case 5: // RATIONAL
		return h.readRationals(r, count, dataOffset)
	case 6: // SBYTE
		return h.readSBytes(r, count, dataOffset, valueOffset)
	case 8: // SSHORT
		return h.readSShorts(r, count, dataOffset, valueOffset)
	case 9: // SLONG
		return h.readSLongs(r, count, dataOffset, valueOffset)
	case 10: // SRATIONAL
		return h.readSRationals(r, count, dataOffset)
	default:
		return nil, fmt.Errorf("unsupported type: %d", tagType)
	}
}

func (h *Handler) readBytes(r *imxbin.Reader, count uint32, dataOffset int64, valueOffset uint32) (any, error) {
	if count <= 4 {
		// Inline data
		buf := make([]byte, 4)
		r.PutUint32(buf, valueOffset)
		if count == 1 {
			return buf[0], nil
		}
		return buf[:count], nil
	}
	data, err := r.ReadBytes(dataOffset, int(count))
	if err != nil {
		return nil, err
	}
	if count == 1 {
		return data[0], nil
	}
	return data, nil
}

func (h *Handler) readASCII(r *imxbin.Reader, count uint32, dataOffset int64, valueOffset uint32) (any, error) {
	var data []byte
	if count <= 4 {
		buf := make([]byte, 4)
		r.PutUint32(buf, valueOffset)
		data = buf[:count]
	} else {
		var err error
		data, err = r.ReadBytes(dataOffset, int(count))
		if err != nil {
			return nil, err
		}
	}
	// Trim null terminator
	for len(data) > 0 && data[len(data)-1] == 0 {
		data = data[:len(data)-1]
	}
	return string(data), nil
}

func (h *Handler) readShorts(r *imxbin.Reader, count uint32, dataOffset int64, valueOffset uint32) (any, error) {
	if count == 0 {
		return []uint16{}, nil
	}

	values := make([]uint16, count)
	if count <= 2 {
		// Inline data
		buf := make([]byte, 4)
		r.PutUint32(buf, valueOffset)
		values[0] = r.Uint16(buf[0:2])
		if count > 1 {
			values[1] = r.Uint16(buf[2:4])
		}
	} else {
		for i := uint32(0); i < count; i++ {
			val, err := r.ReadUint16(dataOffset + int64(i*2))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

func (h *Handler) readLongs(r *imxbin.Reader, count uint32, dataOffset int64, valueOffset uint32) (any, error) {
	if count == 0 {
		return []uint32{}, nil
	}

	values := make([]uint32, count)
	if count == 1 {
		values[0] = valueOffset
	} else {
		for i := uint32(0); i < count; i++ {
			val, err := r.ReadUint32(dataOffset + int64(i*4))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

func (h *Handler) readRationals(r *imxbin.Reader, count uint32, dataOffset int64) (any, error) {
	if count == 0 {
		return []string{}, nil
	}

	values := make([]string, count)
	for i := uint32(0); i < count; i++ {
		offset := dataOffset + int64(i*8)
		num, err := r.ReadUint32(offset)
		if err != nil {
			return nil, err
		}
		denom, err := r.ReadUint32(offset + 4)
		if err != nil {
			return nil, err
		}
		values[i] = fmt.Sprintf("%d/%d", num, denom)
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

func (h *Handler) readSBytes(r *imxbin.Reader, count uint32, dataOffset int64, valueOffset uint32) (any, error) {
	if count <= 4 {
		buf := make([]byte, 4)
		r.PutUint32(buf, valueOffset)
		values := make([]int8, count)
		for i := uint32(0); i < count; i++ {
			values[i] = int8(buf[i])
		}
		if count == 1 {
			return values[0], nil
		}
		return values, nil
	}

	data, err := r.ReadBytes(dataOffset, int(count))
	if err != nil {
		return nil, err
	}
	values := make([]int8, count)
	for i, b := range data {
		values[i] = int8(b)
	}
	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

func (h *Handler) readSShorts(r *imxbin.Reader, count uint32, dataOffset int64, valueOffset uint32) (any, error) {
	if count == 0 {
		return []int16{}, nil
	}

	values := make([]int16, count)
	if count <= 2 {
		buf := make([]byte, 4)
		r.PutUint32(buf, valueOffset)
		values[0] = int16(r.Uint16(buf[0:2]))
		if count > 1 {
			values[1] = int16(r.Uint16(buf[2:4]))
		}
	} else {
		for i := uint32(0); i < count; i++ {
			val, err := r.ReadInt16(dataOffset + int64(i*2))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

func (h *Handler) readSLongs(r *imxbin.Reader, count uint32, dataOffset int64, valueOffset uint32) (any, error) {
	if count == 0 {
		return []int32{}, nil
	}

	values := make([]int32, count)
	if count == 1 {
		values[0] = int32(valueOffset)
	} else {
		for i := uint32(0); i < count; i++ {
			val, err := r.ReadInt32(dataOffset + int64(i*4))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

func (h *Handler) readSRationals(r *imxbin.Reader, count uint32, dataOffset int64) (any, error) {
	if count == 0 {
		return []string{}, nil
	}

	values := make([]string, count)
	for i := uint32(0); i < count; i++ {
		offset := dataOffset + int64(i*8)
		num, err := r.ReadInt32(offset)
		if err != nil {
			return nil, err
		}
		denom, err := r.ReadInt32(offset + 4)
		if err != nil {
			return nil, err
		}
		values[i] = fmt.Sprintf("%d/%d", num, denom)
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

// TagName returns the human-readable name for a Canon tag.
func (h *Handler) TagName(tagID uint16) string {
	if name, ok := tagNames[tagID]; ok {
		return name
	}
	return fmt.Sprintf("0x%04X", tagID)
}

// getTypeSize returns the size in bytes for a TIFF type.
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

// getTypeName returns the name of a TIFF type.
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

// tagNames maps Canon MakerNote tag IDs to human-readable names.
// Reference: https://exiftool.org/TagNames/Canon.html
var tagNames = map[uint16]string{
	0x0001: "CameraSettings",
	0x0002: "FocalLength",
	0x0003: "FlashInfo",
	0x0004: "ShotInfo",
	0x0005: "Panorama",
	0x0006: "ImageType",
	0x0007: "FirmwareVersion",
	0x0008: "FileNumber",
	0x0009: "OwnerName",
	0x000A: "UnknownD30",
	0x000C: "SerialNumber",
	0x000D: "CameraInfo",
	0x000E: "FileLength",
	0x000F: "CustomFunctions",
	0x0010: "ModelID",
	0x0011: "MovieInfo",
	0x0012: "AFInfo",
	0x0013: "ThumbnailImageValidArea",
	0x0015: "SerialNumberFormat",
	0x001A: "SuperMacro",
	0x001C: "DateStampMode",
	0x001D: "MyColors",
	0x001E: "FirmwareRevision",
	0x0023: "Categories",
	0x0024: "FaceDetect1",
	0x0025: "FaceDetect2",
	0x0026: "AFInfo2",
	0x0027: "ContrastInfo",
	0x0028: "ImageUniqueID",
	0x0029: "WBInfo",
	0x002F: "FaceDetect3",
	0x0035: "TimeInfo",
	0x0038: "BatteryType",
	0x003C: "AFInfo3",
	0x0081: "RawDataOffset",
	0x0083: "OriginalDecisionDataOffset",
	0x0090: "CustomFunctions1D",
	0x0091: "PersonalFunctions",
	0x0092: "PersonalFunctionValues",
	0x0093: "CanonFileInfo",
	0x0094: "AFPointsInFocus1D",
	0x0095: "LensModel",
	0x0096: "SerialInfo",
	0x0097: "DustRemovalData",
	0x0098: "CropInfo",
	0x0099: "CustomFunctions2",
	0x009A: "AspectInfo",
	0x00A0: "ProcessingInfo",
	0x00A1: "ToneCurveTable",
	0x00A2: "SharpnessTable",
	0x00A3: "SharpnessFreqTable",
	0x00A4: "WhiteBalanceTable",
	0x00A9: "ColorBalance",
	0x00AA: "MeasuredColor",
	0x00AE: "ColorTemperature",
	0x00B0: "CanonFlags",
	0x00B1: "ModifiedInfo",
	0x00B2: "ToneCurveMatching",
	0x00B3: "WhiteBalanceMatching",
	0x00B4: "ColorSpace",
	0x00B5: "PreviewImageInfo",
	0x00B6: "VRDOffset",
	0x00C0: "SensorInfo",
	0x00D0: "VRDOffset",
	0x00E0: "SensorInfo",
	0x4001: "ColorData",
	0x4002: "CRWParam",
	0x4003: "ColorInfo",
	0x4005: "Flavor",
	0x4008: "PictureStyleUserDef",
	0x4009: "PictureStylePC",
	0x4010: "CustomPictureStyleFileName",
	0x4013: "AFMicroAdj",
	0x4015: "VignettingCorr",
	0x4016: "VignettingCorr2",
	0x4018: "LightingOpt",
	0x4019: "LensInfo",
	0x4020: "AmbienceInfo",
	0x4021: "MultiExp",
	0x4024: "FilterInfo",
	0x4025: "HDRInfo",
	0x4028: "AFConfig",
	0x403F: "RawBurstModeRoll",
}
