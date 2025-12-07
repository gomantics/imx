package jpeg

import (
	"bufio"
	"bytes"
	"os"
	"testing"
)

// BenchmarkParser_Parse benchmarks JPEG marker parsing with typical camera file
func BenchmarkParser_Parse(b *testing.B) {
	data, err := os.ReadFile("../../../testdata/goldens/jpeg/canon_xmp.jpg")
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}

	p := New()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := bufio.NewReader(bytes.NewReader(data))
		_, _ = p.Parse(r)
	}
}
