package xmp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const (
	nsRDF = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	nsXML = "http://www.w3.org/XML/1998/namespace"
)

func parsePacket(data []byte, nodeMap NodeMap, namespaces map[string]string) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	// Initialize handler registry (pre-allocated, reusable)
	handlers := NewHandlerRegistry()

	// Initialize stacks
	nsStack := []*NSFrame{replaceNSFrame(nil, nil)} // Global namespace frame
	ctxStack := []*ContextFrame{{Type: CTX_ROOT}}   // Start in ROOT context

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decode XML token: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			// 1. Manage namespace stack
			parentNS := nsStack[len(nsStack)-1]
			currNS := replaceNSFrame(parentNS, t.Attr)
			nsStack = append(nsStack, currNS)

			// 2. Delegate to state handler
			parent := ctxStack[len(ctxStack)-1]
			handler := handlers.Get(parent.Type)
			newCtx, err := handler.HandleStart(t, parent, currNS, namespaces, nodeMap)
			if err != nil {
				return fmt.Errorf("handle %s start element <%s:%s>: %w",
					parent.Type, t.Name.Space, t.Name.Local, err)
			}
			ctxStack = append(ctxStack, newCtx)

		case xml.EndElement:
			// 3. Delegate to state handler
			curr := ctxStack[len(ctxStack)-1]
			parent := ctxStack[len(ctxStack)-2]
			handler := handlers.Get(curr.Type)
			if err := handler.HandleEnd(curr, parent, nodeMap); err != nil {
				return fmt.Errorf("handle %s end element: %w", curr.Type, err)
			}

			// 4. Pop stacks
			ctxStack = ctxStack[:len(ctxStack)-1]
			nsStack = nsStack[:len(nsStack)-1]

		case xml.CharData:
			// 5. Accumulate character data in current context
			top := ctxStack[len(ctxStack)-1]
			top.text.Write(t)
		}
	}

	return nil
}

func finalizeValue(ctx *ContextFrame) PropertyValue {
	// Priority: Array > Struct > Simple
	if ctx.propKind == KindArray || len(ctx.items) > 0 {
		return PropertyValue{Kind: KindArray, Items: ctx.items}
	}
	if ctx.propKind == KindStruct || len(ctx.fields) > 0 {
		return PropertyValue{Kind: KindStruct, Fields: ctx.fields}
	}

	// Simple
	// Check for empty text but "Struct" implied by attributes?
	// Already handled: propKind set to Struct if attrs found.
	txt := strings.TrimSpace(ctx.text.String())
	return PropertyValue{Kind: KindSimple, Scalar: txt}
}

func parseDescriptionAttrs(attrs []xml.Attr, ns *NSFrame, nodeMap NodeMap, namespaces map[string]string) {
	// Treat attributes as top-level properties
	for _, attr := range attrs {
		if isPropAttr(attr.Name) {
			prefix := resolvePrefix(attr.Name.Space, ns)
			namespaces[attr.Name.Space] = prefix // Capture
			key := PropertyKey{attr.Name.Space, attr.Name.Local}
			val := PropertyValue{Kind: KindSimple, Scalar: attr.Value}
			nodeMap[key] = append(nodeMap[key], val)
		}
	}
}

func parsePropertyAttrs(attrs []xml.Attr, ns *NSFrame, namespaces map[string]string) []StructField {
	var fields []StructField
	for _, attr := range attrs {
		if isPropAttr(attr.Name) {
			prefix := resolvePrefix(attr.Name.Space, ns)
			namespaces[attr.Name.Space] = prefix // Capture
			val := PropertyValue{Kind: KindSimple, Scalar: attr.Value}
			fields = append(fields, StructField{
				Prefix: prefix,
				URI:    attr.Name.Space,
				Name:   attr.Name.Local,
				Value:  val,
			})
		}
	}
	return fields
}
