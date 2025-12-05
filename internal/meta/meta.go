package meta

import (
	"github.com/gomantics/imx/internal/container"
	"github.com/gomantics/imx/internal/types"
)

// Directory is a logical collection of tags for a given namespace and grouping
type Directory struct {
	Namespace types.Namespace
	Name      string
	Tags      map[TagID]Tag
}

// TagID is a unique identifier for a metadata tag
type TagID string

// Tag represents a single metadata attribute
type Tag struct {
	Namespace types.Namespace
	ID        TagID
	Name      string
	Type      string
	Value     any
	Raw       []byte
}

// Parser is the interface for namespace parsers
type Parser interface {
	// Namespace returns the namespace this parser handles
	Namespace() types.Namespace

	// Parse consumes relevant RawBlocks and returns Directories
	Parse(blocks []container.RawBlock, cfg types.ExtractorConfig) ([]Directory, error)
}

// Registry holds all registered namespace parsers
var registry = &Registry{
	parsers: make(map[types.Namespace]Parser),
}

// Registry manages namespace parsers
type Registry struct {
	parsers map[types.Namespace]Parser
}

// Register adds a namespace parser to the registry
func Register(p Parser) {
	registry.parsers[p.Namespace()] = p
}

// Get returns the parser for the given namespace
func Get(ns types.Namespace) (Parser, bool) {
	p, ok := registry.parsers[ns]
	return p, ok
}

// GetRegistry returns the global registry
func GetRegistry() *Registry {
	return registry
}

// All returns all registered parsers
func (r *Registry) All() map[types.Namespace]Parser {
	return r.parsers
}
