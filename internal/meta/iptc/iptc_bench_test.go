package iptc

import (
	"testing"

	"github.com/gomantics/imx/internal/common"
)

// BenchmarkIPTCParse benchmarks IPTC parsing with typical metadata
func BenchmarkIPTCParse(b *testing.B) {
	// Create realistic IPTC-IIM data with typical news/media metadata
	data := buildIPTCData([]dataset{
		{record: RecordApplication, id: 80, value: []byte("Test Byline")},
		{record: RecordApplication, id: 85, value: []byte("Test Byline Title")},
		{record: RecordApplication, id: 90, value: []byte("Test City")},
		{record: RecordApplication, id: 95, value: []byte("Test Province")},
		{record: RecordApplication, id: 101, value: []byte("USA")},
		{record: RecordApplication, id: 5, value: []byte("Test Title")},
		{record: RecordApplication, id: 120, value: []byte("Test caption")},
		{record: RecordApplication, id: 25, value: []byte("keyword1\x00keyword2\x00keyword3")},
	})

	block := common.RawBlock{
		Spec:    common.SpecIPTC,
		Payload: data,
	}

	p := New()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse([]common.RawBlock{block})
	}
}

// Helper types and functions for benchmarks

type dataset struct {
	record Record
	id     byte
	value  []byte
}

func buildIPTCData(datasets []dataset) []byte {
	var data []byte
	for _, ds := range datasets {
		// Marker
		data = append(data, iptcTagMarker)
		// Record
		data = append(data, byte(ds.record))
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

func buildIPTCDataWithExtendedSize(record Record, id byte, value []byte) []byte {
	var data []byte
	// Marker
	data = append(data, iptcTagMarker)
	// Record
	data = append(data, byte(record))
	// Dataset ID
	data = append(data, id)

	// Extended size (size > 32767)
	size := uint32(len(value))
	// Set high bit and specify 4-byte size
	data = append(data, 0x80, 0x04)
	// 4-byte size (big-endian)
	data = append(data, byte(size>>24), byte(size>>16), byte(size>>8), byte(size))
	// Value
	data = append(data, value...)

	return data
}
