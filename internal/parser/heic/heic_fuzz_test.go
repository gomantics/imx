package heic

import (
	"bytes"
	"testing"
)

func FuzzHEICParse(f *testing.F) {
	// Seed with minimal HEIC structures
	seeds := [][]byte{
		// Minimal ftyp
		{0, 0, 0, 12, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c'},
		// ftyp + minimal meta
		{
			0, 0, 0, 12, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c',
			0, 0, 0, 20, 'm', 'e', 't', 'a', 0, 0, 0, 0,
			0, 0, 0, 8, 'h', 'd', 'l', 'r',
		},
		// Valid heif brand
		{0, 0, 0, 12, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'f'},
		// Valid mif1 brand
		{0, 0, 0, 12, 'f', 't', 'y', 'p', 'm', 'i', 'f', '1'},
		// Extended size box
		{0, 0, 0, 1, 'f', 't', 'y', 'p', 0, 0, 0, 0, 0, 0, 0, 20, 'h', 'e', 'i', 'c'},
		// Not HEIC (should be rejected quickly)
		{0xFF, 0xD8, 0xFF, 0xE0},
		// Empty
		{},
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	p := New()

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)

		// Should not panic
		_ = p.Detect(r)
		_, _ = p.Parse(r)
	})
}
