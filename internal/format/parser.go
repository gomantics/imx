package format

import (
	"bufio"

	"github.com/gomantics/imx/internal/common"
)

// Parser is the interface for format parsers
type Parser interface {
	// Detect returns true if this parser supports the given initial bytes
	Detect(peek []byte) bool

	// Parse reads from r and returns all metadata blocks found
	Parse(r *bufio.Reader) ([]common.RawBlock, error)
}
