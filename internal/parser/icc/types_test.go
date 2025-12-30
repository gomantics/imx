package icc

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestParseS15Fixed16(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want float64
	}{
		{
			name: "positive value 1.0",
			data: []byte{0x00, 0x01, 0x00, 0x00},
			want: 1.0,
		},
		{
			name: "positive value 0.5",
			data: []byte{0x00, 0x00, 0x80, 0x00},
			want: 0.5,
		},
		{
			name: "zero",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			want: 0.0,
		},
		{
			name: "negative value -1.0",
			data: []byte{0xFF, 0xFF, 0x00, 0x00},
			want: -1.0,
		},
		{
			name: "too short",
			data: []byte{0x00, 0x01},
			want: 0,
		},
		{
			name: "empty",
			data: []byte{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseS15Fixed16(tt.data)
			if got != tt.want {
				t.Errorf("parseS15Fixed16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseU16Fixed16(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want float64
	}{
		{
			name: "value 1.0",
			data: []byte{0x00, 0x01, 0x00, 0x00},
			want: 1.0,
		},
		{
			name: "value 0.5",
			data: []byte{0x00, 0x00, 0x80, 0x00},
			want: 0.5,
		},
		{
			name: "zero",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			want: 0.0,
		},
		{
			name: "too short",
			data: []byte{0x00},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseU16Fixed16(tt.data)
			if got != tt.want {
				t.Errorf("parseU16Fixed16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseU8Fixed8(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want float64
	}{
		{
			name: "value 1.0",
			data: []byte{0x01, 0x00},
			want: 1.0,
		},
		{
			name: "value 2.2 (approx)",
			data: []byte{0x02, 0x33}, // 2 + 51/256 ≈ 2.199
			want: float64(0x0233) / 256.0,
		},
		{
			name: "zero",
			data: []byte{0x00, 0x00},
			want: 0.0,
		},
		{
			name: "too short",
			data: []byte{0x01},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseU8Fixed8(tt.data)
			if got != tt.want {
				t.Errorf("parseU8Fixed8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseXYZNumber(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want XYZNumber
	}{
		{
			name: "D50 illuminant",
			data: func() []byte {
				d := make([]byte, 12)
				// X = 0.9642, Y = 1.0, Z = 0.8249
				binary.BigEndian.PutUint32(d[0:4], 0x0000F6D6)  // ~0.9642
				binary.BigEndian.PutUint32(d[4:8], 0x00010000)  // 1.0
				binary.BigEndian.PutUint32(d[8:12], 0x0000D32D) // ~0.8249
				return d
			}(),
			want: XYZNumber{
				X: float64(0x0000F6D6) / 65536.0,
				Y: 1.0,
				Z: float64(0x0000D32D) / 65536.0,
			},
		},
		{
			name: "too short",
			data: make([]byte, 8),
			want: XYZNumber{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseXYZNumber(tt.data)
			if got != tt.want {
				t.Errorf("parseXYZNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper to build tag with type signature and data
func buildTagData(typeSig string, data []byte) ([]byte, TagRecord) {
	buf := make([]byte, 8+len(data))
	copy(buf[0:4], typeSig)
	copy(buf[8:], data)
	tag := TagRecord{
		Offset: 0,
		Size:   uint32(len(buf)),
	}
	return buf, tag
}

func TestParseXYZType(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		tag       TagRecord
		wantCount int
		wantErr   bool
	}{
		{
			name: "single XYZ value",
			data: func() []byte {
				d := make([]byte, 20) // 8 header + 12 XYZ
				copy(d[0:4], "XYZ ")
				binary.BigEndian.PutUint32(d[8:12], 0x00010000)  // X = 1.0
				binary.BigEndian.PutUint32(d[12:16], 0x00010000) // Y = 1.0
				binary.BigEndian.PutUint32(d[16:20], 0x00010000) // Z = 1.0
				return d
			}(),
			tag:       TagRecord{Offset: 0, Size: 20},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "tag too small",
			data:      make([]byte, 16),
			tag:       TagRecord{Offset: 0, Size: 16},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "zero count",
			data: func() []byte {
				d := make([]byte, 20)
				copy(d[0:4], "XYZ ")
				return d
			}(),
			tag:       TagRecord{Offset: 0, Size: 20},
			wantCount: 1, // dataSize=12, count=1
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			values, err := parseXYZType(r, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseXYZType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(values) != tt.wantCount {
				t.Errorf("parseXYZType() count = %d, want %d", len(values), tt.wantCount)
			}
		})
	}
}

func TestParseXYZType_ReadError(t *testing.T) {
	tag := TagRecord{Offset: 0, Size: 32}
	r := bytes.NewReader(make([]byte, 10)) // Too short
	_, err := parseXYZType(r, tag)
	if err == nil {
		t.Error("parseXYZType() expected error for read failure")
	}
}

func TestParseXYZType_ZeroCount(t *testing.T) {
	// Create XYZ tag with dataSize < 12 (resulting in count = 0)
	// Size must be >= 20 to pass size check, but dataSize = size - 8
	// For count = 0: dataSize / 12 = 0, so dataSize < 12, so size < 20
	// But size >= 20 is required. So the only way to get count = 0 is
	// if dataSize = 0..11, meaning size = 8..19, but size must be >= 20
	// Actually the check is size < 20, so we can't hit count = 0 through normal path
	// The count = 0 path is dead code unless dataSize < 12 with size >= 20
	// Let's test with size = 20 (dataSize = 12, count = 1) which will work
	d := make([]byte, 20)
	copy(d[0:4], "XYZ ")
	// This will have count = 1, not 0, but tests the edge case
	r := bytes.NewReader(d)
	values, err := parseXYZType(r, TagRecord{Offset: 0, Size: 20})
	if err != nil {
		t.Errorf("parseXYZType() error = %v", err)
	}
	if len(values) != 1 {
		t.Errorf("expected 1 value, got %d", len(values))
	}
}

func TestParseCurvType(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		tag        TagRecord
		wantLinear bool
		wantGamma  bool
		wantErr    bool
	}{
		{
			name: "identity curve (count=0)",
			data: func() []byte {
				d := make([]byte, 12)
				copy(d[0:4], "curv")
				binary.BigEndian.PutUint32(d[8:12], 0) // count = 0
				return d
			}(),
			tag:        TagRecord{Offset: 0, Size: 12},
			wantLinear: true,
			wantGamma:  false,
			wantErr:    false,
		},
		{
			name: "gamma curve (count=1)",
			data: func() []byte {
				d := make([]byte, 14)
				copy(d[0:4], "curv")
				binary.BigEndian.PutUint32(d[8:12], 1)       // count = 1
				binary.BigEndian.PutUint16(d[12:14], 0x0233) // gamma ~2.2
				return d
			}(),
			tag:        TagRecord{Offset: 0, Size: 14},
			wantLinear: false,
			wantGamma:  true,
			wantErr:    false,
		},
		{
			name: "curve with points (count>1)",
			data: func() []byte {
				d := make([]byte, 16)
				copy(d[0:4], "curv")
				binary.BigEndian.PutUint32(d[8:12], 2) // count = 2
				binary.BigEndian.PutUint16(d[12:14], 0x0000)
				binary.BigEndian.PutUint16(d[14:16], 0xFFFF)
				return d
			}(),
			tag:        TagRecord{Offset: 0, Size: 16},
			wantLinear: false,
			wantGamma:  false,
			wantErr:    false,
		},
		{
			name:       "tag too small",
			data:       make([]byte, 8),
			tag:        TagRecord{Offset: 0, Size: 8},
			wantLinear: true,
			wantGamma:  false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			curve, err := parseCurvType(r, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCurvType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if curve.IsLinear != tt.wantLinear {
				t.Errorf("parseCurvType() IsLinear = %v, want %v", curve.IsLinear, tt.wantLinear)
			}
			if curve.IsGamma != tt.wantGamma {
				t.Errorf("parseCurvType() IsGamma = %v, want %v", curve.IsGamma, tt.wantGamma)
			}
		})
	}
}

func TestParseCurvType_ReadErrors(t *testing.T) {
	t.Run("count read error", func(t *testing.T) {
		r := bytes.NewReader(make([]byte, 10))
		_, err := parseCurvType(r, TagRecord{Offset: 0, Size: 12})
		if err == nil {
			t.Error("expected error for count read failure")
		}
	})

	t.Run("gamma read error", func(t *testing.T) {
		d := make([]byte, 12)
		copy(d[0:4], "curv")
		binary.BigEndian.PutUint32(d[8:12], 1) // count = 1, but no gamma data
		r := bytes.NewReader(d)
		curve, err := parseCurvType(r, TagRecord{Offset: 0, Size: 14})
		if err == nil {
			t.Error("expected error for gamma read failure")
		}
		if !curve.IsGamma {
			t.Error("curve should have IsGamma set even on error")
		}
	})

	t.Run("points read error", func(t *testing.T) {
		d := make([]byte, 12)
		copy(d[0:4], "curv")
		binary.BigEndian.PutUint32(d[8:12], 5) // count = 5, but no points
		r := bytes.NewReader(d)
		_, err := parseCurvType(r, TagRecord{Offset: 0, Size: 22})
		if err == nil {
			t.Error("expected error for points read failure")
		}
	})
}

func TestParseParametricCurveType(t *testing.T) {
	tests := []struct {
		name     string
		funcType uint16
		dataLen  int
		wantErr  bool
	}{
		{"type 0 - gamma only", 0, 8, false},
		{"type 1 - gamma,a,b", 1, 16, false},
		{"type 2 - gamma,a,b,c", 2, 20, false},
		{"type 3 - gamma,a,b,c,d", 3, 24, false},
		{"type 4 - gamma,a,b,c,d,e,f", 4, 32, false},
		{"unknown type", 99, 8, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := make([]byte, 8+tt.dataLen)
			copy(d[0:4], "para")
			binary.BigEndian.PutUint16(d[8:10], tt.funcType)
			// Fill with some data
			for i := 12; i < len(d); i += 4 {
				binary.BigEndian.PutUint32(d[i:], 0x00010000) // 1.0
			}

			r := bytes.NewReader(d)
			curve, err := parseParametricCurveType(r, TagRecord{Offset: 0, Size: uint32(len(d))})
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParametricCurveType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if curve.FunctionType != tt.funcType {
				t.Errorf("FunctionType = %d, want %d", curve.FunctionType, tt.funcType)
			}
		})
	}
}

func TestParseParametricCurveType_Errors(t *testing.T) {
	t.Run("tag too small", func(t *testing.T) {
		_, err := parseParametricCurveType(bytes.NewReader(make([]byte, 8)), TagRecord{Offset: 0, Size: 8})
		if err == nil {
			t.Error("expected error for tag too small")
		}
	})

	t.Run("read error", func(t *testing.T) {
		_, err := parseParametricCurveType(bytes.NewReader(make([]byte, 10)), TagRecord{Offset: 0, Size: 16})
		if err == nil {
			t.Error("expected error for read failure")
		}
	})
}

func TestParseParametricCurveType_InsufficientData(t *testing.T) {
	// Test cases where funcType requires more data than available
	tests := []struct {
		name     string
		funcType uint16
		dataLen  int // Data after type signature (excluding 8 byte header)
	}{
		{"type0 insufficient", 0, 4},  // Needs 8 bytes (4 header + 4 gamma), only 4
		{"type1 insufficient", 1, 8},  // Needs 16 bytes, only 8
		{"type2 insufficient", 2, 12}, // Needs 20 bytes, only 12
		{"type3 insufficient", 3, 16}, // Needs 24 bytes, only 16
		{"type4 insufficient", 4, 20}, // Needs 32 bytes, only 20
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := make([]byte, 8+tt.dataLen)
			copy(d[0:4], "para")
			binary.BigEndian.PutUint16(d[8:10], tt.funcType)

			r := bytes.NewReader(d)
			curve, err := parseParametricCurveType(r, TagRecord{Offset: 0, Size: uint32(len(d))})
			if err != nil {
				t.Errorf("parseParametricCurveType() error = %v", err)
			}
			// Gamma should be 0 since data was insufficient
			if curve.Gamma != 0 {
				t.Errorf("Gamma = %v, want 0 (insufficient data)", curve.Gamma)
			}
		})
	}
}

func TestParseMeasurementType(t *testing.T) {
	buildMeasData := func(observer, geometry, illuminant uint32) []byte {
		d := make([]byte, 44)
		copy(d[0:4], "meas")
		binary.BigEndian.PutUint32(d[8:12], observer)
		// XYZ at 12-24
		binary.BigEndian.PutUint32(d[24:28], geometry)
		// Flare at 28-32
		binary.BigEndian.PutUint32(d[32:36], illuminant)
		return d
	}

	tests := []struct {
		observer   uint32
		geometry   uint32
		illuminant uint32
		wantObs    string
		wantGeom   string
		wantIllum  string
	}{
		{1, 1, 1, "CIE1931TwoDegree", "0/45Or45/0", "D50"},
		{2, 2, 2, "CIE1964TenDegree", "0/dOrd/0", "D65"},
		{0, 0, 3, "Unknown", "Unknown", "D93"},
		{0, 0, 4, "Unknown", "Unknown", "F2"},
		{0, 0, 5, "Unknown", "Unknown", "D55"},
		{0, 0, 6, "Unknown", "Unknown", "A"},
		{0, 0, 7, "Unknown", "Unknown", "EquiPower"},
		{0, 0, 8, "Unknown", "Unknown", "F8"},
		{0, 0, 99, "Unknown", "Unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.wantObs+"/"+tt.wantIllum, func(t *testing.T) {
			d := buildMeasData(tt.observer, tt.geometry, tt.illuminant)
			r := bytes.NewReader(d)
			m, err := parseMeasurementType(r, TagRecord{Offset: 0, Size: 44})
			if err != nil {
				t.Errorf("parseMeasurementType() error = %v", err)
			}
			if m.Observer != tt.wantObs {
				t.Errorf("Observer = %q, want %q", m.Observer, tt.wantObs)
			}
			if m.Geometry != tt.wantGeom {
				t.Errorf("Geometry = %q, want %q", m.Geometry, tt.wantGeom)
			}
			if m.Illuminant != tt.wantIllum {
				t.Errorf("Illuminant = %q, want %q", m.Illuminant, tt.wantIllum)
			}
		})
	}
}

func TestParseMeasurementType_Errors(t *testing.T) {
	t.Run("tag too small", func(t *testing.T) {
		_, err := parseMeasurementType(bytes.NewReader(make([]byte, 40)), TagRecord{Offset: 0, Size: 40})
		if err == nil {
			t.Error("expected error for tag too small")
		}
	})

	t.Run("read error", func(t *testing.T) {
		_, err := parseMeasurementType(bytes.NewReader(make([]byte, 10)), TagRecord{Offset: 0, Size: 44})
		if err == nil {
			t.Error("expected error for read failure")
		}
	})
}

func TestParseViewingConditionsType(t *testing.T) {
	buildViewData := func(illuminant uint32) []byte {
		d := make([]byte, 36)
		copy(d[0:4], "view")
		// XYZ illuminant 8-20
		// XYZ surround 20-32
		binary.BigEndian.PutUint32(d[32:36], illuminant)
		return d
	}

	tests := []struct {
		illuminant uint32
		want       string
	}{
		{1, "D50"},
		{2, "D65"},
		{3, "D93"},
		{4, "F2"},
		{5, "D55"},
		{6, "A"},
		{7, "EquiPower"},
		{8, "F8"},
		{99, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			d := buildViewData(tt.illuminant)
			r := bytes.NewReader(d)
			v, err := parseViewingConditionsType(r, TagRecord{Offset: 0, Size: 36})
			if err != nil {
				t.Errorf("parseViewingConditionsType() error = %v", err)
			}
			if v.IlluminantType != tt.want {
				t.Errorf("IlluminantType = %q, want %q", v.IlluminantType, tt.want)
			}
		})
	}
}

func TestParseViewingConditionsType_Errors(t *testing.T) {
	t.Run("tag too small", func(t *testing.T) {
		_, err := parseViewingConditionsType(bytes.NewReader(make([]byte, 32)), TagRecord{Offset: 0, Size: 32})
		if err == nil {
			t.Error("expected error for tag too small")
		}
	})

	t.Run("read error", func(t *testing.T) {
		_, err := parseViewingConditionsType(bytes.NewReader(make([]byte, 10)), TagRecord{Offset: 0, Size: 36})
		if err == nil {
			t.Error("expected error for read failure")
		}
	})
}

func TestParseChromaticityType(t *testing.T) {
	buildChrmData := func(channels uint16, phosphor uint16) []byte {
		d := make([]byte, 12+int(channels)*8)
		copy(d[0:4], "chrm")
		binary.BigEndian.PutUint16(d[8:10], channels)
		binary.BigEndian.PutUint16(d[10:12], phosphor)
		// Coordinates
		for i := 0; i < int(channels); i++ {
			offset := 12 + i*8
			binary.BigEndian.PutUint32(d[offset:], 0x00010000)   // x = 1.0
			binary.BigEndian.PutUint32(d[offset+4:], 0x00010000) // y = 1.0
		}
		return d
	}

	tests := []struct {
		phosphor uint16
		want     string
	}{
		{1, "ITURBT709"},
		{2, "SMPTЕРP145-1994"},
		{3, "EBUTech3213-E"},
		{4, "P22"},
		{99, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			d := buildChrmData(3, tt.phosphor)
			r := bytes.NewReader(d)
			c, err := parseChromaticityType(r, TagRecord{Offset: 0, Size: uint32(len(d))})
			if err != nil {
				t.Errorf("parseChromaticityType() error = %v", err)
			}
			if c.Phosphor != tt.want {
				t.Errorf("Phosphor = %q, want %q", c.Phosphor, tt.want)
			}
			if len(c.Coordinates) != 3 {
				t.Errorf("Coordinates count = %d, want 3", len(c.Coordinates))
			}
		})
	}
}

func TestParseChromaticityType_Errors(t *testing.T) {
	t.Run("tag too small", func(t *testing.T) {
		_, err := parseChromaticityType(bytes.NewReader(make([]byte, 8)), TagRecord{Offset: 0, Size: 8})
		if err == nil {
			t.Error("expected error for tag too small")
		}
	})

	t.Run("read error", func(t *testing.T) {
		_, err := parseChromaticityType(bytes.NewReader(make([]byte, 10)), TagRecord{Offset: 0, Size: 20})
		if err == nil {
			t.Error("expected error for read failure")
		}
	})
}

func TestParseChromaticityType_TruncatedCoordinates(t *testing.T) {
	// Create chromaticity with more channels than data
	d := make([]byte, 16)
	copy(d[0:4], "chrm")
	binary.BigEndian.PutUint16(d[8:10], 10) // 10 channels but not enough data
	binary.BigEndian.PutUint16(d[10:12], 1) // phosphor
	r := bytes.NewReader(d)
	c, err := parseChromaticityType(r, TagRecord{Offset: 0, Size: 16})
	if err != nil {
		t.Errorf("parseChromaticityType() error = %v", err)
	}
	// Should have 0 coordinates since there's not enough data
	if len(c.Coordinates) != 0 {
		t.Errorf("expected 0 coordinates, got %d", len(c.Coordinates))
	}
}

func TestParseChromaticityType_ZeroChannels(t *testing.T) {
	d := make([]byte, 12)
	copy(d[0:4], "chrm")
	binary.BigEndian.PutUint16(d[8:10], 0) // 0 channels
	binary.BigEndian.PutUint16(d[10:12], 1)
	r := bytes.NewReader(d)
	c, err := parseChromaticityType(r, TagRecord{Offset: 0, Size: 12})
	if err != nil {
		t.Errorf("parseChromaticityType() error = %v", err)
	}
	if len(c.Coordinates) != 0 {
		t.Errorf("expected 0 coordinates for 0 channels, got %d", len(c.Coordinates))
	}
}

func TestParseTextType(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		tag     TagRecord
		want    string
		wantErr bool
	}{
		{
			name: "valid text",
			data: func() []byte {
				d := make([]byte, 20)
				copy(d[0:4], "text")
				copy(d[8:], "Hello World\x00")
				return d
			}(),
			tag:     TagRecord{Offset: 0, Size: 20},
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "empty text (size <= 8)",
			data:    make([]byte, 8),
			tag:     TagRecord{Offset: 0, Size: 8},
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got, err := parseTextType(r, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTextType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseTextType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTextType_ReadError(t *testing.T) {
	r := bytes.NewReader(make([]byte, 10))
	_, err := parseTextType(r, TagRecord{Offset: 0, Size: 20})
	if err == nil {
		t.Error("expected error for read failure")
	}
}

func TestParseDescType(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		tag     TagRecord
		want    string
		wantErr bool
	}{
		{
			name: "valid desc",
			data: func() []byte {
				d := make([]byte, 30)
				copy(d[0:4], "desc")
				binary.BigEndian.PutUint32(d[8:12], 12) // count
				copy(d[12:], "Test String\x00")
				return d
			}(),
			tag:     TagRecord{Offset: 0, Size: 30},
			want:    "Test String",
			wantErr: false,
		},
		{
			name: "zero count",
			data: func() []byte {
				d := make([]byte, 16)
				copy(d[0:4], "desc")
				binary.BigEndian.PutUint32(d[8:12], 0)
				return d
			}(),
			tag:     TagRecord{Offset: 0, Size: 16},
			want:    "",
			wantErr: false,
		},
		{
			name: "count > size",
			data: func() []byte {
				d := make([]byte, 16)
				copy(d[0:4], "desc")
				binary.BigEndian.PutUint32(d[8:12], 1000)
				return d
			}(),
			tag:     TagRecord{Offset: 0, Size: 16},
			want:    "",
			wantErr: false,
		},
		{
			name:    "tag too small",
			data:    make([]byte, 8),
			tag:     TagRecord{Offset: 0, Size: 8},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got, err := parseDescType(r, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDescType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseDescType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseDescType_ReadErrors(t *testing.T) {
	t.Run("count read error", func(t *testing.T) {
		r := bytes.NewReader(make([]byte, 10))
		_, err := parseDescType(r, TagRecord{Offset: 0, Size: 16})
		if err == nil {
			t.Error("expected error for count read failure")
		}
	})

	t.Run("string read error", func(t *testing.T) {
		d := make([]byte, 12)
		copy(d[0:4], "desc")
		binary.BigEndian.PutUint32(d[8:12], 10) // count = 10 but no string data
		r := bytes.NewReader(d)
		_, err := parseDescType(r, TagRecord{Offset: 0, Size: 16})
		if err == nil {
			t.Error("expected error for string read failure")
		}
	})
}

func TestParseMlucType(t *testing.T) {
	buildMluc := func(text string) []byte {
		utf16 := make([]byte, len(text)*2)
		for i, c := range text {
			binary.BigEndian.PutUint16(utf16[i*2:], uint16(c))
		}

		d := make([]byte, 28+len(utf16))
		copy(d[0:4], "mluc")
		binary.BigEndian.PutUint32(d[8:12], 1)                   // numRecords
		binary.BigEndian.PutUint32(d[12:16], 12)                 // recordSize
		copy(d[16:18], "en")                                     // language
		copy(d[18:20], "US")                                     // country
		binary.BigEndian.PutUint32(d[20:24], uint32(len(utf16))) // length
		binary.BigEndian.PutUint32(d[24:28], 28)                 // offset
		copy(d[28:], utf16)
		return d
	}

	tests := []struct {
		name    string
		data    []byte
		tag     TagRecord
		want    string
		wantErr bool
	}{
		{
			name:    "valid mluc",
			data:    buildMluc("Test"),
			tag:     TagRecord{Offset: 0, Size: 36},
			want:    "Test",
			wantErr: false,
		},
		{
			name: "zero records",
			data: func() []byte {
				d := make([]byte, 16)
				copy(d[0:4], "mluc")
				binary.BigEndian.PutUint32(d[8:12], 0)
				return d
			}(),
			tag:     TagRecord{Offset: 0, Size: 16},
			want:    "",
			wantErr: false,
		},
		{
			name: "zero length",
			data: func() []byte {
				d := make([]byte, 28)
				copy(d[0:4], "mluc")
				binary.BigEndian.PutUint32(d[8:12], 1)
				binary.BigEndian.PutUint32(d[20:24], 0) // length = 0
				return d
			}(),
			tag:     TagRecord{Offset: 0, Size: 28},
			want:    "",
			wantErr: false,
		},
		{
			name: "length > size",
			data: func() []byte {
				d := make([]byte, 28)
				copy(d[0:4], "mluc")
				binary.BigEndian.PutUint32(d[8:12], 1)
				binary.BigEndian.PutUint32(d[20:24], 1000) // length > size
				return d
			}(),
			tag:     TagRecord{Offset: 0, Size: 28},
			want:    "",
			wantErr: false,
		},
		{
			name:    "tag too small",
			data:    make([]byte, 12),
			tag:     TagRecord{Offset: 0, Size: 12},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got, err := parseMlucType(r, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMlucType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseMlucType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseMlucType_ReadErrors(t *testing.T) {
	t.Run("records read error", func(t *testing.T) {
		r := bytes.NewReader(make([]byte, 10))
		_, err := parseMlucType(r, TagRecord{Offset: 0, Size: 20})
		if err == nil {
			t.Error("expected error for records read failure")
		}
	})

	t.Run("record read error", func(t *testing.T) {
		d := make([]byte, 16)
		copy(d[0:4], "mluc")
		binary.BigEndian.PutUint32(d[8:12], 1) // numRecords = 1
		r := bytes.NewReader(d)
		_, err := parseMlucType(r, TagRecord{Offset: 0, Size: 28})
		if err == nil {
			t.Error("expected error for record read failure")
		}
	})

	t.Run("string read error", func(t *testing.T) {
		d := make([]byte, 28)
		copy(d[0:4], "mluc")
		binary.BigEndian.PutUint32(d[8:12], 1)
		binary.BigEndian.PutUint32(d[20:24], 10) // length
		binary.BigEndian.PutUint32(d[24:28], 28) // offset
		r := bytes.NewReader(d)
		_, err := parseMlucType(r, TagRecord{Offset: 0, Size: 38})
		if err == nil {
			t.Error("expected error for string read failure")
		}
	})
}

func TestParseMlucType_NonASCII(t *testing.T) {
	// Build mluc with non-ASCII character
	d := make([]byte, 32)
	copy(d[0:4], "mluc")
	binary.BigEndian.PutUint32(d[8:12], 1)
	binary.BigEndian.PutUint32(d[20:24], 4)      // length = 4
	binary.BigEndian.PutUint32(d[24:28], 28)     // offset
	binary.BigEndian.PutUint16(d[28:30], 0x00C9) // É (non-ASCII)
	binary.BigEndian.PutUint16(d[30:32], 0x0000) // null terminator

	r := bytes.NewReader(d)
	got, err := parseMlucType(r, TagRecord{Offset: 0, Size: 32})
	if err != nil {
		t.Errorf("parseMlucType() error = %v", err)
	}
	if got != "É" {
		t.Errorf("parseMlucType() = %q, want 'É'", got)
	}
}

func TestParseSigType(t *testing.T) {
	tests := []struct {
		name    string
		sig     uint32
		want    string
		wantErr bool
	}{
		{"CRT", 0x43525420, "CathodeRayTubeDisplay", false},
		{"LCD", 0x4C434420, "LCDDisplay", false},
		{"unknown", 0x12345678, "\x12\x34\x56\x78", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := make([]byte, 12)
			copy(d[0:4], "sig ")
			binary.BigEndian.PutUint32(d[8:12], tt.sig)
			r := bytes.NewReader(d)
			got, err := parseSigType(r, TagRecord{Offset: 0, Size: 12})
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSigType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseSigType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSigType_Errors(t *testing.T) {
	t.Run("tag too small", func(t *testing.T) {
		_, err := parseSigType(bytes.NewReader(make([]byte, 8)), TagRecord{Offset: 0, Size: 8})
		if err == nil {
			t.Error("expected error for tag too small")
		}
	})

	t.Run("read error", func(t *testing.T) {
		_, err := parseSigType(bytes.NewReader(make([]byte, 10)), TagRecord{Offset: 0, Size: 12})
		if err == nil {
			t.Error("expected error for read failure")
		}
	})
}

func TestParseDateTimeType(t *testing.T) {
	buildDateTime := func(year, month, day, hour, minute, second uint16) []byte {
		d := make([]byte, 20)
		copy(d[0:4], "dtim")
		binary.BigEndian.PutUint16(d[8:10], year)
		binary.BigEndian.PutUint16(d[10:12], month)
		binary.BigEndian.PutUint16(d[12:14], day)
		binary.BigEndian.PutUint16(d[14:16], hour)
		binary.BigEndian.PutUint16(d[16:18], minute)
		binary.BigEndian.PutUint16(d[18:20], second)
		return d
	}

	t.Run("valid datetime", func(t *testing.T) {
		d := buildDateTime(2024, 1, 15, 12, 30, 45)
		r := bytes.NewReader(d)
		got, err := parseDateTimeType(r, TagRecord{Offset: 0, Size: 20})
		if err != nil {
			t.Errorf("parseDateTimeType() error = %v", err)
		}
		want := "2024-01-15 12:30:45"
		if got != want {
			t.Errorf("parseDateTimeType() = %q, want %q", got, want)
		}
	})

	t.Run("tag too small", func(t *testing.T) {
		_, err := parseDateTimeType(bytes.NewReader(make([]byte, 16)), TagRecord{Offset: 0, Size: 16})
		if err == nil {
			t.Error("expected error for tag too small")
		}
	})

	t.Run("read error", func(t *testing.T) {
		_, err := parseDateTimeType(bytes.NewReader(make([]byte, 10)), TagRecord{Offset: 0, Size: 20})
		if err == nil {
			t.Error("expected error for read failure")
		}
	})
}

func TestParseS15Fixed16Type(t *testing.T) {
	t.Run("valid array", func(t *testing.T) {
		d := make([]byte, 16) // 8 header + 8 data (2 values)
		copy(d[0:4], "sf32")
		binary.BigEndian.PutUint32(d[8:12], 0x00010000)  // 1.0
		binary.BigEndian.PutUint32(d[12:16], 0x00020000) // 2.0

		r := bytes.NewReader(d)
		values, err := parseS15Fixed16Type(r, TagRecord{Offset: 0, Size: 16})
		if err != nil {
			t.Errorf("parseS15Fixed16Type() error = %v", err)
		}
		if len(values) != 2 {
			t.Errorf("len = %d, want 2", len(values))
		}
	})

	t.Run("tag too small", func(t *testing.T) {
		_, err := parseS15Fixed16Type(bytes.NewReader(make([]byte, 4)), TagRecord{Offset: 0, Size: 4})
		if err == nil {
			t.Error("expected error for tag too small")
		}
	})

	t.Run("zero count", func(t *testing.T) {
		d := make([]byte, 8)
		copy(d[0:4], "sf32")
		r := bytes.NewReader(d)
		values, err := parseS15Fixed16Type(r, TagRecord{Offset: 0, Size: 8})
		if err != nil {
			t.Errorf("parseS15Fixed16Type() error = %v", err)
		}
		if values != nil {
			t.Errorf("expected nil for zero count")
		}
	})

	t.Run("read error", func(t *testing.T) {
		_, err := parseS15Fixed16Type(bytes.NewReader(make([]byte, 10)), TagRecord{Offset: 0, Size: 16})
		if err == nil {
			t.Error("expected error for read failure")
		}
	})
}

func TestParseU16Fixed16ArrayType(t *testing.T) {
	t.Run("valid array", func(t *testing.T) {
		d := make([]byte, 16)
		copy(d[0:4], "uf32")
		binary.BigEndian.PutUint32(d[8:12], 0x00010000)
		binary.BigEndian.PutUint32(d[12:16], 0x00020000)

		r := bytes.NewReader(d)
		values, err := parseU16Fixed16ArrayType(r, TagRecord{Offset: 0, Size: 16})
		if err != nil {
			t.Errorf("parseU16Fixed16ArrayType() error = %v", err)
		}
		if len(values) != 2 {
			t.Errorf("len = %d, want 2", len(values))
		}
	})

	t.Run("tag too small", func(t *testing.T) {
		_, err := parseU16Fixed16ArrayType(bytes.NewReader(make([]byte, 4)), TagRecord{Offset: 0, Size: 4})
		if err == nil {
			t.Error("expected error for tag too small")
		}
	})

	t.Run("zero count", func(t *testing.T) {
		d := make([]byte, 8)
		r := bytes.NewReader(d)
		values, err := parseU16Fixed16ArrayType(r, TagRecord{Offset: 0, Size: 8})
		if err != nil {
			t.Errorf("parseU16Fixed16ArrayType() error = %v", err)
		}
		if values != nil {
			t.Errorf("expected nil for zero count")
		}
	})

	t.Run("read error", func(t *testing.T) {
		_, err := parseU16Fixed16ArrayType(bytes.NewReader(make([]byte, 10)), TagRecord{Offset: 0, Size: 16})
		if err == nil {
			t.Error("expected error for read failure")
		}
	})
}
