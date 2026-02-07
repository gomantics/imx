package fujifilm

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/parser/tiff/makernote"
)

func TestHandler_Manufacturer(t *testing.T) {
	h := New()
	if got := h.Manufacturer(); got != "Fujifilm" {
		t.Errorf("Manufacturer() = %q, want %q", got, "Fujifilm")
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
			name: "FUJIFILM header with offset 12",
			data: func() []byte {
				data := make([]byte, 100)
				copy(data[0:8], []byte("FUJIFILM"))
				binary.LittleEndian.PutUint32(data[8:12], 12) // IFD offset
				return data
			}(),
			wantOK: true,
			wantCfg: &makernote.Config{
				IFDOffset:  12,
				OffsetBase: makernote.OffsetRelativeToMakerNote,
				ByteOrder:  binary.LittleEndian,
				HasNextIFD: false,
				Variant:    "Standard",
			},
		},
		{
			name: "FUJIFILM header with offset 20",
			data: func() []byte {
				data := make([]byte, 100)
				copy(data[0:8], []byte("FUJIFILM"))
				binary.LittleEndian.PutUint32(data[8:12], 20) // Different IFD offset
				return data
			}(),
			wantOK: true,
			wantCfg: &makernote.Config{
				IFDOffset:  20,
				OffsetBase: makernote.OffsetRelativeToMakerNote,
				ByteOrder:  binary.LittleEndian,
			},
		},
		{
			name:   "not Fujifilm - Sony",
			data:   []byte("SONY DSC xxxxxxx"),
			wantOK: false,
		},
		{
			name:   "not Fujifilm - Canon",
			data:   []byte{0x05, 0x00, 0x01, 0x00, 0x02, 0x00},
			wantOK: false,
		},
		{
			name:   "too short",
			data:   []byte("FUJI"),
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
		{0x0010, "SerialNumber"},
		{0x1000, "Quality"},
		{0x1021, "FocusMode"},
		{0x1022, "AFMode"},
		{0x1401, "FilmMode"},
		{0x1002, "WhiteBalance"},
		{0x1400, "DynamicRange"},
		{0x9999, ""}, // Unknown tag
	}

	for _, tt := range tests {
		name := tt.expected
		if name == "" {
			name = "unknown"
		}
		t.Run(name, func(t *testing.T) {
			got := h.TagName(tt.tagID)
			if got != tt.expected {
				t.Errorf("TagName(0x%04X) = %q, want %q", tt.tagID, got, tt.expected)
			}
		})
	}
}

func TestHandler_Parse(t *testing.T) {
	h := New()

	// Build a minimal Fujifilm MakerNote with header and one IFD entry
	// Header: "FUJIFILM" (8 bytes) + IFD offset (4 bytes)
	// Entry count: 1 (2 bytes)
	// Entry: tag=0x1000, type=SHORT(3), count=1, value=1 (Quality)
	data := make([]byte, 100)
	copy(data[0:8], []byte("FUJIFILM"))
	binary.LittleEndian.PutUint32(data[8:12], 12) // IFD at offset 12

	// IFD at offset 12
	binary.LittleEndian.PutUint16(data[12:14], 1) // 1 entry

	// Entry at offset 14
	binary.LittleEndian.PutUint16(data[14:16], 0x1000) // Tag: Quality
	binary.LittleEndian.PutUint16(data[16:18], 3)      // Type: SHORT
	binary.LittleEndian.PutUint32(data[18:22], 1)      // Count: 1
	binary.LittleEndian.PutUint32(data[22:26], 1)      // Value: 1 (inline)

	reader := bytes.NewReader(data)
	cfg := &makernote.Config{
		IFDOffset:  12,
		OffsetBase: makernote.OffsetRelativeToMakerNote,
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
	if tag.Name != "Quality" {
		t.Errorf("Tag.Name = %q, want %q", tag.Name, "Quality")
	}
	if tag.ID != "Fujifilm:0x1000" {
		t.Errorf("Tag.ID = %q, want %q", tag.ID, "Fujifilm:0x1000")
	}
	if val, ok := tag.Value.(uint16); !ok || val != 1 {
		t.Errorf("Tag.Value = %v (%T), want 1 (uint16)", tag.Value, tag.Value)
	}
}

func TestHandler_Parse_MultipleEntries(t *testing.T) {
	h := New()

	// Build Fujifilm MakerNote with 3 entries
	data := make([]byte, 200)
	copy(data[0:8], []byte("FUJIFILM"))
	binary.LittleEndian.PutUint32(data[8:12], 12) // IFD at offset 12

	// IFD at offset 12
	binary.LittleEndian.PutUint16(data[12:14], 3) // 3 entries

	// Entry 1: Quality (0x1000), SHORT, value=2
	binary.LittleEndian.PutUint16(data[14:16], 0x1000)
	binary.LittleEndian.PutUint16(data[16:18], 3)
	binary.LittleEndian.PutUint32(data[18:22], 1)
	binary.LittleEndian.PutUint32(data[22:26], 2)

	// Entry 2: FocusMode (0x1021), SHORT, value=1
	binary.LittleEndian.PutUint16(data[26:28], 0x1021)
	binary.LittleEndian.PutUint16(data[28:30], 3)
	binary.LittleEndian.PutUint32(data[30:34], 1)
	binary.LittleEndian.PutUint32(data[34:38], 1)

	// Entry 3: FilmMode (0x1401), SHORT, value=0
	binary.LittleEndian.PutUint16(data[38:40], 0x1401)
	binary.LittleEndian.PutUint16(data[40:42], 3)
	binary.LittleEndian.PutUint32(data[42:46], 1)
	binary.LittleEndian.PutUint32(data[46:50], 0)

	reader := bytes.NewReader(data)
	cfg := &makernote.Config{
		IFDOffset:  12,
		OffsetBase: makernote.OffsetRelativeToMakerNote,
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
		{"Quality", 2},
		{"FocusMode", 1},
		{"FilmMode", 0},
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

func TestHandler_Parse_RelativeOffset(t *testing.T) {
	h := New()

	// Test that offsets are correctly calculated relative to MakerNote start
	// Build MakerNote with one tag that has value at external offset
	data := make([]byte, 200)
	copy(data[0:8], []byte("FUJIFILM"))
	binary.LittleEndian.PutUint32(data[8:12], 12)

	// IFD at offset 12
	binary.LittleEndian.PutUint16(data[12:14], 1)

	// Entry: SerialNumber (0x0010), ASCII, count=10
	// Value offset 100 (relative to MakerNote start, so absolute 100)
	binary.LittleEndian.PutUint16(data[14:16], 0x0010) // Tag: SerialNumber
	binary.LittleEndian.PutUint16(data[16:18], 2)      // Type: ASCII
	binary.LittleEndian.PutUint32(data[18:22], 10)     // Count: 10
	binary.LittleEndian.PutUint32(data[22:26], 100)    // Offset: 100 (relative)

	// Put serial number at offset 100
	copy(data[100:110], []byte("1234567890"))

	reader := bytes.NewReader(data)
	cfg := &makernote.Config{
		IFDOffset:  12,
		OffsetBase: makernote.OffsetRelativeToMakerNote,
		ByteOrder:  binary.LittleEndian,
	}

	// makerNoteOffset = 0, so relative offset 100 becomes absolute offset 100
	tags, parseErr := h.Parse(reader, 0, 0, cfg)

	if parseErr != nil && parseErr.OrNil() != nil {
		t.Errorf("Parse() returned errors: %v", parseErr)
	}

	if len(tags) != 1 {
		t.Fatalf("Parse() returned %d tags, want 1", len(tags))
	}

	if tags[0].Name != "SerialNumber" {
		t.Errorf("Tag.Name = %q, want %q", tags[0].Name, "SerialNumber")
	}
	if val, ok := tags[0].Value.(string); !ok || val != "1234567890" {
		t.Errorf("Tag.Value = %q, want %q", tags[0].Value, "1234567890")
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
