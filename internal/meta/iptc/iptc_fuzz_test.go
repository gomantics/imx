package iptc

import (
	"testing"

	"github.com/gomantics/imx/internal/common"
)

// FuzzIPTCParse tests the IPTC parser with random/malformed IPTC-IIM data.
// IPTC-IIM uses a tag-length-value format with variable-length encoding.
func FuzzIPTCParse(f *testing.F) {
	// Seed with minimal valid IPTC dataset
	// Format: 0x1C (marker) + Record + Dataset + Size(2 bytes) + Data
	validIPTC := []byte{
		0x1C,       // Tag marker
		0x02,       // Record 2 (Application Record)
		0x05,       // Dataset 5 (Object Name)
		0x00, 0x04, // Size: 4 bytes
		'T', 'e', 's', 't',
	}
	f.Add(validIPTC)

	// Seed with multiple datasets
	multiDatasets := []byte{
		0x1C, 0x02, 0x05, 0x00, 0x02, 'A', 'B',
		0x1C, 0x02, 0x19, 0x00, 0x03, 'X', 'Y', 'Z',
	}
	f.Add(multiDatasets)

	// Seed with extended size format (sizes > 32767 bytes)
	extendedSize := []byte{
		0x1C,       // Tag marker
		0x02,       // Record 2
		0x05,       // Dataset 5
		0x80, 0x04, // Extended size: bit 15 set, length = 4
		0x00, 0x00, 0x01, 0x00, // Actual size: 256 bytes
	}
	f.Add(extendedSize)

	f.Fuzz(func(t *testing.T, data []byte) {
		block := common.RawBlock{
			Spec:    common.SpecIPTC,
			Payload: data,
			Origin:  "APP13",
		}

		parser := New()
		_, _ = parser.Parse([]common.RawBlock{block})
	})
}

// FuzzIPTCParseIPTCIIM tests the low-level IPTC-IIM parsing function directly.
func FuzzIPTCParseIPTCIIM(f *testing.F) {
	f.Add([]byte{0x1C, 0x02, 0x05, 0x00, 0x01, 'A'})
	f.Add([]byte{0x1C, 0x01, 0x5A, 0x00, 0x03, 'X', 'Y', 'Z'})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseIPTCIIM(data)
	})
}
