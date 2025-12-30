package xmp

import (
	"testing"
)

func TestFlattenNodeMap(t *testing.T) {
	tests := []struct {
		name       string
		nodeMap    NodeMap
		namespaces map[string]string
		wantDirs   int
	}{
		{
			name:       "empty nodeMap",
			nodeMap:    NodeMap{},
			namespaces: map[string]string{},
			wantDirs:   0,
		},
		{
			name: "single simple property",
			nodeMap: NodeMap{
				{URI: "http://purl.org/dc/elements/1.1/", Local: "title"}: {
					{Kind: KindSimple, Scalar: "Test Title"},
				},
			},
			namespaces: map[string]string{"http://purl.org/dc/elements/1.1/": "dc"},
			wantDirs:   1,
		},
		{
			name: "multiple properties same namespace",
			nodeMap: NodeMap{
				{URI: "http://purl.org/dc/elements/1.1/", Local: "title"}: {
					{Kind: KindSimple, Scalar: "Title"},
				},
				{URI: "http://purl.org/dc/elements/1.1/", Local: "creator"}: {
					{Kind: KindSimple, Scalar: "Author"},
				},
			},
			namespaces: map[string]string{"http://purl.org/dc/elements/1.1/": "dc"},
			wantDirs:   1,
		},
		{
			name: "properties from different namespaces",
			nodeMap: NodeMap{
				{URI: "http://purl.org/dc/elements/1.1/", Local: "title"}: {
					{Kind: KindSimple, Scalar: "Title"},
				},
				{URI: "http://ns.adobe.com/xap/1.0/", Local: "Rating"}: {
					{Kind: KindSimple, Scalar: "5"},
				},
			},
			namespaces: map[string]string{
				"http://purl.org/dc/elements/1.1/": "dc",
				"http://ns.adobe.com/xap/1.0/":     "xmp",
			},
			wantDirs: 2,
		},
		{
			name: "unknown namespace - use wellKnown",
			nodeMap: NodeMap{
				{URI: "http://purl.org/dc/elements/1.1/", Local: "title"}: {
					{Kind: KindSimple, Scalar: "Title"},
				},
			},
			namespaces: map[string]string{}, // Empty, should fallback to wellKnown
			wantDirs:   1,
		},
		{
			name: "unknown namespace - use default prefix",
			nodeMap: NodeMap{
				{URI: "http://unknown.namespace.com/", Local: "prop"}: {
					{Kind: KindSimple, Scalar: "value"},
				},
			},
			namespaces: map[string]string{},
			wantDirs:   1,
		},
		{
			name: "multiple values for same key",
			nodeMap: NodeMap{
				{URI: "http://purl.org/dc/elements/1.1/", Local: "subject"}: {
					{Kind: KindSimple, Scalar: "keyword1"},
					{Kind: KindSimple, Scalar: "keyword2"},
				},
			},
			namespaces: map[string]string{"http://purl.org/dc/elements/1.1/": "dc"},
			wantDirs:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs := flattenNodeMap(tt.nodeMap, tt.namespaces)
			if len(dirs) != tt.wantDirs {
				t.Errorf("flattenNodeMap() returned %d dirs, want %d", len(dirs), tt.wantDirs)
			}
		})
	}
}

func TestFlattenVal(t *testing.T) {
	tests := []struct {
		name     string
		value    PropertyValue
		wantType string
	}{
		{
			name:     "simple string",
			value:    PropertyValue{Kind: KindSimple, Scalar: "hello"},
			wantType: "string",
		},
		{
			name:     "simple bool true",
			value:    PropertyValue{Kind: KindSimple, Scalar: "true"},
			wantType: "bool",
		},
		{
			name:     "simple bool false",
			value:    PropertyValue{Kind: KindSimple, Scalar: "false"},
			wantType: "bool",
		},
		{
			name:     "simple int",
			value:    PropertyValue{Kind: KindSimple, Scalar: "42"},
			wantType: "int",
		},
		{
			name:     "simple float",
			value:    PropertyValue{Kind: KindSimple, Scalar: "3.14"},
			wantType: "float",
		},
		{
			name: "array of strings",
			value: PropertyValue{
				Kind: KindArray,
				Items: []PropertyValue{
					{Kind: KindSimple, Scalar: "item1"},
					{Kind: KindSimple, Scalar: "item2"},
				},
			},
			wantType: "array",
		},
		{
			name: "struct",
			value: PropertyValue{
				Kind: KindStruct,
				Fields: []StructField{
					{Prefix: "dc", Name: "title", Value: PropertyValue{Kind: KindSimple, Scalar: "Test"}},
				},
			},
			wantType: "struct",
		},
		{
			name:     "unknown kind",
			value:    PropertyValue{Kind: KindUnknown},
			wantType: unknownDataType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, dataType := flattenVal(tt.value)
			if dataType != tt.wantType {
				t.Errorf("flattenVal() type = %q, want %q", dataType, tt.wantType)
			}
		})
	}
}

func TestFlattenVal_Values(t *testing.T) {
	t.Run("bool true", func(t *testing.T) {
		val, _ := flattenVal(PropertyValue{Kind: KindSimple, Scalar: "true"})
		if val != true {
			t.Errorf("flattenVal() = %v, want true", val)
		}
	})

	t.Run("bool false", func(t *testing.T) {
		val, _ := flattenVal(PropertyValue{Kind: KindSimple, Scalar: "FALSE"})
		if val != false {
			t.Errorf("flattenVal() = %v, want false", val)
		}
	})

	t.Run("int value", func(t *testing.T) {
		val, _ := flattenVal(PropertyValue{Kind: KindSimple, Scalar: "123"})
		if v, ok := val.(int); !ok || v != 123 {
			t.Errorf("flattenVal() = %v (%T), want int 123", val, val)
		}
	})

	t.Run("float value", func(t *testing.T) {
		val, _ := flattenVal(PropertyValue{Kind: KindSimple, Scalar: "3.14"})
		if v, ok := val.(float64); !ok || v != 3.14 {
			t.Errorf("flattenVal() = %v (%T), want float 3.14", val, val)
		}
	})

	t.Run("array value", func(t *testing.T) {
		val, _ := flattenVal(PropertyValue{
			Kind: KindArray,
			Items: []PropertyValue{
				{Kind: KindSimple, Scalar: "a"},
				{Kind: KindSimple, Scalar: "b"},
			},
		})
		arr, ok := val.([]any)
		if !ok {
			t.Fatalf("flattenVal() not []any, got %T", val)
		}
		if len(arr) != 2 {
			t.Errorf("array len = %d, want 2", len(arr))
		}
	})

	t.Run("struct value", func(t *testing.T) {
		val, _ := flattenVal(PropertyValue{
			Kind: KindStruct,
			Fields: []StructField{
				{Prefix: "dc", Name: "title", Value: PropertyValue{Kind: KindSimple, Scalar: "Test"}},
			},
		})
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("flattenVal() not map, got %T", val)
		}
		if m["dc:title"] != "Test" {
			t.Errorf("struct field = %v, want 'Test'", m["dc:title"])
		}
	})

	t.Run("empty array", func(t *testing.T) {
		val, dt := flattenVal(PropertyValue{Kind: KindArray, Items: nil})
		arr, ok := val.([]any)
		if !ok {
			t.Fatalf("flattenVal() not []any, got %T", val)
		}
		if len(arr) != 0 {
			t.Errorf("array len = %d, want 0", len(arr))
		}
		if dt != "array" {
			t.Errorf("dataType = %q, want 'array'", dt)
		}
	})

	t.Run("empty struct", func(t *testing.T) {
		val, dt := flattenVal(PropertyValue{Kind: KindStruct, Fields: nil})
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("flattenVal() not map, got %T", val)
		}
		if len(m) != 0 {
			t.Errorf("map len = %d, want 0", len(m))
		}
		if dt != "struct" {
			t.Errorf("dataType = %q, want 'struct'", dt)
		}
	})
}

func TestFinalizeValue(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *ContextFrame
		wantKind PropKind
	}{
		{
			name: "array by kind",
			ctx: &ContextFrame{
				propKind: KindArray,
				items:    []PropertyValue{{Kind: KindSimple, Scalar: "item"}},
			},
			wantKind: KindArray,
		},
		{
			name: "array by items",
			ctx: &ContextFrame{
				propKind: KindUnknown,
				items:    []PropertyValue{{Kind: KindSimple, Scalar: "item"}},
			},
			wantKind: KindArray,
		},
		{
			name: "struct by kind",
			ctx: &ContextFrame{
				propKind: KindStruct,
				fields:   []StructField{{Name: "field"}},
			},
			wantKind: KindStruct,
		},
		{
			name: "struct by fields",
			ctx: &ContextFrame{
				propKind: KindUnknown,
				fields:   []StructField{{Name: "field"}},
			},
			wantKind: KindStruct,
		},
		{
			name: "simple text",
			ctx: &ContextFrame{
				propKind: KindUnknown,
			},
			wantKind: KindSimple,
		},
		{
			name: "simple text with content",
			ctx: func() *ContextFrame {
				c := &ContextFrame{propKind: KindUnknown}
				c.text.WriteString("  hello world  ")
				return c
			}(),
			wantKind: KindSimple,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := finalizeValue(tt.ctx)
			if val.Kind != tt.wantKind {
				t.Errorf("finalizeValue() Kind = %v, want %v", val.Kind, tt.wantKind)
			}
		})
	}
}

func TestFinalizeValue_TextTrimmed(t *testing.T) {
	ctx := &ContextFrame{propKind: KindUnknown}
	ctx.text.WriteString("  trimmed text  ")

	val := finalizeValue(ctx)
	if val.Kind != KindSimple {
		t.Errorf("Kind = %v, want KindSimple", val.Kind)
	}
	if val.Scalar != "trimmed text" {
		t.Errorf("Scalar = %q, want 'trimmed text'", val.Scalar)
	}
}

func TestFlattenNodeMap_TagProperties(t *testing.T) {
	nodeMap := NodeMap{
		{URI: "http://purl.org/dc/elements/1.1/", Local: "title"}: {
			{Kind: KindSimple, Scalar: "My Title"},
		},
	}
	namespaces := map[string]string{"http://purl.org/dc/elements/1.1/": "dc"}

	dirs := flattenNodeMap(nodeMap, namespaces)
	if len(dirs) != 1 {
		t.Fatalf("expected 1 directory, got %d", len(dirs))
	}

	dir := dirs[0]
	if dir.Name != "XMP-dc" {
		t.Errorf("directory name = %q, want 'XMP-dc'", dir.Name)
	}

	if len(dir.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(dir.Tags))
	}

	tag := dir.Tags[0]
	if string(tag.ID) != "XMP-dc:title" {
		t.Errorf("tag ID = %q, want 'XMP-dc:title'", tag.ID)
	}
	if tag.Name != "title" {
		t.Errorf("tag Name = %q, want 'title'", tag.Name)
	}
	if tag.Value != "My Title" {
		t.Errorf("tag Value = %v, want 'My Title'", tag.Value)
	}
}
