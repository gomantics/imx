package meta

import (
	"github.com/gomantics/imx/internal/container"
	"github.com/gomantics/imx/internal/types"
)

// Namespace is a metadata namespace identifier
type Namespace = types.Namespace

// Directory is a logical collection of tags for a given namespace and grouping
type Directory struct {
	Namespace Namespace
	Name      string
	Tags      map[TagID]Tag
}

// TagID is a unique identifier for a metadata tag
type TagID string

// Tag represents a single metadata attribute
type Tag struct {
	Namespace Namespace
	ID        TagID
	Name      string
	Type      string
	Value     any
	Raw       []byte
}

// Parser is the interface for namespace parsers
type Parser interface {
	// Namespace returns the namespace this parser handles
	Namespace() Namespace

	// Parse consumes relevant RawBlocks and returns Directories
	Parse(blocks []container.RawBlock) ([]Directory, error)
}
