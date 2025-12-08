package jpeg

import (
	"bufio"
	"bytes"
	"testing"
)

// FuzzJPEGParse tests the JPEG parser with random/malformed data.
// This ensures the parser handles corrupt JPEGs gracefully without panicking.
func FuzzJPEGParse(f *testing.F) {
	// Seed with valid JPEG header (SOI + APP0 JFIF)
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00})

	// Seed with minimal valid JPEG (SOI + EOI only)
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xD9})

	// Seed with JPEG containing EXIF
	exifJPEG := []byte{0xFF, 0xD8, 0xFF, 0xE1, 0x00, 0x10}
	exifJPEG = append(exifJPEG, []byte{'E', 'x', 'i', 'f', 0x00, 0x00, 0x4D, 0x4D, 0x00, 0x2A}...)
	f.Add(exifJPEG)

	// Seed with JPEG containing ICC profile
	iccJPEG := []byte{0xFF, 0xD8, 0xFF, 0xE2, 0x00, 0x14}
	iccJPEG = append(iccJPEG, []byte{'I', 'C', 'C', '_', 'P', 'R', 'O', 'F', 'I', 'L', 'E', 0x00}...)
	f.Add(iccJPEG)

	f.Fuzz(func(t *testing.T, data []byte) {
		parser := New()
		reader := bufio.NewReader(bytes.NewReader(data))
		_, _ = parser.Parse(reader)
		// We don't check errors - just ensure no panics
	})
}

// FuzzJPEGDetect tests the JPEG detection logic with random data.
func FuzzJPEGDetect(f *testing.F) {
	// Valid JPEG magic bytes
	f.Add([]byte{0xFF, 0xD8})
	// Invalid magic bytes
	f.Add([]byte{0xFF, 0x00})
	f.Add([]byte{0x00, 0xD8})
	// Empty data
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		parser := New()
		_ = parser.Detect(data)
	})
}
