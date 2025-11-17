package imx

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ExtractPNGMetadata extracts metadata from a PNG file.
func ExtractPNGMetadata(r io.ReadSeeker, md *ImageMetadata) error {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Read PNG signature (8 bytes)
	var sig [8]byte
	if _, err = io.ReadFull(r, sig[:]); err != nil {
		return fmt.Errorf("failed to read PNG signature: %w", err)
	}

	// Verify PNG signature
	pngSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i := 0; i < 8; i++ {
		if sig[i] != pngSig[i] {
			return fmt.Errorf("invalid PNG file")
		}
	}

	hasICC := false

	// Read chunks
	for {
		// Read chunk length (4 bytes, big-endian)
		var lengthBytes [4]byte
		if _, err = io.ReadFull(r, lengthBytes[:]); err != nil {
			break
		}
		length := int(binary.BigEndian.Uint32(lengthBytes[:]))

		// Read chunk type (4 bytes)
		var chunkType [4]byte
		if _, err = io.ReadFull(r, chunkType[:]); err != nil {
			break
		}
		chunkTypeStr := string(chunkType[:])

		// Read chunk data
		chunkData := borrowBuffer(length)
		if length > 0 {
			if _, err = io.ReadFull(r, chunkData); err != nil {
				releaseBuffer(chunkData)
				break
			}
		}

		// Read CRC (4 bytes, but we'll skip it)
		var crc [4]byte
		io.ReadFull(r, crc[:])

		// Process IHDR chunk (Image Header)
		if chunkTypeStr == "IHDR" && length >= 13 {
			md.Width = int(binary.BigEndian.Uint32(chunkData[0:4]))
			md.Height = int(binary.BigEndian.Uint32(chunkData[4:8]))
			bitDepth := int(chunkData[8])
			colorType := int(chunkData[9])
			compressionMethod := int(chunkData[10])
			filterMethod := int(chunkData[11])
			interlaceMethod := int(chunkData[12])

			md.ColorDepth = bitDepth
			md.setAdditional("BitDepth", bitDepth)
			md.setAdditional("ColorType", colorType)
			md.setAdditional("CompressionMethod", compressionMethod)
			md.setAdditional("FilterMethod", filterMethod)
			md.setAdditional("InterlaceMethod", interlaceMethod)

			// Determine color space based on color type
			switch colorType {
			case 0:
				md.ColorSpace = "Grayscale"
			case 2:
				md.ColorSpace = "RGB"
			case 3:
				md.ColorSpace = "Indexed"
			case 4:
				md.ColorSpace = "GrayscaleAlpha"
			case 6:
				md.ColorSpace = "RGBA"
			default:
				md.ColorSpace = "Unknown"
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
				md.mergeEXIF(exifData)
			}
		}

		releaseBuffer(chunkData)

		// Stop after IEND chunk
		if chunkTypeStr == "IEND" {
			break
		}
	}

	md.HasICCProfile = hasICC

	return nil
}
