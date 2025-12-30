package gif

import (
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
)

// parseHeader reads and validates the GIF header and Logical Screen Descriptor
// Returns the GIF version, starting position after header, and any parse errors
func parseHeader(r io.ReaderAt, buf *[11]byte) (string, int64, int, int, *parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()

	// Read and verify GIF header (6 bytes)
	_, err := r.ReadAt(buf[:gifHeaderSize], 0)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read GIF header: %w", err))
		return "", 0, 0, 0, nil, parseErr
	}

	version := string(buf[:gifHeaderSize])
	if version != gifVersion87a && version != gifVersion89a {
		parseErr.Add(fmt.Errorf("invalid GIF signature"))
		return "", 0, 0, 0, nil, parseErr
	}

	// Read Logical Screen Descriptor (7 bytes)
	_, err = r.ReadAt(buf[:logicalScreenDescSize], gifHeaderSize)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read logical screen descriptor: %w", err))
		return "", 0, 0, 0, nil, parseErr
	}

	// Extract image dimensions
	width := int(buf[0]) | (int(buf[1]) << 8)
	height := int(buf[2]) | (int(buf[3]) << 8)
	packed := buf[4]
	backgroundColorIndex := buf[5]
	pixelAspectRatio := buf[6]

	// Parse packed field
	hasGCT := (packed & maskGlobalColorTable) != 0
	colorResolution := int((packed&maskColorResolution)>>4) + 1
	sortFlag := (packed & maskSortFlag) != 0
	gctSize := 1 << ((packed & maskColorTableSize) + 1)

	pos := int64(gifHeaderTotalSize)

	// Skip Global Color Table if present
	if hasGCT {
		pos += int64(gctSize * colorTableEntrySize)
	}

	// Create GIF directory with header metadata
	gifDir := &parser.Directory{
		Name: "GIF",
		Tags: []parser.Tag{
			{
				ID:       parser.TagID("GIF:Version"),
				Name:     "GIFVersion",
				Value:    version[3:], // "87a" or "89a"
				DataType: "string",
			},
			{
				ID:       parser.TagID("GIF:ImageWidth"),
				Name:     "ImageWidth",
				Value:    uint16(width),
				DataType: "uint16",
			},
			{
				ID:       parser.TagID("GIF:ImageHeight"),
				Name:     "ImageHeight",
				Value:    uint16(height),
				DataType: "uint16",
			},
			{
				ID:       parser.TagID("GIF:HasColorMap"),
				Name:     "HasColorMap",
				Value:    hasGCT,
				DataType: "bool",
			},
			{
				ID:       parser.TagID("GIF:ColorResolutionDepth"),
				Name:     "ColorResolutionDepth",
				Value:    uint8(colorResolution),
				DataType: "uint8",
			},
			{
				ID:       parser.TagID("GIF:BitsPerPixel"),
				Name:     "BitsPerPixel",
				Value:    uint8((packed & maskColorTableSize) + 1),
				DataType: "uint8",
			},
			{
				ID:       parser.TagID("GIF:BackgroundColor"),
				Name:     "BackgroundColor",
				Value:    uint8(backgroundColorIndex),
				DataType: "uint8",
			},
		},
	}

	// Add optional tags
	if sortFlag {
		gifDir.Tags = append(gifDir.Tags, parser.Tag{
			ID:       parser.TagID("GIF:GlobalColorTableSorted"),
			Name:     "GlobalColorTableSorted",
			Value:    true,
			DataType: "bool",
		})
	}

	if pixelAspectRatio != 0 {
		// Pixel Aspect Ratio = (pixelAspectRatio + 15) / 64
		gifDir.Tags = append(gifDir.Tags, parser.Tag{
			ID:       parser.TagID("GIF:PixelAspectRatio"),
			Name:     "PixelAspectRatio",
			Value:    uint8(pixelAspectRatio),
			DataType: "uint8",
		})
	}

	return version, pos, width, height, gifDir, parseErr.OrNil()
}
