package gif

import (
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/xmp"
)

// Parser parses GIF image files.
//
// Supported metadata:
//   - GIF Header (version, dimensions, color information)
//   - Animation metadata (frame count, loop count)
//   - XMP (in Application Extension blocks)
//   - Comment Extension blocks
//
// The parser is stateless and safe for concurrent use.
type Parser struct {
	xmp *xmp.Parser
}

// New creates a new GIF parser
func New() *Parser {
	return &Parser{
		xmp: xmp.New(),
	}
}

// Name returns the parser name
func (p *Parser) Name() string {
	return "GIF"
}

// Detect checks if the data is a GIF file
func (p *Parser) Detect(r io.ReaderAt) bool {
	var buf [6]byte
	_, err := r.ReadAt(buf[:], 0)
	if err != nil {
		return false
	}

	// Check for GIF87a or GIF89a signature
	return (string(buf[:]) == "GIF87a" || string(buf[:]) == "GIF89a")
}

// Parse extracts metadata from a GIF file
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()
	var dirs []parser.Directory
	var buf [11]byte // Reusable buffer for reads

	// Parse header and get GIF directory with metadata
	version, pos, _, _, gifDir, headerErr := parseHeader(r, &buf)
	if headerErr != nil {
		return nil, headerErr
	}

	// Track animation metadata during parse (no double-scan)
	frameCount := 0
	loopCount := -1 // -1 means not set, 0 means loop forever

	// Create directory for comments
	commentDir := &parser.Directory{
		Name: "GIF-Comments",
		Tags: []parser.Tag{},
	}

	// Parse data stream and count frames in single pass
	for {
		var separator [1]byte
		_, err := r.ReadAt(separator[:], pos)
		if err != nil {
			if err == io.EOF {
				break
			}
			parseErr.Add(err)
			break
		}

		pos++

		switch separator[0] {
		case separatorExtension:
			extensionDirs, commentTags, newLoopCount, newPos := parseExtensionWithLoopCount(r, pos, &buf, p.xmp)
			dirs = append(dirs, extensionDirs...)
			commentDir.Tags = append(commentDir.Tags, commentTags...)
			if newLoopCount >= 0 {
				loopCount = newLoopCount
			}
			pos = newPos

		case separatorImageDescriptor:
			frameCount++
			var ok bool
			pos, ok = skipImage(r, pos, &buf)
			if !ok {
				goto done
			}

		case separatorTrailer:
			goto done

		case separatorBlockTerminator:
			// Continue

		default:
			// Unknown block - try to skip it gracefully instead of aborting
			parseErr.Add(fmt.Errorf("unknown separator 0x%02X at offset %d", separator[0], pos-1))
			// Try to skip this unknown block by reading the next byte to see if it's a size
			var skipBuf [1]byte
			_, err := r.ReadAt(skipBuf[:], pos)
			if err == nil && skipBuf[0] > 0 && skipBuf[0] < 255 {
				// Looks like a block size, try to skip
				pos = skipDataSubBlocks(r, pos, &buf)
			} else {
				// Can't determine block structure, stop parsing
				goto done
			}
		}
	}

done:
	// Add animation metadata if animated
	if frameCount > 1 {
		gifDir.Tags = append(gifDir.Tags, parser.Tag{
			ID:       parser.TagID("GIF:FrameCount"),
			Name:     "FrameCount",
			Value:    uint16(frameCount),
			DataType: "uint16",
		})
	}

	if frameCount > 1 && loopCount >= 0 {
		gifDir.Tags = append(gifDir.Tags, parser.Tag{
			ID:       parser.TagID("GIF:AnimationIterations"),
			Name:     "AnimationIterations",
			Value:    uint16(loopCount),
			DataType: "uint16",
		})
	}

	dirs = append([]parser.Directory{*gifDir}, dirs...)

	// Add comment directory if it has tags
	if len(commentDir.Tags) > 0 {
		dirs = append(dirs, *commentDir)
	}

	// Log parser info
	_ = version // Keep version for potential future use

	return dirs, parseErr.OrNil()
}
