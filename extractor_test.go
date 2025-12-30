package imx

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

// Test helper is defined in api_test.go to avoid duplication

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
			name: "with HTTP timeout",
			opts: []Option{WithHTTPTimeout(60 * time.Second)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New(tt.opts...)
			if e == nil {
				t.Fatal("New() returned nil")
			}
			if len(e.parsers) == 0 {
				t.Error("No parsers registered")
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
			opts:    []Option{WithHTTPTimeout(60 * time.Second)},
			wantErr: false,
		},
		{
			name: "valid PNG signature",
			// PNG signature padded to 64 bytes (PNG parser should recognize this)
			data: append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
				make([]byte, 60)...),
			opts:    nil,
			wantErr: false,
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
			name:    "too short data",
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

			if err == nil && len(metadata.Directories()) == 0 {
				t.Log("Metadata() returned 0 directories, which is valid for some inputs")
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

// buildJPEGWithNoEXIF creates a valid JPEG without any EXIF data
// Must be at least 64 bytes for detection to succeed
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

	// Should succeed but have no directories (or only format metadata)
	if metadata == nil {
		t.Error("metadata should not be nil")
	}
}

func TestExtractor_metadataFromReaderAt_WithOptions(t *testing.T) {
	validJPEG := loadTestJPEG(t)

	e := New()
	r := bytes.NewReader(validJPEG)
	cfg := e.cloneConfig(WithHTTPTimeout(60 * time.Second))
	metadata, err := e.metadataFromReaderAt(r, cfg)

	if err != nil {
		t.Fatalf("metadataFromReaderAt() error = %v", err)
	}

	if metadata == nil {
		t.Fatal("metadataFromReaderAt() returned nil metadata")
	}
}

func TestExtractor_MetadataFromURL_ConfigClone(t *testing.T) {
	server := newIPv4Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validJPEG := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
		validJPEG = append(validJPEG, make([]byte, 54)...) // Pad to 64 bytes
		validJPEG = append(validJPEG, 0xFF, 0xD9)
		w.Write(validJPEG)
	}))
	if server == nil {
		return
	}
	defer server.Close()

	e := New()
	_, _ = e.MetadataFromURL(server.URL, WithHTTPTimeout(5*time.Second))
}
