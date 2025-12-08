package xmp

import (
	"encoding/xml"
	"testing"
)

func TestHandlerRegistry(t *testing.T) {
	t.Run("NewHandlerRegistry creates registry with all handlers", func(t *testing.T) {
		registry := NewHandlerRegistry()
		if registry == nil {
			t.Fatal("NewHandlerRegistry returned nil")
		}

		// Verify all context types have handlers
		contexts := []ContextType{
			CTX_ROOT,
			CTX_RDF,
			CTX_DESCRIPTION,
			CTX_PROPERTY,
			CTX_ARRAY,
			CTX_LI,
			CTX_STRUCT_FIELD,
		}

		for _, ctx := range contexts {
			handler := registry.Get(ctx)
			if handler == nil {
				t.Errorf("No handler registered for context %s", ctx)
			}
		}
	})

	t.Run("Get returns fallback for unknown context", func(t *testing.T) {
		registry := NewHandlerRegistry()

		// Request handler for invalid context type
		unknownCtx := ContextType(999)
		handler := registry.Get(unknownCtx)

		if handler == nil {
			t.Error("Get should return fallback handler for unknown context, not nil")
		}

		// The fallback should be the ROOT handler
		rootHandler := registry.Get(CTX_ROOT)
		if handler != rootHandler {
			t.Error("Fallback handler should be ROOT handler")
		}
	})
}

func TestFinalizeValue(t *testing.T) {
	t.Run("Array value", func(t *testing.T) {
		ctx := &ContextFrame{
			Type:     CTX_PROPERTY,
			propKind: KindArray,
			items: []PropertyValue{
				{Kind: KindSimple, Scalar: "item1"},
				{Kind: KindSimple, Scalar: "item2"},
			},
		}

		val := finalizeValue(ctx)
		if val.Kind != KindArray {
			t.Errorf("finalizeValue kind = %v, want KindArray", val.Kind)
		}
		if len(val.Items) != 2 {
			t.Errorf("finalizeValue items length = %d, want 2", len(val.Items))
		}
	})

	t.Run("Struct value", func(t *testing.T) {
		ctx := &ContextFrame{
			Type:     CTX_PROPERTY,
			propKind: KindStruct,
			fields: []StructField{
				{Prefix: "ns", Name: "field1", Value: PropertyValue{Kind: KindSimple, Scalar: "val1"}},
			},
		}

		val := finalizeValue(ctx)
		if val.Kind != KindStruct {
			t.Errorf("finalizeValue kind = %v, want KindStruct", val.Kind)
		}
		if len(val.Fields) != 1 {
			t.Errorf("finalizeValue fields length = %d, want 1", len(val.Fields))
		}
	})

	t.Run("Simple value", func(t *testing.T) {
		ctx := &ContextFrame{
			Type: CTX_PROPERTY,
		}
		ctx.text.WriteString("  simple text  ")

		val := finalizeValue(ctx)
		if val.Kind != KindSimple {
			t.Errorf("finalizeValue kind = %v, want KindSimple", val.Kind)
		}
		if val.Scalar != "simple text" {
			t.Errorf("finalizeValue scalar = %q, want %q", val.Scalar, "simple text")
		}
	})

	t.Run("Items without explicit kind", func(t *testing.T) {
		ctx := &ContextFrame{
			Type: CTX_PROPERTY,
			items: []PropertyValue{
				{Kind: KindSimple, Scalar: "item"},
			},
		}

		val := finalizeValue(ctx)
		if val.Kind != KindArray {
			t.Errorf("finalizeValue kind = %v, want KindArray (items present)", val.Kind)
		}
	})

	t.Run("Fields without explicit kind", func(t *testing.T) {
		ctx := &ContextFrame{
			Type: CTX_PROPERTY,
			fields: []StructField{
				{Prefix: "ns", Name: "field", Value: PropertyValue{Kind: KindSimple, Scalar: "val"}},
			},
		}

		val := finalizeValue(ctx)
		if val.Kind != KindStruct {
			t.Errorf("finalizeValue kind = %v, want KindStruct (fields present)", val.Kind)
		}
	})
}

func TestParsePropertyAttrs(t *testing.T) {
	t.Run("Creates struct fields from attributes", func(t *testing.T) {
		attrs := []xml.Attr{
			{Name: xml.Name{Space: "http://example.com/ns/", Local: "width"}, Value: "1920"},
			{Name: xml.Name{Space: "http://example.com/ns/", Local: "height"}, Value: "1080"},
		}

		ns := &NSFrame{
			prefixToURI: map[string]string{"ns": "http://example.com/ns/"},
			uriToPrefix: map[string]string{"http://example.com/ns/": "ns"},
		}

		namespaces := make(map[string]string)
		fields := parsePropertyAttrs(attrs, ns, namespaces)

		if len(fields) != 2 {
			t.Fatalf("fields length = %d, want 2", len(fields))
		}

		if fields[0].Prefix != "ns" || fields[0].Name != "width" || fields[0].Value.Scalar != "1920" {
			t.Errorf("First field incorrect: %+v", fields[0])
		}

		if fields[1].Prefix != "ns" || fields[1].Name != "height" || fields[1].Value.Scalar != "1080" {
			t.Errorf("Second field incorrect: %+v", fields[1])
		}
	})

	t.Run("Filters non-property attributes", func(t *testing.T) {
		attrs := []xml.Attr{
			{Name: xml.Name{Space: "xmlns", Local: "ns"}, Value: "http://example.com/ns/"},
			{Name: xml.Name{Local: "xmlns"}, Value: "http://default.ns/"},
			{Name: xml.Name{Space: "http://www.w3.org/1999/02/22-rdf-syntax-ns#", Local: "parseType"}, Value: "Resource"},
			{Name: xml.Name{Space: "http://example.com/ns/", Local: "valid"}, Value: "yes"},
		}

		ns := &NSFrame{
			prefixToURI: map[string]string{},
			uriToPrefix: map[string]string{},
		}

		namespaces := make(map[string]string)
		fields := parsePropertyAttrs(attrs, ns, namespaces)

		// Only the valid property should be included
		if len(fields) != 1 {
			t.Errorf("fields length = %d, want 1", len(fields))
		}

		if fields[0].Name != "valid" {
			t.Errorf("field name = %q, want %q", fields[0].Name, "valid")
		}
	})

	t.Run("Returns empty for no property attributes", func(t *testing.T) {
		attrs := []xml.Attr{
			{Name: xml.Name{Space: "xmlns", Local: "dc"}, Value: "http://purl.org/dc/elements/1.1/"},
		}

		ns := &NSFrame{
			prefixToURI: map[string]string{},
			uriToPrefix: map[string]string{},
		}

		namespaces := make(map[string]string)
		fields := parsePropertyAttrs(attrs, ns, namespaces)

		if len(fields) != 0 {
			t.Errorf("fields length = %d, want 0", len(fields))
		}
	})
}
