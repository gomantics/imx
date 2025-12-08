package xmp

import (
	"encoding/xml"
)

// RDFStateHandler handles the RDF context.
// This state is active when inside an rdf:RDF element.
type RDFStateHandler struct{}

// HandleStart processes start elements in RDF context.
// Transitions to DESCRIPTION context for rdf:Description elements.
func (h *RDFStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) (*ContextFrame, error) {
	if isRDFDescription(elem.Name.Space, elem.Name.Local) {
		// Parse Description attributes as top-level properties
		parseDescriptionAttrs(elem.Attr, ns, nodeMap, namespaces)
		return &ContextFrame{Type: CTX_DESCRIPTION}, nil
	}
	// Fall back to ROOT for unexpected elements
	return &ContextFrame{Type: CTX_ROOT}, nil
}

// HandleEnd is a no-op for RDF context.
// RDF context doesn't produce any output.
func (h *RDFStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) error {
	// No-op: RDF context doesn't store anything
	return nil
}
