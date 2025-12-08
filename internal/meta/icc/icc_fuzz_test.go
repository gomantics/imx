package icc

import (
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/common"
)

// FuzzICCParse tests the ICC profile parser with random/malformed data.
// ICC profiles have a complex binary structure with headers and tag tables.
func FuzzICCParse(f *testing.F) {
	// Seed with minimal valid ICC profile (128-byte header, 0 tags)
	validICC := make([]byte, 128)
	binary.BigEndian.PutUint32(validICC[0:4], 128)          // Profile size
	binary.BigEndian.PutUint32(validICC[36:40], 0x61637370) // 'acsp' signature
	binary.BigEndian.PutUint32(validICC[128-4:128], 0)      // Tag count = 0
	f.Add(validICC)

	// Seed with profile containing 1 tag
	profileWithTag := make([]byte, 128+12+4)
	binary.BigEndian.PutUint32(profileWithTag[0:4], uint32(len(profileWithTag)))
	binary.BigEndian.PutUint32(profileWithTag[36:40], 0x61637370) // 'acsp'
	binary.BigEndian.PutUint32(profileWithTag[128:132], 1)        // 1 tag
	binary.BigEndian.PutUint32(profileWithTag[132:136], 0x64657363) // 'desc' signature
	binary.BigEndian.PutUint32(profileWithTag[136:140], 144)      // Offset
	binary.BigEndian.PutUint32(profileWithTag[140:144], 4)        // Size
	f.Add(profileWithTag)

	f.Fuzz(func(t *testing.T, data []byte) {
		block := common.RawBlock{
			Spec:    common.SpecICC,
			Payload: data,
			Origin:  "APP2",
		}

		parser := New()
		_, _ = parser.Parse([]common.RawBlock{block})
	})
}

// FuzzICCParseHeader tests ICC profile header parsing.
// Headers contain version info, color space, and other metadata.
func FuzzICCParseHeader(f *testing.F) {
	validHeader := make([]byte, 128)
	binary.BigEndian.PutUint32(validHeader[0:4], 128)
	binary.BigEndian.PutUint32(validHeader[36:40], 0x61637370) // 'acsp'
	f.Add(validHeader)

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseHeader(data)
	})
}

// FuzzICCParseTagTable tests tag table parsing.
// Tag tables contain offsets and sizes that must be validated.
func FuzzICCParseTagTable(f *testing.F) {
	tagTable := make([]byte, 4+12)
	binary.BigEndian.PutUint32(tagTable[0:4], 1)          // 1 tag
	binary.BigEndian.PutUint32(tagTable[4:8], 0x64657363) // 'desc'
	binary.BigEndian.PutUint32(tagTable[8:12], 132)       // Offset
	binary.BigEndian.PutUint32(tagTable[12:16], 10)       // Size
	f.Add(tagTable)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Create minimal valid profile with fuzzed tag table
		profile := make([]byte, 128+len(data))
		binary.BigEndian.PutUint32(profile[0:4], uint32(len(profile)))
		binary.BigEndian.PutUint32(profile[36:40], 0x61637370)
		copy(profile[128:], data)

		_, _ = parseTagTable(profile)
	})
}
