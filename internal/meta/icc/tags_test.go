package icc

import (
	"encoding/binary"
	"testing"
)

func TestParseTagTable(t *testing.T) {
	// Build a profile with tag table
	data := make([]byte, 200)

	// Set profile header (minimal)
	binary.BigEndian.PutUint32(data[0:4], 200)            // profile size
	binary.BigEndian.PutUint32(data[36:40], ICCSignature) // 'acsp'

	// Tag count at offset 128
	binary.BigEndian.PutUint32(data[128:132], 3)

	// Tag 1: desc
	copy(data[132:136], "desc")
	binary.BigEndian.PutUint32(data[136:140], 160) // offset
	binary.BigEndian.PutUint32(data[140:144], 20)  // size

	// Tag 2: cprt
	copy(data[144:148], "cprt")
	binary.BigEndian.PutUint32(data[148:152], 180) // offset
	binary.BigEndian.PutUint32(data[152:156], 20)  // size

	// Tag 3: wtpt
	copy(data[156:160], "wtpt")
	binary.BigEndian.PutUint32(data[160:164], 200) // offset
	binary.BigEndian.PutUint32(data[164:168], 20)  // size

	tags, err := parseTagTable(data)
	if err != nil {
		t.Fatalf("parseTagTable() error = %v", err)
	}

	if len(tags) != 3 {
		t.Fatalf("parseTagTable() returned %d tags, want 3", len(tags))
	}

	if tags[0].Signature != "desc" {
		t.Errorf("tags[0].Signature = %q, want %q", tags[0].Signature, "desc")
	}
	if tags[0].Offset != 160 {
		t.Errorf("tags[0].Offset = %d, want 160", tags[0].Offset)
	}
	if tags[0].Size != 20 {
		t.Errorf("tags[0].Size = %d, want 20", tags[0].Size)
	}

	if tags[1].Signature != "cprt" {
		t.Errorf("tags[1].Signature = %q, want %q", tags[1].Signature, "cprt")
	}

	if tags[2].Signature != "wtpt" {
		t.Errorf("tags[2].Signature = %q, want %q", tags[2].Signature, "wtpt")
	}
}

func TestParseTagTable_TooShort(t *testing.T) {
	data := make([]byte, 128) // No room for tag count
	_, err := parseTagTable(data)
	if err == nil {
		t.Error("parseTagTable() expected error for short data")
	}
}

func TestParseTagTable_UnreasonableCount(t *testing.T) {
	data := make([]byte, 200)
	binary.BigEndian.PutUint32(data[128:132], 10000) // Unreasonable count

	_, err := parseTagTable(data)
	if err == nil {
		t.Error("parseTagTable() expected error for unreasonable tag count")
	}
}

func TestParseTagTable_ShortForEntries(t *testing.T) {
	data := make([]byte, 140)                    // Room for header + count + partial entry
	binary.BigEndian.PutUint32(data[128:132], 2) // 2 tags but not enough space

	_, err := parseTagTable(data)
	if err == nil {
		t.Error("parseTagTable() expected error when data too short for entries")
	}
}

func TestGetTagName(t *testing.T) {
	tests := []struct {
		sig  string
		want string
	}{
		{"desc", "ProfileDescription"},
		{"cprt", "ProfileCopyright"},
		{"wtpt", "MediaWhitePoint"},
		{"bkpt", "MediaBlackPoint"},
		{"chad", "ChromaticAdaptation"},
		{"rXYZ", "RedMatrixColumn"},
		{"gXYZ", "GreenMatrixColumn"},
		{"bXYZ", "BlueMatrixColumn"},
		{"rTRC", "RedToneReproductionCurve"},
		{"gTRC", "GreenToneReproductionCurve"},
		{"bTRC", "BlueToneReproductionCurve"},
		{"kTRC", "GrayToneReproductionCurve"},
		{"A2B0", "AToB0Perceptual"},
		{"B2A0", "BToA0Perceptual"},
		{"lumi", "Luminance"},
		{"meas", "Measurement"},
		{"tech", "Technology"},
		{"view", "ViewingConditions"},
		{"ncl2", "NamedColor2"},
		{"xxxx", "xxxx"}, // Unknown tag returns signature
	}

	for _, tt := range tests {
		t.Run(tt.sig, func(t *testing.T) {
			got := getTagName(tt.sig)
			if got != tt.want {
				t.Errorf("getTagName(%q) = %q, want %q", tt.sig, got, tt.want)
			}
		})
	}
}

func TestGetTechnologyName(t *testing.T) {
	tests := []struct {
		sig  uint32
		want string
	}{
		{0x66736E20, "Film Scanner"},
		{0x64636D20, "Digital Camera"},
		{0x7273636E, "Reflective Scanner"},
		{0x696A6574, "Ink Jet Printer"},
		{0x43525420, "Cathode Ray Tube Display"},
		{0x4F4C4544, "OLED Display"},
		{0x4C434420, "LCD Display"},
		{0x12345678, "\x124Vx"}, // Unknown returns signature string
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := getTechnologyName(tt.sig)
			if got != tt.want {
				t.Errorf("getTechnologyName(0x%08X) = %q, want %q", tt.sig, got, tt.want)
			}
		})
	}
}

func TestTagConstants(t *testing.T) {
	// Verify tag constant values match expected signatures
	if TagProfileDescription != "desc" {
		t.Errorf("TagProfileDescription = %q, want %q", TagProfileDescription, "desc")
	}
	if TagCopyright != "cprt" {
		t.Errorf("TagCopyright = %q, want %q", TagCopyright, "cprt")
	}
	if TagMediaWhitePoint != "wtpt" {
		t.Errorf("TagMediaWhitePoint = %q, want %q", TagMediaWhitePoint, "wtpt")
	}
	if TagRedColorant != "rXYZ" {
		t.Errorf("TagRedColorant = %q, want %q", TagRedColorant, "rXYZ")
	}
}

func TestTypeConstants(t *testing.T) {
	// Verify type constant values
	if TypeText != "text" {
		t.Errorf("TypeText = %q, want %q", TypeText, "text")
	}
	if TypeMLUC != "mluc" {
		t.Errorf("TypeMLUC = %q, want %q", TypeMLUC, "mluc")
	}
	if TypeXYZ != "XYZ " {
		t.Errorf("TypeXYZ = %q, want %q", TypeXYZ, "XYZ ")
	}
	if TypeCurve != "curv" {
		t.Errorf("TypeCurve = %q, want %q", TypeCurve, "curv")
	}
	if TypeParametricCurve != "para" {
		t.Errorf("TypeParametricCurve = %q, want %q", TypeParametricCurve, "para")
	}
}
