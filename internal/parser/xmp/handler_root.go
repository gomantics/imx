package xmp

import (
	"encoding/xml"
)

type RootStateHandler struct{}

func (h *RootStateHandler) HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) *ContextFrame {
	// Check for RDF element
	if elem.Name.Space == nsRDF && elem.Name.Local == "RDF" {
		return &ContextFrame{Type: CTX_RDF}
	}

	// Extract attributes from root <x:xmpmeta> element
	// Common attribute: x:xmptk (XMP Toolkit version)
	if elem.Name.Local == "xmpmeta" {
		for _, attr := range elem.Attr {
			// Extract x:xmptk attribute (XMP Toolkit)
			if attr.Name.Local == "xmptk" {
				// Map "x" namespace to XMP-x directory
				attrNS := attr.Name.Space
				if attrNS == "" {
					attrNS = "adobe:ns:meta/" // Default x namespace
				}

				// Store in nodeMap as XMP-x:XMPToolkit
				key := PropertyKey{URI: attrNS, Local: "XMPToolkit"}
				nodeMap[key] = []PropertyValue{
					{
						Kind:   KindSimple,
						Scalar: attr.Value,
					},
				}
			}
		}
	}

	return &ContextFrame{Type: CTX_ROOT}
}

func (h *RootStateHandler) HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) {
}
