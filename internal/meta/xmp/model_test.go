package xmp

import (
	"testing"
)

func TestContextType(t *testing.T) {
	tests := []struct {
		name     string
		val      ContextType
		expected string
	}{
		{"ROOT", CTX_ROOT, "ROOT"},
		{"RDF", CTX_RDF, "RDF"},
		{"DESCRIPTION", CTX_DESCRIPTION, "DESCRIPTION"},
		{"PROPERTY", CTX_PROPERTY, "PROPERTY"},
		{"ARRAY", CTX_ARRAY, "ARRAY"},
		{"LI", CTX_LI, "LI"},
		{"STRUCT_FIELD", CTX_STRUCT_FIELD, "STRUCT_FIELD"},
		{"UNKNOWN", ContextType(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.String(); got != tt.expected {
				t.Errorf("ContextType.String() = %q, want %q", got, tt.expected)
			}
		})
	}

	// Verify they're all different values
	seen := make(map[ContextType]bool)
	for _, tt := range tests[:7] { // Exclude unknown
		if seen[tt.val] {
			t.Errorf("Duplicate ContextType value: %d", tt.val)
		}
		seen[tt.val] = true
	}
}

func TestPropKind(t *testing.T) {
	tests := []struct {
		name     string
		val      PropKind
		expected string
	}{
		{"Unknown", KindUnknown, "Unknown"},
		{"Simple", KindSimple, "Simple"},
		{"Array", KindArray, "Array"},
		{"Struct", KindStruct, "Struct"},
		{"Invalid", PropKind(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.String(); got != tt.expected {
				t.Errorf("PropKind.String() = %q, want %q", got, tt.expected)
			}
		})
	}

	// Verify they're all different values
	seen := make(map[PropKind]bool)
	for _, tt := range tests[:4] { // Exclude invalid
		if seen[tt.val] {
			t.Errorf("Duplicate PropKind value: %d", tt.val)
		}
		seen[tt.val] = true
	}
}

func TestContextFrame(t *testing.T) {
	// Test ContextFrame creation and field access
	frame := &ContextFrame{
		Type:       CTX_PROPERTY,
		propURI:    "http://example.com/ns/",
		propLocal:  "test",
		propPrefix: "ex",
		propKind:   KindSimple,
	}

	if frame.Type != CTX_PROPERTY {
		t.Errorf("frame.Type = %v, want %v", frame.Type, CTX_PROPERTY)
	}

	frame.text.WriteString("test")
	if frame.text.String() != "test" {
		t.Errorf("text builder = %q, want %q", frame.text.String(), "test")
	}

	frame.items = append(frame.items, PropertyValue{Kind: KindSimple, Scalar: "item"})
	if len(frame.items) != 1 {
		t.Errorf("items length = %d, want 1", len(frame.items))
	}

	frame.fields = append(frame.fields, StructField{Prefix: "ex", Name: "field"})
	if len(frame.fields) != 1 {
		t.Errorf("fields length = %d, want 1", len(frame.fields))
	}
}

func TestPropertyKey(t *testing.T) {
	// Test PropertyKey as map key
	nodeMap := make(NodeMap)
	key := PropertyKey{URI: "http://example.com/ns/", Local: "test"}

	nodeMap[key] = []PropertyValue{{Kind: KindSimple, Scalar: "value"}}

	if val, ok := nodeMap[key]; !ok || len(val) != 1 {
		t.Error("PropertyKey not working as map key")
	}
}

func TestStructField(t *testing.T) {
	// Test StructField creation
	field := StructField{
		Prefix: "ex",
		URI:    "http://example.com/ns/",
		Name:   "field",
		Value:  PropertyValue{Kind: KindSimple, Scalar: "value"},
	}

	if field.Prefix != "ex" {
		t.Errorf("field.Prefix = %q, want %q", field.Prefix, "ex")
	}
	if field.Value.Scalar != "value" {
		t.Errorf("field.Value.Scalar = %q, want %q", field.Value.Scalar, "value")
	}
}

func TestPropertyValue(t *testing.T) {
	// Test simple value
	simple := PropertyValue{Kind: KindSimple, Scalar: "test"}
	if simple.Kind != KindSimple || simple.Scalar != "test" {
		t.Error("Simple PropertyValue not created correctly")
	}

	// Test array value
	array := PropertyValue{
		Kind: KindArray,
		Items: []PropertyValue{
			{Kind: KindSimple, Scalar: "item1"},
			{Kind: KindSimple, Scalar: "item2"},
		},
	}
	if array.Kind != KindArray || len(array.Items) != 2 {
		t.Error("Array PropertyValue not created correctly")
	}

	// Test struct value
	strct := PropertyValue{
		Kind: KindStruct,
		Fields: []StructField{
			{Prefix: "ex", Name: "field1", Value: PropertyValue{Kind: KindSimple, Scalar: "val1"}},
		},
	}
	if strct.Kind != KindStruct || len(strct.Fields) != 1 {
		t.Error("Struct PropertyValue not created correctly")
	}
}

func TestNSFrame(t *testing.T) {
	// Test NSFrame creation and usage
	ns := &NSFrame{
		prefixToURI: make(map[string]string),
		uriToPrefix: make(map[string]string),
	}

	ns.prefixToURI["ex"] = "http://example.com/ns/"
	ns.uriToPrefix["http://example.com/ns/"] = "ex"

	if ns.prefixToURI["ex"] != "http://example.com/ns/" {
		t.Error("NSFrame prefixToURI not working")
	}
	if ns.uriToPrefix["http://example.com/ns/"] != "ex" {
		t.Error("NSFrame uriToPrefix not working")
	}
}
