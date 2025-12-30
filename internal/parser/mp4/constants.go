package mp4

// Atom/Box sizes and offsets
const (
	atomHeaderSize    = 8  // 4-byte size + 4-byte type
	fullBoxHeaderSize = 4  // Full box version (1 byte) + flags (3 bytes)
	minMetadataAtom   = 16 // data atom header size
)

// Atom types
const (
	atomFTYP = "ftyp"
	atomMOOV = "moov"
	atomUDTA = "udta"
	atomMETA = "meta"
	atomILST = "ilst"
	atomDATA = "data"
)

// Metadata data type indicators
const (
	dataTypeBinary = 0
	dataTypeUTF8   = 1
	dataTypeSigned = 21
)
