package imx

import (
	"testing"
)

// BenchmarkMetadataFromFile benchmarks full metadata extraction from a file
func BenchmarkMetadataFromFile(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := MetadataFromFile("testdata/goldens/jpeg/canon_xmp.jpg")
		if err != nil {
			b.Fatalf("MetadataFromFile failed: %v", err)
		}
	}
}
