package imx

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// ImageMetadata contains comprehensive metadata extracted from an image file.
type ImageMetadata struct {
	// Format is the detected image format (e.g., "JPEG", "PNG", "GIF", "WebP", "BMP").
	Format string

	// Width and Height are the image dimensions in pixels.
	Width, Height int

	// FileSize is the size of the image file in bytes.
	FileSize int64

	// ColorDepth is the number of bits per pixel or color channel.
	ColorDepth int

	// ColorSpace indicates the color space (e.g., "RGB", "RGBA", "CMYK", "Grayscale", "Indexed").
	ColorSpace string

	// HasICCProfile indicates whether the image contains an ICC color profile.
	HasICCProfile bool

	// EXIF contains EXIF metadata as a map of tag names to values.
	// Common tags include: "DateTime", "Make", "Model", "ISO", "FNumber", "ExposureTime", etc.
	EXIF map[string]interface{}

	// Format-specific additional metadata
	// For JPEG: Quality estimate, subsampling
	// For PNG: Bit depth, compression method, filter method, interlace
	// For GIF: Color table size, has transparency, animation info
	// For WebP: Has animation, has alpha
	// For BMP: Compression type, planes
	Additional map[string]interface{}
}

// setAdditional stores a value lazily in the Additional map.
func (md *ImageMetadata) setAdditional(key string, value interface{}) {
	if md.Additional == nil {
		md.Additional = make(map[string]interface{})
	}
	md.Additional[key] = value
}

// mergeAdditional merges a map into Additional lazily.
func (md *ImageMetadata) mergeAdditional(values map[string]interface{}) {
	if len(values) == 0 {
		return
	}
	if md.Additional == nil {
		md.Additional = make(map[string]interface{}, len(values))
	}
	for k, v := range values {
		md.Additional[k] = v
	}
}

// mergeEXIF merges EXIF metadata lazily to avoid upfront allocations.
func (md *ImageMetadata) mergeEXIF(values map[string]interface{}) {
	if len(values) == 0 {
		return
	}
	if md.EXIF == nil {
		md.EXIF = make(map[string]interface{}, len(values))
	}
	for k, v := range values {
		md.EXIF[k] = v
	}
}

var bytePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 64*1024)
	},
}

func borrowBuffer(size int) []byte {
	if size <= 0 {
		return nil
	}
	buf := bytePool.Get().([]byte)
	if cap(buf) < size {
		buf = make([]byte, size)
	}
	return buf[:size]
}

func releaseBuffer(buf []byte) {
	if buf == nil {
		return
	}
	bytePool.Put(buf)
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
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	return MetadataFromReader(file, fileSize)
}

// MetadataFromReader extracts metadata from any io.ReadSeeker, allowing callers
// to reuse already opened readers for better performance.
func MetadataFromReader(r io.ReadSeeker, fileSize int64) (*ImageMetadata, error) {
	var magicBytes [16]byte
	n, err := r.Read(magicBytes[:])
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}
	header := magicBytes[:n]

	// Detect format
	format := detectFormat(header)
	if format == "" {
		return nil, fmt.Errorf("unsupported image format")
	}

	// Reset reader to beginning
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek reader: %w", err)
	}

	md := &ImageMetadata{
		Format:   format,
		FileSize: fileSize,
	}

	var extractErr error
	switch format {
	case "JPEG":
		extractErr = ExtractJPEGMetadata(r, md)
	case "PNG":
		extractErr = ExtractPNGMetadata(r, md)
	case "GIF":
		extractErr = ExtractGIFMetadata(r, md)
	case "WebP":
		extractErr = ExtractWebPMetadata(r, md)
	case "BMP":
		extractErr = ExtractBMPMetadata(r, md)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if extractErr != nil {
		return nil, fmt.Errorf("failed to extract %s metadata: %w", format, extractErr)
	}

	return md, nil
}
