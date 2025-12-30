package iptc

import (
	"bytes"
	"testing"
)

// FuzzIPTCParse tests the IPTC parser with random inputs to catch panics and edge cases.
func FuzzIPTCParse(f *testing.F) {
	// Add minimal 8BIM signature
	f.Add([]byte("8BIM"))

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
