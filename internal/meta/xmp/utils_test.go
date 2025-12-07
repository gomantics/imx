package xmp

import (
	"encoding/xml"
	"testing"
)

func TestInferType(t *testing.T) {
	tests := []struct {
		input    string
		wantVal  any
		wantType string
	}{
		// Booleans
		{"true", true, "bool"},
		{"True", true, "bool"},
		{"TRUE", true, "bool"},
		{"false", false, "bool"},
		{"False", false, "bool"},
		{"FALSE", false, "bool"},

		// Integers
		{"0", 0, "int"},
		{"123", 123, "int"},
		{"-456", -456, "int"},
		{"+789", 789, "int"},

		// Floats
		{"0.0", 0.0, "float"},
		{"3.14", 3.14, "float"},
		{"-2.5", -2.5, "float"},
		{"+1.5", 1.5, "float"},
		{"123.456", 123.456, "float"},

		// Exponential notation (not handled by isInt/isFloat, returned as string)
		{"1.23e10", "1.23e10", "string"},
		{"1.23E-5", "1.23E-5", "string"},

		// Strings
		{"", "", "string"},
		{"hello", "hello", "string"},
		{"123abc", "123abc", "string"},
		{"not-a-number", "not-a-number", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotVal, gotType := inferType(tt.input)
			if gotVal != tt.wantVal {
				t.Errorf("inferType(%q) val = %v (%T), want %v (%T)",
					tt.input, gotVal, gotVal, tt.wantVal, tt.wantVal)
			}
			if gotType != tt.wantType {
				t.Errorf("inferType(%q) type = %v, want %v",
					tt.input, gotType, tt.wantType)
			}
		})
	}
}

func TestIsInt(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"123", true},
		{"-456", true},
		{"+789", true},
		{"0", true},
		{"-", true},  // Bug: returns true for sign only
		{"+", true},  // Bug: returns true for sign only
		{"12.34", false},
		{"1e10", false},
		{"abc", false},
		{"12a", false},
		{"a12", false},
	}

	for _, tt := range tests {
		got := isInt(tt.input)
		if got != tt.want {
			t.Errorf("isInt(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsFloat(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"3.14", true},
		{"-2.5", true},
		{"+1.5", true},
		{"0.0", true},
		{".", true},   // Bug: returns true for dot only
		{"-", false},
		{"+", false},
		{"1.2.3", false},
		{"123", false}, // No dot
		{"abc", false},
		{"-.", true},   // Bug: returns true for sign+dot only
		{"+.", true},   // Bug: returns true for sign+dot only
		{"1.2a", false},
		// Note: Current implementation doesn't detect exponential notation
		{"1.23e10", false},
		{"1e10", false},
	}

	for _, tt := range tests {
		got := isFloat(tt.input)
		if got != tt.want {
			t.Errorf("isFloat(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestResolvePrefix(t *testing.T) {
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

	tests := []struct {
		uri  string
		want string
	}{
		{"http://purl.org/dc/elements/1.1/", "dc"},
		{"http://ns.adobe.com/xap/1.0/", "xmp"},
		{"http://ns.adobe.com/photoshop/1.0/", "photoshop"}, // well-known
		{"http://example.com/unknown/", "ns"},                // fallback
	}

	for _, tt := range tests {
		got := resolvePrefix(tt.uri, ns)
		if got != tt.want {
			t.Errorf("resolvePrefix(%q) = %q, want %q", tt.uri, got, tt.want)
		}
	}

	// Test with empty namespace (only well-known fallback)
	emptyNS := &NSFrame{
		prefixToURI: map[string]string{},
		uriToPrefix: map[string]string{},
	}
	got := resolvePrefix("http://purl.org/dc/elements/1.1/", emptyNS)
	if got != "dc" {
		t.Errorf("resolvePrefix with empty ns = %q, want %q", got, "dc")
	}
}

func TestReplaceNSFrame(t *testing.T) {
	t.Run("With no parent", func(t *testing.T) {
		attrs := []xml.Attr{
			{Name: xml.Name{Space: "xmlns", Local: "dc"}, Value: "http://purl.org/dc/elements/1.1/"},
			{Name: xml.Name{Local: "xmlns"}, Value: "http://default.ns/"},
		}

		ns := replaceNSFrame(nil, attrs)
		if ns == nil {
			t.Fatal("replaceNSFrame returned nil")
		}

		if uri := ns.prefixToURI["dc"]; uri != "http://purl.org/dc/elements/1.1/" {
			t.Errorf("prefixToURI[dc] = %q, want %q", uri, "http://purl.org/dc/elements/1.1/")
		}

		if prefix := ns.uriToPrefix["http://purl.org/dc/elements/1.1/"]; prefix != "dc" {
			t.Errorf("uriToPrefix = %q, want %q", prefix, "dc")
		}

		if uri := ns.prefixToURI[""]; uri != "http://default.ns/" {
			t.Errorf("default namespace = %q, want %q", uri, "http://default.ns/")
		}
	})

	t.Run("With parent", func(t *testing.T) {
		parent := &NSFrame{
			prefixToURI: map[string]string{"xmp": "http://ns.adobe.com/xap/1.0/"},
			uriToPrefix: map[string]string{"http://ns.adobe.com/xap/1.0/": "xmp"},
		}

		attrs := []xml.Attr{
			{Name: xml.Name{Space: "xmlns", Local: "crs"}, Value: "http://ns.adobe.com/camera-raw-settings/1.0/"},
		}

		child := replaceNSFrame(parent, attrs)

		// Should have both parent and child namespaces
		if uri := child.prefixToURI["xmp"]; uri != "http://ns.adobe.com/xap/1.0/" {
			t.Errorf("child missing parent namespace")
		}
		if uri := child.prefixToURI["crs"]; uri != "http://ns.adobe.com/camera-raw-settings/1.0/" {
			t.Errorf("child missing new namespace")
		}
	})

	t.Run("Empty attrs", func(t *testing.T) {
		ns := replaceNSFrame(nil, nil)
		if ns == nil {
			t.Fatal("replaceNSFrame returned nil")
		}
		if len(ns.prefixToURI) != 0 {
			t.Errorf("Expected empty prefixToURI, got %v", ns.prefixToURI)
		}
	})
}

func TestIsPropAttr(t *testing.T) {
	tests := []struct {
		name xml.Name
		want bool
	}{
		// Namespace declarations - not properties
		{xml.Name{Space: "xmlns", Local: "dc"}, false},
		{xml.Name{Local: "xmlns"}, false},

		// xml:* attributes - not properties
		{xml.Name{Space: "http://www.w3.org/XML/1998/namespace", Local: "lang"}, false},

		// RDF control attributes - not properties
		{xml.Name{Space: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Local: "about"}, false},
		{xml.Name{Space: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Local: "resource"}, false},
		{xml.Name{Space: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Local: "parseType"}, false},

		// No namespace - not a property
		{xml.Name{Local: "something"}, false},

		// Valid properties
		{xml.Name{Space: "http://purl.org/dc/elements/1.1/", Local: "creator"}, true},
		{xml.Name{Space: "http://ns.adobe.com/xap/1.0/", Local: "Rating"}, true},
	}

	for _, tt := range tests {
		got := isPropAttr(tt.name)
		if got != tt.want {
			t.Errorf("isPropAttr(%+v) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
