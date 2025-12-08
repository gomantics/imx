package xmp

import (
	"testing"

	"github.com/gomantics/imx/internal/common"
)

// BenchmarkXMPParse benchmarks XMP parsing with typical Adobe metadata
func BenchmarkXMPParse(b *testing.B) {
	// Realistic XMP packet with common metadata fields
	xmpData := []byte(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description rdf:about=""
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:xmp="http://ns.adobe.com/xap/1.0/"
    xmlns:photoshop="http://ns.adobe.com/photoshop/1.0/">
  <dc:title>
    <rdf:Alt>
      <rdf:li xml:lang="x-default">Test Image Title</rdf:li>
    </rdf:Alt>
  </dc:title>
  <dc:creator>
    <rdf:Seq>
      <rdf:li>Test Photographer</rdf:li>
    </rdf:Seq>
  </dc:creator>
  <dc:description>
    <rdf:Alt>
      <rdf:li xml:lang="x-default">Test description</rdf:li>
    </rdf:Alt>
  </dc:description>
  <dc:subject>
    <rdf:Bag>
      <rdf:li>keyword1</rdf:li>
      <rdf:li>keyword2</rdf:li>
      <rdf:li>keyword3</rdf:li>
    </rdf:Bag>
  </dc:subject>
  <xmp:CreateDate>2024-01-01T12:00:00</xmp:CreateDate>
  <xmp:ModifyDate>2024-01-02T15:30:00</xmp:ModifyDate>
  <xmp:Rating>5</xmp:Rating>
  <photoshop:City>Test City</photoshop:City>
  <photoshop:Country>Test Country</photoshop:Country>
</rdf:Description>
</rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`)

	block := common.RawBlock{
		Spec:    common.SpecXMP,
		Payload: xmpData,
	}

	p := New()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse([]common.RawBlock{block})
	}
}
