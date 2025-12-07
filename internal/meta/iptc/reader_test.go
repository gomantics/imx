package iptc

import (
	"io"
	"testing"
)

func TestNewDatasetReader(t *testing.T) {
	data := []byte{1, 2, 3}
	r := NewDatasetReader(data)

	if r == nil {
		t.Fatal("NewDatasetReader() returned nil")
	}
	if r.offset != 0 {
		t.Errorf("NewDatasetReader() offset = %d, want 0", r.offset)
	}
	if len(r.data) != 3 {
		t.Errorf("NewDatasetReader() data length = %d, want 3", len(r.data))
	}
}

func TestDatasetReader_EOF(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		want   bool
	}{
		{"empty data", []byte{}, 0, true},
		{"at start", []byte{1, 2, 3}, 0, false},
		{"in middle", []byte{1, 2, 3}, 1, false},
		{"at end", []byte{1, 2, 3}, 3, true},
		{"past end", []byte{1, 2, 3}, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DatasetReader{data: tt.data, offset: tt.offset}
			if got := r.EOF(); got != tt.want {
				t.Errorf("EOF() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatasetReader_Skip(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		skip   int
		want   int
	}{
		{"skip normal", []byte{1, 2, 3, 4, 5}, 2, 2},
		{"skip to end", []byte{1, 2, 3}, 3, 3},
		{"skip past end", []byte{1, 2, 3}, 10, 3}, // Should clamp to len
		{"skip zero", []byte{1, 2, 3}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDatasetReader(tt.data)
			r.Skip(tt.skip)
			if r.offset != tt.want {
				t.Errorf("Skip(%d) offset = %d, want %d", tt.skip, r.offset, tt.want)
			}
		})
	}
}

func TestDatasetReader_readByte(t *testing.T) {
	r := NewDatasetReader([]byte{0x1C, 0x02, 0x05})

	// Read first byte
	b, err := r.readByte()
	if err != nil {
		t.Fatalf("readByte() error = %v", err)
	}
	if b != 0x1C {
		t.Errorf("readByte() = 0x%02X, want 0x1C", b)
	}
	if r.offset != 1 {
		t.Errorf("offset = %d, want 1", r.offset)
	}

	// Read second byte
	b, err = r.readByte()
	if err != nil {
		t.Fatalf("readByte() error = %v", err)
	}
	if b != 0x02 {
		t.Errorf("readByte() = 0x%02X, want 0x02", b)
	}

	// Read third byte
	b, err = r.readByte()
	if err != nil {
		t.Fatalf("readByte() error = %v", err)
	}
	if b != 0x05 {
		t.Errorf("readByte() = 0x%02X, want 0x05", b)
	}

	// Read past end
	_, err = r.readByte()
	if err != io.EOF {
		t.Errorf("readByte() at EOF error = %v, want io.EOF", err)
	}
}

func TestDatasetReader_readBytes(t *testing.T) {
	r := NewDatasetReader([]byte{1, 2, 3, 4, 5})

	// Read 2 bytes
	bytes, err := r.readBytes(2)
	if err != nil {
		t.Fatalf("readBytes(2) error = %v", err)
	}
	if len(bytes) != 2 || bytes[0] != 1 || bytes[1] != 2 {
		t.Errorf("readBytes(2) = %v, want [1 2]", bytes)
	}
	if r.offset != 2 {
		t.Errorf("offset = %d, want 2", r.offset)
	}

	// Read 3 more bytes (to end)
	bytes, err = r.readBytes(3)
	if err != nil {
		t.Fatalf("readBytes(3) error = %v", err)
	}
	if len(bytes) != 3 || bytes[0] != 3 || bytes[2] != 5 {
		t.Errorf("readBytes(3) = %v, want [3 4 5]", bytes)
	}

	// Read past end
	_, err = r.readBytes(1)
	if err != io.EOF {
		t.Errorf("readBytes() past EOF error = %v, want io.EOF", err)
	}
}

func TestDatasetReader_readBytes_Partial(t *testing.T) {
	r := NewDatasetReader([]byte{1, 2, 3})
	r.offset = 2 // Position at byte 3

	// Try to read 5 bytes but only 1 available
	_, err := r.readBytes(5)
	if err != io.EOF {
		t.Errorf("readBytes(5) with 1 available error = %v, want io.EOF", err)
	}
}

func TestDatasetReader_expectMarker(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{"valid marker", []byte{0x1C, 0x02}, false, ""},
		{"invalid marker", []byte{0x1D, 0x02}, true, "invalid marker"},
		{"EOF", []byte{}, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDatasetReader(tt.data)
			err := r.expectMarker()

			if tt.wantErr {
				if err == nil {
					t.Error("expectMarker() error = nil, want error")
				} else if tt.errMsg != "" && err.Error()[:14] != tt.errMsg {
					t.Errorf("expectMarker() error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("expectMarker() error = %v, want nil", err)
				}
				if r.offset != 1 {
					t.Errorf("offset = %d, want 1", r.offset)
				}
			}
		})
	}
}

func TestDatasetReader_readSize_Standard(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{"size 0", []byte{0x00, 0x00}, 0},
		{"size 5", []byte{0x00, 0x05}, 5},
		{"size 255", []byte{0x00, 0xFF}, 255},
		{"size 256", []byte{0x01, 0x00}, 256},
		{"size 32767", []byte{0x7F, 0xFF}, 32767},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDatasetReader(tt.data)
			size, err := r.readSize()
			if err != nil {
				t.Fatalf("readSize() error = %v", err)
			}
			if size != tt.want {
				t.Errorf("readSize() = %d, want %d", size, tt.want)
			}
			if r.offset != 2 {
				t.Errorf("offset = %d, want 2", r.offset)
			}
		})
	}
}

func TestDatasetReader_readSize_Extended(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{
			name: "extended 1 byte (size 100)",
			data: []byte{0x80, 0x01, 0x64}, // Extended flag + 1 byte follows, size = 100
			want: 100,
		},
		{
			name: "extended 2 bytes (size 300)",
			data: []byte{0x80, 0x02, 0x01, 0x2C}, // Extended flag + 2 bytes, size = 300
			want: 300,
		},
		{
			name: "extended 4 bytes (size 70000)",
			data: []byte{0x80, 0x04, 0x00, 0x01, 0x11, 0x70}, // Extended flag + 4 bytes
			want: 70000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDatasetReader(tt.data)
			size, err := r.readSize()
			if err != nil {
				t.Fatalf("readSize() error = %v", err)
			}
			if size != tt.want {
				t.Errorf("readSize() = %d, want %d", size, tt.want)
			}
		})
	}
}

func TestDatasetReader_readSize_ExtendedInvalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"extLen too large (5)", []byte{0x80, 0x05, 0, 0, 0, 0, 0}},
		{"extLen truncated", []byte{0x80, 0x04, 0, 0}}, // Says 4 bytes but only 2 available
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDatasetReader(tt.data)
			_, err := r.readSize()
			if err == nil {
				t.Error("readSize() error = nil, want error for invalid extended size")
			}
		})
	}
}

func TestDatasetReader_readSize_TooShort(t *testing.T) {
	r := NewDatasetReader([]byte{0x00}) // Only 1 byte
	_, err := r.readSize()
	if err != io.EOF {
		t.Errorf("readSize() with 1 byte error = %v, want io.EOF", err)
	}
}

func TestDatasetReader_ReadNext_Simple(t *testing.T) {
	// Build a simple dataset: ObjectName = "Test"
	data := buildIPTCDataset(RecordApplication, 5, []byte("Test"))

	r := NewDatasetReader(data)
	ds, err := r.ReadNext()

	if err != nil {
		t.Fatalf("ReadNext() error = %v", err)
	}
	if ds.Record != RecordApplication {
		t.Errorf("Record = %v, want %v", ds.Record, RecordApplication)
	}
	if ds.DatasetID != 5 {
		t.Errorf("DatasetID = %d, want 5", ds.DatasetID)
	}
	if ds.Name != "ObjectName" {
		t.Errorf("Name = %q, want ObjectName", ds.Name)
	}
	if ds.Value != "Test" {
		t.Errorf("Value = %v, want Test", ds.Value)
	}
	if string(ds.Raw) != "Test" {
		t.Errorf("Raw = %v, want Test", ds.Raw)
	}
}

func TestDatasetReader_ReadNext_Multiple(t *testing.T) {
	// Build multiple datasets
	data := buildIPTCDataset(RecordApplication, 5, []byte("Title"))
	data = append(data, buildIPTCDataset(RecordApplication, 80, []byte("Author"))...)

	r := NewDatasetReader(data)

	// Read first
	ds1, err := r.ReadNext()
	if err != nil {
		t.Fatalf("ReadNext() first error = %v", err)
	}
	if ds1.Value != "Title" {
		t.Errorf("First dataset Value = %v, want Title", ds1.Value)
	}

	// Read second
	ds2, err := r.ReadNext()
	if err != nil {
		t.Fatalf("ReadNext() second error = %v", err)
	}
	if ds2.Value != "Author" {
		t.Errorf("Second dataset Value = %v, want Author", ds2.Value)
	}

	// Read past end
	_, err = r.ReadNext()
	if err != io.EOF {
		t.Errorf("ReadNext() past end error = %v, want io.EOF", err)
	}
}

func TestDatasetReader_ReadNext_SkipGarbage(t *testing.T) {
	// Some garbage bytes followed by valid dataset
	data := []byte{0x00, 0xFF, 0x12}
	data = append(data, buildIPTCDataset(RecordApplication, 5, []byte("Test"))...)

	r := NewDatasetReader(data)
	ds, err := r.ReadNext()

	if err != nil {
		t.Fatalf("ReadNext() error = %v", err)
	}
	if ds.Value != "Test" {
		t.Errorf("Value = %v, want Test", ds.Value)
	}
}

func TestDatasetReader_ReadNext_UnknownDataset(t *testing.T) {
	// Unknown dataset ID
	data := buildIPTCDataset(RecordApplication, 255, []byte("Unknown"))

	r := NewDatasetReader(data)
	ds, err := r.ReadNext()

	if err != nil {
		t.Fatalf("ReadNext() error = %v", err)
	}
	if ds.Name != "Dataset2:255" {
		t.Errorf("Name = %q, want Dataset2:255", ds.Name)
	}
}

func TestDatasetReader_ReadNext_ExtendedSize(t *testing.T) {
	// Build dataset with extended size
	data := []byte{
		iptcTagMarker,
		byte(RecordApplication),
		5,          // ObjectName
		0x80, 0x02, // Extended size: 2 bytes follow
		0x00, 0x05, // Size = 5
		'H', 'e', 'l', 'l', 'o',
	}

	r := NewDatasetReader(data)
	ds, err := r.ReadNext()

	if err != nil {
		t.Fatalf("ReadNext() error = %v", err)
	}
	if ds.Value != "Hello" {
		t.Errorf("Value = %v, want Hello", ds.Value)
	}
}

func TestDatasetReader_ReadNext_EOF(t *testing.T) {
	r := NewDatasetReader([]byte{})
	_, err := r.ReadNext()
	if err != io.EOF {
		t.Errorf("ReadNext() on empty data error = %v, want io.EOF", err)
	}
}

func TestDatasetReader_ReadNext_TruncatedRecord(t *testing.T) {
	// Marker but no record byte
	data := []byte{iptcTagMarker}

	r := NewDatasetReader(data)
	_, err := r.ReadNext()

	if err == nil {
		t.Error("ReadNext() error = nil, want error for truncated record")
	}
}

func TestDatasetReader_ReadNext_TruncatedDatasetID(t *testing.T) {
	// Marker + record but no dataset ID
	data := []byte{iptcTagMarker, byte(RecordApplication)}

	r := NewDatasetReader(data)
	_, err := r.ReadNext()

	if err == nil {
		t.Error("ReadNext() error = nil, want error for truncated dataset ID")
	}
}

func TestDatasetReader_ReadNext_TruncatedSize(t *testing.T) {
	// Marker + record + dataset ID but incomplete size
	data := []byte{iptcTagMarker, byte(RecordApplication), 5, 0x00}

	r := NewDatasetReader(data)
	_, err := r.ReadNext()

	if err == nil {
		t.Error("ReadNext() error = nil, want error for truncated size")
	}
}

func TestDatasetReader_ReadNext_TruncatedValue(t *testing.T) {
	// Valid header but value truncated
	data := []byte{
		iptcTagMarker,
		byte(RecordApplication),
		5,     // ObjectName
		0, 10, // Size = 10
		'T', 'e', 's', 't', // Only 4 bytes instead of 10
	}

	r := NewDatasetReader(data)
	_, err := r.ReadNext()

	if err == nil {
		t.Error("ReadNext() error = nil, want error for truncated value")
	}
}

func TestDatasetReader_ReadNext_OnlyGarbage(t *testing.T) {
	// Only garbage, no valid markers
	data := []byte{0x00, 0xFF, 0x12, 0x34}

	r := NewDatasetReader(data)
	_, err := r.ReadNext()

	if err != io.EOF {
		t.Errorf("ReadNext() on garbage-only data error = %v, want io.EOF", err)
	}
}
