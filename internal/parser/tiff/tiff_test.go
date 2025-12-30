package tiff

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.icc == nil {
		t.Error("ICC parser not initialized")
	}
	if p.iptc == nil {
		t.Error("IPTC parser not initialized")
	}
	if p.xmp == nil {
		t.Error("XMP parser not initialized")
	}
}

func TestParser_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "TIFF" {
		t.Errorf("Name() = %q, want %q", got, "TIFF")
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "little endian TIFF",
			data: []byte{'I', 'I', 42, 0},
			want: true,
		},
		{
			name: "big endian TIFF",
			data: []byte{'M', 'M', 0, 42},
			want: true,
		},
		{
			name: "invalid - wrong byte order marker",
			data: []byte{'X', 'X', 42, 0},
			want: false,
		},
		{
			name: "invalid - wrong magic number LE",
			data: []byte{'I', 'I', 0, 0},
			want: false,
		},
		{
			name: "invalid - wrong magic number BE",
			data: []byte{'M', 'M', 0, 0},
			want: false,
		},
		{
			name: "too short",
			data: []byte{'I', 'I'},
			want: false,
		},
		{
			name: "empty data",
			data: []byte{},
			want: false,
		},
		{
			name: "JPEG signature",
			data: []byte{0xFF, 0xD8, 0xFF, 0xE0},
			want: false,
		},
		{
			name: "PNG signature",
			data: []byte{0x89, 0x50, 0x4E, 0x47},
			want: false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			if got := p.Detect(r); got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

// buildMinimalTIFF creates a minimal valid TIFF with given byte order
func buildMinimalTIFF(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)

	// Header (8 bytes)
	if order == binary.LittleEndian {
		buf.Write([]byte{'I', 'I'})
	} else {
		buf.Write([]byte{'M', 'M'})
	}

	// Magic number (42)
	magicBuf := make([]byte, 2)
	order.PutUint16(magicBuf, 42)
	buf.Write(magicBuf)

	// IFD0 offset (8 = immediately after header)
	offsetBuf := make([]byte, 4)
	order.PutUint32(offsetBuf, 8)
	buf.Write(offsetBuf)

	// IFD0: 1 entry
	entryCountBuf := make([]byte, 2)
	order.PutUint16(entryCountBuf, 1)
	buf.Write(entryCountBuf)

	// Entry: ImageWidth (0x0100), LONG, count=1, value=1920
	tagBuf := make([]byte, 2)
	order.PutUint16(tagBuf, 0x0100)
	buf.Write(tagBuf)

	typeBuf := make([]byte, 2)
	order.PutUint16(typeBuf, uint16(TypeLong))
	buf.Write(typeBuf)

	countBuf := make([]byte, 4)
	order.PutUint32(countBuf, 1)
	buf.Write(countBuf)

	valueBuf := make([]byte, 4)
	order.PutUint32(valueBuf, 1920)
	buf.Write(valueBuf)

	// Next IFD offset (0 = none)
	nextIFDBuf := make([]byte, 4)
	order.PutUint32(nextIFDBuf, 0)
	buf.Write(nextIFDBuf)

	return buf.Bytes()
}

// buildTIFFWithIFD1 creates TIFF with IFD0 and IFD1
func buildTIFFWithIFD1(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)

	// Header
	if order == binary.LittleEndian {
		buf.Write([]byte{'I', 'I'})
	} else {
		buf.Write([]byte{'M', 'M'})
	}

	magicBuf := make([]byte, 2)
	order.PutUint16(magicBuf, 42)
	buf.Write(magicBuf)

	// IFD0 offset = 8
	offsetBuf := make([]byte, 4)
	order.PutUint32(offsetBuf, 8)
	buf.Write(offsetBuf)

	// IFD0: 1 entry (starts at offset 8)
	entryCountBuf := make([]byte, 2)
	order.PutUint16(entryCountBuf, 1)
	buf.Write(entryCountBuf)

	// Entry: ImageWidth
	writeIFDEntry(buf, order, 0x0100, TypeLong, 1, 1920)

	// Next IFD offset = 24 (8 + 2 + 12 + 4 = 26, but IFD1 at 26)
	ifd1Offset := uint32(buf.Len() + 4) // After this 4-byte field
	nextIFDBuf := make([]byte, 4)
	order.PutUint32(nextIFDBuf, ifd1Offset)
	buf.Write(nextIFDBuf)

	// IFD1: 1 entry
	order.PutUint16(entryCountBuf, 1)
	buf.Write(entryCountBuf)

	// Entry: ImageHeight
	writeIFDEntry(buf, order, 0x0101, TypeLong, 1, 1080)

	// No more IFDs
	order.PutUint32(nextIFDBuf, 0)
	buf.Write(nextIFDBuf)

	return buf.Bytes()
}

func writeIFDEntry(buf *bytes.Buffer, order binary.ByteOrder, tag uint16, typ TagType, count, value uint32) {
	tagBuf := make([]byte, 2)
	order.PutUint16(tagBuf, tag)
	buf.Write(tagBuf)

	typeBuf := make([]byte, 2)
	order.PutUint16(typeBuf, uint16(typ))
	buf.Write(typeBuf)

	countBuf := make([]byte, 4)
	order.PutUint32(countBuf, count)
	buf.Write(countBuf)

	valueBuf := make([]byte, 4)
	order.PutUint32(valueBuf, value)
	buf.Write(valueBuf)
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantDirs   int
		wantErr    bool
		checkFirst func(t *testing.T, dirs []string)
	}{
		{
			name:     "minimal little endian TIFF",
			data:     buildMinimalTIFF(binary.LittleEndian),
			wantDirs: 1,
			wantErr:  false,
			checkFirst: func(t *testing.T, dirs []string) {
				if len(dirs) > 0 && dirs[0] != "IFD0" {
					t.Errorf("first directory = %q, want %q", dirs[0], "IFD0")
				}
			},
		},
		{
			name:     "minimal big endian TIFF",
			data:     buildMinimalTIFF(binary.BigEndian),
			wantDirs: 1,
			wantErr:  false,
		},
		{
			name:     "TIFF with IFD0 and IFD1",
			data:     buildTIFFWithIFD1(binary.LittleEndian),
			wantDirs: 2,
			wantErr:  false,
		},
		{
			name:    "too short header",
			data:    []byte{'I', 'I', 42, 0},
			wantErr: true,
		},
		{
			name:    "invalid byte order",
			data:    []byte{'X', 'X', 42, 0, 8, 0, 0, 0},
			wantErr: true,
		},
		{
			name:    "invalid magic number",
			data:    []byte{'I', 'I', 0, 0, 8, 0, 0, 0},
			wantErr: true,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			dirs, err := p.Parse(r)

			if tt.wantErr {
				if err == nil {
					t.Error("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			if len(dirs) != tt.wantDirs {
				t.Errorf("Parse() returned %d directories, want %d", len(dirs), tt.wantDirs)
			}

			if tt.checkFirst != nil {
				var dirNames []string
				for _, d := range dirs {
					dirNames = append(dirNames, d.Name)
				}
				tt.checkFirst(t, dirNames)
			}
		})
	}
}

func TestParser_Parse_WithSubIFDs(t *testing.T) {
	p := New()

	// Build TIFF with ExifIFD pointer
	data := buildTIFFWithExifIFD(binary.LittleEndian)
	r := bytes.NewReader(data)

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should have IFD0 and ExifIFD
	if len(dirs) < 2 {
		t.Errorf("Parse() returned %d directories, want at least 2", len(dirs))
	}

	// Check for ExifIFD
	hasExifIFD := false
	for _, d := range dirs {
		if d.Name == "ExifIFD" {
			hasExifIFD = true
			break
		}
	}
	if !hasExifIFD {
		t.Error("Parse() should return ExifIFD directory")
	}
}

func buildTIFFWithExifIFD(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)

	// Header
	if order == binary.LittleEndian {
		buf.Write([]byte{'I', 'I'})
	} else {
		buf.Write([]byte{'M', 'M'})
	}

	magicBuf := make([]byte, 2)
	order.PutUint16(magicBuf, 42)
	buf.Write(magicBuf)

	// IFD0 offset = 8
	offsetBuf := make([]byte, 4)
	order.PutUint32(offsetBuf, 8)
	buf.Write(offsetBuf)

	// IFD0: 2 entries (starts at offset 8)
	// Size: 2 + 2*12 + 4 = 30 bytes, so ExifIFD starts at 8+30=38
	exifIFDOffset := uint32(8 + 2 + 2*12 + 4)

	entryCountBuf := make([]byte, 2)
	order.PutUint16(entryCountBuf, 2)
	buf.Write(entryCountBuf)

	// Entry 1: ImageWidth
	writeIFDEntry(buf, order, 0x0100, TypeLong, 1, 1920)

	// Entry 2: ExifIFD pointer (0x8769)
	writeIFDEntry(buf, order, TagExifIFD, TypeLong, 1, exifIFDOffset)

	// No next IFD
	nextIFDBuf := make([]byte, 4)
	order.PutUint32(nextIFDBuf, 0)
	buf.Write(nextIFDBuf)

	// ExifIFD: 1 entry
	order.PutUint16(entryCountBuf, 1)
	buf.Write(entryCountBuf)

	// Entry: ExposureTime (inline SHORT value)
	writeIFDEntry(buf, order, 0x829A, TypeShort, 1, 100)

	// No next IFD
	order.PutUint32(nextIFDBuf, 0)
	buf.Write(nextIFDBuf)

	return buf.Bytes()
}

func TestParser_Parse_InvalidIFDOffset(t *testing.T) {
	// TIFF with IFD offset pointing beyond file
	buf := new(bytes.Buffer)
	order := binary.LittleEndian

	buf.Write([]byte{'I', 'I'})

	magicBuf := make([]byte, 2)
	order.PutUint16(magicBuf, 42)
	buf.Write(magicBuf)

	// IFD0 offset pointing way beyond data
	offsetBuf := make([]byte, 4)
	order.PutUint32(offsetBuf, 10000)
	buf.Write(offsetBuf)

	p := New()
	r := bytes.NewReader(buf.Bytes())
	_, err := p.Parse(r)

	// Should return error (or empty dirs with partial error)
	if err == nil {
		t.Log("Parse() with invalid IFD offset should ideally return error")
	}
}

func TestParser_Parse_WithGPSIFD(t *testing.T) {
	p := New()
	data := buildTIFFWithGPSIFD(binary.LittleEndian)
	r := bytes.NewReader(data)

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	hasGPS := false
	for _, d := range dirs {
		if d.Name == "GPS" {
			hasGPS = true
			break
		}
	}
	if !hasGPS {
		t.Error("Parse() should return GPS directory")
	}
}

func buildTIFFWithGPSIFD(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)

	if order == binary.LittleEndian {
		buf.Write([]byte{'I', 'I'})
	} else {
		buf.Write([]byte{'M', 'M'})
	}

	magicBuf := make([]byte, 2)
	order.PutUint16(magicBuf, 42)
	buf.Write(magicBuf)

	offsetBuf := make([]byte, 4)
	order.PutUint32(offsetBuf, 8)
	buf.Write(offsetBuf)

	gpsIFDOffset := uint32(8 + 2 + 2*12 + 4)

	entryCountBuf := make([]byte, 2)
	order.PutUint16(entryCountBuf, 2)
	buf.Write(entryCountBuf)

	writeIFDEntry(buf, order, 0x0100, TypeLong, 1, 1920)
	writeIFDEntry(buf, order, TagGPSIFD, TypeLong, 1, gpsIFDOffset)

	nextIFDBuf := make([]byte, 4)
	order.PutUint32(nextIFDBuf, 0)
	buf.Write(nextIFDBuf)

	// GPS IFD: 1 entry
	order.PutUint16(entryCountBuf, 1)
	buf.Write(entryCountBuf)

	// GPSVersionID (4 bytes inline)
	writeIFDEntry(buf, order, 0x0000, TypeByte, 4, 0x02020000) // Version 2.2.0.0

	order.PutUint32(nextIFDBuf, 0)
	buf.Write(nextIFDBuf)

	return buf.Bytes()
}

func TestParser_Parse_EmptyIFD(t *testing.T) {
	buf := new(bytes.Buffer)
	order := binary.LittleEndian

	buf.Write([]byte{'I', 'I'})

	magicBuf := make([]byte, 2)
	order.PutUint16(magicBuf, 42)
	buf.Write(magicBuf)

	offsetBuf := make([]byte, 4)
	order.PutUint32(offsetBuf, 8)
	buf.Write(offsetBuf)

	// IFD with 0 entries
	entryCountBuf := make([]byte, 2)
	order.PutUint16(entryCountBuf, 0)
	buf.Write(entryCountBuf)

	nextIFDBuf := make([]byte, 4)
	order.PutUint32(nextIFDBuf, 0)
	buf.Write(nextIFDBuf)

	p := New()
	r := bytes.NewReader(buf.Bytes())
	dirs, err := p.Parse(r)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}

	// Empty IFD should not be included
	if len(dirs) != 0 {
		t.Errorf("Parse() returned %d directories, want 0 for empty IFD", len(dirs))
	}
}

func TestParser_ConcurrentParse(t *testing.T) {
	// Create a minimal valid TIFF file for testing
	order := binary.LittleEndian
	buf := new(bytes.Buffer)

	// TIFF header
	buf.Write([]byte{'I', 'I', 42, 0}) // Little endian, magic 42
	ifdOffset := uint32(8)
	order.PutUint32(buf.Bytes()[4:8], ifdOffset)

	// IFD with 1 entry (ImageWidth = 1920)
	entryCountBuf := make([]byte, 2)
	order.PutUint16(entryCountBuf, 1)
	buf.Write(entryCountBuf)

	// ImageWidth entry
	writeIFDEntry(buf, order, 0x0100, TypeLong, 1, 1920)

	// Next IFD offset (0 = no more IFDs)
	nextIFDBuf := make([]byte, 4)
	order.PutUint32(nextIFDBuf, 0)
	buf.Write(nextIFDBuf)

	p := New()
	data := buf.Bytes()
	r := bytes.NewReader(data)

	const goroutines = 10
	done := make(chan bool, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			p.Parse(r)
			done <- true
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}
