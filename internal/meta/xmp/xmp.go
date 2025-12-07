package xmp

import (
	"github.com/gomantics/imx/internal/format"
	"github.com/gomantics/imx/internal/meta"
)

// Parser implements meta.Parser for the XMP specification using a streaming approach.
type Parser struct{}

// New creates a new XMP parser.
func New() *Parser {
	return &Parser{}
}

// Spec returns the meta.Spec constant for XMP.
func (p *Parser) Spec() meta.Spec {
	return meta.SpecXMP
}

// Parse parses a list of raw blocks and returns a single Directory containing usage XMP tags.
func (p *Parser) Parse(blocks []format.RawBlock) ([]meta.Directory, error) {
	// NodeMap to accumulate properties from all blocks
	nodeMap := make(NodeMap)
	// Track URI -> Prefix mappings found in packets
	namespaces := make(map[string]string)

	foundAny := false
	var lastErr error

	for _, block := range blocks {
		if meta.Spec(block.Spec) != meta.SpecXMP {
			continue
		}

		payload := stripXPacket(block.Payload)
		if len(payload) == 0 {
			continue
		}

		if err := parsePacket(payload, nodeMap, namespaces); err != nil {
			lastErr = err
			continue // Skip malformed, try next
		}
		foundAny = true
	}

	if !foundAny && lastErr != nil {
		// If we tried parsing and failed everything
		return nil, lastErr
	}

	if len(nodeMap) == 0 {
		return nil, nil
	}

	dir := flattenNodeMap(nodeMap, namespaces)
	return []meta.Directory{dir}, nil
}
