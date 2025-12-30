package webp

import (
	"bytes"
	"os"
	"testing"
)

func FuzzWebPParse(f *testing.F) {
	// Add seed corpus with valid WebP structures

	// Minimal VP8X WebP
	var buf1 bytes.Buffer
	writeRIFFHeader(&buf1, 12)
	vp8x := createVP8X(100, 100, 0)
	writeChunk(&buf1, "VP8X", vp8x)
	f.Add(buf1.Bytes())

	// WebP with EXIF
	var buf2 bytes.Buffer
	writeRIFFHeader(&buf2, 12)
	vp8x2 := createVP8X(200, 200, 0x02) // EXIF flag
	writeChunk(&buf2, "VP8X", vp8x2)
	exifData := []byte("Exif\x00\x00MM\x00\x2A\x00\x00\x00\x08")
	writeChunk(&buf2, "EXIF", exifData)
	f.Add(buf2.Bytes())

	// WebP with VP8 (lossy)
	var buf3 bytes.Buffer
	writeRIFFHeader(&buf3, 12)
	vp8Data := []byte{
		0x00, 0x00, 0x00, // Frame tag
		0x9D, 0x01, 0x2A, // Start code
		0x64, 0x00, // Width (100)
		0x64, 0x00, // Height (100)
	}
	writeChunk(&buf3, "VP8 ", vp8Data)
	f.Add(buf3.Bytes())

	// WebP with VP8L (lossless)
	var buf4 bytes.Buffer
	writeRIFFHeader(&buf4, 12)
	vp8lData := []byte{
		0x2F,                   // Signature
		0x63, 0x00, 0x00, 0x00, // Width and height encoded
	}
	writeChunk(&buf4, "VP8L", vp8lData)
	f.Add(buf4.Bytes())

	// Real-world sample if available
	if data, err := os.ReadFile("../../../testdata/webp/modern_webp.webp"); err == nil {
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 12 {
			return // Skip too short inputs (minimum RIFF header)
		}

		p := New()
		r := bytes.NewReader(data)

		// Test Parse - should not panic or crash
		_, _ = p.Parse(r)

		// Test Detect - should not panic or crash
		_ = p.Detect(r)
	})
}
