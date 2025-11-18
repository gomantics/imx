package formats

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ExtractPNG extracts metadata from a PNG file.
func ExtractPNG(r io.ReadSeeker) (*Result, error) {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Read PNG signature (8 bytes)
	sig := make([]byte, 8)
	_, err = r.Read(sig)
	if err != nil {
		return nil, fmt.Errorf("failed to read PNG signature: %w", err)
	}

	// Verify PNG signature
	pngSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i := 0; i < 8; i++ {
		if sig[i] != pngSig[i] {
			return nil, fmt.Errorf("%w: invalid PNG file", ErrInvalidData)
		}
	}

	result := newResult()
	hasICC := false

	// Read chunks
	for {
		// Read chunk length (4 bytes, big-endian)
		lengthBytes := make([]byte, 4)
		_, err = r.Read(lengthBytes)
		if err != nil {
			break
		}
		length := int(binary.BigEndian.Uint32(lengthBytes))

		// Read chunk type (4 bytes)
		chunkType := make([]byte, 4)
		_, err = r.Read(chunkType)
		if err != nil {
			break
		}
		chunkTypeStr := string(chunkType)

		// Read chunk data
		chunkData := make([]byte, length)
		if length > 0 {
			_, err = r.Read(chunkData)
			if err != nil {
				break
			}
		}

		// Read CRC (4 bytes, but we'll skip it)
		crc := make([]byte, 4)
		r.Read(crc)

		// Process IHDR chunk (Image Header)
		if chunkTypeStr == "IHDR" && length >= 13 {
			result.Width = int(binary.BigEndian.Uint32(chunkData[0:4]))
			result.Height = int(binary.BigEndian.Uint32(chunkData[4:8]))
			bitDepth := int(chunkData[8])
			colorType := int(chunkData[9])
			compressionMethod := int(chunkData[10])
			filterMethod := int(chunkData[11])
			interlaceMethod := int(chunkData[12])

			result.ColorDepth = bitDepth
			result.Additional["BitDepth"] = bitDepth
			result.Additional["ColorType"] = colorType
			result.Additional["CompressionMethod"] = compressionMethod
			result.Additional["FilterMethod"] = filterMethod
			result.Additional["InterlaceMethod"] = interlaceMethod

			// Determine color space based on color type
			switch colorType {
			case 0:
				result.ColorSpace = "Grayscale"
			case 2:
				result.ColorSpace = "RGB"
			case 3:
				result.ColorSpace = "Indexed"
			case 4:
				result.ColorSpace = "GrayscaleAlpha"
			case 6:
				result.ColorSpace = "RGBA"
			default:
				result.ColorSpace = "Unknown"
			}
		}

		// Process iCCP chunk (ICC Profile)
		if chunkTypeStr == "iCCP" {
			hasICC = true
		}

		// Process eXIf chunk (EXIF data)
		if chunkTypeStr == "eXIf" {
			// Parse EXIF from chunk data
			exifData, err := parseTIFF(chunkData)
			if err == nil {
				for k, v := range exifData {
					result.EXIF[k] = v
				}
			}
		}

		// Stop after IEND chunk
		if chunkTypeStr == "IEND" {
			break
		}
	}

	result.HasICCProfile = hasICC

	return result, nil
}
