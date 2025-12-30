package imx

import (
	"io"
)

// Default extractor instance used by package-level functions
var defaultExtractor = New()

// MetadataFromFile extracts metadata from a file path using the default extractor.
//
// The opts parameter accepts functional options to customize extraction behavior.
// Currently available options:
//   - WithHTTPTimeout: Has no effect for file operations (only applies to MetadataFromURL)
//
// The opts parameter is provided for API consistency and forward compatibility
// with future configuration options.
func MetadataFromFile(path string, opts ...Option) (*Metadata, error) {
	return defaultExtractor.MetadataFromFile(path, opts...)
}

// MetadataFromBytes extracts metadata from a byte slice using the default extractor.
//
// The opts parameter accepts functional options to customize extraction behavior.
// Currently available options:
//   - WithHTTPTimeout: Has no effect for byte operations (only applies to MetadataFromURL)
//
// The opts parameter is provided for API consistency and forward compatibility
// with future configuration options.
func MetadataFromBytes(data []byte, opts ...Option) (*Metadata, error) {
	return defaultExtractor.MetadataFromBytes(data, opts...)
}

// MetadataFromReader extracts metadata from an io.Reader using the default extractor.
// This buffers data on-demand using a smart adapter that implements io.ReaderAt.
//
// The opts parameter accepts functional options to customize extraction behavior.
// Currently available options:
//   - WithHTTPTimeout: Has no effect for reader operations (only applies to MetadataFromURL)
//
// The opts parameter is provided for API consistency and forward compatibility
// with future configuration options.
func MetadataFromReader(r io.Reader, opts ...Option) (*Metadata, error) {
	return defaultExtractor.MetadataFromReader(r, opts...)
}

// MetadataFromURL extracts metadata from an HTTP/HTTPS URL using the default extractor.
//
// The opts parameter accepts functional options to customize extraction behavior.
// Available options:
//   - WithHTTPTimeout: Sets the HTTP request timeout (default: 30 seconds)
//
// Example:
//
//	meta, err := imx.MetadataFromURL("https://example.com/photo.jpg",
//		imx.WithHTTPTimeout(60 * time.Second))
func MetadataFromURL(url string, opts ...Option) (*Metadata, error) {
	return defaultExtractor.MetadataFromURL(url, opts...)
}
