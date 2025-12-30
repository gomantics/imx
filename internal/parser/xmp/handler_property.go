package xmp

import (
	"encoding/xml"
)

type PropertyStateHandler struct{}

func (h *PropertyStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame {
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

func (h *PropertyStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) {
	val := finalizeValue(curr)
	key := PropertyKey{curr.propURI, curr.propLocal}
	nodeMap[key] = append(nodeMap[key], val)
}
