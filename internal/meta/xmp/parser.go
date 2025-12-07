package xmp

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

const (
	nsRDF = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	nsXML = "http://www.w3.org/XML/1998/namespace"
)

func parsePacket(data []byte, nodeMap NodeMap, namespaces map[string]string) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	// Initial State
	var nsStack []*NSFrame
	nsStack = append(nsStack, replaceNSFrame(nil, nil)) // Global frame, empty

	var ctxStack []*ContextFrame
	ctxStack = append(ctxStack, &ContextFrame{Type: CTX_ROOT})

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// 1. Manage Namespace Stack
			parentNS := nsStack[len(nsStack)-1]
			currNS := replaceNSFrame(parentNS, t.Attr)
			nsStack = append(nsStack, currNS)

			// 2. Resolve Element Name
			space := t.Name.Space
			local := t.Name.Local

			// 3. State Transition
			parentCtx := ctxStack[len(ctxStack)-1]
			var newCtx *ContextFrame

			switch parentCtx.Type {
			case CTX_ROOT:
				if space == nsRDF && local == "RDF" {
					newCtx = &ContextFrame{Type: CTX_RDF}
				} else {
					newCtx = &ContextFrame{Type: CTX_ROOT}
				}

			case CTX_RDF:
				if space == nsRDF && local == "Description" {
					newCtx = &ContextFrame{Type: CTX_DESCRIPTION}
					// Handle attributes of Description immediately
					parseDescriptionAttrs(t.Attr, currNS, nodeMap, namespaces)
				} else {
					newCtx = &ContextFrame{Type: CTX_ROOT}
				}

			case CTX_DESCRIPTION:
				if space != nsRDF { // It's a property
					prefix := resolvePrefix(space, currNS)
					namespaces[space] = prefix // Capture
					newCtx = &ContextFrame{
						Type:       CTX_PROPERTY,
						propURI:    space,
						propLocal:  local,
						propPrefix: prefix,
						propKind:   KindUnknown,
					}
					// Check attributes of property element (shorthand struct)
					fields := parsePropertyAttrs(t.Attr, currNS, namespaces)
					if len(fields) > 0 {
						newCtx.propKind = KindStruct
						newCtx.fields = fields
					}
				} else {
					newCtx = &ContextFrame{Type: CTX_ROOT}
				}

			case CTX_PROPERTY:
				if space == nsRDF && (local == "Bag" || local == "Seq" || local == "Alt") {
					parentCtx.propKind = KindArray
					newCtx = &ContextFrame{Type: CTX_ARRAY}
				} else if space == nsRDF && local == "Description" {
					parentCtx.propKind = KindStruct
					fields := parsePropertyAttrs(t.Attr, currNS, namespaces)
					parentCtx.fields = append(parentCtx.fields, fields...)
					newCtx = &ContextFrame{Type: CTX_STRUCT_FIELD, propKind: KindStruct}

				} else {
					parentCtx.propKind = KindStruct
					prefix := resolvePrefix(space, currNS)
					namespaces[space] = prefix // Capture
					newCtx = &ContextFrame{
						Type:       CTX_STRUCT_FIELD,
						propURI:    space,
						propLocal:  local,
						propPrefix: prefix,
					}
					fields := parsePropertyAttrs(t.Attr, currNS, namespaces)
					if len(fields) > 0 {
						newCtx.propKind = KindStruct
						newCtx.fields = fields
					}
				}

			case CTX_ARRAY:
				if space == nsRDF && local == "li" {
					newCtx = &ContextFrame{Type: CTX_LI}
					fields := parsePropertyAttrs(t.Attr, currNS, namespaces)
					if len(fields) > 0 {
						newCtx.propKind = KindStruct
						newCtx.fields = fields
					}
				} else {
					newCtx = &ContextFrame{Type: CTX_ROOT}
				}

			case CTX_LI:
				if space == nsRDF && (local == "Bag" || local == "Seq" || local == "Alt") {
					newCtx = &ContextFrame{Type: CTX_ROOT}
				} else if space == nsRDF && local == "Description" {
					parentCtx.propKind = KindStruct
					fields := parsePropertyAttrs(t.Attr, currNS, namespaces)
					parentCtx.fields = append(parentCtx.fields, fields...)
					newCtx = &ContextFrame{Type: CTX_STRUCT_FIELD, propKind: KindStruct}
				} else {
					parentCtx.propKind = KindStruct
					prefix := resolvePrefix(space, currNS)
					namespaces[space] = prefix // Capture
					newCtx = &ContextFrame{
						Type:       CTX_STRUCT_FIELD,
						propURI:    space,
						propLocal:  local,
						propPrefix: prefix,
					}
					fields := parsePropertyAttrs(t.Attr, currNS, namespaces)
					if len(fields) > 0 {
						newCtx.propKind = KindStruct
						newCtx.fields = fields
					}
				}

			case CTX_STRUCT_FIELD:
				if space == nsRDF && (local == "Bag" || local == "Seq" || local == "Alt") {
					parentCtx.propKind = KindArray
					newCtx = &ContextFrame{Type: CTX_ARRAY}
				} else if space == nsRDF && local == "Description" {
					parentCtx.propKind = KindStruct
					fields := parsePropertyAttrs(t.Attr, currNS, namespaces)
					parentCtx.fields = append(parentCtx.fields, fields...)
					newCtx = &ContextFrame{Type: CTX_STRUCT_FIELD, propKind: KindStruct}
				} else {
					parentCtx.propKind = KindStruct
					prefix := resolvePrefix(space, currNS)
					namespaces[space] = prefix // Capture
					newCtx = &ContextFrame{
						Type:       CTX_STRUCT_FIELD,
						propURI:    space,
						propLocal:  local,
						propPrefix: prefix,
					}
					fields := parsePropertyAttrs(t.Attr, currNS, namespaces)
					if len(fields) > 0 {
						newCtx.propKind = KindStruct
						newCtx.fields = fields
					}
				}
			}

			if newCtx == nil {
				newCtx = &ContextFrame{Type: CTX_ROOT}
			}
			ctxStack = append(ctxStack, newCtx)

		case xml.EndElement:
			currCtx := ctxStack[len(ctxStack)-1]
			parentCtx := ctxStack[len(ctxStack)-2]

			switch currCtx.Type {
			case CTX_PROPERTY:
				val := finalizeValue(currCtx)
				key := PropertyKey{currCtx.propURI, currCtx.propLocal}
				nodeMap[key] = append(nodeMap[key], val)

			case CTX_STRUCT_FIELD:
				if currCtx.propLocal != "" {
					val := finalizeValue(currCtx)
					field := StructField{
						Prefix: currCtx.propPrefix,
						URI:    currCtx.propURI,
						Name:   currCtx.propLocal,
						Value:  val,
					}
					if parentCtx.Type == CTX_PROPERTY || parentCtx.Type == CTX_LI || parentCtx.Type == CTX_STRUCT_FIELD {
						parentCtx.fields = append(parentCtx.fields, field)
					}
				} else {
					if parentCtx.Type == CTX_PROPERTY || parentCtx.Type == CTX_LI || parentCtx.Type == CTX_STRUCT_FIELD {
						parentCtx.fields = append(parentCtx.fields, currCtx.fields...)
					}
				}

			case CTX_ARRAY:
				// Array container ended. Transfer items to parent Property.
				if parentCtx.Type == CTX_PROPERTY {
					parentCtx.items = append(parentCtx.items, currCtx.items...)
				}

			case CTX_LI:
				val := finalizeValue(currCtx)
				if parentCtx.Type == CTX_ARRAY {
					parentCtx.items = append(parentCtx.items, val)
				}
			}

			ctxStack = ctxStack[:len(ctxStack)-1]
			nsStack = nsStack[:len(nsStack)-1]

		case xml.CharData:
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
