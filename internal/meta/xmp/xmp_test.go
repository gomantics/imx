package xmp

import (
	"reflect"
	"testing"

	"github.com/gomantics/imx/internal/format"
	"github.com/gomantics/imx/internal/meta"
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
			block := format.RawBlock{
				Spec:    int(meta.SpecXMP),
				Payload: []byte(tt.payload),
			}
			dirs, err := parser.Parse([]format.RawBlock{block})
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if len(dirs) != 1 {
				t.Fatalf("Expected 1 directory, got %d", len(dirs))
			}
			dir := dirs[0]
			for id, wantVal := range tt.want {
				tag, ok := dir.Tags[meta.TagID(id)]
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

func TestParser_Robustness(t *testing.T) {
	parser := New()

	blocks := []format.RawBlock{
		{Spec: int(meta.SpecXMP), Payload: []byte("<bad>xml</broken>")}, // Malformed
		{Spec: int(meta.SpecXMP), Payload: []byte(`
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
	raw := `<?xpacket begin="?" id="W5M0MpCehiHzreSzNTczkc9d"?><root>data</root><?xpacket end="w"?>`
	got := stripXPacket([]byte(raw))
	want := `<root>data</root>`
	if string(got) != want {
		t.Errorf("stripXPacket failed. Got %q, want %q", got, want)
	}

	// Test without wrappers
	raw2 := `<root>data</root>`
	got2 := stripXPacket([]byte(raw2))
	if string(got2) != raw2 {
		t.Errorf("stripXPacket messed up unwrapped data. Got %q", got2)
	}
}
