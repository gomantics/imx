package icc

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// TagRecord represents a tag table entry.
type TagRecord struct {
	Signature [4]byte
	Offset    uint32
	Size      uint32
}

// TagData represents parsed tag data.
type TagData struct {
	Signature string
	Type      string
	Value     any
}

// Helper functions for parsing ICC fixed-point numbers

// parseS15Fixed16 parses a 4-byte s15Fixed16 number using BigEndian.
func parseS15Fixed16(data []byte) float64 {
	if len(data) < 4 {
		return 0
	}
	val := int32(binary.BigEndian.Uint32(data[0:4]))
	return float64(val) / 65536.0
}

// parseU16Fixed16 parses a 4-byte u16Fixed16 number using BigEndian.
func parseU16Fixed16(data []byte) float64 {
	if len(data) < 4 {
		return 0
	}
	val := binary.BigEndian.Uint32(data[0:4])
	return float64(val) / 65536.0
}

// parseU8Fixed8 parses a 2-byte u8Fixed8 number using BigEndian.
func parseU8Fixed8(data []byte) float64 {
	if len(data) < 2 {
		return 0
	}
	val := binary.BigEndian.Uint16(data[0:2])
	return float64(val) / 256.0
}

// XYZNumber represents a CIE XYZ color value.
type XYZNumber struct {
	X float64
	Y float64
	Z float64
}

// parseXYZNumber parses a 12-byte XYZ value (3 s15Fixed16 numbers).
func parseXYZNumber(data []byte) XYZNumber {
	if len(data) < 12 {
		return XYZNumber{}
	}
	return XYZNumber{
		X: parseS15Fixed16(data[0:4]),
		Y: parseS15Fixed16(data[4:8]),
		Z: parseS15Fixed16(data[8:12]),
	}
}

// parseXYZType parses an XYZ type tag (one or more XYZ values).
func parseXYZType(r io.ReaderAt, tag TagRecord) ([]XYZNumber, error) {
	if tag.Size < 20 {
		return nil, fmt.Errorf("XYZ tag too small")
	}

	dataSize := tag.Size - 8
	count := dataSize / 12

	buf := make([]byte, count*12)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return nil, err
	}

	values := make([]XYZNumber, count)
	for i := uint32(0); i < count; i++ {
		values[i] = parseXYZNumber(buf[i*12:])
	}

	return values, nil
}

// CurveData represents parsed curve data.
type CurveData struct {
	IsGamma  bool      // If true, Gamma contains the gamma value
	IsLinear bool      // If true, curve is identity (1.0 gamma)
	Gamma    float64   // Gamma value if IsGamma
	Points   []float64 // Curve points if not gamma
}

// parseCurvType parses a curve type tag.
func parseCurvType(r io.ReaderAt, tag TagRecord) (CurveData, error) {
	if tag.Size < 12 {
		return CurveData{IsLinear: true, Gamma: 1.0}, fmt.Errorf("curv tag too small")
	}

	// Read count
	buf := make([]byte, 4)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return CurveData{IsLinear: true, Gamma: 1.0}, err
	}

	count := binary.BigEndian.Uint32(buf)
	if count == 0 {
		// Identity curve (gamma 1.0)
		return CurveData{IsLinear: true, Gamma: 1.0}, nil
	}

	if count == 1 {
		// Single value is u8Fixed8 gamma
		gammaBuf := make([]byte, 2)
		_, err = r.ReadAt(gammaBuf, int64(tag.Offset+12))
		if err != nil {
			return CurveData{IsGamma: true, Gamma: 1.0}, err
		}
		gamma := parseU8Fixed8(gammaBuf)
		return CurveData{IsGamma: true, Gamma: gamma}, nil
	}

	// Multiple points define a curve
	curveBuf := make([]byte, count*2)
	_, err = r.ReadAt(curveBuf, int64(tag.Offset+12))
	if err != nil {
		return CurveData{}, err
	}

	points := make([]float64, count)
	for i := uint32(0); i < count; i++ {
		offset := i * 2
		// Each point is a uint16 normalized to 0.0-1.0
		val := binary.BigEndian.Uint16(curveBuf[offset : offset+2])
		points[i] = float64(val) / 65535.0
	}

	return CurveData{Points: points}, nil
}

// ParametricCurveData represents a parametric curve.
type ParametricCurveData struct {
	FunctionType uint16
	Gamma        float64
	A, B, C, D   float64
	E, F         float64
}

// parseParametricCurveType parses a parametricCurveType tag.
func parseParametricCurveType(r io.ReaderAt, tag TagRecord) (ParametricCurveData, error) {
	if tag.Size < 12 {
		return ParametricCurveData{}, fmt.Errorf("para tag too small")
	}

	buf := make([]byte, tag.Size-8)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return ParametricCurveData{}, err
	}

	funcType := binary.BigEndian.Uint16(buf[0:2])
	// buf[2:4] is reserved

	curve := ParametricCurveData{FunctionType: funcType}

	// Parse parameters based on function type
	offset := 4
	switch funcType {
	case 0: // Y = X^g
		if len(buf) >= offset+4 {
			curve.Gamma = parseS15Fixed16(buf[offset:])
		}
	case 1: // Y = (aX+b)^g if X >= -b/a, else 0
		if len(buf) >= offset+12 {
			curve.Gamma = parseS15Fixed16(buf[offset:])
			curve.A = parseS15Fixed16(buf[offset+4:])
			curve.B = parseS15Fixed16(buf[offset+8:])
		}
	case 2: // Y = (aX+b)^g + c if X >= -b/a, else c
		if len(buf) >= offset+16 {
			curve.Gamma = parseS15Fixed16(buf[offset:])
			curve.A = parseS15Fixed16(buf[offset+4:])
			curve.B = parseS15Fixed16(buf[offset+8:])
			curve.C = parseS15Fixed16(buf[offset+12:])
		}
	case 3: // Y = (aX+b)^g if X >= d, else cX
		if len(buf) >= offset+20 {
			curve.Gamma = parseS15Fixed16(buf[offset:])
			curve.A = parseS15Fixed16(buf[offset+4:])
			curve.B = parseS15Fixed16(buf[offset+8:])
			curve.C = parseS15Fixed16(buf[offset+12:])
			curve.D = parseS15Fixed16(buf[offset+16:])
		}
	case 4: // Y = (aX+b)^g + e if X >= d, else cX + f
		if len(buf) >= offset+28 {
			curve.Gamma = parseS15Fixed16(buf[offset:])
			curve.A = parseS15Fixed16(buf[offset+4:])
			curve.B = parseS15Fixed16(buf[offset+8:])
			curve.C = parseS15Fixed16(buf[offset+12:])
			curve.D = parseS15Fixed16(buf[offset+16:])
			curve.E = parseS15Fixed16(buf[offset+20:])
			curve.F = parseS15Fixed16(buf[offset+24:])
		}
	}

	return curve, nil
}

// MeasurementData represents measurement conditions.
type MeasurementData struct {
	Observer   string
	Backing    XYZNumber
	Geometry   string
	Flare      float64
	Illuminant string
}

// parseMeasurementType parses a measurementType tag.
func parseMeasurementType(r io.ReaderAt, tag TagRecord) (MeasurementData, error) {
	if tag.Size < 44 {
		return MeasurementData{}, fmt.Errorf("meas tag too small")
	}

	buf := make([]byte, 36)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return MeasurementData{}, err
	}

	m := MeasurementData{}

	// Observer (standard observer)
	observer := binary.BigEndian.Uint32(buf[0:4])
	switch observer {
	case 1:
		m.Observer = "CIE1931TwoDegree"
	case 2:
		m.Observer = "CIE1964TenDegree"
	default:
		m.Observer = "Unknown"
	}

	// Backing XYZ
	m.Backing = parseXYZNumber(buf[4:16])

	// Geometry
	geometry := binary.BigEndian.Uint32(buf[16:20])
	switch geometry {
	case 1:
		m.Geometry = "0/45Or45/0"
	case 2:
		m.Geometry = "0/dOrd/0"
	default:
		m.Geometry = "Unknown"
	}

	// Flare
	m.Flare = parseU16Fixed16(buf[20:24])

	// Illuminant type
	illuminant := binary.BigEndian.Uint32(buf[24:28])
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
		m.Illuminant = "EquiPower"
	case 8:
		m.Illuminant = "F8"
	default:
		m.Illuminant = "Unknown"
	}

	return m, nil
}

// ViewingConditionsData represents viewing condition parameters.
type ViewingConditionsData struct {
	IlluminantXYZ  XYZNumber
	SurroundXYZ    XYZNumber
	IlluminantType string
}

// parseViewingConditionsType parses a viewingConditionsType tag.
func parseViewingConditionsType(r io.ReaderAt, tag TagRecord) (ViewingConditionsData, error) {
	if tag.Size < 36 {
		return ViewingConditionsData{}, fmt.Errorf("view tag too small")
	}

	buf := make([]byte, 28)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return ViewingConditionsData{}, err
	}

	v := ViewingConditionsData{}
	v.IlluminantXYZ = parseXYZNumber(buf[0:12])
	v.SurroundXYZ = parseXYZNumber(buf[12:24])

	illuminant := binary.BigEndian.Uint32(buf[24:28])
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
		v.IlluminantType = "EquiPower"
	case 8:
		v.IlluminantType = "F8"
	default:
		v.IlluminantType = "Unknown"
	}

	return v, nil
}

// ChromaticityData represents chromaticity coordinates.
type ChromaticityData struct {
	Channels    uint16
	Phosphor    string
	Coordinates [][2]float64 // [x, y] for each channel
}

// parseChromaticityType parses a chromaticityType tag.
func parseChromaticityType(r io.ReaderAt, tag TagRecord) (ChromaticityData, error) {
	if tag.Size < 12 {
		return ChromaticityData{}, fmt.Errorf("chrm tag too small")
	}

	buf := make([]byte, tag.Size-8)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return ChromaticityData{}, err
	}

	c := ChromaticityData{}
	c.Channels = binary.BigEndian.Uint16(buf[0:2])
	phosphor := binary.BigEndian.Uint16(buf[2:4])

	switch phosphor {
	case 1:
		c.Phosphor = "ITURBT709"
	case 2:
		c.Phosphor = "SMPTЕРP145-1994"
	case 3:
		c.Phosphor = "EBUTech3213-E"
	case 4:
		c.Phosphor = "P22"
	default:
		c.Phosphor = "Unknown"
	}

	// Parse chromaticity coordinates (u16Fixed16Number pairs)
	for i := uint16(0); i < c.Channels && int(4+i*8+8) <= len(buf); i++ {
		offset := 4 + int(i)*8
		x := parseU16Fixed16(buf[offset:])
		y := parseU16Fixed16(buf[offset+4:])
		c.Coordinates = append(c.Coordinates, [2]float64{x, y})
	}

	return c, nil
}

// Text and string parsers

// parseTextType parses a text type tag.
func parseTextType(r io.ReaderAt, tag TagRecord) (string, error) {
	if tag.Size <= 8 {
		return "", nil
	}

	textLen := tag.Size - 8
	buf := make([]byte, textLen)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return "", err
	}

	// Trim null bytes
	return strings.TrimRight(string(buf), "\x00"), nil
}

// parseDescType parses a description type tag (old style).
func parseDescType(r io.ReaderAt, tag TagRecord) (string, error) {
	if tag.Size < 12 {
		return "", fmt.Errorf("desc tag too small")
	}

	// Read ASCII description count
	buf := make([]byte, 4)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return "", err
	}

	count := binary.BigEndian.Uint32(buf)
	if count == 0 || count > tag.Size {
		return "", nil
	}

	// Read ASCII string
	strBuf := make([]byte, count)
	_, err = r.ReadAt(strBuf, int64(tag.Offset+12))
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(strBuf), "\x00"), nil
}

// parseMlucType parses a multi-localized Unicode type tag.
func parseMlucType(r io.ReaderAt, tag TagRecord) (string, error) {
	if tag.Size < 16 {
		return "", fmt.Errorf("mluc tag too small")
	}

	// Read number of records
	buf := make([]byte, 8)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return "", err
	}

	numRecords := binary.BigEndian.Uint32(buf[0:4])
	if numRecords == 0 {
		return "", nil
	}

	// Read first record (language code, country code, length, offset)
	recordBuf := make([]byte, 12)
	_, err = r.ReadAt(recordBuf, int64(tag.Offset+16))
	if err != nil {
		return "", err
	}

	length := binary.BigEndian.Uint32(recordBuf[4:8])
	offset := binary.BigEndian.Uint32(recordBuf[8:12])

	if length == 0 || length > tag.Size {
		return "", nil
	}

	// Read UTF-16 string
	strBuf := make([]byte, length)
	_, err = r.ReadAt(strBuf, int64(tag.Offset+offset))
	if err != nil {
		return "", err
	}

	// Convert UTF-16 BE to string (simplified)
	var result strings.Builder
	for i := 0; i < len(strBuf)-1; i += 2 {
		if strBuf[i] == 0 && strBuf[i+1] == 0 {
			break
		}
		char := binary.BigEndian.Uint16(strBuf[i : i+2])
		if char < 128 {
			result.WriteByte(byte(char))
		} else {
			result.WriteRune(rune(char))
		}
	}

	return result.String(), nil
}

// parseSigType parses a signature type tag.
func parseSigType(r io.ReaderAt, tag TagRecord) (string, error) {
	if tag.Size < 12 {
		return "", fmt.Errorf("sig tag too small")
	}

	buf := make([]byte, 4)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return "", err
	}

	// Check if it's a technology signature
	sig := binary.BigEndian.Uint32(buf)
	return getTechnologyName(sig), nil
}

// parseDateTimeType parses a dateTimeType tag.
func parseDateTimeType(r io.ReaderAt, tag TagRecord) (string, error) {
	if tag.Size < 20 {
		return "", fmt.Errorf("dtim tag too small")
	}

	buf := make([]byte, 12)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return "", err
	}

	year := binary.BigEndian.Uint16(buf[0:2])
	month := binary.BigEndian.Uint16(buf[2:4])
	day := binary.BigEndian.Uint16(buf[4:6])
	hour := binary.BigEndian.Uint16(buf[6:8])
	minute := binary.BigEndian.Uint16(buf[8:10])
	second := binary.BigEndian.Uint16(buf[10:12])

	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, minute, second), nil
}

// Array parsers

// parseS15Fixed16Type parses an s15Fixed16 array type.
func parseS15Fixed16Type(r io.ReaderAt, tag TagRecord) ([]float64, error) {
	if tag.Size < 8 {
		return nil, fmt.Errorf("sf32 tag too small")
	}

	count := (tag.Size - 8) / 4
	if count == 0 {
		return nil, nil
	}

	buf := make([]byte, count*4)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return nil, err
	}

	values := make([]float64, count)
	for i := uint32(0); i < count; i++ {
		values[i] = parseS15Fixed16(buf[i*4:])
	}

	return values, nil
}

// parseU16Fixed16ArrayType parses a u16Fixed16 array type.
func parseU16Fixed16ArrayType(r io.ReaderAt, tag TagRecord) ([]float64, error) {
	if tag.Size < 8 {
		return nil, fmt.Errorf("uf32 tag too small")
	}

	count := (tag.Size - 8) / 4
	if count == 0 {
		return nil, nil
	}

	buf := make([]byte, count*4)
	_, err := r.ReadAt(buf, int64(tag.Offset+8))
	if err != nil {
		return nil, err
	}

	values := make([]float64, count)
	for i := uint32(0); i < count; i++ {
		values[i] = parseU16Fixed16(buf[i*4:])
	}

	return values, nil
}
