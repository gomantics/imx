package imx

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ExtractGIFMetadata extracts metadata from a GIF file.
func ExtractGIFMetadata(r io.ReadSeeker, md *ImageMetadata) error {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Read GIF signature (6 bytes)
	var sig [6]byte
	if _, err = io.ReadFull(r, sig[:]); err != nil {
		return fmt.Errorf("failed to read GIF signature: %w", err)
	}

	// Verify GIF signature (GIF87a or GIF89a)
	if string(sig[0:3]) != "GIF" || (sig[3] != 0x38 && sig[3] != 0x39) || sig[5] != 0x61 {
		return fmt.Errorf("invalid GIF file")
	}

	version := string(sig[3:6])
	md.setAdditional("Version", version)

	// Read Logical Screen Descriptor (7 bytes)
	var lsd [7]byte
	if _, err = io.ReadFull(r, lsd[:]); err != nil {
		return fmt.Errorf("failed to read GIF logical screen descriptor: %w", err)
	}

	// Width and Height (little-endian, 2 bytes each)
	md.Width = int(binary.LittleEndian.Uint16(lsd[0:2]))
	md.Height = int(binary.LittleEndian.Uint16(lsd[2:4]))

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

	md.ColorSpace = "Indexed"
	md.ColorDepth = colorResolution * 3 // Approximate
	md.setAdditional("GlobalColorTable", globalColorTableFlag)
	md.setAdditional("ColorResolution", colorResolution)
	md.setAdditional("SortFlag", sortFlag)
	md.setAdditional("GlobalColorTableSize", globalColorTableSize)
	md.setAdditional("BackgroundColorIndex", backgroundColorIndex)
	md.setAdditional("PixelAspectRatio", pixelAspectRatio)

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
		var blockType [1]byte
		if _, err = io.ReadFull(r, blockType[:]); err != nil {
			break
		}

		switch blockType[0] {
		case 0x21: // Extension introducer
			var extLabel [1]byte
			if _, err = io.ReadFull(r, extLabel[:]); err != nil {
				return nil
			}

			switch extLabel[0] {
			case 0xF9: // Graphic Control Extension
				// Read block size (should be 4)
				var blockSize [1]byte
				if _, err = io.ReadFull(r, blockSize[:]); err != nil {
					return nil
				}
				if blockSize[0] == 4 {
					var gceData [4]byte
					if _, err = io.ReadFull(r, gceData[:]); err != nil {
						return nil
					}
					// Check transparency flag
					if (gceData[0] & 0x01) != 0 {
						hasTransparency = true
					}
				}
				// Skip terminator
				var terminator [1]byte
				io.ReadFull(r, terminator[:])

			case 0xFF: // Application Extension (may contain animation info)
				var blockSize [1]byte
				if _, err = io.ReadFull(r, blockSize[:]); err != nil {
					return nil
				}
				if blockSize[0] == 11 {
					appData := make([]byte, 11)
					if _, err = io.ReadFull(r, appData); err != nil {
						return nil
					}
					if string(appData) == "NETSCAPE2.0" {
						hasAnimation = true
					}
					// Skip sub-blocks
					if err := skipSubBlocks(r); err != nil {
						return nil
					}
				}

			default:
				// Skip other extensions
				if err := skipSubBlocks(r); err != nil {
					return nil
				}
			}

		case 0x2C: // Image separator (start of image)
			frameCount++
			// Skip image descriptor and data
			var imgDesc [9]byte
			if _, err = io.ReadFull(r, imgDesc[:]); err != nil {
				return nil
			}

			// Check for local color table
			localColorTableFlag := (imgDesc[8] & 0x80) != 0
			if localColorTableFlag {
				localColorTableSize := int(imgDesc[8]&0x07) + 1
				colorTableSize := 3 * (1 << localColorTableSize)
				r.Seek(int64(colorTableSize), io.SeekCurrent)
			}

			// Skip image data
			var lzwMinCodeSize [1]byte
			io.ReadFull(r, lzwMinCodeSize[:])
			// Skip data sub-blocks
			if err := skipSubBlocks(r); err != nil {
				return nil
			}

		case 0x3B: // Trailer (end of GIF)
			return nil

		default:
			// Unknown block, skip
			break
		}
	}

	md.setAdditional("HasTransparency", hasTransparency)
	md.setAdditional("HasAnimation", hasAnimation)
	md.setAdditional("FrameCount", frameCount)

	return nil
}

func skipSubBlocks(r io.Reader) error {
	var sizeBuf [1]byte
	for {
		if _, err := io.ReadFull(r, sizeBuf[:]); err != nil {
			return err
		}
		size := sizeBuf[0]
		if size == 0 {
			return nil
		}
		if _, err := io.CopyN(io.Discard, r, int64(size)); err != nil {
			return err
		}
	}
}
