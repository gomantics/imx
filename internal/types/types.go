package types

// Format represents an image container format (JPEG, PNG, WebP, etc.)
type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatPNG  Format = "png"
	FormatWebP Format = "webp"
	FormatTIFF Format = "tiff"
	FormatHEIF Format = "heif"
)

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

// ExtractorConfig holds configuration options for metadata extraction
type ExtractorConfig struct {
	MaxBytes       int64    // Maximum bytes to read (0 = no limit)
	BufferSize     int      // Buffer size for reading (0 = default 64KB)
	Specs          []Spec   // Metadata specs to extract (nil/empty = all)
	Formats        []Format // Formats to detect (nil/empty = all registered)
	StopOnFirstErr bool     // Stop on first error vs. continue with partial results
}
