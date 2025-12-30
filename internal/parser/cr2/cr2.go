package cr2

import (
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/tiff"
)

// Parser parses Canon CR2 (Canon Raw 2) files.
// CR2 is based on TIFF format with Canon-specific extensions.
// The parser is stateless and safe for concurrent use.
type Parser struct {
	tiff *tiff.Parser
}

// New creates a new CR2 parser.
func New() *Parser {
	return &Parser{
		tiff: tiff.New(),
	}
}

// Name returns the parser name.
func (p *Parser) Name() string {
	return "CR2"
}

// Detect checks if the data is a CR2 file.
// CR2 files have a 16-byte header:
//   - Bytes 0-7: TIFF header (byte order + magic 42 + IFD offset)
//   - Bytes 8-9: CR2 magic "CR" (0x43 0x52)
//   - Byte 10: Major version (0x02 for CR2)
//   - Byte 11: Minor version
func (p *Parser) Detect(r io.ReaderAt) bool {
	// Check TIFF header first
	if !p.tiff.Detect(r) {
		return false
	}

	// Read CR2-specific header bytes (local buffer for thread safety)
	var buf [4]byte
	_, err := r.ReadAt(buf[:], 8)
	if err != nil {
		return false
	}

	// Check CR2 magic bytes "CR" at offset 8-9
	if buf[0] != cr2MagicByte1 || buf[1] != cr2MagicByte2 {
		return false
	}

	// Check major version is 0x02 (CR2)
	if buf[2] != cr2MajorVersion {
		return false
	}

	return true
}

// Parse extracts metadata from a CR2 file.
// Delegates to TIFF parser since CR2 is TIFF-based.
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	return p.tiff.Parse(r)
}
