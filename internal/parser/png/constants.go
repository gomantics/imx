package png

// PNG format constants

// PNG signature (8 bytes at start of file)
var pngSignature = []byte{137, 80, 78, 71, 13, 10, 26, 10}

// Chunk structure constants
const (
	chunkHeaderSize = 8 // Length (4) + Type (4)
	crcSize         = 4 // CRC field size
)

// Standard PNG chunk types
const (
	chunkTypeIHDR = "IHDR" // Image header
	chunkTypePLTE = "PLTE" // Palette
	chunkTypeIDAT = "IDAT" // Image data
	chunkTypeIEND = "IEND" // Image end
	chunkTypecHRM = "cHRM" // Chromaticity
	chunkTypegAMA = "gAMA" // Gamma
	chunkTypepHYs = "pHYs" // Physical dimensions
	chunkTypetIME = "tIME" // Modification time
	chunkTypebKGD = "bKGD" // Background color
	chunkTypeeXIf = "eXIf" // EXIF metadata
	chunkTypeiTXt = "iTXt" // International text
	chunkTypeiCCP = "iCCP" // ICC color profile
	chunkTypetEXt = "tEXt" // Uncompressed text
	chunkTypezTXt = "zTXt" // Compressed text
)

// IHDR chunk constants
const (
	ihdrChunkSize         = 13 // IHDR data size
	ihdrWidthOffset       = 0  // Width field offset
	ihdrHeightOffset      = 4  // Height field offset
	ihdrBitDepthOffset    = 8  // Bit depth field offset
	ihdrColorTypeOffset   = 9  // Color type field offset
	ihdrCompressionOffset = 10 // Compression method offset
	ihdrFilterOffset      = 11 // Filter method offset
	ihdrInterlaceOffset   = 12 // Interlace method offset
)

// Color type values
const (
	colorTypeGrayscale      = 0
	colorTypeRGB            = 2
	colorTypePalette        = 3
	colorTypeGrayscaleAlpha = 4
	colorTypeRGBA           = 6
)

// Compression method values
const (
	compressionDeflate = 0
)

// Filter method values
const (
	filterAdaptive = 0
)

// Interlace method values
const (
	interlaceNone  = 0
	interlaceAdam7 = 1
)

// cHRM chunk constants
const (
	chrmChunkSize = 32       // cHRM data size
	chrmScale     = 100000.0 // Values are stored as int / 100000
)

// gAMA chunk constants
const (
	gamaChunkSize = 4        // gAMA data size
	gamaScale     = 100000.0 // Gamma is stored as int / 100000
)

// pHYs chunk constants
const (
	physChunkSize     = 9 // pHYs data size
	physPixelsXOffset = 0 // Pixels per unit X offset
	physPixelsYOffset = 4 // Pixels per unit Y offset
	physUnitOffset    = 8 // Unit specifier offset
	physUnitUnknown   = 0 // Unit unknown
	physUnitMeter     = 1 // Unit is meters
)

// tIME chunk constants
const (
	timeChunkSize   = 7 // tIME data size
	timeYearOffset  = 0 // Year field offset
	timeMonthOffset = 2 // Month field offset
	timeDayOffset   = 3 // Day field offset
	timeHourOffset  = 4 // Hour field offset
	timeMinOffset   = 5 // Minute field offset
	timeSecOffset   = 6 // Second field offset
)

// iTXt chunk constants
const (
	itxtCompressionFlagOffset   = 1                   // Compression flag offset from after keyword null
	itxtCompressionMethodOffset = 1                   // Compression method offset from compression flag
	itxtKeywordEnd              = 0                   // Null terminator byte value
	itxtCompressionNone         = 0                   // No compression
	itxtCompressionDeflate      = 0                   // Deflate compression (same as none for now)
	itxtXMPKeyword              = "XML:com.adobe.xmp" // XMP keyword
)

// zTXt chunk constants
const (
	ztxtCompressionMethodOffset = 1 // Compression method offset from keyword null terminator
	ztxtCompressionDeflate      = 0 // Deflate compression
)

// iCCP chunk constants
const (
	iccpCompressionDeflate = 0 // Deflate compression for ICC profiles
)

// bKGD chunk constants (size varies by color type)
const (
	bkgdGrayscaleSize   = 1 // Grayscale background (palette index or 8-bit)
	bkgdGrayscale16Size = 2 // 16-bit grayscale background
	bkgdRGBSize         = 6 // RGB background (16-bit per channel)
)

// Text chunk parsing constants
const (
	textKeywordMaxLen = 79          // Maximum keyword length (per PNG spec)
	textValueMaxLen   = 1024 * 1024 // 1MB limit for text values
)
