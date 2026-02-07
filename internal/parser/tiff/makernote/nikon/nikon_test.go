package nikon

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/parser/tiff/makernote"
)

func TestHandler_Manufacturer(t *testing.T) {
	h := New()
	if got := h.Manufacturer(); got != "Nikon" {
		t.Errorf("Manufacturer() = %v, want Nikon", got)
	}
}

func TestHandler_Detect_Type3(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantOK     bool
		wantOffset int64
		wantOrder  binary.ByteOrder
	}{
		{
			name: "Valid Nikon Type 3 LE",
			data: func() []byte {
				// Nikon\x00\x02\x00\x00 + II + 0x002a + IFD offset
				buf := make([]byte, 26)
				copy(buf[0:5], "Nikon")
				buf[5] = 0x00
				buf[6] = 0x02
				buf[7] = 0x00
				buf[8] = 0x00
				buf[9] = 0x00
				buf[10] = 'I'
				buf[11] = 'I'
				binary.LittleEndian.PutUint16(buf[12:14], 0x002a) // TIFF magic
				binary.LittleEndian.PutUint32(buf[14:18], 8)      // IFD offset
				// Minimal IFD
				binary.LittleEndian.PutUint16(buf[18:20], 1) // entry count
				return buf
			}(),
			wantOK:     true,
			wantOffset: 18,
			wantOrder:  binary.LittleEndian,
		},
		{
			name: "Valid Nikon Type 3 BE",
			data: func() []byte {
				buf := make([]byte, 26)
				copy(buf[0:5], "Nikon")
				buf[5] = 0x00
				buf[6] = 0x02
				buf[7] = 0x00
				buf[8] = 0x00
				buf[9] = 0x00
				buf[10] = 'M'
				buf[11] = 'M'
				binary.BigEndian.PutUint16(buf[12:14], 0x002a)
				binary.BigEndian.PutUint32(buf[14:18], 8)
				binary.BigEndian.PutUint16(buf[18:20], 1)
				return buf
			}(),
			wantOK:     true,
			wantOffset: 18,
			wantOrder:  binary.BigEndian,
		},
		{
			name:   "Too short",
			data:   []byte("Nikon\x00\x02"),
			wantOK: false,
		},
		{
			name: "Wrong magic byte (Type 1 format)",
			data: func() []byte {
				buf := make([]byte, 18)
				copy(buf[0:5], "Nikon")
				buf[5] = 0x00
				buf[6] = 0x01 // Type 1, not Type 3
				return buf
			}(),
			wantOK: false,
		},
		{
			name: "Invalid byte order",
			data: func() []byte {
				buf := make([]byte, 18)
				copy(buf[0:5], "Nikon")
				buf[5] = 0x00
				buf[6] = 0x02
				buf[7] = 0x00
				buf[8] = 0x00
				buf[9] = 0x00
				buf[10] = 'X' // Invalid
				buf[11] = 'X'
				return buf
			}(),
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, cfg := makernote.DetectNikonType3(tt.data)
			if ok != tt.wantOK {
				t.Errorf("DetectNikonType3() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok {
				if cfg.IFDOffset != tt.wantOffset {
					t.Errorf("IFDOffset = %d, want %d", cfg.IFDOffset, tt.wantOffset)
				}
				if cfg.ByteOrder != tt.wantOrder {
					t.Errorf("ByteOrder = %v, want %v", cfg.ByteOrder, tt.wantOrder)
				}
				if cfg.OffsetBase != makernote.OffsetRelativeToMakerNote {
					t.Errorf("OffsetBase = %v, want OffsetRelativeToMakerNote", cfg.OffsetBase)
				}
				if cfg.Variant != "Type3" {
					t.Errorf("Variant = %s, want Type3", cfg.Variant)
				}
			}
		})
	}
}

func TestHandler_Detect_Type1(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantOK     bool
		wantOffset int64
	}{
		{
			name: "Valid Nikon Type 1",
			data: func() []byte {
				buf := make([]byte, 20)
				copy(buf[0:5], "Nikon")
				buf[5] = 0x00
				buf[6] = 0x01
				buf[7] = 0x00
				// IFD starts at offset 8
				binary.LittleEndian.PutUint16(buf[8:10], 1) // entry count
				return buf
			}(),
			wantOK:     true,
			wantOffset: 8,
		},
		{
			name:   "Too short",
			data:   []byte("Nikon\x00\x01"),
			wantOK: false,
		},
		{
			name: "Not Nikon",
			data: func() []byte {
				buf := make([]byte, 12)
				copy(buf[0:5], "Canon")
				return buf
			}(),
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, cfg := makernote.DetectNikonType1(tt.data)
			if ok != tt.wantOK {
				t.Errorf("DetectNikonType1() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok {
				if cfg.IFDOffset != tt.wantOffset {
					t.Errorf("IFDOffset = %d, want %d", cfg.IFDOffset, tt.wantOffset)
				}
				if cfg.OffsetBase != makernote.OffsetAbsolute {
					t.Errorf("OffsetBase = %v, want OffsetAbsolute", cfg.OffsetBase)
				}
				if cfg.ByteOrder != nil {
					t.Errorf("ByteOrder = %v, want nil (inherit)", cfg.ByteOrder)
				}
				if cfg.Variant != "Type1" {
					t.Errorf("Variant = %s, want Type1", cfg.Variant)
				}
			}
		})
	}
}

func TestHandler_Detect_Combined(t *testing.T) {
	h := New()

	// Type 3 should be detected
	type3Data := make([]byte, 26)
	copy(type3Data[0:5], "Nikon")
	type3Data[5] = 0x00
	type3Data[6] = 0x02
	type3Data[10] = 'I'
	type3Data[11] = 'I'
	binary.LittleEndian.PutUint16(type3Data[12:14], 0x002a)
	binary.LittleEndian.PutUint32(type3Data[14:18], 8)

	ok, cfg := h.Detect(type3Data)
	if !ok {
		t.Error("Handler.Detect() should detect Type 3")
	}
	if cfg.Variant != "Type3" {
		t.Errorf("Variant = %s, want Type3", cfg.Variant)
	}

	// Type 1 should be detected
	type1Data := make([]byte, 20)
	copy(type1Data[0:5], "Nikon")
	type1Data[5] = 0x00
	type1Data[6] = 0x01
	type1Data[7] = 0x00

	ok, cfg = h.Detect(type1Data)
	if !ok {
		t.Error("Handler.Detect() should detect Type 1")
	}
	if cfg.Variant != "Type1" {
		t.Errorf("Variant = %s, want Type1", cfg.Variant)
	}

	// Sony should not be detected
	sonyData := []byte("SONY DSC \x00\x00\x00")
	ok, _ = h.Detect(sonyData)
	if ok {
		t.Error("Handler.Detect() should not detect Sony data")
	}
}

func TestHandler_Parse_Type1(t *testing.T) {
	// Build a test Type 1 MakerNote
	data := buildTestType1MakerNote()

	h := New()
	cfg := &makernote.Config{
		IFDOffset:  8,
		OffsetBase: makernote.OffsetAbsolute,
		ByteOrder:  binary.LittleEndian,
		HasNextIFD: false,
		Variant:    "Type1",
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
	tagMap := make(map[string]interface{})
	for _, tag := range tags {
		tagMap[tag.Name] = tag.Value
	}

	if _, ok := tagMap["MakerNoteVersion"]; !ok {
		t.Error("Missing MakerNoteVersion tag")
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
		{0x0001, "MakerNoteVersion"},
		{0x0002, "ISO"},
		{0x001D, "SerialNumber"},
		{0x0083, "LensType"},
		{0x0084, "Lens"},
		{0xFFFF, ""}, // Unknown tag returns empty string
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := h.TagName(tt.tagID)
			if got != tt.want {
				t.Errorf("TagName(0x%04X) = %s, want %s", tt.tagID, got, tt.want)
			}
		})
	}
}

func TestHandler_EncryptedTags(t *testing.T) {
	// Verify encrypted tags are skipped
	if !isEncryptedTag(0x0098) {
		t.Error("0x0098 (LensData) should be encrypted")
	}
	if !isEncryptedTag(0x00A8) {
		t.Error("0x00A8 (FlashInfo) should be encrypted")
	}
	if isEncryptedTag(0x0001) {
		t.Error("0x0001 should not be encrypted")
	}
}

func TestGetTypeSize(t *testing.T) {
	tests := []struct {
		typeVal uint16
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
		if got := getTypeSize(tt.typeVal); got != tt.want {
			t.Errorf("getTypeSize(%d) = %d, want %d", tt.typeVal, got, tt.want)
		}
	}
}

// buildTestType1MakerNote creates a test Nikon Type 1 MakerNote with 3 entries.
func buildTestType1MakerNote() []byte {
	// Layout:
	// 0-7: Nikon\x00\x01\x00 header
	// 8-9: entry count (3)
	// 10-21: entry 1 (MakerNoteVersion - UNDEFINED)
	// 22-33: entry 2 (SerialNumber - ASCII)
	// 34-45: entry 3 (Quality - SHORT)
	// 46-49: next IFD offset (0)
	// 50+: string data

	buf := make([]byte, 100)

	// Nikon Type 1 header
	copy(buf[0:5], "Nikon")
	buf[5] = 0x00
	buf[6] = 0x01
	buf[7] = 0x00

	// Entry count: 3
	binary.LittleEndian.PutUint16(buf[8:10], 3)

	// Entry 1: MakerNoteVersion (0x0001), UNDEFINED, 4 bytes inline
	binary.LittleEndian.PutUint16(buf[10:12], 0x0001)
	binary.LittleEndian.PutUint16(buf[12:14], 7) // UNDEFINED
	binary.LittleEndian.PutUint32(buf[14:18], 4)
	copy(buf[18:22], "0210") // Version inline

	// Entry 2: SerialNumber (0x001D), ASCII, 10 bytes at offset 50
	binary.LittleEndian.PutUint16(buf[22:24], 0x001D)
	binary.LittleEndian.PutUint16(buf[24:26], 2) // ASCII
	binary.LittleEndian.PutUint32(buf[26:30], 10)
	binary.LittleEndian.PutUint32(buf[30:34], 50)

	// Entry 3: Quality (0x0004), SHORT, inline
	binary.LittleEndian.PutUint16(buf[34:36], 0x0004)
	binary.LittleEndian.PutUint16(buf[36:38], 3) // SHORT
	binary.LittleEndian.PutUint32(buf[38:42], 1)
	binary.LittleEndian.PutUint16(buf[42:44], 1) // Value: 1 (Fine)

	// Next IFD offset: 0
	binary.LittleEndian.PutUint32(buf[46:50], 0)

	// String data
	copy(buf[50:60], "12345678\x00\x00")

	return buf
}
