package container

import (
	"bufio"

	"github.com/gomantics/imx/internal/types"
)

// MetaKind represents the type of metadata block
type MetaKind int

const (
	MetaKindEXIF MetaKind = iota
	MetaKindIPTC
	MetaKindXMP
	MetaKindICC
)

// RawBlock is a raw metadata payload extracted from a container
type RawBlock struct {
	Kind    MetaKind
	Payload []byte
	Origin  string        // e.g. "APP1 Exif", "APP13 IPTC"
	Format  types.Format
	Index   int           // sequence number for multiple blocks of same type
}

// Parser is the interface for container format parsers
type Parser interface {
	// Detect returns true if this parser supports the given initial bytes
	Detect(peek []byte) bool

	// Parse reads from r and returns all metadata blocks found
	Parse(r *bufio.Reader, cfg types.ExtractorConfig) ([]RawBlock, error)
}

// Registry holds all registered container parsers
var registry = &Registry{
	parsers: make([]Parser, 0),
}

// Registry manages container parsers
type Registry struct {
	parsers []Parser
}

// Register adds a container parser to the registry
func Register(p Parser) {
	registry.parsers = append(registry.parsers, p)
}

// Detect finds the first parser that can handle the given data
func Detect(peek []byte) Parser {
	for _, p := range registry.parsers {
		if p.Detect(peek) {
			return p
		}
	}
	return nil
}

// GetRegistry returns the global registry
func GetRegistry() *Registry {
	return registry
}
