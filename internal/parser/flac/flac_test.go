package flac

import (
	"bytes"
	"os"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestParser_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "FLAC" {
		t.Errorf("Name() = %v, want %v", got, "FLAC")
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid FLAC marker",
			data: []byte("fLaC"),
			want: true,
		},
		{
			name: "invalid first byte (F instead of f)",
			data: []byte("FLaC"),
			want: false,
		},
		{
			name: "invalid second byte (l instead of L)",
			data: []byte("flaC"),
			want: false,
		},
		{
			name: "invalid third byte (A instead of a)",
			data: []byte("fLAC"),
			want: false,
		},
		{
			name: "invalid fourth byte (c instead of C)",
			data: []byte("fLac"),
			want: false,
		},
		{
			name: "wrong magic completely",
			data: []byte("ABCD"),
			want: false,
		},
		{
			name: "too short (3 bytes)",
			data: []byte("fLa"),
			want: false,
		},
		{
			name: "too short (2 bytes)",
			data: []byte("fL"),
			want: false,
		},
		{
			name: "too short (1 byte)",
			data: []byte("f"),
			want: false,
		},
		{
			name: "empty",
			data: []byte{},
			want: false,
		},
		{
			name: "valid with extra data",
			data: []byte("fLaC\x00\x00\x00\x00extra data"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			r := bytes.NewReader(tt.data)
			if got := p.Detect(r); got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
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
			name: "invalid marker",
			data: []byte("fLac"),
		},
		{
			name: "truncated file - marker only",
			data: []byte("fLaC"),
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

// Ensure Parser implements parser.Parser interface
func TestParser_ImplementsInterface(t *testing.T) {
	var _ parser.Parser = (*Parser)(nil)
}

// TestParser_ConcurrentParse tests that the parser can be used concurrently
// This test will expose the data race in the pos field when run with -race flag
func TestParser_ConcurrentParse(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/flac/sample3_hires.flac")
	if err != nil {
		t.Skipf("Test file not found: %v", err)
	}

	p := New()
	r := bytes.NewReader(data)

	// Run Parse concurrently with the same Parser instance
	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, _ = p.Parse(r)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// TestParser_Parse_CueSheetBlock tests parsing a FLAC file with a cue sheet block
func TestParser_Parse_CueSheetBlock(t *testing.T) {
	// Create a minimal FLAC file with a cue sheet block
	buf := &bytes.Buffer{}
	buf.WriteString("fLaC") // FLAC marker
	// Block header: last block (0x80), cue sheet type (0x05), length 100
	buf.Write([]byte{0x85, 0x00, 0x00, 0x64}) // 0x85 = last block + type 0x05
	buf.Write(make([]byte, 100))              // 100 bytes of cue sheet data

	p := New()
	dirs, parseErr := p.Parse(bytes.NewReader(buf.Bytes()))

	if parseErr != nil {
		t.Errorf("Parse() unexpected error: %v", parseErr)
	}

	// Should have one directory for the cue sheet
	if len(dirs) != 1 {
		t.Errorf("Parse() got %d directories, want 1", len(dirs))
	}

	if len(dirs) > 0 {
		if dirs[0].Name != "FLAC-CueSheet" {
			t.Errorf("Directory name = %v, want FLAC-CUESHEET", dirs[0].Name)
		}
	}
}

// TestParser_Parse_UnknownBlockType tests parsing a FLAC file with unknown block type
func TestParser_Parse_UnknownBlockType(t *testing.T) {
	// Create a minimal FLAC file with an unknown block type
	buf := &bytes.Buffer{}
	buf.WriteString("fLaC") // FLAC marker
	// Block header: last block (0x80), unknown type (0x7F), length 10
	buf.Write([]byte{0xFF, 0x00, 0x00, 0x0A}) // 0xFF = last block + type 0x7F
	buf.Write(make([]byte, 10))               // 10 bytes of data

	p := New()
	dirs, parseErr := p.Parse(bytes.NewReader(buf.Bytes()))

	if parseErr != nil {
		t.Errorf("Parse() unexpected error: %v", parseErr)
	}

	// Should have one directory for the unknown block
	if len(dirs) != 1 {
		t.Errorf("Parse() got %d directories, want 1", len(dirs))
	}

	if len(dirs) > 0 {
		if dirs[0].Name != "FLAC Block 127" {
			t.Errorf("Directory name = %v, want FLAC Block 127", dirs[0].Name)
		}
	}
}

// TestParser_Parse tests basic parsing - comprehensive validation is in validation_test.go
func TestParser_Parse(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/flac/sample3_hires.flac")
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

	// Check that we have at least STREAMINFO and VORBIS directories
	hasStreamInfo := false
	hasVorbis := false
	for _, dir := range dirs {
		if dir.Name == "FLAC-StreamInfo" {
			hasStreamInfo = true
			if len(dir.Tags) == 0 {
				t.Error("FLAC-STREAMINFO has no tags")
			}
		}
		if dir.Name == "FLAC-Vorbis" {
			hasVorbis = true
			if len(dir.Tags) == 0 {
				t.Error("FLAC-VORBIS has no tags")
			}
		}
	}

	if !hasStreamInfo {
		t.Error("Missing FLAC-STREAMINFO directory")
	}
	if !hasVorbis {
		t.Error("Missing FLAC-VORBIS directory")
	}
}

// TestParser_Parse_ExcessiveBlockSize tests that block length validation prevents excessive memory allocation
func TestParser_Parse_ExcessiveBlockSize(t *testing.T) {
	p := New()

	var buf bytes.Buffer
	// Write FLAC marker
	buf.WriteString("fLaC")

	// Write STREAMINFO block header with excessive size (10MB, exceeds 8MB limit)
	buf.WriteByte(0x80) // Last block flag set, type = 0 (STREAMINFO)
	excessiveSize := 10 * 1024 * 1024
	buf.WriteByte(byte(excessiveSize >> 16))
	buf.WriteByte(byte(excessiveSize >> 8))
	buf.WriteByte(byte(excessiveSize))

	r := bytes.NewReader(buf.Bytes())
	dirs, parseErr := p.Parse(r)

	// Should return error for excessive block size
	if parseErr == nil {
		t.Error("Parse() should return error for excessive block size")
	}

	errs := parseErr.Unwrap()
	found := false
	for _, err := range errs {
		if err != nil && bytes.Contains([]byte(err.Error()), []byte("exceeds maximum")) {
			found = true
			break
		}
	}

	if !found {
		t.Error("Parse() error should mention exceeding maximum block size")
	}

	// Should return no directories due to error
	if len(dirs) != 0 {
		t.Errorf("Parse() with excessive block size returned %d directories, want 0", len(dirs))
	}
}

// TestParser_Parse_PictureTypes tests different FLAC picture types
func TestParser_Parse_PictureTypes(t *testing.T) {
	testCases := []struct {
		pictureType uint32
		wantType    string
	}{
		{0, "Other"},
		{3, "Cover (front)"},
		{4, "Cover (back)"},
		{17, "A Bright Colored Fish"},
		{99, "Unknown (99)"},
	}

	for _, tc := range testCases {
		t.Run(tc.wantType, func(t *testing.T) {
			p := New()

			var buf bytes.Buffer
			// FLAC marker
			buf.WriteString("fLaC")

			// Minimal STREAMINFO block (required first block)
			buf.WriteByte(0x00) // Not last, type = 0 (STREAMINFO)
			buf.WriteByte(0x00) // Length = 34 bytes
			buf.WriteByte(0x00)
			buf.WriteByte(0x22)
			streamInfo := make([]byte, 34)
			buf.Write(streamInfo)

			// Build PICTURE block
			var picBuf bytes.Buffer
			// Picture type (4 bytes)
			picBuf.WriteByte(byte(tc.pictureType >> 24))
			picBuf.WriteByte(byte(tc.pictureType >> 16))
			picBuf.WriteByte(byte(tc.pictureType >> 8))
			picBuf.WriteByte(byte(tc.pictureType))
			// MIME type length and value
			mimeType := "image/png"
			picBuf.WriteByte(0x00)
			picBuf.WriteByte(0x00)
			picBuf.WriteByte(0x00)
			picBuf.WriteByte(byte(len(mimeType)))
			picBuf.WriteString(mimeType)
			// Description length (0)
			picBuf.WriteByte(0x00)
			picBuf.WriteByte(0x00)
			picBuf.WriteByte(0x00)
			picBuf.WriteByte(0x00)
			// Width, height, depth, colors
			picBuf.Write([]byte{0x00, 0x00, 0x00, 0x64}) // width = 100
			picBuf.Write([]byte{0x00, 0x00, 0x00, 0x64}) // height = 100
			picBuf.Write([]byte{0x00, 0x00, 0x00, 0x18}) // depth = 24
			picBuf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // colors = 0
			// Picture data length
			picBuf.WriteByte(0x00)
			picBuf.WriteByte(0x00)
			picBuf.WriteByte(0x00)
			picBuf.WriteByte(0x00)

			picData := picBuf.Bytes()

			// PICTURE block header (last block, type = 6)
			buf.WriteByte(0x86)
			buf.WriteByte(byte(len(picData) >> 16))
			buf.WriteByte(byte(len(picData) >> 8))
			buf.WriteByte(byte(len(picData)))
			buf.Write(picData)

			r := bytes.NewReader(buf.Bytes())
			dirs, _ := p.Parse(r)

			// Find PICTURE directory
			var picDir *parser.Directory
			for i := range dirs {
				if dirs[i].Name == "FLAC-Picture" {
					picDir = &dirs[i]
					break
				}
			}

			if picDir == nil {
				t.Fatal("Parse() did not return FLAC-PICTURE directory")
			}

			// Verify picture type tag
			var typeTag *parser.Tag
			for i := range picDir.Tags {
				if picDir.Tags[i].Name == "PictureType" {
					typeTag = &picDir.Tags[i]
					break
				}
			}

			if typeTag == nil {
				t.Fatal("FLAC-PICTURE directory missing PictureType tag")
			}

			if typeTag.Value != tc.wantType {
				t.Errorf("PictureType = %v, want %v", typeTag.Value, tc.wantType)
			}
		})
	}
}

// TestParser_Parse_SeekTable tests FLAC with seek table metadata
func TestParser_Parse_SeekTable(t *testing.T) {
	p := New()

	var buf bytes.Buffer
	// FLAC marker
	buf.WriteString("fLaC")

	// Minimal STREAMINFO block
	buf.WriteByte(0x00) // Not last, type = 0
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x22) // 34 bytes
	streamInfo := make([]byte, 34)
	buf.Write(streamInfo)

	// SEEKTABLE block with 2 seek points (2 * 18 = 36 bytes)
	buf.WriteByte(0x83) // Last block, type = 3 (SEEKTABLE)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x24) // 36 bytes
	seekTable := make([]byte, 36)
	buf.Write(seekTable)

	r := bytes.NewReader(buf.Bytes())
	dirs, _ := p.Parse(r)

	// Find SEEKTABLE directory
	var seekDir *parser.Directory
	for i := range dirs {
		if dirs[i].Name == "FLAC-SeekTable" {
			seekDir = &dirs[i]
			break
		}
	}

	if seekDir == nil {
		t.Fatal("Parse() did not return FLAC-SEEKTABLE directory")
	}

	// Verify seek points tag
	var pointsTag *parser.Tag
	for i := range seekDir.Tags {
		if seekDir.Tags[i].Name == "SeekPoints" {
			pointsTag = &seekDir.Tags[i]
			break
		}
	}

	if pointsTag == nil {
		t.Fatal("FLAC-SEEKTABLE directory missing SeekPoints tag")
	}

	if pointsTag.Value != int64(2) {
		t.Errorf("SeekPoints = %v, want %v", pointsTag.Value, int64(2))
	}
}

// TestParser_Parse_CueSheet tests FLAC with cue sheet metadata
func TestParser_Parse_CueSheet(t *testing.T) {
	p := New()

	var buf bytes.Buffer
	// FLAC marker
	buf.WriteString("fLaC")

	// Minimal STREAMINFO block
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x22)
	streamInfo := make([]byte, 34)
	buf.Write(streamInfo)

	// CUESHEET block (256 bytes)
	buf.WriteByte(0x85) // Last block, type = 5 (CUESHEET)
	buf.WriteByte(0x00)
	buf.WriteByte(0x01)
	buf.WriteByte(0x00) // 256 bytes
	cueSheet := make([]byte, 256)
	buf.Write(cueSheet)

	r := bytes.NewReader(buf.Bytes())
	dirs, _ := p.Parse(r)

	// Find CUESHEET directory
	var cueDir *parser.Directory
	for i := range dirs {
		if dirs[i].Name == "FLAC-CueSheet" {
			cueDir = &dirs[i]
			break
		}
	}

	if cueDir == nil {
		t.Fatal("Parse() did not return FLAC-CUESHEET directory")
	}

	// Verify cue sheet size tag
	var sizeTag *parser.Tag
	for i := range cueDir.Tags {
		if cueDir.Tags[i].Name == "CueSheetSize" {
			sizeTag = &cueDir.Tags[i]
			break
		}
	}

	if sizeTag == nil {
		t.Fatal("FLAC-CUESHEET directory missing CueSheetSize tag")
	}

	if sizeTag.Value != "256 bytes" {
		t.Errorf("CueSheetSize = %v, want %v", sizeTag.Value, "256 bytes")
	}
}

// TestParser_Parse_AllBlockTypes tests FLAC with all metadata block types
func TestParser_Parse_AllBlockTypes(t *testing.T) {
	p := New()

	var buf bytes.Buffer
	// FLAC marker
	buf.WriteString("fLaC")

	// STREAMINFO (type 0, required first)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x22)
	buf.Write(make([]byte, 34))

	// PADDING (type 1)
	buf.WriteByte(0x01)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x10) // 16 bytes
	buf.Write(make([]byte, 16))

	// APPLICATION (type 2)
	buf.WriteByte(0x02)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x08) // 8 bytes
	buf.WriteString("TEST")
	buf.Write(make([]byte, 4))

	// SEEKTABLE (type 3)
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x12) // 18 bytes (1 seek point)
	buf.Write(make([]byte, 18))

	// VORBIS_COMMENT (type 4)
	var vorbisBuf bytes.Buffer
	// Vendor length + vendor string
	vorbisBuf.WriteByte(0x04)
	vorbisBuf.WriteByte(0x00)
	vorbisBuf.WriteByte(0x00)
	vorbisBuf.WriteByte(0x00)
	vorbisBuf.WriteString("TEST")
	// Number of comments
	vorbisBuf.WriteByte(0x00)
	vorbisBuf.WriteByte(0x00)
	vorbisBuf.WriteByte(0x00)
	vorbisBuf.WriteByte(0x00)
	vorbisData := vorbisBuf.Bytes()
	buf.WriteByte(0x04)
	buf.WriteByte(byte(len(vorbisData) >> 16))
	buf.WriteByte(byte(len(vorbisData) >> 8))
	buf.WriteByte(byte(len(vorbisData)))
	buf.Write(vorbisData)

	// CUESHEET (type 5)
	buf.WriteByte(0x05)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x20) // 32 bytes
	buf.Write(make([]byte, 32))

	// PICTURE (type 6, last block)
	var picBuf bytes.Buffer
	picBuf.Write([]byte{0x00, 0x00, 0x00, 0x03}) // type = 3 (front cover)
	picBuf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // MIME length = 0
	picBuf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // description length = 0
	picBuf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // width
	picBuf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // height
	picBuf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // depth
	picBuf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // colors
	picBuf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // picture data length
	picData := picBuf.Bytes()
	buf.WriteByte(0x86) // Last block
	buf.WriteByte(byte(len(picData) >> 16))
	buf.WriteByte(byte(len(picData) >> 8))
	buf.WriteByte(byte(len(picData)))
	buf.Write(picData)

	r := bytes.NewReader(buf.Bytes())
	dirs, _ := p.Parse(r)

	// Verify all block types are present
	expectedDirs := []string{
		"FLAC-StreamInfo",
		"FLAC-Padding",
		"FLAC-Application",
		"FLAC-SeekTable",
		"FLAC-Vorbis",
		"FLAC-CueSheet",
		"FLAC-Picture",
	}

	if len(dirs) != len(expectedDirs) {
		t.Errorf("Parse() returned %d directories, want %d", len(dirs), len(expectedDirs))
	}

	for _, expectedName := range expectedDirs {
		found := false
		for _, dir := range dirs {
			if dir.Name == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Parse() missing directory: %s", expectedName)
		}
	}
}
