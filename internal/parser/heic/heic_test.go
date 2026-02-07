package heic

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestParser_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "HEIC" {
		t.Errorf("Name() = %v, want %v", got, "HEIC")
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid heic brand",
			data: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c'},
			want: true,
		},
		{
			name: "valid heif brand",
			data: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'f'},
			want: true,
		},
		{
			name: "valid mif1 brand",
			data: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'm', 'i', 'f', '1'},
			want: true,
		},
		{
			name: "invalid - not ftyp box",
			data: []byte{0, 0, 0, 24, 'm', 'o', 'o', 'v', 'h', 'e', 'i', 'c'},
			want: false,
		},
		{
			name: "invalid - unknown brand",
			data: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'x', 'x', 'x', 'x'},
			want: false,
		},
		{
			name: "too short",
			data: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p'},
			want: false,
		},
		{
			name: "empty",
			data: []byte{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			p := New()
			got := p.Detect(r)
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Detect_AllBrands(t *testing.T) {
	p := New()
	// Test HEIC brands
	for _, brand := range heicBrands {
		t.Run("HEIC_"+brand, func(t *testing.T) {
			data := []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', brand[0], brand[1], brand[2], brand[3]}
			r := bytes.NewReader(data)
			if !p.Detect(r) {
				t.Errorf("Detect() should recognize HEIC brand %s", brand)
			}
		})
	}
	// Test AVIF brands
	for _, brand := range avifBrands {
		t.Run("AVIF_"+brand, func(t *testing.T) {
			data := []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', brand[0], brand[1], brand[2], brand[3]}
			r := bytes.NewReader(data)
			if !p.Detect(r) {
				t.Errorf("Detect() should recognize AVIF brand %s", brand)
			}
		})
	}
}

// TestParser_Parse tests basic parsing - comprehensive validation is in validation_test.go
func TestParser_Parse(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/heic/apple_icc.HEIC")
	if err != nil {
		t.Skipf("Test file not found: %v", err)
	}

	p := New()
	r := bytes.NewReader(data)
	dirs, parseErr := p.Parse(r)

	// Should parse without panicking
	if parseErr != nil {
		t.Fatalf("Parse() error: %v", parseErr)
	}

	// Should have at least some directories
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories")
	}

	// Check that we have at least IFD0 and ExifIFD directories
	hasIFD0 := false
	hasExif := false
	for _, dir := range dirs {
		if dir.Name == "IFD0" {
			hasIFD0 = true
			if len(dir.Tags) == 0 {
				t.Error("IFD0 has no tags")
			}
		}
		if dir.Name == "ExifIFD" {
			hasExif = true
			if len(dir.Tags) == 0 {
				t.Error("ExifIFD has no tags")
			}
		}
	}

	if !hasIFD0 {
		t.Error("Missing IFD0 directory")
	}
	if !hasExif {
		t.Error("Missing ExifIFD directory")
	}
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
			name: "not HEIC",
			data: []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG
		},
		{
			name: "ftyp only - no meta box",
			data: []byte{
				0, 0, 0, 24, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c',
				0, 0, 0, 0, // Compatible brands (empty)
				0, 0, 0, 0,
				0, 0, 0, 0,
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

func TestParser_Parse_NoMetaBox(t *testing.T) {
	p := New()

	// Valid ftyp but no meta box
	data := []byte{
		0, 0, 0, 16, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c', 0, 0, 0, 0,
	}

	r := bytes.NewReader(data)
	dirs, err := p.Parse(r)

	if err == nil {
		t.Error("Parse() expected error for missing meta box")
	}
	if len(dirs) != 0 {
		t.Errorf("Parse() returned %d dirs, want 0", len(dirs))
	}
}

// TODO: Enable this test once TIFF parser race conditions are fixed
func TestParser_ConcurrentParse(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/heic/apple_icc.HEIC")
	if err != nil {
		t.Skipf("Test file not found: %v", err)
	}

	p := New()
	r := bytes.NewReader(data)

	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, _ = p.Parse(r)
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestParser_ImplementsInterface(t *testing.T) {
	var _ parser.Parser = (*Parser)(nil)
}

func TestParser_Parse_ReadError(t *testing.T) {
	p := New()

	// Minimal data that will trigger a read during meta box search
	data := []byte{
		0, 0, 0, 16, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c', 0, 0, 0, 0,
	}

	r := &errorReaderAt{
		data:        data,
		errorOffset: 16,
		customError: io.ErrUnexpectedEOF,
	}

	_, err := p.Parse(r)
	if err == nil {
		t.Error("Parse() expected error")
	}
}
