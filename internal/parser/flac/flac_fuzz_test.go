package flac

import (
	"bytes"
	"testing"
)

// FuzzFLACParse tests the FLAC parser with random inputs to catch panics and edge cases.
func FuzzFLACParse(f *testing.F) {
	// Seed 1: Minimal valid FLAC with empty STREAMINFO (exercises basic parsing)
	f.Add([]byte{
		0x66, 0x4C, 0x61, 0x43, // "fLaC" magic
		0x80, 0x00, 0x00, 0x22, // Last block, type 0 (STREAMINFO), length 34
		// 34 bytes of STREAMINFO data (all zeros for simplicity)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
	})

	// Seed 2: FLAC with STREAMINFO containing actual values (exercises field parsing)
	f.Add([]byte{
		0x66, 0x4C, 0x61, 0x43, // "fLaC"
		0x80, 0x00, 0x00, 0x22, // Last block, STREAMINFO, 34 bytes
		0x10, 0x00, // Min block size = 4096
		0x10, 0x00, // Max block size = 4096
		0x00, 0x00, 0x00, // Min frame size = 0
		0x00, 0x00, 0x00, // Max frame size = 0
		0x0A, 0xC4, 0x42, // Sample rate 44100Hz, channels 2, bits 16
		0xF0, 0x00, 0x00, 0x00, 0x00, // Total samples
		// MD5 signature (16 bytes)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})

	// Seed 3: FLAC with multiple blocks (STREAMINFO + VORBIS_COMMENT)
	var buf3 bytes.Buffer
	buf3.Write([]byte{0x66, 0x4C, 0x61, 0x43}) // "fLaC"
	buf3.Write([]byte{0x00, 0x00, 0x00, 0x22}) // Not last, STREAMINFO, 34 bytes
	buf3.Write(make([]byte, 34))               // STREAMINFO data
	buf3.Write([]byte{0x84, 0x00, 0x00, 0x0F}) // Last block, VORBIS_COMMENT, 15 bytes
	buf3.Write([]byte{0x04, 0x00, 0x00, 0x00}) // Vendor length = 4
	buf3.Write([]byte("Test"))                 // Vendor string
	buf3.Write([]byte{0x01, 0x00, 0x00, 0x00}) // 1 comment
	buf3.Write([]byte{0x03, 0x00, 0x00, 0x00}) // Comment length = 3
	buf3.Write([]byte("A=B"))                  // Comment
	f.Add(buf3.Bytes())

	// Seed 4: FLAC with PADDING block (exercises padding parsing)
	var buf4 bytes.Buffer
	buf4.Write([]byte{0x66, 0x4C, 0x61, 0x43}) // "fLaC"
	buf4.Write([]byte{0x00, 0x00, 0x00, 0x22}) // Not last, STREAMINFO, 34 bytes
	buf4.Write(make([]byte, 34))               // STREAMINFO data
	buf4.Write([]byte{0x81, 0x00, 0x00, 0x10}) // Last block, PADDING, 16 bytes
	buf4.Write(make([]byte, 16))               // Padding data
	f.Add(buf4.Bytes())

	// Seed 5: FLAC with PICTURE block (exercises picture parsing)
	var buf5 bytes.Buffer
	buf5.Write([]byte{0x66, 0x4C, 0x61, 0x43}) // "fLaC"
	buf5.Write([]byte{0x00, 0x00, 0x00, 0x22}) // Not last, STREAMINFO, 34 bytes
	buf5.Write(make([]byte, 34))               // STREAMINFO data
	buf5.Write([]byte{0x86, 0x00, 0x00, 0x20}) // Last block, PICTURE, 32 bytes
	buf5.Write([]byte{0x00, 0x00, 0x00, 0x03}) // Picture type = 3 (Cover front)
	buf5.Write([]byte{0x00, 0x00, 0x00, 0x09}) // MIME length = 9
	buf5.Write([]byte("image/png"))            // MIME type
	buf5.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Description length = 0
	buf5.Write([]byte{0x00, 0x00, 0x00, 0x64}) // Width = 100
	buf5.Write([]byte{0x00, 0x00, 0x00, 0x64}) // Height = 100
	buf5.Write([]byte{0x00, 0x00, 0x00, 0x18}) // Depth = 24
	buf5.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Colors = 0
	buf5.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Picture data length = 0
	f.Add(buf5.Bytes())

	// Seed 6: FLAC with APPLICATION block (exercises application parsing)
	var buf6 bytes.Buffer
	buf6.Write([]byte{0x66, 0x4C, 0x61, 0x43}) // "fLaC"
	buf6.Write([]byte{0x00, 0x00, 0x00, 0x22}) // Not last, STREAMINFO, 34 bytes
	buf6.Write(make([]byte, 34))               // STREAMINFO data
	buf6.Write([]byte{0x82, 0x00, 0x00, 0x08}) // Last block, APPLICATION, 8 bytes
	buf6.Write([]byte("TEST"))                 // Application ID
	buf6.Write([]byte{0x01, 0x02, 0x03, 0x04}) // Application data
	f.Add(buf6.Bytes())

	// Seed 7: FLAC with SEEKTABLE block (exercises seektable parsing)
	var buf7 bytes.Buffer
	buf7.Write([]byte{0x66, 0x4C, 0x61, 0x43}) // "fLaC"
	buf7.Write([]byte{0x00, 0x00, 0x00, 0x22}) // Not last, STREAMINFO, 34 bytes
	buf7.Write(make([]byte, 34))               // STREAMINFO data
	buf7.Write([]byte{0x83, 0x00, 0x00, 0x12}) // Last block, SEEKTABLE, 18 bytes (1 point)
	// Seek point: sample number (8 bytes) + offset (8 bytes) + samples (2 bytes)
	buf7.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) // Sample
	buf7.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) // Offset
	buf7.Write([]byte{0x10, 0x00})                                     // Samples
	f.Add(buf7.Bytes())

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
