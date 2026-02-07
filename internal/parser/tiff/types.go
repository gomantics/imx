package tiff

import (
	"encoding/binary"
)

// ByteOrder represents TIFF byte order
type ByteOrder binary.ByteOrder

var (
	LittleEndian ByteOrder = binary.LittleEndian
	BigEndian    ByteOrder = binary.BigEndian
)

// TagType represents TIFF data type
type TagType uint16

const (
	TypeByte      TagType = 1
	TypeASCII     TagType = 2
	TypeShort     TagType = 3
	TypeLong      TagType = 4
	TypeRational  TagType = 5
	TypeSByte     TagType = 6
	TypeUndefined TagType = 7
	TypeSShort    TagType = 8
	TypeSLong     TagType = 9
	TypeSRational TagType = 10
	TypeFloat     TagType = 11
	TypeDouble    TagType = 12
)

// Special TIFF tags
const (
	TagExifIFD         uint16 = 0x8769
	TagGPSIFD          uint16 = 0x8825
	TagInteropIFD      uint16 = 0xA005
	TagICCProfile      uint16 = 0x8773
	TagIPTC            uint16 = 0x83BB
	TagXMP             uint16 = 0x02BC // XMLPacket (decimal 700)
	TagMakerNote       uint16 = 0x927C
	TagUserComment     uint16 = 0x9286
	TagSubIFDs         uint16 = 0x014A
	TagJPEGInterchange uint16 = 0x0201
	TagJPEGInterLength uint16 = 0x0202
)

// IFDEntry represents a single IFD entry
type IFDEntry struct {
	Tag         uint16
	Type        TagType
	Count       uint32
	ValueOffset uint32
}

// IFD represents an Image File Directory
type IFD struct {
	Entries       []IFDEntry
	NextIFDOffset uint32
}

// TypeSize returns the size in bytes of a TagType
func (t TagType) TypeSize() int {
	switch t {
	case TypeByte, TypeSByte, TypeASCII, TypeUndefined:
		return typeSizeByte
	case TypeShort, TypeSShort:
		return typeSizeShort
	case TypeLong, TypeSLong, TypeFloat:
		return typeSizeLong
	case TypeRational, TypeSRational, TypeDouble:
		return typeSizeRational
	default:
		return 0
	}
}

// String returns the string representation of TagType
func (t TagType) String() string {
	switch t {
	case TypeByte:
		return "BYTE"
	case TypeASCII:
		return "ASCII"
	case TypeShort:
		return "SHORT"
	case TypeLong:
		return "LONG"
	case TypeRational:
		return "RATIONAL"
	case TypeSByte:
		return "SBYTE"
	case TypeUndefined:
		return "UNDEFINED"
	case TypeSShort:
		return "SSHORT"
	case TypeSLong:
		return "SLONG"
	case TypeSRational:
		return "SRATIONAL"
	case TypeFloat:
		return "FLOAT"
	case TypeDouble:
		return "DOUBLE"
	default:
		return "UNKNOWN"
	}
}
