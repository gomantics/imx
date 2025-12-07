package icc

import (
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/common"
)

// BenchmarkParser_Parse benchmarks ICC color profile parsing
func BenchmarkParser_Parse(b *testing.B) {
	// Create realistic ICC profile data with typical tags
	data := buildICCProfileWithTags(10)

	block := common.RawBlock{
		Spec:    common.SpecICC,
		Payload: data,
	}

	p := New()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse([]common.RawBlock{block})
	}
}

// Helper functions for building ICC test data

func buildICCProfileWithTags(tagCount int) []byte {
	// ICC profile minimum size is 132 bytes for header
	data := make([]byte, 132)

	// Profile size (first 4 bytes) - will update later
	headerSize := 132
	tagTableSize := 4 + (tagCount * 12) // tag count (4) + entries (12 bytes each)
	tagDataSize := tagCount * 100       // Approximate tag data size
	totalSize := headerSize + tagTableSize + tagDataSize
	binary.BigEndian.PutUint32(data[0:4], uint32(totalSize))

	// Preferred CMM type (bytes 4-7)
	copy(data[4:8], "appl")

	// Profile version (bytes 8-11)
	data[8] = 0x04  // Major version 4
	data[9] = 0x40  // Minor version 4.4
	data[10] = 0x00
	data[11] = 0x00

	// Profile/Device class (bytes 12-15)
	copy(data[12:16], "mntr") // Display device profile

	// Color space (bytes 16-19)
	copy(data[16:20], "RGB ")

	// PCS (bytes 20-23)
	copy(data[20:24], "XYZ ")

	// Date created (bytes 24-35) - all zeros for simplicity
	// Platform (bytes 40-43)
	copy(data[40:44], "APPL")

	// Rendering intent (bytes 64-67)
	binary.BigEndian.PutUint32(data[64:68], 0) // Perceptual

	// PCS illuminant (bytes 68-79) - D50
	binary.BigEndian.PutUint32(data[68:72], 0x0000F6D6) // X
	binary.BigEndian.PutUint32(data[72:76], 0x00010000) // Y
	binary.BigEndian.PutUint32(data[76:80], 0x0000D32D) // Z

	// Creator (bytes 80-83)
	copy(data[80:84], "appl")

	// Add tag table after header
	tagTable := make([]byte, tagTableSize)

	// Tag count
	binary.BigEndian.PutUint32(tagTable[0:4], uint32(tagCount))

	// Add tag entries
	dataOffset := headerSize + tagTableSize
	for i := 0; i < tagCount; i++ {
		entryOffset := 4 + (i * 12)

		// Tag signature
		sig := []byte("desc")
		if i == 1 {
			sig = []byte("cprt")
		} else if i == 2 {
			sig = []byte("wtpt")
		} else if i == 3 {
			sig = []byte("rXYZ")
		} else if i == 4 {
			sig = []byte("gXYZ")
		}
		copy(tagTable[entryOffset:entryOffset+4], sig)

		// Offset to tag data
		binary.BigEndian.PutUint32(tagTable[entryOffset+4:entryOffset+8], uint32(dataOffset))

		// Tag data size
		binary.BigEndian.PutUint32(tagTable[entryOffset+8:entryOffset+12], 100)

		dataOffset += 100
	}

	// Combine header + tag table + placeholder tag data
	result := make([]byte, totalSize)
	copy(result[0:132], data)
	copy(result[132:132+tagTableSize], tagTable)

	// Fill tag data section with valid-looking data
	for i := 0; i < tagCount; i++ {
		offset := headerSize + tagTableSize + (i * 100)
		// desc type signature
		copy(result[offset:offset+4], "desc")
		// Reserved
		binary.BigEndian.PutUint32(result[offset+4:offset+8], 0)
		// ASCII count
		binary.BigEndian.PutUint32(result[offset+8:offset+12], 20)
		// ASCII string
		copy(result[offset+12:offset+32], "Test Description    ")
	}

	return result
}
