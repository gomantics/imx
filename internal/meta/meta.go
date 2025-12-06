package meta

import (
	"github.com/gomantics/imx/internal/format"
	"github.com/gomantics/imx/internal/types"
)

// Directory is a logical collection of tags for a given metadata spec
type Directory struct {
	Spec types.Spec
	Name string
	Tags map[TagID]Tag
}

// TagID is a unique identifier for a metadata tag
type TagID string

// Tag represents a single metadata attribute
type Tag struct {
	Spec     types.Spec
	ID       TagID
	Name     string
	DataType string
	Value    any
	Raw      []byte
}

// Parser is the interface for metadata parsers
type Parser interface {
	// Spec returns the metadata spec this parser handles
	Spec() types.Spec

	// Parse consumes relevant RawBlocks and returns Directories
	Parse(blocks []format.RawBlock) ([]Directory, error)
}
