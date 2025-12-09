package imx

import (
	"bytes"
	"os"
	"testing"
)

// BenchmarkMetadataFromFile benchmarks full metadata extraction from a file
func BenchmarkMetadataFromFile(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := MetadataFromFile("testdata/jpeg/canon_xmp.jpg")
		if err != nil {
			b.Fatalf("MetadataFromFile failed: %v", err)
		}
	}
}

// BenchmarkMetadataFromBytes benchmarks metadata extraction from byte slice
func BenchmarkMetadataFromBytes(b *testing.B) {
	data, err := os.ReadFile("testdata/jpeg/canon_xmp.jpg")
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := MetadataFromBytes(data)
		if err != nil {
			b.Fatalf("MetadataFromBytes failed: %v", err)
		}
	}
}

// BenchmarkMetadataFromReader benchmarks metadata extraction from io.Reader
func BenchmarkMetadataFromReader(b *testing.B) {
	data, err := os.ReadFile("testdata/jpeg/canon_xmp.jpg")
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, err := MetadataFromReader(reader)
		if err != nil {
			b.Fatalf("MetadataFromReader failed: %v", err)
		}
	}
}

// BenchmarkMetadata_Tag benchmarks single tag lookup
func BenchmarkMetadata_Tag(b *testing.B) {
	meta, err := MetadataFromFile("testdata/jpeg/canon_xmp.jpg")
	if err != nil {
		b.Fatalf("MetadataFromFile failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = meta.Tag(TagMake)
	}
}

// BenchmarkMetadata_GetAll benchmarks batch tag retrieval
func BenchmarkMetadata_GetAll(b *testing.B) {
	meta, err := MetadataFromFile("testdata/jpeg/canon_xmp.jpg")
	if err != nil {
		b.Fatalf("MetadataFromFile failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = meta.GetAll(TagMake, TagModel, TagDateTimeOriginal, TagISO, TagFNumber)
	}
}

// BenchmarkMetadata_Each benchmarks iteration over all tags
func BenchmarkMetadata_Each(b *testing.B) {
	meta, err := MetadataFromFile("testdata/jpeg/canon_xmp.jpg")
	if err != nil {
		b.Fatalf("MetadataFromFile failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		meta.Each(func(dir Directory, tag Tag) bool {
			return true
		})
	}
}
