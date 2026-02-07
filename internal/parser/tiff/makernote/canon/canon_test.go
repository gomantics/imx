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
		t.Errorf("Manufacturer() = %q, want %q", got, "Canon")
	}
}

func TestHandler_Detect(t *testing.T) {
	h := New()

	tests := []struct {
		name      string
		data      []byte
		wantMatch bool
	}{
		{
			name:      "valid Canon IFD with 2 entries",
			data:      makeCanonIFD(2),
			wantMatch: true,
		},
		{
			name:      "valid Canon IFD with 10 entries",
			data:      makeCanonIFD(10),
			wantMatch: true,
		},
		{
			name:      "too few entries (0)",
			data:      makeCanonIFD(0),
			wantMatch: false,
		},
		{
			name:      "too many entries (101)",
			data:      []byte{101, 0},
			wantMatch: false,
		},
		{
			name:      "too short",
			data:      []byte{2, 0}, // claims 2 entries but no data
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, cfg := h.Detect(tt.data)
			if match != tt.wantMatch {
				t.Errorf("Detect() = %v, want %v", match, tt.wantMatch)
			}
			if match {
				if cfg.IFDOffset != 0 {
					t.Errorf("cfg.IFDOffset = %d, want 0", cfg.IFDOffset)
				}
				if cfg.OffsetBase != makernote.OffsetAbsolute {
					t.Errorf("cfg.OffsetBase = %v, want OffsetAbsolute", cfg.OffsetBase)
				}
			}
		})
	}
}

func TestHandler_Parse(t *testing.T) {
	h := New()
	order := binary.LittleEndian

	tests := []struct {
		name        string
		data        []byte
		cfg         *makernote.Config
		wantTags    int
		wantErr     bool
		checkValues map[string]any
	}{
		{
			name: "parse ImageType tag",
			data: func() []byte {
				buf := new(bytes.Buffer)
				// Entry count: 1 (2 bytes)
				binary.Write(buf, order, uint16(1))
				// Entry: ImageType (0x0006), ASCII, count=10, offset=18 (12 bytes)
				// Offset = 2 (count) + 12 (entry) + 4 (next IFD) = 18
				writeIFDEntry(buf, order, TagImageType, 2, 10, 18)
				// Next IFD offset (ignored) (4 bytes)
				binary.Write(buf, order, uint32(0))
				// Actual string data at offset 18
				buf.WriteString("Canon EOS\x00")
				return buf.Bytes()
			}(),
			cfg: &makernote.Config{
				IFDOffset:  0,
				OffsetBase: makernote.OffsetAbsolute,
				ByteOrder:  order,
			},
			wantTags: 1,
			checkValues: map[string]any{
				"ImageType": "Canon EOS",
			},
		},
		{
			name: "parse SerialNumber tag",
			data: func() []byte {
				buf := new(bytes.Buffer)
				// Entry count: 1 (2 bytes)
				binary.Write(buf, order, uint16(1))
				// Entry: SerialNumber (0x000C), ASCII, count=12, offset=18 (12 bytes)
				writeIFDEntry(buf, order, TagSerialNumber, 2, 12, 18)
				// Next IFD offset (4 bytes)
				binary.Write(buf, order, uint32(0))
				// String data at offset 18
				buf.WriteString("123456789\x00\x00\x00")
				return buf.Bytes()
			}(),
			cfg: &makernote.Config{
				IFDOffset:  0,
				OffsetBase: makernote.OffsetAbsolute,
				ByteOrder:  order,
			},
			wantTags: 1,
			checkValues: map[string]any{
				"SerialNumber": "123456789",
			},
		},
		{
			name: "parse inline SHORT value",
			data: func() []byte {
				buf := new(bytes.Buffer)
				// Entry count: 1
				binary.Write(buf, order, uint16(1))
				// Entry: ModelID (0x0010), SHORT, count=1, value=0x1234 inline
				writeIFDEntry(buf, order, TagModelID, 3, 1, 0x1234)
				// Next IFD offset
				binary.Write(buf, order, uint32(0))
				return buf.Bytes()
			}(),
			cfg: &makernote.Config{
				IFDOffset:  0,
				OffsetBase: makernote.OffsetAbsolute,
				ByteOrder:  order,
			},
			wantTags: 1,
			checkValues: map[string]any{
				"ModelID": uint16(0x1234),
			},
		},
		{
			name: "parse CameraSettings1 short array",
			data: func() []byte {
				buf := new(bytes.Buffer)
				// Entry count: 1 (2 bytes)
				binary.Write(buf, order, uint16(1))
				// Entry: CameraSettings1 (0x0001), SHORT array, count=5, offset=18 (12 bytes)
				writeIFDEntry(buf, order, TagCameraSettings1, 3, 5, 18)
				// Next IFD offset (4 bytes)
				binary.Write(buf, order, uint32(0))
				// Array data at offset 18 (5 shorts = 10 bytes)
				for i := uint16(1); i <= 5; i++ {
					binary.Write(buf, order, i*100)
				}
				return buf.Bytes()
			}(),
			cfg: &makernote.Config{
				IFDOffset:  0,
				OffsetBase: makernote.OffsetAbsolute,
				ByteOrder:  order,
			},
			wantTags: 1,
			checkValues: map[string]any{
				"CameraSettings1": []uint16{100, 200, 300, 400, 500},
			},
		},
		{
			name: "parse multiple tags",
			data: func() []byte {
				buf := new(bytes.Buffer)
				// Entry count: 3 (2 bytes)
				binary.Write(buf, order, uint16(3))
				// Entry 1: ImageType at offset 42 (12 bytes)
				// Offset = 2 + 3*12 + 4 = 42
				writeIFDEntry(buf, order, TagImageType, 2, 10, 42)
				// Entry 2: ModelID inline (12 bytes)
				writeIFDEntry(buf, order, TagModelID, 3, 1, 0x80000001)
				// Entry 3: LensModel at offset 52 (12 bytes)
				writeIFDEntry(buf, order, TagLensModel, 2, 15, 52)
				// Next IFD offset (4 bytes)
				binary.Write(buf, order, uint32(0))
				// ImageType data at offset 42 (10 bytes)
				buf.WriteString("Canon EOS\x00")
				// LensModel data at offset 52 (15 bytes)
				buf.WriteString("EF 50mm f/1.4\x00\x00")
				return buf.Bytes()
			}(),
			cfg: &makernote.Config{
				IFDOffset:  0,
				OffsetBase: makernote.OffsetAbsolute,
				ByteOrder:  order,
			},
			wantTags: 3,
		},
		{
			name: "invalid entry count 0",
			data: func() []byte {
				buf := new(bytes.Buffer)
				binary.Write(buf, order, uint16(0))
				return buf.Bytes()
			}(),
			cfg: &makernote.Config{
				IFDOffset:  0,
				OffsetBase: makernote.OffsetAbsolute,
				ByteOrder:  order,
			},
			wantTags: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			tags, parseErr := h.Parse(reader, 0, 0, tt.cfg)

			if tt.wantErr {
				if parseErr == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if len(tags) != tt.wantTags {
				t.Errorf("got %d tags, want %d", len(tags), tt.wantTags)
			}

			// Check specific values
			for wantName, wantValue := range tt.checkValues {
				found := false
				for _, tag := range tags {
					if tag.Name == wantName {
						found = true
						if !compareValues(tag.Value, wantValue) {
							t.Errorf("tag %s: got %v (%T), want %v (%T)",
								wantName, tag.Value, tag.Value, wantValue, wantValue)
						}
						break
					}
				}
				if !found {
					t.Errorf("tag %s not found", wantName)
				}
			}
		})
	}
}

func TestHandler_TagName(t *testing.T) {
	h := New()

	tests := []struct {
		tagID uint16
		want  string
	}{
		{TagCameraSettings1, "CameraSettings1"},
		{TagImageType, "ImageType"},
		{TagSerialNumber, "SerialNumber"},
		{TagLensModel, "LensModel"},
		{0xFFFF, ""}, // Unknown
	}

	for _, tt := range tests {
		got := h.TagName(tt.tagID)
		if got != tt.want {
			t.Errorf("TagName(0x%04X) = %q, want %q", tt.tagID, got, tt.want)
		}
	}
}

func TestGetTagName(t *testing.T) {
	tests := []struct {
		tagID uint16
		want  string
	}{
		{TagCameraSettings1, "CameraSettings1"},
		{TagCameraSettings2, "CameraSettings2"},
		{TagImageType, "ImageType"},
		{TagFirmwareVersion, "FirmwareVersion"},
		{TagOwnerName, "OwnerName"},
		{TagSerialNumber, "SerialNumber"},
		{TagModelID, "ModelID"},
		{TagLensModel, "LensModel"},
		{TagColorData, "ColorData"},
		{0x9999, ""}, // Unknown
	}

	for _, tt := range tests {
		got := GetTagName(tt.tagID)
		if got != tt.want {
			t.Errorf("GetTagName(0x%04X) = %q, want %q", tt.tagID, got, tt.want)
		}
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
		got := getTypeSize(tt.tagType)
		if got != tt.want {
			t.Errorf("getTypeSize(%d) = %d, want %d", tt.tagType, got, tt.want)
		}
	}
}

func TestGetTypeName(t *testing.T) {
	tests := []struct {
		tagType uint16
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
		got := getTypeName(tt.tagType)
		if got != tt.want {
			t.Errorf("getTypeName(%d) = %q, want %q", tt.tagType, got, tt.want)
		}
	}
}

// Helper functions

func makeCanonIFD(entryCount uint16) []byte {
	order := binary.LittleEndian
	// IFD: 2-byte count + 12-byte entries + 4-byte next offset
	size := 2 + int(entryCount)*12 + 4
	data := make([]byte, size)
	order.PutUint16(data[0:2], entryCount)
	// Fill entries with dummy data
	for i := uint16(0); i < entryCount; i++ {
		offset := 2 + int(i)*12
		order.PutUint16(data[offset:offset+2], 0x0001+i)   // Tag ID
		order.PutUint16(data[offset+2:offset+4], 3)        // Type: SHORT
		order.PutUint32(data[offset+4:offset+8], 1)        // Count
		order.PutUint32(data[offset+8:offset+12], uint32(i)) // Value
	}
	return data
}

func writeIFDEntry(buf *bytes.Buffer, order binary.ByteOrder, tag, tagType uint16, count, value uint32) {
	binary.Write(buf, order, tag)
	binary.Write(buf, order, tagType)
	binary.Write(buf, order, count)
	binary.Write(buf, order, value)
}

func compareValues(got, want any) bool {
	switch w := want.(type) {
	case string:
		g, ok := got.(string)
		return ok && g == w
	case uint16:
		g, ok := got.(uint16)
		return ok && g == w
	case uint32:
		g, ok := got.(uint32)
		return ok && g == w
	case []uint16:
		g, ok := got.([]uint16)
		if !ok || len(g) != len(w) {
			return false
		}
		for i := range w {
			if g[i] != w[i] {
				return false
			}
		}
		return true
	default:
		return false
	}
}
