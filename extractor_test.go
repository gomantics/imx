package imx

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gomantics/imx/internal/format"
	"github.com/gomantics/imx/internal/meta"
)

// testJPEGPath is the path to the test JPEG file
const testJPEGPath = "testdata/DSC_1631.jpg"

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

func TestWithMaxBytes(t *testing.T) {
	cfg := ExtractorConfig{}
	opt := WithMaxBytes(1024)
	opt(&cfg)

	if cfg.MaxBytes != 1024 {
		t.Errorf("WithMaxBytes() MaxBytes = %d, want 1024", cfg.MaxBytes)
	}
}

func TestWithBufferSize(t *testing.T) {
	cfg := ExtractorConfig{}
	opt := WithBufferSize(32768)
	opt(&cfg)

	if cfg.BufferSize != 32768 {
		t.Errorf("WithBufferSize() BufferSize = %d, want 32768", cfg.BufferSize)
	}
}

func TestWithSpecs(t *testing.T) {
	cfg := ExtractorConfig{}
	opt := WithSpecs(SpecEXIF, SpecXMP)
	opt(&cfg)

	if len(cfg.Specs) != 2 {
		t.Errorf("WithSpecs() len(Specs) = %d, want 2", len(cfg.Specs))
	}
	if cfg.Specs[0] != SpecEXIF || cfg.Specs[1] != SpecXMP {
		t.Error("WithSpecs() Specs not set correctly")
	}
}

func TestWithFormats(t *testing.T) {
	cfg := ExtractorConfig{}
	opt := WithFormats(FormatJPEG, FormatPNG)
	opt(&cfg)

	if len(cfg.Formats) != 2 {
		t.Errorf("WithFormats() len(Formats) = %d, want 2", len(cfg.Formats))
	}
}

func TestWithStopOnFirstError(t *testing.T) {
	cfg := ExtractorConfig{}
	opt := WithStopOnFirstError()
	opt(&cfg)

	if !cfg.StopOnFirstErr {
		t.Error("WithStopOnFirstError() StopOnFirstErr should be true")
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
			metadata, err := e.Metadata(r, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Metadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != nil {
				if !strings.Contains(err.Error(), tt.errType.Error()) {
					t.Errorf("Metadata() error = %v, want error containing %v", err, tt.errType)
				}
			}

			if err == nil && len(metadata.Directories) == 0 {
				t.Log("Metadata() returned 0 directories, which is valid for some inputs")
			}
		})
	}
}

func TestExtractor_Metadata_WithMaxBytes(t *testing.T) {
	e := New(WithMaxBytes(100))
	validJPEG := loadTestJPEG(t)

	r := bytes.NewReader(validJPEG)
	_, err := e.Metadata(r)

	// With MaxBytes limiting the read, parsing may or may not succeed
	// depending on where the limit cuts off
	_ = err // Error is acceptable
}

func TestExtractor_Metadata_WithSpecFilter(t *testing.T) {
	e := New(WithSpecs(SpecXMP)) // Only want XMP, not EXIF
	validJPEG := loadTestJPEG(t) // Has EXIF

	r := bytes.NewReader(validJPEG)
	metadata, err := e.Metadata(r)

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

func TestExtractor_Metadata_StopOnFirstError(t *testing.T) {
	// Create extractor with StopOnFirstError
	e := New(WithStopOnFirstError())

	// Valid JPEG that parses successfully
	validJPEG := loadTestJPEG(t)
	r := bytes.NewReader(validJPEG)

	_, err := e.Metadata(r)
	if err != nil {
		t.Errorf("Metadata() error = %v for valid JPEG", err)
	}
}

func TestFilterBlocksForSpec(t *testing.T) {
	blocks := []format.RawBlock{
		{Spec: int(meta.SpecEXIF), Payload: []byte("exif1")},
		{Spec: int(meta.SpecXMP), Payload: []byte("xmp1")},
		{Spec: int(meta.SpecEXIF), Payload: []byte("exif2")},
		{Spec: int(meta.SpecICC), Payload: []byte("icc1")},
	}

	tests := []struct {
		name string
		spec Spec
		want int
	}{
		{
			name: "filter EXIF",
			spec: SpecEXIF,
			want: 2,
		},
		{
			name: "filter XMP",
			spec: SpecXMP,
			want: 1,
		},
		{
			name: "filter ICC",
			spec: SpecICC,
			want: 1,
		},
		{
			name: "filter IPTC (none)",
			spec: SpecIPTC,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterBlocksForSpec(blocks, tt.spec)
			if len(result) != tt.want {
				t.Errorf("filterBlocksForSpec() = %d blocks, want %d", len(result), tt.want)
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
			name:  "item present",
			slice: []Spec{SpecEXIF, SpecXMP, SpecICC},
			item:  SpecXMP,
			want:  true,
		},
		{
			name:  "item not present",
			slice: []Spec{SpecEXIF, SpecXMP},
			item:  SpecIPTC,
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

func TestFormat_Constants(t *testing.T) {
	// Verify Format constants are exported correctly
	tests := []struct {
		name   string
		format Format
		want   int
	}{
		{"FormatJPEG", FormatJPEG, int(format.FormatJPEG)},
		{"FormatPNG", FormatPNG, int(format.FormatPNG)},
		{"FormatWebP", FormatWebP, int(format.FormatWebP)},
		{"FormatTIFF", FormatTIFF, int(format.FormatTIFF)},
		{"FormatHEIF", FormatHEIF, int(format.FormatHEIF)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.format) != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.format, tt.want)
			}
		})
	}
}

// Custom reader that always fails on Read
type failingReader struct{}

func (r failingReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestExtractor_Metadata_ReaderError(t *testing.T) {
	e := New()
	_, err := e.Metadata(failingReader{})

	if err == nil {
		t.Error("Metadata() expected error for failing reader")
	}
}
