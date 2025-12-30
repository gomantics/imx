package icc

import (
	"bytes"
	"testing"
)

// FuzzICCParse tests the ICC parser with random inputs to catch panics and edge cases.
func FuzzICCParse(f *testing.F) {
	// Add minimal ICC profile header with signature
	minimalICC := make([]byte, 128)
	copy(minimalICC[36:40], []byte("acsp")) // ICC signature at offset 36
	f.Add(minimalICC)

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
