package id3

import (
	"bytes"
	"os"
	"testing"
)

// BenchmarkID3Parse benchmarks parsing ID3v2 metadata from audio files.
func BenchmarkID3Parse(b *testing.B) {
	// Use real MP3 sample with rich ID3 metadata
	data, err := os.ReadFile("../../../testdata/mp3/sample1_rich_metadata.mp3")
	if err != nil {
		b.Skipf("test MP3 not found: %v", err)
	}

	p := New()
	r := bytes.NewReader(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Parse(r)
	}
}
