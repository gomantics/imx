package canon

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/parser/tiff/makernote"
)

func TestHandler_Manufacturer(t *testing.T) {
	h := New()
	if got := h.Manufacturer(); got != "Canon" {
		t.Errorf("Manufacturer() = %v, want Canon", got)
	}
}

func TestHandler_Detect(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantOK     bool
		wantOffset int64
	}{
		{
			name: "Valid Canon LE - 2 entries",
			data: func() []byte {
				// Build a minimal valid IFD with 2 entries
				buf := make([]byte, 2+2*12+4) // count + 2 entries + next IFD
				binary.LittleEndian.PutUint16(buf[0:2], 2)
				// Entry 1: tag=0x0001, type=3 (SHORT), count=1, value=0
				binary.LittleEndian.PutUint16(buf[2:4], 0x0001)
				binary.LittleEndian.PutUint16(buf[4:6], 3)
				binary.LittleEndian.PutUint32(buf[6:10], 1)
				binary.LittleEndian.PutUint32(buf[10:14], 0)
				// Entry 2: tag=0x0006, type=2 (ASCII), count=1, value=0
				binary.LittleEndian.PutUint16(buf[14:16], 0x0006)
				binary.LittleEndian.PutUint16(buf[16:18], 2)
				binary.LittleEndian.PutUint32(buf[18:22], 1)
				binary.LittleEndian.PutUint32(buf[22:26], 0)
				return buf
			}(),
			wantOK:     true,
			wantOffset: 0,
		},
		{
			name: "Valid Canon BE - 2 entries",
			data: func() []byte {
				buf := make([]byte, 2+2*12+4)
				binary.BigEndian.PutUint16(buf[0:2], 2)
				binary.BigEndian.PutUint16(buf[2:4], 0x0001)
				binary.BigEndian.PutUint16(buf[4:6], 3)
				binary.BigEndian.PutUint32(buf[6:10], 1)
				binary.BigEndian.PutUint32(buf[10:14], 0)
				binary.BigEndian.PutUint16(buf[14:16], 0x0006)
				binary.BigEndian.PutUint16(buf[16:18], 2)
				binary.BigEndian.PutUint32(buf[18:22], 1)
				binary.BigEndian.PutUint32(buf[22:26], 0)
				return buf
			}(),
			wantOK:     true,
			wantOffset: 0,
		},
		{
			name:   "Too short",
			data:   []byte{0x01, 0x00},
			wantOK: false,
		},
		{
			name: "Zero entries",
			data: func() []byte {
				buf := make([]byte, 14)
				binary.LittleEndian.PutUint16(buf[0:2], 0)
				return buf
			}(),
			wantOK: false,
		},
		{
			name: "Too many entries",
			data: func() []byte {
				buf := make([]byte, 14)
				binary.LittleEndian.PutUint16(buf[0:2], 101)
				return buf
			}(),
			wantOK: false,
		},
		{
			name: "Invalid tag type",
			data: func() []byte {
				buf := make([]byte, 2+12+4)
				binary.LittleEndian.PutUint16(buf[0:2], 1)
				binary.LittleEndian.PutUint16(buf[2:4], 0x0001)
				binary.LittleEndian.PutUint16(buf[4:6], 99) // Invalid type
				binary.LittleEndian.PutUint32(buf[6:10], 1)
				return buf
			}(),
			wantOK: false,
		},
		{
			name:   "Nikon header - should not match",
			data:   []byte("Nikon\x00\x02\x00\x00\x00II"),
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New()
			ok, cfg := h.Detect(tt.data)
			if ok != tt.wantOK {
				t.Errorf("Detect() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok {
				if cfg.IFDOffset != tt.wantOffset {
					t.Errorf("IFDOffset = %d, want %d", cfg.IFDOffset, tt.wantOffset)
				}
				if cfg.OffsetBase != makernote.OffsetAbsolute {
					t.Errorf("OffsetBase = %v, want OffsetAbsolute", cfg.OffsetBase)
				}
				if cfg.HasNextIFD {
					t.Error("HasNextIFD = true, want false")
				}
			}
		})
	}
}

func TestHandler_Parse(t *testing.T) {
	// Build a test MakerNote with known tags
	data := buildTestMakerNote()

	h := New()
	cfg := &makernote.Config{
		IFDOffset:  0,
		OffsetBase: makernote.OffsetAbsolute,
		ByteOrder:  binary.LittleEndian,
		HasNextIFD: false,
	}

	tags, parseErr := h.Parse(bytes.NewReader(data), 0, 0, cfg)
	if parseErr != nil {
		t.Fatalf("Parse() error = %v", parseErr)
	}

	if tags == nil {
		t.Fatal("Parse() returned nil tags")
	}

	if len(tags) != 3 {
		t.Errorf("Got %d tags, want 3", len(tags))
	}

	// Check specific tags
	tagMap := make(map[string]any)
	for _, tag := range tags {
		tagMap[tag.Name] = tag.Value
	}

	if _, ok := tagMap["ImageType"]; !ok {
		t.Error("Missing ImageType tag")
	}

	if _, ok := tagMap["SerialNumber"]; !ok {
		t.Error("Missing SerialNumber tag")
	}
}

func TestHandler_TagName(t *testing.T) {
	h := New()

	tests := []struct {
		tagID uint16
		want  string
	}{
		{0x0001, "CameraSettings"},
		{0x0006, "ImageType"},
		{0x000C, "SerialNumber"},
		{0x0095, "LensModel"},
		{0xFFFF, "0xFFFF"}, // Unknown tag
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := h.TagName(tt.tagID); got != tt.want {
				t.Errorf("TagName(0x%04X) = %s, want %s", tt.tagID, got, tt.want)
			}
		})
	}
}

func TestGetTypeSize(t *testing.T) {
	tests := []struct {
		tagType uint16
		want    int
	}{
		{1, 1},  // BYTE
		{2, 1},  // ASCII
		{3, 2},  // SHORT
		{4, 4},  // LONG
		{5, 8},  // RATIONAL
		{6, 1},  // SBYTE
		{7, 1},  // UNDEFINED
		{8, 2},  // SSHORT
		{9, 4},  // SLONG
		{10, 8}, // SRATIONAL
		{11, 4}, // FLOAT
		{12, 8}, // DOUBLE
		{99, 0}, // Unknown
	}

	for _, tt := range tests {
		if got := getTypeSize(tt.tagType); got != tt.want {
			t.Errorf("getTypeSize(%d) = %d, want %d", tt.tagType, got, tt.want)
		}
	}
}

// buildTestMakerNote creates a test Canon MakerNote with 3 entries.
func buildTestMakerNote() []byte {
	// Layout:
	// 0-1: entry count (3)
	// 2-13: entry 1 (ImageType - ASCII string)
	// 14-25: entry 2 (SerialNumber - LONG)
	// 26-37: entry 3 (LensModel - ASCII string)
	// 38-41: next IFD offset (0)
	// 42+: string data

	buf := make([]byte, 100)

	// Entry count: 3
	binary.LittleEndian.PutUint16(buf[0:2], 3)

	// Entry 1: ImageType (0x0006), ASCII, 10 bytes at offset 42
	binary.LittleEndian.PutUint16(buf[2:4], 0x0006)
	binary.LittleEndian.PutUint16(buf[4:6], 2) // ASCII
	binary.LittleEndian.PutUint32(buf[6:10], 10)
	binary.LittleEndian.PutUint32(buf[10:14], 42)

	// Entry 2: SerialNumber (0x000C), LONG, inline
	binary.LittleEndian.PutUint16(buf[14:16], 0x000C)
	binary.LittleEndian.PutUint16(buf[16:18], 4) // LONG
	binary.LittleEndian.PutUint32(buf[18:22], 1)
	binary.LittleEndian.PutUint32(buf[22:26], 12345678)

	// Entry 3: LensModel (0x0095), ASCII, 15 bytes at offset 52
	binary.LittleEndian.PutUint16(buf[26:28], 0x0095)
	binary.LittleEndian.PutUint16(buf[28:30], 2) // ASCII
	binary.LittleEndian.PutUint32(buf[30:34], 15)
	binary.LittleEndian.PutUint32(buf[34:38], 52)

	// Next IFD offset: 0
	binary.LittleEndian.PutUint32(buf[38:42], 0)

	// String data
	copy(buf[42:52], "Canon EOS\x00")
	copy(buf[52:67], "EF 50mm f/1.4\x00")

	return buf
}

func TestHandler_ParseInlineData(t *testing.T) {
	// Test that inline data (<=4 bytes) is correctly read
	buf := make([]byte, 50)

	// Entry count: 2
	binary.LittleEndian.PutUint16(buf[0:2], 2)

	// Entry 1: SHORT value inline
	binary.LittleEndian.PutUint16(buf[2:4], 0x0001)
	binary.LittleEndian.PutUint16(buf[4:6], 3) // SHORT
	binary.LittleEndian.PutUint32(buf[6:10], 1)
	binary.LittleEndian.PutUint16(buf[10:12], 42) // Value inline

	// Entry 2: LONG value inline
	binary.LittleEndian.PutUint16(buf[14:16], 0x0002)
	binary.LittleEndian.PutUint16(buf[16:18], 4) // LONG
	binary.LittleEndian.PutUint32(buf[18:22], 1)
	binary.LittleEndian.PutUint32(buf[22:26], 12345)

	h := New()
	cfg := &makernote.Config{
		IFDOffset:  0,
		OffsetBase: makernote.OffsetAbsolute,
		ByteOrder:  binary.LittleEndian,
	}

	tags, parseErr := h.Parse(bytes.NewReader(buf), 0, 0, cfg)
	if parseErr != nil {
		t.Fatalf("Parse() error = %v", parseErr)
	}

	if len(tags) != 2 {
		t.Fatalf("Got %d tags, want 2", len(tags))
	}

	// Check SHORT value
	if tags[0].Value != uint16(42) {
		t.Errorf("SHORT value = %v, want 42", tags[0].Value)
	}

	// Check LONG value
	if tags[1].Value != uint32(12345) {
		t.Errorf("LONG value = %v, want 12345", tags[1].Value)
	}
}
