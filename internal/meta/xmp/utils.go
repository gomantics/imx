package xmp

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
)

var wellKnownPrefixes = map[string]string{
	"http://ns.adobe.com/xap/1.0/":                         "xmp",
	"http://ns.adobe.com/xap/1.0/mm/":                      "xmpMM",
	"http://ns.adobe.com/xap/1.0/st/":                      "xmpST",
	"http://ns.adobe.com/xap/1.0/rights/":                  "xmpRights",
	"http://purl.org/dc/elements/1.1/":                     "dc",
	"http://iptc.org/std/Iptc4xmpCore/1.0/xmlns/":          "Iptc4xmpCore",
	"http://ns.adobe.com/photoshop/1.0/":                   "photoshop",
	"http://ns.adobe.com/tiff/1.0/":                        "tiff",
	"http://ns.adobe.com/exif/1.0/":                        "exif",
	"http://ns.adobe.com/camera-raw-settings/1.0/":         "crs",
	"http://www.metadataworkinggroup.com/schemas/regions/": "mwg-rs",
	"http://ns.apple.com/faceinfo/1.0/":                    "apple-fi",
	"http://ns.adobe.com/xmp/sType/Area#":                  "stArea",
	"http://ns.adobe.com/xap/1.0/sType/Dimensions#":        "stDim",
	"http://ns.adobe.com/xap/1.0/sType/ResourceEvent#":     "stEvt",
}

func replaceNSFrame(parent *NSFrame, attrs []xml.Attr) *NSFrame {
	// Clone or Create
	newFrame := &NSFrame{
		prefixToURI: make(map[string]string),
		uriToPrefix: make(map[string]string),
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
			// xmlns:prefix="uri"
			prefix := attr.Name.Local
			uri := attr.Value
			newFrame.prefixToURI[prefix] = uri
			newFrame.uriToPrefix[uri] = prefix
		} else if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
			// default ns
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
	return "ns"
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

// isArrayContainer checks if an element is an RDF array container (Bag, Seq, Alt).
func isArrayContainer(space, local string) bool {
	return space == nsRDF && (local == "Bag" || local == "Seq" || local == "Alt")
}

// isRDFDescription checks if an element is an RDF Description.
func isRDFDescription(space, local string) bool {
	return space == nsRDF && local == "Description"
}

// isRDFLi checks if an element is an RDF li (list item).
func isRDFLi(space, local string) bool {
	return space == nsRDF && local == "li"
}

// createStructFieldContext creates a new struct field context frame.
// This helper reduces duplication across multiple state handlers.
func createStructFieldContext(space, local string, ns *NSFrame, attrs []xml.Attr, namespaces map[string]string) *ContextFrame {
	prefix := resolvePrefix(space, ns)
	namespaces[space] = prefix

	ctx := &ContextFrame{
		Type:       CTX_STRUCT_FIELD,
		propURI:    space,
		propLocal:  local,
		propPrefix: prefix,
	}

	// Check for struct attributes (shorthand struct notation)
	fields := parsePropertyAttrs(attrs, ns, namespaces)
	if len(fields) > 0 {
		ctx.propKind = KindStruct
		ctx.fields = fields
	}

	return ctx
}
