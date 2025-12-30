package gif

import (
	"bytes"
	"os"
	"testing"
)

// BenchmarkGIFParse benchmarks parsing GIF (Graphics Interchange Format) files.
func BenchmarkGIFParse(b *testing.B) {
	data, err := os.ReadFile("../../../testdata/gif/animated_art.gif")
	if err != nil {
		b.Skipf("Test file not found: %v", err)
	}

	p := New()
	r := bytes.NewReader(data)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(r)
	}
}
