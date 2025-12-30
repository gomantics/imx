package jpeg

import (
	"bytes"
	"testing"
)

// FuzzJPEGParse tests the JPEG parser with random inputs to catch panics and edge cases.
func FuzzJPEGParse(f *testing.F) {
	// Add minimal valid JPEG (SOI + EOI markers)
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xD9})

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked: %v", r)
			}
		}()

		reader := bytes.NewReader(data)
		parser := New()

		// Just call Parse - we don't care about errors, only panics
		_, _ = parser.Parse(reader)
	})
}
