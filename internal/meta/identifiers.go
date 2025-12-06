package meta

// Magic bytes for identifying metadata types embedded in image files.
// These identifiers are format-agnostic and used across different container
// formats (JPEG, PNG, TIFF, etc.) to recognize metadata blocks.
var (
	// MagicEXIF identifies EXIF metadata blocks
	MagicEXIF = []byte("Exif\x00\x00")

	// MagicXMP identifies XMP metadata blocks (Adobe XMP namespace)
	MagicXMP = []byte("http://ns.adobe.com/xap/1.0/\x00")

	// MagicICC identifies ICC color profile blocks
	MagicICC = []byte("ICC_PROFILE\x00")

	// MagicIPTC identifies IPTC/Photoshop metadata blocks
	MagicIPTC = []byte("Photoshop 3.0\x00")
)
