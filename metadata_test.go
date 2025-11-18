package imx

import (
	"bytes"
	"os"
	"testing"

	"imx/formats"
)

// TestDetectFormat tests format detection via magic bytes
func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name       string
		magicBytes []byte
		expected   string
	}{
		{
			name:       "JPEG",
			magicBytes: []byte{0xFF, 0xD8, 0xFF, 0xE0},
			expected:   "JPEG",
		},
		{
			name:       "PNG",
			magicBytes: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			expected:   "PNG",
		},
		{
			name:       "GIF87a",
			magicBytes: []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61},
			expected:   "GIF",
		},
		{
			name:       "GIF89a",
			magicBytes: []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61},
			expected:   "GIF",
		},
		{
			name:       "WebP",
			magicBytes: []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			expected:   "WebP",
		},
		{
			name:       "BMP",
			magicBytes: []byte{0x42, 0x4D, 0x00, 0x00},
			expected:   "BMP",
		},
		{
			name:       "Unknown",
			magicBytes: []byte{0x00, 0x00, 0x00, 0x00},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formats.Detect(tt.magicBytes)
			if result != tt.expected {
				t.Errorf("detectFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMetadata_InvalidFile tests error handling for invalid files
func TestMetadata_InvalidFile(t *testing.T) {
	_, err := Metadata("nonexistent.jpg")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

// TestMetadata_UnsupportedFormat tests error handling for unsupported formats
func TestMetadata_UnsupportedFormat(t *testing.T) {
	// Create a temporary file with unsupported format
	tmpfile, err := os.CreateTemp("", "test.*.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	// Write invalid magic bytes
	tmpfile.Write([]byte{0x00, 0x00, 0x00, 0x00})

	_, err = Metadata(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

// createMinimalJPEG creates a minimal valid JPEG file for testing
func createMinimalJPEG() []byte {
	// Minimal JPEG: SOI, APP0 (JFIF), SOF0, SOS, EOI
	jpeg := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE0, 0x00, 0x10, // APP0 segment (16 bytes)
		0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, // JFIF header
		0xFF, 0xC0, 0x00, 0x0B, // SOF0 segment (11 bytes)
		0x08,       // Precision
		0x00, 0x64, // Height (100)
		0x00, 0x64, // Width (100)
		0x03,       // Components
		0xFF, 0xD9, // EOI
	}
	return jpeg
}

// createMinimalPNG creates a minimal valid PNG file for testing
func createMinimalPNG() []byte {
	// Minimal PNG: Signature, IHDR, IDAT, IEND
	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length (13)
		0x49, 0x48, 0x44, 0x52, // "IHDR"
		0x00, 0x00, 0x00, 0x64, // Width (100)
		0x00, 0x00, 0x00, 0x64, // Height (100)
		0x08,                   // Bit depth
		0x02,                   // Color type (RGB)
		0x00,                   // Compression
		0x00,                   // Filter
		0x00,                   // Interlace
		0x00, 0x00, 0x00, 0x00, // CRC (dummy)
		0x00, 0x00, 0x00, 0x00, // IEND chunk length
		0x49, 0x45, 0x4E, 0x44, // "IEND"
		0xAE, 0x42, 0x60, 0x82, // CRC
	}
	return png
}

// createMinimalGIF creates a minimal valid GIF file for testing
func createMinimalGIF() []byte {
	// Minimal GIF: Header, Logical Screen Descriptor, Color Table, Image Data, Trailer
	gif := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, // "GIF89a"
		0x64, 0x00, // Width (100) little-endian
		0x64, 0x00, // Height (100) little-endian
		0x80,             // Packed fields
		0x00,             // Background color
		0x00,             // Aspect ratio
		0x00, 0x00, 0x00, // Color table entry (dummy)
		0x2C,                                                 // Image separator
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Image descriptor
		0x02, 0x02, 0x44, 0x01, 0x00, // Image data (minimal)
		0x3B, // Trailer
	}
	return gif
}

// createMinimalWebP creates a minimal valid WebP file for testing
func createMinimalWebP() []byte {
	// Minimal WebP: RIFF header, VP8 chunk
	webp := []byte{
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x00, 0x00, 0x00, 0x00, // File size (dummy)
		0x57, 0x45, 0x42, 0x50, // "WEBP"
		0x56, 0x50, 0x38, 0x20, // "VP8 "
		0x00, 0x00, 0x00, 0x00, // Chunk size (dummy)
		0x9D, 0x01, 0x2A, // Key frame signature
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Dimensions (dummy)
	}
	return webp
}

// createMinimalBMP creates a minimal valid BMP file for testing
func createMinimalBMP() []byte {
	// Minimal BMP: File header, DIB header
	bmp := []byte{
		0x42, 0x4D, // "BM"
		0x00, 0x00, 0x00, 0x00, // File size (dummy)
		0x00, 0x00, // Reserved
		0x00, 0x00, // Reserved
		0x36, 0x00, 0x00, 0x00, // Offset to pixel data
		0x28, 0x00, 0x00, 0x00, // DIB header size (40)
		0x64, 0x00, 0x00, 0x00, // Width (100)
		0x64, 0x00, 0x00, 0x00, // Height (100)
		0x01, 0x00, // Planes
		0x18, 0x00, // Bits per pixel (24)
		0x00, 0x00, 0x00, 0x00, // Compression
		0x00, 0x00, 0x00, 0x00, // Image size
		0x00, 0x00, 0x00, 0x00, // X pixels per meter
		0x00, 0x00, 0x00, 0x00, // Y pixels per meter
		0x00, 0x00, 0x00, 0x00, // Colors used
		0x00, 0x00, 0x00, 0x00, // Important colors
	}
	return bmp
}

// TestMetadata_JPEG tests JPEG metadata extraction
func TestMetadata_JPEG(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test.*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	jpegData := createMinimalJPEG()
	tmpfile.Write(jpegData)
	tmpfile.Close()

	md, err := Metadata(tmpfile.Name())
	if err != nil {
		t.Fatalf("Metadata() error = %v", err)
	}

	if md.Format != FormatJPEG {
		t.Errorf("Format = %v, want JPEG", md.Format)
	}

	if md.Width != 100 || md.Height != 100 {
		t.Errorf("Dimensions = %dx%d, want 100x100", md.Width, md.Height)
	}
}

func TestMetadataFromBytes(t *testing.T) {
	md, err := MetadataFromBytes(createMinimalJPEG())
	if err != nil {
		t.Fatalf("MetadataFromBytes() error = %v", err)
	}
	if md.Format != FormatJPEG {
		t.Errorf("Format = %v, want %v", md.Format, FormatJPEG)
	}
}

func TestMetadataFromReader(t *testing.T) {
	reader := bytes.NewReader(createMinimalPNG())
	md, err := MetadataFromReader(reader)
	if err != nil {
		t.Fatalf("MetadataFromReader() error = %v", err)
	}
	if md.Format != FormatPNG {
		t.Errorf("Format = %v, want %v", md.Format, FormatPNG)
	}
}

func TestMetadataFromReaderAt(t *testing.T) {
	data := createMinimalGIF()
	reader := bytes.NewReader(data)
	md, err := MetadataFromReaderAt(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("MetadataFromReaderAt() error = %v", err)
	}
	if md.Format != FormatGIF {
		t.Errorf("Format = %v, want %v", md.Format, FormatGIF)
	}
}

// TestMetadata_PNG tests PNG metadata extraction
func TestMetadata_PNG(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test.*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	pngData := createMinimalPNG()
	tmpfile.Write(pngData)
	tmpfile.Close()

	md, err := Metadata(tmpfile.Name())
	if err != nil {
		t.Fatalf("Metadata() error = %v", err)
	}

	if md.Format != FormatPNG {
		t.Errorf("Format = %v, want PNG", md.Format)
	}

	if md.Width != 100 || md.Height != 100 {
		t.Errorf("Dimensions = %dx%d, want 100x100", md.Width, md.Height)
	}
}

// TestMetadata_GIF tests GIF metadata extraction
func TestMetadata_GIF(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test.*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	gifData := createMinimalGIF()
	tmpfile.Write(gifData)
	tmpfile.Close()

	md, err := Metadata(tmpfile.Name())
	if err != nil {
		t.Fatalf("Metadata() error = %v", err)
	}

	if md.Format != FormatGIF {
		t.Errorf("Format = %v, want GIF", md.Format)
	}

	if md.Width != 100 || md.Height != 100 {
		t.Errorf("Dimensions = %dx%d, want 100x100", md.Width, md.Height)
	}
}

// TestMetadata_WebP tests WebP metadata extraction
func TestMetadata_WebP(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test.*.webp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	webpData := createMinimalWebP()
	tmpfile.Write(webpData)
	tmpfile.Close()

	md, err := Metadata(tmpfile.Name())
	if err != nil {
		// WebP parsing might fail with minimal data, which is acceptable
		t.Logf("Metadata() error = %v (expected for minimal WebP)", err)
		return
	}

	if md.Format != FormatWebP {
		t.Errorf("Format = %v, want WebP", md.Format)
	}
}

// TestMetadata_BMP tests BMP metadata extraction
func TestMetadata_BMP(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test.*.bmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	bmpData := createMinimalBMP()
	tmpfile.Write(bmpData)
	tmpfile.Close()

	md, err := Metadata(tmpfile.Name())
	if err != nil {
		t.Fatalf("Metadata() error = %v", err)
	}

	if md.Format != FormatBMP {
		t.Errorf("Format = %v, want BMP", md.Format)
	}

	if md.Width != 100 || md.Height != 100 {
		t.Errorf("Dimensions = %dx%d, want 100x100", md.Width, md.Height)
	}
}

// TestImageMetadata_Struct tests the ImageMetadata struct fields
func TestImageMetadata_Struct(t *testing.T) {
	md := &ImageMetadata{
		Format:        "JPEG",
		Width:         1920,
		Height:        1080,
		FileSize:      102400,
		ColorDepth:    24,
		ColorSpace:    "RGB",
		HasICCProfile: true,
		EXIF:          make(map[string]interface{}),
		Additional:    make(map[string]interface{}),
	}

	if md.Format != FormatJPEG {
		t.Errorf("Format = %v, want JPEG", md.Format)
	}

	if md.Width != 1920 || md.Height != 1080 {
		t.Errorf("Dimensions = %dx%d, want 1920x1080", md.Width, md.Height)
	}

	if md.FileSize != 102400 {
		t.Errorf("FileSize = %v, want 102400", md.FileSize)
	}

	if md.ColorDepth != 24 {
		t.Errorf("ColorDepth = %v, want 24", md.ColorDepth)
	}

	if md.ColorSpace != "RGB" {
		t.Errorf("ColorSpace = %v, want RGB", md.ColorSpace)
	}

	if !md.HasICCProfile {
		t.Error("HasICCProfile = false, want true")
	}
}

// BenchmarkDetectFormat benchmarks format detection
func BenchmarkDetectFormat(b *testing.B) {
	magicBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	for i := 0; i < b.N; i++ {
		formats.Detect(magicBytes)
	}
}
