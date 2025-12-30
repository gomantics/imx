package heic

import (
	"bytes"
	"io"
	"testing"

	"github.com/gomantics/imx/internal/parser/icc"
	"github.com/gomantics/imx/internal/parser/tiff"
	"github.com/gomantics/imx/internal/parser/xmp"
)

func TestDescribesPrimaryItem(t *testing.T) {
	tests := []struct {
		name      string
		item      *HeifItem
		primaryID uint32
		want      bool
	}{
		{
			name: "item references primary",
			item: &HeifItem{
				ItemID:     1,
				References: []uint32{2, 3, 4},
			},
			primaryID: 3,
			want:      true,
		},
		{
			name: "item does not reference primary",
			item: &HeifItem{
				ItemID:     1,
				References: []uint32{2, 4},
			},
			primaryID: 3,
			want:      false,
		},
		{
			name: "empty references",
			item: &HeifItem{
				ItemID:     1,
				References: []uint32{},
			},
			primaryID: 3,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := describesPrimaryItem(tt.item, tt.primaryID)
			if got != tt.want {
				t.Errorf("describesPrimaryItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindTIFFHeader(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{
			name: "big-endian TIFF at start",
			data: []byte{'M', 'M', 0x00, 0x2A},
			want: 0,
		},
		{
			name: "little-endian TIFF at start",
			data: []byte{'I', 'I', 0x2A, 0x00},
			want: 0,
		},
		{
			name: "TIFF after padding",
			data: []byte{0x00, 0x00, 0x00, 0x00, 'M', 'M', 0x00, 0x2A},
			want: 4,
		},
		{
			name: "no TIFF header",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			want: -1,
		},
		{
			name: "empty data",
			data: []byte{},
			want: -1,
		},
		{
			name: "TIFF header beyond scan limit",
			data: append(make([]byte, 25), 'M', 'M'),
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findTIFFHeader(tt.data)
			if got != tt.want {
				t.Errorf("findTIFFHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsXMPData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "xpacket signature",
			data: []byte("<?xpacket begin='...'"),
			want: true,
		},
		{
			name: "xmpmeta signature",
			data: []byte("<x:xmpmeta xmlns:x='...'>"),
			want: true,
		},
		{
			name: "no XMP signature",
			data: []byte("<html><body>Hello</body></html>"),
			want: false,
		},
		{
			name: "empty data",
			data: []byte{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isXMPData(tt.data)
			if got != tt.want {
				t.Errorf("isXMPData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveNullBytes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{
			name: "no null bytes",
			data: []byte("hello world"),
			want: []byte("hello world"),
		},
		{
			name: "null bytes in middle",
			data: []byte{'h', 'e', 0, 'l', 0, 'l', 'o'},
			want: []byte("hello"),
		},
		{
			name: "only null bytes",
			data: []byte{0, 0, 0, 0},
			want: []byte{},
		},
		{
			name: "empty data",
			data: []byte{},
			want: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeNullBytes(tt.data)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("removeNullBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadItemData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		item    *HeifItem
		want    []byte
		wantErr bool
	}{
		{
			name: "single extent",
			data: []byte("hello world"),
			item: &HeifItem{
				Location: ItemLocation{
					BaseOffset: 0,
					Extents:    []Extent{{Offset: 0, Length: 5}},
				},
			},
			want:    []byte("hello"),
			wantErr: false,
		},
		{
			name: "multiple extents",
			data: []byte("hello world test"),
			item: &HeifItem{
				Location: ItemLocation{
					BaseOffset: 0,
					Extents: []Extent{
						{Offset: 0, Length: 5},
						{Offset: 6, Length: 5},
					},
				},
			},
			want:    []byte("helloworld"),
			wantErr: false,
		},
		{
			name: "with base offset",
			data: []byte("XXXXhello"),
			item: &HeifItem{
				Location: ItemLocation{
					BaseOffset: 4,
					Extents:    []Extent{{Offset: 0, Length: 5}},
				},
			},
			want:    []byte("hello"),
			wantErr: false,
		},
		{
			name: "empty extents",
			data: []byte("hello"),
			item: &HeifItem{
				Location: ItemLocation{
					BaseOffset: 0,
					Extents:    []Extent{},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "read error - offset beyond data",
			data: []byte("hello"),
			item: &HeifItem{
				Location: ItemLocation{
					BaseOffset: 100,
					Extents:    []Extent{{Offset: 0, Length: 5}},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got, err := readItemData(r, tt.item)

			if tt.wantErr {
				if err == nil {
					t.Error("readItemData() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("readItemData() error = %v", err)
			}

			if !bytes.Equal(got, tt.want) {
				t.Errorf("readItemData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractMetadata_NoPrimaryItem(t *testing.T) {
	p := New()
	index := &HeifIndex{
		PrimaryItemID: 999, // Non-existent
		Items:         make(map[uint32]*HeifItem),
	}

	r := bytes.NewReader([]byte{})
	dirs := p.extractMetadata(r, index)

	if len(dirs) != 0 {
		t.Errorf("extractMetadata() returned %d dirs, want 0", len(dirs))
	}
}

func TestExtractExif_Errors(t *testing.T) {
	p := &Parser{
		tiff: tiff.New(),
		xmp:  xmp.New(),
		icc:  icc.New(),
	}

	tests := []struct {
		name string
		data []byte
		item *HeifItem
	}{
		{
			name: "data too short",
			data: []byte{0, 0, 0, 0, 0, 0, 0},
			item: &HeifItem{
				Location: ItemLocation{
					Extents: []Extent{{Offset: 0, Length: 7}},
				},
			},
		},
		{
			name: "invalid tiff offset - too small",
			data: []byte{0, 0, 0, 2, 'M', 'M', 0x00, 0x2A}, // offset=2, but TIFF at offset 4
			item: &HeifItem{
				Location: ItemLocation{
					Extents: []Extent{{Offset: 0, Length: 8}},
				},
			},
		},
		{
			name: "invalid tiff offset - beyond data",
			data: []byte{0, 0, 0, 100, 'M', 'M', 0x00, 0x2A}, // offset=100, beyond length
			item: &HeifItem{
				Location: ItemLocation{
					Extents: []Extent{{Offset: 0, Length: 8}},
				},
			},
		},
		{
			name: "no TIFF header found",
			data: []byte{0, 0, 0, 4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			item: &HeifItem{
				Location: ItemLocation{
					Extents: []Extent{{Offset: 0, Length: 12}},
				},
			},
		},
		{
			name: "TIFF data too short after header",
			data: []byte{0, 0, 0, 4, 'M', 'M', 0x00, 0x2A}, // only 4 bytes after offset
			item: &HeifItem{
				Location: ItemLocation{
					Extents: []Extent{{Offset: 0, Length: 8}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			dirs := p.extractExif(r, tt.item)
			if dirs != nil {
				t.Errorf("extractExif() = %v, want nil", dirs)
			}
		})
	}
}

func TestExtractExif_ReadError(t *testing.T) {
	p := &Parser{tiff: tiff.New()}

	item := &HeifItem{
		Location: ItemLocation{
			BaseOffset: 1000, // Beyond data
			Extents:    []Extent{{Offset: 0, Length: 100}},
		},
	}

	r := bytes.NewReader([]byte("small data"))
	dirs := p.extractExif(r, item)

	if dirs != nil {
		t.Errorf("extractExif() = %v, want nil", dirs)
	}
}

func TestExtractXMP_Errors(t *testing.T) {
	p := &Parser{xmp: xmp.New()}

	tests := []struct {
		name string
		data []byte
		item *HeifItem
	}{
		{
			name: "empty data",
			data: []byte{},
			item: &HeifItem{
				Location: ItemLocation{
					Extents: []Extent{},
				},
			},
		},
		{
			name: "not XMP data",
			data: []byte("<html>not xmp</html>"),
			item: &HeifItem{
				Location: ItemLocation{
					Extents: []Extent{{Offset: 0, Length: 20}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			dirs := p.extractXMP(r, tt.item)
			if dirs != nil {
				t.Errorf("extractXMP() = %v, want nil", dirs)
			}
		})
	}
}

func TestExtractXMP_ReadError(t *testing.T) {
	p := &Parser{xmp: xmp.New()}

	item := &HeifItem{
		Location: ItemLocation{
			BaseOffset: 1000,
			Extents:    []Extent{{Offset: 0, Length: 100}},
		},
	}

	r := bytes.NewReader([]byte("small"))
	dirs := p.extractXMP(r, item)

	if dirs != nil {
		t.Errorf("extractXMP() = %v, want nil", dirs)
	}
}

func TestExtractICC_Errors(t *testing.T) {
	p := &Parser{icc: icc.New()}

	tests := []struct {
		name string
		data []byte
		item *HeifItem
	}{
		{
			name: "no ICC property",
			data: []byte{},
			item: &HeifItem{
				ICCProperty: nil,
			},
		},
		{
			name: "not ICC color type",
			data: []byte{'n', 'c', 'l', 'x', 0, 0, 0, 0}, // nclx instead of rICC/prof
			item: &HeifItem{
				ICCProperty: &Box{
					Type:    "colr",
					Size:    12,
					Offset:  0,
					Payload: 0,
				},
			},
		},
		{
			name: "zero ICC size",
			data: []byte{'r', 'I', 'C', 'C'},
			item: &HeifItem{
				ICCProperty: &Box{
					Type:    "colr",
					Size:    8, // Size equals header, so ICC data size is 0
					Offset:  0,
					Payload: 4,
				},
			},
		},
		{
			name: "negative ICC size",
			data: []byte{0, 0, 0, 0, 'p', 'r', 'o', 'f'}, // Header at offset 0, color type at offset 4
			item: &HeifItem{
				ICCProperty: &Box{
					Type:    "colr",
					Size:    5, // Very small size
					Offset:  0,
					Payload: 4, // iccSize = 5 - (4 - 0) - 4 = -3 < 0
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			dirs := p.extractICC(r, tt.item)
			if dirs != nil {
				t.Errorf("extractICC() = %v, want nil", dirs)
			}
		})
	}
}

func TestExtractICC_ReadErrors(t *testing.T) {
	p := &Parser{icc: icc.New()}

	// Test header read error
	t.Run("header read error", func(t *testing.T) {
		item := &HeifItem{
			ICCProperty: &Box{
				Type:    "colr",
				Size:    100,
				Offset:  0,
				Payload: 50, // Beyond data
			},
		}

		r := bytes.NewReader([]byte("small"))
		dirs := p.extractICC(r, item)

		if dirs != nil {
			t.Errorf("extractICC() = %v, want nil", dirs)
		}
	})

	// Test ICC data read error
	t.Run("ICC data read error", func(t *testing.T) {
		item := &HeifItem{
			ICCProperty: &Box{
				Type:    "colr",
				Size:    100,
				Offset:  0,
				Payload: 0,
			},
		}

		r := bytes.NewReader([]byte{'r', 'I', 'C', 'C'}) // Only header, no ICC data
		dirs := p.extractICC(r, item)

		if dirs != nil {
			t.Errorf("extractICC() = %v, want nil", dirs)
		}
	})
}

// errorReaderAt for testing read errors
type errorReaderAt struct {
	data        []byte
	errorOffset int64
	customError error
}

func (e *errorReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off >= e.errorOffset {
		return 0, e.customError
	}
	if off >= int64(len(e.data)) {
		return 0, io.EOF
	}
	n := copy(p, e.data[off:])
	return n, nil
}

func TestReadItemData_IncompleteRead(t *testing.T) {
	// Create a reader that returns fewer bytes than requested
	item := &HeifItem{
		Location: ItemLocation{
			BaseOffset: 0,
			Extents:    []Extent{{Offset: 0, Length: 100}}, // Request 100 bytes
		},
	}

	r := bytes.NewReader([]byte("short")) // Only 5 bytes available
	_, err := readItemData(r, item)

	if err == nil {
		t.Error("readItemData() expected error for incomplete read")
	}
}

// shortReadReaderAt returns fewer bytes than requested but no error
type shortReadReaderAt struct {
	data      []byte
	maxReturn int
}

func (s *shortReadReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(s.data)) {
		return 0, io.EOF
	}
	n := copy(p, s.data[off:])
	if n > s.maxReturn {
		n = s.maxReturn
	}
	return n, nil
}

func TestReadItemData_PartialRead(t *testing.T) {
	// Reader that returns partial data without error
	item := &HeifItem{
		Location: ItemLocation{
			BaseOffset: 0,
			Extents:    []Extent{{Offset: 0, Length: 20}}, // Request 20 bytes
		},
	}

	r := &shortReadReaderAt{
		data:      make([]byte, 100),
		maxReturn: 5, // Only return 5 bytes at a time
	}

	_, err := readItemData(r, item)
	if err == nil {
		t.Error("readItemData() expected error for partial read")
	}
}
