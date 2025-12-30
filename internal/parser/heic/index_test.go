package heic

import (
	"bytes"
	"io"
	"testing"
)

func TestBuildHeifIndex(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		metaBox *Box
		maxScan int64
		wantErr bool
	}{
		{
			name: "missing hdlr box",
			data: []byte{
				// meta box with no hdlr
				0, 0, 0, 0, // version/flags
			},
			metaBox: &Box{
				Type:    "meta",
				Size:    12,
				Offset:  0,
				Payload: 0,
			},
			maxScan: 100,
			wantErr: true,
		},
		{
			name: "missing iinf box",
			data: append([]byte{
				0, 0, 0, 0, // version/flags for meta
				// hdlr box
				0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0,
			}, make([]byte, 100)...),
			metaBox: &Box{
				Type:    "meta",
				Size:    120,
				Offset:  0,
				Payload: 0,
			},
			maxScan: 200,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			_, err := buildHeifIndex(r, tt.metaBox, tt.maxScan)

			if tt.wantErr {
				if err == nil {
					t.Error("buildHeifIndex() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("buildHeifIndex() error = %v", err)
			}
		})
	}
}

func TestBuildHeifIndex_PitmError(t *testing.T) {
	// Create data with hdlr, pitm (that will fail), and iinf
	data := append([]byte{
		0, 0, 0, 0, // version/flags for meta
		// hdlr box
		0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0,
		// pitm box (too small to parse)
		0, 0, 0, 10, 'p', 'i', 't', 'm', 0, 0,
	}, make([]byte, 100)...)

	metaBox := &Box{
		Type:    "meta",
		Size:    130,
		Offset:  0,
		Payload: 0,
	}

	r := &errorReaderAt{
		data:        data,
		errorOffset: 20, // Error when reading pitm payload
		customError: io.ErrUnexpectedEOF,
	}

	_, err := buildHeifIndex(r, metaBox, 200)
	if err == nil {
		t.Error("buildHeifIndex() expected error for pitm parse failure")
	}
}

func TestParsePitm(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		box     *Box
		wantID  uint32
		wantErr bool
	}{
		{
			name: "version 0",
			data: []byte{
				0, 0, 0, 0, // version and flags
				0, 42, // item ID = 42
				0, 0, // padding
			},
			box: &Box{
				Type:    "pitm",
				Size:    14,
				Offset:  0,
				Payload: 0,
			},
			wantID:  42,
			wantErr: false,
		},
		{
			name: "version 1",
			data: []byte{
				1, 0, 0, 0, // version 1 and flags
				0, 0, 0, 99, // item ID = 99
			},
			box: &Box{
				Type:    "pitm",
				Size:    16,
				Offset:  0,
				Payload: 0,
			},
			wantID:  99,
			wantErr: false,
		},
		{
			name:    "read error",
			data:    []byte{0, 0}, // Too short
			box:     &Box{Payload: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			index := &HeifIndex{}

			err := parsePitm(r, tt.box, index)

			if tt.wantErr {
				if err == nil {
					t.Error("parsePitm() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parsePitm() error = %v", err)
			}

			if index.PrimaryItemID != tt.wantID {
				t.Errorf("PrimaryItemID = %v, want %v", index.PrimaryItemID, tt.wantID)
			}
		})
	}
}

func TestParseIlocItemID(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		version    uint8
		wantPos    int
		wantItemID uint32
	}{
		{
			name:       "version 0/1 - 16-bit ID",
			data:       []byte{0, 42, 0, 0, 0, 0},
			version:    0,
			wantPos:    2,
			wantItemID: 42,
		},
		{
			name:       "version 2 - 32-bit ID",
			data:       []byte{0, 0, 0, 99, 0, 0},
			version:    2,
			wantPos:    4,
			wantItemID: 99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, itemID := parseIlocItemID(tt.data, tt.version)

			if pos != tt.wantPos {
				t.Errorf("pos = %v, want %v", pos, tt.wantPos)
			}
			if itemID != tt.wantItemID {
				t.Errorf("itemID = %v, want %v", itemID, tt.wantItemID)
			}
		})
	}
}

func TestParseIlocExtents(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		count      uint16
		version    uint8
		indexSize  uint8
		offsetSize uint8
		lengthSize uint8
		wantCount  int
	}{
		{
			name:       "single extent, 4-byte offset and length",
			data:       []byte{0, 0, 0, 10, 0, 0, 0, 20},
			count:      1,
			version:    0,
			indexSize:  0,
			offsetSize: 4,
			lengthSize: 4,
			wantCount:  1,
		},
		{
			name:       "two extents, 2-byte offset and length",
			data:       []byte{0, 10, 0, 20, 0, 30, 0, 40},
			count:      2,
			version:    0,
			indexSize:  0,
			offsetSize: 2,
			lengthSize: 2,
			wantCount:  2,
		},
		{
			name:       "with index size (version 1)",
			data:       []byte{0, 0, 0, 10, 0, 0, 0, 20}, // 2-byte index + 2-byte offset + 2-byte length
			count:      1,
			version:    1,
			indexSize:  2,
			offsetSize: 2,
			lengthSize: 2,
			wantCount:  1,
		},
		{
			name:       "zero offset size",
			data:       []byte{0, 0, 0, 20},
			count:      1,
			version:    0,
			indexSize:  0,
			offsetSize: 0,
			lengthSize: 4,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extents := parseIlocExtents(tt.data, tt.count, tt.version, tt.indexSize, tt.offsetSize, tt.lengthSize)

			if len(extents) != tt.wantCount {
				t.Errorf("len(extents) = %v, want %v", len(extents), tt.wantCount)
			}
		})
	}
}

func TestParseIpmaItemID(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		version    uint8
		wantPos    int
		wantItemID uint32
	}{
		{
			name:       "version 0 - 16-bit ID",
			data:       []byte{0, 42, 0, 0, 0, 0},
			version:    0,
			wantPos:    2,
			wantItemID: 42,
		},
		{
			name:       "version 1 - 32-bit ID",
			data:       []byte{0, 0, 0, 99, 0, 0},
			version:    1,
			wantPos:    4,
			wantItemID: 99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, itemID := parseIpmaItemID(tt.data, tt.version)

			if pos != tt.wantPos {
				t.Errorf("pos = %v, want %v", pos, tt.wantPos)
			}
			if itemID != tt.wantItemID {
				t.Errorf("itemID = %v, want %v", itemID, tt.wantItemID)
			}
		})
	}
}

func TestParseIpmaProperty(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		flags         uint32
		wantPropIndex uint32
		wantBytes     int
	}{
		{
			name:          "7-bit mode (flags=0)",
			data:          []byte{0x42}, // index=66 (0x42 & 0x7F)
			flags:         0,
			wantPropIndex: 66,
			wantBytes:     1,
		},
		{
			name:          "7-bit mode with essential flag",
			data:          []byte{0x82}, // essential=1, index=2
			flags:         0,
			wantPropIndex: 2,
			wantBytes:     1,
		},
		{
			name:          "15-bit mode (flags=1)",
			data:          []byte{0x01, 0x23}, // index=291 (0x0123 & 0x7FFF)
			flags:         1,
			wantPropIndex: 291,
			wantBytes:     2,
		},
		{
			name:          "15-bit mode with essential flag",
			data:          []byte{0x80, 0x05}, // essential=1, index=5
			flags:         1,
			wantPropIndex: 5,
			wantBytes:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			propIndex, bytesRead := parseIpmaProperty(tt.data, tt.flags)

			if propIndex != tt.wantPropIndex {
				t.Errorf("propIndex = %v, want %v", propIndex, tt.wantPropIndex)
			}
			if bytesRead != tt.wantBytes {
				t.Errorf("bytesRead = %v, want %v", bytesRead, tt.wantBytes)
			}
		})
	}
}

func TestParseIrefCdsc(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		version uint8
		index   *HeifIndex
	}{
		{
			name: "version 0 - single reference",
			data: []byte{
				0, 1, // from ID = 1
				0, 1, // ref count = 1
				0, 2, // to ID = 2
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // padding to 32 bytes
			},
			version: 0,
			index: &HeifIndex{
				Items: map[uint32]*HeifItem{
					1: {ItemID: 1},
					2: {ItemID: 2},
				},
			},
		},
		{
			name: "version 1 - single reference",
			data: []byte{
				0, 0, 0, 1, // from ID = 1
				0, 1, // ref count = 1
				0, 0, 0, 2, // to ID = 2
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // padding to 32 bytes
			},
			version: 1,
			index: &HeifIndex{
				Items: map[uint32]*HeifItem{
					1: {ItemID: 1},
					2: {ItemID: 2},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			refBox := &Box{
				Type:    "cdsc",
				Size:    uint64(len(tt.data) + 8),
				Offset:  0,
				Payload: 0,
			}

			parseIrefCdsc(r, refBox, tt.version, tt.index)

			// Check that references were added
			if item1, ok := tt.index.Items[1]; ok {
				if len(item1.References) == 0 {
					t.Error("item 1 should have references")
				}
			}
		})
	}
}

func TestParseIrefCdsc_ReadError(t *testing.T) {
	r := bytes.NewReader([]byte{})
	refBox := &Box{
		Type:    "cdsc",
		Size:    100,
		Offset:  0,
		Payload: 0,
	}
	index := &HeifIndex{Items: make(map[uint32]*HeifItem)}

	// Should not panic
	parseIrefCdsc(r, refBox, 0, index)
}

func TestParseInfe(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		infeBox   *Box
		parentBox *Box
		wantID    uint32
		wantType  string
		wantErr   bool
	}{
		{
			name: "version 2 infe",
			data: []byte{
				2, 0, 0, 0, // version 2 and flags
				0, 42, // item ID = 42
				0, 0, // protection index
				'E', 'x', 'i', 'f', // item type
				0, 0, 0, 0, // padding
			},
			infeBox: &Box{
				Type:    "infe",
				Size:    20,
				Offset:  0,
				Payload: 0,
			},
			parentBox: &Box{Size: 100, Offset: 0},
			wantID:    42,
			wantType:  "Exif",
			wantErr:   false,
		},
		{
			name: "version 3 infe",
			data: []byte{
				3, 0, 0, 0, // version 3 and flags
				0, 99, // item ID = 99
				0, 0, // protection index
				'm', 'i', 'm', 'e', // item type
				0, 0, 0, 0, // padding
			},
			infeBox: &Box{
				Type:    "infe",
				Size:    20,
				Offset:  0,
				Payload: 0,
			},
			parentBox: &Box{Size: 100, Offset: 0},
			wantID:    99,
			wantType:  "mime",
			wantErr:   false,
		},
		{
			name:      "read error",
			data:      []byte{0, 0}, // Too short
			infeBox:   &Box{Payload: 0},
			parentBox: &Box{Size: 100, Offset: 0},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			index := &HeifIndex{Items: make(map[uint32]*HeifItem)}

			err := parseInfe(r, tt.infeBox, tt.parentBox, index)

			if tt.wantErr {
				if err == nil {
					t.Error("parseInfe() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseInfe() error = %v", err)
			}

			item, exists := index.Items[tt.wantID]
			if !exists {
				t.Fatalf("item %d not found in index", tt.wantID)
			}

			if item.ItemType != tt.wantType {
				t.Errorf("ItemType = %v, want %v", item.ItemType, tt.wantType)
			}
		})
	}
}

func TestParseIinf(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		box       *Box
		wantCount int
		wantErr   bool
	}{
		{
			name: "version 0 - single item",
			data: []byte{
				0, 0, 0, 0, // version 0 and flags
				0, 1, // item count = 1
				// infe box (with enough padding for parseInfe to read type data)
				0, 0, 0, 24, 'i', 'n', 'f', 'e',
				2, 0, 0, 0, // version 2
				0, 1, // item ID
				0, 0, // protection index
				'E', 'x', 'i', 'f', // item type
				0, 0, 0, 0, // padding for type read
			},
			box: &Box{
				Type:    "iinf",
				Size:    30,
				Offset:  0,
				Payload: 0,
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "version 1 - 32-bit count",
			data: []byte{
				1, 0, 0, 0, // version 1 and flags
				0, 0, 0, 1, // item count = 1
				// infe box
				0, 0, 0, 24, 'i', 'n', 'f', 'e',
				2, 0, 0, 0, // version 2
				0, 2, // item ID
				0, 0, // protection index
				'm', 'i', 'm', 'e',
				0, 0, 0, 0, // padding
			},
			box: &Box{
				Type:    "iinf",
				Size:    32,
				Offset:  0,
				Payload: 0,
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:    "read error",
			data:    []byte{0, 0}, // Too short
			box:     &Box{Payload: 0},
			wantErr: true,
		},
		{
			name: "wrong box type in infe",
			data: []byte{
				0, 0, 0, 0, // version 0
				0, 1, // count = 1
				0, 0, 0, 12, 'x', 'x', 'x', 'x', // wrong type
				0, 0, 0, 0,
			},
			box: &Box{
				Type:    "iinf",
				Size:    18,
				Offset:  0,
				Payload: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			index := &HeifIndex{Items: make(map[uint32]*HeifItem)}

			err := parseIinf(r, tt.box, index)

			if tt.wantErr {
				if err == nil {
					t.Error("parseIinf() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseIinf() error = %v", err)
			}

			if len(index.Items) != tt.wantCount {
				t.Errorf("len(Items) = %v, want %v", len(index.Items), tt.wantCount)
			}
		})
	}
}

func TestParseIloc(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		box     *Box
		wantErr bool
	}{
		{
			name: "version 0",
			data: append([]byte{
				0, 0, 0, 0, // version 0 and flags
				0x44, // offset_size=4, length_size=4
				0x00, // base_offset_size=0, index_size=0
				0, 1, // item count = 1
				0, 1, // item ID = 1
				0, 0, // data ref index
				0, 1, // extent count = 1
				0, 0, 0, 10, // extent offset = 10
				0, 0, 0, 20, // extent length = 20
			}, make([]byte, 64)...), // plenty of padding
			box: &Box{
				Type:    "iloc",
				Size:    100,
				Offset:  0,
				Payload: 0,
			},
			wantErr: false,
		},
		{
			name: "version 1 with construction method",
			data: append([]byte{
				1, 0, 0, 0, // version 1 and flags
				0x44, // offset_size=4, length_size=4
				0x00, // base_offset_size=0, index_size=0
				0, 1, // item count = 1
				0, 1, // item ID = 1
				0, 0, // construction method
				0, 0, // data ref index
				0, 1, // extent count = 1
				0, 0, 0, 10, // extent offset
				0, 0, 0, 20, // extent length
			}, make([]byte, 64)...), // plenty of padding
			box: &Box{
				Type:    "iloc",
				Size:    100,
				Offset:  0,
				Payload: 0,
			},
			wantErr: false,
		},
		{
			name: "version 2 with 32-bit item count",
			data: append([]byte{
				2, 0, 0, 0, // version 2 and flags
				0x44, // offset_size=4, length_size=4
				0x00, // base_offset_size=0, index_size=0
				0, 0, // reserved
				0, 0, 0, 1, // item count = 1
				0, 0, 0, 1, // item ID = 1 (32-bit)
				0, 0, // construction method
				0, 0, // data ref index
				0, 1, // extent count = 1
				0, 0, 0, 10, // extent offset
				0, 0, 0, 20, // extent length
			}, make([]byte, 64)...), // plenty of padding
			box: &Box{
				Type:    "iloc",
				Size:    100,
				Offset:  0,
				Payload: 0,
			},
			wantErr: false,
		},
		{
			name:    "read error",
			data:    []byte{0, 0}, // Too short
			box:     &Box{Payload: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			index := &HeifIndex{Items: make(map[uint32]*HeifItem)}

			err := parseIloc(r, tt.box, index)

			if tt.wantErr {
				if err == nil {
					t.Error("parseIloc() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseIloc() error = %v", err)
			}
		})
	}
}

func TestParseIref(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		box     *Box
		wantErr bool
	}{
		{
			name: "single cdsc reference",
			data: []byte{
				0, 0, 0, 0, // version 0 and flags
				// cdsc reference box
				0, 0, 0, 14, 'c', 'd', 's', 'c',
				0, 1, // from ID = 1
				0, 1, // ref count = 1
				0, 2, // to ID = 2
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // padding
			},
			box: &Box{
				Type:    "iref",
				Size:    36,
				Offset:  0,
				Payload: 0,
			},
			wantErr: false,
		},
		{
			name: "non-cdsc reference (skipped)",
			data: []byte{
				0, 0, 0, 0, // version 0
				// dimg reference box (not cdsc)
				0, 0, 0, 14, 'd', 'i', 'm', 'g',
				0, 1, // from ID
				0, 1, // ref count
				0, 2, // to ID
			},
			box: &Box{
				Type:    "iref",
				Size:    18,
				Offset:  0,
				Payload: 0,
			},
			wantErr: false,
		},
		{
			name:    "read error",
			data:    []byte{0, 0}, // Too short
			box:     &Box{Payload: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			index := &HeifIndex{
				Items: map[uint32]*HeifItem{
					1: {ItemID: 1},
					2: {ItemID: 2},
				},
			}

			err := parseIref(r, tt.box, index)

			if tt.wantErr {
				if err == nil {
					t.Error("parseIref() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseIref() error = %v", err)
			}
		})
	}
}

func TestParseIprp(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		box     *Box
		wantErr bool
	}{
		{
			name: "ipco with colr property",
			data: []byte{
				// ipco box
				0, 0, 0, 20, 'i', 'p', 'c', 'o',
				0, 0, 0, 12, 'c', 'o', 'l', 'r', // colr property
				'r', 'I', 'C', 'C',
				// ipma box
				0, 0, 0, 16, 'i', 'p', 'm', 'a',
				0, 0, 0, 0, // version 0 and flags
				0, 0, 0, 1, // entry count = 1
				0, 1, // item ID = 1
				1,    // assoc count = 1
				0x01, // property index = 1
			},
			box: &Box{
				Type:    "iprp",
				Size:    36,
				Offset:  0,
				Payload: 0,
			},
			wantErr: false,
		},
		{
			name: "no ipco - optional",
			data: []byte{
				// empty iprp - no ipco
			},
			box: &Box{
				Type:    "iprp",
				Size:    8,
				Offset:  0,
				Payload: 0,
			},
			wantErr: false, // ipco is optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			index := &HeifIndex{
				Items: map[uint32]*HeifItem{
					1: {ItemID: 1},
				},
			}

			err := parseIprp(r, tt.box, index)

			if tt.wantErr {
				if err == nil {
					t.Error("parseIprp() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseIprp() error = %v", err)
			}
		})
	}
}

func TestParseIpma(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		box        *Box
		properties []PropertyEntry
		wantErr    bool
	}{
		{
			name: "version 0 with 7-bit index",
			data: []byte{
				0, 0, 0, 0, // version 0, flags = 0 (7-bit index)
				0, 0, 0, 1, // entry count = 1
				0, 1, // item ID = 1
				2,          // assoc count = 2
				0x01, 0x02, // property indices
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // padding
			},
			box: &Box{
				Type:    "ipma",
				Size:    32,
				Offset:  0,
				Payload: 0,
			},
			properties: []PropertyEntry{
				{Index: 1, Type: "ispe"},
				{Index: 2, Type: "colr", Box: &Box{Type: "colr", Size: 12, Offset: 0, Payload: 0}},
			},
			wantErr: false,
		},
		{
			name: "version 0 with 15-bit index (flags=1)",
			data: []byte{
				0, 0, 0, 1, // version 0, flags = 1 (15-bit index)
				0, 0, 0, 1, // entry count = 1
				0, 1, // item ID = 1
				1,          // assoc count = 1
				0x00, 0x01, // property index = 1 (15-bit)
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // padding
			},
			box: &Box{
				Type:    "ipma",
				Size:    32,
				Offset:  0,
				Payload: 0,
			},
			properties: []PropertyEntry{
				{Index: 1, Type: "ispe"},
			},
			wantErr: false,
		},
		{
			name: "version 1 with 32-bit item ID",
			data: []byte{
				1, 0, 0, 0, // version 1, flags = 0
				0, 0, 0, 1, // entry count = 1
				0, 0, 0, 1, // item ID = 1 (32-bit)
				1,                                                    // assoc count = 1
				0x01,                                                 // property index = 1
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // padding
			},
			box: &Box{
				Type:    "ipma",
				Size:    32,
				Offset:  0,
				Payload: 0,
			},
			properties: []PropertyEntry{
				{Index: 1, Type: "ispe"},
			},
			wantErr: false,
		},
		{
			name:    "read error",
			data:    []byte{0, 0}, // Too short
			box:     &Box{Payload: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			index := &HeifIndex{
				Items: map[uint32]*HeifItem{
					1: {ItemID: 1},
				},
			}

			err := parseIpma(r, tt.box, index, tt.properties)

			if tt.wantErr {
				if err == nil {
					t.Error("parseIpma() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseIpma() error = %v", err)
			}
		})
	}
}

func TestParseInfe_UpdateExisting(t *testing.T) {
	r := bytes.NewReader([]byte{
		2, 0, 0, 0, // version 2
		0, 42, // item ID = 42
		0, 0, // protection index
		'E', 'x', 'i', 'f', // item type
		0, 0, 0, 0, // padding
	})

	infeBox := &Box{Type: "infe", Size: 20, Offset: 0, Payload: 0}
	parentBox := &Box{Size: 100, Offset: 0}
	index := &HeifIndex{
		Items: map[uint32]*HeifItem{
			42: {ItemID: 42, ItemType: "old"},
		},
	}

	err := parseInfe(r, infeBox, parentBox, index)
	if err != nil {
		t.Fatalf("parseInfe() error = %v", err)
	}

	if index.Items[42].ItemType != "Exif" {
		t.Errorf("ItemType = %v, want Exif", index.Items[42].ItemType)
	}
}

func TestParseIinf_InfeReadError(t *testing.T) {
	// Data that will fail when reading infe box header
	data := []byte{
		0, 0, 0, 0, // version 0 and flags
		0, 1, // item count = 1
		// infe box header starts here but truncated (less than 8 bytes)
		0, 0, 0, 12,
	}

	box := &Box{
		Type:    "iinf",
		Size:    18,
		Offset:  0,
		Payload: 0,
	}

	r := bytes.NewReader(data)
	index := &HeifIndex{Items: make(map[uint32]*HeifItem)}

	err := parseIinf(r, box, index)
	if err == nil {
		t.Error("parseIinf() expected error for infe read failure")
	}
}

func TestParseIloc_EntryReadError(t *testing.T) {
	// Data with valid header but entry read will fail
	data := []byte{
		0, 0, 0, 0, // version 0 and flags
		0x44, // offset_size=4, length_size=4
		0x00, // base_offset_size=0
		0, 1, // item count = 1
		// Entry data missing/truncated
	}

	box := &Box{
		Type:    "iloc",
		Size:    16,
		Offset:  0,
		Payload: 0,
	}

	r := bytes.NewReader(data)
	index := &HeifIndex{Items: make(map[uint32]*HeifItem)}

	err := parseIloc(r, box, index)
	if err == nil {
		t.Error("parseIloc() expected error for entry read failure")
	}
}

func TestParseIprp_IterateError(t *testing.T) {
	// ipco with invalid child box - this should return error
	data := []byte{
		// ipco box
		0, 0, 0, 20, 'i', 'p', 'c', 'o',
		// invalid child box (non-printable type)
		0, 0, 0, 12, 0x01, 0x02, 0x03, 0x04,
		0, 0, 0, 0,
	}

	box := &Box{
		Type:    "iprp",
		Size:    20,
		Offset:  0,
		Payload: 0,
	}

	r := bytes.NewReader(data)
	index := &HeifIndex{Items: make(map[uint32]*HeifItem)}

	// The iterate error gets propagated
	err := parseIprp(r, box, index)
	if err == nil {
		t.Error("parseIprp() expected error for invalid child box")
	}
}

func TestParseIloc_WithBaseOffset(t *testing.T) {
	// Test iloc with base_offset_size > 0
	data := append([]byte{
		0, 0, 0, 0, // version 0 and flags
		0x44, // offset_size=4, length_size=4
		0x40, // base_offset_size=4, index_size=0
		0, 1, // item count = 1
		0, 1, // item ID = 1
		0, 0, // data ref index
		0, 0, 0, 100, // base offset = 100
		0, 1, // extent count = 1
		0, 0, 0, 10, // extent offset = 10
		0, 0, 0, 20, // extent length = 20
	}, make([]byte, 64)...)

	box := &Box{
		Type:    "iloc",
		Size:    100,
		Offset:  0,
		Payload: 0,
	}

	r := bytes.NewReader(data)
	index := &HeifIndex{Items: make(map[uint32]*HeifItem)}

	err := parseIloc(r, box, index)
	if err != nil {
		t.Fatalf("parseIloc() error = %v", err)
	}

	item, exists := index.Items[1]
	if !exists {
		t.Fatal("item 1 not found")
	}

	if item.Location.BaseOffset != 100 {
		t.Errorf("BaseOffset = %v, want 100", item.Location.BaseOffset)
	}
}

func TestBuildHeifIndex_MissingIloc(t *testing.T) {
	// Create data with hdlr, iinf, but no iloc
	data := append([]byte{
		0, 0, 0, 0, // version/flags for meta
		// hdlr box
		0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0,
		// iinf box (minimal valid)
		0, 0, 0, 14, 'i', 'i', 'n', 'f',
		0, 0, 0, 0, // version 0
		0, 0, // count = 0
		// No iloc box
	}, make([]byte, 50)...)

	metaBox := &Box{
		Type:    "meta",
		Size:    uint64(len(data) + 8),
		Offset:  0,
		Payload: 0,
	}

	r := bytes.NewReader(data)
	_, err := buildHeifIndex(r, metaBox, 200)
	// Should error because iloc is required
	if err == nil {
		t.Error("buildHeifIndex() expected error for missing iloc")
	}
}

func TestBuildHeifIndex_IlocParseError(t *testing.T) {
	// Valid structure but iloc has invalid version that causes error
	data := append([]byte{
		0, 0, 0, 0, // version/flags for meta
		// hdlr box
		0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0,
		// iinf box
		0, 0, 0, 14, 'i', 'i', 'n', 'f',
		0, 0, 0, 0, // version 0
		0, 0, // count = 0
		// iloc box with item that will fail to parse
		0, 0, 0, 20, 'i', 'l', 'o', 'c',
		0, 0, 0, 0, // version 0
		0x44, 0x00, // sizes: offset=4, length=4, base=0, index=0
		0, 1, // item count = 1
		// item entry truncated - will cause read error
	}, make([]byte, 10)...)

	metaBox := &Box{
		Type:    "meta",
		Size:    uint64(len(data) + 8),
		Offset:  0,
		Payload: 0,
	}

	r := bytes.NewReader(data)
	_, err := buildHeifIndex(r, metaBox, 200)
	if err == nil {
		t.Error("buildHeifIndex() expected error for iloc parse failure")
	}
}

func TestBuildHeifIndex_WithIrefAndIprp(t *testing.T) {
	// Valid structure with iloc, iref, and iprp to cover more paths
	data := append([]byte{
		0, 0, 0, 0, // version/flags for meta
		// hdlr box
		0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0,
		// iinf box
		0, 0, 0, 14, 'i', 'i', 'n', 'f',
		0, 0, 0, 0, // version 0
		0, 0, // count = 0
		// iloc box (valid, empty)
		0, 0, 0, 16, 'i', 'l', 'o', 'c',
		0, 0, 0, 0, // version 0
		0x00, 0x00, // sizes
		0, 0, // item count = 0
		// iref box (valid but empty children)
		0, 0, 0, 12, 'i', 'r', 'e', 'f',
		0, 0, 0, 0, // version 0
		// iprp box with ipco and ipma
		0, 0, 0, 28, 'i', 'p', 'r', 'p',
		// ipco (empty)
		0, 0, 0, 8, 'i', 'p', 'c', 'o',
		// ipma (valid, empty)
		0, 0, 0, 12, 'i', 'p', 'm', 'a',
		0, 0, 0, 0, // version 0
		0, 0, 0, 0, // entry count = 0
	}, make([]byte, 20)...)

	metaBox := &Box{
		Type:    "meta",
		Size:    uint64(len(data) + 8),
		Offset:  0,
		Payload: 0,
	}

	r := bytes.NewReader(data)
	index, err := buildHeifIndex(r, metaBox, 200)
	if err != nil {
		t.Fatalf("buildHeifIndex() error = %v", err)
	}
	if index == nil {
		t.Error("buildHeifIndex() returned nil index")
	}
}

func TestBuildHeifIndex_IprpIterateError(t *testing.T) {
	// Structure where iprp's ipco has invalid child, causing iterateChildren error
	data := append([]byte{
		0, 0, 0, 0, // version/flags for meta
		// hdlr box
		0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0,
		// iinf box
		0, 0, 0, 14, 'i', 'i', 'n', 'f',
		0, 0, 0, 0, // version 0
		0, 0, // count = 0
		// iloc box (valid, empty)
		0, 0, 0, 16, 'i', 'l', 'o', 'c',
		0, 0, 0, 0, // version 0
		0x00, 0x00, // sizes
		0, 0, // item count = 0
		// iprp box
		0, 0, 0, 24, 'i', 'p', 'r', 'p',
		// ipco with invalid child box type (non-printable ASCII)
		0, 0, 0, 16, 'i', 'p', 'c', 'o',
		0, 0, 0, 8, 0x01, 0x02, 0x03, 0x04, // Invalid box type
	}, make([]byte, 20)...)

	metaBox := &Box{
		Type:    "meta",
		Size:    uint64(len(data) + 8),
		Offset:  0,
		Payload: 0,
	}

	r := bytes.NewReader(data)
	_, err := buildHeifIndex(r, metaBox, 200)
	if err == nil {
		t.Error("buildHeifIndex() expected error for iprp iterate failure")
	}
}

func TestBuildHeifIndex_IrefReadError(t *testing.T) {
	// Structure with iref that fails on initial read
	data := []byte{
		0, 0, 0, 0, // version/flags for meta
		// hdlr box
		0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0,
		// iinf box
		0, 0, 0, 14, 'i', 'i', 'n', 'f',
		0, 0, 0, 0, // version 0
		0, 0, // count = 0
		// iloc box (valid, empty)
		0, 0, 0, 16, 'i', 'l', 'o', 'c',
		0, 0, 0, 0, // version 0
		0x00, 0x00, // sizes
		0, 0, // item count = 0
		// iref box header (payload will fail to read)
		0, 0, 0, 20, 'i', 'r', 'e', 'f',
		// No payload data - will cause read error
	}

	metaBox := &Box{
		Type:    "meta",
		Size:    uint64(len(data) + 8),
		Offset:  0,
		Payload: 0,
	}

	// Use error reader that fails when reading iref payload
	r := &errorReaderAt{
		data:        data,
		errorOffset: 54, // Offset where iref payload starts
		customError: io.ErrUnexpectedEOF,
	}

	_, err := buildHeifIndex(r, metaBox, 200)
	if err == nil {
		t.Error("buildHeifIndex() expected error for iref read failure")
	}
}

func TestParseIinf_InfeParseError(t *testing.T) {
	// Valid iinf header but infe will fail to parse
	data := []byte{
		0, 0, 0, 0, // version 0 and flags
		0, 1, // item count = 1
		// infe box header
		0, 0, 0, 20, 'i', 'n', 'f', 'e',
		// infe payload too short
	}

	box := &Box{
		Type:    "iinf",
		Size:    26,
		Offset:  0,
		Payload: 0,
	}

	r := &errorReaderAt{
		data:        data,
		errorOffset: 14, // Fail reading infe payload
		customError: io.ErrUnexpectedEOF,
	}

	index := &HeifIndex{Items: make(map[uint32]*HeifItem)}
	err := parseIinf(r, box, index)
	if err == nil {
		t.Error("parseIinf() expected error for infe parse failure")
	}
}

func TestParseIprp_MissingIpma(t *testing.T) {
	// iprp with ipco but no ipma
	data := []byte{
		// ipco box (valid, empty)
		0, 0, 0, 8, 'i', 'p', 'c', 'o',
		// no ipma box follows
	}

	box := &Box{
		Type:    "iprp",
		Size:    16,
		Offset:  0,
		Payload: 0,
	}

	r := bytes.NewReader(data)
	index := &HeifIndex{Items: make(map[uint32]*HeifItem)}

	// Should not error - ipma is optional
	err := parseIprp(r, box, index)
	if err != nil {
		t.Errorf("parseIprp() unexpected error: %v", err)
	}
}

func TestBuildHeifIndex_ParseIinfError(t *testing.T) {
	// Valid structure but iinf parsing will fail
	data := append([]byte{
		0, 0, 0, 0, // version/flags for meta
		// hdlr box
		0, 0, 0, 12, 'h', 'd', 'l', 'r', 0, 0, 0, 0,
		// iinf box with bad content
		0, 0, 0, 14, 'i', 'i', 'n', 'f',
	}, make([]byte, 20)...)

	metaBox := &Box{
		Type:    "meta",
		Size:    50,
		Offset:  0,
		Payload: 0,
	}

	// Error reader that fails on iinf payload read
	r := &errorReaderAt{
		data:        data,
		errorOffset: 24, // Fail when reading iinf payload
		customError: io.ErrUnexpectedEOF,
	}

	_, err := buildHeifIndex(r, metaBox, 100)
	if err == nil {
		t.Error("buildHeifIndex() expected error for iinf parse failure")
	}
}
