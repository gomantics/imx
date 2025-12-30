package gif

import (
	"bytes"
	"io"
	"testing"

	"github.com/gomantics/imx/internal/parser/xmp"
)

func TestRemoveMagicTrailer(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantLen  int
		wantData string
	}{
		{
			name:     "short data - no trailer possible",
			input:    []byte("short XMP data"),
			wantLen:  14,
			wantData: "short XMP data",
		},
		{
			name:     "exactly 257 bytes - no trailer",
			input:    make([]byte, 257),
			wantLen:  257,
			wantData: "",
		},
		{
			name: "data with valid magic trailer",
			input: func() []byte {
				data := []byte("XMP DATA HERE")
				trailer := make([]byte, 257)
				trailer[0] = 0x01 // Magic byte
				// Rest are zeros by default
				return append(data, trailer...)
			}(),
			wantLen:  13,
			wantData: "XMP DATA HERE",
		},
		{
			name: "data with 0x01 but not all zeros",
			input: func() []byte {
				data := []byte("XMP DATA")
				trailer := make([]byte, 257)
				trailer[0] = 0x01
				trailer[100] = 0xFF // Not all zeros
				return append(data, trailer...)
			}(),
			wantLen: 265, // 8 + 257, no trimming
		},
		{
			name: "data without 0x01 at trailer position",
			input: func() []byte {
				data := []byte("XMP DATA")
				trailer := make([]byte, 257)
				trailer[0] = 0x00 // Not 0x01
				return append(data, trailer...)
			}(),
			wantLen: 265, // No trimming
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeMagicTrailer(tt.input)
			if len(result) != tt.wantLen {
				t.Errorf("removeMagicTrailer() length = %d, want %d", len(result), tt.wantLen)
			}
			if tt.wantData != "" && string(result) != tt.wantData {
				t.Errorf("removeMagicTrailer() data = %q, want %q", string(result), tt.wantData)
			}
		})
	}
}

func TestParseCommentExtension(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		wantComment string
		wantErr     bool
	}{
		{
			name: "simple comment",
			data: []byte{
				0x0C,                                                       // Block size
				'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!', // Comment data
				0x00, // Block terminator
			},
			wantComment: "Hello World!",
		},
		{
			name: "multi-block comment",
			data: []byte{
				0x05, 'H', 'e', 'l', 'l', 'o', // First block
				0x06, ' ', 'W', 'o', 'r', 'l', 'd', // Second block
				0x00, // Terminator
			},
			wantComment: "Hello World",
		},
		{
			name: "empty comment",
			data: []byte{
				0x00, // Immediate terminator
			},
			wantComment: "",
		},
		{
			name: "read error",
			data: []byte{
				0x05, // Says 5 bytes but only 3 follow
				'A', 'B', 'C',
			},
			wantComment: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			var buf [11]byte
			tag, _ := parseCommentExtension(r, 0, &buf)

			if tt.wantComment == "" && tag != nil {
				t.Errorf("parseCommentExtension() expected nil tag for empty comment")
				return
			}

			if tt.wantComment != "" {
				if tag == nil {
					t.Fatalf("parseCommentExtension() returned nil tag, want comment")
				}
				if tag.Value != tt.wantComment {
					t.Errorf("parseCommentExtension() comment = %q, want %q", tag.Value, tt.wantComment)
				}
				if tag.Name != "Comment" {
					t.Errorf("parseCommentExtension() tag name = %q, want %q", tag.Name, "Comment")
				}
			}
		})
	}
}

func TestParseApplicationExtension(t *testing.T) {
	xmpParser := xmp.New()

	tests := []struct {
		name     string
		data     []byte
		wantDirs int
		wantXMP  bool
	}{
		{
			name: "read error on block size",
			data: []byte{
				// Empty - will cause read error on first byte
			},
			wantDirs: 0,
			wantXMP:  false,
		},
		{
			name: "invalid block size",
			data: []byte{
				0x0A, // Wrong size (should be 11)
				'X', 'M', 'P', ' ', 'D', 'a', 't', 'a', 'X', 'M',
				0x00,
			},
			wantDirs: 0,
			wantXMP:  false,
		},
		{
			name: "read error on app ID",
			data: []byte{
				0x0B,          // Correct block size
				'X', 'M', 'P', // Only 3 bytes instead of 11
			},
			wantDirs: 0,
			wantXMP:  false,
		},
		{
			name: "not XMP application",
			data: []byte{
				0x0B,                                   // Block size = 11
				'N', 'E', 'T', 'S', 'C', 'A', 'P', 'E', // App ID
				'2', '.', '0', // Auth code
				0x03, 0x01, 0x00, 0x00, // NETSCAPE extension data
				0x00, // Terminator
			},
			wantDirs: 0,
			wantXMP:  false,
		},
		{
			name: "XMP with standard format (sub-blocks)",
			data: []byte{
				0x0B,                                   // Block size = 11
				'X', 'M', 'P', ' ', 'D', 'a', 't', 'a', // App ID
				'X', 'M', 'P', // Auth code
				0x05, '<', '?', 'x', 'm', 'l', // Sub-block with minimal XML
				0x00, // Terminator
			},
			wantDirs: 0, // Won't parse successfully but tests the path
			wantXMP:  false,
		},
		{
			name: "XMP read error",
			data: []byte{
				0x0B,                                   // Block size = 11
				'X', 'M', 'P', ' ', 'D', 'a', 't', 'a', // App ID
				'X', 'M', 'P', // Auth code
				// Missing data - read error
			},
			wantDirs: 0,
			wantXMP:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			var buf [11]byte
			dirs, _ := parseApplicationExtension(r, 0, &buf, xmpParser)

			if len(dirs) != tt.wantDirs {
				t.Errorf("parseApplicationExtension() dirs count = %d, want %d", len(dirs), tt.wantDirs)
			}
		})
	}
}

func TestReadOldFormatXMP(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantData string
		wantPos  int64
	}{
		{
			name:     "no end marker found",
			data:     []byte("<?xpacket begin but no end"),
			wantData: "",
			wantPos:  0,
		},
		{
			name:     "end marker but no closing tag",
			data:     []byte("<?xpacket begin='x'?><test/><?xpacket end="),
			wantData: "",
			wantPos:  0,
		},
		{
			name:     "complete XMP with terminator",
			data:     []byte("<?xpacket begin='x'?><x:xmpmeta xmlns:x='test'></x:xmpmeta><?xpacket end='w'?>\x00"),
			wantData: "<?xpacket begin='x'?><x:xmpmeta xmlns:x='test'></x:xmpmeta><?xpacket end='w'?>",
			wantPos:  79, // Length of XMP + terminator position
		},
		{
			name: "XMP with magic trailer",
			data: func() []byte {
				xmp := []byte("<?xpacket begin='x'?><test/><?xpacket end='w'?>")
				trailer := make([]byte, 257)
				trailer[0] = 0x01 // Magic byte
				// Rest are zeros
				trailer[256] = 0x00 // Block terminator at end
				return append(xmp, trailer...)
			}(),
			wantData: "<?xpacket begin='x'?><test/><?xpacket end='w'?>",
			wantPos:  49, // Just XMP length since we consume the trailing terminator differently
		},
		{
			name:     "EOF before terminator",
			data:     []byte("<?xpacket begin='x'?><test/><?xpacket end='w'?>"),
			wantData: "",
			wantPos:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			var buf [11]byte
			xmpData, pos := readOldFormatXMP(r, 0, &buf)

			if tt.wantData == "" {
				if len(xmpData) != 0 {
					t.Errorf("readOldFormatXMP() expected empty data, got %d bytes", len(xmpData))
				}
			} else {
				if string(xmpData) != tt.wantData {
					t.Errorf("readOldFormatXMP() data = %q, want %q", string(xmpData), tt.wantData)
				}
			}

			if pos != tt.wantPos {
				t.Errorf("readOldFormatXMP() pos = %d, want %d", pos, tt.wantPos)
			}
		})
	}
}

func TestReadDataSubBlocks(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantData string
		wantErr  bool
	}{
		{
			name: "single block",
			data: []byte{
				0x05, 'H', 'e', 'l', 'l', 'o',
				0x00,
			},
			wantData: "Hello",
		},
		{
			name: "multiple blocks",
			data: []byte{
				0x05, 'H', 'e', 'l', 'l', 'o',
				0x01, ' ',
				0x05, 'W', 'o', 'r', 'l', 'd',
				0x00,
			},
			wantData: "Hello World",
		},
		{
			name: "empty blocks",
			data: []byte{
				0x00,
			},
			wantData: "",
		},
		{
			name: "read error on block size",
			data: []byte{
				// Empty - will cause read error
			},
			wantData: "",
		},
		{
			name: "read error mid-block",
			data: []byte{
				0x0A,          // Says 10 bytes
				'A', 'B', 'C', // Only 3 bytes
			},
			wantData: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			var buf [11]byte
			result, _ := readDataSubBlocks(r, 0, &buf)

			if string(result) != tt.wantData {
				t.Errorf("readDataSubBlocks() = %q, want %q", string(result), tt.wantData)
			}
		})
	}
}

func TestSkipDataSubBlocks(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantPos int64
	}{
		{
			name: "single block",
			data: []byte{
				0x05, 'H', 'e', 'l', 'l', 'o',
				0x00,
			},
			wantPos: 7,
		},
		{
			name: "multiple blocks",
			data: []byte{
				0x03, 'A', 'B', 'C',
				0x02, 'D', 'E',
				0x00,
			},
			wantPos: 8,
		},
		{
			name: "immediate terminator",
			data: []byte{
				0x00,
			},
			wantPos: 1,
		},
		{
			name: "read error",
			data: []byte{
				0x0A, // Says 10 bytes but fewer follow
				'A', 'B',
			},
			wantPos: 11, // Advances past block size even if read fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			var buf [11]byte
			pos := skipDataSubBlocks(r, 0, &buf)

			if pos != tt.wantPos {
				t.Errorf("skipDataSubBlocks() pos = %d, want %d", pos, tt.wantPos)
			}
		})
	}
}

func TestParseExtension(t *testing.T) {
	xmpParser := xmp.New()

	tests := []struct {
		name        string
		data        []byte
		wantDirLen  int
		wantTagLen  int
		description string
	}{
		{
			name: "comment extension",
			data: []byte{
				0xFE, // Comment extension label
				0x05, 'H', 'e', 'l', 'l', 'o',
				0x00,
			},
			wantDirLen:  0,
			wantTagLen:  1,
			description: "Should parse comment",
		},
		{
			name: "graphic control extension",
			data: []byte{
				0xF9,                         // Graphic Control Extension
				0x04, 0x00, 0x00, 0x00, 0x00, // GCE data
				0x00, // Terminator
			},
			wantDirLen:  0,
			wantTagLen:  0,
			description: "Should skip graphic control",
		},
		{
			name: "plain text extension",
			data: []byte{
				0x01, // Plain Text Extension
				0x05, 'A', 'B', 'C', 'D', 'E',
				0x00,
			},
			wantDirLen:  0,
			wantTagLen:  0,
			description: "Should skip plain text",
		},
		{
			name: "unknown extension",
			data: []byte{
				0x99, // Unknown label
				0x03, 'X', 'Y', 'Z',
				0x00,
			},
			wantDirLen:  0,
			wantTagLen:  0,
			description: "Should skip unknown",
		},
		{
			name:        "read error",
			data:        []byte{}, // Empty, will cause read error
			wantDirLen:  0,
			wantTagLen:  0,
			description: "Should handle read error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			var buf [11]byte
			dirs, tags, _ := parseExtension(r, 0, &buf, xmpParser)

			if len(dirs) != tt.wantDirLen {
				t.Errorf("%s: dirs length = %d, want %d", tt.description, len(dirs), tt.wantDirLen)
			}
			if len(tags) != tt.wantTagLen {
				t.Errorf("%s: tags length = %d, want %d", tt.description, len(tags), tt.wantTagLen)
			}
		})
	}
}

// errorReaderAt wraps bytes and returns custom error at specific offset
type errorReaderAt struct {
	data        []byte
	errorOffset int64
	customError error
}

func (e *errorReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= e.errorOffset {
		return 0, e.customError
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

// TestReadOldFormatXMP_NonEOFError tests handling of non-EOF errors during chunk reading
func TestReadOldFormatXMP_NonEOFError(t *testing.T) {
	// Create some XMP data
	xmpData := []byte("<?xpacket begin='x'?><test/>")

	// Create reader that returns custom error immediately
	customErr := io.ErrUnexpectedEOF
	r := &errorReaderAt{
		data:        xmpData,
		errorOffset: 0, // Error on first read
		customError: customErr,
	}

	var buf [11]byte
	result, pos := readOldFormatXMP(r, 0, &buf)

	// Should return empty on non-EOF error
	if len(result) != 0 {
		t.Errorf("readOldFormatXMP() expected empty result on non-EOF error, got %d bytes", len(result))
	}
	if pos != 0 {
		t.Errorf("readOldFormatXMP() expected pos=0 on error, got %d", pos)
	}
}
