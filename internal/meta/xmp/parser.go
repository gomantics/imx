package xmp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func parsePacket(data []byte, nodeMap NodeMap, namespaces map[string]string) error {
	// Validate inputs
	if len(data) == 0 {
		return fmt.Errorf("empty XMP data")
	}
	if nodeMap == nil {
		return fmt.Errorf("nodeMap cannot be nil")
	}
	if namespaces == nil {
		return fmt.Errorf("namespaces map cannot be nil")
	}

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
			if len(nsStack) == 0 {
				return fmt.Errorf("namespace stack underflow on start element")
			}
			parentNS := nsStack[len(nsStack)-1]
			currNS := replaceNSFrame(parentNS, t.Attr)
			nsStack = append(nsStack, currNS)

			// 2. Delegate to state handler
			if len(ctxStack) == 0 {
				return fmt.Errorf("context stack underflow on start element")
			}
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
			if len(ctxStack) < 2 {
				return fmt.Errorf("context stack underflow on end element (need at least 2, have %d)", len(ctxStack))
			}
			curr := ctxStack[len(ctxStack)-1]
			parent := ctxStack[len(ctxStack)-2]
			handler := handlers.Get(curr.Type)
			if err := handler.HandleEnd(curr, parent, nodeMap); err != nil {
				return fmt.Errorf("handle %s end element: %w", curr.Type, err)
			}

			// 4. Pop stacks
			ctxStack = ctxStack[:len(ctxStack)-1]
			if len(nsStack) == 0 {
				return fmt.Errorf("namespace stack underflow on stack pop")
			}
			nsStack = nsStack[:len(nsStack)-1]

		case xml.CharData:
			// 5. Accumulate character data in current context
			if len(ctxStack) == 0 {
				return fmt.Errorf("context stack underflow on char data")
			}
			top := ctxStack[len(ctxStack)-1]
			top.text.Write(t)
		}
	}

	return nil
}

// finalizeValue converts a ContextFrame into a PropertyValue.
// It determines the value kind based on accumulated data with priority: Array > Struct > Simple.
// Arrays are identified by propKind=Array or non-empty items slice.
// Structs are identified by propKind=Struct or non-empty fields slice.
// Simple values are trimmed text content from the text builder.
func finalizeValue(ctx *ContextFrame) PropertyValue {
	// Priority: Array > Struct > Simple
	if ctx.propKind == KindArray || len(ctx.items) > 0 {
		return PropertyValue{Kind: KindArray, Items: ctx.items}
	}
	if ctx.propKind == KindStruct || len(ctx.fields) > 0 {
		return PropertyValue{Kind: KindStruct, Fields: ctx.fields}
	}

	// Simple value - trim whitespace from accumulated text
	txt := strings.TrimSpace(ctx.text.String())
	return PropertyValue{Kind: KindSimple, Scalar: txt}
}

// parseDescriptionAttrs extracts XMP properties from rdf:Description attributes.
// In XMP, Description element attributes represent top-level properties in shorthand notation.
// Only property attributes (non-xmlns, non-rdf) are processed and added to the nodeMap.
func parseDescriptionAttrs(attrs []xml.Attr, ns *NSFrame, nodeMap NodeMap, namespaces map[string]string) {
	for _, attr := range attrs {
		if isPropAttr(attr.Name) {
			prefix := resolvePrefix(attr.Name.Space, ns)
			namespaces[attr.Name.Space] = prefix // Capture namespace mapping
			key := PropertyKey{attr.Name.Space, attr.Name.Local}
			val := PropertyValue{Kind: KindSimple, Scalar: attr.Value}
			nodeMap[key] = append(nodeMap[key], val)
		}
	}
}

// parsePropertyAttrs extracts struct fields from element attributes.
// In XMP, attributes on property elements represent fields of a struct (shorthand struct notation).
// Returns a slice of StructField representing each property attribute.
func parsePropertyAttrs(attrs []xml.Attr, ns *NSFrame, namespaces map[string]string) []StructField {
	var fields []StructField
	for _, attr := range attrs {
		if isPropAttr(attr.Name) {
			prefix := resolvePrefix(attr.Name.Space, ns)
			namespaces[attr.Name.Space] = prefix // Capture namespace mapping
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
