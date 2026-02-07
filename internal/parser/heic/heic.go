// Package heic implements a parser for HEIC/HEIF and AVIF image files.
//
// HEIC (High Efficiency Image Container) and AVIF (AV1 Image File Format)
// are both based on the ISO Base Media File Format (ISOBMFF). They share the
// same container structure, differing only in the image codec used (HEVC vs AV1).
// This parser extracts EXIF, XMP, and ICC metadata from both formats by parsing
// the box structure and building an item index.
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

// Parser parses HEIC/HEIF and AVIF image files.
// Both formats use the ISO Base Media File Format (ISOBMFF) container.
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
// This parser handles both HEIC/HEIF and AVIF formats.
func (p *Parser) Name() string {
	return "HEIC"
}

// Detect checks if the data is a HEIC/HEIF or AVIF file.
// Both formats use the same ISOBMFF container structure.
func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, 12)
	if _, err := r.ReadAt(buf[:12], 0); err != nil {
		return false
	}

	// Must start with ftyp box
	if !boxTypeEquals(buf[4:8], boxTypeFtyp) {
		return false
	}

	// Check major brand (HEIC or AVIF)
	brand := string(buf[8:12])
	for _, valid := range heicBrands {
		if brand == valid {
			return true
		}
	}
	for _, valid := range avifBrands {
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
