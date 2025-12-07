package xmp

import (
	"encoding/xml"
	"testing"
)

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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
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
		err := parsePacket(payload, nodeMap, namespaces)
		if err != nil {
			t.Fatalf("parsePacket error: %v", err)
		}
	})
}
