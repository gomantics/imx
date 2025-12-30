package bufpool

import "testing"

func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		size         int
		wantCapacity int
	}{
		{"size 1 gets 2-byte buffer", 1, Size2},
		{"size 2 gets 2-byte buffer", 2, Size2},
		{"size 3 gets 4-byte buffer", 3, Size4},
		{"size 4 gets 4-byte buffer", 4, Size4},
		{"size 5 gets 8-byte buffer", 5, Size8},
		{"size 8 gets 8-byte buffer", 8, Size8},
		{"size 9 gets 16-byte buffer", 9, Size16},
		{"size 16 gets 16-byte buffer", 16, Size16},
		{"size 17 gets 256-byte buffer", 17, Size256},
		{"size 256 gets 256-byte buffer", 256, Size256},
		{"size 257 gets 4096-byte buffer", 257, Size4096},
		{"size 4096 gets 4096-byte buffer", 4096, Size4096},
		{"size 4097 gets exact-size buffer", 4097, 4097},
		{"size 10000 gets exact-size buffer", 10000, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := Get(tt.size)
			if cap(buf) != tt.wantCapacity {
				t.Errorf("Get(%d) returned buffer with capacity %d, want %d", tt.size, cap(buf), tt.wantCapacity)
			}
			if len(buf) != tt.wantCapacity {
				t.Errorf("Get(%d) returned buffer with length %d, want %d", tt.size, len(buf), tt.wantCapacity)
			}
			Put(buf)
		})
	}
}

func TestPut(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"put 2-byte buffer", Size2},
		{"put 4-byte buffer", Size4},
		{"put 8-byte buffer", Size8},
		{"put 16-byte buffer", Size16},
		{"put 256-byte buffer", Size256},
		{"put 4096-byte buffer", Size4096},
		{"put non-standard buffer", 100},
		{"put large buffer", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := Get(tt.size)
			// Put should not panic
			Put(buf)
		})
	}
}

func TestPut_Nil(t *testing.T) {
	// Should not panic
	Put(nil)
}
