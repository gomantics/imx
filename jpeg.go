package imx

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ExtractJPEGMetadata extracts metadata from a JPEG file.
func ExtractJPEGMetadata(r io.ReadSeeker, md *ImageMetadata) error {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	var header [2]byte
	if _, err = io.ReadFull(r, header[:]); err != nil {
		return fmt.Errorf("failed to read JPEG header: %w", err)
	}

	// Verify JPEG SOI marker
	if header[0] != 0xFF || header[1] != 0xD8 {
		return fmt.Errorf("invalid JPEG file")
	}

	hasICC := false

	// Read through JPEG segments
	for {
		var marker [2]byte
		if _, err = io.ReadFull(r, marker[:]); err != nil {
			break
		}

		// Check for marker
		if marker[0] != 0xFF {
			break
		}

		markerType := marker[1]

		// Skip padding bytes (0xFF)
		for markerType == 0xFF {
			if _, err = io.ReadFull(r, marker[:]); err != nil {
				return nil
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
		var lengthBytes [2]byte
		if _, err = io.ReadFull(r, lengthBytes[:]); err != nil {
			break
		}
		length := int(binary.BigEndian.Uint16(lengthBytes[:])) - 2

		// Handle different segment types
		switch markerType {
		case 0xE0: // APP0 (JFIF)
			// Skip JFIF data
			r.Seek(int64(length), io.SeekCurrent)

		case 0xE1: // APP1 (EXIF)
			segmentData := borrowBuffer(length)
			if _, err = io.ReadFull(r, segmentData); err != nil {
				releaseBuffer(segmentData)
				continue
			}
			// Check for EXIF identifier
			if len(segmentData) >= 6 && string(segmentData[0:6]) == "Exif\x00\x00" {
				// Parse EXIF from segment data
				exifData, err := parseTIFF(segmentData[6:])
				if err == nil {
					md.mergeEXIF(exifData)
				}
			}
			releaseBuffer(segmentData)

		case 0xE2: // APP2 (ICC Profile)
			segmentData := borrowBuffer(length)
			if _, err = io.ReadFull(r, segmentData); err != nil {
				releaseBuffer(segmentData)
				continue
			}
			// Check for ICC profile identifier
			if len(segmentData) >= 11 && string(segmentData[0:11]) == "ICC_PROFILE" {
				hasICC = true
			}
			releaseBuffer(segmentData)

		case 0xC0, 0xC1, 0xC2, 0xC3, 0xC5, 0xC6, 0xC7, 0xC9, 0xCA, 0xCB, 0xCD, 0xCE, 0xCF:
			// SOF (Start of Frame) segments - contain image dimensions
			readLen := 9
			if length < readLen {
				readLen = length
			}
			var sofBuf [9]byte
			if _, err = io.ReadFull(r, sofBuf[:readLen]); err != nil {
				continue
			}
			sofData := sofBuf[:readLen]
			if len(sofData) >= 5 {
				// Precision (bits per sample)
				precision := int(sofData[0])
				md.ColorDepth = precision * 3 // Assuming RGB
				md.setAdditional("BitsPerSample", precision)

				// Height and Width (big-endian)
				height := int(binary.BigEndian.Uint16(sofData[1:3]))
				width := int(binary.BigEndian.Uint16(sofData[3:5]))
				md.Height = height
				md.Width = width

				// Number of components
				if len(sofData) >= 6 {
					numComponents := int(sofData[5])
					md.setAdditional("Components", numComponents)
					switch numComponents {
					case 1:
						md.ColorSpace = "Grayscale"
					case 3:
						md.ColorSpace = "RGB"
					case 4:
						md.ColorSpace = "CMYK"
					default:
						md.ColorSpace = "Unknown"
					}
				}
			}
			// Skip remaining segment data
			if length > readLen {
				r.Seek(int64(length-readLen), io.SeekCurrent)
			}

		default:
			// Skip unknown segments
			r.Seek(int64(length), io.SeekCurrent)
		}
	}

	md.HasICCProfile = hasICC

	// Set default color space if not set
	if md.ColorSpace == "" {
		md.ColorSpace = "RGB"
	}

	return nil
}
