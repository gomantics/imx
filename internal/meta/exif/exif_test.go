package exif

import (
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/common"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Error("New() returned nil")
	}
}

func TestParser_Spec(t *testing.T) {
	p := New()
	if p.Spec() != common.SpecEXIF {
		t.Errorf("Spec() = %v, want %v", p.Spec(), common.SpecEXIF)
	}
}

// buildTIFF creates a TIFF structure for testing
func buildTIFF(bigEndian bool, entries []ifdEntry) []byte {
	var buf []byte
	var byteOrder binary.ByteOrder

	// Byte order marker
	if bigEndian {
		buf = append(buf, 'M', 'M')
		byteOrder = binary.BigEndian
	} else {
		buf = append(buf, 'I', 'I')
		byteOrder = binary.LittleEndian
	}

	// TIFF magic number (42)
	magic := make([]byte, 2)
	byteOrder.PutUint16(magic, 42)
	buf = append(buf, magic...)

	// Offset to first IFD (starts at byte 8 for us)
	offset := make([]byte, 4)
	byteOrder.PutUint32(offset, 8)
	buf = append(buf, offset...)

	// IFD0
	entryCount := make([]byte, 2)
	byteOrder.PutUint16(entryCount, uint16(len(entries)))
	buf = append(buf, entryCount...)

	// Write each entry (12 bytes each)
	for _, entry := range entries {
		entryBuf := make([]byte, 12)
		byteOrder.PutUint16(entryBuf[0:2], entry.tagID)
		byteOrder.PutUint16(entryBuf[2:4], entry.dataType)
		byteOrder.PutUint32(entryBuf[4:8], entry.count)
		copy(entryBuf[8:12], entry.valueOrOffset)
		buf = append(buf, entryBuf...)
	}

	// Next IFD offset (0 = no more IFDs)
	nextIFD := make([]byte, 4)
	buf = append(buf, nextIFD...)

	return buf
}

type ifdEntry struct {
	tagID         uint16 // TIFF tag identifier
	dataType      uint16 // TIFF data type (1=BYTE, 2=ASCII, 3=SHORT, 4=LONG, 5=RATIONAL, etc.)
	count         uint32 // Number of values
	valueOrOffset []byte // Value (if ≤4 bytes) or offset to value data
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []common.RawBlock
		wantDirs int
		wantErr  bool
		checkFn  func(t *testing.T, dirs []common.Directory)
	}{
		{
			name:     "empty blocks returns empty",
			blocks:   []common.RawBlock{},
			wantDirs: 0,
			wantErr:  false,
		},
		{
			name: "non-EXIF blocks are ignored",
			blocks: []common.RawBlock{
				{Spec: common.SpecXMP, Payload: []byte("some xmp data")},
				{Spec: common.SpecICC, Payload: []byte("some icc data")},
			},
			wantDirs: 0,
			wantErr:  false,
		},
		{
			name: "valid EXIF with little-endian",
			blocks: []common.RawBlock{
				{
					Spec:    common.SpecEXIF,
					Payload: buildTIFF(false, nil),
				},
			},
			wantDirs: 1,
			wantErr:  false,
			checkFn: func(t *testing.T, dirs []common.Directory) {
				if dirs[0].Name != "IFD0" {
					t.Errorf("Directory name = %q, want %q", dirs[0].Name, "IFD0")
				}
			},
		},
		{
			name: "valid EXIF with big-endian",
			blocks: []common.RawBlock{
				{
					Spec:    common.SpecEXIF,
					Payload: buildTIFF(true, nil),
				},
			},
			wantDirs: 1,
			wantErr:  false,
		},
		{
			name: "TIFF header too short",
			blocks: []common.RawBlock{
				{
					Spec:    common.SpecEXIF,
					Payload: []byte{0x49, 0x49, 0x2A, 0x00}, // Only 4 bytes
				},
			},
			wantDirs: 0,
			wantErr:  true,
		},
		{
			name: "invalid byte order",
			blocks: []common.RawBlock{
				{
					Spec:    common.SpecEXIF,
					Payload: []byte{'X', 'X', 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00},
				},
			},
			wantDirs: 0,
			wantErr:  true,
		},
		{
			name: "invalid TIFF magic number",
			blocks: []common.RawBlock{
				{
					Spec:    common.SpecEXIF,
					Payload: []byte{'I', 'I', 0x00, 0x00, 0x08, 0x00, 0x00, 0x00}, // Magic = 0, not 42
				},
			},
			wantDirs: 0,
			wantErr:  true,
		},
		{
			name: "IFD offset beyond data",
			blocks: []common.RawBlock{
				{
					Spec:    common.SpecEXIF,
					Payload: []byte{'I', 'I', 0x2A, 0x00, 0xFF, 0x00, 0x00, 0x00}, // Offset = 255
				},
			},
			wantDirs: 0, // IFD offset out of bounds, no dirs parsed
			wantErr:  false,
		},
		{
			name: "IFD offset at zero skips parsing",
			blocks: []common.RawBlock{
				{
					Spec:    common.SpecEXIF,
					Payload: []byte{'I', 'I', 0x2A, 0x00, 0x00, 0x00, 0x00, 0x00}, // Offset = 0
				},
			},
			wantDirs: 0,
			wantErr:  false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs, err := p.Parse(tt.blocks)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if len(dirs) != tt.wantDirs {
				t.Errorf("Parse() returned %d dirs, want %d", len(dirs), tt.wantDirs)
				return
			}

			if tt.checkFn != nil {
				tt.checkFn(t, dirs)
			}
		})
	}
}

func TestParser_ParseWithTags(t *testing.T) {
	p := New()

	// Build a more complete TIFF with string data
	tiff := buildTIFFWithStrings(false, []tagWithValue{
		{tagID: 0x010F, value: "Canon"},  // Make
		{tagID: 0x0110, value: "EOS 5D"}, // Model
	})

	blocks := []common.RawBlock{
		{
			Spec:    common.SpecEXIF,
			Payload: tiff,
		},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) == 0 {
		t.Fatal("Parse() returned 0 dirs")
	}

	// Check for Make tag
	if tag, ok := dirs[0].Tags["EXIF:Make"]; ok {
		if tag.Value != "Canon" {
			t.Errorf("Make value = %v, want %q", tag.Value, "Canon")
		}
	} else {
		t.Error("Make tag not found")
	}
}

type tagWithValue struct {
	tagID uint16
	value string
}

// buildTIFFWithStrings creates a TIFF with string tags stored at offsets beyond the IFD
func buildTIFFWithStrings(bigEndian bool, tags []tagWithValue) []byte {
	var buf []byte
	var byteOrder binary.ByteOrder

	// Byte order marker
	if bigEndian {
		buf = append(buf, 'M', 'M')
		byteOrder = binary.BigEndian
	} else {
		buf = append(buf, 'I', 'I')
		byteOrder = binary.LittleEndian
	}

	// TIFF magic number (42)
	magic := make([]byte, 2)
	byteOrder.PutUint16(magic, 42)
	buf = append(buf, magic...)

	// Offset to first IFD (starts at byte 8)
	offset := make([]byte, 4)
	byteOrder.PutUint32(offset, 8)
	buf = append(buf, offset...)

	// IFD entry count
	entryCount := make([]byte, 2)
	byteOrder.PutUint16(entryCount, uint16(len(tags)))
	buf = append(buf, entryCount...)

	// Calculate where string data will start
	// Header (8) + count (2) + entries (12 each) + next IFD offset (4)
	stringDataOffset := 8 + 2 + len(tags)*12 + 4

	// Collect string data
	var stringData []byte

	// Write entries
	for _, tag := range tags {
		entry := make([]byte, 12)
		byteOrder.PutUint16(entry[0:2], tag.tagID)
		byteOrder.PutUint16(entry[2:4], 2)  // ASCII type
		count := uint32(len(tag.value) + 1) // Include null terminator
		byteOrder.PutUint32(entry[4:8], count)

		if count <= 4 {
			// Value fits in offset field
			copy(entry[8:12], []byte(tag.value+"\x00"))
		} else {
			// Store offset to string data
			byteOrder.PutUint32(entry[8:12], uint32(stringDataOffset+len(stringData)))
			stringData = append(stringData, []byte(tag.value+"\x00")...)
		}
		buf = append(buf, entry...)
	}

	// Next IFD offset (0 = no more)
	nextIFD := make([]byte, 4)
	buf = append(buf, nextIFD...)

	// Append string data
	buf = append(buf, stringData...)

	return buf
}

func TestParser_ParseValue(t *testing.T) {
	p := New()
	byteOrder := binary.LittleEndian

	tests := []struct {
		name      string
		tagType   uint16
		count     uint32
		data      []byte
		offset    int
		wantValue any
		wantType  string
	}{
		{
			name:      "BYTE single value",
			tagType:   1,
			count:     1,
			data:      []byte{0x42, 0x00, 0x00, 0x00},
			offset:    0,
			wantValue: 66,
			wantType:  "byte",
		},
		{
			name:      "BYTE multiple values",
			tagType:   1,
			count:     3,
			data:      []byte{0x01, 0x02, 0x03, 0x00},
			offset:    0,
			wantValue: []byte{0x01, 0x02, 0x03},
			wantType:  "bytes",
		},
		{
			name:      "ASCII string (fits in 4 bytes)",
			tagType:   2,
			count:     4,
			data:      []byte{0x41, 0x42, 0x43, 0x00}, // "ABC\0"
			offset:    0,
			wantValue: "ABC",
			wantType:  "string",
		},
		{
			name:      "SHORT single value",
			tagType:   3,
			count:     1,
			data:      []byte{0x64, 0x00, 0x00, 0x00}, // 100
			offset:    0,
			wantValue: 100,
			wantType:  "short",
		},
		{
			name:      "SHORT multiple values (fits in 4 bytes)",
			tagType:   3,
			count:     2,
			data:      []byte{0x64, 0x00, 0xC8, 0x00}, // 100, 200
			offset:    0,
			wantValue: []int{100, 200},
			wantType:  "shorts",
		},
		{
			name:      "LONG single value",
			tagType:   4,
			count:     1,
			data:      []byte{0xE8, 0x03, 0x00, 0x00}, // 1000
			offset:    0,
			wantValue: 1000,
			wantType:  "long",
		},
		{
			name:    "LONG multiple values at offset",
			tagType: 4,
			count:   2,
			// First 4 bytes: offset=4 pointing to remaining data
			// Remaining: two LONGs (1000, 2000)
			data:      []byte{0x04, 0x00, 0x00, 0x00, 0xE8, 0x03, 0x00, 0x00, 0xD0, 0x07, 0x00, 0x00},
			offset:    0,
			wantValue: []int{1000, 2000},
			wantType:  "longs",
		},
		{
			name:    "RATIONAL single value at offset",
			tagType: 5,
			count:   1,
			// First 4 bytes: offset=4 pointing to rational data
			// Remaining: num=100, denom=10
			data:      []byte{0x04, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00},
			offset:    0,
			wantValue: 10.0,
			wantType:  "rational",
		},
		{
			name:    "RATIONAL with zero denominator",
			tagType: 5,
			count:   1,
			// First 4 bytes: offset=4 pointing to rational data
			// Remaining: num=100, denom=0
			data:      []byte{0x04, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			offset:    0,
			wantValue: 0.0,
			wantType:  "rational",
		},
		{
			name:    "RATIONAL multiple values at offset",
			tagType: 5,
			count:   2,
			// First 4 bytes: offset=4 pointing to rational data
			// Remaining: two rationals (100/10, 200/20)
			data:      []byte{0x04, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0xC8, 0x00, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00},
			offset:    0,
			wantValue: []float64{10.0, 10.0},
			wantType:  "rationals",
		},
		{
			name:    "RATIONAL multiple with zero denom",
			tagType: 5,
			count:   2,
			// First 4 bytes: offset=4 pointing to rational data
			// Remaining: two rationals (100/0, 200/20)
			data:      []byte{0x04, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC8, 0x00, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00},
			offset:    0,
			wantValue: []float64{0.0, 10.0},
			wantType:  "rationals",
		},
		{
			name:      "UNDEFINED",
			tagType:   7,
			count:     4,
			data:      []byte{0x01, 0x02, 0x03, 0x04},
			offset:    0,
			wantValue: []byte{0x01, 0x02, 0x03, 0x04},
			wantType:  "undefined",
		},
		{
			name:      "unknown type",
			tagType:   99,
			count:     1,
			data:      []byte{0x00, 0x00, 0x00, 0x00},
			offset:    0,
			wantValue: nil,
			wantType:  "unknown",
		},
		{
			name:      "invalid offset for large value",
			tagType:   4,
			count:     2,                              // Needs 8 bytes, stored at offset
			data:      []byte{0xFF, 0x00, 0x00, 0x00}, // Invalid offset (255) - beyond data
			offset:    0,
			wantValue: nil,
			wantType:  "invalid_offset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, typeName := p.parseValue(tt.data, tt.tagType, tt.count, tt.offset, byteOrder)

			if typeName != tt.wantType {
				t.Errorf("parseValue() type = %q, want %q", typeName, tt.wantType)
			}

			// For byte slices, compare contents
			if wantBytes, ok := tt.wantValue.([]byte); ok {
				gotBytes, ok := value.([]byte)
				if !ok {
					t.Errorf("parseValue() value type = %T, want []byte", value)
					return
				}
				if len(gotBytes) != len(wantBytes) {
					t.Errorf("parseValue() value len = %d, want %d", len(gotBytes), len(wantBytes))
					return
				}
				for i := range wantBytes {
					if gotBytes[i] != wantBytes[i] {
						t.Errorf("parseValue() value[%d] = %v, want %v", i, gotBytes[i], wantBytes[i])
					}
				}
				return
			}

			// For int slices
			if wantInts, ok := tt.wantValue.([]int); ok {
				gotInts, ok := value.([]int)
				if !ok {
					t.Errorf("parseValue() value type = %T, want []int", value)
					return
				}
				if len(gotInts) != len(wantInts) {
					t.Errorf("parseValue() value len = %d, want %d", len(gotInts), len(wantInts))
					return
				}
				for i := range wantInts {
					if gotInts[i] != wantInts[i] {
						t.Errorf("parseValue() value[%d] = %v, want %v", i, gotInts[i], wantInts[i])
					}
				}
				return
			}

			// For float64 slices
			if wantFloats, ok := tt.wantValue.([]float64); ok {
				gotFloats, ok := value.([]float64)
				if !ok {
					t.Errorf("parseValue() value type = %T, want []float64", value)
					return
				}
				if len(gotFloats) != len(wantFloats) {
					t.Errorf("parseValue() value len = %d, want %d", len(gotFloats), len(wantFloats))
					return
				}
				for i := range wantFloats {
					if gotFloats[i] != wantFloats[i] {
						t.Errorf("parseValue() value[%d] = %v, want %v", i, gotFloats[i], wantFloats[i])
					}
				}
				return
			}

			// Simple equality check
			if value != tt.wantValue {
				t.Errorf("parseValue() value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

func TestParser_ParseIFD_OutOfBounds(t *testing.T) {
	p := New()
	data := make([]byte, 10)

	// Test offset out of bounds
	_, _, err := p.parseIFD(data, 20, binary.LittleEndian, "Test")
	if err == nil {
		t.Error("parseIFD() expected error for out of bounds offset")
	}
}

func TestParser_ParseIFD_TruncatedEntries(t *testing.T) {
	p := New()

	// Create minimal data with entry count claiming more entries than available
	data := make([]byte, 20)
	binary.LittleEndian.PutUint16(data[0:2], 5) // Claim 5 entries, but only have space for ~1

	dir, _, err := p.parseIFD(data, 0, binary.LittleEndian, "Test")
	if err != nil {
		t.Errorf("parseIFD() unexpected error = %v", err)
	}

	// Should have parsed whatever entries fit
	if len(dir.Tags) > 1 {
		t.Logf("parseIFD() parsed %d tags despite truncated data", len(dir.Tags))
	}
}

func TestParser_ParseWithExifSubIFD(t *testing.T) {
	p := New()

	// Build TIFF with ExifOffset pointer
	var buf []byte
	byteOrder := binary.LittleEndian

	// Header
	buf = append(buf, 'I', 'I') // Little-endian
	tmp := make([]byte, 2)
	byteOrder.PutUint16(tmp, 42)
	buf = append(buf, tmp...) // TIFF magic

	tmp = make([]byte, 4)
	byteOrder.PutUint32(tmp, 8)
	buf = append(buf, tmp...) // IFD0 offset

	// IFD0 with ExifOffset tag
	byteOrder.PutUint16(tmp[:2], 1) // 1 entry
	buf = append(buf, tmp[:2]...)

	// ExifOffset entry (tag 0x8769 = ExifOffset)
	entry := make([]byte, 12)
	byteOrder.PutUint16(entry[0:2], 0x8769) // ExifOffset tag
	byteOrder.PutUint16(entry[2:4], 4)      // LONG type
	byteOrder.PutUint32(entry[4:8], 1)      // count = 1
	byteOrder.PutUint32(entry[8:12], 26)    // Offset to ExifIFD
	buf = append(buf, entry...)

	// Next IFD offset = 0
	buf = append(buf, 0, 0, 0, 0)

	// ExifIFD at offset 26
	byteOrder.PutUint16(tmp[:2], 0) // 0 entries
	buf = append(buf, tmp[:2]...)
	buf = append(buf, 0, 0, 0, 0) // Next IFD = 0

	blocks := []common.RawBlock{
		{Spec: common.SpecEXIF, Payload: buf},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should have IFD0 and ExifIFD
	if len(dirs) < 2 {
		t.Errorf("Parse() returned %d dirs, want at least 2", len(dirs))
	}
}

func TestParser_ParseWithGPSSubIFD(t *testing.T) {
	p := New()

	// Build TIFF with GPSInfo pointer
	var buf []byte
	byteOrder := binary.LittleEndian

	// Header
	buf = append(buf, 'I', 'I')
	tmp := make([]byte, 2)
	byteOrder.PutUint16(tmp, 42)
	buf = append(buf, tmp...)

	tmp = make([]byte, 4)
	byteOrder.PutUint32(tmp, 8)
	buf = append(buf, tmp...)

	// IFD0 with GPSInfo tag
	byteOrder.PutUint16(tmp[:2], 1)
	buf = append(buf, tmp[:2]...)

	// GPSInfo entry (tag 0x8825 = GPSInfo)
	entry := make([]byte, 12)
	byteOrder.PutUint16(entry[0:2], 0x8825) // GPSInfo tag
	byteOrder.PutUint16(entry[2:4], 4)      // LONG type
	byteOrder.PutUint32(entry[4:8], 1)      // count = 1
	byteOrder.PutUint32(entry[8:12], 26)    // Offset to GPS IFD
	buf = append(buf, entry...)

	// Next IFD offset = 0
	buf = append(buf, 0, 0, 0, 0)

	// GPS IFD at offset 26
	byteOrder.PutUint16(tmp[:2], 1) // 1 entry
	buf = append(buf, tmp[:2]...)

	// GPS tag entry (0x0001 = GPSLatitudeRef)
	gpsEntry := make([]byte, 12)
	byteOrder.PutUint16(gpsEntry[0:2], 0x0001) // GPSLatitudeRef tag
	byteOrder.PutUint16(gpsEntry[2:4], 2)      // ASCII type
	byteOrder.PutUint32(gpsEntry[4:8], 2)      // count = 2 (includes null terminator)
	copy(gpsEntry[8:12], []byte("N\x00\x00\x00"))
	buf = append(buf, gpsEntry...)

	buf = append(buf, 0, 0, 0, 0) // Next IFD = 0

	blocks := []common.RawBlock{
		{Spec: common.SpecEXIF, Payload: buf},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should have IFD0 and GPS
	if len(dirs) < 2 {
		t.Errorf("Parse() returned %d dirs, want at least 2", len(dirs))
	}

	// Check GPS tag was parsed correctly
	for _, dir := range dirs {
		if dir.Name == "GPS" {
			if tag, ok := dir.Tags["EXIF:GPSLatitudeRef"]; ok {
				if tag.Value != "N" {
					t.Errorf("GPSLatitudeRef = %v, want %q", tag.Value, "N")
				}
			}
		}
	}
}

func TestParser_ParseWithIFD1(t *testing.T) {
	p := New()

	// Build TIFF with IFD1 (thumbnail)
	var buf []byte
	byteOrder := binary.LittleEndian

	// Header
	buf = append(buf, 'I', 'I')
	tmp := make([]byte, 2)
	byteOrder.PutUint16(tmp, 42)
	buf = append(buf, tmp...)

	tmp = make([]byte, 4)
	byteOrder.PutUint32(tmp, 8)
	buf = append(buf, tmp...)

	// IFD0 with 0 entries
	byteOrder.PutUint16(tmp[:2], 0)
	buf = append(buf, tmp[:2]...)

	// Next IFD offset pointing to IFD1
	byteOrder.PutUint32(tmp, 14) // IFD1 at offset 14
	buf = append(buf, tmp...)

	// IFD1 with 0 entries
	byteOrder.PutUint16(tmp[:2], 0)
	buf = append(buf, tmp[:2]...)
	buf = append(buf, 0, 0, 0, 0)

	blocks := []common.RawBlock{
		{Spec: common.SpecEXIF, Payload: buf},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should have IFD0 and IFD1
	if len(dirs) != 2 {
		t.Errorf("Parse() returned %d dirs, want 2", len(dirs))
	}

	foundIFD1 := false
	for _, dir := range dirs {
		if dir.Name == "IFD1" {
			foundIFD1 = true
		}
	}
	if !foundIFD1 {
		t.Error("IFD1 not found in directories")
	}
}

func TestParser_ParseEntry_UnknownTag(t *testing.T) {
	p := New()

	// Create entry with unknown tag ID
	data := make([]byte, 12)
	byteOrder := binary.LittleEndian
	byteOrder.PutUint16(data[0:2], 0xFFFF) // Unknown/undefined tag ID
	byteOrder.PutUint16(data[2:4], 3)      // SHORT type
	byteOrder.PutUint32(data[4:8], 1)      // count = 1
	byteOrder.PutUint16(data[8:10], 42)    // value = 42

	tag := p.parseEntry(data, 0, byteOrder, "IFD0")

	if tag.Name != "TagFFFF" {
		t.Errorf("Unknown tag name = %q, want %q", tag.Name, "TagFFFF")
	}
}

func TestParser_ParseEntry_GPSTags(t *testing.T) {
	p := New()

	// Create entry with GPS tag
	data := make([]byte, 12)
	byteOrder := binary.LittleEndian
	byteOrder.PutUint16(data[0:2], 0x0001) // 0x0001 = GPSLatitudeRef tag
	byteOrder.PutUint16(data[2:4], 2)      // ASCII type
	byteOrder.PutUint32(data[4:8], 2)      // count = 2 (includes null terminator)
	copy(data[8:12], []byte("N\x00\x00\x00"))

	tag := p.parseEntry(data, 0, byteOrder, "GPS")

	if tag.Name != "GPSLatitudeRef" {
		t.Errorf("GPS tag name = %q, want %q", tag.Name, "GPSLatitudeRef")
	}
}

func TestParser_Parse_IFD0ParseError(t *testing.T) {
	p := New()

	// Create TIFF header where IFD0 offset passes check but parseIFD fails
	// We need: ifd0Offset < len(data) to enter the block
	// Then parseIFD needs offset+2 > len(data) to fail
	// With offset=8 and len=9, check passes (8<9) but parseIFD fails (8+2=10>9)
	data := []byte{
		'I', 'I', // Little-endian
		0x2A, 0x00, // TIFF magic
		0x08, 0x00, 0x00, 0x00, // IFD0 offset = 8
		0x00, // Only 1 byte after offset, need 2 for entry count
	}

	blocks := []common.RawBlock{
		{Spec: common.SpecEXIF, Payload: data},
	}

	_, err := p.Parse(blocks)
	if err == nil {
		t.Error("expected error when parseIFD fails due to truncated entry count, got nil")
	}
}

func TestParser_ParseValue_SBYTE(t *testing.T) {
	p := New()

	// Type 6 (SBYTE) is now properly handled by SByteParser
	data := make([]byte, 20)
	byteOrder := binary.LittleEndian

	// Put value data at offset 0 (signed byte value: -1)
	copy(data[0:4], []byte{0xFF, 0x00, 0x00, 0x00})

	// Type 6 is SBYTE - should be parsed correctly now
	value, typeName := p.parseValue(data, 6, 1, 0, byteOrder)

	if typeName != "sbyte" {
		t.Errorf("parseValue() typeName = %q, want %q", typeName, "sbyte")
	}
	if value != -1 {
		t.Errorf("parseValue() value = %v, want -1", value)
	}
}

func TestParser_ParseValue_TrulyUnknownType(t *testing.T) {
	p := New()

	// Type 99 doesn't exist in TIFF spec - not in TIFFTypeSizes
	data := make([]byte, 20)
	byteOrder := binary.LittleEndian

	// Put some data at offset 0
	copy(data[0:4], []byte{0x01, 0x02, 0x03, 0x04})

	// Type 99 is unknown - should return nil and "unknown"
	value, typeName := p.parseValue(data, 99, 4, 0, byteOrder)

	if typeName != "unknown" {
		t.Errorf("parseValue() typeName = %q, want %q", typeName, "unknown")
	}

	if value != nil {
		t.Errorf("parseValue() value = %v, want nil for unknown type", value)
	}
}
