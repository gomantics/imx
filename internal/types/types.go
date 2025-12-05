package types

// Format represents a container format (JPEG, PNG, WebP, etc.)
type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatPNG  Format = "png"
	FormatWebP Format = "webp"
	FormatTIFF Format = "tiff"
	FormatHEIF Format = "heif"
)

// Namespace represents a metadata namespace
type Namespace string

const (
	NamespaceEXIF Namespace = "exif"
	NamespaceIPTC Namespace = "iptc"
	NamespaceXMP  Namespace = "xmp"
	NamespaceICC  Namespace = "icc"
)

// ExtractorConfig holds configuration options for metadata extraction
type ExtractorConfig struct {
	MaxBytes       int64       // Maximum bytes to read (0 = no limit)
	BufferSize     int         // Buffer size for reading (0 = default 64KB)
	Namespaces     []Namespace // Namespaces to extract (nil/empty = all)
	Formats        []Format    // Formats to detect (nil/empty = all registered)
	StopOnFirstErr bool        // Stop on first error vs. continue with partial results
}
