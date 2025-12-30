package tiff

import (
	"bytes"
	"testing"
)

// FuzzTIFFParse tests the TIFF parser with random inputs to catch panics and edge cases.
func FuzzTIFFParse(f *testing.F) {
	// Add minimal valid TIFF headers
	f.Add([]byte{'I', 'I', 42, 0, 8, 0, 0, 0}) // Little-endian TIFF
	f.Add([]byte{'M', 'M', 0, 42, 0, 0, 0, 8}) // Big-endian TIFF

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
