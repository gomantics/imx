package xmp

import (
	"encoding/xml"
)

// RootStateHandler handles the ROOT context.
// This is the initial state before encountering any XMP structure.
type RootStateHandler struct{}

// HandleStart processes start elements in ROOT context.
// Transitions to RDF context if rdf:RDF element is encountered.
func (h *RootStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) (*ContextFrame, error) {
	if elem.Name.Space == nsRDF && elem.Name.Local == "RDF" {
		return &ContextFrame{Type: CTX_RDF}, nil
	}
	// Stay in ROOT context for other elements
	return &ContextFrame{Type: CTX_ROOT}, nil
}

// HandleEnd is a no-op for ROOT context.
// ROOT context doesn't produce any output.
func (h *RootStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) error {
	// No-op: ROOT context doesn't store anything
	return nil
}
