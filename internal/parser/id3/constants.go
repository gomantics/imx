package id3

// ID3 Magic Bytes
var (
	id3v2Signature = [3]byte{'I', 'D', '3'}
)

// Header Sizes
const (
	id3v2HeaderSize = 10 // ID3v2 header is always 10 bytes
	frameHeaderV2   = 6  // ID3v2.2: 3-byte ID + 3-byte size
	frameHeaderV3   = 10 // ID3v2.3/2.4: 4-byte ID + 4-byte size + 2-byte flags
)

// Frame Header Component Sizes
const (
	frameSizeV2    = 3 // ID3v2.2 uses 3-byte frame size
	frameSizeV3    = 4 // ID3v2.3/2.4 use 4-byte frame size
	frameFlagsSize = 2 // ID3v2.3/2.4 use 2-byte frame flags
)

// Frame ID Sizes
const (
	frameIDSizeV2 = 3 // ID3v2.2 uses 3-character frame IDs
	frameIDSizeV3 = 4 // ID3v2.3/2.4 use 4-character frame IDs
)

// Version Numbers
const (
	versionID3v22 = 2
	versionID3v23 = 3
	versionID3v24 = 4
)

// Header Flag Bits
const (
	flagUnsynchronisation = 0x80
	flagExtendedHeader    = 0x40
	flagExperimental      = 0x20
	flagFooter            = 0x10
)

// Text Encoding Values
const (
	encodingISO88591 = 0x00 // ISO-8859-1 (Latin-1)
	encodingUTF16BOM = 0x01 // UTF-16 with BOM
	encodingUTF16BE  = 0x02 // UTF-16BE without BOM
	encodingUTF8     = 0x03 // UTF-8
)

// UTF-16 Byte Order Marks
var (
	bomUTF16LE = [2]byte{0xFF, 0xFE}
	bomUTF16BE = [2]byte{0xFE, 0xFF}
)

// Limits
const (
	maxFrameSize  = 16 * 1024 * 1024 // 16MB reasonable limit per frame
	maxFrameCount = 4096             // Maximum number of frames to parse
)

// Synchsafe Integer Constants
const (
	synchsafeBits = 7    // Each byte uses only 7 bits
	synchsafeMask = 0x7F // Mask for lower 7 bits
)

// Common Frame IDs - ID3v2.3/2.4 (4 characters)
const (
	frameTitle         = "TIT2"
	frameArtist        = "TPE1"
	frameAlbum         = "TALB"
	frameRecordingTime = "TDRC"
	frameYear          = "TYER"
	frameTrack         = "TRCK"
	frameDisc          = "TPOS"
	frameGenre         = "TCON"
	frameAlbumArtist   = "TPE2"
	frameComposer      = "TCOM"
	frameLyricist      = "TEXT"
	framePublisher     = "TPUB"
	frameCopyright     = "TCOP"
	frameEncodedBy     = "TENC"
	frameBPM           = "TBPM"
	frameISRC          = "TSRC"
	frameUserText      = "TXXX"
	frameComment       = "COMM"
	frameLyrics        = "USLT"
	framePicture       = "APIC"
	framePrivate       = "PRIV"
	frameUniqueFileID  = "UFID"
)

// Common Frame IDs - ID3v2.2 (3 characters)
const (
	frameV2Title   = "TT2"
	frameV2Artist  = "TP1"
	frameV2Album   = "TAL"
	frameV2Year    = "TYE"
	frameV2Track   = "TRK"
	frameV2Genre   = "TCO"
	frameV2Comment = "COM"
	frameV2Picture = "PIC"
)
