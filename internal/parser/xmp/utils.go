package xmp

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
)

func replaceNSFrame(parent *NSFrame, attrs []xml.Attr) *NSFrame {
	hasXMLNS := false
	for _, attr := range attrs {
		if attr.Name.Space == "xmlns" || (attr.Name.Local == "xmlns" && attr.Name.Space == "") {
			hasXMLNS = true
			break
		}
	}

	if !hasXMLNS && parent != nil {
		return parent
	}

	parentSize := 0
	if parent != nil {
		parentSize = len(parent.prefixToURI)
	}
	expectedSize := parentSize + len(attrs)

	newFrame := &NSFrame{
		prefixToURI: make(map[string]string, expectedSize),
		uriToPrefix: make(map[string]string, expectedSize),
	}

	if parent != nil {
		for k, v := range parent.prefixToURI {
			newFrame.prefixToURI[k] = v
		}
		for k, v := range parent.uriToPrefix {
			newFrame.uriToPrefix[k] = v
		}
	}

	for _, attr := range attrs {
		if attr.Name.Space == "xmlns" {
			prefix := attr.Name.Local
			uri := attr.Value
			newFrame.prefixToURI[prefix] = uri
			newFrame.uriToPrefix[uri] = prefix
		} else if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
			newFrame.prefixToURI[""] = attr.Value
		}
	}
	return newFrame
}

func resolvePrefix(uri string, ns *NSFrame) string {
	if p, ok := ns.uriToPrefix[uri]; ok {
		return p
	}
	if p, ok := wellKnownPrefixes[uri]; ok {
		return p
	}
	return defaultPrefix
}

func isPropAttr(name xml.Name) bool {
	if name.Space == "xmlns" || (name.Space == "" && name.Local == "xmlns") {
		return false
	}
	if name.Space == nsXML {
		return false
	}
	if name.Space == nsRDF && (name.Local == "about" || name.Local == "resource" || name.Local == "parseType") {
		return false
	}
	return name.Space != ""
}

func stripXPacket(data []byte) []byte {
	if i := bytes.Index(data, []byte("<?xpacket begin=")); i >= 0 {
		if end := bytes.Index(data[i:], []byte("?>")); end >= 0 {
			data = data[i+end+2:]
		}
	}
	if i := bytes.Index(data, []byte("<?xpacket end=")); i >= 0 {
		data = data[:i]
	}
	return bytes.TrimSpace(data)
}

func inferType(s string) (any, string) {
	lower := strings.ToLower(s)
	if lower == "true" {
		return true, "bool"
	}
	if lower == "false" {
		return false, "bool"
	}
	if isInt(s) {
		if i, err := strconv.Atoi(s); err == nil {
			return i, "int"
		}
	}
	if isFloat(s) {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f, "float"
		}
	}
	return s, "string"
}

func isInt(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		if i == 0 && (c == '-' || c == '+') {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isFloat(s string) bool {
	if s == "" {
		return false
	}
	hasDot := false
	for i, c := range s {
		if i == 0 && (c == '-' || c == '+') {
			continue
		}
		if c == '.' {
			if hasDot {
				return false
			}
			hasDot = true
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return hasDot
}

func isArrayContainer(space, local string) bool {
	return space == nsRDF && (local == "Bag" || local == "Seq" || local == "Alt")
}

func isRDFDescription(space, local string) bool {
	return space == nsRDF && local == "Description"
}

func isRDFLi(space, local string) bool {
	return space == nsRDF && local == "li"
}

func createStructFieldContext(space, local string, ns *NSFrame, attrs []xml.Attr, namespaces map[string]string) *ContextFrame {
	prefix := resolvePrefix(space, ns)
	namespaces[space] = prefix

	ctx := &ContextFrame{
		Type:       CTX_STRUCT_FIELD,
		propURI:    space,
		propLocal:  local,
		propPrefix: prefix,
	}

	fields := parsePropertyAttrs(attrs, ns, namespaces)
	if len(fields) > 0 {
		ctx.propKind = KindStruct
		ctx.fields = fields
	}

	return ctx
}

func parsePropertyAttrs(attrs []xml.Attr, ns *NSFrame, namespaces map[string]string) []StructField {
	var fields []StructField
	for _, attr := range attrs {
		if isPropAttr(attr.Name) {
			prefix := resolvePrefix(attr.Name.Space, ns)
			namespaces[attr.Name.Space] = prefix
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
