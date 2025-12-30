package xmp

import (
	"bytes"
	"testing"
)

// FuzzXMPParse tests the XMP parser with random inputs to catch panics and edge cases.
func FuzzXMPParse(f *testing.F) {
	// Add minimal XMP packet
	f.Add([]byte(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?><x:xmpmeta xmlns:x="adobe:ns:meta/"></x:xmpmeta><?xpacket end="w"?>`))

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
