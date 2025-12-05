package imx

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

// MetadataFromFile extracts metadata from a file path
func MetadataFromFile(path string, opts ...Option) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: open file: %w", err)
	}
	defer f.Close()

	return Extract(f, opts...)
}

// MetadataFromBytes extracts metadata from a byte slice
func MetadataFromBytes(data []byte, opts ...Option) (Metadata, error) {
	return Extract(bytes.NewReader(data), opts...)
}

// MetadataFromURL extracts metadata from an HTTP/HTTPS URL
func MetadataFromURL(url string, opts ...Option) (Metadata, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Metadata{}, fmt.Errorf("imx: http status %d", resp.StatusCode)
	}

	return Extract(resp.Body, opts...)
}

// Extractor methods that mirror top-level functions

// MetadataFromFile extracts metadata from a file path using this extractor
func (e *Extractor) MetadataFromFile(path string, opts ...Option) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: open file: %w", err)
	}
	defer f.Close()

	return e.Metadata(f, opts...)
}

// MetadataFromBytes extracts metadata from a byte slice using this extractor
func (e *Extractor) MetadataFromBytes(data []byte, opts ...Option) (Metadata, error) {
	return e.Metadata(bytes.NewReader(data), opts...)
}

// MetadataFromURL extracts metadata from an HTTP/HTTPS URL using this extractor
func (e *Extractor) MetadataFromURL(url string, opts ...Option) (Metadata, error) {
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
