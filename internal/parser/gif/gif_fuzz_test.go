package gif

import (
	"bytes"
	"testing"
)

// FuzzGIFParse tests the GIF parser with random inputs to find panics
func FuzzGIFParse(f *testing.F) {
	// Seed 1: Minimal valid GIF89a
	f.Add([]byte("GIF89a\x0A\x00\x0A\x00\x00\x00\x00\x3B"))

	// Seed 2: GIF with comment extension
	f.Add([]byte("GIF89a\x0A\x00\x0A\x00\x00\x00\x00\x21\xFE\x05Hello\x00\x3B"))

	// Seed 3: GIF with image descriptor
	f.Add([]byte("GIF89a\x0A\x00\x0A\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x0A\x00\x0A\x00\x00\x02\x00\x3B"))

	// Seed 4: GIF with graphic control extension
	f.Add([]byte("GIF89a\x0A\x00\x0A\x00\x00\x00\x00\x21\xF9\x04\x00\x00\x00\x00\x00\x3B"))

	// Seed 5: GIF with application extension (non-XMP)
	f.Add([]byte("GIF89a\x0A\x00\x0A\x00\x00\x00\x00\x21\xFF\x0BNETSCAPE2.0\x03\x01\x00\x00\x00\x3B"))

	// Seed 6: GIF with global color table
	f.Add([]byte("GIF89a\x0A\x00\x0A\x00\x80\x00\x00\xFF\x00\x00\x00\xFF\x00\x3B"))

	f.Fuzz(func(t *testing.T, data []byte) {
		p := New()
		r := bytes.NewReader(data)

		// Should not panic
		_, _ = p.Parse(r)
	})
}
