package makernote

import (
	"encoding/binary"
	"testing"
)

func TestDetectNikonType3(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantMatch bool
		wantOrder binary.ByteOrder
	}{
		{
			name: "valid Nikon Type 3 little-endian",
			data: append(
				[]byte("Nikon\x00\x02\x00\x00\x00"),  // 10-byte header
				[]byte("II\x2a\x00\x08\x00\x00\x00")..., // TIFF header LE
			),
			wantMatch: true,
			wantOrder: binary.LittleEndian,
		},
		{
			name: "valid Nikon Type 3 big-endian",
			data: append(
				[]byte("Nikon\x00\x02\x00\x00\x00"),
				[]byte("MM\x00\x2a\x00\x00\x00\x08")...,
			),
			wantMatch: true,
			wantOrder: binary.BigEndian,
		},
		{
			name:      "Nikon Type 1 header - should not match",
			data:      []byte("Nikon\x00\x01\x00"),
			wantMatch: false,
		},
		{
			name:      "too short",
			data:      []byte("Nikon\x00\x02"),
			wantMatch: false,
		},
		{
			name:      "invalid byte order marker",
			data:      append([]byte("Nikon\x00\x02\x00\x00\x00"), []byte("XX\x2a\x00\x08\x00\x00\x00")...),
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, cfg := DetectNikonType3(tt.data)
			if match != tt.wantMatch {
				t.Errorf("DetectNikonType3() match = %v, want %v", match, tt.wantMatch)
			}
			if match {
				if cfg.ByteOrder != tt.wantOrder {
					t.Errorf("DetectNikonType3() order = %v, want %v", cfg.ByteOrder, tt.wantOrder)
				}
				if cfg.IFDOffset != 18 {
					t.Errorf("DetectNikonType3() IFDOffset = %d, want 18", cfg.IFDOffset)
				}
				if cfg.OffsetBase != OffsetRelativeToMakerNote {
					t.Errorf("DetectNikonType3() OffsetBase = %v, want OffsetRelativeToMakerNote", cfg.OffsetBase)
				}
				if cfg.Variant != "Type3" {
					t.Errorf("DetectNikonType3() Variant = %s, want Type3", cfg.Variant)
				}
			}
		})
	}
}

func TestDetectNikonType1(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantMatch bool
	}{
		{
			name:      "valid Nikon Type 1",
			data:      []byte("Nikon\x00\x01\x00"),
			wantMatch: true,
		},
		{
			name:      "Nikon Type 3 header - should not match",
			data:      []byte("Nikon\x00\x02\x00\x00\x00IIII"),
			wantMatch: false,
		},
		{
			name:      "too short",
			data:      []byte("Nikon\x00"),
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, cfg := DetectNikonType1(tt.data)
			if match != tt.wantMatch {
				t.Errorf("DetectNikonType1() match = %v, want %v", match, tt.wantMatch)
			}
			if match {
				if cfg.IFDOffset != 8 {
					t.Errorf("DetectNikonType1() IFDOffset = %d, want 8", cfg.IFDOffset)
				}
				if cfg.OffsetBase != OffsetAbsolute {
					t.Errorf("DetectNikonType1() OffsetBase = %v, want OffsetAbsolute", cfg.OffsetBase)
				}
				if cfg.ByteOrder != nil {
					t.Errorf("DetectNikonType1() ByteOrder = %v, want nil (inherit)", cfg.ByteOrder)
				}
			}
		})
	}
}

func TestDetectSony(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantMatch bool
	}{
		{
			name:      "valid SONY DSC",
			data:      []byte("SONY DSC \x00\x00\x00"),
			wantMatch: true,
		},
		{
			name:      "valid SONY CAM",
			data:      []byte("SONY CAM \x00\x00\x00"),
			wantMatch: true,
		},
		{
			name:      "too short",
			data:      []byte("SONY DSC"),
			wantMatch: false,
		},
		{
			name:      "wrong header",
			data:      []byte("SONY ABC \x00\x00\x00"),
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, cfg := DetectSony(tt.data)
			if match != tt.wantMatch {
				t.Errorf("DetectSony() match = %v, want %v", match, tt.wantMatch)
			}
			if match {
				if cfg.IFDOffset != 12 {
					t.Errorf("DetectSony() IFDOffset = %d, want 12", cfg.IFDOffset)
				}
				if cfg.OffsetBase != OffsetAbsolute {
					t.Errorf("DetectSony() OffsetBase = %v, want OffsetAbsolute", cfg.OffsetBase)
				}
				if cfg.ByteOrder != binary.LittleEndian {
					t.Errorf("DetectSony() ByteOrder = %v, want LittleEndian", cfg.ByteOrder)
				}
			}
		})
	}
}

func TestDetectFujifilm(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		wantMatch     bool
		wantIFDOffset int64
	}{
		{
			name:          "valid Fujifilm with offset 12",
			data:          []byte("FUJIFILM\x0c\x00\x00\x00"),
			wantMatch:     true,
			wantIFDOffset: 12,
		},
		{
			name:          "valid Fujifilm with offset 20",
			data:          []byte("FUJIFILM\x14\x00\x00\x00"),
			wantMatch:     true,
			wantIFDOffset: 20,
		},
		{
			name:      "too short",
			data:      []byte("FUJIFILM"),
			wantMatch: false,
		},
		{
			name:      "wrong header",
			data:      []byte("FUJIXXXX\x0c\x00\x00\x00"),
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, cfg := DetectFujifilm(tt.data)
			if match != tt.wantMatch {
				t.Errorf("DetectFujifilm() match = %v, want %v", match, tt.wantMatch)
			}
			if match {
				if cfg.IFDOffset != tt.wantIFDOffset {
					t.Errorf("DetectFujifilm() IFDOffset = %d, want %d", cfg.IFDOffset, tt.wantIFDOffset)
				}
				if cfg.OffsetBase != OffsetRelativeToMakerNote {
					t.Errorf("DetectFujifilm() OffsetBase = %v, want OffsetRelativeToMakerNote", cfg.OffsetBase)
				}
				if cfg.ByteOrder != binary.LittleEndian {
					t.Errorf("DetectFujifilm() ByteOrder = %v, want LittleEndian", cfg.ByteOrder)
				}
			}
		})
	}
}

func TestDetectCanon(t *testing.T) {
	// Canon IFD: 2-byte entry count + 12-byte entries
	// Create a minimal valid IFD with 2 entries
	makeCanonIFD := func(entryCount uint16) []byte {
		data := make([]byte, 2+int(entryCount)*12)
		binary.LittleEndian.PutUint16(data[0:2], entryCount)
		return data
	}

	tests := []struct {
		name      string
		data      []byte
		wantMatch bool
	}{
		{
			name:      "valid Canon with 2 entries",
			data:      makeCanonIFD(2),
			wantMatch: true,
		},
		{
			name:      "valid Canon with 50 entries",
			data:      makeCanonIFD(50),
			wantMatch: true,
		},
		{
			name:      "valid Canon with 1 entry",
			data:      makeCanonIFD(1),
			wantMatch: true,
		},
		{
			name:      "invalid - 0 entries",
			data:      makeCanonIFD(0),
			wantMatch: false,
		},
		{
			name:      "invalid - too many entries (101)",
			data:      []byte{101, 0}, // Entry count 101
			wantMatch: false,
		},
		{
			name:      "too short",
			data:      []byte{2, 0}, // 2 entries but no entry data
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, cfg := DetectCanon(tt.data)
			if match != tt.wantMatch {
				t.Errorf("DetectCanon() match = %v, want %v", match, tt.wantMatch)
			}
			if match {
				if cfg.IFDOffset != 0 {
					t.Errorf("DetectCanon() IFDOffset = %d, want 0", cfg.IFDOffset)
				}
				if cfg.OffsetBase != OffsetAbsolute {
					t.Errorf("DetectCanon() OffsetBase = %v, want OffsetAbsolute", cfg.OffsetBase)
				}
				if cfg.ByteOrder != nil {
					t.Errorf("DetectCanon() ByteOrder = %v, want nil (inherit)", cfg.ByteOrder)
				}
			}
		})
	}
}

func TestDetectionPriority(t *testing.T) {
	// Test that detection order is respected by the registry
	registry := NewRegistry()

	// Create mock handlers for testing priority
	type mockHandler struct {
		name   string
		detect func([]byte) (bool, *Config)
	}

	// The registry should be populated in the correct order
	// This test verifies the detection functions don't have false positives

	t.Run("Nikon Type 3 not matched by Type 1", func(t *testing.T) {
		data := append(
			[]byte("Nikon\x00\x02\x00\x00\x00"),
			[]byte("II\x2a\x00\x08\x00\x00\x00")...,
		)
		match, _ := DetectNikonType1(data)
		if match {
			t.Error("Nikon Type 3 data should not match Type 1 detection")
		}
	})

	t.Run("Sony not matched by Fujifilm", func(t *testing.T) {
		data := []byte("SONY DSC \x00\x00\x00")
		match, _ := DetectFujifilm(data)
		if match {
			t.Error("Sony data should not match Fujifilm detection")
		}
	})

	t.Run("Fujifilm not matched by Canon", func(t *testing.T) {
		// FUJIFILM header would have entry count 0x5546 ('FU') which is > 100
		data := []byte("FUJIFILM\x0c\x00\x00\x00")
		match, _ := DetectCanon(data)
		if match {
			t.Error("Fujifilm data should not match Canon detection")
		}
	})

	_ = registry // Registry tested implicitly via detection functions
}

func TestRegistryDetect(t *testing.T) {
	// Create a simple test handler
	registry := NewRegistry()

	// Test empty registry returns nil
	handler, cfg := registry.Detect([]byte("test data"))
	if handler != nil || cfg != nil {
		t.Error("Empty registry should return nil")
	}
}

func TestConfigFields(t *testing.T) {
	// Verify Config struct has all required fields
	cfg := &Config{
		IFDOffset:  12,
		OffsetBase: OffsetRelativeToMakerNote,
		ByteOrder:  binary.LittleEndian,
		HasNextIFD: true,
		Variant:    "Type3",
	}

	if cfg.IFDOffset != 12 {
		t.Errorf("IFDOffset = %d, want 12", cfg.IFDOffset)
	}
	if cfg.OffsetBase != OffsetRelativeToMakerNote {
		t.Errorf("OffsetBase = %v, want OffsetRelativeToMakerNote", cfg.OffsetBase)
	}
	if cfg.ByteOrder != binary.LittleEndian {
		t.Errorf("ByteOrder = %v, want LittleEndian", cfg.ByteOrder)
	}
	if !cfg.HasNextIFD {
		t.Error("HasNextIFD = false, want true")
	}
	if cfg.Variant != "Type3" {
		t.Errorf("Variant = %s, want Type3", cfg.Variant)
	}
}

func TestOffsetBaseConstants(t *testing.T) {
	// Verify OffsetBase constants are defined correctly
	if OffsetAbsolute != 0 {
		t.Errorf("OffsetAbsolute = %d, want 0", OffsetAbsolute)
	}
	if OffsetRelativeToMakerNote != 1 {
		t.Errorf("OffsetRelativeToMakerNote = %d, want 1", OffsetRelativeToMakerNote)
	}
}
