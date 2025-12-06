package meta

import (
	"github.com/gomantics/imx/internal/format"
)

// Parser is the interface for metadata parsers
type Parser interface {
	// Spec returns the metadata spec this parser handles
	Spec() Spec

	// Parse consumes relevant RawBlocks and returns Directories
	Parse(blocks []format.RawBlock) ([]Directory, error)
}
