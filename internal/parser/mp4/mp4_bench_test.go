package mp4

import (
	"bytes"
	"os"
	"testing"
)

// BenchmarkMP4Parse benchmarks parsing MP4 (MPEG-4 Part 14) files.
func BenchmarkMP4Parse(b *testing.B) {
	data, err := os.ReadFile("../../../testdata/m4a/sample4_itunes.m4a")
	if err != nil {
		b.Skipf("sample M4A not found: %v", err)
	}
	p := New()
	r := bytes.NewReader(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Parse(r)
	}
}
