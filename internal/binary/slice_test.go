package binary

import (
	"encoding/binary"
	"testing"
)

func TestUint16BE(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		want   uint16
	}{
		{"offset 0", []byte{0x12, 0x34, 0x56, 0x78}, 0, 0x1234},
		{"offset 1", []byte{0x00, 0x12, 0x34, 0x56}, 1, 0x1234},
		{"offset 2", []byte{0x00, 0x00, 0xAB, 0xCD}, 2, 0xABCD},
		{"max value", []byte{0xFF, 0xFF}, 0, 0xFFFF},
		{"zero", []byte{0x00, 0x00}, 0, 0x0000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Uint16BE(tt.data, tt.offset)
			if got != tt.want {
				t.Errorf("Uint16BE() = 0x%04X, want 0x%04X", got, tt.want)
			}
		})
	}
}

func TestUint16LE(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		want   uint16
	}{
		{"offset 0", []byte{0x34, 0x12, 0x78, 0x56}, 0, 0x1234},
		{"offset 1", []byte{0x00, 0x34, 0x12, 0x56}, 1, 0x1234},
		{"offset 2", []byte{0x00, 0x00, 0xCD, 0xAB}, 2, 0xABCD},
		{"max value", []byte{0xFF, 0xFF}, 0, 0xFFFF},
		{"zero", []byte{0x00, 0x00}, 0, 0x0000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Uint16LE(tt.data, tt.offset)
			if got != tt.want {
				t.Errorf("Uint16LE() = 0x%04X, want 0x%04X", got, tt.want)
			}
		})
	}
}

func TestUint32BE(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		want   uint32
	}{
		{"offset 0", []byte{0x12, 0x34, 0x56, 0x78}, 0, 0x12345678},
		{"offset 1", []byte{0x00, 0x12, 0x34, 0x56, 0x78}, 1, 0x12345678},
		{"offset 2", []byte{0x00, 0x00, 0xAB, 0xCD, 0xEF, 0x01}, 2, 0xABCDEF01},
		{"max value", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFF},
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0, 0x00000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Uint32BE(tt.data, tt.offset)
			if got != tt.want {
				t.Errorf("Uint32BE() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestUint32LE(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		want   uint32
	}{
		{"offset 0", []byte{0x78, 0x56, 0x34, 0x12}, 0, 0x12345678},
		{"offset 1", []byte{0x00, 0x78, 0x56, 0x34, 0x12}, 1, 0x12345678},
		{"offset 2", []byte{0x00, 0x00, 0x01, 0xEF, 0xCD, 0xAB}, 2, 0xABCDEF01},
		{"max value", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFF},
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0, 0x00000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Uint32LE(tt.data, tt.offset)
			if got != tt.want {
				t.Errorf("Uint32LE() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestUint64BE(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		want   uint64
	}{
		{"offset 0", []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}, 0, 0x123456789ABCDEF0},
		{"offset 1", []byte{0x00, 0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}, 1, 0x123456789ABCDEF0},
		{"max value", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFFFFFFFFFF},
		{"zero", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0, 0x0000000000000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Uint64BE(tt.data, tt.offset)
			if got != tt.want {
				t.Errorf("Uint64BE() = 0x%016X, want 0x%016X", got, tt.want)
			}
		})
	}
}

func TestUint64LE(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		offset int
		want   uint64
	}{
		{"offset 0", []byte{0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12}, 0, 0x123456789ABCDEF0},
		{"offset 1", []byte{0x00, 0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12}, 1, 0x123456789ABCDEF0},
		{"max value", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFFFFFFFFFF},
		{"zero", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0, 0x0000000000000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Uint64LE(tt.data, tt.offset)
			if got != tt.want {
				t.Errorf("Uint64LE() = 0x%016X, want 0x%016X", got, tt.want)
			}
		})
	}
}

func TestPutUint16BE(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		value  uint16
		want   []byte
	}{
		{"offset 0", 0, 0x1234, []byte{0x12, 0x34, 0x00, 0x00}},
		{"offset 1", 1, 0x1234, []byte{0x00, 0x12, 0x34, 0x00}},
		{"offset 2", 2, 0xABCD, []byte{0x00, 0x00, 0xAB, 0xCD}},
		{"max value", 0, 0xFFFF, []byte{0xFF, 0xFF}},
		{"zero", 0, 0x0000, []byte{0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, len(tt.want))
			PutUint16BE(buf, tt.offset, tt.value)
			for i := range tt.want {
				if buf[i] != tt.want[i] {
					t.Errorf("PutUint16BE() = %v, want %v", buf, tt.want)
					break
				}
			}
		})
	}
}

func TestPutUint16LE(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		value  uint16
		want   []byte
	}{
		{"offset 0", 0, 0x1234, []byte{0x34, 0x12, 0x00, 0x00}},
		{"offset 1", 1, 0x1234, []byte{0x00, 0x34, 0x12, 0x00}},
		{"offset 2", 2, 0xABCD, []byte{0x00, 0x00, 0xCD, 0xAB}},
		{"max value", 0, 0xFFFF, []byte{0xFF, 0xFF}},
		{"zero", 0, 0x0000, []byte{0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, len(tt.want))
			PutUint16LE(buf, tt.offset, tt.value)
			for i := range tt.want {
				if buf[i] != tt.want[i] {
					t.Errorf("PutUint16LE() = %v, want %v", buf, tt.want)
					break
				}
			}
		})
	}
}

func TestPutUint32BE(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		value  uint32
		want   []byte
	}{
		{"offset 0", 0, 0x12345678, []byte{0x12, 0x34, 0x56, 0x78}},
		{"offset 1", 1, 0x12345678, []byte{0x00, 0x12, 0x34, 0x56, 0x78}},
		{"max value", 0, 0xFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"zero", 0, 0x00000000, []byte{0x00, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, len(tt.want))
			PutUint32BE(buf, tt.offset, tt.value)
			for i := range tt.want {
				if buf[i] != tt.want[i] {
					t.Errorf("PutUint32BE() = %v, want %v", buf, tt.want)
					break
				}
			}
		})
	}
}

func TestPutUint32LE(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		value  uint32
		want   []byte
	}{
		{"offset 0", 0, 0x12345678, []byte{0x78, 0x56, 0x34, 0x12}},
		{"offset 1", 1, 0x12345678, []byte{0x00, 0x78, 0x56, 0x34, 0x12}},
		{"max value", 0, 0xFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"zero", 0, 0x00000000, []byte{0x00, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, len(tt.want))
			PutUint32LE(buf, tt.offset, tt.value)
			for i := range tt.want {
				if buf[i] != tt.want[i] {
					t.Errorf("PutUint32LE() = %v, want %v", buf, tt.want)
					break
				}
			}
		})
	}
}

func TestPutUint64BE(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		value  uint64
		want   []byte
	}{
		{"offset 0", 0, 0x123456789ABCDEF0, []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}},
		{"offset 1", 1, 0x123456789ABCDEF0, []byte{0x00, 0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}},
		{"max value", 0, 0xFFFFFFFFFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"zero", 0, 0x0000000000000000, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, len(tt.want))
			PutUint64BE(buf, tt.offset, tt.value)
			for i := range tt.want {
				if buf[i] != tt.want[i] {
					t.Errorf("PutUint64BE() = %v, want %v", buf, tt.want)
					break
				}
			}
		})
	}
}

func TestPutUint64LE(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		value  uint64
		want   []byte
	}{
		{"offset 0", 0, 0x123456789ABCDEF0, []byte{0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12}},
		{"offset 1", 1, 0x123456789ABCDEF0, []byte{0x00, 0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12}},
		{"max value", 0, 0xFFFFFFFFFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"zero", 0, 0x0000000000000000, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, len(tt.want))
			PutUint64LE(buf, tt.offset, tt.value)
			for i := range tt.want {
				if buf[i] != tt.want[i] {
					t.Errorf("PutUint64LE() = %v, want %v", buf, tt.want)
					break
				}
			}
		})
	}
}

// TestSliceConsistency verifies that our slice functions match encoding/binary behavior
func TestSliceConsistency(t *testing.T) {
	testData := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}

	t.Run("Uint16BE matches encoding/binary", func(t *testing.T) {
		for offset := 0; offset < len(testData)-1; offset++ {
			want := binary.BigEndian.Uint16(testData[offset:])
			got := Uint16BE(testData, offset)
			if got != want {
				t.Errorf("offset %d: Uint16BE() = 0x%04X, encoding/binary = 0x%04X", offset, got, want)
			}
		}
	})

	t.Run("Uint32LE matches encoding/binary", func(t *testing.T) {
		for offset := 0; offset < len(testData)-3; offset++ {
			want := binary.LittleEndian.Uint32(testData[offset:])
			got := Uint32LE(testData, offset)
			if got != want {
				t.Errorf("offset %d: Uint32LE() = 0x%08X, encoding/binary = 0x%08X", offset, got, want)
			}
		}
	})
}
