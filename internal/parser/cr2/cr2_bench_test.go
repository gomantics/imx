package cr2

import (
	"bytes"
	"os"
	"testing"
)

// BenchmarkCR2Parse benchmarks parsing Canon CR2 (Canon Raw 2) files.
func BenchmarkCR2Parse(b *testing.B) {
	// Read test file into memory
	data, err := os.ReadFile("../../../testdata/cr2/sample1.cr2")
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
