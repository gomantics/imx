package id3

import (
	"bytes"
	"testing"
)

func FuzzID3Parse(f *testing.F) {
	// Seed with valid ID3v2 headers
	f.Add([]byte{'I', 'D', '3', 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	f.Add([]byte{'I', 'D', '3', 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// Minimal tag with frame
	var buf bytes.Buffer
	buf.Write([]byte{'I', 'D', '3', 0x03, 0x00, 0x00})
	buf.Write(encodeSynchsafeInt(0x15))
	buf.Write([]byte{'T', 'I', 'T', '2'})
	buf.Write([]byte{0x00, 0x00, 0x00, 0x0B, 0x00, 0x00, 0x00})
	buf.Write([]byte("Test\x00"))
	f.Add(buf.Bytes())

	f.Fuzz(func(t *testing.T, data []byte) {
		p := New()
		r := bytes.NewReader(data)

		// Parser should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Parser panicked: %v", r)
			}
		}()

		// Just ensure it doesn't crash
		p.Parse(r)
	})
}
