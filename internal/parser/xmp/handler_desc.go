package xmp

import (
	"encoding/xml"
)

type DescriptionStateHandler struct{}

func (h *DescriptionStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame {
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

		fields := parsePropertyAttrs(elem.Attr, ns, namespaces)
		if len(fields) > 0 {
			ctx.propKind = KindStruct
			ctx.fields = fields
		}

		return ctx
	}

	return &ContextFrame{Type: CTX_ROOT}
}

func (h *DescriptionStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) {
}
