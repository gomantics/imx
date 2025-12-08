package xmp

import (
	"encoding/xml"
)

// RDFStateHandler handles the RDF context.
// This state is active when inside an rdf:RDF element.
type RDFStateHandler struct{}

// HandleStart processes start elements in RDF context.
// Transitions to DESCRIPTION context for rdf:Description elements.
func (h *RDFStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame {
	if isRDFDescription(elem.Name.Space, elem.Name.Local) {
		// Parse Description attributes as top-level properties
		parseDescriptionAttrs(elem.Attr, ns, nodeMap, namespaces)
		return &ContextFrame{Type: CTX_DESCRIPTION}
	}
	// Fall back to ROOT for unexpected elements
	return &ContextFrame{Type: CTX_ROOT}
}

// HandleEnd is a no-op for RDF context.
// RDF context doesn't produce any output.
func (h *RDFStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) {
	// No-op: RDF context doesn't store anything
}

// parseDescriptionAttrs extracts XMP properties from rdf:Description attributes.
// In XMP, Description element attributes represent top-level properties in shorthand notation.
// Only property attributes (non-xmlns, non-rdf) are processed and added to the nodeMap.
func parseDescriptionAttrs(attrs []xml.Attr, ns *NSFrame, nodeMap NodeMap, namespaces map[string]string) {
	for _, attr := range attrs {
		if isPropAttr(attr.Name) {
			prefix := resolvePrefix(attr.Name.Space, ns)
			namespaces[attr.Name.Space] = prefix // Capture namespace mapping
			key := PropertyKey{attr.Name.Space, attr.Name.Local}
			val := PropertyValue{Kind: KindSimple, Scalar: attr.Value}
			nodeMap[key] = append(nodeMap[key], val)
		}
	}
}
