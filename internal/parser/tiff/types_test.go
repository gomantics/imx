package tiff

import (
	"testing"
)

func TestTagType_TypeSize(t *testing.T) {
	tests := []struct {
		name string
		t    TagType
		want int
	}{
		{"TypeByte", TypeByte, 1},
		{"TypeASCII", TypeASCII, 1},
		{"TypeSByte", TypeSByte, 1},
		{"TypeUndefined", TypeUndefined, 1},
		{"TypeShort", TypeShort, 2},
		{"TypeSShort", TypeSShort, 2},
		{"TypeLong", TypeLong, 4},
		{"TypeSLong", TypeSLong, 4},
		{"TypeFloat", TypeFloat, 4},
		{"TypeRational", TypeRational, 8},
		{"TypeSRational", TypeSRational, 8},
		{"TypeDouble", TypeDouble, 8},
		{"Unknown type 0", TagType(0), 0},
		{"Unknown type 99", TagType(99), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.TypeSize(); got != tt.want {
				t.Errorf("TagType.TypeSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagType_String(t *testing.T) {
	tests := []struct {
		name string
		t    TagType
		want string
	}{
		{"TypeByte", TypeByte, "BYTE"},
		{"TypeASCII", TypeASCII, "ASCII"},
		{"TypeShort", TypeShort, "SHORT"},
		{"TypeLong", TypeLong, "LONG"},
		{"TypeRational", TypeRational, "RATIONAL"},
		{"TypeSByte", TypeSByte, "SBYTE"},
		{"TypeUndefined", TypeUndefined, "UNDEFINED"},
		{"TypeSShort", TypeSShort, "SSHORT"},
		{"TypeSLong", TypeSLong, "SLONG"},
		{"TypeSRational", TypeSRational, "SRATIONAL"},
		{"TypeFloat", TypeFloat, "FLOAT"},
		{"TypeDouble", TypeDouble, "DOUBLE"},
		{"Unknown type 0", TagType(0), "UNKNOWN"},
		{"Unknown type 99", TagType(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("TagType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestByteOrderConstants(t *testing.T) {
	// Verify byte order constants are set
	if LittleEndian == nil {
		t.Error("LittleEndian should not be nil")
	}
	if BigEndian == nil {
		t.Error("BigEndian should not be nil")
	}
}

func TestSpecialTagConstants(t *testing.T) {
	tests := []struct {
		name string
		tag  uint16
		want uint16
	}{
		{"TagExifIFD", TagExifIFD, 0x8769},
		{"TagGPSIFD", TagGPSIFD, 0x8825},
		{"TagInteropIFD", TagInteropIFD, 0xA005},
		{"TagICCProfile", TagICCProfile, 0x8773},
		{"TagIPTC", TagIPTC, 0x83BB},
		{"TagXMP", TagXMP, 0x02BC},
		{"TagMakerNote", TagMakerNote, 0x927C},
		{"TagSubIFDs", TagSubIFDs, 0x014A},
		{"TagJPEGInterchange", TagJPEGInterchange, 0x0201},
		{"TagJPEGInterLength", TagJPEGInterLength, 0x0202},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tag != tt.want {
				t.Errorf("%s = 0x%04X, want 0x%04X", tt.name, tt.tag, tt.want)
			}
		})
	}
}

func TestIFDEntry(t *testing.T) {
	entry := IFDEntry{
		Tag:         0x0100,
		Type:        TypeLong,
		Count:       1,
		ValueOffset: 1920,
	}

	if entry.Tag != 0x0100 {
		t.Errorf("Tag = 0x%04X, want 0x0100", entry.Tag)
	}
	if entry.Type != TypeLong {
		t.Errorf("Type = %v, want TypeLong", entry.Type)
	}
	if entry.Count != 1 {
		t.Errorf("Count = %d, want 1", entry.Count)
	}
	if entry.ValueOffset != 1920 {
		t.Errorf("ValueOffset = %d, want 1920", entry.ValueOffset)
	}
}

func TestIFD(t *testing.T) {
	ifd := IFD{
		Entries: []IFDEntry{
			{Tag: 0x0100, Type: TypeLong, Count: 1, ValueOffset: 1920},
			{Tag: 0x0101, Type: TypeLong, Count: 1, ValueOffset: 1080},
		},
		NextIFDOffset: 1024,
	}

	if len(ifd.Entries) != 2 {
		t.Errorf("len(Entries) = %d, want 2", len(ifd.Entries))
	}
	if ifd.NextIFDOffset != 1024 {
		t.Errorf("NextIFDOffset = %d, want 1024", ifd.NextIFDOffset)
	}
}
