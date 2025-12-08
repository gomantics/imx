package exif

import (
	"testing"

	"github.com/gomantics/imx/internal/common"
)

// BenchmarkEXIFParse benchmarks EXIF parsing with typical camera data
func BenchmarkEXIFParse(b *testing.B) {
	// Create a realistic TIFF structure with typical camera metadata
	data := buildTIFF(true, []ifdEntry{
		{tagID: 0x010F, dataType: 2, count: 6, valueOrOffset: []byte("Canon\x00\x00\x00")},        // Make
		{tagID: 0x0110, dataType: 2, count: 10, valueOrOffset: []byte("EOS 5D\x00\x00")},          // Model
		{tagID: 0x0112, dataType: 3, count: 1, valueOrOffset: []byte{0x00, 0x01, 0x00, 0x00}},     // Orientation
		{tagID: 0x011A, dataType: 5, count: 1, valueOrOffset: []byte{0, 0, 0, 72}},                // XResolution offset
		{tagID: 0x011B, dataType: 5, count: 1, valueOrOffset: []byte{0, 0, 0, 80}},                // YResolution offset
	})

	exifBlock := common.RawBlock{
		Spec:    common.SpecEXIF,
		Payload: data,
	}

	p := New()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse([]common.RawBlock{exifBlock})
	}
}
