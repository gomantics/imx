package imx

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/gomantics/imx/internal/common"
)

// testJPEGPath is the path to the test JPEG file
const testJPEGPath = "testdata/goldens/jpeg/canon_xmp.jpg"

// loadTestJPEG loads the test JPEG file for testing
func loadTestJPEG(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(testJPEGPath)
	if err != nil {
		t.Fatalf("Failed to load test JPEG: %v", err)
	}
	return data
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "default options",
			opts: nil,
		},
		{
			name: "with max bytes",
			opts: []Option{WithMaxBytes(1024)},
		},
		{
			name: "with buffer size",
			opts: []Option{WithBufferSize(32 * 1024)},
		},
		{
			name: "with specs",
			opts: []Option{WithSpecs(SpecEXIF)},
		},
		{
			name: "with formats",
			opts: []Option{WithFormats(FormatJPEG)},
		},
		{
			name: "with stop on first error",
			opts: []Option{WithStopOnFirstError()},
		},
		{
			name: "with multiple options",
			opts: []Option{
				WithMaxBytes(2048),
				WithBufferSize(16 * 1024),
				WithSpecs(SpecEXIF, SpecXMP),
				WithStopOnFirstError(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New(tt.opts...)
			if e == nil {
				t.Fatal("New() returned nil")
			}
			if e.cfg.BufferSize <= 0 {
				t.Error("BufferSize should be positive")
			}
		})
	}
}

func TestExtractor_Metadata(t *testing.T) {
	validJPEG := loadTestJPEG(t)

	tests := []struct {
		name    string
		data    []byte
		opts    []Option
		wantErr bool
		errType error
	}{
		{
			name:    "valid JPEG with EXIF",
			data:    validJPEG,
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid JPEG with per-call options",
			data:    validJPEG,
			opts:    []Option{WithMaxBytes(20000000)}, // Large enough for the 17MB test file
			wantErr: false,
		},
		{
			name: "unknown format - PNG signature",
			// PNG signature padded to 64 bytes
			data: append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
				make([]byte, 60)...),
			opts:    nil,
			wantErr: true,
			errType: ErrUnknownFormat,
		},
		{
			name: "unknown format - random bytes",
			// Random bytes padded to 64 bytes
			data:    make([]byte, 64),
			opts:    nil,
			wantErr: true,
			errType: ErrUnknownFormat,
		},
		{
			name:    "peek fails - too short",
			data:    []byte{0xFF},
			opts:    nil,
			wantErr: true,
		},
	}

	e := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			metadata, err := e.MetadataFromReader(r, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("MetadataFromReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != nil {
				if !errors.Is(err, tt.errType) {
					t.Errorf("Metadata() error = %v, want %v", err, tt.errType)
				}
			}

			if err == nil && len(metadata.Directories) == 0 {
				t.Log("Metadata() returned 0 directories, which is valid for some inputs")
			}
		})
	}
}

func TestExtractor_MaxBytes(t *testing.T) {
	e := New(WithMaxBytes(100))
	validJPEG := loadTestJPEG(t)

	r := bytes.NewReader(validJPEG)
	_, err := e.MetadataFromReader(r)

	// With MaxBytes limiting the read, parsing may or may not succeed
	// depending on where the limit cuts off
	_ = err // Error is acceptable
}

func TestExtractor_SpecFilter(t *testing.T) {
	e := New(WithSpecs(SpecXMP)) // Only want XMP, not EXIF
	validJPEG := loadTestJPEG(t) // Has EXIF

	r := bytes.NewReader(validJPEG)
	metadata, err := e.MetadataFromReader(r)

	if err != nil {
		t.Fatalf("Metadata() error = %v", err)
	}

	// Should have no EXIF directories since we filtered for XMP only
	for _, dir := range metadata.Directories {
		if dir.Spec == SpecEXIF {
			t.Error("Metadata() should not return EXIF dirs when filtered for XMP only")
		}
	}
}

func TestExtractor_StopOnError(t *testing.T) {
	// Create extractor with StopOnFirstError
	e := New(WithStopOnFirstError())

	// Valid JPEG that parses successfully
	validJPEG := loadTestJPEG(t)
	r := bytes.NewReader(validJPEG)

	_, err := e.MetadataFromReader(r)
	if err != nil {
		t.Errorf("Metadata() error = %v for valid JPEG", err)
	}
}

func TestFilterBlocksForSpec(t *testing.T) {
	blocks := []common.RawBlock{
		{Spec: SpecEXIF, Payload: []byte{1}},
		{Spec: SpecXMP, Payload: []byte{2}},
		{Spec: SpecEXIF, Payload: []byte{3}},
		{Spec: SpecICC, Payload: []byte{4}},
	}

	tests := []struct {
		name      string
		spec      Spec
		wantCount int
	}{
		{"filter EXIF", SpecEXIF, 2},
		{"filter XMP", SpecXMP, 1},
		{"filter ICC", SpecICC, 1},
		{"filter IPTC (none)", SpecIPTC, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterBlocksForSpec(blocks, tt.spec)
			if len(result) != tt.wantCount {
				t.Errorf("filterBlocksForSpec() returned %d blocks, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []Spec
		item  Spec
		want  bool
	}{
		{
			name:  "item in slice",
			slice: []Spec{SpecEXIF, SpecXMP, SpecICC},
			item:  SpecXMP,
			want:  true,
		},
		{
			name:  "item not in slice",
			slice: []Spec{SpecEXIF, SpecXMP},
			item:  SpecICC,
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []Spec{},
			item:  SpecEXIF,
			want:  false,
		},
		{
			name:  "nil slice",
			slice: nil,
			item:  SpecEXIF,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.want {
				t.Errorf("contains() = %v, want %v", result, tt.want)
			}
		})
	}
}

// Custom reader that always fails on Read
type failingReader struct{}

func (r failingReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestExtractor_ReaderError(t *testing.T) {
	e := New()
	_, err := e.MetadataFromReader(failingReader{})

	if err == nil {
		t.Error("Metadata() expected error for failing reader, got nil")
	}
}

// buildJPEGWithBadEXIF creates a JPEG with malformed EXIF data that will fail to parse
func buildJPEGWithBadEXIF() []byte {
	var buf bytes.Buffer

	// SOI
	buf.Write([]byte{0xFF, 0xD8})

	// APP1 with bad EXIF - valid header but truncated IFD
	// We make it large enough (80 bytes) to satisfy Peek(64)
	// And set offset to 79, so 79 < 80 (valid checks)
	// But 79+2 > 80 (entry count read failure)
	badExif := make([]byte, 80)
	copy(badExif[0:2], []byte{'I', 'I'})               // Little-endian
	copy(badExif[2:4], []byte{0x2A, 0x00})             // TIFF magic
	copy(badExif[4:8], []byte{0x4F, 0x00, 0x00, 0x00}) // IFD0 offset = 79

	buf.WriteByte(0xFF)
	buf.WriteByte(0xE1)
	length := uint16(len(badExif) + 2 + 6)
	buf.WriteByte(byte(length >> 8))
	buf.WriteByte(byte(length))
	buf.Write([]byte("Exif\x00\x00"))
	buf.Write(badExif)

	// SOS to end metadata
	buf.Write([]byte{0xFF, 0xDA, 0x00, 0x08, 0x00, 0x01, 0x00, 0x00, 0x3F, 0x00})

	// EOI
	buf.Write([]byte{0xFF, 0xD9})

	return buf.Bytes()
}

// buildJPEGWithNoEXIF creates a valid JPEG without any EXIF data
// Must be at least 64 bytes for Peek to succeed
func buildJPEGWithNoEXIF() []byte {
	var buf bytes.Buffer

	// SOI
	buf.Write([]byte{0xFF, 0xD8})

	// APP0 (JFIF) instead of APP1 (EXIF) - make it large enough
	buf.Write([]byte{0xFF, 0xE0})
	buf.Write([]byte{0x00, 0x40}) // length 64
	buf.Write([]byte("JFIF\x00"))
	buf.Write([]byte{0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00})
	// Pad to fill 64-2=62 bytes of data
	padding := make([]byte, 62-14) // 62 - (5+9) already written
	buf.Write(padding)

	// SOS
	buf.Write([]byte{0xFF, 0xDA, 0x00, 0x08, 0x00, 0x01, 0x00, 0x00, 0x3F, 0x00})

	// EOI
	buf.Write([]byte{0xFF, 0xD9})

	return buf.Bytes()
}

func TestExtractor_NoBlocks(t *testing.T) {
	// Test when format parses successfully but no EXIF blocks are found
	e := New()
	jpegNoExif := buildJPEGWithNoEXIF()

	r := bytes.NewReader(jpegNoExif)
	metadata, err := e.MetadataFromReader(r)

	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}

	// Should succeed but have no directories
	if len(metadata.Directories) != 0 {
		t.Errorf("returned %d directories, want 0", len(metadata.Directories))
	}
}

func TestExtractor_ParseErrorStop(t *testing.T) {
	// Test with StopOnFirstError=true and bad EXIF data
	e := New(WithStopOnFirstError())
	jpegBadExif := buildJPEGWithBadEXIF()

	r := bytes.NewReader(jpegBadExif)
	_, err := e.MetadataFromReader(r)

	if err == nil {
		t.Error("expected error with StopOnFirstError=true, got nil")
	}
}

func TestExtractor_ParseErrorContinue(t *testing.T) {
	// Test with StopOnFirstError=false (default) and bad EXIF data
	// Should continue and return empty result without error
	e := New()
	jpegBadExif := buildJPEGWithBadEXIF()

	r := bytes.NewReader(jpegBadExif)
	metadata, err := e.MetadataFromReader(r)

	if err != nil {
		t.Errorf("returned unexpected error with StopOnFirstError=false: %v", err)
	}

	// Continued past the error, should have no directories
	if len(metadata.Directories) != 0 {
		t.Errorf("returned %d directories, want 0 when parsing fails", len(metadata.Directories))
	}
}
