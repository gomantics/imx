package testing

import "testing"

// TestValuesEqual tests the ValuesEqual function with various type combinations
func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name string
		got  interface{}
		want interface{}
		eq   bool
	}{
		{name: "identical strings", got: "hello", want: "hello", eq: true},
		{name: "different strings", got: "hello", want: "world", eq: false},
		{name: "identical uint16", got: uint16(100), want: uint16(100), eq: true},
		{name: "different uint16", got: uint16(100), want: uint16(200), eq: false},
		{name: "identical uint32", got: uint32(1000), want: uint32(1000), eq: true},
		{name: "different uint32", got: uint32(1000), want: uint32(2000), eq: false},
		{name: "identical int", got: int(42), want: int(42), eq: true},
		{name: "different int", got: int(42), want: int(84), eq: false},
		{name: "int vs uint32 same value", got: uint32(42), want: int(42), eq: true},
		{name: "int vs uint32 different value", got: uint32(42), want: int(84), eq: false},
		{name: "identical int64", got: int64(123456), want: int64(123456), eq: true},
		{name: "different int64", got: int64(123456), want: int64(654321), eq: false},
		{name: "different types no conversion", got: "100", want: uint16(100), eq: false},
		{name: "identical uint8", got: uint8(25), want: uint8(25), eq: true},
		{name: "different uint8", got: uint8(25), want: uint8(50), eq: false},
		{name: "identical uint64", got: uint64(123456789), want: uint64(123456789), eq: true},
		{name: "different uint64", got: uint64(123456789), want: uint64(987654321), eq: false},
		{name: "bool true equals", got: true, want: true, eq: true},
		{name: "bool false equals", got: false, want: false, eq: true},
		{name: "bool different", got: true, want: false, eq: false},
		{name: "[]uint16 identical", got: []uint16{1, 2, 3}, want: []uint16{1, 2, 3}, eq: true},
		{name: "[]uint16 different", got: []uint16{1, 2, 3}, want: []uint16{1, 2, 4}, eq: false},
		{name: "[]uint16 different length", got: []uint16{1, 2}, want: []uint16{1, 2, 3}, eq: false},
		{name: "[]string identical", got: []string{"a", "b"}, want: []string{"a", "b"}, eq: true},
		{name: "[]string different", got: []string{"a", "b"}, want: []string{"a", "c"}, eq: false},
		{name: "[]byte identical", got: []byte{1, 2, 3}, want: []byte{1, 2, 3}, eq: true},
		{name: "[]byte different", got: []byte{1, 2, 3}, want: []byte{1, 2, 4}, eq: false},
		{name: "float64 identical", got: float64(3.14159), want: float64(3.14159), eq: true},
		{name: "float64 different", got: float64(3.14159), want: float64(2.71828), eq: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValuesEqual(tt.got, tt.want)
			if got != tt.eq {
				t.Errorf("ValuesEqual(%v, %v) = %v, want %v", tt.got, tt.want, got, tt.eq)
			}
		})
	}
}

// TestTypeMatches tests the TypeMatches function with various types
func TestTypeMatches(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		typeName string
		matches  bool
	}{
		{name: "string matches", value: "hello", typeName: "string", matches: true},
		{name: "string mismatch", value: "hello", typeName: "int", matches: false},
		{name: "uint16 matches", value: uint16(100), typeName: "uint16", matches: true},
		{name: "uint16 mismatch", value: uint16(100), typeName: "uint32", matches: false},
		{name: "uint32 matches", value: uint32(1000), typeName: "uint32", matches: true},
		{name: "uint32 mismatch", value: uint32(1000), typeName: "string", matches: false},
		{name: "int matches", value: int(42), typeName: "int", matches: true},
		{name: "int mismatch", value: int(42), typeName: "int64", matches: false},
		{name: "int64 matches", value: int64(123456), typeName: "int64", matches: true},
		{name: "int64 mismatch", value: int64(123456), typeName: "int", matches: false},
		{name: "[]uint16 matches", value: []uint16{1, 2, 3}, typeName: "[]uint16", matches: true},
		{name: "[]uint16 mismatch", value: []uint16{1, 2, 3}, typeName: "[]uint8", matches: false},
		{name: "[]uint8 matches", value: []uint8{1, 2, 3}, typeName: "[]uint8", matches: true},
		{name: "[]byte matches", value: []byte{1, 2, 3}, typeName: "[]byte", matches: true},
		{name: "[]byte vs []uint8", value: []byte{1, 2, 3}, typeName: "[]uint8", matches: true}, // []byte is alias for []uint8
		{name: "[]string matches", value: []string{"a", "b"}, typeName: "[]string", matches: true},
		{name: "[]string mismatch", value: []string{"a", "b"}, typeName: "string", matches: false},
		{name: "bool true matches", value: true, typeName: "bool", matches: true},
		{name: "bool false matches", value: false, typeName: "bool", matches: true},
		{name: "bool mismatch", value: true, typeName: "string", matches: false},
		{name: "uint8 matches", value: uint8(5), typeName: "uint8", matches: true},
		{name: "uint8 mismatch", value: uint8(5), typeName: "uint16", matches: false},
		{name: "uint64 matches", value: uint64(12345), typeName: "uint64", matches: true},
		{name: "uint64 mismatch", value: uint64(12345), typeName: "uint32", matches: false},
		{name: "float64 matches", value: float64(3.14159), typeName: "float64", matches: true},
		{name: "float64 mismatch", value: float64(3.14159), typeName: "int", matches: false},
		{name: "unknown type", value: "test", typeName: "unknown", matches: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TypeMatches(tt.value, tt.typeName)
			if got != tt.matches {
				t.Errorf("TypeMatches(%v [%T], %q) = %v, want %v",
					tt.value, tt.value, tt.typeName, got, tt.matches)
			}
		})
	}
}
