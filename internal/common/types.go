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

// TagID is a unique identifier for a metadata tag.
// Standard format: SPEC[-Namespace]:LocalName
// - SPEC: EXIF, IPTC, XMP, ICC (uppercase)
// - Namespace: Only for XMP (e.g., dc, xmp, photoshop)
// - LocalName: CamelCase, no spaces
//
// Examples:
//   - "EXIF:Make"
//   - "XMP-dc:Title"
//   - "IPTC:Byline"
//   - "ICC:ProfileDescription"
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

// Spec returns the spec portion of the tag ID.
// Example: "EXIF:Make" → "EXIF"
// Example: "XMP-dc:Title" → "XMP"
func (id TagID) Spec() string {
	s := string(id)
	// Find colon
	colonIdx := -1
	for i, c := range s {
		if c == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx < 0 {
		return ""
	}

	// Get spec part before colon
	specPart := s[:colonIdx]

	// Handle XMP with namespace: "XMP-dc" → "XMP"
	dashIdx := -1
	for i, c := range specPart {
		if c == '-' {
			dashIdx = i
			break
		}
	}
	if dashIdx > 0 {
		return specPart[:dashIdx]
	}

	return specPart
}

// Name returns the local name portion of the tag ID.
// Example: "EXIF:Make" → "Make"
// Example: "XMP-dc:Title" → "Title"
func (id TagID) Name() string {
	s := string(id)
	// Find colon
	colonIdx := -1
	for i, c := range s {
		if c == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx < 0 {
		return s
	}
	return s[colonIdx+1:]
}

// Namespace returns the namespace for XMP tags, empty for others.
// Example: "XMP-dc:Title" → "dc"
// Example: "EXIF:Make" → ""
func (id TagID) Namespace() string {
	s := string(id)
	// Find colon
	colonIdx := -1
	for i, c := range s {
		if c == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx < 0 {
		return ""
	}

	// Get spec part before colon
	specPart := s[:colonIdx]

	// Find dash
	dashIdx := -1
	for i, c := range specPart {
		if c == '-' {
			dashIdx = i
			break
		}
	}
	if dashIdx > 0 {
		return specPart[dashIdx+1:]
	}

	return ""
}

// IsValid returns true if the tag ID follows the standard format.
// Valid format: SPEC[-Namespace]:LocalName
// - Must contain a colon
// - Spec must be uppercase
// - Local name must not be empty
// - No spaces allowed
func (id TagID) IsValid() bool {
	s := string(id)

	// Check for spaces
	for _, c := range s {
		if c == ' ' {
			return false
		}
	}

	// Must contain colon
	colonIdx := -1
	for i, c := range s {
		if c == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx < 0 {
		return false
	}

	// Get parts
	specPart := s[:colonIdx]
	namePart := s[colonIdx+1:]

	// Name must not be empty
	if namePart == "" {
		return false
	}

	// Get spec (before dash if XMP namespace)
	spec := specPart
	dashIdx := -1
	for i, c := range specPart {
		if c == '-' {
			dashIdx = i
			break
		}
	}
	if dashIdx > 0 {
		spec = specPart[:dashIdx]
	}

	// Spec must be uppercase
	for _, c := range spec {
		if c >= 'a' && c <= 'z' {
			return false
		}
	}

	return true
}
