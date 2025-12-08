package xmp

import (
	"encoding/xml"
	"testing"
)

func TestParseDescriptionAttrs(t *testing.T) {
	t.Run("Simple property attributes", func(t *testing.T) {
		attrs := []xml.Attr{
			{Name: xml.Name{Space: "http://purl.org/dc/elements/1.1/", Local: "format"}, Value: "image/jpeg"},
			{Name: xml.Name{Space: "http://ns.adobe.com/xap/1.0/", Local: "Rating"}, Value: "5"},
		}

		ns := &NSFrame{
			prefixToURI: map[string]string{
				"dc":  "http://purl.org/dc/elements/1.1/",
				"xmp": "http://ns.adobe.com/xap/1.0/",
			},
			uriToPrefix: map[string]string{
				"http://purl.org/dc/elements/1.1/": "dc",
				"http://ns.adobe.com/xap/1.0/":     "xmp",
			},
		}

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)

		parseDescriptionAttrs(attrs, ns, nodeMap, namespaces)

		if len(nodeMap) != 2 {
			t.Errorf("nodeMap length = %d, want 2", len(nodeMap))
		}

		key1 := PropertyKey{URI: "http://purl.org/dc/elements/1.1/", Local: "format"}
		if val, ok := nodeMap[key1]; !ok || len(val) != 1 || val[0].Scalar != "image/jpeg" {
			t.Errorf("Missing or incorrect dc:format")
		}

		key2 := PropertyKey{URI: "http://ns.adobe.com/xap/1.0/", Local: "Rating"}
		if val, ok := nodeMap[key2]; !ok || len(val) != 1 || val[0].Scalar != "5" {
			t.Errorf("Missing or incorrect xmp:Rating")
		}

		if namespaces["http://purl.org/dc/elements/1.1/"] != "dc" {
			t.Errorf("Namespace not captured for dc")
		}
	})

	t.Run("Filters non-property attributes", func(t *testing.T) {
		attrs := []xml.Attr{
			{Name: xml.Name{Space: "xmlns", Local: "dc"}, Value: "http://purl.org/dc/elements/1.1/"},
			{Name: xml.Name{Space: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Local: "about"}, Value: ""},
			{Name: xml.Name{Space: "http://purl.org/dc/elements/1.1/", Local: "format"}, Value: "jpeg"},
		}

		ns := &NSFrame{
			prefixToURI: map[string]string{},
			uriToPrefix: map[string]string{},
		}

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)

		parseDescriptionAttrs(attrs, ns, nodeMap, namespaces)

		// Only dc:format should be added
		if len(nodeMap) != 1 {
			t.Errorf("nodeMap length = %d, want 1 (only actual properties)", len(nodeMap))
		}
	})
}
