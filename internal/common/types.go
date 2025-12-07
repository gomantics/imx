package common

// Spec represents a metadata specification type
type Spec int

const (
	SpecEXIF Spec = iota
	SpecIPTC
	SpecXMP
	SpecICC
)

// String returns the string representation of the spec
func (s Spec) String() string {
	switch s {
	case SpecEXIF:
		return "exif"
	case SpecIPTC:
		return "iptc"
	case SpecXMP:
		return "xmp"
	case SpecICC:
		return "icc"
	default:
		return "unknown"
	}
}

// Format represents an image container format (JPEG, PNG, WebP, etc.)
type Format int

const (
	FormatJPEG Format = iota
	FormatPNG
	FormatWebP
	FormatTIFF
	FormatHEIF
)

// String returns the string representation of the format
func (f Format) String() string {
	switch f {
	case FormatJPEG:
		return "jpeg"
	case FormatPNG:
		return "png"
	case FormatWebP:
		return "webp"
	case FormatTIFF:
		return "tiff"
	case FormatHEIF:
		return "heif"
	default:
		return "unknown"
	}
}

// RawBlock is a raw metadata payload extracted from an image format
type RawBlock struct {
	Spec    Spec
	Payload []byte
	Origin  string // e.g. "APP1 Exif", "eXIf chunk"
	Format  Format
	Index   int // sequence number for multiple blocks of same type
}

// TagID is a unique identifier for a metadata tag
type TagID string

// Directory is a logical collection of tags for a given metadata spec
type Directory struct {
	Spec Spec
	Name string
	Tags map[TagID]Tag
}

// Tag represents a single metadata attribute
type Tag struct {
	Spec     Spec
	ID       TagID
	Name     string
	DataType string
	Value    any
	Raw      []byte
}
