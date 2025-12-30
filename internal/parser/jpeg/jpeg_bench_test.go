package jpeg

import (
	"bytes"
	"os"
	"testing"
)

// BenchmarkJPEGParse benchmarks parsing a complete JPEG file with metadata.
func BenchmarkJPEGParse(b *testing.B) {
	// Read test file into memory
	data, err := os.ReadFile("../../../testdata/jpeg/olympus_micro43.jpg")
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
