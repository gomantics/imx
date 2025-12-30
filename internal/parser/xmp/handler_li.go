package xmp

import (
	"encoding/xml"
)

type LiStateHandler struct{}

func (h *LiStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame {
	space := elem.Name.Space
	local := elem.Name.Local

	if isArrayContainer(space, local) {
		return &ContextFrame{Type: CTX_ROOT}
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

func (h *LiStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) {
	val := finalizeValue(curr)
	if parent.Type == CTX_ARRAY {
		parent.items = append(parent.items, val)
	}
}
