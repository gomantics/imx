package imx

import (
	"io"
)

// Default extractor instance used by package-level functions
var defaultExtractor = New()

// Extract extracts metadata from an io.Reader using the default extractor
func Extract(r io.Reader, opts ...Option) (Metadata, error) {
	return defaultExtractor.Metadata(r, opts...)
}

// ExtractFromFile extracts metadata from a file path using the default extractor
func ExtractFromFile(path string, opts ...Option) (Metadata, error) {
	return defaultExtractor.ExtractFromFile(path, opts...)
}

// ExtractFromBytes extracts metadata from a byte slice using the default extractor
func ExtractFromBytes(data []byte, opts ...Option) (Metadata, error) {
	return defaultExtractor.ExtractFromBytes(data, opts...)
}

// ExtractFromURL extracts metadata from an HTTP/HTTPS URL using the default extractor
func ExtractFromURL(url string, opts ...Option) (Metadata, error) {
	return defaultExtractor.ExtractFromURL(url, opts...)
}
