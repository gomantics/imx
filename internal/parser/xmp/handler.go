package xmp

import (
	"encoding/xml"
)

type StateHandler interface {
	HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame
	HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap)
}
