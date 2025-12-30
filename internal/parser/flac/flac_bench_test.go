package flac

import (
	"bytes"
	"os"
	"testing"
)

// BenchmarkFLACParse benchmarks parsing FLAC (Free Lossless Audio Codec) files.
func BenchmarkFLACParse(b *testing.B) {
	data, err := os.ReadFile("../../../testdata/flac/sample3_hires.flac")
	if err != nil {
		b.Skipf("Test file not found: %v", err)
	}

	p := New()
	r := bytes.NewReader(data) // Create reader once outside loop

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(r)
	}
}
