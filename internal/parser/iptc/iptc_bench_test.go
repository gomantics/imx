package iptc

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// BenchmarkIPTCParse benchmarks IPTC parsing with typical metadata.
func BenchmarkIPTCParse(b *testing.B) {
	// Create realistic IPTC-IIM data wrapped in Photoshop 8BIM structure
	data := buildPhotoshopIPTC([]iptcDataset{
		{record: 2, id: 80, value: []byte("Test Byline")},
		{record: 2, id: 85, value: []byte("Test Byline Title")},
		{record: 2, id: 90, value: []byte("Test City")},
		{record: 2, id: 95, value: []byte("Test Province")},
		{record: 2, id: 101, value: []byte("USA")},
		{record: 2, id: 5, value: []byte("Test Title")},
		{record: 2, id: 120, value: []byte("Test caption describing the image content")},
		{record: 2, id: 25, value: []byte("keyword1")},
		{record: 2, id: 25, value: []byte("keyword2")},
		{record: 2, id: 25, value: []byte("keyword3")},
	})

	reader := bytes.NewReader(data)
	p := New()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(reader)
	}
}

// iptcDataset represents a single IPTC dataset entry.
type iptcDataset struct {
	record byte
	id     byte
	value  []byte
}

// buildPhotoshopIPTC creates Photoshop 8BIM structure containing IPTC-IIM data.
func buildPhotoshopIPTC(datasets []iptcDataset) []byte {
	// First build the IPTC-IIM data
	iptcData := buildIPTCIIM(datasets)

	buf := new(bytes.Buffer)

	// 8BIM signature
	buf.WriteString("8BIM")

	// Resource ID for IPTC-NAA record (0x0404)
	binary.Write(buf, binary.BigEndian, ResourceIPTC)

	// Pascal string (resource name) - empty
	buf.WriteByte(0) // length = 0
	buf.WriteByte(0) // padding to make it even

	// Resource data size
	binary.Write(buf, binary.BigEndian, uint32(len(iptcData)))

	// IPTC data
	buf.Write(iptcData)

	// Pad to even if necessary
	if len(iptcData)%2 != 0 {
		buf.WriteByte(0)
	}

	return buf.Bytes()
}

// buildIPTCIIM creates raw IPTC-IIM format data.
func buildIPTCIIM(datasets []iptcDataset) []byte {
	var data []byte

	for _, ds := range datasets {
		// Tag marker
		data = append(data, iptcTagMarker)
		// Record number
		data = append(data, ds.record)
		// Dataset ID
		data = append(data, ds.id)
		// Size (big-endian uint16)
		size := uint16(len(ds.value))
		data = append(data, byte(size>>8), byte(size))
		// Value
		data = append(data, ds.value...)
	}

	return data
}
