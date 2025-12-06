package format

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
// Note: Spec is defined in internal/meta but we can't import it here due to circular dependency
// So we use int as the underlying type and document that it should be meta.Spec
type RawBlock struct {
	Spec    int // Should be meta.Spec but avoiding circular import
	Payload []byte
	Origin  string // e.g. "APP1 Exif", "eXIf chunk"
	Format  Format
	Index   int // sequence number for multiple blocks of same type
}
