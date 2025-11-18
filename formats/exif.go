package formats

import (
	"encoding/binary"
	"fmt"
	"io"
)

// EXIF tag IDs (commonly used)
const (
	exifTagDateTime          = 0x0132
	exifTagMake              = 0x010F
	exifTagModel             = 0x0110
	exifTagOrientation       = 0x0112
	exifTagXResolution       = 0x011A
	exifTagYResolution       = 0x011B
	exifTagResolutionUnit    = 0x0128
	exifTagSoftware          = 0x0131
	exifTagArtist            = 0x013B
	exifTagCopyright         = 0x8298
	exifTagExifIFD           = 0x8769
	exifTagGPSIFD            = 0x8825
	exifTagISO               = 0x8827
	exifTagExposureTime      = 0x829A
	exifTagFNumber           = 0x829D
	exifTagDateTimeOriginal  = 0x9003
	exifTagDateTimeDigitized = 0x9004
)

// EXIF data types
const (
	exifTypeByte      = 1
	exifTypeASCII     = 2
	exifTypeShort     = 3
	exifTypeLong      = 4
	exifTypeRational  = 5
	exifTypeUndefined = 7
	exifTypeSLong     = 9
	exifTypeSRational = 10
)

// ParseEXIF extracts EXIF data from a JPEG or PNG file.
// It searches for the EXIF APP1 segment and parses the TIFF structure.
func ParseEXIF(r io.ReadSeeker) (map[string]interface{}, error) {
	exif := make(map[string]interface{})

	// For JPEG, EXIF is in APP1 segment (0xFFE1)
	// For PNG, EXIF is in eXIf chunk
	// This is a simplified parser that handles common cases

	// Try to find EXIF APP1 marker in JPEG
	pos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return exif, nil // Return empty, not an error
	}

	buf := make([]byte, 4)
	_, err = r.Read(buf)
	if err != nil {
		return exif, nil
	}

	// Check for JPEG APP1 marker (0xFFE1)
	if buf[0] == 0xFF && buf[1] == 0xE1 {
		// Read segment length
		length := int(buf[2])<<8 | int(buf[3])
		if length < 6 {
			return exif, nil
		}

		// Read segment data
		segmentData := make([]byte, length-2)
		_, err = r.Read(segmentData)
		if err != nil {
			return exif, nil
		}

		// Check for "Exif\0\0" identifier
		if len(segmentData) >= 6 && string(segmentData[0:6]) == "Exif\x00\x00" {
			// Parse TIFF header and IFD
			exifData, err := parseTIFF(segmentData[6:])
			if err == nil {
				for k, v := range exifData {
					exif[k] = v
				}
			}
		}
	}

	// Reset position
	r.Seek(pos, io.SeekStart)
	return exif, nil
}

// parseTIFF parses a TIFF structure (used by EXIF)
func parseTIFF(data []byte) (map[string]interface{}, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for TIFF header")
	}

	exif := make(map[string]interface{})

	// Check byte order (II for little-endian, MM for big-endian)
	var byteOrder binary.ByteOrder
	if data[0] == 0x49 && data[1] == 0x49 {
		byteOrder = binary.LittleEndian
	} else if data[0] == 0x4D && data[1] == 0x4D {
		byteOrder = binary.BigEndian
	} else {
		return nil, fmt.Errorf("invalid TIFF byte order")
	}

	// Check TIFF magic number (42)
	if byteOrder.Uint16(data[2:4]) != 42 {
		return nil, fmt.Errorf("invalid TIFF magic number")
	}

	// Get offset to first IFD
	ifdOffset := int(byteOrder.Uint32(data[4:8]))
	if ifdOffset >= len(data) {
		return nil, fmt.Errorf("IFD offset out of bounds")
	}

	// Parse IFD
	parseIFD(data, ifdOffset, byteOrder, exif, 0)

	return exif, nil
}

// parseIFD parses an Image File Directory
func parseIFD(data []byte, offset int, byteOrder binary.ByteOrder, exif map[string]interface{}, depth int) {
	if depth > 10 || offset+2 > len(data) {
		return // Prevent infinite recursion
	}

	// Get number of directory entries
	if offset+2 > len(data) {
		return
	}
	numEntries := int(byteOrder.Uint16(data[offset : offset+2]))

	offset += 2

	// Parse each entry
	for i := 0; i < numEntries && offset+12 <= len(data); i++ {
		tag := byteOrder.Uint16(data[offset : offset+2])
		dataType := byteOrder.Uint16(data[offset+2 : offset+4])
		count := byteOrder.Uint32(data[offset+4 : offset+8])
		valueOffset := byteOrder.Uint32(data[offset+8 : offset+12])

		// Read tag value
		var value interface{}
		valueSize := getDataTypeSize(dataType) * int(count)

		if valueSize <= 4 {
			// Value is stored directly in the offset field
			value = readTagValue(data[offset+8:offset+12], dataType, count, byteOrder)
		} else {
			// Value is stored at the offset
			valOffset := int(valueOffset)
			if valOffset < len(data) && valOffset+valueSize <= len(data) {
				value = readTagValue(data[valOffset:valOffset+valueSize], dataType, count, byteOrder)
			}
		}

		// Map tag to name and store
		if tagName := getEXIFTagName(tag); tagName != "" {
			exif[tagName] = value
		}

		// Handle IFD pointers
		if tag == exifTagExifIFD && valueSize <= 4 {
			ifdPtr := int(valueOffset)
			if ifdPtr < len(data) {
				parseIFD(data, ifdPtr, byteOrder, exif, depth+1)
			}
		}

		offset += 12
	}
}

// getDataTypeSize returns the size in bytes of an EXIF data type
func getDataTypeSize(dataType uint16) int {
	switch dataType {
	case exifTypeByte, exifTypeASCII, exifTypeUndefined:
		return 1
	case exifTypeShort:
		return 2
	case exifTypeLong, exifTypeSLong:
		return 4
	case exifTypeRational, exifTypeSRational:
		return 8
	default:
		return 1
	}
}

// readTagValue reads a tag value based on its data type
func readTagValue(data []byte, dataType uint16, count uint32, byteOrder binary.ByteOrder) interface{} {
	switch dataType {
	case exifTypeByte, exifTypeUndefined:
		if count == 1 && len(data) >= 1 {
			return uint8(data[0])
		}
		return data[:min(int(count), len(data))]

	case exifTypeASCII:
		if len(data) >= int(count) {
			str := string(data[:count])
			// Remove null terminator
			if len(str) > 0 && str[len(str)-1] == 0 {
				str = str[:len(str)-1]
			}
			return str
		}
		return ""

	case exifTypeShort:
		if count == 1 && len(data) >= 2 {
			return byteOrder.Uint16(data[0:2])
		}
		vals := make([]uint16, min(int(count), len(data)/2))
		for i := range vals {
			if i*2+2 <= len(data) {
				vals[i] = byteOrder.Uint16(data[i*2 : i*2+2])
			}
		}
		return vals

	case exifTypeLong, exifTypeSLong:
		if count == 1 && len(data) >= 4 {
			val := byteOrder.Uint32(data[0:4])
			if dataType == exifTypeSLong {
				return int32(val)
			}
			return val
		}
		vals := make([]uint32, min(int(count), len(data)/4))
		for i := range vals {
			if i*4+4 <= len(data) {
				vals[i] = byteOrder.Uint32(data[i*4 : i*4+4])
			}
		}
		return vals

	case exifTypeRational, exifTypeSRational:
		if count == 1 && len(data) >= 8 {
			num := byteOrder.Uint32(data[0:4])
			den := byteOrder.Uint32(data[4:8])
			if den == 0 {
				return float64(0)
			}
			if dataType == exifTypeSRational {
				return float64(int32(num)) / float64(int32(den))
			}
			return float64(num) / float64(den)
		}
		return nil

	default:
		return nil
	}
}

// getEXIFTagName returns the human-readable name for an EXIF tag
func getEXIFTagName(tag uint16) string {
	switch tag {
	case exifTagDateTime:
		return "DateTime"
	case exifTagMake:
		return "Make"
	case exifTagModel:
		return "Model"
	case exifTagOrientation:
		return "Orientation"
	case exifTagXResolution:
		return "XResolution"
	case exifTagYResolution:
		return "YResolution"
	case exifTagResolutionUnit:
		return "ResolutionUnit"
	case exifTagSoftware:
		return "Software"
	case exifTagArtist:
		return "Artist"
	case exifTagCopyright:
		return "Copyright"
	case exifTagISO:
		return "ISO"
	case exifTagExposureTime:
		return "ExposureTime"
	case exifTagFNumber:
		return "FNumber"
	case exifTagDateTimeOriginal:
		return "DateTimeOriginal"
	case exifTagDateTimeDigitized:
		return "DateTimeDigitized"
	default:
		return ""
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
