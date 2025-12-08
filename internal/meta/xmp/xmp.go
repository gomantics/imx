package xmp

import (
	"fmt"

	"github.com/gomantics/imx/internal/common"
)

// Parser implements meta.Parser for the XMP specification using a streaming approach.
type Parser struct{}

// New creates a new XMP parser.
func New() *Parser {
	return &Parser{}
}

// Spec returns the meta.Spec constant for XMP.
func (p *Parser) Spec() common.Spec {
	return common.SpecXMP
}

// Parse parses a list of raw blocks and returns a single Directory containing usage XMP tags.
func (p *Parser) Parse(blocks []common.RawBlock) ([]common.Directory, error) {
	// NodeMap to accumulate properties from all blocks
	nodeMap := make(NodeMap)
	// Track URI -> Prefix mappings found in packets
	namespaces := make(map[string]string)

	foundAny := false
	var lastErr error
	var lastBlockIdx int

	for idx, block := range blocks {
		if block.Spec != common.SpecXMP {
			continue
		}

		payload := stripXPacket(block.Payload)
		if len(payload) == 0 {
			continue
		}

		if err := parsePacket(payload, nodeMap, namespaces); err != nil {
			lastErr = err
			lastBlockIdx = idx
			continue // Skip malformed, try next
		}
		foundAny = true
	}

	if !foundAny && lastErr != nil {
		// If we tried parsing and failed everything
		return nil, fmt.Errorf("parse XMP block %d (size=%d bytes): %w",
			lastBlockIdx, len(blocks[lastBlockIdx].Payload), lastErr)
	}

	if len(nodeMap) == 0 {
		return nil, nil
	}

	dir := flattenNodeMap(nodeMap, namespaces)
	return []common.Directory{dir}, nil
}
