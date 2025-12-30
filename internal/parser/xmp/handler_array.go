package xmp

import (
	"encoding/xml"
)

type ArrayStateHandler struct{}

func (h *ArrayStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame {
	if isRDFLi(elem.Name.Space, elem.Name.Local) {
		ctx := &ContextFrame{Type: CTX_LI}

		fields := parsePropertyAttrs(elem.Attr, ns, namespaces)
		if len(fields) > 0 {
			ctx.propKind = KindStruct
			ctx.fields = fields
		}

		return ctx
	}

	return &ContextFrame{Type: CTX_ROOT}
}

func (h *ArrayStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) {
	if parent.Type == CTX_PROPERTY {
		parent.items = append(parent.items, curr.items...)
	}
}
