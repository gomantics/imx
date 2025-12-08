package xmp

import (
	"testing"

	"github.com/gomantics/imx/internal/common"
)

// FuzzXMPParse tests the XMP parser with random/malformed XML.
// XMP uses RDF/XML which has complex nested structures and namespaces.
func FuzzXMPParse(f *testing.F) {
	// Seed with minimal valid XMP
	minimalXMP := []byte(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""/>
 </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`)
	f.Add(minimalXMP)

	// Seed with XMP containing a simple property
	simpleProperty := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:title>Test</dc:title>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
	f.Add(simpleProperty)

	// Seed with XMP containing an array
	arrayProperty := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
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
	f.Add(arrayProperty)

	// Seed with XMP containing a struct
	structProperty := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:ns="http://example.com/ns/">
   <ns:prop>
    <rdf:Description ns:field="value"/>
   </ns:prop>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
	f.Add(structProperty)

	f.Fuzz(func(t *testing.T, data []byte) {
		block := common.RawBlock{
			Spec:    common.SpecXMP,
			Payload: data,
			Origin:  "APP1",
		}

		parser := New()
		_, _ = parser.Parse([]common.RawBlock{block})
	})
}

// FuzzXMPParsePacket tests the XMP packet parser directly with random XML data.
// This tests the low-level XML parsing and namespace handling.
func FuzzXMPParsePacket(f *testing.F) {
	// Seed with minimal valid XMP packet
	f.Add([]byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description rdf:about=""/></rdf:RDF></x:xmpmeta>`))

	// Seed with simple property
	f.Add([]byte(`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:dc="http://purl.org/dc/elements/1.1/"><rdf:Description rdf:about=""><dc:title>Test</dc:title></rdf:Description></rdf:RDF>`))

	f.Fuzz(func(t *testing.T, data []byte) {
		parser := New()
		nodeMap := make(NodeMap)
		namespaces := make(map[string]string)
		_ = parser.parsePacket(data, nodeMap, namespaces)
	})
}
