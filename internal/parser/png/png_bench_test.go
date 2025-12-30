package png

import (
	"bytes"
	"testing"
)

// BenchmarkPNGParse benchmarks parsing PNG (Portable Network Graphics) files.
func BenchmarkPNGParse(b *testing.B) {
	// Create test PNG data
	data := createMinimalPNG()

	p := New()
	r := bytes.NewReader(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(r)
		r.Seek(0, 0) // Reset reader for next iteration
	}
}
