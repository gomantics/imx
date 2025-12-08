package xmp

import (
	"encoding/xml"
)

// LiStateHandler handles the LI (list item) context.
// This state is active when inside an rdf:li element within an array.
type LiStateHandler struct{}

// HandleStart processes start elements in LI context.
// Handles nested arrays, structs, and struct fields within list items.
func (h *LiStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) (*ContextFrame, error) {
	space := elem.Name.Space
	local := elem.Name.Local

	// Nested array containers are not supported (fall back to ROOT)
	if isArrayContainer(space, local) {
		return &ContextFrame{Type: CTX_ROOT}, nil
	}

	// Check for struct (rdf:Description)
	if isRDFDescription(space, local) {
		parent.propKind = KindStruct
		fields := parsePropertyAttrs(elem.Attr, ns, namespaces)
		parent.fields = append(parent.fields, fields...)
		return &ContextFrame{Type: CTX_STRUCT_FIELD, propKind: KindStruct}, nil
	}

	// Otherwise, it's a struct field
	parent.propKind = KindStruct
	return createStructFieldContext(space, local, ns, elem.Attr, namespaces), nil
}

// HandleEnd finalizes the list item and adds it to the parent array.
func (h *LiStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) error {
	val := finalizeValue(curr)
	if parent.Type == CTX_ARRAY {
		parent.items = append(parent.items, val)
	}
	return nil
}
