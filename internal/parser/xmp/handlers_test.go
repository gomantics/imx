package xmp

import (
	"encoding/xml"
	"testing"
)

func makeNSFrame() *NSFrame {
	return &NSFrame{
		prefixToURI: map[string]string{
			"dc":  "http://purl.org/dc/elements/1.1/",
			"xmp": "http://ns.adobe.com/xap/1.0/",
			"rdf": nsRDF,
		},
		uriToPrefix: map[string]string{
			"http://purl.org/dc/elements/1.1/": "dc",
			"http://ns.adobe.com/xap/1.0/":     "xmp",
			nsRDF:                              "rdf",
		},
	}
}

// --- RootStateHandler Tests ---

func TestRootStateHandler_HandleStart(t *testing.T) {
	tests := []struct {
		name    string
		elem    xml.StartElement
		wantCtx ContextType
	}{
		{
			name: "rdf:RDF element",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "RDF"},
			},
			wantCtx: CTX_RDF,
		},
		{
			name: "other element",
			elem: xml.StartElement{
				Name: xml.Name{Space: "http://other.ns/", Local: "element"},
			},
			wantCtx: CTX_ROOT,
		},
		{
			name: "x:xmpmeta element",
			elem: xml.StartElement{
				Name: xml.Name{Space: "adobe:ns:meta/", Local: "xmpmeta"},
			},
			wantCtx: CTX_ROOT,
		},
		{
			name: "x:xmpmeta element with xmptk attribute",
			elem: xml.StartElement{
				Name: xml.Name{Space: "adobe:ns:meta/", Local: "xmpmeta"},
				Attr: []xml.Attr{
					{Name: xml.Name{Space: "adobe:ns:meta/", Local: "xmptk"}, Value: "Adobe XMP Core 5.6-c140"},
				},
			},
			wantCtx: CTX_ROOT,
		},
		{
			name: "x:xmpmeta element with xmptk attribute no namespace",
			elem: xml.StartElement{
				Name: xml.Name{Space: "adobe:ns:meta/", Local: "xmpmeta"},
				Attr: []xml.Attr{
					{Name: xml.Name{Space: "", Local: "xmptk"}, Value: "Test XMP Toolkit"},
				},
			},
			wantCtx: CTX_ROOT,
		},
	}

	h := &RootStateHandler{}
	ns := makeNSFrame()
	namespaces := make(map[string]string)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeMap := make(NodeMap)
			parent := &ContextFrame{Type: CTX_ROOT}
			ctx := h.HandleStart(tt.elem, parent, ns, namespaces, nodeMap)
			if ctx.Type != tt.wantCtx {
				t.Errorf("HandleStart() Type = %v, want %v", ctx.Type, tt.wantCtx)
			}

			// If this is an xmptk test, verify that the attribute was extracted
			if tt.name == "x:xmpmeta element with xmptk attribute" {
				key := PropertyKey{URI: "adobe:ns:meta/", Local: "XMPToolkit"}
				if vals, ok := nodeMap[key]; !ok {
					t.Error("Expected XMPToolkit to be in nodeMap")
				} else if len(vals) != 1 {
					t.Errorf("Expected 1 XMPToolkit value, got %d", len(vals))
				} else if vals[0].Scalar != "Adobe XMP Core 5.6-c140" {
					t.Errorf("Expected XMPToolkit value = 'Adobe XMP Core 5.6-c140', got '%s'", vals[0].Scalar)
				}
			} else if tt.name == "x:xmpmeta element with xmptk attribute no namespace" {
				// When namespace is empty, it defaults to "adobe:ns:meta/"
				key := PropertyKey{URI: "adobe:ns:meta/", Local: "XMPToolkit"}
				if vals, ok := nodeMap[key]; !ok {
					t.Error("Expected XMPToolkit to be in nodeMap")
				} else if len(vals) != 1 {
					t.Errorf("Expected 1 XMPToolkit value, got %d", len(vals))
				} else if vals[0].Scalar != "Test XMP Toolkit" {
					t.Errorf("Expected XMPToolkit value = 'Test XMP Toolkit', got '%s'", vals[0].Scalar)
				}
			}
		})
	}
}

func TestRootStateHandler_HandleEnd(t *testing.T) {
	h := &RootStateHandler{}
	// HandleEnd is a no-op, just ensure it doesn't panic
	h.HandleEnd(&ContextFrame{Type: CTX_ROOT}, &ContextFrame{Type: CTX_ROOT}, nil)
}

// --- RDFStateHandler Tests ---

func TestRDFStateHandler_HandleStart(t *testing.T) {
	tests := []struct {
		name    string
		elem    xml.StartElement
		wantCtx ContextType
	}{
		{
			name: "rdf:Description element",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Description"},
			},
			wantCtx: CTX_DESCRIPTION,
		},
		{
			name: "rdf:Description with attrs",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Description"},
				Attr: []xml.Attr{
					{Name: xml.Name{Space: "http://purl.org/dc/elements/1.1/", Local: "title"}, Value: "Test"},
				},
			},
			wantCtx: CTX_DESCRIPTION,
		},
		{
			name: "other element",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Bag"},
			},
			wantCtx: CTX_ROOT,
		},
	}

	h := &RDFStateHandler{}
	ns := makeNSFrame()
	namespaces := make(map[string]string)
	nodeMap := make(NodeMap)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &ContextFrame{Type: CTX_RDF}
			ctx := h.HandleStart(tt.elem, parent, ns, namespaces, nodeMap)
			if ctx.Type != tt.wantCtx {
				t.Errorf("HandleStart() Type = %v, want %v", ctx.Type, tt.wantCtx)
			}
		})
	}
}

func TestRDFStateHandler_HandleEnd(t *testing.T) {
	h := &RDFStateHandler{}
	// HandleEnd is a no-op
	h.HandleEnd(&ContextFrame{Type: CTX_RDF}, &ContextFrame{Type: CTX_ROOT}, nil)
}

func TestParseDescriptionAttrs(t *testing.T) {
	ns := makeNSFrame()
	nodeMap := make(NodeMap)
	namespaces := make(map[string]string)

	attrs := []xml.Attr{
		{Name: xml.Name{Space: "http://purl.org/dc/elements/1.1/", Local: "title"}, Value: "Test Title"},
		{Name: xml.Name{Space: "xmlns", Local: "dc"}, Value: "http://purl.org/dc/elements/1.1/"}, // Should be filtered
		{Name: xml.Name{Space: nsRDF, Local: "about"}, Value: ""},                                // Should be filtered
	}

	parseDescriptionAttrs(attrs, ns, nodeMap, namespaces)

	key := PropertyKey{URI: "http://purl.org/dc/elements/1.1/", Local: "title"}
	if vals, ok := nodeMap[key]; !ok || len(vals) != 1 {
		t.Errorf("expected 1 value for dc:title, got %d", len(vals))
	}
}

// --- DescriptionStateHandler Tests ---

func TestDescriptionStateHandler_HandleStart(t *testing.T) {
	tests := []struct {
		name    string
		elem    xml.StartElement
		wantCtx ContextType
	}{
		{
			name: "property element",
			elem: xml.StartElement{
				Name: xml.Name{Space: "http://purl.org/dc/elements/1.1/", Local: "title"},
			},
			wantCtx: CTX_PROPERTY,
		},
		{
			name: "property with attrs",
			elem: xml.StartElement{
				Name: xml.Name{Space: "http://purl.org/dc/elements/1.1/", Local: "creator"},
				Attr: []xml.Attr{
					{Name: xml.Name{Space: "http://purl.org/dc/elements/1.1/", Local: "format"}, Value: "text"},
				},
			},
			wantCtx: CTX_PROPERTY,
		},
		{
			name: "rdf element",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Bag"},
			},
			wantCtx: CTX_ROOT,
		},
	}

	h := &DescriptionStateHandler{}
	ns := makeNSFrame()
	namespaces := make(map[string]string)
	nodeMap := make(NodeMap)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &ContextFrame{Type: CTX_DESCRIPTION}
			ctx := h.HandleStart(tt.elem, parent, ns, namespaces, nodeMap)
			if ctx.Type != tt.wantCtx {
				t.Errorf("HandleStart() Type = %v, want %v", ctx.Type, tt.wantCtx)
			}
		})
	}
}

func TestDescriptionStateHandler_HandleEnd(t *testing.T) {
	h := &DescriptionStateHandler{}
	// HandleEnd is a no-op
	h.HandleEnd(&ContextFrame{Type: CTX_DESCRIPTION}, &ContextFrame{Type: CTX_RDF}, nil)
}

// --- PropertyStateHandler Tests ---

func TestPropertyStateHandler_HandleStart(t *testing.T) {
	tests := []struct {
		name       string
		elem       xml.StartElement
		wantCtx    ContextType
		wantParent PropKind
	}{
		{
			name: "array container - Bag",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Bag"},
			},
			wantCtx:    CTX_ARRAY,
			wantParent: KindArray,
		},
		{
			name: "array container - Seq",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Seq"},
			},
			wantCtx:    CTX_ARRAY,
			wantParent: KindArray,
		},
		{
			name: "array container - Alt",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Alt"},
			},
			wantCtx:    CTX_ARRAY,
			wantParent: KindArray,
		},
		{
			name: "rdf:Description",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Description"},
			},
			wantCtx:    CTX_STRUCT_FIELD,
			wantParent: KindStruct,
		},
		{
			name: "struct field element",
			elem: xml.StartElement{
				Name: xml.Name{Space: "http://ns.adobe.com/xap/1.0/", Local: "field"},
			},
			wantCtx:    CTX_STRUCT_FIELD,
			wantParent: KindStruct,
		},
	}

	h := &PropertyStateHandler{}
	ns := makeNSFrame()
	namespaces := make(map[string]string)
	nodeMap := make(NodeMap)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &ContextFrame{Type: CTX_PROPERTY, propKind: KindUnknown}
			ctx := h.HandleStart(tt.elem, parent, ns, namespaces, nodeMap)
			if ctx.Type != tt.wantCtx {
				t.Errorf("HandleStart() Type = %v, want %v", ctx.Type, tt.wantCtx)
			}
			if parent.propKind != tt.wantParent {
				t.Errorf("parent.propKind = %v, want %v", parent.propKind, tt.wantParent)
			}
		})
	}
}

func TestPropertyStateHandler_HandleEnd(t *testing.T) {
	h := &PropertyStateHandler{}
	nodeMap := make(NodeMap)

	curr := &ContextFrame{
		Type:      CTX_PROPERTY,
		propURI:   "http://purl.org/dc/elements/1.1/",
		propLocal: "title",
	}
	curr.text.WriteString("Test Value")

	parent := &ContextFrame{Type: CTX_DESCRIPTION}

	h.HandleEnd(curr, parent, nodeMap)

	key := PropertyKey{URI: "http://purl.org/dc/elements/1.1/", Local: "title"}
	if vals, ok := nodeMap[key]; !ok || len(vals) != 1 {
		t.Errorf("expected value in nodeMap")
	}
}

// --- ArrayStateHandler Tests ---

func TestArrayStateHandler_HandleStart(t *testing.T) {
	tests := []struct {
		name    string
		elem    xml.StartElement
		wantCtx ContextType
	}{
		{
			name: "rdf:li element",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "li"},
			},
			wantCtx: CTX_LI,
		},
		{
			name: "rdf:li with attrs",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "li"},
				Attr: []xml.Attr{
					{Name: xml.Name{Space: "http://example.com/", Local: "attr"}, Value: "val"},
				},
			},
			wantCtx: CTX_LI,
		},
		{
			name: "other element",
			elem: xml.StartElement{
				Name: xml.Name{Space: "http://example.com/", Local: "other"},
			},
			wantCtx: CTX_ROOT,
		},
	}

	h := &ArrayStateHandler{}
	ns := makeNSFrame()
	namespaces := make(map[string]string)
	nodeMap := make(NodeMap)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &ContextFrame{Type: CTX_ARRAY}
			ctx := h.HandleStart(tt.elem, parent, ns, namespaces, nodeMap)
			if ctx.Type != tt.wantCtx {
				t.Errorf("HandleStart() Type = %v, want %v", ctx.Type, tt.wantCtx)
			}
		})
	}
}

func TestArrayStateHandler_HandleEnd(t *testing.T) {
	h := &ArrayStateHandler{}

	t.Run("append items to property parent", func(t *testing.T) {
		curr := &ContextFrame{
			Type:  CTX_ARRAY,
			items: []PropertyValue{{Kind: KindSimple, Scalar: "item1"}},
		}
		parent := &ContextFrame{Type: CTX_PROPERTY}

		h.HandleEnd(curr, parent, nil)

		if len(parent.items) != 1 {
			t.Errorf("parent.items = %d, want 1", len(parent.items))
		}
	})

	t.Run("non-property parent", func(t *testing.T) {
		curr := &ContextFrame{
			Type:  CTX_ARRAY,
			items: []PropertyValue{{Kind: KindSimple, Scalar: "item1"}},
		}
		parent := &ContextFrame{Type: CTX_ROOT}

		h.HandleEnd(curr, parent, nil)

		if len(parent.items) != 0 {
			t.Errorf("parent.items = %d, want 0", len(parent.items))
		}
	})
}

// --- LiStateHandler Tests ---

func TestLiStateHandler_HandleStart(t *testing.T) {
	tests := []struct {
		name       string
		elem       xml.StartElement
		wantCtx    ContextType
		wantParent PropKind
	}{
		{
			name: "array container",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Bag"},
			},
			wantCtx:    CTX_ROOT,
			wantParent: KindUnknown,
		},
		{
			name: "rdf:Description",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Description"},
			},
			wantCtx:    CTX_STRUCT_FIELD,
			wantParent: KindStruct,
		},
		{
			name: "struct field",
			elem: xml.StartElement{
				Name: xml.Name{Space: "http://ns.adobe.com/xap/1.0/", Local: "field"},
			},
			wantCtx:    CTX_STRUCT_FIELD,
			wantParent: KindStruct,
		},
	}

	h := &LiStateHandler{}
	ns := makeNSFrame()
	namespaces := make(map[string]string)
	nodeMap := make(NodeMap)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &ContextFrame{Type: CTX_LI, propKind: KindUnknown}
			ctx := h.HandleStart(tt.elem, parent, ns, namespaces, nodeMap)
			if ctx.Type != tt.wantCtx {
				t.Errorf("HandleStart() Type = %v, want %v", ctx.Type, tt.wantCtx)
			}
			if parent.propKind != tt.wantParent {
				t.Errorf("parent.propKind = %v, want %v", parent.propKind, tt.wantParent)
			}
		})
	}
}

func TestLiStateHandler_HandleEnd(t *testing.T) {
	h := &LiStateHandler{}

	t.Run("append to array parent", func(t *testing.T) {
		curr := &ContextFrame{Type: CTX_LI}
		curr.text.WriteString("item value")
		parent := &ContextFrame{Type: CTX_ARRAY}

		h.HandleEnd(curr, parent, nil)

		if len(parent.items) != 1 {
			t.Errorf("parent.items = %d, want 1", len(parent.items))
		}
	})

	t.Run("non-array parent", func(t *testing.T) {
		curr := &ContextFrame{Type: CTX_LI}
		curr.text.WriteString("item value")
		parent := &ContextFrame{Type: CTX_ROOT}

		h.HandleEnd(curr, parent, nil)

		if len(parent.items) != 0 {
			t.Errorf("parent.items = %d, want 0", len(parent.items))
		}
	})
}

// --- StructFieldStateHandler Tests ---

func TestStructFieldStateHandler_HandleStart(t *testing.T) {
	tests := []struct {
		name       string
		elem       xml.StartElement
		wantCtx    ContextType
		wantParent PropKind
	}{
		{
			name: "array container",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Seq"},
			},
			wantCtx:    CTX_ARRAY,
			wantParent: KindArray,
		},
		{
			name: "rdf:Description",
			elem: xml.StartElement{
				Name: xml.Name{Space: nsRDF, Local: "Description"},
			},
			wantCtx:    CTX_STRUCT_FIELD,
			wantParent: KindStruct,
		},
		{
			name: "nested struct field",
			elem: xml.StartElement{
				Name: xml.Name{Space: "http://ns.adobe.com/xap/1.0/", Local: "nested"},
			},
			wantCtx:    CTX_STRUCT_FIELD,
			wantParent: KindStruct,
		},
	}

	h := &StructFieldStateHandler{}
	ns := makeNSFrame()
	namespaces := make(map[string]string)
	nodeMap := make(NodeMap)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &ContextFrame{Type: CTX_STRUCT_FIELD, propKind: KindUnknown}
			ctx := h.HandleStart(tt.elem, parent, ns, namespaces, nodeMap)
			if ctx.Type != tt.wantCtx {
				t.Errorf("HandleStart() Type = %v, want %v", ctx.Type, tt.wantCtx)
			}
			if parent.propKind != tt.wantParent {
				t.Errorf("parent.propKind = %v, want %v", parent.propKind, tt.wantParent)
			}
		})
	}
}

func TestStructFieldStateHandler_HandleEnd(t *testing.T) {
	h := &StructFieldStateHandler{}

	t.Run("with propLocal - add field to parent", func(t *testing.T) {
		curr := &ContextFrame{
			Type:       CTX_STRUCT_FIELD,
			propURI:    "http://ns.adobe.com/xap/1.0/",
			propLocal:  "field",
			propPrefix: "xmp",
		}
		curr.text.WriteString("value")
		parent := &ContextFrame{Type: CTX_PROPERTY}

		h.HandleEnd(curr, parent, nil)

		if len(parent.fields) != 1 {
			t.Errorf("parent.fields = %d, want 1", len(parent.fields))
		}
		if parent.fields[0].Name != "field" {
			t.Errorf("field Name = %q, want 'field'", parent.fields[0].Name)
		}
	})

	t.Run("without propLocal - merge fields", func(t *testing.T) {
		curr := &ContextFrame{
			Type:   CTX_STRUCT_FIELD,
			fields: []StructField{{Name: "merged"}},
		}
		parent := &ContextFrame{Type: CTX_LI}

		h.HandleEnd(curr, parent, nil)

		if len(parent.fields) != 1 {
			t.Errorf("parent.fields = %d, want 1", len(parent.fields))
		}
	})

	t.Run("add to STRUCT_FIELD parent", func(t *testing.T) {
		curr := &ContextFrame{
			Type:       CTX_STRUCT_FIELD,
			propURI:    "http://ns.adobe.com/xap/1.0/",
			propLocal:  "nested",
			propPrefix: "xmp",
		}
		parent := &ContextFrame{Type: CTX_STRUCT_FIELD}

		h.HandleEnd(curr, parent, nil)

		if len(parent.fields) != 1 {
			t.Errorf("parent.fields = %d, want 1", len(parent.fields))
		}
	})

	t.Run("non-matching parent type", func(t *testing.T) {
		curr := &ContextFrame{
			Type:       CTX_STRUCT_FIELD,
			propURI:    "http://ns.adobe.com/xap/1.0/",
			propLocal:  "field",
			propPrefix: "xmp",
		}
		parent := &ContextFrame{Type: CTX_ROOT}

		h.HandleEnd(curr, parent, nil)

		// Should not add to non-matching parent
		if len(parent.fields) != 0 {
			t.Errorf("parent.fields = %d, want 0", len(parent.fields))
		}
	})
}

// --- Test Handler Interface Implementation ---

func TestHandlerInterfaceImplementations(t *testing.T) {
	var _ StateHandler = (*RootStateHandler)(nil)
	var _ StateHandler = (*RDFStateHandler)(nil)
	var _ StateHandler = (*DescriptionStateHandler)(nil)
	var _ StateHandler = (*PropertyStateHandler)(nil)
	var _ StateHandler = (*ArrayStateHandler)(nil)
	var _ StateHandler = (*LiStateHandler)(nil)
	var _ StateHandler = (*StructFieldStateHandler)(nil)
}
