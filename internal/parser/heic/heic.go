// Package heic implements a parser for HEIC/HEIF image files.
//
// HEIC (High Efficiency Image Container) is based on the ISO Base Media File
// Format (ISOBMFF). This parser extracts EXIF, XMP, and ICC metadata from
// HEIC/HEIF files by parsing the box structure and building an item index.
package heic

import (
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/icc"
	"github.com/gomantics/imx/internal/parser/tiff"
	"github.com/gomantics/imx/internal/parser/xmp"
)

// maxBoxScan is the maximum number of bytes to scan when searching for boxes.
const maxBoxScan = 100 * 1024 * 1024 // 100MB

// Parser parses HEIC/HEIF image files.
type Parser struct {
	tiff *tiff.Parser
	xmp  *xmp.Parser
	icc  *icc.Parser
}

// New creates a new HEIC parser.
func New() *Parser {
	return &Parser{
		tiff: tiff.New(),
		xmp:  xmp.New(),
		icc:  icc.New(),
	}
}

// Name returns the parser name.
func (p *Parser) Name() string {
	return "HEIC"
}

// Detect checks if the data is a HEIC/HEIF file.
func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, 12)
	if _, err := r.ReadAt(buf[:12], 0); err != nil {
		return false
	}

	// Must start with ftyp box
	if !boxTypeEquals(buf[4:8], boxTypeFtyp) {
		return false
	}

	// Check major brand
	brand := string(buf[8:12])
	for _, valid := range validBrands {
		if brand == valid {
			return true
		}
	}

	return false
}

// Parse extracts metadata from HEIC/HEIF file.
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()

	// Find meta box (required for all metadata)
	metaBox, err := findBox(r, boxTypeMeta, 0, maxBoxScan)
	if err != nil {
		parseErr.Add(fmt.Errorf("meta box not found (file may not be a valid HEIC/HEIF or may be corrupted): %w", err))
		return nil, parseErr
	}

	// Build HEIF index from meta box
	index, err := buildHeifIndex(r, metaBox, maxBoxScan)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to build HEIF index (file structure may be invalid or unsupported): %w", err))
		return nil, parseErr
	}

	// Extract metadata using index
	dirs := p.extractMetadata(r, index)

	return dirs, parseErr.OrNil()
}
