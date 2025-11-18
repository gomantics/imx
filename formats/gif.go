package formats

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ExtractGIF extracts metadata from a GIF file.
func ExtractGIF(r io.ReadSeeker) (*Result, error) {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Read GIF signature (6 bytes)
	sig := make([]byte, 6)
	_, err = r.Read(sig)
	if err != nil {
		return nil, fmt.Errorf("failed to read GIF signature: %w", err)
	}

	// Verify GIF signature (GIF87a or GIF89a)
	if string(sig[0:3]) != "GIF" || (sig[3] != 0x38 && sig[3] != 0x39) || sig[5] != 0x61 {
		return nil, fmt.Errorf("%w: invalid GIF file", ErrInvalidData)
	}

	version := string(sig[3:6])
	result := newResult()
	result.Additional["Version"] = version

	// Read Logical Screen Descriptor (7 bytes)
	lsd := make([]byte, 7)
	_, err = r.Read(lsd)
	if err != nil {
		return nil, fmt.Errorf("failed to read GIF logical screen descriptor: %w", err)
	}

	// Width and Height (little-endian, 2 bytes each)
	result.Width = int(binary.LittleEndian.Uint16(lsd[0:2]))
	result.Height = int(binary.LittleEndian.Uint16(lsd[2:4]))

	// Packed fields
	packed := lsd[4]
	globalColorTableFlag := (packed & 0x80) != 0
	colorResolution := int((packed>>4)&0x07) + 1
	sortFlag := (packed & 0x08) != 0
	globalColorTableSize := int(packed&0x07) + 1

	// Background color index
	backgroundColorIndex := int(lsd[5])

	// Pixel aspect ratio
	pixelAspectRatio := lsd[6]

	result.ColorSpace = "Indexed"
	result.ColorDepth = colorResolution * 3 // Approximate
	result.Additional["GlobalColorTable"] = globalColorTableFlag
	result.Additional["ColorResolution"] = colorResolution
	result.Additional["SortFlag"] = sortFlag
	result.Additional["GlobalColorTableSize"] = globalColorTableSize
	result.Additional["BackgroundColorIndex"] = backgroundColorIndex
	result.Additional["PixelAspectRatio"] = pixelAspectRatio

	// Skip global color table if present
	if globalColorTableFlag {
		colorTableSize := 3 * (1 << globalColorTableSize)
		r.Seek(int64(colorTableSize), io.SeekCurrent)
	}

	// Check for transparency and animation by scanning extension blocks
	hasTransparency := false
	hasAnimation := false
	frameCount := 0

	for {
		blockType := make([]byte, 1)
		_, err = r.Read(blockType)
		if err != nil {
			break
		}

		switch blockType[0] {
		case 0x21: // Extension introducer
			extLabel := make([]byte, 1)
			_, err = r.Read(extLabel)
			if err != nil {
				return nil, err
			}

			switch extLabel[0] {
			case 0xF9: // Graphic Control Extension
				// Read block size (should be 4)
				blockSize := make([]byte, 1)
				r.Read(blockSize)
				if blockSize[0] == 4 {
					gceData := make([]byte, 4)
					r.Read(gceData)
					// Check transparency flag
					if (gceData[0] & 0x01) != 0 {
						hasTransparency = true
					}
				}
				// Skip terminator
				r.Read(make([]byte, 1))

			case 0xFF: // Application Extension (may contain animation info)
				blockSize := make([]byte, 1)
				r.Read(blockSize)
				if blockSize[0] == 11 {
					appData := make([]byte, 11)
					r.Read(appData)
					if string(appData) == "NETSCAPE2.0" {
						hasAnimation = true
					}
					// Skip sub-blocks
					for {
						subBlockSize := make([]byte, 1)
						r.Read(subBlockSize)
						if subBlockSize[0] == 0 {
							break
						}
						subBlockData := make([]byte, int(subBlockSize[0]))
						r.Read(subBlockData)
					}
				}

			default:
				// Skip other extensions
				for {
					subBlockSize := make([]byte, 1)
					r.Read(subBlockSize)
					if subBlockSize[0] == 0 {
						break
					}
					subBlockData := make([]byte, int(subBlockSize[0]))
					r.Read(subBlockData)
				}
			}

		case 0x2C: // Image separator (start of image)
			frameCount++
			// Skip image descriptor and data
			imgDesc := make([]byte, 9)
			r.Read(imgDesc)

			// Check for local color table
			localColorTableFlag := (imgDesc[8] & 0x80) != 0
			if localColorTableFlag {
				localColorTableSize := int(imgDesc[8]&0x07) + 1
				colorTableSize := 3 * (1 << localColorTableSize)
				r.Seek(int64(colorTableSize), io.SeekCurrent)
			}

			// Skip image data
			lzwMinCodeSize := make([]byte, 1)
			r.Read(lzwMinCodeSize)
			// Skip data sub-blocks
			for {
				subBlockSize := make([]byte, 1)
				r.Read(subBlockSize)
				if subBlockSize[0] == 0 {
					break
				}
				subBlockData := make([]byte, int(subBlockSize[0]))
				r.Read(subBlockData)
			}

		case 0x3B: // Trailer (end of GIF)
			return result, nil

		default:
			// Unknown block, skip
		}
	}

	result.Additional["HasTransparency"] = hasTransparency
	result.Additional["HasAnimation"] = hasAnimation
	result.Additional["FrameCount"] = frameCount

	return result, nil
}
