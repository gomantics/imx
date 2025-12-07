package icc

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"

	"github.com/gomantics/imx/internal/common"
)

// ParsedTag represents a parsed tag value with its metadata
type ParsedTag struct {
	Signature string
	Name      string
	TypeSig   string
	Value     any
	Raw       []byte
}

// parseTagValue parses a tag's data based on its type signature
func parseTagValue(data []byte, entry TagEntry, fullData []byte) ParsedTag {
	tag := ParsedTag{
		Signature: entry.Signature,
		Name:      getTagName(entry.Signature),
	}

	// Bounds check
	if int(entry.Offset)+int(entry.Size) > len(fullData) {
		return tag
	}

	tagData := fullData[entry.Offset : entry.Offset+entry.Size]
	if len(tagData) < 8 {
		return tag
	}

	tag.Raw = tagData

	// First 4 bytes are the type signature
	tag.TypeSig = string(tagData[0:4])

	// Next 4 bytes are reserved (should be 0)
	// Actual value data starts at offset 8
	valueData := tagData[8:]

	switch tag.TypeSig {
	case TypeText:
		tag.Value = parseTextType(valueData)

	case TypeDesc:
		tag.Value = parseTextDescriptionType(valueData)

	case TypeMLUC:
		// MLUC string offsets are relative to start of tag data, not value data
		tag.Value = parseMultiLocalizedUnicode(tagData)

	case TypeXYZ:
		tag.Value = parseXYZType(valueData)

	case TypeCurve:
		tag.Value = parseCurveType(valueData)

	case TypeParametricCurve:
		tag.Value = parseParametricCurveType(valueData)

	case TypeSignature:
		tag.Value = parseSignatureType(valueData)

	case TypeDateTime:
		tag.Value = parseDateTimeType(valueData)

	case TypeMeasurement:
		tag.Value = parseMeasurementType(valueData)

	case TypeViewingConditions:
		tag.Value = parseViewingConditionsType(valueData)

	case TypeS15Fixed16Array:
		tag.Value = parseS15Fixed16ArrayType(valueData)

	case TypeU16Fixed16Array:
		tag.Value = parseU16Fixed16ArrayType(valueData)

	case TypeChromaticity:
		tag.Value = parseChromaticityType(valueData)

	default:
		// For unknown types, return raw data size
		tag.Value = len(tagData)
	}

	return tag
}

// parseTextType parses a textType tag (7-bit ASCII)
func parseTextType(data []byte) string {
	// Text ends at null byte or end of data
	end := bytes.IndexByte(data, 0)
	if end == -1 {
		end = len(data)
	}
	return string(data[:end])
}

// parseTextDescriptionType parses a textDescriptionType tag (v2)
func parseTextDescriptionType(data []byte) string {
	if len(data) < 4 {
		return ""
	}

	// ASCII count (including null terminator)
	asciiLen := binary.BigEndian.Uint32(data[0:4])
	if asciiLen == 0 {
		return ""
	}

	if len(data) < 4+int(asciiLen) {
		// Partial data
		asciiLen = uint32(len(data) - 4)
	}

	// ASCII string
	text := data[4 : 4+asciiLen]
	end := bytes.IndexByte(text, 0)
	if end == -1 {
		end = len(text)
	}

	return string(text[:end])
}

// parseMultiLocalizedUnicode parses a multiLocalizedUnicodeType tag (v4+)
// Note: data should be the full tag data including the 8-byte header (type sig + reserved)
func parseMultiLocalizedUnicode(data []byte) string {
	// MLUC structure:
	// Bytes 0-4: Type signature ('mluc')
	// Bytes 4-8: Reserved (zeros)
	// Bytes 8-12: Number of records
	// Bytes 12-16: Record size (should be 12)
	// Bytes 16+: Records
	// String offsets are relative to byte 0 of the tag

	if len(data) < 16 {
		return ""
	}

	// Number of records (at offset 8)
	recordCount := binary.BigEndian.Uint32(data[8:12])
	// Record size (should be 12, at offset 12)
	recordSize := binary.BigEndian.Uint32(data[12:16])

	if recordCount == 0 || recordSize < 12 {
		return ""
	}

	// Try to find English (en) first, otherwise use first record
	bestOffset := uint32(0)
	bestLength := uint32(0)
	foundEnglish := false

	// Records start at offset 16
	for i := uint32(0); i < recordCount && int(16+i*recordSize+12) <= len(data); i++ {
		recordStart := 16 + i*recordSize
		langCode := string(data[recordStart : recordStart+2])
		// countryCode := string(data[recordStart+2 : recordStart+4])
		strLength := binary.BigEndian.Uint32(data[recordStart+4 : recordStart+8])
		strOffset := binary.BigEndian.Uint32(data[recordStart+8 : recordStart+12])

		if i == 0 || (!foundEnglish && langCode == "en") {
			bestOffset = strOffset
			bestLength = strLength
			if langCode == "en" {
				foundEnglish = true
			}
		}
	}

	if bestOffset == 0 || bestLength == 0 {
		return ""
	}

	// String offset is relative to start of tag data
	if int(bestOffset+bestLength) > len(data) {
		return ""
	}

	return decodeUTF16BE(data[bestOffset : bestOffset+bestLength])
}

// decodeUTF16BE decodes a UTF-16 big-endian string
func decodeUTF16BE(data []byte) string {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}

	u16s := make([]uint16, len(data)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = binary.BigEndian.Uint16(data[i*2:])
	}

	// Remove null terminator if present
	for len(u16s) > 0 && u16s[len(u16s)-1] == 0 {
		u16s = u16s[:len(u16s)-1]
	}

	return string(utf16.Decode(u16s))
}

// parseXYZType parses an XYZType tag (one or more XYZ values)
func parseXYZType(data []byte) []XYZNumber {
	count := len(data) / 12
	if count == 0 {
		return nil
	}

	values := make([]XYZNumber, count)
	for i := 0; i < count; i++ {
		values[i] = parseXYZNumber(data[i*12:])
	}

	return values
}

// CurveData represents parsed curve data
type CurveData struct {
	IsGamma  bool      // If true, Values[0] is gamma value
	IsLinear bool      // If true, curve is identity (1.0 gamma)
	Gamma    float64   // Gamma value if IsGamma
	Points   []float64 // Curve points if not gamma
}

// parseCurveType parses a curveType tag
func parseCurveType(data []byte) CurveData {
	if len(data) < 4 {
		return CurveData{IsLinear: true, Gamma: 1.0}
	}

	pointCount := binary.BigEndian.Uint32(data[0:4])

	if pointCount == 0 {
		// Identity curve (gamma 1.0)
		return CurveData{IsLinear: true, Gamma: 1.0}
	}

	if pointCount == 1 {
		// Single value is u8Fixed8 gamma
		if len(data) < 6 {
			return CurveData{IsGamma: true, Gamma: 1.0}
		}
		gammaSlice, _ := common.SafeSlice(data, 4, 2)
		gamma, _ := common.ParseU8Fixed8(gammaSlice)
		return CurveData{IsGamma: true, Gamma: gamma}
	}

	// Multiple points define a curve
	points := make([]float64, 0, pointCount)
	for i := uint32(0); i < pointCount && int(4+i*2+2) <= len(data); i++ {
		// Each point is a uint16 normalized to 0.0-1.0
		val := binary.BigEndian.Uint16(data[4+i*2:])
		points = append(points, float64(val)/65535.0)
	}

	return CurveData{Points: points}
}

// ParametricCurveData represents a parametric curve
type ParametricCurveData struct {
	FunctionType uint16
	Gamma        float64
	A, B, C, D   float64
	E, F, G      float64
}

// parseParametricCurveType parses a parametricCurveType tag
func parseParametricCurveType(data []byte) ParametricCurveData {
	if len(data) < 4 {
		return ParametricCurveData{}
	}

	funcType := binary.BigEndian.Uint16(data[0:2])
	// data[2:4] is reserved

	curve := ParametricCurveData{FunctionType: funcType}

	// Parse parameters based on function type
	offset := 4
	switch funcType {
	case 0: // Y = X^g
		if len(data) >= offset+4 {
			s, _ := common.SafeSlice(data, offset, 4)
			curve.Gamma, _ = common.ParseS15Fixed16(s)
		}
	case 1: // Y = (aX+b)^g if X >= -b/a, else 0
		if len(data) >= offset+12 {
			s, _ := common.SafeSlice(data, offset, 4)
			curve.Gamma, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+4, 4)
			curve.A, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+8, 4)
			curve.B, _ = common.ParseS15Fixed16(s)
		}
	case 2: // Y = (aX+b)^g + c if X >= -b/a, else c
		if len(data) >= offset+16 {
			s, _ := common.SafeSlice(data, offset, 4)
			curve.Gamma, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+4, 4)
			curve.A, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+8, 4)
			curve.B, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+12, 4)
			curve.C, _ = common.ParseS15Fixed16(s)
		}
	case 3: // Y = (aX+b)^g if X >= d, else cX
		if len(data) >= offset+20 {
			s, _ := common.SafeSlice(data, offset, 4)
			curve.Gamma, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+4, 4)
			curve.A, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+8, 4)
			curve.B, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+12, 4)
			curve.C, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+16, 4)
			curve.D, _ = common.ParseS15Fixed16(s)
		}
	case 4: // Y = (aX+b)^g + e if X >= d, else cX + f
		if len(data) >= offset+28 {
			s, _ := common.SafeSlice(data, offset, 4)
			curve.Gamma, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+4, 4)
			curve.A, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+8, 4)
			curve.B, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+12, 4)
			curve.C, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+16, 4)
			curve.D, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+20, 4)
			curve.E, _ = common.ParseS15Fixed16(s)
			s, _ = common.SafeSlice(data, offset+24, 4)
			curve.F, _ = common.ParseS15Fixed16(s)
		}
	}

	return curve
}

// parseSignatureType parses a signatureType tag
func parseSignatureType(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	sig := binary.BigEndian.Uint32(data[0:4])

	// Check if it's a technology signature
	if name := getTechnologyName(sig); name != signatureToString(sig) {
		return name
	}

	return signatureToString(sig)
}

// parseDateTimeType parses a dateTimeType tag
func parseDateTimeType(data []byte) string {
	if len(data) < 12 {
		return ""
	}
	t := parseDateTimeNumber(data)
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// MeasurementData represents measurement conditions
type MeasurementData struct {
	Observer   string
	Backing    XYZNumber
	Geometry   string
	Flare      float64
	Illuminant string
}

// parseMeasurementType parses a measurementType tag
func parseMeasurementType(data []byte) MeasurementData {
	if len(data) < 36 {
		return MeasurementData{}
	}

	m := MeasurementData{}

	// Observer (standard observer)
	observer := binary.BigEndian.Uint32(data[0:4])
	switch observer {
	case 1:
		m.Observer = "CIE 1931 (2°)"
	case 2:
		m.Observer = "CIE 1964 (10°)"
	default:
		m.Observer = "Unknown"
	}

	// Backing XYZ
	m.Backing = parseXYZNumber(data[4:16])

	// Geometry
	geometry := binary.BigEndian.Uint32(data[16:20])
	switch geometry {
	case 1:
		m.Geometry = "0/45 or 45/0"
	case 2:
		m.Geometry = "0/d or d/0"
	default:
		m.Geometry = "Unknown"
	}

	// Flare
	flareSlice, _ := common.SafeSlice(data, 20, 4)
	m.Flare, _ = common.ParseU16Fixed16(flareSlice)

	// Illuminant type
	illuminant := binary.BigEndian.Uint32(data[24:28])
	switch illuminant {
	case 1:
		m.Illuminant = "D50"
	case 2:
		m.Illuminant = "D65"
	case 3:
		m.Illuminant = "D93"
	case 4:
		m.Illuminant = "F2"
	case 5:
		m.Illuminant = "D55"
	case 6:
		m.Illuminant = "A"
	case 7:
		m.Illuminant = "E (Equi-Power)"
	case 8:
		m.Illuminant = "F8"
	default:
		m.Illuminant = "Unknown"
	}

	return m
}

// ViewingConditionsData represents viewing condition parameters
type ViewingConditionsData struct {
	IlluminantXYZ  XYZNumber
	SurroundXYZ    XYZNumber
	IlluminantType string
}

// parseViewingConditionsType parses a viewingConditionsType tag
func parseViewingConditionsType(data []byte) ViewingConditionsData {
	if len(data) < 28 {
		return ViewingConditionsData{}
	}

	v := ViewingConditionsData{}
	v.IlluminantXYZ = parseXYZNumber(data[0:12])
	v.SurroundXYZ = parseXYZNumber(data[12:24])

	illuminant := binary.BigEndian.Uint32(data[24:28])
	switch illuminant {
	case 1:
		v.IlluminantType = "D50"
	case 2:
		v.IlluminantType = "D65"
	case 3:
		v.IlluminantType = "D93"
	case 4:
		v.IlluminantType = "F2"
	case 5:
		v.IlluminantType = "D55"
	case 6:
		v.IlluminantType = "A"
	case 7:
		v.IlluminantType = "E (Equi-Power)"
	case 8:
		v.IlluminantType = "F8"
	default:
		v.IlluminantType = "Unknown"
	}

	return v
}

// parseS15Fixed16ArrayType parses an s15Fixed16ArrayType tag
func parseS15Fixed16ArrayType(data []byte) []float64 {
	count := len(data) / 4
	if count == 0 {
		return nil
	}

	values := make([]float64, count)
	for i := 0; i < count; i++ {
		s, _ := common.SafeSlice(data, i*4, 4)
		values[i], _ = common.ParseS15Fixed16(s)
	}

	return values
}

// parseU16Fixed16ArrayType parses a u16Fixed16ArrayType tag
func parseU16Fixed16ArrayType(data []byte) []float64 {
	count := len(data) / 4
	if count == 0 {
		return nil
	}

	values := make([]float64, count)
	for i := 0; i < count; i++ {
		s, _ := common.SafeSlice(data, i*4, 4)
		values[i], _ = common.ParseU16Fixed16(s)
	}

	return values
}

// ChromaticityData represents chromaticity coordinates
type ChromaticityData struct {
	Channels    uint16
	Phosphor    string
	Coordinates [][2]float64 // [x, y] for each channel
}

// parseChromaticityType parses a chromaticityType tag
func parseChromaticityType(data []byte) ChromaticityData {
	if len(data) < 4 {
		return ChromaticityData{}
	}

	c := ChromaticityData{}
	c.Channels = binary.BigEndian.Uint16(data[0:2])
	phosphor := binary.BigEndian.Uint16(data[2:4])

	switch phosphor {
	case 1:
		c.Phosphor = "ITU-R BT.709"
	case 2:
		c.Phosphor = "SMPTE RP145-1994"
	case 3:
		c.Phosphor = "EBU Tech.3213-E"
	case 4:
		c.Phosphor = "P22"
	default:
		c.Phosphor = "Unknown"
	}

	// Parse chromaticity coordinates (u16Fixed16Number pairs)
	for i := uint16(0); i < c.Channels && int(4+i*8+8) <= len(data); i++ {
		offset := 4 + int(i)*8
		xSlice, _ := common.SafeSlice(data, offset, 4)
		x, _ := common.ParseU16Fixed16(xSlice)
		ySlice, _ := common.SafeSlice(data, offset+4, 4)
		y, _ := common.ParseU16Fixed16(ySlice)
		c.Coordinates = append(c.Coordinates, [2]float64{x, y})
	}

	return c
}
