package xmp

import (
	"encoding/xml"
)

// DescriptionStateHandler handles the DESCRIPTION context.
// This state is active when inside an rdf:Description element.
type DescriptionStateHandler struct{}

// HandleStart processes start elements in DESCRIPTION context.
// Creates PROPERTY contexts for non-RDF elements (actual XMP properties).
func (h *DescriptionStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame {
	// Only non-RDF elements are properties in Description context
	if elem.Name.Space != nsRDF {
		prefix := resolvePrefix(elem.Name.Space, ns)
		namespaces[elem.Name.Space] = prefix

		ctx := &ContextFrame{
			Type:       CTX_PROPERTY,
			propURI:    elem.Name.Space,
			propLocal:  elem.Name.Local,
			propPrefix: prefix,
			propKind:   KindUnknown,
		}

		// Check for struct attributes (shorthand struct notation)
		fields := parsePropertyAttrs(elem.Attr, ns, namespaces)
		if len(fields) > 0 {
			ctx.propKind = KindStruct
			ctx.fields = fields
		}

		return ctx
	}

	// Fall back to ROOT for RDF elements in Description (unexpected)
	return &ContextFrame{Type: CTX_ROOT}
}

// HandleEnd is a no-op for DESCRIPTION context.
// Description context doesn't produce any output.
func (h *DescriptionStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) {
	// No-op: DESCRIPTION context doesn't store anything
}
