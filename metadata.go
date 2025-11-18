package imx

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"imx/formats"
)

var defaultHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

// Metadata reads an image file and extracts comprehensive metadata including
// format, dimensions, color information, EXIF data, and ICC profiles.
//
// The function detects the image format by reading the file's magic bytes,
// then extracts format-specific metadata.
//
// Example:
//
//	md, err := imx.Metadata("image.jpg")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Format: %s, Dimensions: %dx%d\n", md.Format, md.Width, md.Height)
func Metadata(filepath string) (*ImageMetadata, error) {
	return MetadataFromFile(filepath)
}

// MetadataFromFile extracts metadata from an image on disk.
func MetadataFromFile(path string) (*ImageMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return metadataFromSeeker(file, info.Size())
}

// MetadataFromBytes extracts metadata from an in-memory byte slice.
func MetadataFromBytes(data []byte) (*ImageMetadata, error) {
	reader := bytes.NewReader(data)
	return metadataFromSeeker(reader, int64(len(data)))
}

// MetadataFromReader reads all data from r into memory and extracts metadata.
func MetadataFromReader(r io.Reader) (*ImageMetadata, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSource, err)
	}
	return MetadataFromBytes(data)
}

// MetadataFromReaderAt extracts metadata from any io.ReaderAt with a known size.
func MetadataFromReaderAt(r io.ReaderAt, size int64) (*ImageMetadata, error) {
	section := io.NewSectionReader(r, 0, size)
	return metadataFromSeeker(section, size)
}

// MetadataFromURL downloads an image from a URL and extracts metadata.
func MetadataFromURL(url string) (*ImageMetadata, error) {
	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: unexpected status code %d from %s", ErrFetchFailed, resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return MetadataFromBytes(data)
}

func metadataFromSeeker(rs io.ReadSeeker, size int64) (*ImageMetadata, error) {
	magicBytes := make([]byte, 16)
	n, err := rs.Read(magicBytes)
	if err != nil && n == 0 {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSource, err)
	}
	magicBytes = magicBytes[:n]

	format := formats.Detect(magicBytes)
	if format == "" {
		return nil, ErrUnsupportedFormat
	}

	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSource, err)
	}

	md := &ImageMetadata{
		Format:     Format(format),
		FileSize:   size,
		EXIF:       make(map[string]interface{}),
		Additional: make(map[string]interface{}),
	}

	result, err := formats.Extract(format, rs)
	if err != nil {
		return nil, fmt.Errorf("failed to extract %s metadata: %w", format, err)
	}

	md.Width = result.Width
	md.Height = result.Height
	md.ColorDepth = result.ColorDepth
	md.ColorSpace = ColorSpace(result.ColorSpace)
	md.HasICCProfile = result.HasICCProfile
	if len(result.EXIF) > 0 {
		md.EXIF = result.EXIF
	}
	if len(result.Additional) > 0 {
		md.Additional = result.Additional
	}

	return md, nil
}
