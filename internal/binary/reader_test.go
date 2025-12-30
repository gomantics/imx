package binary

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestReader_ReadUint16(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78}

	tests := []struct {
		name    string
		order   binary.ByteOrder
		offset  int64
		want    uint16
		wantErr bool
	}{
		{"big endian at 0", binary.BigEndian, 0, 0x1234, false},
		{"big endian at 2", binary.BigEndian, 2, 0x5678, false},
		{"little endian at 0", binary.LittleEndian, 0, 0x3412, false},
		{"little endian at 2", binary.LittleEndian, 2, 0x7856, false},
		{"offset beyond data", binary.BigEndian, 100, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(newMockReader(data), tt.order)
			got, err := r.ReadUint16(tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint16() = 0x%04X, want 0x%04X", got, tt.want)
			}
		})
	}
}

func TestReader_ReadUint32(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}

	tests := []struct {
		name    string
		order   binary.ByteOrder
		offset  int64
		want    uint32
		wantErr bool
	}{
		{"big endian at 0", binary.BigEndian, 0, 0x12345678, false},
		{"big endian at 4", binary.BigEndian, 4, 0x9ABCDEF0, false},
		{"little endian at 0", binary.LittleEndian, 0, 0x78563412, false},
		{"little endian at 4", binary.LittleEndian, 4, 0xF0DEBC9A, false},
		{"offset beyond data", binary.BigEndian, 100, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(newMockReader(data), tt.order)
			got, err := r.ReadUint32(tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint32() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestReader_ReadUint64(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}

	tests := []struct {
		name    string
		order   binary.ByteOrder
		offset  int64
		want    uint64
		wantErr bool
	}{
		{"big endian", binary.BigEndian, 0, 0x123456789ABCDEF0, false},
		{"little endian", binary.LittleEndian, 0, 0xF0DEBC9A78563412, false},
		{"offset beyond data", binary.BigEndian, 100, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(newMockReader(data), tt.order)
			got, err := r.ReadUint64(tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint64() = 0x%016X, want 0x%016X", got, tt.want)
			}
		})
	}
}

func TestReader_ReadInt16(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		order   binary.ByteOrder
		offset  int64
		want    int16
		wantErr bool
	}{
		{"positive big endian", []byte{0x12, 0x34}, binary.BigEndian, 0, 0x1234, false},
		{"negative big endian", []byte{0xFF, 0xFE}, binary.BigEndian, 0, -2, false},
		{"positive little endian", []byte{0x34, 0x12}, binary.LittleEndian, 0, 0x1234, false},
		{"offset beyond data", []byte{0x12, 0x34}, binary.BigEndian, 100, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(newMockReader(tt.data), tt.order)
			got, err := r.ReadInt16(tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadInt16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadInt16() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestReader_ReadInt32(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		order   binary.ByteOrder
		offset  int64
		want    int32
		wantErr bool
	}{
		{"positive big endian", []byte{0x12, 0x34, 0x56, 0x78}, binary.BigEndian, 0, 0x12345678, false},
		{"negative big endian", []byte{0xFF, 0xFF, 0xFF, 0xFE}, binary.BigEndian, 0, -2, false},
		{"positive little endian", []byte{0x78, 0x56, 0x34, 0x12}, binary.LittleEndian, 0, 0x12345678, false},
		{"offset beyond data", []byte{0x12, 0x34, 0x56, 0x78}, binary.BigEndian, 100, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(newMockReader(tt.data), tt.order)
			got, err := r.ReadInt32(tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadInt32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadInt32() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestReader_ReadBytes(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}

	tests := []struct {
		name    string
		offset  int64
		n       int
		want    []byte
		wantErr bool
	}{
		{"read 4 bytes at 0", 0, 4, []byte{0x12, 0x34, 0x56, 0x78}, false},
		{"read 2 bytes at 2", 2, 2, []byte{0x56, 0x78}, false},
		{"read all bytes", 0, 6, []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}, false},
		{"offset beyond data", 100, 4, nil, true},
		{"read beyond available", 4, 4, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(newMockReader(data), binary.BigEndian)
			got, err := r.ReadBytes(tt.offset, tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("ReadBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
