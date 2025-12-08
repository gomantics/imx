package xmp

import (
	"reflect"
	"testing"

	"github.com/gomantics/imx/internal/common"
)

func TestParse(t *testing.T) {
	parser := New()

	tests := []struct {
		name    string
		payload string
		want    map[string]any // ID -> Value
	}{
		{
			name: "Simple Attributes",
			payload: `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:xmp="http://ns.adobe.com/xap/1.0/"
    xmlns:crs="http://ns.adobe.com/camera-raw-settings/1.0/"
    xmp:CreatorTool="TestTool"
    xmp:Rating="5"
    crs:Exposure2012="0.50"
    xmp:Switch="True">
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`,
			want: map[string]any{
				"XMP-xmp:CreatorTool":  "TestTool",
				"XMP-xmp:Rating":       5,    // int
				"XMP-crs:Exposure2012": 0.50, // float
				"XMP-xmp:Switch":       true, // bool
			},
		},
		{
			name: "Arrays (Bag/Seq/Alt)",
			payload: `<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:subject>
    <rdf:Bag>
     <rdf:li>keyword1</rdf:li>
     <rdf:li>keyword2</rdf:li>
    </rdf:Bag>
   </dc:subject>
   <dc:title>
    <rdf:Alt>
     <rdf:li xml:lang="x-default">My Title</rdf:li>
    </rdf:Alt>
   </dc:title>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`,
			want: map[string]any{
				"XMP-dc:subject": []any{"keyword1", "keyword2"},
				"XMP-dc:title":   []any{"My Title"},
			},
		},
		{
			name: "Nested Structs",
			payload: `<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:StructProp>
    <rdf:Description ns:Field1="Val1">
     <ns:Field2>Val2</ns:Field2>
    </rdf:Description>
   </ns:StructProp>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`,
			want: map[string]any{
				"XMP-ns:StructProp": map[string]any{
					"ns:Field1": "Val1",
					"ns:Field2": "Val2",
				},
			},
		},
		{
			name: "Unknown Namespace",
			payload: `<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:unknown="http://example.com/unknown/">
   <unknown:Prop>Value</unknown:Prop>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`,
			want: map[string]any{
				"XMP-unknown:Prop": "Value", // extracted prefix
			},
		},
		{
			name: "History Struct Array",
			payload: `<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
          xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/"
          xmlns:stEvt="http://ns.adobe.com/xap/1.0/sType/ResourceEvent#">
  <rdf:Description rdf:about="">
   <xmpMM:History>
    <rdf:Seq>
     <rdf:li rdf:parseType="Resource">
      <stEvt:action>saved</stEvt:action>
      <stEvt:instanceID>xmp.iid:123</stEvt:instanceID>
     </rdf:li>
     <rdf:li rdf:parseType="Resource">
      <stEvt:action>saved</stEvt:action>
      <stEvt:instanceID>xmp.iid:456</stEvt:instanceID>
     </rdf:li>
    </rdf:Seq>
   </xmpMM:History>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`,
			want: map[string]any{
				"XMP-xmpMM:History": []any{
					map[string]any{"stEvt:action": "saved", "stEvt:instanceID": "xmp.iid:123"},
					map[string]any{"stEvt:action": "saved", "stEvt:instanceID": "xmp.iid:456"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := common.RawBlock{
				Spec:    common.SpecXMP,
				Payload: []byte(tt.payload),
			}
			dirs, err := parser.Parse([]common.RawBlock{block})
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if len(dirs) != 1 {
				t.Fatalf("Expected 1 directory, got %d", len(dirs))
			}
			dir := dirs[0]
			for id, wantVal := range tt.want {
				tag, ok := dir.Tags[common.TagID(id)]
				if !ok {
					t.Errorf("Tag %s missing", id)
					continue
				}
				if !reflect.DeepEqual(tag.Value, wantVal) {
					t.Errorf("Tag %s value mismatch: got %v (%T), want %v (%T)", id, tag.Value, tag.Value, wantVal, wantVal)
				}
			}
		})
	}
}

func TestParser_AllBlocksFail(t *testing.T) {
	parser := New()

	// All blocks are malformed XML
	blocks := []common.RawBlock{
		{Spec: common.SpecXMP, Payload: []byte("<bad>xml</broken>")},
		{Spec: common.SpecXMP, Payload: []byte("<another><bad>")},
	}

	dirs, err := parser.Parse(blocks)
	if err == nil {
		t.Error("Expected error when all blocks fail to parse")
	}
	if len(dirs) != 0 {
		t.Errorf("Expected no directories when all parsing fails, got %d", len(dirs))
	}
}

func TestParser_Robustness(t *testing.T) {
	parser := New()

	blocks := []common.RawBlock{
		{Spec: common.SpecXMP, Payload: []byte("<bad>xml</broken>")}, // Malformed
		{Spec: common.SpecXMP, Payload: []byte(`
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/" dc:valid="true"/>
 </rdf:RDF>
</x:xmpmeta>`)}, // Valid
	}

	dirs, err := parser.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse failed even with one valid block: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("Expected 1 directory, got %d", len(dirs))
	}

	if _, ok := dirs[0].Tags["XMP-dc:valid"]; !ok {
		t.Errorf("Expected valid tag to be parsed")
	}
}

func TestStripXPacket(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Full wrapper",
			input: `<?xpacket begin="?" id="W5M0MpCehiHzreSzNTczkc9d"?><root>data</root><?xpacket end="w"?>`,
			want:  `<root>data</root>`,
		},
		{
			name:  "No wrapper",
			input: `<root>data</root>`,
			want:  `<root>data</root>`,
		},
		{
			name:  "Only begin",
			input: `<?xpacket begin="?"?><data/>`,
			want:  `<data/>`,
		},
		{
			name:  "Only end",
			input: `<data/><?xpacket end="w"?>`,
			want:  `<data/>`,
		},
		{
			name:  "With whitespace",
			input: `<?xpacket begin="?"?>  <data/>  <?xpacket end="w"?>`,
			want:  `<data/>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(stripXPacket([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("stripXPacket() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSpec tests the Spec() method
func TestSpec(t *testing.T) {
	parser := New()
	if got := parser.Spec(); got != common.SpecXMP {
		t.Errorf("Spec() = %v, want %v", got, common.SpecXMP)
	}
}

// TestParse_EdgeCases tests Parse with various edge cases
func TestParse_EdgeCases(t *testing.T) {
	parser := New()

	t.Run("Empty blocks", func(t *testing.T) {
		dirs, err := parser.Parse([]common.RawBlock{})
		if err != nil {
			t.Errorf("Parse(empty) error: %v", err)
		}
		if len(dirs) != 0 {
			t.Errorf("Parse(empty) returned %d dirs, want 0", len(dirs))
		}
	})

	t.Run("Non-XMP blocks", func(t *testing.T) {
		dirs, err := parser.Parse([]common.RawBlock{
			{Spec: common.SpecEXIF, Payload: []byte("not xmp")},
		})
		if err != nil {
			t.Errorf("Parse(non-xmp) error: %v", err)
		}
		if len(dirs) != 0 {
			t.Errorf("Parse(non-xmp) returned %d dirs, want 0", len(dirs))
		}
	})

	t.Run("Empty payload", func(t *testing.T) {
		dirs, err := parser.Parse([]common.RawBlock{
			{Spec: common.SpecXMP, Payload: []byte("")},
		})
		if err != nil {
			t.Errorf("Parse(empty payload) error: %v", err)
		}
		if len(dirs) != 0 {
			t.Errorf("Parse(empty payload) returned %d dirs, want 0", len(dirs))
		}
	})

	t.Run("Whitespace only", func(t *testing.T) {
		dirs, err := parser.Parse([]common.RawBlock{
			{Spec: common.SpecXMP, Payload: []byte("   \n\t  ")},
		})
		if err != nil {
			t.Errorf("Parse(whitespace) error: %v", err)
		}
		if len(dirs) != 0 {
			t.Errorf("Parse(whitespace) returned %d dirs, want 0", len(dirs))
		}
	})

	t.Run("Multiple blocks", func(t *testing.T) {
		blocks := []common.RawBlock{
			{
				Spec: common.SpecXMP,
				Payload: []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
 <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/" dc:format="jpeg"/>
</rdf:RDF></x:xmpmeta>`),
			},
			{
				Spec: common.SpecXMP,
				Payload: []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
 <rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/" xmp:Rating="3"/>
</rdf:RDF></x:xmpmeta>`),
			},
		}

		dirs, err := parser.Parse(blocks)
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}

		if len(dirs) != 1 {
			t.Fatalf("Expected 1 directory, got %d", len(dirs))
		}

		// Should have both properties
		if _, ok := dirs[0].Tags["XMP-dc:format"]; !ok {
			t.Error("Missing dc:format from first block")
		}
		if _, ok := dirs[0].Tags["XMP-xmp:Rating"]; !ok {
			t.Error("Missing xmp:Rating from second block")
		}
	})
}
func TestParsePacket_EdgeCases(t *testing.T) {
	t.Run("Nested array in struct field", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:container>
    <rdf:Description>
     <ns:items>
      <rdf:Bag>
       <rdf:li>item1</rdf:li>
       <rdf:li>item2</rdf:li>
      </rdf:Bag>
     </ns:items>
    </rdf:Description>
   </ns:container>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}

		if len(nodeMap) == 0 {
			t.Error("Expected parsed data")
		}
	})

	t.Run("Array inside list item", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:outer>
    <rdf:Seq>
     <rdf:li>
      <rdf:Bag>
       <rdf:li>nested1</rdf:li>
      </rdf:Bag>
     </rdf:li>
    </rdf:Seq>
   </ns:outer>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Struct field with nested struct", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:outer>
    <rdf:Description>
     <ns:inner>
      <rdf:Description ns:field="value"/>
     </ns:inner>
    </rdf:Description>
   </ns:outer>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("CharData in various contexts", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:title>Simple Text</dc:title>
   <dc:subject>
    <rdf:Bag>
     <rdf:li>keyword</rdf:li>
    </rdf:Bag>
   </dc:subject>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}

		// Verify dc:title has simple text
		key := PropertyKey{URI: "http://purl.org/dc/elements/1.1/", Local: "title"}
		if val, ok := nodeMap[key]; !ok || len(val) == 0 || val[0].Scalar != "Simple Text" {
			t.Errorf("dc:title not parsed correctly")
		}
	})

	t.Run("Multiple Description blocks at RDF level", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
          xmlns:dc="http://purl.org/dc/elements/1.1/"
          xmlns:xmp="http://ns.adobe.com/xap/1.0/">
  <rdf:Description rdf:about="" dc:format="jpeg"/>
  <rdf:Description rdf:about="" xmp:Rating="4"/>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}

		if len(nodeMap) != 2 {
			t.Errorf("Expected 2 properties from multiple Description blocks")
		}
	})

	t.Run("Empty elements", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:empty></dc:empty>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}

		key := PropertyKey{URI: "http://purl.org/dc/elements/1.1/", Local: "empty"}
		if val, ok := nodeMap[key]; !ok || len(val) == 0 || val[0].Scalar != "" {
			t.Errorf("Empty element not parsed correctly")
		}
	})

	t.Run("Malformed XML", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta><broken>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err == nil {
			t.Error("Expected error for malformed XML")
		}
	})

	t.Run("Non-RDF element under root", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <x:other>content</x:other>
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/" dc:test="value"/>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Non-Description element under RDF", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Other>content</rdf:Other>
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/" dc:test="value"/>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("RDF element directly in property", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:prop>
    <rdf:Description dc:inner="value"/>
   </dc:prop>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Unknown element in array context", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:array>
    <rdf:Bag>
     <rdf:unknown>should fallback to root</rdf:unknown>
    </rdf:Bag>
   </dc:array>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("LI with Description and attributes", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
          xmlns:stEvt="http://ns.adobe.com/xap/1.0/sType/ResourceEvent#">
  <rdf:Description rdf:about="" xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/">
   <xmpMM:History>
    <rdf:Seq>
     <rdf:li>
      <rdf:Description stEvt:action="saved" stEvt:when="2023-01-01"/>
     </rdf:li>
    </rdf:Seq>
   </xmpMM:History>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("LI with nested property element", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:items>
    <rdf:Seq>
     <rdf:li>
      <ns:field>value</ns:field>
     </rdf:li>
    </rdf:Seq>
   </ns:items>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Struct field without propLocal", func(t *testing.T) {
		// This tests the case where STRUCT_FIELD has propLocal == ""
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:outer>
    <rdf:Description>
     <rdf:Description ns:inner="value"/>
    </rdf:Description>
   </ns:outer>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Array item transfer to property", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:subject>
    <rdf:Bag>
     <rdf:li>item1</rdf:li>
     <rdf:li>item2</rdf:li>
     <rdf:li>item3</rdf:li>
    </rdf:Bag>
   </dc:subject>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}

		key := PropertyKey{URI: "http://purl.org/dc/elements/1.1/", Local: "subject"}
		if val, ok := nodeMap[key]; !ok || len(val) == 0 || len(val[0].Items) != 3 {
			t.Errorf("Array items not properly transferred to property")
		}
	})

	t.Run("Property with non-RDF non-Description child", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:container>
    <ns:nested>value</ns:nested>
   </ns:container>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Property with attributes becomes struct", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:prop ns:attr1="val1" ns:attr2="val2">text</ns:prop>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}

		key := PropertyKey{URI: "http://example.com/ns/", Local: "prop"}
		if val, ok := nodeMap[key]; !ok || len(val) == 0 || val[0].Kind != KindStruct {
			t.Errorf("Property with attributes should be KindStruct")
		}
	})

	t.Run("LI with attributes becomes struct", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:items>
    <rdf:Seq>
     <rdf:li ns:attr="value">text</rdf:li>
    </rdf:Seq>
   </ns:items>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Struct field with attributes", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:outer>
    <rdf:Description>
     <ns:inner ns:attr="attrval">text</ns:inner>
    </rdf:Description>
   </ns:outer>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Struct field with array child", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:outer>
    <rdf:Description>
     <ns:items>
      <rdf:Bag>
       <rdf:li>item</rdf:li>
      </rdf:Bag>
     </ns:items>
    </rdf:Description>
   </ns:outer>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Struct field parent is LI", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:array>
    <rdf:Seq>
     <rdf:li>
      <rdf:Description>
       <ns:field>val</ns:field>
      </rdf:Description>
     </rdf:li>
    </rdf:Seq>
   </ns:array>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Deeply nested struct fields", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:level1>
    <rdf:Description>
     <ns:level2>
      <rdf:Description>
       <ns:level3>value</ns:level3>
      </rdf:Description>
     </ns:level2>
    </rdf:Description>
   </ns:level1>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("LI containing Bag/Seq/Alt", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:outer>
    <rdf:Seq>
     <rdf:li>
      <rdf:Bag>
       <rdf:li>nested</rdf:li>
      </rdf:Bag>
     </rdf:li>
    </rdf:Seq>
   </ns:outer>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Array end with non-Property parent", func(t *testing.T) {
		// Test CTX_ARRAY ending when parent is not CTX_PROPERTY
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:items>
    <rdf:Seq>
     <rdf:li>item1</rdf:li>
     <rdf:li>item2</rdf:li>
    </rdf:Seq>
   </ns:items>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("LI with non-Array parent", func(t *testing.T) {
		// This shouldn't normally happen in well-formed XMP but tests the else branch
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:subject>
    <rdf:Bag>
     <rdf:li>keyword1</rdf:li>
     <rdf:li>keyword2</rdf:li>
    </rdf:Bag>
   </dc:subject>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Struct field parent is STRUCT_FIELD", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:outer>
    <rdf:Description>
     <ns:middle>
      <rdf:Description>
       <ns:inner>value</ns:inner>
      </rdf:Description>
     </ns:middle>
    </rdf:Description>
   </ns:outer>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Parse type resource in LI", func(t *testing.T) {
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
          xmlns:stEvt="http://ns.adobe.com/xap/1.0/sType/ResourceEvent#">
  <rdf:Description rdf:about="" xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/">
   <xmpMM:History>
    <rdf:Seq>
     <rdf:li rdf:parseType="Resource" stEvt:action="created" stEvt:when="2023"/>
    </rdf:Seq>
   </xmpMM:History>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("RDF non-Description under Description (line 83-85)", func(t *testing.T) {
		// Test the else branch at line 83-85: RDF element under Description that's not Description
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:valid>test</dc:valid>
   <rdf:SomeWeirdElement>ignored</rdf:SomeWeirdElement>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("Property struct field with attributes (line 108-111)", func(t *testing.T) {
		// Test line 108-111: struct field with attributes under Property
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:container>
    <ns:nested ns:attr1="val1" ns:attr2="val2">text</ns:nested>
   </ns:container>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})

	t.Run("LI struct field with attributes (line 145-148)", func(t *testing.T) {
		// Test line 145-148: struct field with attributes under LI
		payload := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:array>
    <rdf:Seq>
     <rdf:li>
      <ns:field ns:attr1="val1" ns:attr2="val2">text</ns:field>
     </rdf:li>
    </rdf:Seq>
   </ns:array>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		p := New()
		err := p.parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})
}
