package xmp

import (
	"encoding/xml"
)

// ArrayStateHandler handles the ARRAY context.
// This state is active when inside an rdf:Bag, rdf:Seq, or rdf:Alt element.
type ArrayStateHandler struct{}

// HandleStart processes start elements in ARRAY context.
// Creates LI contexts for rdf:li elements.
func (h *ArrayStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) (*ContextFrame, error) {
	if isRDFLi(elem.Name.Space, elem.Name.Local) {
		ctx := &ContextFrame{Type: CTX_LI}

		// Check for struct attributes
		fields := parsePropertyAttrs(elem.Attr, ns, namespaces)
		if len(fields) > 0 {
			ctx.propKind = KindStruct
			ctx.fields = fields
		}

		return ctx, nil
	}

	// Fall back to ROOT for unexpected elements
	return &ContextFrame{Type: CTX_ROOT}, nil
}

// HandleEnd transfers array items to the parent property.
func (h *ArrayStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) error {
	// Transfer items from array to parent property
	if parent.Type == CTX_PROPERTY {
		parent.items = append(parent.items, curr.items...)
	}
	return nil
}
