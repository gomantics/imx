package common

import (
	"encoding/binary"
	"testing"
)

func TestByteParser(t *testing.T) {
	p := ByteParser{}

	// Single byte
	val, typ := p.Parse([]byte{42}, 1, binary.BigEndian)
	if val != 42 {
		t.Errorf("got %v, want 42", val)
	}
	if typ != "byte" {
		t.Errorf("got type %q, want byte", typ)
	}

	// Multiple bytes
	val, typ = p.Parse([]byte{1, 2, 3}, 3, binary.BigEndian)
	bytes, ok := val.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", val)
	}
	if len(bytes) != 3 || bytes[0] != 1 || bytes[1] != 2 || bytes[2] != 3 {
		t.Errorf("got %v, want [1 2 3]", bytes)
	}
	if typ != "bytes" {
		t.Errorf("got type %q, want bytes", typ)
	}
}

func TestASCIIParser(t *testing.T) {
	p := ASCIIParser{}

	// String with null terminator
	val, typ := p.Parse([]byte{'H', 'e', 'l', 'l', 'o', 0}, 6, binary.BigEndian)
	if val != "Hello" {
		t.Errorf("got %q, want %q", val, "Hello")
	}
	if typ != "string" {
		t.Errorf("got type %q, want string", typ)
	}

	// String without null
	val, typ = p.Parse([]byte{'T', 'e', 's', 't'}, 4, binary.BigEndian)
	if val != "Test" {
		t.Errorf("got %q, want %q", val, "Test")
	}
}

func TestShortParser(t *testing.T) {
	p := ShortParser{}

	// Single short (big-endian)
	data := []byte{0x12, 0x34}
	val, typ := p.Parse(data, 1, binary.BigEndian)
	if val != 0x1234 {
		t.Errorf("got %v, want 0x1234", val)
	}
	if typ != "short" {
		t.Errorf("got type %q, want short", typ)
	}

	// Single short (little-endian)
	val, typ = p.Parse(data, 1, binary.LittleEndian)
	if val != 0x3412 {
		t.Errorf("got %v, want 0x3412", val)
	}

	// Multiple shorts
	data = []byte{0x00, 0x01, 0x00, 0x02, 0x00, 0x03}
	val, typ = p.Parse(data, 3, binary.BigEndian)
	shorts, ok := val.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", val)
	}
	if len(shorts) != 3 || shorts[0] != 1 || shorts[1] != 2 || shorts[2] != 3 {
		t.Errorf("got %v, want [1 2 3]", shorts)
	}
	if typ != "shorts" {
		t.Errorf("got type %q, want shorts", typ)
	}
}

func TestLongParser(t *testing.T) {
	p := LongParser{}

	// Single long (big-endian)
	data := []byte{0x12, 0x34, 0x56, 0x78}
	val, typ := p.Parse(data, 1, binary.BigEndian)
	if val != 0x12345678 {
		t.Errorf("got %v, want 0x12345678", val)
	}
	if typ != "long" {
		t.Errorf("got type %q, want long", typ)
	}

	// Multiple longs
	data = []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02}
	val, typ = p.Parse(data, 2, binary.BigEndian)
	longs, ok := val.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", val)
	}
	if len(longs) != 2 || longs[0] != 1 || longs[1] != 2 {
		t.Errorf("got %v, want [1 2]", longs)
	}
	if typ != "longs" {
		t.Errorf("got type %q, want longs", typ)
	}
}

func TestRationalParser(t *testing.T) {
	p := RationalParser{}

	// Single rational: 100/10 = 10.0
	data := []byte{0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x0A}
	val, typ := p.Parse(data, 1, binary.BigEndian)
	if val != 10.0 {
		t.Errorf("got %v, want 10.0", val)
	}
	if typ != "rational" {
		t.Errorf("got type %q, want rational", typ)
	}

	// Rational with zero denominator
	data = []byte{0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x00}
	val, typ = p.Parse(data, 1, binary.BigEndian)
	if val != 0.0 {
		t.Errorf("got %v, want 0.0 for zero denominator", val)
	}

	// Multiple rationals
	data = []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, // 1/2 = 0.5
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, // 3/4 = 0.75
	}
	val, typ = p.Parse(data, 2, binary.BigEndian)
	rationals, ok := val.([]float64)
	if !ok {
		t.Fatalf("expected []float64, got %T", val)
	}
	if len(rationals) != 2 || rationals[0] != 0.5 || rationals[1] != 0.75 {
		t.Errorf("got %v, want [0.5 0.75]", rationals)
	}
	if typ != "rationals" {
		t.Errorf("got type %q, want rationals", typ)
	}

	// Multiple rationals with zero denominator in array
	data = []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, // 1/2 = 0.5
		0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00, // 10/0 = 0.0 (zero denom)
	}
	val, typ = p.Parse(data, 2, binary.BigEndian)
	rationals, ok = val.([]float64)
	if !ok {
		t.Fatalf("expected []float64, got %T", val)
	}
	if len(rationals) != 2 || rationals[0] != 0.5 || rationals[1] != 0.0 {
		t.Errorf("got %v, want [0.5 0.0]", rationals)
	}
}

func TestSByteParser(t *testing.T) {
	p := SByteParser{}

	// Single sbyte (-1)
	val, typ := p.Parse([]byte{0xFF}, 1, binary.BigEndian)
	if val != -1 {
		t.Errorf("got %v, want -1", val)
	}
	if typ != "sbyte" {
		t.Errorf("got type %q, want sbyte", typ)
	}

	// Multiple sbytes
	val, typ = p.Parse([]byte{0xFF, 0x00, 0x7F}, 3, binary.BigEndian)
	sbytes, ok := val.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", val)
	}
	if len(sbytes) != 3 || sbytes[0] != -1 || sbytes[1] != 0 || sbytes[2] != 127 {
		t.Errorf("got %v, want [-1 0 127]", sbytes)
	}
	if typ != "sbytes" {
		t.Errorf("got type %q, want sbytes", typ)
	}
}

func TestUndefinedParser(t *testing.T) {
	p := UndefinedParser{}

	// Raw bytes
	data := []byte{0x01, 0x02, 0x03, 0x04}
	val, typ := p.Parse(data, 4, binary.BigEndian)
	bytes, ok := val.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", val)
	}
	if len(bytes) != 4 || bytes[0] != 1 || bytes[3] != 4 {
		t.Errorf("got %v, want [1 2 3 4]", bytes)
	}
	if typ != "undefined" {
		t.Errorf("got type %q, want undefined", typ)
	}
}

func TestSShortParser(t *testing.T) {
	p := SShortParser{}

	// Single sshort (-1)
	data := []byte{0xFF, 0xFF}
	val, typ := p.Parse(data, 1, binary.BigEndian)
	if val != -1 {
		t.Errorf("got %v, want -1", val)
	}
	if typ != "sshort" {
		t.Errorf("got type %q, want sshort", typ)
	}

	// Multiple sshorts
	data = []byte{0xFF, 0xFF, 0x00, 0x00, 0x7F, 0xFF}
	val, typ = p.Parse(data, 3, binary.BigEndian)
	sshorts, ok := val.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", val)
	}
	if len(sshorts) != 3 || sshorts[0] != -1 || sshorts[1] != 0 || sshorts[2] != 32767 {
		t.Errorf("got %v, want [-1 0 32767]", sshorts)
	}
	if typ != "sshorts" {
		t.Errorf("got type %q, want sshorts", typ)
	}
}

func TestSLongParser(t *testing.T) {
	p := SLongParser{}

	// Single slong (-1)
	data := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	val, typ := p.Parse(data, 1, binary.BigEndian)
	if val != -1 {
		t.Errorf("got %v, want -1", val)
	}
	if typ != "slong" {
		t.Errorf("got type %q, want slong", typ)
	}

	// Multiple slongs
	data = []byte{
		0xFF, 0xFF, 0xFF, 0xFF, // -1
		0x00, 0x00, 0x00, 0x00, // 0
		0x7F, 0xFF, 0xFF, 0xFF, // 2147483647
	}
	val, typ = p.Parse(data, 3, binary.BigEndian)
	slongs, ok := val.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", val)
	}
	if len(slongs) != 3 || slongs[0] != -1 || slongs[1] != 0 || slongs[2] != 2147483647 {
		t.Errorf("got %v, want [-1 0 2147483647]", slongs)
	}
	if typ != "slongs" {
		t.Errorf("got type %q, want slongs", typ)
	}
}

func TestSRationalParser(t *testing.T) {
	p := SRationalParser{}

	// Single srational: -100/10 = -10.0
	data := []byte{
		0xFF, 0xFF, 0xFF, 0x9C, // -100 (as int32)
		0x00, 0x00, 0x00, 0x0A, // 10
	}
	val, typ := p.Parse(data, 1, binary.BigEndian)
	if val != -10.0 {
		t.Errorf("got %v, want -10.0", val)
	}
	if typ != "srational" {
		t.Errorf("got type %q, want srational", typ)
	}

	// Srational with zero denominator
	data = []byte{
		0xFF, 0xFF, 0xFF, 0x9C, // -100
		0x00, 0x00, 0x00, 0x00, // 0
	}
	val, typ = p.Parse(data, 1, binary.BigEndian)
	if val != 0.0 {
		t.Errorf("got %v, want 0.0 for zero denominator", val)
	}

	// Multiple srationals
	data = []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x02, // -1/2 = -0.5
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, // 3/4 = 0.75
	}
	val, typ = p.Parse(data, 2, binary.BigEndian)
	srationals, ok := val.([]float64)
	if !ok {
		t.Fatalf("expected []float64, got %T", val)
	}
	if len(srationals) != 2 || srationals[0] != -0.5 || srationals[1] != 0.75 {
		t.Errorf("got %v, want [-0.5 0.75]", srationals)
	}
	if typ != "srationals" {
		t.Errorf("got type %q, want srationals", typ)
	}

	// Multiple srationals with zero denominator in array
	data = []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x02, // -1/2 = -0.5
		0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00, // 10/0 = 0.0 (zero denom)
	}
	val, typ = p.Parse(data, 2, binary.BigEndian)
	srationals, ok = val.([]float64)
	if !ok {
		t.Fatalf("expected []float64, got %T", val)
	}
	if len(srationals) != 2 || srationals[0] != -0.5 || srationals[1] != 0.0 {
		t.Errorf("got %v, want [-0.5 0.0]", srationals)
	}
}

func TestTIFFTypeParsers_AllTypesRegistered(t *testing.T) {
	// Verify all 10 TIFF types are registered
	expectedTypes := []uint16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, typeID := range expectedTypes {
		if _, ok := TIFFTypeParsers[typeID]; !ok {
			t.Errorf("TIFF type %d not registered in TIFFTypeParsers", typeID)
		}
	}

	if len(TIFFTypeParsers) != 10 {
		t.Errorf("TIFFTypeParsers has %d parsers, want 10", len(TIFFTypeParsers))
	}
}

func TestTIFFTypeSizes_AllTypesRegistered(t *testing.T) {
	// Verify all 10 TIFF types have sizes defined
	expectedSizes := map[uint16]int{
		1:  1, // BYTE
		2:  1, // ASCII
		3:  2, // SHORT
		4:  4, // LONG
		5:  8, // RATIONAL
		6:  1, // SBYTE
		7:  1, // UNDEFINED
		8:  2, // SSHORT
		9:  4, // SLONG
		10: 8, // SRATIONAL
	}

	for typeID, expectedSize := range expectedSizes {
		size, ok := TIFFTypeSizes[typeID]
		if !ok {
			t.Errorf("TIFF type %d not in TIFFTypeSizes", typeID)
			continue
		}
		if size != expectedSize {
			t.Errorf("TIFF type %d size = %d, want %d", typeID, size, expectedSize)
		}
	}

	if len(TIFFTypeSizes) != 10 {
		t.Errorf("TIFFTypeSizes has %d entries, want 10", len(TIFFTypeSizes))
	}
}
