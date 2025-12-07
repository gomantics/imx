package icc

import (
	"encoding/binary"
	"testing"
)

func TestParseTextType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"simple", []byte("Hello World\x00"), "Hello World"},
		{"no null", []byte("Hello World"), "Hello World"},
		{"empty", []byte{}, ""},
		{"only null", []byte{0}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTextType(tt.data)
			if got != tt.want {
				t.Errorf("parseTextType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTextDescriptionType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "simple",
			data: func() []byte {
				d := make([]byte, 20)
				binary.BigEndian.PutUint32(d[0:4], 12) // ASCII length
				copy(d[4:16], "Hello World\x00")
				return d
			}(),
			want: "Hello World",
		},
		{
			name: "short data",
			data: []byte{0, 0},
			want: "",
		},
		{
			name: "zero length",
			data: make([]byte, 8),
			want: "",
		},
		{
			name: "truncated",
			data: func() []byte {
				d := make([]byte, 8)
				binary.BigEndian.PutUint32(d[0:4], 100) // Length longer than data
				copy(d[4:8], "test")
				return d
			}(),
			want: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTextDescriptionType(tt.data)
			if got != tt.want {
				t.Errorf("parseTextDescriptionType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseMultiLocalizedUnicode(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "simple English",
			data: buildMLUC("enUS", "Hello World"),
			want: "Hello World",
		},
		{
			name: "short data",
			data: make([]byte, 10),
			want: "",
		},
		{
			name: "zero records",
			data: func() []byte {
				d := make([]byte, 20)
				copy(d[0:4], "mluc")
				return d // recordCount = 0
			}(),
			want: "",
		},
		{
			name: "invalid record size",
			data: func() []byte {
				d := make([]byte, 20)
				copy(d[0:4], "mluc")
				binary.BigEndian.PutUint32(d[8:12], 1)  // recordCount
				binary.BigEndian.PutUint32(d[12:16], 4) // recordSize too small
				return d
			}(),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMultiLocalizedUnicode(tt.data)
			if got != tt.want {
				t.Errorf("parseMultiLocalizedUnicode() = %q, want %q", got, tt.want)
			}
		})
	}
}

// buildMLUC creates a valid MLUC tag data with given language and text
func buildMLUC(langCountry, text string) []byte {
	// Calculate sizes
	utf16Bytes := make([]byte, len(text)*2+2) // +2 for null terminator
	for i, r := range text {
		binary.BigEndian.PutUint16(utf16Bytes[i*2:], uint16(r))
	}
	stringLen := uint32(len(utf16Bytes))
	stringOffset := uint32(28) // 8 (header) + 8 (counts) + 12 (record)

	data := make([]byte, int(stringOffset)+len(utf16Bytes))

	// Type signature
	copy(data[0:4], "mluc")
	// Reserved
	binary.BigEndian.PutUint32(data[4:8], 0)
	// Record count
	binary.BigEndian.PutUint32(data[8:12], 1)
	// Record size
	binary.BigEndian.PutUint32(data[12:16], 12)

	// Record
	copy(data[16:18], langCountry[0:2]) // language
	copy(data[18:20], langCountry[2:4]) // country
	binary.BigEndian.PutUint32(data[20:24], stringLen)
	binary.BigEndian.PutUint32(data[24:28], stringOffset)

	// String data
	copy(data[stringOffset:], utf16Bytes)

	return data
}

func TestDecodeUTF16BE(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "simple",
			data: []byte{0x00, 'H', 0x00, 'i'},
			want: "Hi",
		},
		{
			name: "with null terminator",
			data: []byte{0x00, 'H', 0x00, 'i', 0x00, 0x00},
			want: "Hi",
		},
		{
			name: "odd length",
			data: []byte{0x00, 'H', 0x00},
			want: "H",
		},
		{
			name: "empty",
			data: []byte{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeUTF16BE(tt.data)
			if got != tt.want {
				t.Errorf("decodeUTF16BE() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseXYZType(t *testing.T) {
	// Create XYZ data with one value
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[0:4], 0x00010000)  // X = 1.0
	binary.BigEndian.PutUint32(data[4:8], 0x00008000)  // Y = 0.5
	binary.BigEndian.PutUint32(data[8:12], 0x00004000) // Z = 0.25

	got := parseXYZType(data)

	if len(got) != 1 {
		t.Fatalf("parseXYZType() returned %d values, want 1", len(got))
	}
	if got[0].X != 1.0 {
		t.Errorf("X = %f, want 1.0", got[0].X)
	}
	if got[0].Y != 0.5 {
		t.Errorf("Y = %f, want 0.5", got[0].Y)
	}
	if got[0].Z != 0.25 {
		t.Errorf("Z = %f, want 0.25", got[0].Z)
	}
}

func TestParseXYZType_Empty(t *testing.T) {
	got := parseXYZType([]byte{})
	if got != nil {
		t.Errorf("parseXYZType([]) = %v, want nil", got)
	}
}

func TestParseCurveType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		isGamma  bool
		isLinear bool
		gamma    float64
		points   int
	}{
		{
			name: "identity (0 points)",
			data: func() []byte {
				d := make([]byte, 4)
				binary.BigEndian.PutUint32(d, 0)
				return d
			}(),
			isLinear: true,
			gamma:    1.0,
		},
		{
			name: "gamma 2.2",
			data: func() []byte {
				d := make([]byte, 6)
				binary.BigEndian.PutUint32(d[0:4], 1)
				binary.BigEndian.PutUint16(d[4:6], 0x0233) // ~2.2 in u8Fixed8
				return d
			}(),
			isGamma: true,
			gamma:   float64(0x0233) / 256.0,
		},
		{
			name: "curve with points",
			data: func() []byte {
				d := make([]byte, 12)
				binary.BigEndian.PutUint32(d[0:4], 4)        // 4 points
				binary.BigEndian.PutUint16(d[4:6], 0x0000)   // 0.0
				binary.BigEndian.PutUint16(d[6:8], 0x5555)   // ~0.33
				binary.BigEndian.PutUint16(d[8:10], 0xAAAA)  // ~0.67
				binary.BigEndian.PutUint16(d[10:12], 0xFFFF) // 1.0
				return d
			}(),
			points: 4,
		},
		{
			name:     "short data",
			data:     []byte{0, 0},
			isLinear: true,
			gamma:    1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCurveType(tt.data)

			if got.IsGamma != tt.isGamma {
				t.Errorf("IsGamma = %v, want %v", got.IsGamma, tt.isGamma)
			}
			if got.IsLinear != tt.isLinear {
				t.Errorf("IsLinear = %v, want %v", got.IsLinear, tt.isLinear)
			}
			if tt.isGamma || tt.isLinear {
				if got.Gamma != tt.gamma {
					t.Errorf("Gamma = %f, want %f", got.Gamma, tt.gamma)
				}
			}
			if tt.points > 0 && len(got.Points) != tt.points {
				t.Errorf("len(Points) = %d, want %d", len(got.Points), tt.points)
			}
		})
	}
}

func TestParseParametricCurveType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		funcType uint16
		gamma    float64
	}{
		{
			name: "type 0 (simple gamma)",
			data: func() []byte {
				d := make([]byte, 8)
				binary.BigEndian.PutUint16(d[0:2], 0)          // function type
				binary.BigEndian.PutUint32(d[4:8], 0x00024000) // gamma = 2.25
				return d
			}(),
			funcType: 0,
			gamma:    2.25,
		},
		{
			name:     "short data",
			data:     []byte{0, 0},
			funcType: 0,
			gamma:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseParametricCurveType(tt.data)

			if got.FunctionType != tt.funcType {
				t.Errorf("FunctionType = %d, want %d", got.FunctionType, tt.funcType)
			}
			if got.Gamma != tt.gamma {
				t.Errorf("Gamma = %f, want %f", got.Gamma, tt.gamma)
			}
		})
	}
}

func TestParseSignatureType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "technology - LCD",
			data: func() []byte {
				d := make([]byte, 4)
				binary.BigEndian.PutUint32(d, 0x4C434420) // 'LCD '
				return d
			}(),
			want: "LCD Display",
		},
		{
			name: "unknown signature",
			data: func() []byte {
				d := make([]byte, 4)
				copy(d, "test")
				return d
			}(),
			want: "test",
		},
		{
			name: "short data",
			data: []byte{0, 0},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSignatureType(tt.data)
			if got != tt.want {
				t.Errorf("parseSignatureType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseDateTimeType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "valid date",
			data: func() []byte {
				d := make([]byte, 12)
				binary.BigEndian.PutUint16(d[0:2], 2023)
				binary.BigEndian.PutUint16(d[2:4], 3)
				binary.BigEndian.PutUint16(d[4:6], 9)
				binary.BigEndian.PutUint16(d[6:8], 10)
				binary.BigEndian.PutUint16(d[8:10], 57)
				binary.BigEndian.PutUint16(d[10:12], 30)
				return d
			}(),
			want: "2023-03-09 10:57:30",
		},
		{
			name: "short data",
			data: make([]byte, 6),
			want: "",
		},
		{
			name: "invalid date",
			data: make([]byte, 12), // all zeros
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDateTimeType(tt.data)
			if got != tt.want {
				t.Errorf("parseDateTimeType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseMeasurementType(t *testing.T) {
	data := make([]byte, 36)
	binary.BigEndian.PutUint32(data[0:4], 1) // CIE 1931 observer
	// Backing XYZ (12 bytes)
	binary.BigEndian.PutUint32(data[4:8], 0x00010000)   // X = 1.0
	binary.BigEndian.PutUint32(data[8:12], 0x00010000)  // Y = 1.0
	binary.BigEndian.PutUint32(data[12:16], 0x00010000) // Z = 1.0
	binary.BigEndian.PutUint32(data[16:20], 1)          // 0/45 geometry
	binary.BigEndian.PutUint32(data[20:24], 0x00001000) // flare
	binary.BigEndian.PutUint32(data[24:28], 1)          // D50 illuminant

	got := parseMeasurementType(data)

	if got.Observer != "CIE 1931 (2°)" {
		t.Errorf("Observer = %q, want %q", got.Observer, "CIE 1931 (2°)")
	}
	if got.Geometry != "0/45 or 45/0" {
		t.Errorf("Geometry = %q, want %q", got.Geometry, "0/45 or 45/0")
	}
	if got.Illuminant != "D50" {
		t.Errorf("Illuminant = %q, want %q", got.Illuminant, "D50")
	}
}

func TestParseMeasurementType_Short(t *testing.T) {
	got := parseMeasurementType(make([]byte, 20))
	if got.Observer != "" {
		t.Error("parseMeasurementType() should return empty for short data")
	}
}

func TestParseViewingConditionsType(t *testing.T) {
	data := make([]byte, 28)
	// Illuminant XYZ
	binary.BigEndian.PutUint32(data[0:4], 0x00010000)
	binary.BigEndian.PutUint32(data[4:8], 0x00010000)
	binary.BigEndian.PutUint32(data[8:12], 0x00010000)
	// Surround XYZ
	binary.BigEndian.PutUint32(data[12:16], 0x00008000)
	binary.BigEndian.PutUint32(data[16:20], 0x00008000)
	binary.BigEndian.PutUint32(data[20:24], 0x00008000)
	// Illuminant type
	binary.BigEndian.PutUint32(data[24:28], 2) // D65

	got := parseViewingConditionsType(data)

	if got.IlluminantType != "D65" {
		t.Errorf("IlluminantType = %q, want %q", got.IlluminantType, "D65")
	}
	if got.IlluminantXYZ.X != 1.0 {
		t.Errorf("IlluminantXYZ.X = %f, want 1.0", got.IlluminantXYZ.X)
	}
}

func TestParseViewingConditionsType_Short(t *testing.T) {
	got := parseViewingConditionsType(make([]byte, 20))
	if got.IlluminantType != "" {
		t.Error("parseViewingConditionsType() should return empty for short data")
	}
}

func TestParseS15Fixed16ArrayType(t *testing.T) {
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[0:4], 0x00010000)  // 1.0
	binary.BigEndian.PutUint32(data[4:8], 0x00008000)  // 0.5
	binary.BigEndian.PutUint32(data[8:12], 0xFFFF0000) // -1.0

	got := parseS15Fixed16ArrayType(data)

	if len(got) != 3 {
		t.Fatalf("parseS15Fixed16ArrayType() returned %d values, want 3", len(got))
	}
	if got[0] != 1.0 {
		t.Errorf("got[0] = %f, want 1.0", got[0])
	}
	if got[1] != 0.5 {
		t.Errorf("got[1] = %f, want 0.5", got[1])
	}
	if got[2] != -1.0 {
		t.Errorf("got[2] = %f, want -1.0", got[2])
	}
}

func TestParseS15Fixed16ArrayType_Empty(t *testing.T) {
	got := parseS15Fixed16ArrayType([]byte{})
	if got != nil {
		t.Errorf("parseS15Fixed16ArrayType([]) = %v, want nil", got)
	}
}

func TestParseU16Fixed16ArrayType(t *testing.T) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], 0x00010000) // 1.0
	binary.BigEndian.PutUint32(data[4:8], 0x00020000) // 2.0

	got := parseU16Fixed16ArrayType(data)

	if len(got) != 2 {
		t.Fatalf("parseU16Fixed16ArrayType() returned %d values, want 2", len(got))
	}
	if got[0] != 1.0 {
		t.Errorf("got[0] = %f, want 1.0", got[0])
	}
	if got[1] != 2.0 {
		t.Errorf("got[1] = %f, want 2.0", got[1])
	}
}

func TestParseU16Fixed16ArrayType_Empty(t *testing.T) {
	got := parseU16Fixed16ArrayType([]byte{})
	if got != nil {
		t.Errorf("parseU16Fixed16ArrayType([]) = %v, want nil", got)
	}
}

func TestParseChromaticityType(t *testing.T) {
	data := make([]byte, 20)
	binary.BigEndian.PutUint16(data[0:2], 3) // 3 channels
	binary.BigEndian.PutUint16(data[2:4], 1) // ITU-R BT.709
	// Channel 0 coordinates
	binary.BigEndian.PutUint32(data[4:8], 0x0000A800)  // x
	binary.BigEndian.PutUint32(data[8:12], 0x00005400) // y

	got := parseChromaticityType(data)

	if got.Channels != 3 {
		t.Errorf("Channels = %d, want 3", got.Channels)
	}
	if got.Phosphor != "ITU-R BT.709" {
		t.Errorf("Phosphor = %q, want %q", got.Phosphor, "ITU-R BT.709")
	}
	if len(got.Coordinates) < 1 {
		t.Fatal("Expected at least one coordinate pair")
	}
}

func TestParseChromaticityType_Short(t *testing.T) {
	got := parseChromaticityType([]byte{0, 0})
	if got.Channels != 0 {
		t.Error("parseChromaticityType() should return empty for short data")
	}
}

func TestParseTagValue(t *testing.T) {
	// Build a profile with a text tag
	fullData := make([]byte, 200)
	copy(fullData[100:104], "text")      // type signature
	copy(fullData[108:120], "Hello\x00") // text value

	entry := TagEntry{
		Signature: "cprt",
		Offset:    100,
		Size:      20,
	}

	got := parseTagValue(nil, entry, fullData)

	if got.Signature != "cprt" {
		t.Errorf("Signature = %q, want %q", got.Signature, "cprt")
	}
	if got.TypeSig != "text" {
		t.Errorf("TypeSig = %q, want %q", got.TypeSig, "text")
	}
	if got.Value != "Hello" {
		t.Errorf("Value = %q, want %q", got.Value, "Hello")
	}
}

func TestParseTagValue_OutOfBounds(t *testing.T) {
	fullData := make([]byte, 50)
	entry := TagEntry{
		Signature: "test",
		Offset:    100, // Out of bounds
		Size:      20,
	}

	got := parseTagValue(nil, entry, fullData)

	if got.Value != nil {
		t.Errorf("parseTagValue() should return nil value for out-of-bounds offset")
	}
}

func TestParseTagValue_ShortData(t *testing.T) {
	fullData := make([]byte, 110)
	entry := TagEntry{
		Signature: "test",
		Offset:    100,
		Size:      5, // Less than 8 bytes for header
	}

	got := parseTagValue(nil, entry, fullData)

	if got.Value != nil {
		t.Errorf("parseTagValue() should return nil value for short tag data")
	}
}

func TestParseTagValue_UnknownType(t *testing.T) {
	fullData := make([]byte, 150)
	copy(fullData[100:104], "xxxx") // unknown type
	entry := TagEntry{
		Signature: "test",
		Offset:    100,
		Size:      20,
	}

	got := parseTagValue(nil, entry, fullData)

	// Unknown types return the data size
	if got.Value != 20 {
		t.Errorf("parseTagValue() for unknown type should return size, got %v", got.Value)
	}
}

func TestParseTagValue_AllTypes(t *testing.T) {
	tests := []struct {
		name    string
		typeSig string
		data    []byte
	}{
		{
			name:    "desc type",
			typeSig: "desc",
			data: func() []byte {
				d := make([]byte, 20)
				binary.BigEndian.PutUint32(d[0:4], 5)
				copy(d[4:9], "test\x00")
				return d
			}(),
		},
		{
			name:    "XYZ type",
			typeSig: "XYZ ",
			data: func() []byte {
				d := make([]byte, 12)
				binary.BigEndian.PutUint32(d[0:4], 0x00010000)
				binary.BigEndian.PutUint32(d[4:8], 0x00010000)
				binary.BigEndian.PutUint32(d[8:12], 0x00010000)
				return d
			}(),
		},
		{
			name:    "curv type",
			typeSig: "curv",
			data: func() []byte {
				d := make([]byte, 4)
				binary.BigEndian.PutUint32(d, 0)
				return d
			}(),
		},
		{
			name:    "para type",
			typeSig: "para",
			data: func() []byte {
				d := make([]byte, 8)
				binary.BigEndian.PutUint16(d[0:2], 0)
				binary.BigEndian.PutUint32(d[4:8], 0x00010000)
				return d
			}(),
		},
		{
			name:    "sig type",
			typeSig: "sig ",
			data: func() []byte {
				d := make([]byte, 4)
				copy(d, "test")
				return d
			}(),
		},
		{
			name:    "dtim type",
			typeSig: "dtim",
			data: func() []byte {
				d := make([]byte, 12)
				binary.BigEndian.PutUint16(d[0:2], 2023)
				binary.BigEndian.PutUint16(d[2:4], 1)
				binary.BigEndian.PutUint16(d[4:6], 1)
				return d
			}(),
		},
		{
			name:    "meas type",
			typeSig: "meas",
			data:    make([]byte, 36),
		},
		{
			name:    "view type",
			typeSig: "view",
			data:    make([]byte, 28),
		},
		{
			name:    "sf32 type",
			typeSig: "sf32",
			data: func() []byte {
				d := make([]byte, 4)
				binary.BigEndian.PutUint32(d, 0x00010000)
				return d
			}(),
		},
		{
			name:    "uf32 type",
			typeSig: "uf32",
			data: func() []byte {
				d := make([]byte, 4)
				binary.BigEndian.PutUint32(d, 0x00010000)
				return d
			}(),
		},
		{
			name:    "chrm type",
			typeSig: "chrm",
			data: func() []byte {
				d := make([]byte, 12)
				binary.BigEndian.PutUint16(d[0:2], 1)
				binary.BigEndian.PutUint16(d[2:4], 1)
				return d
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullData := make([]byte, 200)
			copy(fullData[100:104], tt.typeSig)
			copy(fullData[108:], tt.data)

			entry := TagEntry{
				Signature: "test",
				Offset:    100,
				Size:      uint32(8 + len(tt.data)),
			}

			got := parseTagValue(nil, entry, fullData)
			if got.TypeSig != tt.typeSig {
				t.Errorf("TypeSig = %q, want %q", got.TypeSig, tt.typeSig)
			}
		})
	}
}

func TestParseParametricCurveType_AllFunctionTypes(t *testing.T) {
	tests := []struct {
		name     string
		funcType uint16
		data     []byte
	}{
		{
			name:     "type 1",
			funcType: 1,
			data: func() []byte {
				d := make([]byte, 16)
				binary.BigEndian.PutUint16(d[0:2], 1)
				binary.BigEndian.PutUint32(d[4:8], 0x00024000)   // gamma
				binary.BigEndian.PutUint32(d[8:12], 0x00010000)  // a
				binary.BigEndian.PutUint32(d[12:16], 0x00008000) // b
				return d
			}(),
		},
		{
			name:     "type 2",
			funcType: 2,
			data: func() []byte {
				d := make([]byte, 20)
				binary.BigEndian.PutUint16(d[0:2], 2)
				binary.BigEndian.PutUint32(d[4:8], 0x00024000)   // gamma
				binary.BigEndian.PutUint32(d[8:12], 0x00010000)  // a
				binary.BigEndian.PutUint32(d[12:16], 0x00008000) // b
				binary.BigEndian.PutUint32(d[16:20], 0x00004000) // c
				return d
			}(),
		},
		{
			name:     "type 3",
			funcType: 3,
			data: func() []byte {
				d := make([]byte, 24)
				binary.BigEndian.PutUint16(d[0:2], 3)
				binary.BigEndian.PutUint32(d[4:8], 0x00024000)   // gamma
				binary.BigEndian.PutUint32(d[8:12], 0x00010000)  // a
				binary.BigEndian.PutUint32(d[12:16], 0x00008000) // b
				binary.BigEndian.PutUint32(d[16:20], 0x00004000) // c
				binary.BigEndian.PutUint32(d[20:24], 0x00002000) // d
				return d
			}(),
		},
		{
			name:     "type 4",
			funcType: 4,
			data: func() []byte {
				d := make([]byte, 32)
				binary.BigEndian.PutUint16(d[0:2], 4)
				binary.BigEndian.PutUint32(d[4:8], 0x00024000)   // gamma
				binary.BigEndian.PutUint32(d[8:12], 0x00010000)  // a
				binary.BigEndian.PutUint32(d[12:16], 0x00008000) // b
				binary.BigEndian.PutUint32(d[16:20], 0x00004000) // c
				binary.BigEndian.PutUint32(d[20:24], 0x00002000) // d
				binary.BigEndian.PutUint32(d[24:28], 0x00001000) // e
				binary.BigEndian.PutUint32(d[28:32], 0x00000800) // f
				return d
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseParametricCurveType(tt.data)
			if got.FunctionType != tt.funcType {
				t.Errorf("FunctionType = %d, want %d", got.FunctionType, tt.funcType)
			}
		})
	}
}

func TestParseMeasurementType_AllValues(t *testing.T) {
	tests := []struct {
		name       string
		observer   uint32
		geometry   uint32
		illuminant uint32
		wantObs    string
		wantGeom   string
		wantIllum  string
	}{
		{"obs1_geom1_D50", 1, 1, 1, "CIE 1931 (2°)", "0/45 or 45/0", "D50"},
		{"obs2_geom2_D65", 2, 2, 2, "CIE 1964 (10°)", "0/d or d/0", "D65"},
		{"unknown_D93", 99, 99, 3, "Unknown", "Unknown", "D93"},
		{"F2", 1, 1, 4, "CIE 1931 (2°)", "0/45 or 45/0", "F2"},
		{"D55", 1, 1, 5, "CIE 1931 (2°)", "0/45 or 45/0", "D55"},
		{"A", 1, 1, 6, "CIE 1931 (2°)", "0/45 or 45/0", "A"},
		{"E", 1, 1, 7, "CIE 1931 (2°)", "0/45 or 45/0", "E (Equi-Power)"},
		{"F8", 1, 1, 8, "CIE 1931 (2°)", "0/45 or 45/0", "F8"},
		{"unknown_illum", 1, 1, 99, "CIE 1931 (2°)", "0/45 or 45/0", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.wantIllum, func(t *testing.T) {
			data := make([]byte, 36)
			binary.BigEndian.PutUint32(data[0:4], tt.observer)
			binary.BigEndian.PutUint32(data[16:20], tt.geometry)
			binary.BigEndian.PutUint32(data[24:28], tt.illuminant)

			got := parseMeasurementType(data)
			if got.Observer != tt.wantObs {
				t.Errorf("Observer = %q, want %q", got.Observer, tt.wantObs)
			}
			if got.Geometry != tt.wantGeom {
				t.Errorf("Geometry = %q, want %q", got.Geometry, tt.wantGeom)
			}
			if got.Illuminant != tt.wantIllum {
				t.Errorf("Illuminant = %q, want %q", got.Illuminant, tt.wantIllum)
			}
		})
	}
}

func TestParseViewingConditionsType_AllIlluminants(t *testing.T) {
	illuminants := []struct {
		val  uint32
		want string
	}{
		{1, "D50"},
		{2, "D65"},
		{3, "D93"},
		{4, "F2"},
		{5, "D55"},
		{6, "A"},
		{7, "E (Equi-Power)"},
		{8, "F8"},
		{99, "Unknown"},
	}

	for _, tt := range illuminants {
		t.Run(tt.want, func(t *testing.T) {
			data := make([]byte, 28)
			binary.BigEndian.PutUint32(data[24:28], tt.val)

			got := parseViewingConditionsType(data)
			if got.IlluminantType != tt.want {
				t.Errorf("IlluminantType = %q, want %q", got.IlluminantType, tt.want)
			}
		})
	}
}

func TestParseChromaticityType_AllPhosphors(t *testing.T) {
	phosphors := []struct {
		val  uint16
		want string
	}{
		{1, "ITU-R BT.709"},
		{2, "SMPTE RP145-1994"},
		{3, "EBU Tech.3213-E"},
		{4, "P22"},
		{99, "Unknown"},
	}

	for _, tt := range phosphors {
		t.Run(tt.want, func(t *testing.T) {
			data := make([]byte, 12)
			binary.BigEndian.PutUint16(data[0:2], 1) // 1 channel
			binary.BigEndian.PutUint16(data[2:4], tt.val)

			got := parseChromaticityType(data)
			if got.Phosphor != tt.want {
				t.Errorf("Phosphor = %q, want %q", got.Phosphor, tt.want)
			}
		})
	}
}

func TestParseMultiLocalizedUnicode_NonEnglish(t *testing.T) {
	// Build MLUC with French only
	data := make([]byte, 40)
	copy(data[0:4], "mluc")
	binary.BigEndian.PutUint32(data[8:12], 1)   // 1 record
	binary.BigEndian.PutUint32(data[12:16], 12) // record size
	copy(data[16:18], "fr")                     // French
	copy(data[18:20], "FR")
	binary.BigEndian.PutUint32(data[20:24], 8)  // string length
	binary.BigEndian.PutUint32(data[24:28], 28) // string offset
	// "Test" in UTF-16BE
	binary.BigEndian.PutUint16(data[28:30], 'T')
	binary.BigEndian.PutUint16(data[30:32], 'e')
	binary.BigEndian.PutUint16(data[32:34], 's')
	binary.BigEndian.PutUint16(data[34:36], 't')

	got := parseMultiLocalizedUnicode(data)
	if got != "Test" {
		t.Errorf("parseMultiLocalizedUnicode() = %q, want %q", got, "Test")
	}
}

func TestParseMultiLocalizedUnicode_ZeroOffset(t *testing.T) {
	data := make([]byte, 28)
	copy(data[0:4], "mluc")
	binary.BigEndian.PutUint32(data[8:12], 1)   // 1 record
	binary.BigEndian.PutUint32(data[12:16], 12) // record size
	copy(data[16:18], "en")
	copy(data[18:20], "US")
	binary.BigEndian.PutUint32(data[20:24], 0) // zero length
	binary.BigEndian.PutUint32(data[24:28], 0) // zero offset

	got := parseMultiLocalizedUnicode(data)
	if got != "" {
		t.Errorf("parseMultiLocalizedUnicode() with zero offset = %q, want empty", got)
	}
}

func TestParseMultiLocalizedUnicode_OffsetOutOfBounds(t *testing.T) {
	data := make([]byte, 28)
	copy(data[0:4], "mluc")
	binary.BigEndian.PutUint32(data[8:12], 1)   // 1 record
	binary.BigEndian.PutUint32(data[12:16], 12) // record size
	copy(data[16:18], "en")
	copy(data[18:20], "US")
	binary.BigEndian.PutUint32(data[20:24], 100)  // length
	binary.BigEndian.PutUint32(data[24:28], 1000) // offset out of bounds

	got := parseMultiLocalizedUnicode(data)
	if got != "" {
		t.Errorf("parseMultiLocalizedUnicode() with out-of-bounds offset = %q, want empty", got)
	}
}

func TestParseCurveType_GammaShortData(t *testing.T) {
	data := make([]byte, 5) // count=1 but only 1 byte for value
	binary.BigEndian.PutUint32(data[0:4], 1)

	got := parseCurveType(data)
	if !got.IsGamma {
		t.Error("parseCurveType() should return IsGamma for count=1")
	}
	if got.Gamma != 1.0 {
		t.Errorf("Gamma = %f, want 1.0 for short data", got.Gamma)
	}
}
