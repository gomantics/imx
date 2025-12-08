package exif

import (
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/common"
)

// FuzzEXIFParse tests the EXIF parser with random/malformed TIFF data.
// EXIF uses TIFF format which has complex IFD structures that need robust parsing.
func FuzzEXIFParse(f *testing.F) {
	// Seed with minimal valid EXIF (big-endian)
	validExifBE := []byte{
		'M', 'M',       // Big-endian marker
		0x00, 0x2A,     // TIFF magic number
		0x00, 0x00, 0x00, 0x08, // IFD0 offset
	}
	f.Add(validExifBE)

	// Seed with little-endian EXIF
	validExifLE := []byte{
		'I', 'I',       // Little-endian marker
		0x2A, 0x00,     // TIFF magic number
		0x08, 0x00, 0x00, 0x00, // IFD0 offset
	}
	f.Add(validExifLE)

	f.Fuzz(func(t *testing.T, data []byte) {
		block := common.RawBlock{
			Spec:    common.SpecEXIF,
			Payload: data,
			Origin:  "APP1",
		}

		parser := New()
		_, _ = parser.Parse([]common.RawBlock{block})
	})
}

// FuzzEXIFParseIFD tests IFD (Image File Directory) parsing specifically.
// IFDs contain tag entries and offsets that can cause issues if malformed.
func FuzzEXIFParseIFD(f *testing.F) {
	// Seed with empty IFD (0 entries)
	f.Add([]byte{0x00, 0x00})

	// Seed with IFD containing 1 entry
	ifdData := make([]byte, 2+12)
	binary.BigEndian.PutUint16(ifdData[0:2], 1) // 1 entry
	f.Add(ifdData)

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 8 {
			return
		}

		// Create valid TIFF header + fuzzed IFD data
		fullData := make([]byte, 8+len(data))
		copy(fullData[0:2], []byte{'M', 'M'})           // Big-endian
		binary.BigEndian.PutUint16(fullData[2:4], 0x2A) // Magic
		binary.BigEndian.PutUint32(fullData[4:8], 8)    // IFD offset
		copy(fullData[8:], data)

		block := common.RawBlock{
			Spec:    common.SpecEXIF,
			Payload: fullData,
			Origin:  "APP1",
		}

		parser := New()
		_, _ = parser.Parse([]common.RawBlock{block})
	})
}
