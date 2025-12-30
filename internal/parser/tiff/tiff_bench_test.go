package tiff

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// BenchmarkTIFFParse benchmarks TIFF/EXIF parsing with typical camera data.
func BenchmarkTIFFParse(b *testing.B) {
	// Create a realistic TIFF structure with typical camera metadata
	data := buildBenchmarkTIFF()
	reader := bytes.NewReader(data)

	p := New()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(reader)
	}
}

// buildBenchmarkTIFF creates a TIFF with typical camera metadata entries.
func buildBenchmarkTIFF() []byte {
	buf := new(bytes.Buffer)
	order := binary.LittleEndian

	// Header (8 bytes)
	buf.WriteString("II")                // Little endian
	binary.Write(buf, order, uint16(42)) // TIFF magic
	binary.Write(buf, order, uint32(8))  // Offset to first IFD

	// IFD0 starts at offset 8
	numEntries := uint16(10)
	binary.Write(buf, order, numEntries)

	// Calculate data offset (after IFD entries and next IFD pointer)
	// IFD: 2 (count) + 10*12 (entries) + 4 (next IFD) = 126
	dataOffset := uint32(8 + 2 + 10*12 + 4)

	// Entry 1: Make (ASCII string)
	writeIFDEntry(buf, order, 0x010F, TypeASCII, 6, dataOffset)
	makeStr := []byte("Canon\x00")
	dataOffset += 6

	// Entry 2: Model (ASCII string)
	writeIFDEntry(buf, order, 0x0110, TypeASCII, 14, dataOffset)
	modelStr := []byte("EOS 5D Mark IV\x00")
	dataOffset += 14

	// Entry 3: Orientation (SHORT, inline)
	writeIFDEntry(buf, order, 0x0112, TypeShort, 1, 1)

	// Entry 4: XResolution (RATIONAL)
	writeIFDEntry(buf, order, 0x011A, TypeRational, 1, dataOffset)
	dataOffset += 8

	// Entry 5: YResolution (RATIONAL)
	writeIFDEntry(buf, order, 0x011B, TypeRational, 1, dataOffset)
	dataOffset += 8

	// Entry 6: ResolutionUnit (SHORT, inline)
	writeIFDEntry(buf, order, 0x0128, TypeShort, 1, 2) // inches

	// Entry 7: Software (ASCII)
	writeIFDEntry(buf, order, 0x0131, TypeASCII, 12, dataOffset)
	softwareStr := []byte("Adobe PS CC\x00")
	dataOffset += 12

	// Entry 8: DateTime (ASCII)
	writeIFDEntry(buf, order, 0x0132, TypeASCII, 20, dataOffset)
	dateTimeStr := []byte("2024:01:15 10:30:00\x00")
	dataOffset += 20

	// Entry 9: Artist (ASCII)
	writeIFDEntry(buf, order, 0x013B, TypeASCII, 14, dataOffset)
	artistStr := []byte("Photographer\x00\x00")
	dataOffset += 14

	// Entry 10: Copyright (ASCII)
	writeIFDEntry(buf, order, 0x8298, TypeASCII, 16, dataOffset)
	copyrightStr := []byte("(c) 2024 Author\x00")

	// Next IFD pointer (0 = no more IFDs)
	binary.Write(buf, order, uint32(0))

	// Write data section
	buf.Write(makeStr)
	buf.Write(modelStr)

	// XResolution: 300/1
	binary.Write(buf, order, uint32(300))
	binary.Write(buf, order, uint32(1))

	// YResolution: 300/1
	binary.Write(buf, order, uint32(300))
	binary.Write(buf, order, uint32(1))

	buf.Write(softwareStr)
	buf.Write(dateTimeStr)
	buf.Write(artistStr)
	buf.Write(copyrightStr)

	return buf.Bytes()
}
