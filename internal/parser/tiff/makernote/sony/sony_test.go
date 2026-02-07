package sony

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/parser/tiff/makernote"
)

func TestHandler_Manufacturer(t *testing.T) {
	h := New()
	if got := h.Manufacturer(); got != "Sony" {
		t.Errorf("Manufacturer() = %q, want %q", got, "Sony")
	}
}

func TestHandler_Detect(t *testing.T) {
	h := New()

	tests := []struct {
		name    string
		data    []byte
		wantOK  bool
		wantCfg *makernote.Config
	}{
		{
			name:   "SONY DSC header",
			data:   append([]byte("SONY DSC "), make([]byte, 100)...),
			wantOK: true,
			wantCfg: &makernote.Config{
				IFDOffset:  12,
				OffsetBase: makernote.OffsetAbsolute,
				ByteOrder:  binary.LittleEndian,
				HasNextIFD: false,
				Variant:    "Standard",
			},
		},
		{
			name:   "SONY CAM header",
			data:   append([]byte("SONY CAM "), make([]byte, 100)...),
			wantOK: true,
			wantCfg: &makernote.Config{
				IFDOffset:  12,
				OffsetBase: makernote.OffsetAbsolute,
				ByteOrder:  binary.LittleEndian,
				HasNextIFD: false,
				Variant:    "Standard",
			},
		},
		{
			name:   "not Sony - Canon",
			data:   []byte{0x05, 0x00, 0x01, 0x00, 0x02, 0x00}, // Looks like Canon IFD
			wantOK: false,
		},
		{
			name:   "not Sony - Nikon",
			data:   []byte("Nikon\x00\x02\x00"),
			wantOK: false,
		},
		{
			name:   "too short",
			data:   []byte("SONY"),
			wantOK: false,
		},
		{
			name:   "empty data",
			data:   []byte{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOK, gotCfg := h.Detect(tt.data)
			if gotOK != tt.wantOK {
				t.Errorf("Detect() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if tt.wantOK && gotCfg != nil {
				if gotCfg.IFDOffset != tt.wantCfg.IFDOffset {
					t.Errorf("Config.IFDOffset = %d, want %d", gotCfg.IFDOffset, tt.wantCfg.IFDOffset)
				}
				if gotCfg.OffsetBase != tt.wantCfg.OffsetBase {
					t.Errorf("Config.OffsetBase = %v, want %v", gotCfg.OffsetBase, tt.wantCfg.OffsetBase)
				}
				if gotCfg.ByteOrder != tt.wantCfg.ByteOrder {
					t.Errorf("Config.ByteOrder = %v, want %v", gotCfg.ByteOrder, tt.wantCfg.ByteOrder)
				}
			}
		})
	}
}

func TestHandler_TagName(t *testing.T) {
	h := New()

	tests := []struct {
		tagID    uint16
		expected string
	}{
		{0x0102, "Quality"},
		{0x0115, "WhiteBalance"},
		{0x2031, "SerialNumber"},
		{0xb001, "SonyModelID"},
		{0xb027, "LensID"},
		{0xb020, "CreativeStyle"},
		{0x9999, ""}, // Unknown tag
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := h.TagName(tt.tagID)
			if got != tt.expected {
				t.Errorf("TagName(0x%04X) = %q, want %q", tt.tagID, got, tt.expected)
			}
		})
	}
}

func TestHandler_Parse(t *testing.T) {
	h := New()

	// Build a minimal Sony MakerNote with header and one IFD entry
	// Header: "SONY DSC " (12 bytes)
	// Entry count: 1 (2 bytes)
	// Entry: tag=0xb001, type=SHORT(3), count=1, value=123
	data := make([]byte, 100)
	copy(data[0:12], []byte("SONY DSC "))

	// IFD at offset 12
	binary.LittleEndian.PutUint16(data[12:14], 1) // 1 entry

	// Entry at offset 14
	binary.LittleEndian.PutUint16(data[14:16], 0xb001) // Tag: SonyModelID
	binary.LittleEndian.PutUint16(data[16:18], 3)      // Type: SHORT
	binary.LittleEndian.PutUint32(data[18:22], 1)      // Count: 1
	binary.LittleEndian.PutUint32(data[22:26], 123)    // Value: 123 (inline)

	reader := bytes.NewReader(data)
	cfg := &makernote.Config{
		IFDOffset:  12,
		OffsetBase: makernote.OffsetAbsolute,
		ByteOrder:  binary.LittleEndian,
	}

	tags, parseErr := h.Parse(reader, 0, 0, cfg)

	if parseErr != nil && parseErr.OrNil() != nil {
		t.Errorf("Parse() returned errors: %v", parseErr)
	}

	if len(tags) != 1 {
		t.Fatalf("Parse() returned %d tags, want 1", len(tags))
	}

	tag := tags[0]
	if tag.Name != "SonyModelID" {
		t.Errorf("Tag.Name = %q, want %q", tag.Name, "SonyModelID")
	}
	if tag.ID != "Sony:0xB001" {
		t.Errorf("Tag.ID = %q, want %q", tag.ID, "Sony:0xB001")
	}
	if val, ok := tag.Value.(uint16); !ok || val != 123 {
		t.Errorf("Tag.Value = %v (%T), want 123 (uint16)", tag.Value, tag.Value)
	}
}

func TestHandler_Parse_MultipleEntries(t *testing.T) {
	h := New()

	// Build Sony MakerNote with 3 entries
	data := make([]byte, 200)
	copy(data[0:12], []byte("SONY DSC "))

	// IFD at offset 12
	binary.LittleEndian.PutUint16(data[12:14], 3) // 3 entries

	// Entry 1: Quality (0x0102), SHORT, value=1
	binary.LittleEndian.PutUint16(data[14:16], 0x0102)
	binary.LittleEndian.PutUint16(data[16:18], 3)
	binary.LittleEndian.PutUint32(data[18:22], 1)
	binary.LittleEndian.PutUint32(data[22:26], 1)

	// Entry 2: WhiteBalance (0x0115), SHORT, value=0
	binary.LittleEndian.PutUint16(data[26:28], 0x0115)
	binary.LittleEndian.PutUint16(data[28:30], 3)
	binary.LittleEndian.PutUint32(data[30:34], 1)
	binary.LittleEndian.PutUint32(data[34:38], 0)

	// Entry 3: SonyModelID (0xb001), SHORT, value=456
	binary.LittleEndian.PutUint16(data[38:40], 0xb001)
	binary.LittleEndian.PutUint16(data[40:42], 3)
	binary.LittleEndian.PutUint32(data[42:46], 1)
	binary.LittleEndian.PutUint32(data[46:50], 456)

	reader := bytes.NewReader(data)
	cfg := &makernote.Config{
		IFDOffset:  12,
		OffsetBase: makernote.OffsetAbsolute,
		ByteOrder:  binary.LittleEndian,
	}

	tags, parseErr := h.Parse(reader, 0, 0, cfg)

	if parseErr != nil && parseErr.OrNil() != nil {
		t.Errorf("Parse() returned errors: %v", parseErr)
	}

	if len(tags) != 3 {
		t.Fatalf("Parse() returned %d tags, want 3", len(tags))
	}

	expectedTags := []struct {
		name  string
		value uint16
	}{
		{"Quality", 1},
		{"WhiteBalance", 0},
		{"SonyModelID", 456},
	}

	for i, expected := range expectedTags {
		if tags[i].Name != expected.name {
			t.Errorf("tags[%d].Name = %q, want %q", i, tags[i].Name, expected.name)
		}
		if val, ok := tags[i].Value.(uint16); !ok || val != expected.value {
			t.Errorf("tags[%d].Value = %v, want %d", i, tags[i].Value, expected.value)
		}
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
		got := getTypeSize(tt.typeVal)
		if got != tt.want {
			t.Errorf("getTypeSize(%d) = %d, want %d", tt.typeVal, got, tt.want)
		}
	}
}

func TestGetTypeName(t *testing.T) {
	tests := []struct {
		typeVal uint16
		want    string
	}{
		{1, "BYTE"},
		{2, "ASCII"},
		{3, "SHORT"},
		{4, "LONG"},
		{5, "RATIONAL"},
		{6, "SBYTE"},
		{7, "UNDEFINED"},
		{8, "SSHORT"},
		{9, "SLONG"},
		{10, "SRATIONAL"},
		{11, "FLOAT"},
		{12, "DOUBLE"},
		{99, "UNKNOWN"},
	}

	for _, tt := range tests {
		got := getTypeName(tt.typeVal)
		if got != tt.want {
			t.Errorf("getTypeName(%d) = %q, want %q", tt.typeVal, got, tt.want)
		}
	}
}
