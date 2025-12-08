package xmp

import (
	"encoding/xml"
)

// StructFieldStateHandler handles the STRUCT_FIELD context.
// This state is active when processing a field within a struct (nested property).
type StructFieldStateHandler struct{}

// HandleStart processes start elements in STRUCT_FIELD context.
// Handles nested arrays, structs, and additional struct fields.
func (h *StructFieldStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) (*ContextFrame, error) {
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

	// Otherwise, it's a nested struct field
	parent.propKind = KindStruct
	return createStructFieldContext(space, local, ns, elem.Attr, namespaces), nil
}

// HandleEnd finalizes the struct field and adds it to the parent struct.
func (h *StructFieldStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) error {
	// If this field has a name, create a StructField
	if curr.propLocal != "" {
		val := finalizeValue(curr)
		field := StructField{
			Prefix: curr.propPrefix,
			URI:    curr.propURI,
			Name:   curr.propLocal,
			Value:  val,
		}

		// Add field to parent if parent can contain fields
		if parent.Type == CTX_PROPERTY || parent.Type == CTX_LI || parent.Type == CTX_STRUCT_FIELD {
			parent.fields = append(parent.fields, field)
		}
	} else {
		// Anonymous struct (from rdf:Description) - merge fields into parent
		if parent.Type == CTX_PROPERTY || parent.Type == CTX_LI || parent.Type == CTX_STRUCT_FIELD {
			parent.fields = append(parent.fields, curr.fields...)
		}
	}
	return nil
}
