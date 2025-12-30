package heic

import (
	"bytes"
	"io"
	"testing"
)

func TestReadBoxHeader(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		offset   int64
		wantType string
		wantSize uint64
		wantErr  bool
	}{
		{
			name:     "valid ftyp box",
			data:     []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c'},
			offset:   0,
			wantType: "ftyp",
			wantSize: 24,
			wantErr:  false,
		},
		{
			name:     "valid meta box",
			data:     []byte{0, 0, 0, 100, 'm', 'e', 't', 'a'},
			offset:   0,
			wantType: "meta",
			wantSize: 100,
			wantErr:  false,
		},
		{
			name: "64-bit size (extended)",
			data: []byte{
				0, 0, 0, 1, 'f', 't', 'y', 'p', // size=1 means extended
				0, 0, 0, 0, 0, 0, 1, 0, // 64-bit size = 256
			},
			offset:   0,
			wantType: "ftyp",
			wantSize: 256,
			wantErr:  false,
		},
		{
			name:    "size=0 not supported",
			data:    []byte{0, 0, 0, 0, 'f', 't', 'y', 'p'},
			offset:  0,
			wantErr: true,
		},
		{
			name:    "invalid box type - non-printable",
			data:    []byte{0, 0, 0, 24, 0x01, 0x02, 0x03, 0x04},
			offset:  0,
			wantErr: true,
		},
		{
			name:    "truncated header",
			data:    []byte{0, 0, 0, 24},
			offset:  0,
			wantErr: true,
		},
		{
			name:    "read error",
			data:    []byte{},
			offset:  0,
			wantErr: true,
		},
		{
			name: "extended size read error",
			data: []byte{
				0, 0, 0, 1, 'f', 't', 'y', 'p', // size=1 means extended
				0, 0, 0, 0, // only 4 bytes instead of 8
			},
			offset:  0,
			wantErr: true,
		},
		{
			name: "extended size too small - causes infinite loop without validation",
			data: []byte{
				0, 0, 0, 1, 'f', 't', 'y', 'p', // size=1 means extended
				0, 0, 0, 0, 0, 0, 0, 4, // large size = 4 (invalid, less than boxHeaderSize=8)
			},
			offset:  0,
			wantErr: true,
		},
		{
			name: "extended size exceeds maxBoxSize",
			data: []byte{
				0, 0, 0, 1, 'f', 't', 'y', 'p', // size=1 means extended
				0, 0, 0, 0, 0x10, 0, 0, 0, // 64-bit size > 100MB (maxBoxSize)
			},
			offset:  0,
			wantErr: true,
		},
		{
			name: "standard size exceeds maxBoxSize",
			data: []byte{
				0x10, 0, 0, 0, 'f', 't', 'y', 'p', // size > 100MB (maxBoxSize)
			},
			offset:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			box, err := readBoxHeader(r, tt.offset)

			if tt.wantErr {
				if err == nil {
					t.Error("readBoxHeader() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("readBoxHeader() error = %v", err)
			}

			if box.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", box.Type, tt.wantType)
			}
			if box.Size != tt.wantSize {
				t.Errorf("Size = %v, want %v", box.Size, tt.wantSize)
			}
		})
	}
}

func TestFindBox(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		boxType string
		maxScan int64
		wantErr bool
	}{
		{
			name: "find meta box",
			data: []byte{
				0, 0, 0, 16, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c', 0, 0, 0, 0,
				0, 0, 0, 16, 'm', 'e', 't', 'a', 0, 0, 0, 0, 0, 0, 0, 0,
			},
			boxType: "meta",
			maxScan: 1000,
			wantErr: false,
		},
		{
			name: "box not found within limit",
			data: []byte{
				0, 0, 0, 16, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c', 0, 0, 0, 0,
			},
			boxType: "meta",
			maxScan: 100,
			wantErr: true,
		},
		{
			name:    "empty data - EOF error",
			data:    []byte{},
			boxType: "meta",
			maxScan: 100,
			wantErr: true,
		},
		{
			name:    "invalid box type error propagation",
			data:    []byte{0, 0, 0, 24, 0x01, 0x02, 0x03, 0x04}, // non-printable
			boxType: "meta",
			maxScan: 100,
			wantErr: true,
		},
		{
			name: "box not found within maxScan - scan exhausted",
			data: []byte{
				0, 0, 0, 12, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c',
				0, 0, 0, 12, 'm', 'd', 'a', 't', 0, 0, 0, 0,
			},
			boxType: "meta",
			maxScan: 24, // Exactly matches data size, so we scan all boxes but don't find meta
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			box, err := findBox(r, tt.boxType, 0, tt.maxScan)

			if tt.wantErr {
				if err == nil {
					t.Error("findBox() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("findBox() error = %v", err)
			}

			if box.Type != tt.boxType {
				t.Errorf("Type = %v, want %v", box.Type, tt.boxType)
			}
		})
	}
}

func TestFindChildBox(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		parent    *Box
		childType string
		wantErr   bool
	}{
		{
			name: "find child box",
			data: []byte{
				0, 0, 0, 32, 'm', 'e', 't', 'a', // parent
				0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0, // child 1
				0, 0, 0, 12, 'p', 'i', 't', 'm', 0, 0, 0, 0, // child 2
			},
			parent: &Box{
				Type:    "meta",
				Size:    32,
				Offset:  0,
				Payload: 8,
			},
			childType: "pitm",
			wantErr:   false,
		},
		{
			name: "child not found",
			data: []byte{
				0, 0, 0, 20, 'm', 'e', 't', 'a', // parent
				0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0, // child
			},
			parent: &Box{
				Type:    "meta",
				Size:    20,
				Offset:  0,
				Payload: 8,
			},
			childType: "pitm",
			wantErr:   true,
		},
		{
			name: "read error in child",
			data: []byte{
				0, 0, 0, 100, 'm', 'e', 't', 'a', // parent claims size 100
			},
			parent: &Box{
				Type:    "meta",
				Size:    100,
				Offset:  0,
				Payload: 8,
			},
			childType: "hdlr",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			box, err := findChildBox(r, tt.parent, tt.childType)

			if tt.wantErr {
				if err == nil {
					t.Error("findChildBox() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("findChildBox() error = %v", err)
			}

			if box.Type != tt.childType {
				t.Errorf("Type = %v, want %v", box.Type, tt.childType)
			}
		})
	}
}

func TestIterateChildren(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		parent    *Box
		wantCount int
		wantErr   bool
	}{
		{
			name: "iterate two children",
			data: []byte{
				0, 0, 0, 32, 'p', 'a', 'r', 'n', // parent
				0, 0, 0, 12, 'c', 'h', 'd', '1', 0, 0, 0, 0, // child 1
				0, 0, 0, 12, 'c', 'h', 'd', '2', 0, 0, 0, 0, // child 2
			},
			parent: &Box{
				Type:    "parn",
				Size:    32,
				Offset:  0,
				Payload: 8,
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "not enough space for header",
			data: []byte{
				0, 0, 0, 12, 'p', 'a', 'r', 'n', // parent with 4 bytes payload
				0, 0, 0, 0, // only 4 bytes, not enough for box header
			},
			parent: &Box{
				Type:    "parn",
				Size:    12,
				Offset:  0,
				Payload: 8,
			},
			wantCount: 0,
			wantErr:   false, // Should just stop iterating
		},
		{
			name: "read error",
			data: []byte{
				0, 0, 0, 100, 'p', 'a', 'r', 'n', // claims 100 bytes
			},
			parent: &Box{
				Type:    "parn",
				Size:    100,
				Offset:  0,
				Payload: 8,
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			count := 0
			err := iterateChildren(r, tt.parent, func(box *Box) error {
				count++
				return nil
			})

			if tt.wantErr {
				if err == nil {
					t.Error("iterateChildren() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("iterateChildren() error = %v", err)
			}

			if count != tt.wantCount {
				t.Errorf("count = %v, want %v", count, tt.wantCount)
			}
		})
	}
}

func TestIterateChildren_CallbackError(t *testing.T) {
	data := []byte{
		0, 0, 0, 20, 'p', 'a', 'r', 'n',
		0, 0, 0, 12, 'c', 'h', 'd', '1', 0, 0, 0, 0,
	}
	parent := &Box{
		Type:    "parn",
		Size:    20,
		Offset:  0,
		Payload: 8,
	}

	r := bytes.NewReader(data)
	callbackErr := io.ErrUnexpectedEOF
	err := iterateChildren(r, parent, func(box *Box) error {
		return callbackErr
	})

	if err != callbackErr {
		t.Errorf("expected callback error, got %v", err)
	}
}

func TestBoxTypeEquals(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
		want     bool
	}{
		{
			name:     "matching type",
			data:     []byte{'f', 't', 'y', 'p'},
			expected: "ftyp",
			want:     true,
		},
		{
			name:     "non-matching type",
			data:     []byte{'m', 'e', 't', 'a'},
			expected: "ftyp",
			want:     false,
		},
		{
			name:     "data too short",
			data:     []byte{'f', 't'},
			expected: "ftyp",
			want:     false,
		},
		{
			name:     "expected wrong length",
			data:     []byte{'f', 't', 'y', 'p'},
			expected: "fty",
			want:     false,
		},
		{
			name:     "empty data",
			data:     []byte{},
			expected: "ftyp",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boxTypeEquals(tt.data, tt.expected)
			if got != tt.want {
				t.Errorf("boxTypeEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadUint(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		size int
		want uint64
	}{
		{
			name: "1 byte",
			data: []byte{0xFF},
			size: 1,
			want: 255,
		},
		{
			name: "2 bytes",
			data: []byte{0x01, 0x00},
			size: 2,
			want: 256,
		},
		{
			name: "4 bytes",
			data: []byte{0x00, 0x01, 0x00, 0x00},
			size: 4,
			want: 65536,
		},
		{
			name: "8 bytes",
			data: []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
			size: 8,
			want: 4294967296,
		},
		{
			name: "size 0",
			data: []byte{0xFF},
			size: 0,
			want: 0,
		},
		{
			name: "negative size",
			data: []byte{0xFF},
			size: -1,
			want: 0,
		},
		{
			name: "size > 8",
			data: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			size: 9,
			want: 0,
		},
		{
			name: "data too short",
			data: []byte{0xFF},
			size: 4,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readUint(tt.data, tt.size)
			if got != tt.want {
				t.Errorf("readUint() = %v, want %v", got, tt.want)
			}
		})
	}
}
