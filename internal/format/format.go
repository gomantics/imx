package format

import (
	"bufio"

	"github.com/gomantics/imx/internal/types"
)

// RawBlock is a raw metadata payload extracted from an image format
type RawBlock struct {
	Kind    types.MetaKind
	Payload []byte
	Origin  string       // e.g. "APP1 Exif", "eXIf chunk"
	Format  types.Format
	Index   int          // sequence number for multiple blocks of same type
}

// Parser is the interface for format parsers
type Parser interface {
	// Detect returns true if this parser supports the given initial bytes
	Detect(peek []byte) bool

	// Parse reads from r and returns all metadata blocks found
	Parse(r *bufio.Reader) ([]RawBlock, error)
}
