package webp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

// Helper function to write RIFF header
func writeRIFFHeader(buf *bytes.Buffer, fileSize uint32) {
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, fileSize)
	buf.WriteString("WEBP")
}

// Helper function to write a WebP chunk
func writeChunk(buf *bytes.Buffer, fourCC string, data []byte) {
	buf.WriteString(fourCC)
	binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	buf.Write(data)
	// Add padding byte if size is odd
	if len(data)%2 != 0 {
		buf.WriteByte(0)
	}
}

// Helper function to create minimal VP8X chunk data
func createVP8X(width, height uint32, flags byte) []byte {
	data := make([]byte, 10)
	data[0] = flags
	// Width and height are stored as 24-bit values minus 1
	w := width - 1
	h := height - 1
	data[4] = byte(w)
	data[5] = byte(w >> 8)
	data[6] = byte(w >> 16)
	data[7] = byte(h)
	data[8] = byte(h >> 8)
	data[9] = byte(h >> 16)
	return data
}

// Helper function to create VP8 lossy chunk data
func createVP8(width, height uint32) []byte {
	data := make([]byte, 10)
	// Frame tag (3 bytes)
	frameTag := uint32(0x00) // Version 0, show_frame=0
	data[0] = byte(frameTag)
	data[1] = byte(frameTag >> 8)
	data[2] = byte(frameTag >> 16)
	// Start code
	data[3] = 0x9D
	data[4] = 0x01
	data[5] = 0x2A
	// Width (14 bits) and horizontal scale (2 bits)
	data[6] = byte(width)
	data[7] = byte(width >> 8)
	// Height (14 bits) and vertical scale (2 bits)
	data[8] = byte(height)
	data[9] = byte(height >> 8)
	return data
}

// Helper function to create VP8L lossless chunk data
func createVP8L(width, height uint32) []byte {
	data := make([]byte, 5)
	data[0] = 0x2F // Signature
	// Width and height are 14-bit values minus 1
	w := width - 1
	h := height - 1
	// Pack width and height into 5 bytes
	val := uint32(w) | (uint32(h) << 14)
	data[1] = byte(val)
	data[2] = byte(val >> 8)
	data[3] = byte(val >> 16)
	data[4] = byte(val >> 24)
	return data
}

// Helper functions
func findDir(dirs []parser.Directory, name string) *parser.Directory {
	for i := range dirs {
		if dirs[i].Name == name {
			return &dirs[i]
		}
	}
	return nil
}

func findTag(tags []parser.Tag, name string) *parser.Tag {
	for i := range tags {
		if tags[i].Name == name {
			return &tags[i]
		}
	}
	return nil
}

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.tiff == nil {
		t.Error("New() created parser with nil tiff parser")
	}
	if p.xmp == nil {
		t.Error("New() created parser with nil xmp parser")
	}
	if p.icc == nil {
		t.Error("New() created parser with nil icc parser")
	}
}

func TestParser_Name(t *testing.T) {
	p := New()
	got := p.Name()
	want := "WebP"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid WebP",
			data: []byte{
				'R', 'I', 'F', 'F', // RIFF signature
				0x00, 0x00, 0x00, 0x00, // File size (doesn't matter for detect)
				'W', 'E', 'B', 'P', // WEBP form type
			},
			want: true,
		},
		{
			name: "invalid RIFF signature",
			data: []byte{
				'X', 'I', 'F', 'F',
				0x00, 0x00, 0x00, 0x00,
				'W', 'E', 'B', 'P',
			},
			want: false,
		},
		{
			name: "invalid WEBP form type",
			data: []byte{
				'R', 'I', 'F', 'F',
				0x00, 0x00, 0x00, 0x00,
				'W', 'A', 'V', 'E', // WAVE, not WEBP
			},
			want: false,
		},
		{
			name: "too short",
			data: []byte{'R', 'I', 'F', 'F'},
			want: false,
		},
		{
			name: "empty",
			data: []byte{},
			want: false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got := p.Detect(r)
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Parse_VP8X(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)
	vp8x := createVP8X(1920, 1080, 0)
	writeChunk(&buf, "VP8X", vp8x)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}

	if dirs[0].Name != "WebP" {
		t.Errorf("Directory name = %q, want %q", dirs[0].Name, "WebP")
	}

	// Check dimensions
	widthTag := findTag(dirs[0].Tags, "ImageWidth")
	if widthTag == nil {
		t.Fatal("ImageWidth tag not found")
	}
	if widthTag.Value != uint32(1920) {
		t.Errorf("ImageWidth = %v, want 1920", widthTag.Value)
	}

	heightTag := findTag(dirs[0].Tags, "ImageHeight")
	if heightTag == nil {
		t.Fatal("ImageHeight tag not found")
	}
	if heightTag.Value != uint32(1080) {
		t.Errorf("ImageHeight = %v, want 1080", heightTag.Value)
	}
}

func TestParser_Parse_VP8X_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flags    byte
		expected string
	}{
		{"No flags", 0x00, "None"},
		{"EXIF flag", 0x02, "EXIF"},
		{"XMP flag", 0x04, "XMP"},
		{"ICCP flag", 0x08, "ICCP"},
		{"Alpha flag", 0x10, "Alpha"},
		{"Animation flag", 0x20, "Animation"},
		{"Multiple flags", 0x1E, "EXIF, XMP, ICCP, Alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writeRIFFHeader(&buf, 12)
			vp8x := createVP8X(100, 100, tt.flags)
			writeChunk(&buf, "VP8X", vp8x)

			r := bytes.NewReader(buf.Bytes())
			p := New()
			dirs, _ := p.Parse(r)

			flagsTag := findTag(dirs[0].Tags, "WebPFlags")
			if flagsTag == nil {
				t.Fatal("WebPFlags tag not found")
			}

			if flagsTag.Value != tt.expected {
				t.Errorf("WebPFlags = %q, want %q", flagsTag.Value, tt.expected)
			}
		})
	}
}

func TestParser_Parse_VP8_Lossy(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)
	vp8 := createVP8(640, 480)
	writeChunk(&buf, "VP8 ", vp8)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}

	webpDir := dirs[0]

	// Check dimensions
	widthTag := findTag(webpDir.Tags, "ImageWidth")
	if widthTag == nil {
		t.Fatal("ImageWidth tag not found")
	}
	if widthTag.Value != uint32(640) {
		t.Errorf("ImageWidth = %v, want 640", widthTag.Value)
	}

	heightTag := findTag(webpDir.Tags, "ImageHeight")
	if heightTag == nil {
		t.Fatal("ImageHeight tag not found")
	}
	if heightTag.Value != uint32(480) {
		t.Errorf("ImageHeight = %v, want 480", heightTag.Value)
	}

	// Check version tag
	versionTag := findTag(webpDir.Tags, "VP8Version")
	if versionTag == nil {
		t.Fatal("VP8Version tag not found")
	}
}

// TestParser_Parse_VP8_Versions tests different VP8 version numbers
func TestParser_Parse_VP8_Versions(t *testing.T) {
	tests := []struct {
		name       string
		version    uint32
		wantString string
	}{
		{"version 0", 0, "0 (bicubic reconstruction, normal loop)"},
		{"version 1", 1, "1 (simple/no loop filter)"},
		{"version 2", 2, "2 (complex/normal loop filter)"},
		{"version 3", 3, "3 (complex/simple loop filter)"},
		{"version 4", 4, "4 (bicubic reconstruction, normal loop)"}, // Falls through to default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writeRIFFHeader(&buf, 12)

			// Create VP8 with specific version
			data := make([]byte, 10)
			frameTag := (tt.version << 1) // version in bits 1-3
			data[0] = byte(frameTag)
			data[1] = byte(frameTag >> 8)
			data[2] = byte(frameTag >> 16)
			// Start code
			data[3] = 0x9D
			data[4] = 0x01
			data[5] = 0x2A
			// Dimensions
			data[6] = 100
			data[7] = 0
			data[8] = 100
			data[9] = 0

			writeChunk(&buf, "VP8 ", data)

			r := bytes.NewReader(buf.Bytes())
			p := New()
			dirs, _ := p.Parse(r)

			if len(dirs) == 0 {
				t.Fatal("Expected WebP directory")
			}

			versionTag := findTag(dirs[0].Tags, "VP8Version")
			if versionTag == nil {
				t.Fatal("VP8Version tag not found")
			}

			if versionTag.Value != tt.wantString {
				t.Errorf("VP8Version = %q, want %q", versionTag.Value, tt.wantString)
			}
		})
	}
}

// TestParser_Parse_VP8_ShowFrame tests VP8 showFrame flag
func TestParser_Parse_VP8_ShowFrame(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// Create VP8 with showFrame=1
	data := make([]byte, 10)
	frameTag := uint32(0x01) // showFrame bit set
	data[0] = byte(frameTag)
	data[1] = byte(frameTag >> 8)
	data[2] = byte(frameTag >> 16)
	// Start code
	data[3] = 0x9D
	data[4] = 0x01
	data[5] = 0x2A
	// Dimensions
	data[6] = 100
	data[7] = 0
	data[8] = 100
	data[9] = 0

	writeChunk(&buf, "VP8 ", data)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	if len(dirs) == 0 {
		t.Fatal("Expected WebP directory")
	}

	showFrameTag := findTag(dirs[0].Tags, "ShowFrame")
	if showFrameTag == nil {
		t.Error("ShowFrame tag not found when showFrame=1")
	}
}

func TestParser_Parse_VP8L_Lossless(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)
	vp8l := createVP8L(800, 600)
	writeChunk(&buf, "VP8L", vp8l)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}

	webpDir := dirs[0]

	// Check dimensions
	widthTag := findTag(webpDir.Tags, "ImageWidth")
	if widthTag == nil {
		t.Fatal("ImageWidth tag not found")
	}
	if widthTag.Value != uint32(800) {
		t.Errorf("ImageWidth = %v, want 800", widthTag.Value)
	}

	heightTag := findTag(webpDir.Tags, "ImageHeight")
	if heightTag == nil {
		t.Fatal("ImageHeight tag not found")
	}
	if heightTag.Value != uint32(600) {
		t.Errorf("ImageHeight = %v, want 600", heightTag.Value)
	}

	// Check format tag
	formatTag := findTag(webpDir.Tags, "Format")
	if formatTag == nil {
		t.Fatal("Format tag not found")
	}
	if formatTag.Value != "Lossless" {
		t.Errorf("Format = %q, want Lossless", formatTag.Value)
	}
}

func TestParser_Parse_EXIF_WithPrefix(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// Create minimal EXIF data with "Exif\x00\x00" header + TIFF
	var exifData bytes.Buffer
	exifData.WriteString("Exif\x00\x00")
	// Minimal TIFF: little-endian + magic + IFD offset + empty IFD
	exifData.Write([]byte{0x49, 0x49})             // Little-endian
	exifData.Write([]byte{0x2A, 0x00})             // Magic number 42
	exifData.Write([]byte{0x08, 0x00, 0x00, 0x00}) // IFD offset
	exifData.Write([]byte{0x00, 0x00})             // 0 entries

	writeChunk(&buf, "EXIF", exifData.Bytes())

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// TIFF parser returns empty for empty IFD, so we might have 0 dirs
	t.Logf("EXIF chunk with prefix parsed successfully, got %d directories", len(dirs))
}

func TestParser_Parse_EXIF_WithoutPrefix(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// EXIF data without "Exif\x00\x00" header (direct TIFF)
	var exifData bytes.Buffer
	exifData.Write([]byte{0x4D, 0x4D})             // Big-endian
	exifData.Write([]byte{0x00, 0x2A})             // Magic number 42
	exifData.Write([]byte{0x00, 0x00, 0x00, 0x08}) // IFD offset
	exifData.Write([]byte{0x00, 0x00})             // 0 entries

	writeChunk(&buf, "EXIF", exifData.Bytes())

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	t.Logf("EXIF chunk without prefix parsed successfully, got %d directories", len(dirs))
}

// TestParser_Parse_ExifDetectionBug tests the specific bug fix for EXIF detection
func TestParser_Parse_ExifDetectionBug(t *testing.T) {
	// Test case 1: EXIF chunk starting with 0xFF but not JPEG SOI (should NOT be rejected)
	t.Run("EXIF starts with 0xFF but not JPEG", func(t *testing.T) {
		var buf bytes.Buffer
		writeRIFFHeader(&buf, 12)

		// Create EXIF chunk starting with [0xFF, 0x00, ...] (NOT JPEG SOI)
		// This should be parsed, not rejected
		var exifData bytes.Buffer
		exifData.WriteByte(0xFF)                       // First byte is 0xFF
		exifData.WriteByte(0x00)                       // Second byte is NOT 0xD8
		exifData.Write([]byte{0x49, 0x49})             // Little-endian TIFF
		exifData.Write([]byte{0x2A, 0x00})             // Magic
		exifData.Write([]byte{0x08, 0x00, 0x00, 0x00}) // IFD offset
		exifData.Write([]byte{0x00, 0x00})             // 0 entries

		writeChunk(&buf, "EXIF", exifData.Bytes())

		r := bytes.NewReader(buf.Bytes())
		p := New()
		_, err := p.Parse(r)

		// Should NOT error - this is valid (though unusual) EXIF data
		if err != nil {
			t.Errorf("Parse() should accept EXIF starting with 0xFF but not JPEG SOI, got error: %v", err)
		}
	})

	// Test case 2: EXIF chunk starting with JPEG SOI should be rejected
	t.Run("EXIF starts with JPEG SOI", func(t *testing.T) {
		var buf bytes.Buffer
		writeRIFFHeader(&buf, 12)

		// Create EXIF chunk starting with JPEG SOI marker [0xFF, 0xD8]
		// This SHOULD be rejected (malformed EXIF)
		var exifData bytes.Buffer
		exifData.WriteByte(0xFF)           // JPEG SOI byte 1
		exifData.WriteByte(0xD8)           // JPEG SOI byte 2
		exifData.Write([]byte{0xFF, 0xE1}) // JPEG APP1 marker (typical for EXIF in JPEG)
		exifData.Write([]byte{0x00, 0x10}) // Segment size
		exifData.WriteString("Exif\x00\x00")
		exifData.Write([]byte{0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00})

		writeChunk(&buf, "EXIF", exifData.Bytes())

		r := bytes.NewReader(buf.Bytes())
		p := New()
		dirs, _ := p.Parse(r)

		// Should NOT have EXIF directories (rejected)
		// Only WebP directory should be present (if any)
		for _, dir := range dirs {
			if dir.Name != "WebP" {
				t.Errorf("Parse() should reject EXIF starting with JPEG SOI, but got directory: %s", dir.Name)
			}
		}
	})
}

func TestParser_Parse_ErrorCases(t *testing.T) {
	p := New()

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "invalid RIFF signature",
			data: []byte{'X', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P'},
		},
		{
			name: "invalid WEBP signature",
			data: []byte{'R', 'I', 'F', 'F', 0x04, 0, 0, 0, 'W', 'A', 'V', 'E'},
		},
		{
			name: "truncated chunk",
			data: []byte{
				'R', 'I', 'F', 'F',
				0x10, 0x00, 0x00, 0x00,
				'W', 'E', 'B', 'P',
				'V', 'P', '8', ' ',
				0xFF, 0xFF, 0xFF, 0xFF, // Huge chunk size
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			// Should not panic
			_, _ = p.Parse(r)
		})
	}
}

func TestParser_ImplementsInterface(t *testing.T) {
	var _ parser.Parser = (*Parser)(nil)
}

func TestParser_ConcurrentParse(t *testing.T) {
	// Create minimal valid WebP data with multiple chunks to exercise more code paths
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// Add VP8X chunk
	vp8x := createVP8X(100, 100, 0x02) // EXIF flag
	writeChunk(&buf, "VP8X", vp8x)

	// Add EXIF chunk
	exifData := []byte("Exif\x00\x00MM\x00\x2A\x00\x00\x00\x08")
	writeChunk(&buf, "EXIF", exifData)

	p := New()
	r := bytes.NewReader(buf.Bytes())

	// Run concurrent Parse operations
	const goroutines = 50 // Increased to make race conditions more likely
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			dirs, parseErr := p.Parse(r)

			// Verify results are correct
			if len(dirs) == 0 {
				errors <- fmt.Errorf("goroutine %d: expected directories, got 0", id)
				return
			}

			// Check for parse errors (OrNil returns nil if no errors)
			if parseErr != nil {
				errors <- fmt.Errorf("goroutine %d: unexpected parse error: %v", id, parseErr)
				return
			}

			// Verify we got the WebP directory
			foundWebP := false
			for _, dir := range dirs {
				if dir.Name == "WebP" {
					foundWebP = true
					break
				}
			}
			if !foundWebP {
				errors <- fmt.Errorf("goroutine %d: expected WebP directory", id)
				return
			}

			errors <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < goroutines; i++ {
		if err := <-errors; err != nil {
			t.Error(err)
		}
	}
}

// TestParser_Parse_MultipleChunks tests parsing WebP with multiple metadata chunks
func TestParser_Parse_MultipleChunks(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// VP8X with flags
	vp8x := createVP8X(1024, 768, 0x02) // EXIF flag
	writeChunk(&buf, "VP8X", vp8x)

	// EXIF chunk
	var exifData bytes.Buffer
	exifData.Write([]byte{0x49, 0x49})             // Little-endian (no "Exif" header)
	exifData.Write([]byte{0x2A, 0x00})             // Magic
	exifData.Write([]byte{0x08, 0x00, 0x00, 0x00}) // IFD offset
	exifData.Write([]byte{0x00, 0x00})             // 0 entries
	writeChunk(&buf, "EXIF", exifData.Bytes())

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should have WebP directory with VP8X tags
	if len(dirs) < 1 {
		t.Fatal("Expected at least WebP directory")
	}

	webpDir := findDir(dirs, "WebP")
	if webpDir == nil {
		t.Fatal("WebP directory not found")
	}

	// Verify dimensions from VP8X
	widthTag := findTag(webpDir.Tags, "ImageWidth")
	if widthTag == nil {
		t.Error("ImageWidth tag not found")
	}
}

// Edge case tests
func TestParser_Parse_VP8X_TruncatedData(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// VP8X chunk with insufficient data (only 5 bytes instead of 10)
	buf.WriteString("VP8X")
	binary.Write(&buf, binary.LittleEndian, uint32(5))
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00})

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should handle gracefully (skip the chunk or return empty)
	if len(dirs) > 0 && len(dirs[0].Tags) > 0 {
		t.Error("Parse() should not extract tags from truncated VP8X")
	}
}

func TestParser_Parse_VP8_InvalidStartCode(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// VP8 chunk with invalid start code
	vp8Data := make([]byte, 10)
	vp8Data[0] = 0x00
	vp8Data[1] = 0x00
	vp8Data[2] = 0x00
	vp8Data[3] = 0xFF // Invalid (should be 0x9D)
	vp8Data[4] = 0xFF // Invalid (should be 0x01)
	vp8Data[5] = 0xFF // Invalid (should be 0x2A)
	vp8Data[6] = 100
	vp8Data[7] = 0
	vp8Data[8] = 100
	vp8Data[9] = 0

	writeChunk(&buf, "VP8 ", vp8Data)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should skip invalid VP8 data
	if len(dirs) > 0 && len(dirs[0].Tags) > 0 {
		widthTag := findTag(dirs[0].Tags, "ImageWidth")
		if widthTag != nil {
			t.Error("Parse() should not extract tags from invalid VP8")
		}
	}
}

func TestParser_Parse_VP8L_InvalidSignature(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// VP8L chunk with invalid signature
	vp8lData := make([]byte, 5)
	vp8lData[0] = 0xFF // Invalid (should be 0x2F)
	vp8lData[1] = 0x00
	vp8lData[2] = 0x00
	vp8lData[3] = 0x00
	vp8lData[4] = 0x00

	writeChunk(&buf, "VP8L", vp8lData)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should skip invalid VP8L data
	if len(dirs) > 0 && len(dirs[0].Tags) > 0 {
		widthTag := findTag(dirs[0].Tags, "ImageWidth")
		if widthTag != nil {
			t.Error("Parse() should not extract tags from invalid VP8L")
		}
	}
}

func TestParser_Parse_EXIF_TooSmall(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// EXIF chunk with only 4 bytes (too small)
	writeChunk(&buf, "EXIF", []byte{0x00, 0x00, 0x00, 0x00})

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should handle gracefully (skip the chunk)
	t.Logf("Got %d directories from too-small EXIF", len(dirs))
}

func TestParser_Parse_XMP_Empty(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// Empty XMP chunk
	writeChunk(&buf, "XMP ", []byte{})

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should handle gracefully
	t.Logf("Got %d directories from empty XMP", len(dirs))
}

func TestParser_Parse_ICCP_Empty(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// Empty ICCP chunk
	writeChunk(&buf, "ICCP", []byte{})

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should handle gracefully
	t.Logf("Got %d directories from empty ICCP", len(dirs))
}

func TestParser_Parse_XMP_WithData(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// Create minimal XMP data
	xmpData := `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:creator><rdf:Seq><rdf:li>Test Author</rdf:li></rdf:Seq></dc:creator>
</rdf:Description>
</rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`

	writeChunk(&buf, "XMP ", []byte(xmpData))

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// XMP parser should parse the data (may or may not extract tags depending on XMP parser implementation)
	t.Logf("Got %d directories from XMP chunk with data", len(dirs))
}

func TestParser_Parse_ICCP_WithData(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// Create minimal ICC profile (128 bytes header minimum)
	minimalICC := make([]byte, 128)
	binary.BigEndian.PutUint32(minimalICC[0:4], 128) // Profile size

	writeChunk(&buf, "ICCP", minimalICC)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// ICC parser should parse the data (may or may not extract tags depending on ICC parser implementation)
	t.Logf("Got %d directories from ICCP chunk with data", len(dirs))
}

// failingReader is a mock io.ReaderAt that fails on specific read offsets
type failingReader struct {
	data        []byte
	failOffsets map[int64]bool
}

func (fr *failingReader) ReadAt(p []byte, off int64) (n int, err error) {
	if fr.failOffsets[off] {
		return 0, io.ErrUnexpectedEOF
	}
	if int(off) >= len(fr.data) {
		return 0, io.EOF
	}
	n = copy(p, fr.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

// TestParser_Parse_ReadErrors tests error handling when ReadAt fails
func TestParser_Parse_ReadErrors(t *testing.T) {
	// Create valid WebP data
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)
	vp8x := createVP8X(100, 100, 0x02)
	writeChunk(&buf, "VP8X", vp8x)

	data := buf.Bytes()

	t.Run("chunk header read error", func(t *testing.T) {
		// Fail when reading chunk header
		fr := &failingReader{
			data:        data,
			failOffsets: map[int64]bool{12: true}, // Fail at first chunk header
		}

		p := New()
		dirs, parseErr := p.Parse(fr)

		// Should handle error gracefully
		if parseErr == nil {
			t.Error("Expected parse error when chunk header read fails")
		}
		t.Logf("Got %d directories, error: %v", len(dirs), parseErr)
	})

	t.Run("VP8X data read error", func(t *testing.T) {
		// Fail when reading VP8X chunk data
		fr := &failingReader{
			data:        data,
			failOffsets: map[int64]bool{20: true}, // Fail when reading VP8X data (chunk.DataOffset)
		}

		p := New()
		dirs, _ := p.Parse(fr)

		// Should skip the chunk gracefully
		t.Logf("Got %d directories after VP8X read error", len(dirs))
	})
}

// TestParser_Detect_ReadError tests Detect with read failures
func TestParser_Detect_ReadError(t *testing.T) {
	fr := &failingReader{
		data:        []byte{},
		failOffsets: map[int64]bool{0: true},
	}

	p := New()
	if p.Detect(fr) {
		t.Error("Detect() should return false when read fails")
	}
}

// TestParser_Parse_EXIF_ReadError tests parseExifChunk read error
func TestParser_Parse_EXIF_ReadError(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// Create minimal EXIF chunk
	var exifData bytes.Buffer
	exifData.WriteString("Exif\x00\x00")
	exifData.Write([]byte{0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00})
	writeChunk(&buf, "EXIF", exifData.Bytes())

	data := buf.Bytes()

	// Calculate the offset where EXIF chunk data starts
	exifDataOffset := int64(20) // RIFF(12) + chunk header(8) = 20

	fr := &failingReader{
		data:        data,
		failOffsets: map[int64]bool{exifDataOffset: true}, // Fail reading EXIF chunk first 4 bytes
	}

	p := New()
	dirs, _ := p.Parse(fr)

	// Should handle gracefully (skip EXIF chunk)
	t.Logf("Got %d directories after EXIF read error", len(dirs))
}

// TestParser_Parse_VP8_ReadError tests parseImageChunk VP8 read error
func TestParser_Parse_VP8_ReadError(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)
	vp8 := createVP8(640, 480)
	writeChunk(&buf, "VP8 ", vp8)

	data := buf.Bytes()

	// VP8 data starts at offset 20 (RIFF header 12 + chunk header 8)
	vp8DataOffset := int64(20)

	fr := &failingReader{
		data:        data,
		failOffsets: map[int64]bool{vp8DataOffset: true}, // Fail reading VP8 data
	}

	p := New()
	dirs, _ := p.Parse(fr)

	// Should handle gracefully (skip VP8 chunk)
	t.Logf("Got %d directories after VP8 read error", len(dirs))
}

// TestParser_Parse_VP8L_ReadError tests parseImageChunk VP8L read error
func TestParser_Parse_VP8L_ReadError(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)
	vp8l := createVP8L(800, 600)
	writeChunk(&buf, "VP8L", vp8l)

	data := buf.Bytes()

	// VP8L data starts at offset 20
	vp8lDataOffset := int64(20)

	fr := &failingReader{
		data:        data,
		failOffsets: map[int64]bool{vp8lDataOffset: true}, // Fail reading VP8L data
	}

	p := New()
	dirs, _ := p.Parse(fr)

	// Should handle gracefully (skip VP8L chunk)
	t.Logf("Got %d directories after VP8L read error", len(dirs))
}

// TestParser_Parse_ChunkPadding tests odd-sized chunks that require padding
func TestParser_Parse_ChunkPadding(t *testing.T) {
	var buf bytes.Buffer

	// Build the body first to calculate correct file size
	var body bytes.Buffer

	// Create a chunk with odd size (will trigger padding logic on line 131-132)
	// EXIF chunk with size 7 bytes (odd)
	body.WriteString("EXIF")
	binary.Write(&body, binary.LittleEndian, uint32(7)) // Odd size
	body.WriteString("Exif123")                         // 7 bytes of data
	body.WriteByte(0)                                   // Padding byte

	// Add VP8X chunk after to verify padding was handled correctly
	vp8x := createVP8X(100, 100, 0)
	body.WriteString("VP8X")
	binary.Write(&body, binary.LittleEndian, uint32(len(vp8x)))
	body.Write(vp8x)

	// Now write RIFF header with correct size
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(4+body.Len())) // 4 for "WEBP"
	buf.WriteString("WEBP")
	buf.Write(body.Bytes())

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should successfully parse both chunks (padding handled correctly)
	// We should get WebP directory from VP8X
	if len(dirs) == 0 {
		t.Error("Expected at least one directory after parsing chunks with padding")
	}
	t.Logf("Successfully parsed %d directories with chunk padding", len(dirs))
}

// TestParser_Parse_EOFDuringChunkRead tests io.EOF when reading chunk header
func TestParser_Parse_EOFDuringChunkRead(t *testing.T) {
	var buf bytes.Buffer

	// Create RIFF header with fileSize that claims more data than actually present
	// This will cause the parser to try reading beyond the end of the file
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(100)) // Claims 100 bytes after "WEBP"
	buf.WriteString("WEBP")

	// Add one valid chunk
	vp8x := createVP8X(100, 100, 0)
	writeChunk(&buf, "VP8X", vp8x)

	// File ends here, but RIFF header claims more data exists
	// When parser tries to read next chunk at position = end of file,
	// ReadAt will return io.EOF (line 99-100)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should break cleanly on io.EOF
	// Should still have parsed the VP8X chunk successfully
	if len(dirs) == 0 {
		t.Error("Expected VP8X directory before EOF")
	}
	t.Logf("Got %d directories before EOF", len(dirs))
}

// TestParser_Parse_EXIF_OnlyHeader tests EXIF chunk with only "Exif" header, no data
func TestParser_Parse_EXIF_OnlyHeader(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// EXIF chunk with exactly 6 bytes: "Exif\x00\x00" - no TIFF data after
	// This triggers the size <= 0 check on line 201-202
	var exifData bytes.Buffer
	exifData.WriteString("Exif\x00\x00")

	writeChunk(&buf, "EXIF", exifData.Bytes())

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should handle gracefully (skip EXIF chunk with no data)
	t.Logf("Got %d directories from EXIF with only header", len(dirs))
}

// TestParser_Parse_EXIF_TooSmallForHeader tests EXIF chunk smaller than 6 bytes
func TestParser_Parse_EXIF_SmallChunk(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// EXIF chunk with only 5 bytes (less than minimum)
	// This triggers the chunk.Size < 6 check on line 173-174
	writeChunk(&buf, "EXIF", []byte("Exif\x00"))

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should skip this malformed EXIF chunk
	t.Logf("Got %d directories from undersized EXIF chunk", len(dirs))
}

// TestParser_Parse_VP8_TooSmall tests VP8 chunk with size < 10
func TestParser_Parse_VP8_TooSmall(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// VP8 chunk with only 8 bytes (less than minimum 10 required)
	// This triggers the chunk.Size < 10 check on line 293-295
	writeChunk(&buf, "VP8 ", []byte("12345678"))

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should skip this malformed VP8 chunk (no WebP directory created)
	if len(dirs) > 0 {
		t.Errorf("Expected no directories from undersized VP8 chunk, got %d", len(dirs))
	}
	t.Logf("Correctly skipped undersized VP8 chunk")
}

// TestParser_Parse_VP8L_TooSmall tests VP8L chunk with size < 5
func TestParser_Parse_VP8L_TooSmall(t *testing.T) {
	var buf bytes.Buffer
	writeRIFFHeader(&buf, 12)

	// VP8L chunk with only 3 bytes (less than minimum 5 required)
	// This triggers the chunk.Size < 5 check on line 348-350
	writeChunk(&buf, "VP8L", []byte("123"))

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should skip this malformed VP8L chunk (no WebP directory created)
	if len(dirs) > 0 {
		t.Errorf("Expected no directories from undersized VP8L chunk, got %d", len(dirs))
	}
	t.Logf("Correctly skipped undersized VP8L chunk")
}
