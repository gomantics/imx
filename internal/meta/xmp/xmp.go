package xmp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/common"
)

// Parser implements meta.Parser for the XMP specification using a streaming approach.
type Parser struct {
	handlers *HandlerRegistry
}

// New creates a new XMP parser with an initialized handler registry.
func New() *Parser {
	return &Parser{
		handlers: NewHandlerRegistry(),
	}
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

		if err := p.parsePacket(payload, nodeMap, namespaces); err != nil {
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

// parsePacket parses a single XMP packet using streaming XML parsing.
// It uses a stack-based state machine with namespace tracking to convert
// RDF/XML into a flat property map suitable for the public API.
func (p *Parser) parsePacket(data []byte, nodeMap NodeMap, namespaces map[string]string) error {
	// Validate inputs
	if len(data) == 0 {
		return fmt.Errorf("empty XMP data")
	}
	if nodeMap == nil {
		return fmt.Errorf("nodeMap cannot be nil")
	}
	if namespaces == nil {
		return fmt.Errorf("namespaces map cannot be nil")
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))

	// Initialize stacks
	nsStack := []*NSFrame{replaceNSFrame(nil, nil)} // Global namespace frame
	ctxStack := []*ContextFrame{{Type: CTX_ROOT}}   // Start in ROOT context

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decode XML token: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			// 1. Manage namespace stack
			parentNS := nsStack[len(nsStack)-1]
			currNS := replaceNSFrame(parentNS, t.Attr)
			nsStack = append(nsStack, currNS)

			// 2. Delegate to state handler
			parent := ctxStack[len(ctxStack)-1]
			handler := p.handlers.Get(parent.Type)
			newCtx := handler.HandleStart(t, parent, currNS, namespaces, nodeMap)
			ctxStack = append(ctxStack, newCtx)

		case xml.EndElement:
			// 3. Delegate to state handler
			curr := ctxStack[len(ctxStack)-1]
			parent := ctxStack[len(ctxStack)-2]
			handler := p.handlers.Get(curr.Type)
			handler.HandleEnd(curr, parent, nodeMap)

			// 4. Pop stacks
			ctxStack = ctxStack[:len(ctxStack)-1]
			nsStack = nsStack[:len(nsStack)-1]

		case xml.CharData:
			// 5. Accumulate character data in current context
			top := ctxStack[len(ctxStack)-1]
			top.text.Write(t)
		}
	}

	return nil
}
