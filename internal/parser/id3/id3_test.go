package id3

import (
	"bytes"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestParser_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "ID3" {
		t.Errorf("Name() = %v, want %v", got, "ID3")
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid ID3v2 header",
			data: []byte{'I', 'D', '3', 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			want: true,
		},
		{
			name: "invalid header",
			data: []byte{'I', 'D', '4', 0x04, 0x00},
			want: false,
		},
		{
			name: "too short",
			data: []byte{'I', 'D'},
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
			p := New()
			r := bytes.NewReader(tt.data)
			if got := p.Detect(r); got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Parse_MinimalTag(t *testing.T) {
	// Create minimal ID3v2.3 tag with one text frame (TIT2 = Title)
	var buf bytes.Buffer

	// ID3v2.3 header
	buf.Write([]byte{'I', 'D', '3'})    // Identifier
	buf.WriteByte(0x03)                 // Version
	buf.WriteByte(0x00)                 // Revision
	buf.WriteByte(0x00)                 // Flags
	buf.Write(encodeSynchsafeInt(0x15)) // Size (21 bytes for one frame)

	// TIT2 frame (Title)
	buf.Write([]byte{'T', 'I', 'T', '2'})                                // Frame ID
	buf.Write([]byte{0x00, 0x00, 0x00, 0x0B})                            // Size (11 bytes)
	buf.Write([]byte{0x00, 0x00})                                        // Flags
	buf.WriteByte(0x00)                                                  // Text encoding (ISO-8859-1)
	buf.Write([]byte{'T', 'e', 's', 't', ' ', 'S', 'o', 'n', 'g', 0x00}) // Text + null terminator

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}

	dir := dirs[0]
	if dir.Name != "ID3v2_3" {
		t.Errorf("Directory name = %v, want ID3v2.3", dir.Name)
	}

	// Should have version tag + title frame
	if len(dir.Tags) < 2 {
		t.Errorf("Parse() got %d tags, want at least 2", len(dir.Tags))
	}

	// Check for title frame
	foundTitle := false
	for _, tag := range dir.Tags {
		if tag.Name == "Title" { // getFrameDescription returns "Title" for "TIT2"
			foundTitle = true
			if tag.Value != "Test Song" {
				t.Errorf("Title value = %v, want 'Test Song'", tag.Value)
			}
		}
	}
	if !foundTitle {
		t.Error("Title frame not found")
	}
}

func TestParser_Parse_EmptyTag(t *testing.T) {
	// Create empty ID3v2.4 tag (only header, no frames)
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})    // Identifier
	buf.WriteByte(0x04)                 // Version
	buf.WriteByte(0x00)                 // Revision
	buf.WriteByte(0x00)                 // Flags
	buf.Write(encodeSynchsafeInt(0x00)) // Size (0 bytes)

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}

	// Should have at least version tag
	if len(dirs[0].Tags) < 1 {
		t.Error("Expected at least version tag")
	}
}

func TestParser_Parse_InvalidHeader(t *testing.T) {
	// Invalid magic bytes
	data := []byte{'I', 'D', '4', 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	p := New()
	r := bytes.NewReader(data)

	dirs, _ := p.Parse(r)
	if len(dirs) != 0 {
		t.Errorf("Parse() with invalid header returned %d directories, want 0", len(dirs))
	}
}

func TestDecodeSynchsafeInt(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want uint32
	}{
		{
			name: "zero",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			want: 0,
		},
		{
			name: "small value",
			data: []byte{0x00, 0x00, 0x00, 0x15},
			want: 21,
		},
		{
			name: "larger value",
			data: []byte{0x00, 0x00, 0x02, 0x01},
			want: 257,
		},
		{
			name: "max 28-bit value",
			data: []byte{0x7F, 0x7F, 0x7F, 0x7F},
			want: 268435455, // 2^28 - 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeSynchsafeInt(tt.data); got != tt.want {
				t.Errorf("decodeSynchsafeInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeTextFrame(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "ISO-8859-1",
			data: append([]byte{0x00}, []byte("Hello")...),
			want: "Hello",
		},
		{
			name: "UTF-8",
			data: append([]byte{0x03}, []byte("Hello 世界")...),
			want: "Hello 世界",
		},
		{
			name: "empty",
			data: []byte{},
			want: "",
		},
		{
			name: "only encoding byte",
			data: []byte{0x00},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeTextFrame(tt.data); got != tt.want {
				t.Errorf("decodeTextFrame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeUTF16LE(t *testing.T) {
	// "Hello" in UTF-16LE
	data := []byte{
		'H', 0x00,
		'e', 0x00,
		'l', 0x00,
		'l', 0x00,
		'o', 0x00,
	}

	got := decodeUTF16LE(data)
	want := "Hello"
	if got != want {
		t.Errorf("decodeUTF16LE() = %v, want %v", got, want)
	}
}

func TestDecodeUTF16BE(t *testing.T) {
	// "Hello" in UTF-16BE
	data := []byte{
		0x00, 'H',
		0x00, 'e',
		0x00, 'l',
		0x00, 'l',
		0x00, 'o',
	}

	got := decodeUTF16BE(data)
	want := "Hello"
	if got != want {
		t.Errorf("decodeUTF16BE() = %v, want %v", got, want)
	}
}

func TestIsTextFrame(t *testing.T) {
	tests := []struct {
		frameID string
		want    bool
	}{
		{"TIT2", true},
		{"TALB", true},
		{"TPE1", true},
		{"TXXX", false}, // User-defined text, special case
		{"APIC", false},
		{"COMM", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.frameID, func(t *testing.T) {
			if got := isTextFrame(tt.frameID); got != tt.want {
				t.Errorf("isTextFrame(%v) = %v, want %v", tt.frameID, got, tt.want)
			}
		})
	}
}

func TestGetFrameDescription(t *testing.T) {
	tests := []struct {
		frameID string
		want    string
	}{
		{"TIT2", "Title"},
		{"TPE1", "Artist"},
		{"TALB", "Album"},
		{"TRCK", "Track Number"},
		{"UNKNOWN", "UNKNOWN"}, // Unknown frame returns itself
	}

	for _, tt := range tests {
		t.Run(tt.frameID, func(t *testing.T) {
			if got := getFrameDescription(tt.frameID); got != tt.want {
				t.Errorf("getFrameDescription(%v) = %v, want %v", tt.frameID, got, tt.want)
			}
		})
	}
}

func TestTrimNull(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{
			name: "no nulls",
			data: []byte("Hello"),
			want: []byte("Hello"),
		},
		{
			name: "trailing nulls",
			data: []byte("Hello\x00\x00"),
			want: []byte("Hello"),
		},
		{
			name: "all nulls",
			data: []byte("\x00\x00\x00"),
			want: []byte{},
		},
		{
			name: "empty",
			data: []byte{},
			want: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimNull(tt.data)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("trimNull() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to encode synchsafe integer for tests
func encodeSynchsafeInt(n uint32) []byte {
	buf := make([]byte, 4)
	buf[0] = byte((n >> 21) & 0x7F)
	buf[1] = byte((n >> 14) & 0x7F)
	buf[2] = byte((n >> 7) & 0x7F)
	buf[3] = byte(n & 0x7F)
	return buf
}

// TestParser_Parse_MultipleFrames tests parsing multiple frames
func TestParser_Parse_MultipleFrames(t *testing.T) {
	var buf bytes.Buffer

	// ID3v2.3 header
	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x2A)) // Size for two frames

	// TIT2 frame (Title)
	buf.Write([]byte{'T', 'I', 'T', '2'})
	buf.Write([]byte{0x00, 0x00, 0x00, 0x0B})
	buf.Write([]byte{0x00, 0x00})
	buf.WriteByte(0x00)
	buf.Write([]byte("Test Song\x00"))

	// TPE1 frame (Artist)
	buf.Write([]byte{'T', 'P', 'E', '1'})
	buf.Write([]byte{0x00, 0x00, 0x00, 0x0C})
	buf.Write([]byte{0x00, 0x00})
	buf.WriteByte(0x00)
	buf.Write([]byte("Test Artist\x00"))

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}

	// Should have version + 2 frames = 3 tags minimum
	if len(dirs[0].Tags) < 3 {
		t.Errorf("Parse() got %d tags, want at least 3", len(dirs[0].Tags))
	}
}

// Ensure Parser implements parser.Parser interface
func TestParser_ImplementsInterface(t *testing.T) {
	var _ parser.Parser = (*Parser)(nil)
}

func TestParser_ConcurrentParse(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x15))
	buf.Write([]byte{'T', 'I', 'T', '2'})
	buf.Write([]byte{0x00, 0x00, 0x00, 0x0B})
	buf.Write([]byte{0x00, 0x00})
	buf.WriteByte(0x00)
	buf.Write([]byte("Test Song\x00"))

	data := buf.Bytes()
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

func TestDecodeCommentFrame(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "ISO-8859-1 comment",
			data: append([]byte{0x00, 'e', 'n', 'g'}, []byte("Hello comment")...),
			want: "Hello comment",
		},
		{
			name: "UTF-8 comment",
			data: append([]byte{0x03, 'e', 'n', 'g'}, []byte("Hello UTF-8")...),
			want: "Hello UTF-8",
		},
		{
			name: "too short",
			data: []byte{0x00, 'e'},
			want: "",
		},
		{
			name: "empty",
			data: []byte{},
			want: "",
		},
		{
			name: "unknown encoding",
			data: append([]byte{0x05, 'e', 'n', 'g'}, []byte("Hello")...),
			want: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeCommentFrame(tt.data); got != tt.want {
				t.Errorf("decodeCommentFrame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeUTF16WithBOM(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "UTF-16LE with BOM",
			data: []byte{0xFF, 0xFE, 'H', 0x00, 'i', 0x00},
			want: "Hi",
		},
		{
			name: "UTF-16BE with BOM",
			data: []byte{0xFE, 0xFF, 0x00, 'H', 0x00, 'i'},
			want: "Hi",
		},
		{
			name: "no BOM assumes LE",
			data: []byte{'H', 0x00, 'i', 0x00},
			want: "Hi",
		},
		{
			name: "too short",
			data: []byte{0xFF},
			want: "",
		},
		{
			name: "empty",
			data: []byte{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeUTF16WithBOM(tt.data); got != tt.want {
				t.Errorf("decodeUTF16WithBOM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeTextFrame_UTF16(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "UTF-16 with BOM (LE)",
			// Note: null bytes in UTF-16 are part of the encoding, test without trailing nulls
			data: []byte{0x01, 0xFF, 0xFE, 'H', 0x00, 'i', 0x00, 0x00, 0x00}, // BOM + "Hi" + null terminator
			want: "Hi",
		},
		{
			name: "UTF-16BE without BOM",
			data: []byte{0x02, 0x00, 'H', 0x00, 'i'},
			want: "Hi",
		},
		{
			name: "unknown encoding falls back",
			data: append([]byte{0x99}, []byte("test")...),
			want: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeTextFrame(tt.data); got != tt.want {
				t.Errorf("decodeTextFrame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeUTF16_OddLength(t *testing.T) {
	// Test odd-length data (should truncate last byte)
	dataLE := []byte{'H', 0x00, 'i', 0x00, 'x'} // 5 bytes
	gotLE := decodeUTF16LE(dataLE)
	if gotLE != "Hi" {
		t.Errorf("decodeUTF16LE() odd length = %v, want Hi", gotLE)
	}

	dataBE := []byte{0x00, 'H', 0x00, 'i', 'x'} // 5 bytes
	gotBE := decodeUTF16BE(dataBE)
	if gotBE != "Hi" {
		t.Errorf("decodeUTF16BE() odd length = %v, want Hi", gotBE)
	}
}

func TestParser_Parse_ID3v22(t *testing.T) {
	// Create ID3v2.2 tag
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x02)                 // Version 2.2
	buf.WriteByte(0x00)                 // Revision
	buf.WriteByte(0x00)                 // Flags
	buf.Write(encodeSynchsafeInt(0x10)) // Size

	// TT2 frame (Title for v2.2)
	buf.Write([]byte{'T', 'T', '2'})    // Frame ID (3 chars)
	buf.Write([]byte{0x00, 0x00, 0x07}) // Size (3 bytes, 24-bit)
	buf.WriteByte(0x00)                 // Encoding
	buf.Write([]byte("Title\x00"))

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}

	if dirs[0].Name != "ID3v2_2" {
		t.Errorf("Directory name = %v, want ID3v2.2", dirs[0].Name)
	}
}

func TestParser_Parse_ID3v24(t *testing.T) {
	// Create ID3v2.4 tag
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x04)                 // Version 2.4
	buf.WriteByte(0x00)                 // Revision
	buf.WriteByte(0x00)                 // Flags
	buf.Write(encodeSynchsafeInt(0x15)) // Size

	// TIT2 frame
	buf.Write([]byte{'T', 'I', 'T', '2'})
	buf.Write(encodeSynchsafeInt(0x0B)) // Synchsafe size for v2.4
	buf.Write([]byte{0x00, 0x00})       // Flags
	buf.WriteByte(0x00)                 // Encoding
	buf.Write([]byte("Test Song\x00"))

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}

	if dirs[0].Name != "ID3v2_4" {
		t.Errorf("Directory name = %v, want ID3v2.4", dirs[0].Name)
	}
}

func TestParser_Parse_WithFlags(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x04)
	buf.WriteByte(0x00)
	buf.WriteByte(0x70) // ExtHeader + Experimental + Footer flags
	buf.Write(encodeSynchsafeInt(0x10))
	// Extended header (minimal)
	buf.Write(encodeSynchsafeInt(0x06)) // Size of extended header
	buf.WriteByte(0x01)                 // Number of flag bytes
	buf.WriteByte(0x00)                 // No extended flags

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, _ := p.Parse(r)

	// Should have parsed with flags
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}

	// Check flags tag exists
	foundFlags := false
	for _, tag := range dirs[0].Tags {
		if tag.Name == "Flags" {
			foundFlags = true
		}
	}
	if !foundFlags {
		t.Error("Flags tag not found")
	}
}

func TestParser_Parse_PictureFrame(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x20))

	// APIC frame
	buf.Write([]byte{'A', 'P', 'I', 'C'})
	buf.Write([]byte{0x00, 0x00, 0x00, 0x10}) // Size
	buf.Write([]byte{0x00, 0x00})             // Flags
	buf.Write(make([]byte, 0x10))             // Fake picture data

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check for picture frame
	foundPic := false
	for _, tag := range dirs[0].Tags {
		if tag.Name == "Attached Picture" {
			foundPic = true
			if tag.DataType != "binary" {
				t.Errorf("Picture DataType = %v, want binary", tag.DataType)
			}
		}
	}
	if !foundPic {
		t.Error("Picture frame not found")
	}
}

func TestParser_Parse_CommentFrame(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x20))

	// COMM frame
	buf.Write([]byte{'C', 'O', 'M', 'M'})
	buf.Write([]byte{0x00, 0x00, 0x00, 0x10}) // Size
	buf.Write([]byte{0x00, 0x00})             // Flags
	buf.WriteByte(0x00)                       // Encoding
	buf.Write([]byte("eng"))                  // Language
	buf.Write([]byte("Test comment\x00"))     // Comment

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check for comment frame
	foundComment := false
	for _, tag := range dirs[0].Tags {
		if tag.Name == "Comment" {
			foundComment = true
		}
	}
	if !foundComment {
		t.Error("Comment frame not found")
	}
}

func TestParser_Parse_BinaryFrame(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x20))

	// PRIV frame (generic binary)
	buf.Write([]byte{'P', 'R', 'I', 'V'})
	buf.Write([]byte{0x00, 0x00, 0x00, 0x10}) // Size
	buf.Write([]byte{0x00, 0x00})             // Flags
	buf.Write(make([]byte, 0x10))             // Private data

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check for private frame
	foundPriv := false
	for _, tag := range dirs[0].Tags {
		if tag.Name == "Private" {
			foundPriv = true
			if tag.DataType != "binary" {
				t.Errorf("Private DataType = %v, want binary", tag.DataType)
			}
		}
	}
	if !foundPriv {
		t.Error("Private frame not found")
	}
}

func TestParser_Parse_InvalidFrameSize(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x20))

	// Frame with size 0
	buf.Write([]byte{'T', 'I', 'T', '2'})
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Size = 0 (invalid)
	buf.Write([]byte{0x00, 0x00})

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, parseErr := p.Parse(r)
	if parseErr == nil {
		t.Error("Expected error for invalid frame size")
	}

	// Should still return directory with version tag
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestParser_Parse_ExtendedHeaderError(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x04)
	buf.WriteByte(0x00)
	buf.WriteByte(0x40) // Extended header flag
	buf.Write(encodeSynchsafeInt(0x10))
	// No extended header data - will cause read error

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, parseErr := p.Parse(r)
	if parseErr == nil {
		t.Error("Expected error for missing extended header")
	}

	// Should return directory anyway
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestParser_Parse_ReadError(t *testing.T) {
	// Empty data - can't even read header
	p := New()
	r := bytes.NewReader([]byte{})

	dirs, parseErr := p.Parse(r)
	if parseErr == nil {
		t.Error("Expected error for empty data")
	}
	if len(dirs) != 0 {
		t.Errorf("Parse() got %d directories, want 0", len(dirs))
	}
}

func TestDecodeCommentFrame_UTF16(t *testing.T) {
	// UTF-16 with BOM - null terminator at end
	data := []byte{0x01, 'e', 'n', 'g', 0xFF, 0xFE, 'H', 0x00, 'i', 0x00, 0x00, 0x00}
	got := decodeCommentFrame(data)
	if got != "Hi" {
		t.Errorf("decodeCommentFrame UTF-16 LE = %v, want Hi", got)
	}

	// UTF-16BE
	data2 := []byte{0x02, 'e', 'n', 'g', 0x00, 'H', 0x00, 'i'}
	got2 := decodeCommentFrame(data2)
	if got2 != "Hi" {
		t.Errorf("decodeCommentFrame UTF-16 BE = %v, want Hi", got2)
	}
}

func TestParser_Parse_PaddingFrame(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x20))

	// Frame with null ID (padding)
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Null frame ID
	buf.Write(make([]byte, 0x1C))             // Padding

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, _ := p.Parse(r)
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestParser_Parse_FrameReadError(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x100)) // Large size

	// Only write partial frame data (will cause read error)
	buf.Write([]byte{'T', 'I', 'T', '2'})

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, _ := p.Parse(r)
	// Should still return directory
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestParser_Parse_ID3v22_ReadError(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x02) // Version 2.2
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x50))

	// Write frame ID but truncate size
	buf.Write([]byte{'T', 'T', '2'})

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, _ := p.Parse(r)
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestDecodeUTF16BE_Empty(t *testing.T) {
	got := decodeUTF16BE([]byte{})
	if got != "" {
		t.Errorf("decodeUTF16BE empty = %v, want empty", got)
	}
}

func TestReadSynchsafeInt(t *testing.T) {
	data := []byte{0x00, 0x00, 0x02, 0x01}
	r := bytes.NewReader(data)

	got, err := readSynchsafeInt(r, 0, 4)
	if err != nil {
		t.Fatalf("readSynchsafeInt() error = %v", err)
	}
	if got != 257 {
		t.Errorf("readSynchsafeInt() = %v, want 257", got)
	}
}

func TestReadSynchsafeInt_Error(t *testing.T) {
	r := bytes.NewReader([]byte{0x00})

	_, err := readSynchsafeInt(r, 0, 4)
	if err == nil {
		t.Error("readSynchsafeInt() expected error for short data")
	}
}

func TestParser_Parse_ID3v24_SynchsafeError(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x04) // Version 2.4
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x50))

	// Write frame ID but truncate size (synchsafe for v2.4)
	buf.Write([]byte{'T', 'I', 'T', '2'})

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, _ := p.Parse(r)
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestParser_Parse_FrameNil(t *testing.T) {
	// This tests the case where parseFrame returns nil
	// Create tag with just padding after header
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x10))

	// Write null byte at start of frame area (padding)
	buf.Write(make([]byte, 0x10))

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, _ := p.Parse(r)
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestDecodeUTF16BE_SingleByte(t *testing.T) {
	// Odd length - should truncate
	got := decodeUTF16BE([]byte{0x00})
	if got != "" {
		t.Errorf("decodeUTF16BE single byte = %v, want empty", got)
	}
}

func TestDecodeUTF16BE_WithNullTerminator(t *testing.T) {
	// UTF-16BE with null terminator in the middle
	data := []byte{0x00, 'H', 0x00, 'i', 0x00, 0x00} // "Hi" + null terminator
	got := decodeUTF16BE(data)
	if got != "Hi" {
		t.Errorf("decodeUTF16BE with null = %v, want Hi", got)
	}
}

func TestParser_Parse_FrameIDReadError(t *testing.T) {
	// Tag with size that says there are frames, but no frame data
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x100)) // Large size
	// No frame data at all

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, _ := p.Parse(r)
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestParser_Parse_FrameFlagsReadError(t *testing.T) {
	// Frame header present but missing flags bytes (v2.3 path)
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x10))

	// Frame ID and size, but omit flags and data
	buf.Write([]byte{'T', 'I', 'T', '2'})
	buf.Write([]byte{0x00, 0x00, 0x00, 0x01}) // size = 1
	// no flags/data written

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, _ := p.Parse(r)
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestParser_Parse_V22SizeReadError(t *testing.T) {
	// ID3v2.2 tag with frame ID but no size bytes
	var buf bytes.Buffer

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x02) // Version 2.2
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x10))
	buf.Write([]byte{'T', 'T', '2'}) // Frame ID only, no size

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, _ := p.Parse(r)
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestParser_Parse_TooManyFrames(t *testing.T) {
	// Build v2.3 tag with 4,097 tiny frames to hit maxFrameCount guard
	var buf bytes.Buffer

	frameCount := maxFrameCount + 1
	frameSize := 1               // 1 byte payload
	frameHeaderSize := 4 + 4 + 2 // ID + size + flags
	tagPayloadSize := frameCount * (frameHeaderSize + frameSize)

	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03) // Version 2.3
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(uint32(tagPayloadSize)))

	for i := 0; i < frameCount; i++ {
		buf.Write([]byte{'T', 'I', 'T', '2'})
		buf.Write([]byte{0x00, 0x00, 0x00, byte(frameSize)}) // size = 1
		buf.Write([]byte{0x00, 0x00})                        // flags
		buf.WriteByte('A')                                   // payload
	}

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, parseErr := p.Parse(r)

	if parseErr == nil {
		t.Fatal("expected parseErr for too many frames")
	}
	if len(dirs) != 1 {
		t.Fatalf("Parse() got %d directories, want 1", len(dirs))
	}
}

func TestParser_Parse_CommentFrame_WithDescription(t *testing.T) {
	var buf bytes.Buffer

	// COMM frame with proper structure: encoding + lang + description\0 + comment\0
	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x30))

	// COMM frame: encoding + lang + desc\0 + comment\0
	commFrame := []byte{
		0x00,                                   // Encoding: ISO-8859-1
		'e', 'n', 'g',                          // Language
		'S', 'h', 'o', 'r', 't', 0x00,          // Description: "Short\0"
		'T', 'h', 'i', 's', ' ', 'i', 's', ' ', // Comment text
		't', 'h', 'e', ' ', 'c', 'o', 'm', 'm',
		'e', 'n', 't', 0x00,
	}

	buf.Write([]byte{'C', 'O', 'M', 'M'})
	buf.Write([]byte{0x00, 0x00, 0x00, byte(len(commFrame))})
	buf.Write([]byte{0x00, 0x00}) // Flags
	buf.Write(commFrame)

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check for comment frame with correct value
	foundComment := false
	for _, tag := range dirs[0].Tags {
		if tag.Name == "Comment" {
			foundComment = true
			if tag.Value != "This is the comment" {
				t.Errorf("Comment value = %q, want %q", tag.Value, "This is the comment")
			}
		}
	}
	if !foundComment {
		t.Error("Comment frame not found")
	}
}

func TestParser_Parse_UserTextFrame(t *testing.T) {
	var buf bytes.Buffer

	// TXXX frame with structure: encoding + description\0 + value\0
	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x03)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x30))

	// TXXX frame: encoding + desc\0 + value\0
	txxxFrame := []byte{
		0x00,                                    // Encoding: ISO-8859-1
		'M', 'y', 'D', 'e', 's', 'c', 0x00,      // Description: "MyDesc\0"
		'M', 'y', 'V', 'a', 'l', 'u', 'e', 0x00, // Value: "MyValue\0"
	}

	buf.Write([]byte{'T', 'X', 'X', 'X'})
	buf.Write([]byte{0x00, 0x00, 0x00, byte(len(txxxFrame))})
	buf.Write([]byte{0x00, 0x00}) // Flags
	buf.Write(txxxFrame)

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check for user text frame with correct value
	foundTXXX := false
	for _, tag := range dirs[0].Tags {
		if tag.Name == "User Defined Text" {
			foundTXXX = true
			if tag.Value != "MyValue" {
				t.Errorf("User Defined Text value = %q, want %q", tag.Value, "MyValue")
			}
		}
	}
	if !foundTXXX {
		t.Error("User Defined Text frame not found")
	}
}

func TestParser_Parse_ID3v22_CommentFrame(t *testing.T) {
	var buf bytes.Buffer

	// ID3v2.2 COM frame
	buf.Write([]byte{'I', 'D', '3'})
	buf.WriteByte(0x02) // Version 2.2
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.Write(encodeSynchsafeInt(0x20))

	// COM frame (3-char ID for v2.2): encoding + lang + comment
	comFrame := []byte{
		0x00,                              // Encoding: ISO-8859-1
		'e', 'n', 'g',                     // Language
		'H', 'e', 'l', 'l', 'o', 0x00,     // Comment: "Hello\0"
	}

	buf.Write([]byte{'C', 'O', 'M'})
	buf.Write([]byte{0x00, 0x00, byte(len(comFrame))}) // 3-byte size for v2.2
	buf.Write(comFrame)

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check for comment frame
	foundComment := false
	for _, tag := range dirs[0].Tags {
		if tag.Name == "Comment" {
			foundComment = true
			if tag.Value != "Hello" {
				t.Errorf("Comment value = %q, want %q", tag.Value, "Hello")
			}
		}
	}
	if !foundComment {
		t.Error("ID3v2.2 Comment frame not found")
	}
}

func TestDecodeUserTextFrame(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "ISO-8859-1 with description",
			data: []byte{0x00, 'd', 'e', 's', 'c', 0x00, 'v', 'a', 'l', 'u', 'e', 0x00},
			want: "value",
		},
		{
			name: "ISO-8859-1 no description",
			data: []byte{0x00, 0x00, 'v', 'a', 'l', 'u', 'e', 0x00},
			want: "value",
		},
		{
			name: "UTF-8 with description",
			data: []byte{0x03, 'd', 'e', 's', 'c', 0x00, 'v', 'a', 'l', 'u', 'e', 0x00},
			want: "value",
		},
		{
			name: "too short",
			data: []byte{0x00},
			want: "",
		},
		{
			name: "empty",
			data: []byte{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeUserTextFrame(tt.data)
			if got != tt.want {
				t.Errorf("decodeUserTextFrame() = %q, want %q", got, tt.want)
			}
		})
	}
}
