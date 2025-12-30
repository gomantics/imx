package xmp

import (
	"encoding/xml"
)

type RDFStateHandler struct{}

func (h *RDFStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame {
	if isRDFDescription(elem.Name.Space, elem.Name.Local) {
		parseDescriptionAttrs(elem.Attr, ns, nodeMap, namespaces)
		return &ContextFrame{Type: CTX_DESCRIPTION}
	}
	return &ContextFrame{Type: CTX_ROOT}
}

func (h *RDFStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) {
}

func parseDescriptionAttrs(attrs []xml.Attr, ns *NSFrame, nodeMap NodeMap, namespaces map[string]string) {
	for _, attr := range attrs {
		if isPropAttr(attr.Name) {
			prefix := resolvePrefix(attr.Name.Space, ns)
			namespaces[attr.Name.Space] = prefix
			key := PropertyKey{attr.Name.Space, attr.Name.Local}
			val := PropertyValue{Kind: KindSimple, Scalar: attr.Value}
			nodeMap[key] = append(nodeMap[key], val)
		}
	}
}
