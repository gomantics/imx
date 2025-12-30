package flac

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
)

// Parser parses FLAC (Free Lossless Audio Codec) files.
//
// FLAC file structure:
//   - 4-byte marker "fLaC"
//   - Metadata blocks (STREAMINFO, VORBIS_COMMENT, PICTURE, etc.)
//   - Audio frames
//
// The parser uses io.ReaderAt for efficient random access without
// loading the entire file into memory. The parser is stateless and
// safe for concurrent use.
type Parser struct {
	// Stateless parser - no fields needed
}

// New creates a new FLAC parser
func New() *Parser {
	return &Parser{}
}

// Name returns the parser name
func (p *Parser) Name() string {
	return "FLAC"
}

// Detect checks if the data is a FLAC file by looking for "fLaC" marker
func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, 4)
	_, err := r.ReadAt(buf, 0)
	return err == nil && bytes.Equal(buf, []byte("fLaC"))
}

// Parse extracts metadata from a FLAC file
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()
	var dirs []parser.Directory

	// Verify FLAC marker
	marker := make([]byte, 4)
	_, err := r.ReadAt(marker, 0)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read FLAC marker: %w", err))
		return nil, parseErr
	}

	if !bytes.Equal(marker, []byte("fLaC")) {
		parseErr.Add(fmt.Errorf("invalid FLAC marker: %s", string(marker)))
		return nil, parseErr
	}

	pos := int64(4)

	// Parse metadata blocks
	for {
		blockDir, isLast, newPos, err := p.parseMetadataBlock(r, pos)
		if err != nil {
			parseErr.Add(err)
			break
		}

		pos = newPos

		if blockDir != nil && len(blockDir.Tags) > 0 {
			dirs = append(dirs, *blockDir)
		}

		if isLast {
			break
		}
	}

	return dirs, parseErr.OrNil()
}

// parseMetadataBlock parses a single FLAC metadata block and returns the directory, isLast flag, new position, and error
func (p *Parser) parseMetadataBlock(r io.ReaderAt, pos int64) (*parser.Directory, bool, int64, error) {
	// Read block header (4 bytes)
	header := make([]byte, 4)
	_, err := r.ReadAt(header, pos)
	if err != nil {
		return nil, false, pos, fmt.Errorf("failed to read metadata block header: %w", err)
	}
	pos += 4

	// Parse header
	isLast := (header[0] & 0x80) != 0
	blockType := header[0] & 0x7F
	blockLength := int64(header[1])<<16 | int64(header[2])<<8 | int64(header[3])

	// Validate block length to prevent excessive memory allocation
	if blockLength > maxBlockSize {
		return nil, false, pos, fmt.Errorf("metadata block size %d exceeds maximum %d", blockLength, maxBlockSize)
	}

	// Parse block based on type
	blockStart := pos
	pos += blockLength

	var dir *parser.Directory

	switch blockType {
	case blockTypeStreamInfo:
		dir = p.parseStreamInfo(r, blockStart, blockLength)
	case blockTypePadding:
		dir = p.parsePadding(blockLength)
	case blockTypeApplication:
		dir = p.parseApplication(r, blockStart, blockLength)
	case blockTypeSeekTable:
		dir = p.parseSeekTable(r, blockStart, blockLength)
	case blockTypeVorbisComment:
		dir = p.parseVorbisComment(r, blockStart, blockLength)
	case blockTypeCueSheet:
		dir = p.parseCueSheet(blockLength)
	case blockTypePicture:
		dir = p.parsePicture(r, blockStart, blockLength)
	default:
		// Unknown block type (>127 reserved)
		dir = &parser.Directory{
			Name: fmt.Sprintf("FLAC Block %d", blockType),
			Tags: []parser.Tag{
				{
					ID:       parser.TagID(fmt.Sprintf("FLAC:Block%d:Size", blockType)),
					Name:     "Size",
					Value:    blockLength,
					DataType: "uint32",
				},
			},
		}
	}

	return dir, isLast, pos, nil
}
