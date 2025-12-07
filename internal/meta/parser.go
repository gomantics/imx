package meta

import (
	"github.com/gomantics/imx/internal/common"
)

// Parser is the interface for metadata parsers
type Parser interface {
	// Spec returns the metadata spec this parser handles
	Spec() common.Spec

	// Parse consumes relevant RawBlocks and returns Directories
	Parse(blocks []common.RawBlock) ([]common.Directory, error)
}
