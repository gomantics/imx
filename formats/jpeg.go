package formats

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ExtractJPEG extracts metadata from a JPEG file.
func ExtractJPEG(r io.ReadSeeker) (*Result, error) {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 2)
	_, err = r.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read JPEG header: %w", err)
	}

	// Verify JPEG SOI marker
	if buf[0] != 0xFF || buf[1] != 0xD8 {
		return nil, fmt.Errorf("%w: invalid JPEG file", ErrInvalidData)
	}

	result := newResult()
	hasICC := false

	// Read through JPEG segments
	for {
		marker := make([]byte, 2)
		_, err = r.Read(marker)
		if err != nil {
			break
		}

		// Check for marker
		if marker[0] != 0xFF {
			break
		}

		markerType := marker[1]

		// Skip padding bytes (0xFF)
		for markerType == 0xFF {
			_, err = r.Read(marker)
			if err != nil {
				return nil, err
			}
			markerType = marker[1]
		}

		// End of image
		if markerType == 0xD9 {
			break
		}

		// Restart markers (0xD0-0xD7) have no length
		if markerType >= 0xD0 && markerType <= 0xD7 {
			continue
		}

		// Read segment length
		lengthBytes := make([]byte, 2)
		_, err = r.Read(lengthBytes)
		if err != nil {
			break
		}
		length := int(binary.BigEndian.Uint16(lengthBytes)) - 2

		// Handle different segment types
		switch markerType {
		case 0xE0: // APP0 (JFIF)
			// Skip JFIF data
			r.Seek(int64(length), io.SeekCurrent)

		case 0xE1: // APP1 (EXIF)
			segmentData := make([]byte, length)
			_, err = r.Read(segmentData)
			if err != nil {
				continue
			}
			// Check for EXIF identifier
			if len(segmentData) >= 6 && string(segmentData[0:6]) == "Exif\x00\x00" {
				// Parse EXIF from segment data
				exifData, err := parseTIFF(segmentData[6:])
				if err == nil {
					for k, v := range exifData {
						result.EXIF[k] = v
					}
				}
			}

		case 0xE2: // APP2 (ICC Profile)
			segmentData := make([]byte, length)
			_, err = r.Read(segmentData)
			if err != nil {
				continue
			}
			// Check for ICC profile identifier
			if len(segmentData) >= 11 && string(segmentData[0:11]) == "ICC_PROFILE" {
				hasICC = true
			}

		case 0xC0, 0xC1, 0xC2, 0xC3, 0xC5, 0xC6, 0xC7, 0xC9, 0xCA, 0xCB, 0xCD, 0xCE, 0xCF:
			// SOF (Start of Frame) segments - contain image dimensions
			readLen := length
			if readLen > 9 {
				readLen = 9
			}
			sofData := make([]byte, readLen)
			_, err = r.Read(sofData)
			if err != nil {
				continue
			}
			if len(sofData) >= 5 {
				// Precision (bits per sample)
				precision := int(sofData[0])
				result.ColorDepth = precision * 3 // Assuming RGB
				result.Additional["BitsPerSample"] = precision

				// Height and Width (big-endian)
				height := int(binary.BigEndian.Uint16(sofData[1:3]))
				width := int(binary.BigEndian.Uint16(sofData[3:5]))
				result.Height = height
				result.Width = width

				// Number of components
				if len(sofData) >= 6 {
					numComponents := int(sofData[5])
					result.Additional["Components"] = numComponents
					switch numComponents {
					case 1:
						result.ColorSpace = "Grayscale"
					case 3:
						result.ColorSpace = "RGB"
					case 4:
						result.ColorSpace = "CMYK"
					default:
						result.ColorSpace = "Unknown"
					}
				}
			}
			// Skip remaining segment data
			if length > 9 {
				r.Seek(int64(length-9), io.SeekCurrent)
			}

		default:
			// Skip unknown segments
			r.Seek(int64(length), io.SeekCurrent)
		}
	}

	result.HasICCProfile = hasICC

	// Set default color space if not set
	if result.ColorSpace == "" {
		result.ColorSpace = "RGB"
	}

	return result, nil
}
