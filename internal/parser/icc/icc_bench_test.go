package icc

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// BenchmarkICCParse benchmarks parsing ICC (International Color Consortium) profiles.
func BenchmarkICCParse(b *testing.B) {
	// Build a realistic ICC profile with typical tags
	data := buildBenchmarkICCProfile()
	p := New()
	r := bytes.NewReader(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(r)
	}
}

// buildBenchmarkICCProfile creates a valid ICC profile similar to Display P3.
func buildBenchmarkICCProfile() []byte {
	tagCount := 9
	tagTableSize := 4 + (tagCount * tagRecordSize)
	tagDataSize := tagCount * 100
	totalSize := headerSize + tagTableSize + tagDataSize

	data := make([]byte, totalSize)

	// Header
	binary.BigEndian.PutUint32(data[offsetProfileSize:], uint32(totalSize))
	copy(data[offsetCMMType:], "appl")
	data[offsetProfileVersion] = 0x04
	data[offsetProfileVersion+1] = 0x40
	copy(data[offsetProfileClass:], "mntr")
	copy(data[offsetColorSpace:], "RGB ")
	copy(data[offsetPCS:], "XYZ ")
	binary.BigEndian.PutUint16(data[offsetDateTime:], 2024)
	binary.BigEndian.PutUint16(data[offsetDateTime+2:], 1)
	binary.BigEndian.PutUint16(data[offsetDateTime+4:], 1)
	copy(data[offsetSignature:], iccSignature[:])
	copy(data[offsetPlatform:], "APPL")
	binary.BigEndian.PutUint32(data[offsetIlluminant:], 0x0000F6D6)
	binary.BigEndian.PutUint32(data[offsetIlluminant+4:], 0x00010000)
	binary.BigEndian.PutUint32(data[offsetIlluminant+8:], 0x0000D32D)
	copy(data[offsetProfileCreator:], "appl")

	// Tag table
	binary.BigEndian.PutUint32(data[offsetTagTableCount:], uint32(tagCount))

	tags := []string{"desc", "cprt", "wtpt", "rXYZ", "gXYZ", "bXYZ", "rTRC", "gTRC", "bTRC"}
	dataOffset := headerSize + tagTableSize

	for i, sig := range tags {
		entryOffset := offsetTagTableEntries + (i * tagRecordSize)
		copy(data[entryOffset:], sig)
		binary.BigEndian.PutUint32(data[entryOffset+4:], uint32(dataOffset))
		binary.BigEndian.PutUint32(data[entryOffset+8:], 100)

		// Tag data (desc type)
		copy(data[dataOffset:], typeDesc)
		binary.BigEndian.PutUint32(data[dataOffset+descCountOffset:], 20)
		copy(data[dataOffset+descStringOffset:], "Display P3          ")

		dataOffset += 100
	}

	return data
}
