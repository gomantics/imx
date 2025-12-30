package tiff

// TIFF file format constants
const (
	// TIFF magic number (42 in decimal)
	tiffMagicNumber = 42

	// TIFF header sizes
	tiffHeaderSize       = 8 // Complete TIFF header (byte order + magic + IFD offset)
	tiffHeaderPrefixSize = 4 // Just byte order + magic for detection

	// IFD entry structure
	ifdEntrySize        = 12 // Size of one IFD entry in bytes
	ifdEntryCountSize   = 2  // Size of entry count field
	ifdEntryTagOffset   = 0  // Offset of tag field in entry
	ifdEntryTypeOffset  = 2  // Offset of type field in entry
	ifdEntryCountOffset = 4  // Offset of count field in entry
	ifdEntryValueOffset = 8  // Offset of value/offset field in entry

	// Inline data threshold (values <= 4 bytes are stored inline)
	inlineDataThreshold = 4

	// Data type sizes in bytes
	typeSizeByte      = 1
	typeSizeASCII     = 1
	typeSizeShort     = 2
	typeSizeLong      = 4
	typeSizeRational  = 8
	typeSizeSByte     = 1
	typeSizeSShort    = 2
	typeSizeSLong     = 4
	typeSizeSRational = 8
	typeSizeFloat     = 4
	typeSizeDouble    = 8

	// Buffer sizes for reading
	bufferSizeUint16 = 2
	bufferSizeUint32 = 4
	bufferSizeUint64 = 8

	// Special tag values
	tagGPSVersionID = 0x0000 // GPS Version ID tag

	// Byte order markers
	byteOrderLittleEndian = 'I'
	byteOrderBigEndian    = 'M'
)
