package gif

// GIF Format Separators and Block Types
const (
	// Block separators
	separatorExtension       = 0x21 // Extension block
	separatorImageDescriptor = 0x2C // Image Descriptor
	separatorTrailer         = 0x3B // Trailer (end of GIF)
	separatorBlockTerminator = 0x00 // Block terminator or padding
)

// Extension Labels
const (
	labelPlainText      = 0x01 // Plain Text Extension
	labelGraphicControl = 0xF9 // Graphic Control Extension
	labelComment        = 0xFE // Comment Extension
	labelApplicationExt = 0xFF // Application Extension
)

// GIF Header and Structure Constants
const (
	// Header sizes
	gifHeaderSize           = 6  // Size of GIF header ("GIF87a" or "GIF89a")
	logicalScreenDescSize   = 7  // Size of Logical Screen Descriptor
	gifHeaderTotalSize      = 13 // Total size (header + LSD)
	imageDescriptorSize     = 9  // Size of Image Descriptor
	applicationExtBlockSize = 11 // Standard Application Extension block size
	applicationIDLength     = 8  // Application identifier length
	authCodeLength          = 3  // Authentication code length
	colorTableEntrySize     = 3  // RGB bytes per color table entry

	// Packed field bit masks (for flags in Logical Screen Descriptor and Image Descriptor)
	maskGlobalColorTable = 0x80 // Global/Local Color Table flag
	maskColorResolution  = 0x70 // Color resolution
	maskSortFlag         = 0x08 // Sort flag
	maskColorTableSize   = 0x07 // Color table size
)

// XMP and Application Extension Constants
const (
	xmpApplicationID      = "XMP Data" // XMP Application identifier
	netscapeApplicationID = "NETSCAPE" // NETSCAPE Application identifier
	netscapeAuthCode      = "2.0"      // NETSCAPE authentication code
	xmpPacketStartChar    = 0x3C       // '<' character, indicates old-format XMP
	xmpMagicTrailerSize   = 257        // Size of XMP magic trailer (1 + 256)
	xmpMagicTrailerMarker = 0x01       // First byte of magic trailer
	xmpMagicTrailerFill   = 0x00       // Fill byte in magic trailer
	xmpReadChunkSize      = 64 * 1024  // 64KB chunks for reading old-format XMP

	// NETSCAPE animation extension
	netscapeSubBlockSize    = 3 // Size of NETSCAPE sub-block
	netscapeSubBlockID      = 1 // Sub-block ID for loop count
	netscapeLoopCountOffset = 2 // Offset to loop count in sub-block
)

// GIF Version Strings
const (
	gifVersion87a = "GIF87a"
	gifVersion89a = "GIF89a"
)
