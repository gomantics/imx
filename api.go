package imx

import (
	"io"
)

// Default extractor instance used by package-level functions
var defaultExtractor = New()

// MetadataFromReader extracts metadata from an io.Reader using the default extractor
func MetadataFromReader(r io.Reader, opts ...Option) (Metadata, error) {
	return defaultExtractor.MetadataFromReader(r, opts...)
}

// MetadataFromFile extracts metadata from a file path using the default extractor
func MetadataFromFile(path string, opts ...Option) (Metadata, error) {
	return defaultExtractor.MetadataFromFile(path, opts...)
}

// MetadataFromBytes extracts metadata from a byte slice using the default extractor
func MetadataFromBytes(data []byte, opts ...Option) (Metadata, error) {
	return defaultExtractor.MetadataFromBytes(data, opts...)
}

// MetadataFromURL extracts metadata from an HTTP/HTTPS URL using the default extractor
func MetadataFromURL(url string, opts ...Option) (Metadata, error) {
	return defaultExtractor.MetadataFromURL(url, opts...)
}
