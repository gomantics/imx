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
