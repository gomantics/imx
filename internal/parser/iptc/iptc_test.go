package iptc

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
}

func TestParser_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "IPTC" {
		t.Errorf("Name() = %q, want %q", got, "IPTC")
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"valid 8BIM signature", []byte("8BIM"), true},
		{"invalid signature", []byte("JPEG"), false},
		{"too short", []byte("8BI"), false},
		{"empty", []byte{}, false},
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

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantDirs bool
		wantErr  bool
	}{
		{
			name:     "no IPTC data",
			data:     []byte{0x00, 0x00, 0x00, 0x00},
			wantDirs: false,
			wantErr:  false,
		},
		{
			name: "simple IPTC dataset",
			data: []byte{
				0x1C, 0x02, 0x50, // Marker, Record=Application, DatasetID=0x50 (Byline)
				0x00, 0x04, // Size = 4
				't', 'e', 's', 't', // Data
			},
			wantDirs: true,
			wantErr:  false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs, parseErr := p.Parse(bytes.NewReader(tt.data))
			if (parseErr != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", parseErr, tt.wantErr)
			}
			if (len(dirs) > 0) != tt.wantDirs {
				t.Errorf("Parse() dirs present = %v, want %v", len(dirs) > 0, tt.wantDirs)
			}
		})
	}
}

func TestParser_findIPTCResource(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantOffset bool
		wantSize   int64
		wantErr    bool
	}{
		{
			name: "valid IPTC resource",
			data: func() []byte {
				data := make([]byte, 20)
				copy(data[0:4], "8BIM")                             // Signature
				binary.BigEndian.PutUint16(data[4:6], ResourceIPTC) // Resource ID
				data[6] = 0                                         // Name length = 0
				// namePadded = 0, but (0+1)%2 != 0, so namePadded becomes 1
				binary.BigEndian.PutUint32(data[8:12], 4) // Data size at offset 7 + 1
				return data
			}(),
			wantOffset: true,
			wantSize:   4,
			wantErr:    false,
		},
		{
			name: "8BIM with name padding odd",
			data: func() []byte {
				data := make([]byte, 30)
				copy(data[0:4], "8BIM")                             // Signature
				binary.BigEndian.PutUint16(data[4:6], ResourceIPTC) // Resource ID
				data[6] = 3                                         // Name length = 3
				copy(data[7:10], "foo")                             // Name
				// namePadded: (3+1)%2 = 0, so namePadded = 3
				binary.BigEndian.PutUint32(data[10:14], 8) // Data size at offset 7 + 3
				return data
			}(),
			wantOffset: true,
			wantSize:   8,
			wantErr:    false,
		},
		{
			name: "8BIM with name padding even",
			data: func() []byte {
				data := make([]byte, 30)
				copy(data[0:4], "8BIM")                             // Signature
				binary.BigEndian.PutUint16(data[4:6], ResourceIPTC) // Resource ID
				data[6] = 4                                         // Name length = 4
				copy(data[7:11], "test")                            // Name
				// namePadded: (4+1)%2 = 1 != 0, so namePadded = 5
				binary.BigEndian.PutUint32(data[12:16], 8) // Data size at offset 7 + 5
				return data
			}(),
			wantOffset: true,
			wantSize:   8,
			wantErr:    false,
		},
		{
			name: "skip wrong resource ID",
			data: func() []byte {
				data := make([]byte, 40)
				// First 8BIM with wrong ID
				copy(data[0:4], "8BIM")
				binary.BigEndian.PutUint16(data[4:6], 0x0400) // Wrong ID
				data[6] = 0
				binary.BigEndian.PutUint32(data[8:12], 4) // Size at 8
				// Skip to next: 8 + 4 + 4 = 16 (size is even, no padding)
				// Second 8BIM at offset 16
				copy(data[16:20], "8BIM")
				binary.BigEndian.PutUint16(data[20:22], ResourceIPTC)
				data[22] = 0
				binary.BigEndian.PutUint32(data[24:28], 6)
				return data
			}(),
			wantOffset: true,
			wantSize:   6,
			wantErr:    false,
		},
		{
			name: "odd size padding",
			data: func() []byte {
				data := make([]byte, 50)
				// First 8BIM with odd size
				copy(data[0:4], "8BIM")
				binary.BigEndian.PutUint16(data[4:6], 0x0400)
				data[6] = 0
				binary.BigEndian.PutUint32(data[8:12], 5) // Odd size = 5
				// Skip: 8 + 4 + 5 + 1(padding) = 18
				// Second 8BIM at offset 18
				copy(data[18:22], "8BIM")
				binary.BigEndian.PutUint16(data[22:24], ResourceIPTC)
				data[24] = 0
				binary.BigEndian.PutUint32(data[26:30], 4)
				return data
			}(),
			wantOffset: true,
			wantSize:   4,
			wantErr:    false,
		},
		{
			name:       "no IPTC resource",
			data:       []byte{0x00, 0x00, 0x00},
			wantOffset: false,
			wantSize:   0,
			wantErr:    false,
		},
		{
			name: "read error on size",
			data: func() []byte {
				data := make([]byte, 7)
				copy(data[0:4], "8BIM")
				binary.BigEndian.PutUint16(data[4:6], ResourceIPTC)
				data[6] = 0
				// Size should be at offset 8, but data ends at 7
				return data
			}(),
			wantOffset: false,
			wantSize:   0,
			wantErr:    false,
		},
		{
			name:       "no 8BIM found",
			data:       []byte("JFIF\x00\x01\x00test data"),
			wantOffset: false,
			wantSize:   0,
			wantErr:    false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, size, err := p.findIPTCResource(bytes.NewReader(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("findIPTCResource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (offset != 0) != tt.wantOffset {
				t.Errorf("findIPTCResource() offset = %d, wantOffset %v", offset, tt.wantOffset)
			}
			if size != tt.wantSize {
				t.Errorf("findIPTCResource() size = %d, want %d", size, tt.wantSize)
			}
		})
	}
}

func TestParser_readByte(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantB   byte
		wantPos int64
		wantErr bool
	}{
		{"valid read", []byte{0x42, 0x43}, 0x42, 1, false},
		{"empty data", []byte{}, 0, 0, true},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			pos := int64(0)
			b, err := p.readByte(r, &pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("readByte() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if b != tt.wantB {
					t.Errorf("readByte() byte = 0x%02X, want 0x%02X", b, tt.wantB)
				}
				if pos != tt.wantPos {
					t.Errorf("readByte() pos = %d, want %d", pos, tt.wantPos)
				}
			}
		})
	}
}

func TestParser_readSize(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantSize int
		wantPos  int64
		wantErr  bool
	}{
		{
			name:     "simple size",
			data:     []byte{0x00, 0x10},
			wantSize: 16,
			wantPos:  2,
			wantErr:  false,
		},
		{
			name:     "extended size 1 byte",
			data:     []byte{0x80, 0x01, 0xFF},
			wantSize: 255,
			wantPos:  3,
			wantErr:  false,
		},
		{
			name:     "extended size 2 bytes",
			data:     []byte{0x80, 0x02, 0x01, 0x00},
			wantSize: 256,
			wantPos:  4,
			wantErr:  false,
		},
		{
			name:     "extended size 4 bytes",
			data:     []byte{0x80, 0x04, 0x00, 0x01, 0x00, 0x00},
			wantSize: 65536,
			wantPos:  6,
			wantErr:  false,
		},
		{
			name:     "invalid extended size length 0",
			data:     []byte{0x80, 0x00},
			wantSize: 0,
			wantPos:  2,
			wantErr:  true,
		},
		{
			name:     "invalid extended size length > 4",
			data:     []byte{0x80, 0x05},
			wantSize: 0,
			wantPos:  2,
			wantErr:  true,
		},
		{
			name:     "extended size read error",
			data:     []byte{0x80, 0x02},
			wantSize: 0,
			wantPos:  2,
			wantErr:  true,
		},
		{
			name:     "read error first 2 bytes",
			data:     []byte{0x80},
			wantSize: 0,
			wantPos:  0,
			wantErr:  true,
		},
		{
			name:     "extended size overflow protection - exceeds limit",
			data:     []byte{0x80, 0x04, 0x01, 0x00, 0x00, 0x00}, // 16MB > 10MB limit
			wantSize: 0,
			wantPos:  6,
			wantErr:  true,
		},
		{
			name:     "extended size at limit boundary",
			data:     []byte{0x80, 0x04, 0x00, 0xA0, 0x00, 0x00}, // Exactly 10MB
			wantSize: 10485760,
			wantPos:  6,
			wantErr:  false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			pos := int64(0)
			size, err := p.readSize(r, &pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("readSize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if size != tt.wantSize {
					t.Errorf("readSize() size = %d, want %d", size, tt.wantSize)
				}
				if pos != tt.wantPos {
					t.Errorf("readSize() pos = %d, want %d", pos, tt.wantPos)
				}
			}
		})
	}
}

func TestParser_parseIPTCIIM(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantCount int
		wantErr   bool
	}{
		{
			name: "multiple datasets",
			data: []byte{
				0x1C, 0x02, 0x50, 0x00, 0x03, 'f', 'o', 'o',
				0x1C, 0x02, 0x78, 0x00, 0x03, 'b', 'a', 'r',
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "empty data",
			data:      []byte{},
			wantCount: 0,
			wantErr:   false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datasets, err := p.parseIPTCIIM(bytes.NewReader(tt.data), 0, int64(len(tt.data)))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIPTCIIM() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(datasets) != tt.wantCount {
				t.Errorf("parseIPTCIIM() count = %d, want %d", len(datasets), tt.wantCount)
			}
		})
	}
}

func TestParser_readDataset(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantCount int
		wantErr   bool
	}{
		{
			name:      "valid dataset",
			data:      []byte{0x1C, 0x02, 0x50, 0x00, 0x03, 'f', 'o', 'o'},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "invalid marker",
			data:      []byte{0x00, 0x02, 0x50, 0x00, 0x03, 'f', 'o', 'o'},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "zero size dataset",
			data:      []byte{0x1C, 0x02, 0x50, 0x00, 0x00},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "EOF at marker",
			data:      []byte{},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "read error on record",
			data:      []byte{0x1C},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "read error on datasetID",
			data:      []byte{0x1C, 0x02},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "read error on size",
			data:      []byte{0x1C, 0x02, 0x50},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "read error on data",
			data:      []byte{0x1C, 0x02, 0x50, 0x00, 0x10},
			wantCount: 0,
			wantErr:   true,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			pos := int64(0)
			end := int64(len(tt.data))
			if tt.name == "read error on data" {
				end = 100
			}
			var datasets []Dataset

			err := p.readDataset(r, &pos, end, &datasets)
			if (err != nil) != tt.wantErr {
				t.Errorf("readDataset() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(datasets) != tt.wantCount {
				t.Errorf("readDataset() count = %d, want %d", len(datasets), tt.wantCount)
			}
		})
	}
}

func TestParser_buildDirectories(t *testing.T) {
	tests := []struct {
		name     string
		datasets []Dataset
		wantDirs int
		wantTags int
	}{
		{
			name: "multiple datasets same record",
			datasets: []Dataset{
				{Record: RecordApplication, DatasetID: 25, Name: "Keywords", Value: "foo", Raw: []byte("foo")},
				{Record: RecordApplication, DatasetID: 25, Name: "Keywords", Value: "bar", Raw: []byte("bar")},
				{Record: RecordApplication, DatasetID: 80, Name: "Byline", Value: "test", Raw: []byte("test")},
			},
			wantDirs: 1,
			wantTags: 2,
		},
		{
			name: "multiple records",
			datasets: []Dataset{
				{Record: RecordEnvelope, DatasetID: 0, Name: "RecordVersion", Value: 4, Raw: []byte{0, 4}},
				{Record: RecordApplication, DatasetID: 80, Name: "Byline", Value: "test", Raw: []byte("test")},
			},
			wantDirs: 2,
			wantTags: 0, // Not checking tags for this test
		},
		{
			name:     "empty datasets",
			datasets: []Dataset{},
			wantDirs: 0,
			wantTags: 0,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs := p.buildDirectories(tt.datasets)
			if len(dirs) != tt.wantDirs {
				t.Errorf("buildDirectories() dirs = %d, want %d", len(dirs), tt.wantDirs)
			}
			if tt.wantTags > 0 && len(dirs) > 0 && len(dirs[0].Tags) != tt.wantTags {
				t.Errorf("buildDirectories() tags = %d, want %d", len(dirs[0].Tags), tt.wantTags)
			}
		})
	}
}

func TestRecord_String(t *testing.T) {
	tests := []struct {
		record Record
		want   string
	}{
		{RecordEnvelope, "Envelope"},
		{RecordApplication, "Application"},
		{RecordNewsPhoto, "NewsPhoto"},
		{RecordPreObjectData, "PreObjectData"},
		{RecordObjectData, "ObjectData"},
		{RecordPostObjectData, "PostObjectData"},
		{Record(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.record.String(); got != tt.want {
				t.Errorf("Record.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Additional tests for 100% coverage - covering specific error paths

// errorReader is a custom io.ReaderAt that returns errors for testing
type errorReader struct {
	data      []byte
	errorAt   int64 // Position at which to return an error
	errorType string
}

func (e *errorReader) ReadAt(p []byte, off int64) (int, error) {
	if e.errorType == "specific" && off >= e.errorAt {
		return 0, bytes.ErrTooLarge
	}
	if off >= int64(len(e.data)) {
		return 0, io.EOF
	}
	n := copy(p, e.data[off:])
	if n < len(p) {
		return n, io.ErrUnexpectedEOF
	}
	return n, nil
}

func TestParser_Parse_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		reader  io.ReaderAt
		wantErr bool
	}{
		{
			name: "findIPTCResource returns error",
			reader: &errorReader{
				data:      []byte{0xFF}, // Minimal data to trigger immediate error
				errorAt:   0,
				errorType: "specific",
			},
			wantErr: true,
		},
		{
			name: "parseIPTCIIM error on non-EOF error",
			reader: &errorReader{
				data: []byte{
					0x1C, 0x02, 0x50, 0x00, 0x04, 't', 'e', 's', 't', // Valid dataset (9 bytes)
					0x1C, // Start of next dataset at offset 9
				},
				errorAt:   10, // Trigger error when reading next dataset
				errorType: "specific",
			},
			wantErr: true, // parseIPTCIIM now returns errors
		},
		{
			name: "valid 8BIM with IPTC data",
			reader: func() io.ReaderAt {
				data := make([]byte, 50)
				copy(data[0:4], "8BIM")
				binary.BigEndian.PutUint16(data[4:6], ResourceIPTC)
				data[6] = 0
				binary.BigEndian.PutUint32(data[8:12], 10)
				copy(data[12:], []byte{
					0x1C, 0x02, 0x50, 0x00, 0x04, 't', 'e', 's', 't',
				})
				return bytes.NewReader(data)
			}(),
			wantErr: false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs, parseErr := p.Parse(tt.reader)
			hasErr := parseErr != nil && parseErr.Error() != ""
			if hasErr != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", parseErr, tt.wantErr)
			}
			if !tt.wantErr && len(dirs) == 0 {
				// Should have parsed some directories for valid data
				if tt.name == "valid 8BIM with IPTC data" {
					t.Errorf("Parse() returned no directories for valid data")
				}
			}
		})
	}
}

func TestParser_findIPTCResource_HeaderReadError(t *testing.T) {
	tests := []struct {
		name       string
		reader     io.ReaderAt
		wantOffset int64
		wantSize   int64
		wantErr    bool
	}{
		{
			name: "error reading header at offset 0",
			reader: &errorReader{
				data:      make([]byte, 0),
				errorAt:   0,
				errorType: "specific",
			},
			wantOffset: 0,
			wantSize:   0,
			wantErr:    true, // Now returns error
		},
		{
			name: "error reading header after finding 8BIM",
			reader: &errorReader{
				data: func() []byte {
					data := make([]byte, 20)
					// First 8BIM that's incomplete
					copy(data[0:4], "JUNK")
					// At offset 4, start a valid 8BIM but cause error
					copy(data[4:8], "8BIM")
					return data
				}(),
				errorAt:   11, // Error when trying to read full header at offset 4
				errorType: "specific",
			},
			wantOffset: 0,
			wantSize:   0,
			wantErr:    true, // Now returns error
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, size, err := p.findIPTCResource(tt.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("findIPTCResource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if offset != tt.wantOffset {
					t.Errorf("findIPTCResource() offset = %d, want %d", offset, tt.wantOffset)
				}
				if size != tt.wantSize {
					t.Errorf("findIPTCResource() size = %d, want %d", size, tt.wantSize)
				}
			}
		})
	}
}

func TestParser_parseIPTCIIM_NonEOFError(t *testing.T) {
	tests := []struct {
		name      string
		reader    io.ReaderAt
		offset    int64
		maxSize   int64
		wantCount int
		wantErr   bool
	}{
		{
			name: "readDataset with errorReader triggers non-EOF error",
			reader: &errorReader{
				data: []byte{
					0x1C, 0x02, 0x50, 0x00, 0x04, 't', 'e', 's', 't', // Valid dataset (9 bytes)
					0x1C, // Start of next dataset at offset 9
				},
				errorAt:   10, // Trigger error when trying to read next dataset
				errorType: "specific",
			},
			offset:    0,
			maxSize:   100,
			wantCount: 1,
			wantErr:   true, // Returns custom error (not EOF/ErrUnexpectedEOF)
		},
		{
			name:      "empty data returns no error",
			reader:    bytes.NewReader([]byte{}),
			offset:    0,
			maxSize:   0,
			wantCount: 0,
			wantErr:   false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datasets, err := p.parseIPTCIIM(tt.reader, tt.offset, tt.maxSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIPTCIIM() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(datasets) < tt.wantCount {
				t.Errorf("parseIPTCIIM() count = %d, want at least %d", len(datasets), tt.wantCount)
			}
		})
	}
}

func TestParser_ConcurrentParse(t *testing.T) {
	// Create minimal valid IPTC data
	data := make([]byte, 28)
	// 8BIM header
	copy(data[0:4], "8BIM")
	binary.BigEndian.PutUint16(data[4:6], 1028) // Tag (IPTC-NAA)
	binary.BigEndian.PutUint16(data[6:8], 0)    // Name length
	// Size (18 bytes)
	binary.BigEndian.PutUint32(data[8:12], 18)
	// IPTC data: Tag marker (0x1C), Record number (2), Dataset number (0), size (0x80)
	data[12] = 0x1C
	data[13] = 2
	data[14] = 0
	data[15] = 0x80
	data[16] = 0x00 // Extended length high byte
	data[17] = 0x0A // Extended length low byte (10 bytes)
	// 10 bytes of data
	copy(data[18:28], "TestData\x00\x00")

	p := New()
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
