package common

import (
	"encoding/binary"
)

// TIFFTypeParser parses TIFF tag values of a specific type.
// TIFF types are defined in the TIFF 6.0 specification and are used by:
// - EXIF metadata (stored as TIFF IFDs)
// - TIFF image files
// - Other formats that embed TIFF data structures
type TIFFTypeParser interface {
	Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string)
}

// ByteParser handles TIFF BYTE type (unsigned 8-bit, type ID 1)
type ByteParser struct{}

func (p ByteParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	if count == 1 {
		return int(data[0]), "byte"
	}
	slice, _ := SafeSlice(data, 0, int(count))
	return slice, "bytes"
}

// ASCIIParser handles TIFF ASCII string type (type ID 2)
type ASCIIParser struct{}

func (p ASCIIParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	slice, _ := SafeSlice(data, 0, int(count))
	str := TrimNullBytesFromSlice(slice)
	return str, "string"
}

// ShortParser handles TIFF SHORT type (unsigned 16-bit, type ID 3)
type ShortParser struct{}

func (p ShortParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	if count == 1 {
		val, _ := ReadUint16(data, 0, byteOrder)
		return int(val), "short"
	}
	vals := make([]int, count)
	for i := uint32(0); i < count; i++ {
		val, _ := ReadUint16(data, int(i*2), byteOrder)
		vals[i] = int(val)
	}
	return vals, "shorts"
}

// LongParser handles TIFF LONG type (unsigned 32-bit, type ID 4)
type LongParser struct{}

func (p LongParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	if count == 1 {
		val, _ := ReadUint32(data, 0, byteOrder)
		return int(val), "long"
	}
	vals := make([]int, count)
	for i := uint32(0); i < count; i++ {
		val, _ := ReadUint32(data, int(i*4), byteOrder)
		vals[i] = int(val)
	}
	return vals, "longs"
}

// RationalParser handles TIFF RATIONAL type (two 32-bit unsigned integers, type ID 5)
type RationalParser struct{}

func (p RationalParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	if count == 1 {
		num, _ := ReadUint32(data, 0, byteOrder)
		denom, _ := ReadUint32(data, 4, byteOrder)
		if denom == 0 {
			return 0.0, "rational"
		}
		return float64(num) / float64(denom), "rational"
	}
	vals := make([]float64, count)
	for i := uint32(0); i < count; i++ {
		num, _ := ReadUint32(data, int(i*8), byteOrder)
		denom, _ := ReadUint32(data, int(i*8+4), byteOrder)
		if denom == 0 {
			vals[i] = 0
		} else {
			vals[i] = float64(num) / float64(denom)
		}
	}
	return vals, "rationals"
}

// SByteParser handles TIFF SBYTE type (signed 8-bit, type ID 6)
type SByteParser struct{}

func (p SByteParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	if count == 1 {
		return int(int8(data[0])), "sbyte"
	}
	vals := make([]int, count)
	for i := uint32(0); i < count; i++ {
		vals[i] = int(int8(data[i]))
	}
	return vals, "sbytes"
}

// UndefinedParser handles TIFF UNDEFINED type (raw bytes, type ID 7)
type UndefinedParser struct{}

func (p UndefinedParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	slice, _ := SafeSlice(data, 0, int(count))
	return slice, "undefined"
}

// SShortParser handles TIFF SSHORT type (signed 16-bit, type ID 8)
type SShortParser struct{}

func (p SShortParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	if count == 1 {
		val, _ := ReadUint16(data, 0, byteOrder)
		return int(int16(val)), "sshort"
	}
	vals := make([]int, count)
	for i := uint32(0); i < count; i++ {
		val, _ := ReadUint16(data, int(i*2), byteOrder)
		vals[i] = int(int16(val))
	}
	return vals, "sshorts"
}

// SLongParser handles TIFF SLONG type (signed 32-bit, type ID 9)
type SLongParser struct{}

func (p SLongParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	if count == 1 {
		val, _ := ReadUint32(data, 0, byteOrder)
		return int(int32(val)), "slong"
	}
	vals := make([]int, count)
	for i := uint32(0); i < count; i++ {
		val, _ := ReadUint32(data, int(i*4), byteOrder)
		vals[i] = int(int32(val))
	}
	return vals, "slongs"
}

// SRationalParser handles TIFF SRATIONAL type (two 32-bit signed integers, type ID 10)
type SRationalParser struct{}

func (p SRationalParser) Parse(data []byte, count uint32, byteOrder binary.ByteOrder) (any, string) {
	if count == 1 {
		numVal, _ := ReadUint32(data, 0, byteOrder)
		denomVal, _ := ReadUint32(data, 4, byteOrder)
		num := int32(numVal)
		denom := int32(denomVal)
		if denom == 0 {
			return 0.0, "srational"
		}
		return float64(num) / float64(denom), "srational"
	}
	vals := make([]float64, count)
	for i := uint32(0); i < count; i++ {
		numVal, _ := ReadUint32(data, int(i*8), byteOrder)
		denomVal, _ := ReadUint32(data, int(i*8+4), byteOrder)
		num := int32(numVal)
		denom := int32(denomVal)
		if denom == 0 {
			vals[i] = 0
		} else {
			vals[i] = float64(num) / float64(denom)
		}
	}
	return vals, "srationals"
}

// TIFFTypeSizes defines the size in bytes for each TIFF type (TIFF 6.0 specification)
var TIFFTypeSizes = map[uint16]int{
	1:  1, // BYTE
	2:  1, // ASCII
	3:  2, // SHORT
	4:  4, // LONG
	5:  8, // RATIONAL (2x uint32)
	6:  1, // SBYTE
	7:  1, // UNDEFINED
	8:  2, // SSHORT
	9:  4, // SLONG
	10: 8, // SRATIONAL (2x int32)
}

// TIFFTypeParsers is the registry of TIFF type parsers by type ID
var TIFFTypeParsers = map[uint16]TIFFTypeParser{
	1:  ByteParser{},
	2:  ASCIIParser{},
	3:  ShortParser{},
	4:  LongParser{},
	5:  RationalParser{},
	6:  SByteParser{},
	7:  UndefinedParser{},
	8:  SShortParser{},
	9:  SLongParser{},
	10: SRationalParser{},
}
