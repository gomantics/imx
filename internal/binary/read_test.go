package binary

import (
	"io"
	"testing"
)

// mockReaderAt is a simple in-memory io.ReaderAt implementation for testing.
type mockReaderAt struct {
	data []byte
}

func (m *mockReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || off >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n := copy(p, m.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func newMockReader(data []byte) *mockReaderAt {
	return &mockReaderAt{data: data}
}

func TestReadUint16BE(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		offset  int64
		want    uint16
		wantErr bool
	}{
		{"read 0x1234", []byte{0x12, 0x34, 0x56, 0x78}, 0, 0x1234, false},
		{"read at offset 2", []byte{0x00, 0x00, 0xAB, 0xCD}, 2, 0xABCD, false},
		{"read 0xFFFF", []byte{0xFF, 0xFF}, 0, 0xFFFF, false},
		{"read 0x0000", []byte{0x00, 0x00}, 0, 0x0000, false},
		{"offset beyond data", []byte{0x12, 0x34}, 10, 0, true},
		{"insufficient data", []byte{0x12}, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newMockReader(tt.data)
			got, err := ReadUint16BE(r, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint16BE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint16BE() = 0x%04X, want 0x%04X", got, tt.want)
			}
		})
	}
}

func TestReadUint16LE(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		offset  int64
		want    uint16
		wantErr bool
	}{
		{"read 0x3412 (bytes 0x12,0x34)", []byte{0x12, 0x34, 0x56, 0x78}, 0, 0x3412, false},
		{"read at offset 2", []byte{0x00, 0x00, 0xAB, 0xCD}, 2, 0xCDAB, false},
		{"read 0xFFFF", []byte{0xFF, 0xFF}, 0, 0xFFFF, false},
		{"offset beyond data", []byte{0x12, 0x34}, 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newMockReader(tt.data)
			got, err := ReadUint16LE(r, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint16LE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint16LE() = 0x%04X, want 0x%04X", got, tt.want)
			}
		})
	}
}

func TestReadUint32BE(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		offset  int64
		want    uint32
		wantErr bool
	}{
		{"read 0x12345678", []byte{0x12, 0x34, 0x56, 0x78}, 0, 0x12345678, false},
		{"read at offset 4", []byte{0x00, 0x00, 0x00, 0x00, 0xAB, 0xCD, 0xEF, 0x01}, 4, 0xABCDEF01, false},
		{"read 0xFFFFFFFF", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFF, false},
		{"read 0x00000000", []byte{0x00, 0x00, 0x00, 0x00}, 0, 0x00000000, false},
		{"offset beyond data", []byte{0x12, 0x34, 0x56, 0x78}, 10, 0, true},
		{"insufficient data", []byte{0x12, 0x34, 0x56}, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newMockReader(tt.data)
			got, err := ReadUint32BE(r, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint32BE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint32BE() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestReadUint32LE(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		offset  int64
		want    uint32
		wantErr bool
	}{
		{"read 0x78563412", []byte{0x12, 0x34, 0x56, 0x78}, 0, 0x78563412, false},
		{"read at offset 4", []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0xEF, 0xCD, 0xAB}, 4, 0xABCDEF01, false},
		{"read 0xFFFFFFFF", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFF, false},
		{"offset beyond data", []byte{0x12, 0x34, 0x56, 0x78}, 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newMockReader(tt.data)
			got, err := ReadUint32LE(r, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint32LE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint32LE() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestReadUint64BE(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		offset  int64
		want    uint64
		wantErr bool
	}{
		{"read 0x123456789ABCDEF0", []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}, 0, 0x123456789ABCDEF0, false},
		{"read at offset 8", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFE, 0xDC, 0xBA, 0x98, 0x76, 0x54, 0x32, 0x10}, 8, 0xFEDCBA9876543210, false},
		{"read 0xFFFFFFFFFFFFFFFF", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFFFFFFFFFF, false},
		{"read 0x0000000000000000", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0, 0x0000000000000000, false},
		{"offset beyond data", []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}, 10, 0, true},
		{"insufficient data", []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE}, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newMockReader(tt.data)
			got, err := ReadUint64BE(r, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint64BE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint64BE() = 0x%016X, want 0x%016X", got, tt.want)
			}
		})
	}
}

func TestReadUint64LE(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		offset  int64
		want    uint64
		wantErr bool
	}{
		{"read 0xF0DEBC9A78563412", []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}, 0, 0xF0DEBC9A78563412, false},
		{"read at offset 8", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x32, 0x54, 0x76, 0x98, 0xBA, 0xDC, 0xFE}, 8, 0xFEDCBA9876543210, false},
		{"read 0xFFFFFFFFFFFFFFFF", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFFFFFFFFFF, false},
		{"offset beyond data", []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}, 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newMockReader(tt.data)
			got, err := ReadUint64LE(r, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint64LE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint64LE() = 0x%016X, want 0x%016X", got, tt.want)
			}
		})
	}
}
