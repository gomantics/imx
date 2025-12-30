package tiff

import (
	"bytes"
	"encoding/binary"
	"testing"

	imxbin "github.com/gomantics/imx/internal/binary"
	"github.com/gomantics/imx/internal/parser"
)

// writeIFDEntry is defined in tiff_test.go

func TestParser_parseIFD(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name           string
		data           []byte
		wantDir        bool
		wantErr        bool
		wantNumEntries uint16
		wantSubIFDs    int
	}{
		{
			name: "valid IFD with 3 entries",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 2)
				order.PutUint16(b, 3)
				buf.Write(b)
				writeIFDEntry(buf, order, 0x0100, TypeLong, 1, 1920)        // ImageWidth
				writeIFDEntry(buf, order, 0x010F, TypeASCII, 4, 0x00636162) // Make
				writeIFDEntry(buf, order, 0x0101, TypeLong, 1, 1080)        // ImageHeight
				b4 := make([]byte, 4)
				order.PutUint32(b4, 0)
				buf.Write(b4)
				return buf.Bytes()
			}(),
			wantDir:        true,
			wantErr:        false,
			wantNumEntries: 3,
			wantSubIFDs:    0,
		},
		{
			name: "IFD with SubIFD pointers",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 2)
				order.PutUint16(b, 3)
				buf.Write(b)
				writeIFDEntry(buf, order, TagExifIFD, TypeLong, 1, 1000)
				writeIFDEntry(buf, order, TagGPSIFD, TypeLong, 1, 2000)
				writeIFDEntry(buf, order, TagInteropIFD, TypeLong, 1, 3000)
				b4 := make([]byte, 4)
				order.PutUint32(b4, 0)
				buf.Write(b4)
				return buf.Bytes()
			}(),
			wantDir:        true,
			wantErr:        false,
			wantNumEntries: 3,
			wantSubIFDs:    3,
		},
		{
			name:           "empty data - read error",
			data:           []byte{},
			wantDir:        false,
			wantErr:        true,
			wantNumEntries: 0,
			wantSubIFDs:    0,
		},
		{
			name: "truncated entries",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 2)
				order.PutUint16(b, 5) // Claims 5 entries but no data
				buf.Write(b)
				return buf.Bytes()
			}(),
			wantDir:        true,
			wantErr:        true,
			wantNumEntries: 5,
			wantSubIFDs:    0,
		},
		{
			name: "tag with unknown type",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 2)
				order.PutUint16(b, 1)
				buf.Write(b)
				writeIFDEntry(buf, order, 0x0100, TagType(0), 1, 100) // Unknown type
				b4 := make([]byte, 4)
				order.PutUint32(b4, 0)
				buf.Write(b4)
				return buf.Bytes()
			}(),
			wantDir:        true,
			wantErr:        true,
			wantNumEntries: 1,
			wantSubIFDs:    0,
		},
		{
			name: "with embedded metadata tags",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 2)
				order.PutUint16(b, 3)
				buf.Write(b)
				writeIFDEntry(buf, order, TagICCProfile, TypeUndefined, 4, 0)
				writeIFDEntry(buf, order, TagIPTC, TypeUndefined, 4, 0)
				writeIFDEntry(buf, order, TagXMP, TypeByte, 4, 0)
				b4 := make([]byte, 4)
				order.PutUint32(b4, 0)
				buf.Write(b4)
				return buf.Bytes()
			}(),
			wantDir:        true,
			wantErr:        true, // Errors because offsets point to invalid data
			wantNumEntries: 3,
			wantSubIFDs:    0,
		},
		{
			name: "with SubIFDs tag",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 2)
				order.PutUint16(b, 1)
				buf.Write(b)
				writeIFDEntry(buf, order, TagSubIFDs, TypeLong, 1, 1000)
				b4 := make([]byte, 4)
				order.PutUint32(b4, 0)
				buf.Write(b4)
				return buf.Bytes()
			}(),
			wantDir:        true,
			wantErr:        false,
			wantNumEntries: 1,
			wantSubIFDs:    1,
		},
		{
			name: "parseIFD with nil sharedParseErr",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 2)
				order.PutUint16(b, 1)
				buf.Write(b)
				writeIFDEntry(buf, order, 0x0100, TypeLong, 1, 1920) // ImageWidth
				b4 := make([]byte, 4)
				order.PutUint32(b4, 0)
				buf.Write(b4)
				return buf.Bytes()
			}(),
			wantDir:        true,
			wantErr:        false,
			wantNumEntries: 1,
			wantSubIFDs:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			var iccDirs, iptcDirs, xmpDirs []parser.Directory

			// Test both with shared error and without
			var parseErr *parser.ParseError
			if tt.name == "parseIFD with nil sharedParseErr" {
				// Test the nil path
				parseErr = nil
			} else {
				parseErr = parser.NewParseError()
			}
			dir, _, subIFDs, numEntries := p.parseIFD(reader, 0, "IFD0", &iccDirs, &iptcDirs, &xmpDirs, parseErr)

			if (dir != nil) != tt.wantDir {
				t.Errorf("dir = %v, wantDir %v", dir != nil, tt.wantDir)
			}

			// Check errors - parseIFD returns a parseErr which could be the one we passed or a new one
			if tt.name == "parseIFD with nil sharedParseErr" {
				// When we pass nil, parseIFD creates a new one internally
				// We get it back in the return value (second return), but we're testing that it creates one
				// For this specific test, we just verify it doesn't panic and returns expected results
			} else {
				hasErr := parseErr != nil && parseErr.OrNil() != nil
				if hasErr != tt.wantErr {
					t.Errorf("error = %v, wantErr %v", hasErr, tt.wantErr)
				}
			}

			if numEntries != tt.wantNumEntries {
				t.Errorf("numEntries = %d, want %d", numEntries, tt.wantNumEntries)
			}

			if len(subIFDs) != tt.wantSubIFDs {
				t.Errorf("subIFDs = %d, want %d", len(subIFDs), tt.wantSubIFDs)
			}
		})
	}
}

func TestParser_readIFDEntry(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		order      binary.ByteOrder
		wantTag    uint16
		wantType   TagType
		wantCount  uint32
		wantOffset uint32
		wantErr    bool
	}{
		{
			name: "valid entry little endian",
			data: func() []byte {
				buf := new(bytes.Buffer)
				order := binary.LittleEndian
				writeIFDEntry(buf, order, 0x0100, TypeLong, 1, 1920)
				return buf.Bytes()
			}(),
			order:      binary.LittleEndian,
			wantTag:    0x0100,
			wantType:   TypeLong,
			wantCount:  1,
			wantOffset: 1920,
		},
		{
			name: "valid entry big endian",
			data: func() []byte {
				buf := new(bytes.Buffer)
				order := binary.BigEndian
				writeIFDEntry(buf, order, 0x010F, TypeASCII, 5, 100)
				return buf.Bytes()
			}(),
			order:      binary.BigEndian,
			wantTag:    0x010F,
			wantType:   TypeASCII,
			wantCount:  5,
			wantOffset: 100,
		},
		{
			name:    "truncated - only tag",
			data:    []byte{0x00, 0x01},
			order:   binary.LittleEndian,
			wantErr: true,
		},
		{
			name:    "truncated - missing type",
			data:    []byte{0x00, 0x01, 0x00},
			order:   binary.LittleEndian,
			wantErr: true,
		},
		{
			name:    "truncated - missing count",
			data:    []byte{0x00, 0x01, 0x03, 0x00, 0x00},
			order:   binary.LittleEndian,
			wantErr: true,
		},
		{
			name:    "truncated - missing value",
			data:    []byte{0x00, 0x01, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00},
			order:   binary.LittleEndian,
			wantErr: true,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), tt.order)
			entry, err := p.readIFDEntry(reader, 0)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if entry.Tag != tt.wantTag {
				t.Errorf("Tag = 0x%04X, want 0x%04X", entry.Tag, tt.wantTag)
			}
			if entry.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", entry.Type, tt.wantType)
			}
			if entry.Count != tt.wantCount {
				t.Errorf("Count = %d, want %d", entry.Count, tt.wantCount)
			}
			if entry.ValueOffset != tt.wantOffset {
				t.Errorf("ValueOffset = %d, want %d", entry.ValueOffset, tt.wantOffset)
			}
		})
	}
}

func TestParser_parseTag(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name    string
		data    []byte
		entry   *IFDEntry
		dirName string
		wantVal interface{}
		wantErr bool
	}{
		{
			name: "GPS Version ID formatting",
			data: []byte{2, 2, 0, 0},
			entry: &IFDEntry{
				Tag:         0x0000,
				Type:        TypeByte,
				Count:       4,
				ValueOffset: 0x00000202,
			},
			dirName: "GPS",
			wantVal: "2.2.0.0",
		},
		{
			name: "regular tag",
			data: []byte{},
			entry: &IFDEntry{
				Tag:         0x0100,
				Type:        TypeLong,
				Count:       1,
				ValueOffset: 1920,
			},
			dirName: "IFD0",
			wantVal: uint32(1920),
		},
		{
			name: "unknown type returns error",
			data: []byte{},
			entry: &IFDEntry{
				Tag:         0x0100,
				Type:        TagType(0),
				Count:       1,
				ValueOffset: 0,
			},
			dirName: "IFD0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			tag, err := p.parseTag(reader, tt.entry, tt.dirName)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if tag == nil {
				t.Fatal("tag is nil")
			}
			if tag.Value != tt.wantVal {
				t.Errorf("Value = %v, want %v", tag.Value, tt.wantVal)
			}
		})
	}
}

func TestParser_readTagValue(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name    string
		data    []byte
		entry   *IFDEntry
		wantErr bool
	}{
		{"TypeByte", []byte{}, &IFDEntry{Type: TypeByte, Count: 1, ValueOffset: 0x42}, false},
		{"TypeASCII", []byte{}, &IFDEntry{Type: TypeASCII, Count: 3, ValueOffset: 0x00626100}, false},
		{"TypeShort", []byte{}, &IFDEntry{Type: TypeShort, Count: 1, ValueOffset: 100}, false},
		{"TypeLong", []byte{}, &IFDEntry{Type: TypeLong, Count: 1, ValueOffset: 12345}, false},
		{"TypeUndefined", []byte{}, &IFDEntry{Type: TypeUndefined, Count: 4, ValueOffset: 0x04030201}, false},
		{"TypeSByte", []byte{}, &IFDEntry{Type: TypeSByte, Count: 1, ValueOffset: 0xFF}, false},
		{"TypeSShort", []byte{}, &IFDEntry{Type: TypeSShort, Count: 1, ValueOffset: 0xFFFF}, false},
		{"TypeSLong", []byte{}, &IFDEntry{Type: TypeSLong, Count: 1, ValueOffset: 0xFFFFFFFF}, false},
		{"TypeRational", func() []byte {
			buf := new(bytes.Buffer)
			b := make([]byte, 4)
			order.PutUint32(b, 1)
			buf.Write(b)
			order.PutUint32(b, 100)
			buf.Write(b)
			return buf.Bytes()
		}(), &IFDEntry{Type: TypeRational, Count: 1, ValueOffset: 0}, false},
		{"TypeSRational", func() []byte {
			buf := new(bytes.Buffer)
			b := make([]byte, 4)
			order.PutUint32(b, 0xFFFFFFFF) // -1
			buf.Write(b)
			order.PutUint32(b, 3)
			buf.Write(b)
			return buf.Bytes()
		}(), &IFDEntry{Type: TypeSRational, Count: 1, ValueOffset: 0}, false},
		{"Unknown type 0", []byte{}, &IFDEntry{Type: TagType(0), Count: 1, ValueOffset: 0}, true},
		{"TypeFloat unsupported", []byte{0, 0, 0, 0}, &IFDEntry{Type: TypeFloat, Count: 1, ValueOffset: 0}, true},
		{"TypeDouble unsupported", []byte{0, 0, 0, 0, 0, 0, 0, 0}, &IFDEntry{Type: TypeDouble, Count: 1, ValueOffset: 0}, true},
		{
			name:    "Integer overflow protection - count * typeSize exceeds limit",
			data:    []byte{},
			entry:   &IFDEntry{Type: TypeRational, Count: 0x20000000, ValueOffset: 0}, // 536870912 * 8 = 4GB would overflow
			wantErr: true,
		},
		{
			name:    "Integer overflow protection - total size exceeds MaxTIFFTagDataSize",
			data:    []byte{},
			entry:   &IFDEntry{Type: TypeByte, Count: 100 * 1024 * 1024, ValueOffset: 0}, // 100MB > 50MB limit
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			_, err := p.readTagValue(reader, tt.entry)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParser_readBytes(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name       string
		data       []byte
		entry      *IFDEntry
		dataOffset int64
		wantValue  interface{}
		wantErr    bool
	}{
		{
			name:       "single byte inline",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeByte, Count: 1, ValueOffset: 0xAB},
			dataOffset: -1,
			wantValue:  byte(0xAB),
		},
		{
			name:       "multiple bytes inline",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeByte, Count: 4, ValueOffset: 0x04030201},
			dataOffset: -1,
			wantValue:  []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name:       "bytes from offset",
			data:       []byte{0x00, 0x00, 0x00, 0x00, 0xAA, 0xBB, 0xCC},
			entry:      &IFDEntry{Type: TypeByte, Count: 3, ValueOffset: 4},
			dataOffset: 4,
			wantValue:  []byte{0xAA, 0xBB, 0xCC},
		},
		{
			name:       "read error",
			data:       []byte{0x01},
			entry:      &IFDEntry{Type: TypeByte, Count: 100, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			got, err := p.readBytes(reader, tt.entry, tt.dataOffset)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			switch want := tt.wantValue.(type) {
			case byte:
				if got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			case []byte:
				gotBytes, _ := got.([]byte)
				if !bytes.Equal(gotBytes, want) {
					t.Errorf("got %v, want %v", gotBytes, want)
				}
			}
		})
	}
}

func TestParser_readASCII(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name       string
		data       []byte
		entry      *IFDEntry
		dataOffset int64
		wantValue  string
		wantErr    bool
	}{
		{
			name:       "inline ASCII",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeASCII, Count: 4, ValueOffset: 0x00636261},
			dataOffset: -1,
			wantValue:  "abc",
		},
		{
			name:       "ASCII from offset",
			data:       []byte{0x00, 0x00, 0x00, 0x00, 'H', 'e', 'l', 'l', 'o', 0x00},
			entry:      &IFDEntry{Type: TypeASCII, Count: 6, ValueOffset: 4},
			dataOffset: 4,
			wantValue:  "Hello",
		},
		{
			name:       "read error",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeASCII, Count: 100, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			got, err := p.readASCII(reader, tt.entry, tt.dataOffset)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("got %q, want %q", got, tt.wantValue)
			}
		})
	}
}

func TestParser_readShorts(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name       string
		data       []byte
		entry      *IFDEntry
		dataOffset int64
		wantValue  interface{}
		wantErr    bool
	}{
		{
			name:       "single short inline",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeShort, Count: 1, ValueOffset: 0x1234},
			dataOffset: -1,
			wantValue:  uint16(0x1234),
		},
		{
			name:       "two shorts inline",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeShort, Count: 2, ValueOffset: 0x56781234},
			dataOffset: -1,
			wantValue:  []uint16{0x1234, 0x5678},
		},
		{
			name: "shorts from offset",
			data: func() []byte {
				buf := new(bytes.Buffer)
				buf.Write([]byte{0x00, 0x00, 0x00, 0x00})
				b := make([]byte, 2)
				order.PutUint16(b, 100)
				buf.Write(b)
				order.PutUint16(b, 200)
				buf.Write(b)
				return buf.Bytes()
			}(),
			entry:      &IFDEntry{Type: TypeShort, Count: 2, ValueOffset: 4},
			dataOffset: 4,
			wantValue:  []uint16{100, 200},
		},
		{
			name:       "read error",
			data:       []byte{0x01},
			entry:      &IFDEntry{Type: TypeShort, Count: 10, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
		{
			name:       "zero count",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeShort, Count: 0, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []uint16{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			got, err := p.readShorts(reader, tt.entry, tt.dataOffset)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			switch want := tt.wantValue.(type) {
			case uint16:
				if got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			case []uint16:
				gotSlice := got.([]uint16)
				for i := range want {
					if gotSlice[i] != want[i] {
						t.Errorf("[%d] got %d, want %d", i, gotSlice[i], want[i])
					}
				}
			}
		})
	}
}

func TestParser_readLongs(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name       string
		data       []byte
		entry      *IFDEntry
		dataOffset int64
		wantValue  interface{}
		wantErr    bool
	}{
		{
			name:       "single long inline",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeLong, Count: 1, ValueOffset: 12345678},
			dataOffset: -1,
			wantValue:  uint32(12345678),
		},
		{
			name: "longs from offset",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 4)
				order.PutUint32(b, 1000)
				buf.Write(b)
				order.PutUint32(b, 2000)
				buf.Write(b)
				return buf.Bytes()
			}(),
			entry:      &IFDEntry{Type: TypeLong, Count: 2, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []uint32{1000, 2000},
		},
		{
			name:       "read error",
			data:       []byte{0x01},
			entry:      &IFDEntry{Type: TypeLong, Count: 10, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
		{
			name:       "zero count",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeLong, Count: 0, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []uint32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			got, err := p.readLongs(reader, tt.entry, tt.dataOffset)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			switch want := tt.wantValue.(type) {
			case uint32:
				if got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			case []uint32:
				gotSlice := got.([]uint32)
				for i := range want {
					if gotSlice[i] != want[i] {
						t.Errorf("[%d] got %d, want %d", i, gotSlice[i], want[i])
					}
				}
			}
		})
	}
}

func TestParser_readRationals(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name       string
		data       []byte
		entry      *IFDEntry
		dataOffset int64
		wantValue  interface{}
		wantErr    bool
	}{
		{
			name: "single rational",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 4)
				order.PutUint32(b, 1)
				buf.Write(b)
				order.PutUint32(b, 100)
				buf.Write(b)
				return buf.Bytes()
			}(),
			entry:      &IFDEntry{Type: TypeRational, Count: 1, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  "1/100",
		},
		{
			name: "multiple rationals",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 4)
				order.PutUint32(b, 72)
				buf.Write(b)
				order.PutUint32(b, 1)
				buf.Write(b)
				order.PutUint32(b, 300)
				buf.Write(b)
				order.PutUint32(b, 1)
				buf.Write(b)
				return buf.Bytes()
			}(),
			entry:      &IFDEntry{Type: TypeRational, Count: 2, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []string{"72/1", "300/1"},
		},
		{
			name:       "read numerator error",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeRational, Count: 1, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
		{
			name:       "read denominator error",
			data:       []byte{0x01, 0x00, 0x00, 0x00}, // Only numerator
			entry:      &IFDEntry{Type: TypeRational, Count: 1, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
		{
			name:       "zero count",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeRational, Count: 0, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			got, err := p.readRationals(reader, tt.entry, tt.dataOffset)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			switch want := tt.wantValue.(type) {
			case string:
				if got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			case []string:
				gotSlice := got.([]string)
				for i := range want {
					if gotSlice[i] != want[i] {
						t.Errorf("[%d] got %q, want %q", i, gotSlice[i], want[i])
					}
				}
			}
		})
	}
}

func TestParser_readSBytes(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name       string
		data       []byte
		entry      *IFDEntry
		dataOffset int64
		wantValue  interface{}
		wantErr    bool
	}{
		{
			name:       "single sbyte inline positive",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSByte, Count: 1, ValueOffset: 127},
			dataOffset: -1,
			wantValue:  int8(127),
		},
		{
			name:       "single sbyte inline negative",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSByte, Count: 1, ValueOffset: 0xFF},
			dataOffset: -1,
			wantValue:  int8(-1),
		},
		{
			name:       "sbytes from offset",
			data:       []byte{0x00, 0x00, 0x00, 0x00, 0x7F, 0xFF, 0x80},
			entry:      &IFDEntry{Type: TypeSByte, Count: 3, ValueOffset: 4},
			dataOffset: 4,
			wantValue:  []int8{127, -1, -128},
		},
		{
			name:       "read error",
			data:       []byte{0x01},
			entry:      &IFDEntry{Type: TypeSByte, Count: 100, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
		{
			name:       "zero count",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSByte, Count: 0, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []int8{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			got, err := p.readSBytes(reader, tt.entry, tt.dataOffset)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			switch want := tt.wantValue.(type) {
			case int8:
				if got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			case []int8:
				gotSlice := got.([]int8)
				for i := range want {
					if gotSlice[i] != want[i] {
						t.Errorf("[%d] got %d, want %d", i, gotSlice[i], want[i])
					}
				}
			}
		})
	}
}

func TestParser_readSShorts(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name       string
		data       []byte
		entry      *IFDEntry
		dataOffset int64
		wantValue  interface{}
		wantErr    bool
	}{
		{
			name:       "single sshort inline positive",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSShort, Count: 1, ValueOffset: 1000},
			dataOffset: -1,
			wantValue:  int16(1000),
		},
		{
			name:       "single sshort inline negative",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSShort, Count: 1, ValueOffset: 0xFFFF},
			dataOffset: -1,
			wantValue:  int16(-1),
		},
		{
			name:       "two sshorts inline",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSShort, Count: 2, ValueOffset: 0x0002FFFF},
			dataOffset: -1,
			wantValue:  []int16{-1, 2},
		},
		{
			name: "sshorts from offset",
			data: func() []byte {
				buf := new(bytes.Buffer)
				buf.Write([]byte{0x00, 0x00, 0x00, 0x00})
				b := make([]byte, 2)
				order.PutUint16(b, 0xFFFF) // -1
				buf.Write(b)
				order.PutUint16(b, 100)
				buf.Write(b)
				return buf.Bytes()
			}(),
			entry:      &IFDEntry{Type: TypeSShort, Count: 2, ValueOffset: 4},
			dataOffset: 4,
			wantValue:  []int16{-1, 100},
		},
		{
			name:       "read error",
			data:       []byte{0x01},
			entry:      &IFDEntry{Type: TypeSShort, Count: 10, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
		{
			name:       "zero count",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSShort, Count: 0, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []int16{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			got, err := p.readSShorts(reader, tt.entry, tt.dataOffset)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			switch want := tt.wantValue.(type) {
			case int16:
				if got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			case []int16:
				gotSlice := got.([]int16)
				for i := range want {
					if gotSlice[i] != want[i] {
						t.Errorf("[%d] got %d, want %d", i, gotSlice[i], want[i])
					}
				}
			}
		})
	}
}

func TestParser_readSLongs(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name       string
		data       []byte
		entry      *IFDEntry
		dataOffset int64
		wantValue  interface{}
		wantErr    bool
	}{
		{
			name:       "single slong inline positive",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSLong, Count: 1, ValueOffset: 12345},
			dataOffset: -1,
			wantValue:  int32(12345),
		},
		{
			name:       "single slong inline negative",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSLong, Count: 1, ValueOffset: 0xFFFFFFFF},
			dataOffset: -1,
			wantValue:  int32(-1),
		},
		{
			name: "slongs from offset",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 4)
				order.PutUint32(b, 0xFFFFFF9C) // -100
				buf.Write(b)
				order.PutUint32(b, 200)
				buf.Write(b)
				return buf.Bytes()
			}(),
			entry:      &IFDEntry{Type: TypeSLong, Count: 2, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []int32{-100, 200},
		},
		{
			name:       "read error",
			data:       []byte{0x01},
			entry:      &IFDEntry{Type: TypeSLong, Count: 10, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
		{
			name:       "zero count",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSLong, Count: 0, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []int32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			got, err := p.readSLongs(reader, tt.entry, tt.dataOffset)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			switch want := tt.wantValue.(type) {
			case int32:
				if got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			case []int32:
				gotSlice := got.([]int32)
				for i := range want {
					if gotSlice[i] != want[i] {
						t.Errorf("[%d] got %d, want %d", i, gotSlice[i], want[i])
					}
				}
			}
		})
	}
}

func TestParser_readSRationals(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name       string
		data       []byte
		entry      *IFDEntry
		dataOffset int64
		wantValue  interface{}
		wantErr    bool
	}{
		{
			name: "single srational",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 4)
				order.PutUint32(b, 0xFFFFFFFF) // -1
				buf.Write(b)
				order.PutUint32(b, 3)
				buf.Write(b)
				return buf.Bytes()
			}(),
			entry:      &IFDEntry{Type: TypeSRational, Count: 1, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  "-1/3",
		},
		{
			name: "multiple srationals",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 4)
				order.PutUint32(b, 0xFFFFFFFF) // -1
				buf.Write(b)
				order.PutUint32(b, 2)
				buf.Write(b)
				order.PutUint32(b, 10)
				buf.Write(b)
				order.PutUint32(b, 3)
				buf.Write(b)
				return buf.Bytes()
			}(),
			entry:      &IFDEntry{Type: TypeSRational, Count: 2, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []string{"-1/2", "10/3"},
		},
		{
			name:       "read numerator error",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSRational, Count: 1, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
		{
			name:       "read denominator error",
			data:       []byte{0x01, 0x00, 0x00, 0x00}, // Only numerator
			entry:      &IFDEntry{Type: TypeSRational, Count: 1, ValueOffset: 0},
			dataOffset: 0,
			wantErr:    true,
		},
		{
			name:       "zero count",
			data:       []byte{},
			entry:      &IFDEntry{Type: TypeSRational, Count: 0, ValueOffset: 0},
			dataOffset: 0,
			wantValue:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			got, err := p.readSRationals(reader, tt.entry, tt.dataOffset)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			switch want := tt.wantValue.(type) {
			case string:
				if got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			case []string:
				gotSlice := got.([]string)
				for i := range want {
					if gotSlice[i] != want[i] {
						t.Errorf("[%d] got %q, want %q", i, gotSlice[i], want[i])
					}
				}
			}
		})
	}
}

func TestParser_handleICCProfile(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	t.Run("reads data and calls ICC parser", func(t *testing.T) {
		// Provide enough data to read
		data := make([]byte, 136)
		entry := &IFDEntry{Tag: TagICCProfile, Type: TypeUndefined, Count: 132, ValueOffset: 4}

		reader := imxbin.NewReader(bytes.NewReader(data), order)
		parseErr := parser.NewParseError()
		var tags []parser.Tag

		// This will read data successfully, ICC parser may produce errors for invalid data
		var iccDirs []parser.Directory
		p.handleICCProfile(reader, entry, &tags, parseErr, &iccDirs)
		// No assertions - we're testing that it doesn't panic
	})

	t.Run("read error", func(t *testing.T) {
		data := []byte{0x00}
		entry := &IFDEntry{Tag: TagICCProfile, Type: TypeUndefined, Count: 1000, ValueOffset: 1000}

		reader := imxbin.NewReader(bytes.NewReader(data), order)
		parseErr := parser.NewParseError()
		var tags []parser.Tag

		var iccDirs []parser.Directory
		p.handleICCProfile(reader, entry, &tags, parseErr, &iccDirs)

		if parseErr.OrNil() == nil {
			t.Error("expected read error")
		}
	})
}

func TestParser_handleIPTC(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	t.Run("reads data and calls IPTC parser", func(t *testing.T) {
		data := append([]byte{0x00, 0x00, 0x00, 0x00}, []byte{0x1C, 0x02, 0x00, 0x00, 0x00}...)
		entry := &IFDEntry{Tag: TagIPTC, Type: TypeUndefined, Count: 5, ValueOffset: 4}

		reader := imxbin.NewReader(bytes.NewReader(data), order)
		parseErr := parser.NewParseError()

		var iptcDirs []parser.Directory
		p.handleIPTC(reader, entry, parseErr, &iptcDirs)
		// IPTC parser may produce errors for minimal/invalid data - that's OK
	})

	t.Run("malformed IPTC with invalid extended size", func(t *testing.T) {
		// IPTC data with invalid extended size length:
		// 0x1C = IPTC marker
		// 0x02 = record
		// 0x00 = dataset ID
		// 0x80, 0x00 = size with high bit set, extLen = 0 (invalid: must be 1-4)
		iptcData := []byte{0x1C, 0x02, 0x00, 0x80, 0x00}
		data := append([]byte{0x00, 0x00, 0x00, 0x00}, iptcData...)
		entry := &IFDEntry{Tag: TagIPTC, Type: TypeUndefined, Count: uint32(len(iptcData)), ValueOffset: 4}

		reader := imxbin.NewReader(bytes.NewReader(data), order)
		parseErr := parser.NewParseError()
		var iptcDirs []parser.Directory

		p.handleIPTC(reader, entry, parseErr, &iptcDirs)

		if parseErr.OrNil() == nil {
			t.Error("expected error for malformed IPTC data")
		}
	})

	t.Run("read error", func(t *testing.T) {
		data := []byte{0x00}
		entry := &IFDEntry{Tag: TagIPTC, Type: TypeUndefined, Count: 1000, ValueOffset: 1000}

		reader := imxbin.NewReader(bytes.NewReader(data), order)
		parseErr := parser.NewParseError()
		var iptcDirs []parser.Directory

		p.handleIPTC(reader, entry, parseErr, &iptcDirs)

		if parseErr.OrNil() == nil {
			t.Error("expected read error")
		}
	})
}

func TestParser_handleXMP(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	t.Run("reads data and calls XMP parser", func(t *testing.T) {
		xmpData := []byte("<?xml version=\"1.0\"?><x:xmpmeta></x:xmpmeta>\x00")
		data := append([]byte{0x00, 0x00, 0x00, 0x00}, xmpData...)
		entry := &IFDEntry{Tag: TagXMP, Type: TypeByte, Count: uint32(len(xmpData)), ValueOffset: 4}

		reader := imxbin.NewReader(bytes.NewReader(data), order)
		parseErr := parser.NewParseError()

		var xmpDirs []parser.Directory
		p.handleXMP(reader, entry, parseErr, &xmpDirs)
		// XMP parser may produce errors for minimal/invalid data - that's OK
	})

	t.Run("read error", func(t *testing.T) {
		data := []byte{0x00}
		entry := &IFDEntry{Tag: TagXMP, Type: TypeByte, Count: 1000, ValueOffset: 1000}

		reader := imxbin.NewReader(bytes.NewReader(data), order)
		parseErr := parser.NewParseError()
		var xmpDirs []parser.Directory

		p.handleXMP(reader, entry, parseErr, &xmpDirs)

		if parseErr.OrNil() == nil {
			t.Error("expected read error")
		}
	})
}

func TestParser_handleSubIFDs(t *testing.T) {
	p := New()
	order := binary.LittleEndian

	tests := []struct {
		name      string
		data      []byte
		entry     *IFDEntry
		wantCount int
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "single SubIFD inline",
			data:      []byte{},
			entry:     &IFDEntry{Tag: TagSubIFDs, Type: TypeLong, Count: 1, ValueOffset: 1000},
			wantCount: 1,
			wantNames: []string{"SubIFD"},
		},
		{
			name: "multiple SubIFDs from offset",
			data: func() []byte {
				buf := new(bytes.Buffer)
				b := make([]byte, 4)
				order.PutUint32(b, 2000)
				buf.Write(b)
				order.PutUint32(b, 3000)
				buf.Write(b)
				return buf.Bytes()
			}(),
			entry:     &IFDEntry{Tag: TagSubIFDs, Type: TypeLong, Count: 2, ValueOffset: 0},
			wantCount: 2,
			wantNames: []string{"SubIFD", "SubIFD1"},
		},
		{
			name:      "read error",
			data:      []byte{0x00},
			entry:     &IFDEntry{Tag: TagSubIFDs, Type: TypeLong, Count: 5, ValueOffset: 1000},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := imxbin.NewReader(bytes.NewReader(tt.data), order)
			parseErr := parser.NewParseError()
			var subIFDs []SubIFD

			p.handleSubIFDs(reader, tt.entry, &subIFDs, parseErr)

			if len(subIFDs) != tt.wantCount {
				t.Errorf("subIFDs = %d, want %d", len(subIFDs), tt.wantCount)
			}

			for i, wantName := range tt.wantNames {
				if i < len(subIFDs) && subIFDs[i].Name != wantName {
					t.Errorf("SubIFD[%d].Name = %q, want %q", i, subIFDs[i].Name, wantName)
				}
			}
		})
	}
}

func TestGetTagName(t *testing.T) {
	tests := []struct {
		tag     uint16
		dirName string
		want    string
	}{
		{0x0100, "IFD0", "ImageWidth"},
		{0x010F, "IFD1", "Make"},
		{0x829A, "ExifIFD", "ExposureTime"},
		{0x0002, "GPS", "GPSLatitude"},
		{0x0001, "Interoperability", "InteroperabilityIndex"},
		{0xFFFF, "IFD0", "0xFFFF"}, // Unknown falls back to hex
		{0x0100, "SubIFD", "ImageWidth"},
		{0x0100, "Unknown", "ImageWidth"}, // Default to TIFF tags
	}

	for _, tt := range tests {
		got := getTagName(tt.tag, tt.dirName)
		if got != tt.want {
			t.Errorf("getTagName(0x%04X, %q) = %q, want %q", tt.tag, tt.dirName, got, tt.want)
		}
	}
}

func TestGetTagNameForDir(t *testing.T) {
	tests := []struct {
		tag     uint16
		dirName string
		want    string
	}{
		{0x0100, "IFD0", "ImageWidth"},
		{0x0100, "ifd0", "ImageWidth"}, // Case insensitive
		{0x0100, "IFD1", "ImageWidth"},
		{0x010F, "TIFF", "Make"},
		{0x829A, "exififd", "ExposureTime"},
		{0x0002, "gps", "GPSLatitude"},
		{0x0001, "interoperability", "InteroperabilityIndex"},
		{0x0100, "subifd", "ImageWidth"},
		{0x0100, "subifd2", "ImageWidth"},
		{0x0100, "something", "ImageWidth"}, // Default to TIFF
		{0xFFFF, "IFD0", ""},                // Unknown returns empty
	}

	for _, tt := range tests {
		got := getTagNameForDir(tt.tag, tt.dirName)
		if got != tt.want {
			t.Errorf("getTagNameForDir(0x%04X, %q) = %q, want %q", tt.tag, tt.dirName, got, tt.want)
		}
	}
}
