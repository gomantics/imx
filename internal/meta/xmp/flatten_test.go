package xmp

import (
	"testing"

	"github.com/gomantics/imx/internal/common"
)

func TestFlattenNodeMap(t *testing.T) {
	t.Run("Simple values", func(t *testing.T) {
		nodeMap := NodeMap{
			PropertyKey{URI: "http://purl.org/dc/elements/1.1/", Local: "creator"}: []PropertyValue{
				{Kind: KindSimple, Scalar: "John Doe"},
			},
		}

		namespaces := map[string]string{
			"http://purl.org/dc/elements/1.1/": "dc",
		}

		dir := flattenNodeMap(nodeMap, namespaces)

		if dir.Spec != common.SpecXMP {
			t.Errorf("dir.Spec = %v, want %v", dir.Spec, common.SpecXMP)
		}

		if dir.Name != "XMP" {
			t.Errorf("dir.Name = %q, want %q", dir.Name, "XMP")
		}

		tag, ok := dir.Tags["XMP-dc:creator"]
		if !ok {
			t.Fatal("Missing creator tag")
		}
		if tag.Value != "John Doe" {
			t.Errorf("creator value = %v, want %v", tag.Value, "John Doe")
		}
	})

	t.Run("Array values", func(t *testing.T) {
		nodeMap := NodeMap{
			PropertyKey{URI: "http://purl.org/dc/elements/1.1/", Local: "subject"}: []PropertyValue{
				{
					Kind: KindArray,
					Items: []PropertyValue{
						{Kind: KindSimple, Scalar: "keyword1"},
						{Kind: KindSimple, Scalar: "keyword2"},
					},
				},
			},
		}

		namespaces := map[string]string{
			"http://purl.org/dc/elements/1.1/": "dc",
		}

		dir := flattenNodeMap(nodeMap, namespaces)

		tag, ok := dir.Tags["XMP-dc:subject"]
		if !ok {
			t.Fatal("Missing subject tag")
		}

		arr, ok := tag.Value.([]any)
		if !ok {
			t.Fatalf("subject value is not array: %T", tag.Value)
		}
		if len(arr) != 2 {
			t.Errorf("subject array length = %d, want 2", len(arr))
		}
	})

	t.Run("Struct values", func(t *testing.T) {
		nodeMap := NodeMap{
			PropertyKey{URI: "http://example.com/ns/", Local: "dimensions"}: []PropertyValue{
				{
					Kind: KindStruct,
					Fields: []StructField{
						{Prefix: "ns", Name: "width", Value: PropertyValue{Kind: KindSimple, Scalar: "1920"}},
						{Prefix: "ns", Name: "height", Value: PropertyValue{Kind: KindSimple, Scalar: "1080"}},
					},
				},
			},
		}

		namespaces := map[string]string{
			"http://example.com/ns/": "ns",
		}

		dir := flattenNodeMap(nodeMap, namespaces)

		tag, ok := dir.Tags["XMP-ns:dimensions"]
		if !ok {
			t.Fatal("Missing dimensions tag")
		}

		m, ok := tag.Value.(map[string]any)
		if !ok {
			t.Fatalf("dimensions value is not map: %T", tag.Value)
		}

		if m["ns:width"] != 1920 {
			t.Errorf("width = %v, want 1920", m["ns:width"])
		}
		if m["ns:height"] != 1080 {
			t.Errorf("height = %v, want 1080", m["ns:height"])
		}
	})

	t.Run("Multiple values become array", func(t *testing.T) {
		nodeMap := NodeMap{
			PropertyKey{URI: "http://purl.org/dc/elements/1.1/", Local: "creator"}: []PropertyValue{
				{Kind: KindSimple, Scalar: "Author 1"},
				{Kind: KindSimple, Scalar: "Author 2"},
			},
		}

		namespaces := map[string]string{
			"http://purl.org/dc/elements/1.1/": "dc",
		}

		dir := flattenNodeMap(nodeMap, namespaces)

		tag, ok := dir.Tags["XMP-dc:creator"]
		if !ok {
			t.Fatal("Missing creator tag")
		}

		arr, ok := tag.Value.([]any)
		if !ok {
			t.Fatalf("Expected array for multiple values, got %T", tag.Value)
		}

		if len(arr) != 2 {
			t.Errorf("Array length = %d, want 2", len(arr))
		}
	})

	t.Run("Unknown namespace uses well-known prefix", func(t *testing.T) {
		nodeMap := NodeMap{
			PropertyKey{URI: "http://ns.adobe.com/photoshop/1.0/", Local: "Credit"}: []PropertyValue{
				{Kind: KindSimple, Scalar: "Test"},
			},
		}

		namespaces := map[string]string{} // Empty - should use well-known

		dir := flattenNodeMap(nodeMap, namespaces)

		tag, ok := dir.Tags["XMP-photoshop:Credit"]
		if !ok {
			t.Fatal("Missing Credit tag with photoshop prefix")
		}
		if tag.Value != "Test" {
			t.Errorf("Credit value = %v, want Test", tag.Value)
		}
	})

	t.Run("Truly unknown namespace uses ns fallback", func(t *testing.T) {
		nodeMap := NodeMap{
			PropertyKey{URI: "http://example.com/unknown/", Local: "test"}: []PropertyValue{
				{Kind: KindSimple, Scalar: "value"},
			},
		}

		namespaces := map[string]string{} // Empty

		dir := flattenNodeMap(nodeMap, namespaces)

		tag, ok := dir.Tags["XMP-ns:test"]
		if !ok {
			t.Fatal("Missing test tag with ns fallback prefix")
		}
		if tag.Value != "value" {
			t.Errorf("test value = %v, want value", tag.Value)
		}
	})
}

func TestFlattenVal(t *testing.T) {
	t.Run("Simple value", func(t *testing.T) {
		v := PropertyValue{
			Kind:   KindSimple,
			Scalar: "123",
		}

		val, dataType := flattenVal(v)
		if val != 123 {
			t.Errorf("flattenVal(simple) = %v, want 123", val)
		}
		if dataType != "int" {
			t.Errorf("dataType = %s, want int", dataType)
		}
	})

	t.Run("Array value", func(t *testing.T) {
		v := PropertyValue{
			Kind: KindArray,
			Items: []PropertyValue{
				{Kind: KindSimple, Scalar: "item1"},
				{Kind: KindSimple, Scalar: "item2"},
			},
		}

		val, dataType := flattenVal(v)
		arr, ok := val.([]any)
		if !ok {
			t.Fatalf("flattenVal(array) returned %T, want []any", val)
		}
		if len(arr) != 2 {
			t.Errorf("array length = %d, want 2", len(arr))
		}
		if dataType != "array" {
			t.Errorf("dataType = %s, want array", dataType)
		}
	})

	t.Run("Struct value", func(t *testing.T) {
		v := PropertyValue{
			Kind: KindStruct,
			Fields: []StructField{
				{Prefix: "ns", Name: "field1", Value: PropertyValue{Kind: KindSimple, Scalar: "val1"}},
			},
		}

		val, dataType := flattenVal(v)
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("flattenVal(struct) returned %T, want map", val)
		}
		if m["ns:field1"] != "val1" {
			t.Errorf("field1 = %v, want val1", m["ns:field1"])
		}
		if dataType != "struct" {
			t.Errorf("dataType = %s, want struct", dataType)
		}
	})

	t.Run("Nested array in struct", func(t *testing.T) {
		v := PropertyValue{
			Kind: KindStruct,
			Fields: []StructField{
				{
					Prefix: "ns",
					Name:   "keywords",
					Value: PropertyValue{
						Kind: KindArray,
						Items: []PropertyValue{
							{Kind: KindSimple, Scalar: "kw1"},
							{Kind: KindSimple, Scalar: "kw2"},
						},
					},
				},
			},
		}

		val, dataType := flattenVal(v)
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("flattenVal returned %T, want map", val)
		}

		arr, ok := m["ns:keywords"].([]any)
		if !ok {
			t.Fatalf("keywords is %T, want []any", m["ns:keywords"])
		}
		if len(arr) != 2 {
			t.Errorf("keywords length = %d, want 2", len(arr))
		}
		if dataType != "struct" {
			t.Errorf("dataType = %s, want struct", dataType)
		}
	})

	t.Run("Unknown kind", func(t *testing.T) {
		v := PropertyValue{
			Kind: KindUnknown,
		}

		val, dataType := flattenVal(v)
		if val != nil {
			t.Errorf("flattenVal(unknown) = %v, want nil", val)
		}
		if dataType != "unknown" {
			t.Errorf("dataType = %s, want unknown", dataType)
		}
	})
}
