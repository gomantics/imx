package png

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestParse_EmptyFile(t *testing.T) {
	r := bytes.NewReader([]byte{})
	p := New()

	dirs, err := p.Parse(r)

	if err == nil {
		t.Error("Parse() expected error for empty file")
	}

	if dirs != nil {
		t.Error("Parse() should return nil dirs on error")
	}
}

func TestParse_OnlySignature(t *testing.T) {
	data := pngSignature
	r := bytes.NewReader(data)
	p := New()

	dirs, err := p.Parse(r)

	// Should either return error or empty directories (both are acceptable)
	if err == nil && len(dirs) > 0 {
		t.Error("Parse() should return error or empty dirs for PNG with only signature")
	}
}

func TestName(t *testing.T) {
	p := New()
	if p.Name() != "PNG" {
		t.Errorf("Name() = %q, want %q", p.Name(), "PNG")
	}
}

// Legacy tests using external files

func TestDetect_ExternalFiles(t *testing.T) {
	tests := []struct {
		name string
		file string
		want bool
	}{
		{
			name: "valid png",
			file: "../../../testdata/png/basic.png",
			want: true,
		},
		{
			name: "not png",
			file: "../../../testdata/jpeg/canon_xmp.jpg",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.file)
			if err != nil {
				t.Skip("test file not found")
			}
			defer f.Close()

			p := New()
			got := p.Detect(f)
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse_ExternalFiles(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
		minDirs int
	}{
		{
			name:    "valid png file",
			file:    "../../../testdata/png/basic.png",
			wantErr: false,
			minDirs: 0, // May have text/EXIF depending on file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.file)
			if err != nil {
				t.Skip("test file not found")
			}
			defer f.Close()

			p := New()
			dirs, parseErr := p.Parse(f)

			if (parseErr != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", parseErr, tt.wantErr)
				return
			}

			if len(dirs) < tt.minDirs {
				t.Errorf("Parse() got %d directories, want at least %d", len(dirs), tt.minDirs)
			}
		})
	}
}

// Test with empty reader to trigger potential edge cases
func TestParse_EmptyReader(t *testing.T) {
	p := New()
	r := bytes.NewReader([]byte{})
	dirs, err := p.Parse(r)

	// Should return error for invalid PNG
	if err == nil {
		t.Error("Expected error for empty reader")
	}
	if len(dirs) != 0 {
		t.Errorf("Expected no directories for empty reader, got %d", len(dirs))
	}
}

// Core Parser Tests

func TestDetect_ValidSignature(t *testing.T) {
	data := createMinimalPNG()
	r := bytes.NewReader(data)
	p := New()

	if !p.Detect(r) {
		t.Error("Detect() failed for valid PNG signature")
	}
}

func TestDetect_InvalidSignature(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	r := bytes.NewReader(data)
	p := New()

	if p.Detect(r) {
		t.Error("Detect() succeeded for invalid signature")
	}
}

func TestDetect_TooShort(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E}
	r := bytes.NewReader(data)
	p := New()

	if p.Detect(r) {
		t.Error("Detect() succeeded for too short data")
	}
}

func TestParse_MinimalValid(t *testing.T) {
	data := createMinimalPNG()
	r := bytes.NewReader(data)
	p := New()

	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(dirs) == 0 {
		t.Error("Parse() returned no directories for minimal valid PNG")
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Error("Parse() did not return PNG directory")
	}
}

func TestParse_InvalidSignature(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()

	dirs, err := p.Parse(r)

	if err == nil {
		t.Error("Parse() expected error for invalid signature")
	}

	if dirs != nil {
		t.Error("Parse() should return nil dirs on signature error")
	}
}

func TestParse_UnknownChunkType(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// Unknown chunk type
	writeChunk(&buf, "xYZa", []byte{0x01, 0x02, 0x03, 0x04})

	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()

	dirs, err := p.Parse(r)

	// Should not error on unknown chunks, just skip them
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Error("Parse() should still parse IHDR despite unknown chunk")
	}
}

func TestParse_TruncatedChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// Write chunk header but truncate data
	chunkLen := uint32(100)
	binary.Write(&buf, binary.BigEndian, chunkLen)
	buf.WriteString("tEXt")
	buf.Write([]byte{0x01, 0x02}) // Only 2 bytes instead of 100

	r := bytes.NewReader(buf.Bytes())
	p := New()

	dirs, err := p.Parse(r)

	// May error or may skip truncated chunk
	// Either way, IHDR should have been parsed
	if err == nil && len(dirs) > 0 {
		pngDir := findDir(dirs, "PNG")
		if pngDir == nil {
			t.Error("Parse() should have parsed IHDR before truncated chunk")
		}
	}
}

func TestParse_MultipleChunks(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(100, 100, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// Add gAMA chunk
	gammaData := make([]byte, 4)
	binary.BigEndian.PutUint32(gammaData, 220000)
	writeChunk(&buf, "gAMA", gammaData)

	// Add tEXt chunk
	textData := []byte("Author\x00John Doe")
	writeChunk(&buf, "tEXt", textData)

	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()

	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Parse() did not return PNG directory")
	}

	// Should have IHDR tags
	if findTag(pngDir.Tags, "ImageWidth") == nil {
		t.Error("Parse() did not parse ImageWidth from IHDR")
	}

	// Should have gAMA tag
	if findTag(pngDir.Tags, "Gamma") == nil {
		t.Error("Parse() did not parse Gamma")
	}

	// Should have text directory
	textDir := findDir(dirs, "PNG-Text")
	if textDir == nil {
		t.Error("Parse() did not create PNG Text directory")
	}
}

func TestParse_ReadChunkError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()

	// Error when reading chunk header after signature
	errorOffset := int64(len(pngSignature))
	r := &errorReaderAt{data: data, errorAtOffset: errorOffset}

	p := New()
	dirs, err := p.Parse(r)

	// Should error when chunk header can't be read
	if err == nil {
		t.Error("Expected error when chunk header read fails")
	}

	if dirs != nil {
		t.Error("Expected nil dirs when chunk read fails")
	}
}

func TestParser_ConcurrentParse(t *testing.T) {
	data := createMinimalPNG()

	p := New()
	r := bytes.NewReader(data)

	// Run Parse concurrently with the same Parser instance
	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = p.Parse(r)
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

// ReadAt Error Tests

func TestParse_SignatureReadError(t *testing.T) {
	data := createMinimalPNG()
	// Error at signature offset (0)
	r := &errorReaderAt{data: data, errorAtOffset: 0}

	p := New()
	dirs, err := p.Parse(r)

	if err == nil {
		t.Error("Expected error when signature read fails")
	}
	if dirs != nil {
		t.Error("Expected nil dirs when signature read fails")
	}
}

// Helper functions

func writeChunk(buf *bytes.Buffer, chunkType string, data []byte) {
	// Write length
	length := uint32(0)
	if data != nil {
		length = uint32(len(data))
	}
	binary.Write(buf, binary.BigEndian, length)

	// Write type
	buf.WriteString(chunkType)

	// Write data
	if data != nil {
		buf.Write(data)
	}

	// Calculate and write CRC
	crc := crc32.NewIEEE()
	crc.Write([]byte(chunkType))
	if data != nil {
		crc.Write(data)
	}
	binary.Write(buf, binary.BigEndian, crc.Sum32())
}

func createIHDR(width, height uint32, bitDepth, colorType byte) []byte {
	ihdr := make([]byte, ihdrChunkSize)
	binary.BigEndian.PutUint32(ihdr[ihdrWidthOffset:ihdrWidthOffset+4], width)
	binary.BigEndian.PutUint32(ihdr[ihdrHeightOffset:ihdrHeightOffset+4], height)
	ihdr[ihdrBitDepthOffset] = bitDepth
	ihdr[ihdrColorTypeOffset] = colorType
	ihdr[ihdrCompressionOffset] = compressionDeflate
	ihdr[ihdrFilterOffset] = filterAdaptive
	ihdr[ihdrInterlaceOffset] = interlaceNone
	return ihdr
}

func createMinimalPNG() []byte {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(1, 1, 8, colorTypeRGB)
	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	return buf.Bytes()
}

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

type errorReaderAt struct {
	data          []byte
	errorAtOffset int64
}

func (e *errorReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= e.errorAtOffset {
		return 0, errors.New("forced read error")
	}
	if off >= int64(len(e.data)) {
		return 0, io.EOF
	}
	n = copy(p, e.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func findSubstring(data []byte, substr string) int {
	return bytes.Index(data, []byte(substr))
}
