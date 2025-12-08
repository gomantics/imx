package xmp

import (
	"encoding/xml"
)

// PropertyStateHandler handles the PROPERTY context.
// This state is active when processing an XMP property element.
type PropertyStateHandler struct{}

// HandleStart processes start elements in PROPERTY context.
// Handles arrays (rdf:Bag/Seq/Alt), structs (rdf:Description), and nested struct fields.
func (h *PropertyStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) (*ContextFrame, error) {
	space := elem.Name.Space
	local := elem.Name.Local

	// Check for array containers
	if isArrayContainer(space, local) {
		parent.propKind = KindArray
		return &ContextFrame{Type: CTX_ARRAY}, nil
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

// HandleEnd finalizes the property and stores it in the node map.
func (h *PropertyStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) error {
	val := finalizeValue(curr)
	key := PropertyKey{curr.propURI, curr.propLocal}
	nodeMap[key] = append(nodeMap[key], val)
	return nil
}
