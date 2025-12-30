package heic

// Box type identifiers (4-character codes)
const (
	boxTypeFtyp = "ftyp" // File type box
	boxTypeMeta = "meta" // Metadata box
	boxTypeMdat = "mdat" // Media data box
	boxTypeHdlr = "hdlr" // Handler box
	boxTypePitm = "pitm" // Primary item box
	boxTypeIinf = "iinf" // Item information box
	boxTypeInfe = "infe" // Item information entry
	boxTypeIloc = "iloc" // Item location box
	boxTypeIref = "iref" // Item reference box
	boxTypeIprp = "iprp" // Item properties box
	boxTypeIpco = "ipco" // Item property container
	boxTypeIpma = "ipma" // Item property association
	boxTypeCdsc = "cdsc" // Content describes reference
	boxTypeColr = "colr" // Color information box
)

// Item type identifiers
const (
	itemTypeExif = "Exif" // EXIF metadata item
	itemTypeMime = "mime" // MIME type item (used for XMP)
)

// Color type identifiers
const (
	colorTypeRICC = "rICC" // Restricted ICC profile
	colorTypeProf = "prof" // Unrestricted ICC profile
)

// Valid HEIC/HEIF major brands
var validBrands = []string{
	"heic", "heif", "heix", "hevc", "heim", "heis",
	"mif1", "msf1", "heiv", "hevx",
}

// Box header sizes
const (
	boxHeaderSize      = 8  // Standard box header (size + type)
	boxHeaderLargeSize = 16 // Extended box header (size=1 + type + 64-bit size)
	fullBoxHeaderSize  = 4  // Version (1 byte) + flags (3 bytes)
)

// Size field special values
const (
	sizeExtended = 1 // Indicates 64-bit size follows
	sizeToEOF    = 0 // Box extends to end of file (not supported)
)

// Bit masks for iloc parsing
const (
	maskOffsetSize     = 0xF0 // Upper nibble: offset_size
	maskLengthSize     = 0x0F // Lower nibble: length_size
	maskBaseOffsetSize = 0xF0 // Upper nibble: base_offset_size
	maskIndexSize      = 0x0F // Lower nibble: index_size (v1+)
)

// Bit masks for ipma parsing
const (
	maskEssentialFlag15 = 0x8000 // Essential flag (15-bit mode)
	maskPropertyIndex15 = 0x7FFF // Property index (15-bit mode)
	maskEssentialFlag7  = 0x80   // Essential flag (7-bit mode)
	maskPropertyIndex7  = 0x7F   // Property index (7-bit mode)
)

// XMP detection signatures
var (
	xmpPacketSignature = []byte("<?xpacket")
	xmpXmMetaSignature = []byte("<x:xmpmeta")
)

// TIFF header signatures
const (
	tiffBigEndian    = "MM"
	tiffLittleEndian = "II"
)

// Limits for parsing
const (
	maxTIFFScanOffset = 20 // Max bytes to scan for TIFF header in EXIF
)
