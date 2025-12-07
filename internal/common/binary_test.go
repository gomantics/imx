package common

import (
	"encoding/binary"
	"testing"
)

func TestReadUint16(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		order  binary.ByteOrder
		want   uint16
		isErr  bool
	}{
		{
			name:   "big endian at offset 0",
			data:   []byte{0x12, 0x34, 0x56, 0x78},
			offset: 0,
			order:  binary.BigEndian,
			want:   0x1234,
		},
		{
			name:   "little endian at offset 0",
			data:   []byte{0x12, 0x34, 0x56, 0x78},
			offset: 0,
			order:  binary.LittleEndian,
			want:   0x3412,
		},
		{
			name:   "big endian at offset 2",
			data:   []byte{0x12, 0x34, 0x56, 0x78},
			offset: 2,
			order:  binary.BigEndian,
			want:   0x5678,
		},
		{
			name:   "out of bounds at offset 3",
			data:   []byte{0x12, 0x34, 0x56, 0x78},
			offset: 3,
			order:  binary.BigEndian,
			isErr:  true,
		},
		{
			name:   "out of bounds at offset 4",
			data:   []byte{0x12, 0x34, 0x56, 0x78},
			offset: 4,
			order:  binary.BigEndian,
			isErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadUint16(tt.data, tt.offset, tt.order)
			if tt.isErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got 0x%04X, want 0x%04X", got, tt.want)
			}
		})
	}
}

func TestReadUint32(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		order  binary.ByteOrder
		want   uint32
		isErr  bool
	}{
		{
			name:   "big endian at offset 0",
			data:   []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC},
			offset: 0,
			order:  binary.BigEndian,
			want:   0x12345678,
		},
		{
			name:   "little endian at offset 0",
			data:   []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC},
			offset: 0,
			order:  binary.LittleEndian,
			want:   0x78563412,
		},
		{
			name:   "big endian at offset 2",
			data:   []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC},
			offset: 2,
			order:  binary.BigEndian,
			want:   0x56789ABC,
		},
		{
			name:   "out of bounds at offset 3",
			data:   []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC},
			offset: 3,
			order:  binary.BigEndian,
			isErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadUint32(tt.data, tt.offset, tt.order)
			if tt.isErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestReadUint64(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		order  binary.ByteOrder
		want   uint64
		isErr  bool
	}{
		{
			name:   "big endian at offset 0",
			data:   []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0, 0x11, 0x22},
			offset: 0,
			order:  binary.BigEndian,
			want:   0x123456789ABCDEF0,
		},
		{
			name:   "little endian at offset 0",
			data:   []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0, 0x11, 0x22},
			offset: 0,
			order:  binary.LittleEndian,
			want:   0xF0DEBC9A78563412,
		},
		{
			name:   "big endian at offset 2",
			data:   []byte{0x00, 0x00, 0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0},
			offset: 2,
			order:  binary.BigEndian,
			want:   0x123456789ABCDEF0,
		},
		{
			name:   "out of bounds at offset 3",
			data:   []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0, 0x11, 0x22},
			offset: 3,
			order:  binary.BigEndian,
			isErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadUint64(tt.data, tt.offset, tt.order)
			if tt.isErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got 0x%016X, want 0x%016X", got, tt.want)
			}
		})
	}
}

func TestSafeSlice(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		size   int
		want   []byte
		isErr  bool
	}{
		{
			name:   "valid slice from middle",
			data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			offset: 1,
			size:   3,
			want:   []byte{0x01, 0x02, 0x03},
		},
		{
			name:   "zero size",
			data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			offset: 0,
			size:   0,
			want:   []byte{},
		},
		{
			name:   "full slice",
			data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			offset: 0,
			size:   6,
			want:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
		},
		{
			name:   "out of bounds size",
			data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			offset: 0,
			size:   7,
			isErr:  true,
		},
		{
			name:   "out of bounds offset+size",
			data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			offset: 5,
			size:   2,
			isErr:  true,
		},
		{
			name:   "negative offset",
			data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			offset: -1,
			size:   3,
			isErr:  true,
		},
		{
			name:   "negative size",
			data:   []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			offset: 0,
			size:   -1,
			isErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeSlice(tt.data, tt.offset, tt.size)
			if tt.isErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Errorf("got len=%d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %v, want %v", got, tt.want)
					break
				}
			}
		})
	}
}

func TestParseS15Fixed16(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		want  float64
		isErr bool
	}{
		{
			name: "positive value 1.5",
			data: []byte{0x00, 0x01, 0x80, 0x00}, // 1.5 in s15.16
			want: 1.5,
		},
		{
			name: "zero",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			want: 0.0,
		},
		{
			name: "negative value -1.5",
			data: []byte{0xFF, 0xFE, 0x80, 0x00}, // -1.5 in s15.16
			want: -1.5,
		},
		{
			name:  "insufficient data",
			data:  []byte{0x00, 0x01},
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseS15Fixed16(tt.data)
			if tt.isErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

func TestParseU16Fixed16(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		want  float64
		isErr bool
	}{
		{
			name: "value 1.5",
			data: []byte{0x00, 0x01, 0x80, 0x00}, // 1.5 in u16.16
			want: 1.5,
		},
		{
			name: "zero",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			want: 0.0,
		},
		{
			name: "value 256.0",
			data: []byte{0x01, 0x00, 0x00, 0x00}, // 256.0 in u16.16
			want: 256.0,
		},
		{
			name:  "insufficient data",
			data:  []byte{0x00, 0x01},
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseU16Fixed16(tt.data)
			if tt.isErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

func TestParseU8Fixed8(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		want  float64
		isErr bool
	}{
		{
			name: "value 1.5",
			data: []byte{0x01, 0x80}, // 1.5 in u8.8
			want: 1.5,
		},
		{
			name: "zero",
			data: []byte{0x00, 0x00},
			want: 0.0,
		},
		{
			name: "value 10.25",
			data: []byte{0x0A, 0x40}, // 10.25 in u8.8
			want: 10.25,
		},
		{
			name:  "insufficient data",
			data:  []byte{0x00},
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseU8Fixed8(tt.data)
			if tt.isErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

func TestTrimNullBytes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no nulls",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "null at end",
			input: "hello\x00",
			want:  "hello",
		},
		{
			name:  "null in middle",
			input: "hello\x00world",
			want:  "hello",
		},
		{
			name:  "multiple nulls",
			input: "hello\x00\x00\x00",
			want:  "hello",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only nulls",
			input: "\x00\x00\x00",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimNullBytes(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrimNullBytesFromSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "no nulls",
			input: []byte("hello"),
			want:  "hello",
		},
		{
			name:  "null at end",
			input: []byte{'h', 'e', 'l', 'l', 'o', 0x00},
			want:  "hello",
		},
		{
			name:  "null in middle",
			input: []byte{'h', 'e', 'l', 'l', 'o', 0x00, 'w', 'o', 'r', 'l', 'd'},
			want:  "hello",
		},
		{
			name:  "empty slice",
			input: []byte{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimNullBytesFromSlice(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
