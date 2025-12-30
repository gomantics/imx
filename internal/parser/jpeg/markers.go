package jpeg

// JPEG markers
const (
	markerPrefix = 0xFF // Marker prefix
	markerSOI    = 0xD8 // Start of Image
	markerEOI    = 0xD9 // End of Image
	markerSOS    = 0xDA // Start of Scan (image data follows)
	markerAPP0   = 0xE0 // APP0 - JFIF
	markerAPP1   = 0xE1 // APP1 - EXIF, XMP
	markerAPP2   = 0xE2 // APP2 - ICC Profile
	markerAPP13  = 0xED // APP13 - IPTC/Photoshop
)

// Metadata identifiers in APP segments
var (
	identEXIF      = []byte("Exif\x00\x00")
	identXMP       = []byte("http://ns.adobe.com/xap/1.0/\x00")
	identICC       = []byte("ICC_PROFILE\x00")
	identPhotoshop = []byte("Photoshop 3.0\x00")
)
