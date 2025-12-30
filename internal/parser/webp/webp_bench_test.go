package webp

import (
	"bytes"
	"os"
	"testing"
)

// BenchmarkWebPParse benchmarks parsing WebP files.
func BenchmarkWebPParse(b *testing.B) {
	// Read test file into memory
	data, err := os.ReadFile("../../../testdata/webp/modern_webp.webp")
	if err != nil {
		b.Fatalf("failed to read test file: %v", err)
	}

	reader := bytes.NewReader(data)
	parser := New()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(reader)
	}
}
