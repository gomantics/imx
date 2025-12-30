package mp4

import (
	"bytes"
	"os"
	"testing"
)

func FuzzMP4Parse(f *testing.F) {
	// Seed with valid ftyp boxes
	f.Add(createFtypBox("M4A ", 0, []string{}))
	f.Add(createFtypBox("mp41", 0, []string{"isom"}))
	f.Add(createFtypBox("isom", 0, []string{"mp42"}))
	if data, err := os.ReadFile("../../../testdata/m4a/sample4_itunes.m4a"); err == nil {
		f.Add(data)
	}

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
