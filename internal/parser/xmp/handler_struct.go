package xmp

import (
	"encoding/xml"
)

type StructFieldStateHandler struct{}

func (h *StructFieldStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame {
	space := elem.Name.Space
	local := elem.Name.Local

	if isArrayContainer(space, local) {
		parent.propKind = KindArray
		return &ContextFrame{Type: CTX_ARRAY}
	}

	if isRDFDescription(space, local) {
		parent.propKind = KindStruct
		fields := parsePropertyAttrs(elem.Attr, ns, namespaces)
		parent.fields = append(parent.fields, fields...)
		return &ContextFrame{Type: CTX_STRUCT_FIELD, propKind: KindStruct}
	}

	parent.propKind = KindStruct
	return createStructFieldContext(space, local, ns, elem.Attr, namespaces)
}

func (h *StructFieldStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) {
	if curr.propLocal != "" {
		val := finalizeValue(curr)
		field := StructField{
			Prefix: curr.propPrefix,
			URI:    curr.propURI,
			Name:   curr.propLocal,
			Value:  val,
		}

		if parent.Type == CTX_PROPERTY || parent.Type == CTX_LI || parent.Type == CTX_STRUCT_FIELD {
			parent.fields = append(parent.fields, field)
		}
	} else {
		if parent.Type == CTX_PROPERTY || parent.Type == CTX_LI || parent.Type == CTX_STRUCT_FIELD {
			parent.fields = append(parent.fields, curr.fields...)
		}
	}
}
