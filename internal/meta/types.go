package meta

// Spec represents a metadata specification (EXIF, IPTC, XMP, ICC, etc.)
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
