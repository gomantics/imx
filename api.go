package imx

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
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

// Extractor convenience methods

// ExtractFromFile extracts metadata from a file path using this extractor
func (e *Extractor) ExtractFromFile(path string, opts ...Option) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: open file: %w", err)
	}
	defer f.Close()

	return e.Metadata(f, opts...)
}

// ExtractFromBytes extracts metadata from a byte slice using this extractor
func (e *Extractor) ExtractFromBytes(data []byte, opts ...Option) (Metadata, error) {
	return e.Metadata(bytes.NewReader(data), opts...)
}

// ExtractFromURL extracts metadata from an HTTP/HTTPS URL using this extractor
func (e *Extractor) ExtractFromURL(url string, opts ...Option) (Metadata, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Metadata{}, fmt.Errorf("imx: http status %d", resp.StatusCode)
	}

	return e.Metadata(resp.Body, opts...)
}
