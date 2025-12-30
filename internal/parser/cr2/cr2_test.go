package cr2

import (
	"bytes"
	"os"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.tiff == nil {
		t.Error("New() created parser with nil tiff parser")
	}
}

func TestParser_Name(t *testing.T) {
	p := New()
	got := p.Name()
	want := "CR2"
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
			name: "valid CR2 little-endian",
			// II (little-endian) + 0x002A (42) + IFD offset + CR2 magic + version
			data: []byte{
				0x49, 0x49, 0x2A, 0x00, // TIFF header: "II" + 42 little-endian
				0x10, 0x00, 0x00, 0x00, // IFD offset (16)
				0x43, 0x52, // CR2 magic "CR"
				0x02, 0x00, // Major version 0x02, minor 0x00
			},
			want: true,
		},
		{
			name: "valid CR2 big-endian",
			// MM (big-endian) + 0x002A (42) + IFD offset + CR2 magic + version
			data: []byte{
				0x4D, 0x4D, 0x00, 0x2A, // TIFF header: "MM" + 42 big-endian
				0x00, 0x00, 0x00, 0x10, // IFD offset (16)
				0x43, 0x52, // CR2 magic "CR"
				0x02, 0x01, // Major version 0x02, minor 0x01
			},
			want: true,
		},
		{
			name: "valid CR2 different minor version",
			data: []byte{
				0x49, 0x49, 0x2A, 0x00, // TIFF header
				0x10, 0x00, 0x00, 0x00, // IFD offset
				0x43, 0x52, // CR2 magic
				0x02, 0xFF, // Major 0x02, minor 0xFF
			},
			want: true,
		},
		{
			name: "invalid TIFF header",
			data: []byte{
				0x00, 0x00, 0x00, 0x00, // Invalid TIFF header
				0x00, 0x00, 0x00, 0x00,
				0x43, 0x52, // CR2 magic
				0x02, 0x00, // Version
			},
			want: false,
		},
		{
			name: "valid TIFF but wrong CR2 magic first byte",
			data: []byte{
				0x49, 0x49, 0x2A, 0x00, // Valid TIFF header
				0x10, 0x00, 0x00, 0x00,
				0x00, 0x52, // Wrong magic (not "CR")
				0x02, 0x00,
			},
			want: false,
		},
		{
			name: "valid TIFF but wrong CR2 magic second byte",
			data: []byte{
				0x49, 0x49, 0x2A, 0x00, // Valid TIFF header
				0x10, 0x00, 0x00, 0x00,
				0x43, 0x00, // Wrong magic (not "CR")
				0x02, 0x00,
			},
			want: false,
		},
		{
			name: "valid TIFF but wrong major version",
			data: []byte{
				0x49, 0x49, 0x2A, 0x00, // Valid TIFF header
				0x10, 0x00, 0x00, 0x00,
				0x43, 0x52, // Correct CR2 magic
				0x01, 0x00, // Wrong major version (not 0x02)
			},
			want: false,
		},
		{
			name: "valid TIFF but wrong major version 0x03",
			data: []byte{
				0x49, 0x49, 0x2A, 0x00, // Valid TIFF header
				0x10, 0x00, 0x00, 0x00,
				0x43, 0x52, // Correct CR2 magic
				0x03, 0x00, // Wrong major version
			},
			want: false,
		},
		{
			name: "too short - only TIFF header",
			data: []byte{
				0x49, 0x49, 0x2A, 0x00,
				0x10, 0x00, 0x00, 0x00,
			},
			want: false,
		},
		{
			name: "too short - TIFF + partial CR2",
			data: []byte{
				0x49, 0x49, 0x2A, 0x00,
				0x10, 0x00, 0x00, 0x00,
				0x43, // Only one byte of CR2 magic
			},
			want: false,
		},
		{
			name: "empty data",
			data: []byte{},
			want: false,
		},
		{
			name: "JPEG signature (not TIFF)",
			data: []byte{
				0xFF, 0xD8, 0xFF, 0xE0,
				0x00, 0x10, 0x4A, 0x46,
				0x43, 0x52, 0x02, 0x00,
			},
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

// TestParser_Parse tests basic parsing functionality with real file
func TestParser_Parse(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/cr2/sample1.cr2")
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

	// Check that we have at least IFD0 and ExifIFD
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

// TestParser_Parse_ErrorCases tests error handling
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
			name: "invalid data",
			data: []byte{0x00, 0x00, 0x00},
		},
		{
			name: "minimal valid CR2 with empty IFD",
			data: []byte{
				0x49, 0x49, 0x2A, 0x00, // TIFF header
				0x10, 0x00, 0x00, 0x00, // IFD offset = 16
				0x43, 0x52, // CR2 magic
				0x02, 0x00, // Version
				0x00, 0x00, 0x00, 0x00, // Padding
				0x00, 0x00, // 0 entries
				0x00, 0x00, 0x00, 0x00, // Next IFD = 0
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
	// Verify that Parser implements parser.Parser interface
	var _ parser.Parser = (*Parser)(nil)
}

func TestParser_ConcurrentParse(t *testing.T) {
	// Create minimal valid CR2 data
	data := []byte{
		0x49, 0x49, 0x2A, 0x00, // TIFF header: "II" + 42 little-endian
		0x10, 0x00, 0x00, 0x00, // IFD offset (16)
		0x43, 0x52, // CR2 magic "CR"
		0x02, 0x00, // Major version 0x02, minor 0x00
		// IFD at offset 16
		0x00, 0x00, // Number of entries (0)
		0x00, 0x00, 0x00, 0x00, // Next IFD offset (0)
	}

	p := New()
	r := bytes.NewReader(data)

	const goroutines = 10
	done := make(chan bool, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			p.Parse(r)
			done <- true
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}
