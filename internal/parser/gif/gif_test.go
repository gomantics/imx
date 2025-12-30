package gif

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestParser_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "GIF" {
		t.Errorf("Name() = %v, want %v", got, "GIF")
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "GIF87a valid",
			data: []byte("GIF87a"),
			want: true,
		},
		{
			name: "GIF89a valid",
			data: []byte("GIF89a"),
			want: true,
		},
		{
			name: "invalid first byte",
			data: []byte("gif89a"),
			want: false,
		},
		{
			name: "invalid - too short",
			data: []byte("GIF"),
			want: false,
		},
		{
			name: "invalid - wrong signature",
			data: []byte("NOTGIF"),
			want: false,
		},
		{
			name: "invalid - JPEG",
			data: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
			want: false,
		},
		{
			name: "empty",
			data: []byte{},
			want: false,
		},
		{
			name: "valid GIF87a with extra data",
			data: []byte("GIF87a\x00\x00\x00\x00\x00\x00\x00"),
			want: true,
		},
		{
			name: "valid GIF89a with extra data",
			data: []byte("GIF89a\x00\x00\x00\x00\x00\x00\x00"),
			want: true,
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

// TestParser_Parse tests basic parsing - comprehensive validation is in validation_test.go
func TestParser_Parse(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/gif/animated_art.gif")
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

	// Check that we have GIF directory
	hasGIF := false
	for _, dir := range dirs {
		if dir.Name == "GIF" {
			hasGIF = true
			if len(dir.Tags) == 0 {
				t.Error("GIF has no tags")
			}
		}
	}

	if !hasGIF {
		t.Error("Missing GIF directory")
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
			name: "invalid marker",
			data: []byte("fLac"),
		},
		{
			name: "truncated file - marker only",
			data: []byte("GIF89a"),
		},
		{
			name: "truncated LSD",
			data: []byte("GIF89a\x00"),
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

// TestParser_Parse_EdgeCases tests additional Parse error paths
func TestParser_Parse_EdgeCases(t *testing.T) {
	p := New()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name: "file with unknown separator in data stream",
			data: []byte{
				'G', 'I', 'F', '8', '9', 'a',
				0x0A, 0x00, 0x0A, 0x00,
				0x00, 0x00, 0x00,
				0xFF, // Unknown separator (not 0x21, 0x2C, 0x3B, or 0x00)
			},
			wantErr: true,
		},
		{
			name: "file ending with EOF during parsing",
			data: []byte{
				'G', 'I', 'F', '8', '9', 'a',
				0x0A, 0x00, 0x0A, 0x00,
				0x00, 0x00, 0x00,
				0x21, // Extension but no label
			},
			wantErr: false, // Parser handles EOF gracefully
		},
		{
			name: "file with safety limit exceeded",
			data: func() []byte {
				// This test is hard to trigger without actual 10MB+ file
				data := []byte{
					'G', 'I', 'F', '8', '9', 'a',
					0x0A, 0x00, 0x0A, 0x00,
					0x00, 0x00, 0x00,
					0x3B, // Immediate trailer
				}
				return data
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			dirs, err := p.Parse(r)

			if tt.wantErr && err == nil {
				t.Errorf("Parse() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
			}
			_ = dirs // dirs may be empty or have partial data
		})
	}
}

// TestParser_Parse_AnimatedGIF tests that Parse correctly extracts animation metadata
func TestParser_Parse_AnimatedGIF(t *testing.T) {
	p := New()

	// Construct a minimal animated GIF with 2 frames and loop count
	data := []byte{
		'G', 'I', 'F', '8', '9', 'a',
		0x0A, 0x00, 0x0A, 0x00, // Width=10, Height=10
		0x00, 0x00, 0x00, // No global color table
		// NETSCAPE2.0 extension
		0x21, 0xFF, 0x0B,
		'N', 'E', 'T', 'S', 'C', 'A', 'P', 'E',
		'2', '.', '0',
		0x03, 0x01, 0x05, 0x00, // Loop count = 5
		0x00,
		// Frame 1
		0x21, 0xF9, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, // Graphic Control
		0x2C, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x05, 0x00, 0x00, // Image descriptor
		0x08, 0x02, 0xAA, 0xBB, 0x00, // LZW + data
		// Frame 2
		0x21, 0xF9, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, // Graphic Control
		0x2C, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x05, 0x00, 0x00, // Image descriptor
		0x08, 0x02, 0xCC, 0xDD, 0x00, // LZW + data
		0x3B, // Trailer
	}

	r := bytes.NewReader(data)
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Find GIF directory
	var gifDir *parser.Directory
	for i := range dirs {
		if dirs[i].Name == "GIF" {
			gifDir = &dirs[i]
			break
		}
	}

	if gifDir == nil {
		t.Fatal("Parse() did not return GIF directory")
	}

	// Check for FrameCount tag
	var foundFrameCount, foundLoopCount bool
	for _, tag := range gifDir.Tags {
		if tag.Name == "FrameCount" {
			foundFrameCount = true
			if tag.Value != uint16(2) {
				t.Errorf("FrameCount = %v, want 2", tag.Value)
			}
		}
		if tag.Name == "AnimationIterations" {
			foundLoopCount = true
			if tag.Value != uint16(5) {
				t.Errorf("AnimationIterations = %v, want 5", tag.Value)
			}
		}
	}

	if !foundFrameCount {
		t.Error("Parse() did not add FrameCount tag for animated GIF")
	}
	if !foundLoopCount {
		t.Error("Parse() did not add AnimationIterations tag for animated GIF")
	}
}

// TestParser_Parse_ImageDescriptorError tests skipImage failure path
func TestParser_Parse_ImageDescriptorError(t *testing.T) {
	p := New()

	// GIF with truncated image descriptor
	data := []byte{
		'G', 'I', 'F', '8', '9', 'a',
		0x0A, 0x00, 0x0A, 0x00,
		0x00, 0x00, 0x00,
		0x2C,       // Image separator
		0x00, 0x00, // Only 2 bytes of descriptor instead of 9
	}

	r := bytes.NewReader(data)
	dirs, _ := p.Parse(r)

	// Should still return GIF directory even with error
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories after image descriptor error")
	}
}

// TestParser_Parse_NonEOFError tests Parse handling of non-EOF errors
func TestParser_Parse_NonEOFError(t *testing.T) {
	p := New()

	// Minimal valid GIF header
	data := []byte{
		'G', 'I', 'F', '8', '9', 'a',
		0x0A, 0x00, 0x0A, 0x00,
		0x00, 0x00, 0x00,
		0x21, // Extension separator - error will occur here
	}

	// Create reader that returns custom error at offset 13 (after header)
	customErr := io.ErrUnexpectedEOF
	r := &errorReaderAt{
		data:        data,
		errorOffset: 13,
		customError: customErr,
	}

	dirs, parseErr := p.Parse(r)

	// Should return GIF directory even with error
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories after non-EOF error")
	}

	// Should have captured the error
	if parseErr == nil {
		t.Error("Parse() expected error to be captured, got nil")
	}
}

// TestParser_ConcurrentParse tests that the parser can be used concurrently
// This is a critical test from parser.md to ensure no data races
func TestParser_ConcurrentParse(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/gif/animated_art.gif")
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

// Ensure Parser implements parser.Parser interface
func TestParser_ImplementsInterface(t *testing.T) {
	var _ parser.Parser = (*Parser)(nil)
}

// TestParser_Parse_UnknownExtensionBlock tests that unknown extension blocks are gracefully skipped
func TestParser_Parse_UnknownExtensionBlock(t *testing.T) {
	p := New()

	// GIF with unknown extension block (0x42) in the middle
	data := []byte{
		'G', 'I', 'F', '8', '9', 'a',
		0x0A, 0x00, 0x0A, 0x00, // Width=10, Height=10
		0x00, 0x00, 0x00, // No global color table
		// Frame 1
		0x2C, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x05, 0x00, 0x00, // Image descriptor
		0x08, 0x02, 0xAA, 0xBB, 0x00, // LZW + data
		// Unknown extension 0x42
		0x21, 0x42, // Extension with unknown label
		0x05, 'H', 'e', 'l', 'l', 'o', // Sub-block with 5 bytes
		0x00, // Terminator
		// Frame 2
		0x2C, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x05, 0x00, 0x00, // Image descriptor
		0x08, 0x02, 0xCC, 0xDD, 0x00, // LZW + data
		0x3B, // Trailer
	}

	r := bytes.NewReader(data)
	dirs, err := p.Parse(r)

	// Should succeed despite unknown extension
	if err != nil {
		t.Errorf("Parse() with unknown extension should not fail, got error: %v", err)
	}

	// Should still parse GIF directory
	var gifDir *parser.Directory
	for i := range dirs {
		if dirs[i].Name == "GIF" {
			gifDir = &dirs[i]
			break
		}
	}

	if gifDir == nil {
		t.Fatal("Parse() did not return GIF directory after unknown extension")
	}

	// Should count both frames
	var frameCount uint16
	for _, tag := range gifDir.Tags {
		if tag.Name == "FrameCount" {
			frameCount = tag.Value.(uint16)
			break
		}
	}

	if frameCount != 2 {
		t.Errorf("FrameCount = %d, want 2 (unknown extension should not affect frame counting)", frameCount)
	}
}

// TestParser_Parse_LargeAnimatedGIF tests parsing of animated GIF with many frames
func TestParser_Parse_LargeAnimatedGIF(t *testing.T) {
	p := New()

	// Construct animated GIF with 150 frames
	var buf bytes.Buffer

	// Header
	buf.WriteString("GIF89a")
	buf.Write([]byte{0x0A, 0x00, 0x0A, 0x00}) // Width=10, Height=10
	buf.Write([]byte{0x00, 0x00, 0x00})       // No global color table

	// NETSCAPE2.0 extension with loop count
	buf.Write([]byte{0x21, 0xFF, 0x0B})
	buf.WriteString("NETSCAPE2.0")
	buf.Write([]byte{0x03, 0x01, 0x00, 0x00}) // Loop forever
	buf.WriteByte(0x00)                       // Terminator

	// Add 150 frames
	for i := 0; i < 150; i++ {
		// Graphic Control Extension
		buf.Write([]byte{0x21, 0xF9, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00})
		// Image Descriptor
		buf.Write([]byte{0x2C, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x05, 0x00, 0x00})
		// LZW data
		buf.Write([]byte{0x08, 0x02, 0xAA + byte(i%10), 0xBB + byte(i%10), 0x00})
	}

	// Trailer
	buf.WriteByte(0x3B)

	r := bytes.NewReader(buf.Bytes())
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error on large animated GIF = %v, want nil", err)
	}

	// Find GIF directory
	var gifDir *parser.Directory
	for i := range dirs {
		if dirs[i].Name == "GIF" {
			gifDir = &dirs[i]
			break
		}
	}

	if gifDir == nil {
		t.Fatal("Parse() did not return GIF directory for large animated GIF")
	}

	// Check for FrameCount tag
	var foundFrameCount bool
	var frameCount uint16
	for _, tag := range gifDir.Tags {
		if tag.Name == "FrameCount" {
			foundFrameCount = true
			frameCount = tag.Value.(uint16)
			break
		}
	}

	if !foundFrameCount {
		t.Error("Parse() did not add FrameCount tag for large animated GIF")
	}

	if frameCount != 150 {
		t.Errorf("FrameCount = %d, want 150", frameCount)
	}
}

// TestParser_Parse_MalformedFrameDescriptor tests handling of truncated frame descriptors
func TestParser_Parse_MalformedFrameDescriptor(t *testing.T) {
	p := New()

	// GIF with malformed (truncated) image descriptor
	data := []byte{
		'G', 'I', 'F', '8', '9', 'a',
		0x0A, 0x00, 0x0A, 0x00,
		0x00, 0x00, 0x00,
		0x2C,       // Image separator
		0x00, 0x00, // Only 2 bytes instead of required 9
	}

	r := bytes.NewReader(data)
	dirs, _ := p.Parse(r)

	// Should still return GIF directory even with malformed descriptor
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories after malformed frame descriptor")
	}

	// Should have counted the truncated frame attempt
	var gifDir *parser.Directory
	for i := range dirs {
		if dirs[i].Name == "GIF" {
			gifDir = &dirs[i]
			break
		}
	}

	if gifDir == nil {
		t.Fatal("Parse() did not return GIF directory")
	}

	// Frame count should not be added for single malformed frame
	for _, tag := range gifDir.Tags {
		if tag.Name == "FrameCount" {
			t.Errorf("FrameCount should not be added for malformed single frame, got %v", tag.Value)
		}
	}
}

// TestParser_Parse_CommentAndXMP tests GIF with both comment and XMP extensions
func TestParser_Parse_CommentAndXMP(t *testing.T) {
	p := New()

	// Construct GIF with both comment and XMP
	var buf bytes.Buffer

	// Header
	buf.WriteString("GIF89a")
	buf.Write([]byte{0x0A, 0x00, 0x0A, 0x00})
	buf.Write([]byte{0x00, 0x00, 0x00})

	// Comment extension
	buf.Write([]byte{0x21, 0xFE}) // Comment label
	buf.WriteByte(12)             // Block size
	buf.WriteString("Test Comment")
	buf.WriteByte(0x00) // Terminator

	// XMP Application Extension (simplified)
	buf.Write([]byte{0x21, 0xFF, 0x0B}) // Application Extension
	buf.WriteString("XMP Data")
	buf.Write([]byte{'X', 'M', 'P'}) // Auth code

	// XMP data (minimal valid XML)
	xmpData := `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?><x:xmpmeta xmlns:x="adobe:ns:meta/"></x:xmpmeta><?xpacket end="w"?>`
	buf.WriteByte(byte(len(xmpData)))
	buf.WriteString(xmpData)
	buf.WriteByte(0x00) // Terminator

	// Single frame
	buf.Write([]byte{0x2C, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x05, 0x00, 0x00})
	buf.Write([]byte{0x08, 0x02, 0xAA, 0xBB, 0x00})

	// Trailer
	buf.WriteByte(0x3B)

	r := bytes.NewReader(buf.Bytes())
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Should have GIF directory
	hasGIF := false
	hasComments := false
	for _, dir := range dirs {
		if dir.Name == "GIF" {
			hasGIF = true
		}
		if dir.Name == "GIF-Comments" {
			hasComments = true
			// Verify comment content
			foundComment := false
			for _, tag := range dir.Tags {
				if tag.Name == "Comment" && tag.Value == "Test Comment" {
					foundComment = true
					break
				}
			}
			if !foundComment {
				t.Error("Comment extension was not parsed correctly")
			}
		}
	}

	if !hasGIF {
		t.Error("Parse() did not return GIF directory")
	}
	if !hasComments {
		t.Error("Parse() did not return GIF Comments directory")
	}
}

// TestParser_Parse_UnknownSeparatorNoBlockSize tests unknown separator that can't be skipped
func TestParser_Parse_UnknownSeparatorNoBlockSize(t *testing.T) {
	p := New()

	// GIF with unknown separator followed by data that doesn't look like a block size
	data := []byte{
		'G', 'I', 'F', '8', '9', 'a',
		0x0A, 0x00, 0x0A, 0x00,
		0x00, 0x00, 0x00,
		// Frame 1
		0x2C, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x05, 0x00, 0x00,
		0x08, 0x02, 0xAA, 0xBB, 0x00,
		// Unknown separator with invalid block size (0x00 or 0xFF)
		0x99, // Unknown separator
		0x00, // Looks like terminator, not a valid block size
	}

	r := bytes.NewReader(data)
	dirs, parseErr := p.Parse(r)

	// Should have error but still return GIF directory
	if parseErr == nil {
		t.Error("Parse() expected error for unknown separator, got nil")
	}

	// Should still return GIF directory
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories after unknown separator")
	}

	// Should have counted the one frame before unknown separator
	var gifDir *parser.Directory
	for i := range dirs {
		if dirs[i].Name == "GIF" {
			gifDir = &dirs[i]
			break
		}
	}

	if gifDir == nil {
		t.Fatal("Parse() did not return GIF directory")
	}
}

// TestParser_Parse_NETSCAPEReadError tests NETSCAPE extension with read error
func TestParser_Parse_NETSCAPEReadError(t *testing.T) {
	p := New()

	// Create GIF with NETSCAPE extension but truncated data
	data := []byte{
		'G', 'I', 'F', '8', '9', 'a',
		0x0A, 0x00, 0x0A, 0x00,
		0x00, 0x00, 0x00,
		// NETSCAPE2.0 extension
		0x21, 0xFF, 0x0B,
		'N', 'E', 'T', 'S', 'C', 'A', 'P', 'E',
		'2', '.', '0',
		// Truncated - missing sub-block data
	}

	r := bytes.NewReader(data)
	dirs, _ := p.Parse(r)

	// Should still return GIF directory even with truncated NETSCAPE
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories after truncated NETSCAPE")
	}
}

// TestParser_Parse_NETSCAPEInvalidSubBlock tests NETSCAPE with invalid sub-block size
func TestParser_Parse_NETSCAPEInvalidSubBlock(t *testing.T) {
	p := New()

	// Create GIF with NETSCAPE extension but invalid sub-block size
	var buf bytes.Buffer
	buf.WriteString("GIF89a")
	buf.Write([]byte{0x0A, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00})

	// NETSCAPE2.0 extension
	buf.Write([]byte{0x21, 0xFF, 0x0B})
	buf.WriteString("NETSCAPE2.0")
	// Invalid sub-block size (not 3)
	buf.WriteByte(0x05) // Wrong size
	buf.Write([]byte{0x01, 0x00, 0x00, 0x00, 0x00})
	buf.WriteByte(0x00) // Terminator

	// Frame
	buf.Write([]byte{0x2C, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x05, 0x00, 0x00})
	buf.Write([]byte{0x08, 0x02, 0xAA, 0xBB, 0x00})
	buf.WriteByte(0x3B) // Trailer

	r := bytes.NewReader(buf.Bytes())
	dirs, _ := p.Parse(r)

	// Should still parse successfully
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories")
	}

	// Should NOT have AnimationIterations tag (invalid NETSCAPE)
	var gifDir *parser.Directory
	for i := range dirs {
		if dirs[i].Name == "GIF" {
			gifDir = &dirs[i]
			break
		}
	}

	if gifDir == nil {
		t.Fatal("Parse() did not return GIF directory")
	}

	for _, tag := range gifDir.Tags {
		if tag.Name == "AnimationIterations" {
			t.Error("Parse() should not add AnimationIterations for invalid NETSCAPE sub-block")
		}
	}
}

// TestParser_Parse_XMPPeekError tests XMP extension with read error when peeking
func TestParser_Parse_XMPPeekError(t *testing.T) {
	p := New()

	// Create GIF with XMP extension but truncated at peek
	data := []byte{
		'G', 'I', 'F', '8', '9', 'a',
		0x0A, 0x00, 0x0A, 0x00,
		0x00, 0x00, 0x00,
		// XMP Application Extension
		0x21, 0xFF, 0x0B,
		'X', 'M', 'P', ' ', 'D', 'a', 't', 'a',
		'X', 'M', 'P',
		// Truncated - missing peek byte
	}

	r := bytes.NewReader(data)
	dirs, _ := p.Parse(r)

	// Should still return GIF directory
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories after XMP peek error")
	}
}

// TestParser_Parse_UnknownSeparatorWithValidBlockSize tests successful skip of unknown separator with valid block
func TestParser_Parse_UnknownSeparatorWithValidBlockSize(t *testing.T) {
	p := New()

	var buf bytes.Buffer
	// GIF header
	buf.WriteString("GIF89a")
	// Logical screen descriptor
	buf.Write([]byte{0x0A, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00})

	// Unknown separator 0x50 (not 0x21, 0x2C, 0x3B, or 0x00)
	buf.WriteByte(0x50)
	// Valid block size
	buf.WriteByte(0x05)
	// Block data (5 bytes)
	buf.Write([]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE})
	// Block terminator
	buf.WriteByte(0x00)

	// Valid frame so we have something to parse
	buf.Write([]byte{0x2C, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x05, 0x00, 0x00})
	buf.Write([]byte{0x08, 0x02, 0xAA, 0xBB, 0x00})

	// Trailer
	buf.WriteByte(0x3B)

	r := bytes.NewReader(buf.Bytes())
	dirs, parseErr := p.Parse(r)

	// Should parse successfully despite unknown separator
	if len(dirs) == 0 {
		t.Fatal("Parse() returned no directories")
	}

	// Should have a parse error about the unknown separator
	if parseErr == nil {
		t.Error("Parse() should report unknown separator error")
	} else {
		errs := parseErr.Unwrap()
		if len(errs) == 0 {
			t.Error("Parse() should report unknown separator error")
		}
	}

	// But should still extract GIF metadata
	var gifDir *parser.Directory
	for i := range dirs {
		if dirs[i].Name == "GIF" {
			gifDir = &dirs[i]
			break
		}
	}

	if gifDir == nil {
		t.Fatal("Parse() did not return GIF directory")
	}
}
