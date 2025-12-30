package png

import (
	"bytes"
	"testing"
)

func FuzzPNGParse(f *testing.F) {
	// Add seed corpus with valid PNG structures
	f.Add(createMinimalPNG())

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 16 {
			return // Skip too short inputs
		}

		p := New()
		r := bytes.NewReader(data)
		_, _ = p.Parse(r) // Ignore errors, we're testing for crashes
	})
}
