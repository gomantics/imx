package icc

// ICC Profile Structure Offsets
const (
	// Header offsets
	offsetProfileSize     = 0
	offsetCMMType         = 4
	offsetProfileVersion  = 8
	offsetProfileClass    = 12
	offsetColorSpace      = 16
	offsetPCS             = 20
	offsetDateTime        = 24
	offsetSignature       = 36
	offsetPlatform        = 40
	offsetProfileFlags    = 44
	offsetDeviceManuf     = 48
	offsetDeviceModel     = 52
	offsetDeviceAttrs     = 56
	offsetRenderingIntent = 64
	offsetIlluminant      = 68
	offsetProfileCreator  = 80
	offsetProfileID       = 84
	offsetProfileIDEnd    = 100
	offsetTagTable        = 128
	offsetTagTableCount   = 128
	offsetTagTableEntries = 132
)

// ICC Profile Sizes
const (
	headerSize     = 128
	tagRecordSize  = 12
	signatureSize  = 4
	dateTimeSize   = 12
	illuminantSize = 12
	profileIDSize  = 16
	minTagDataSize = 8 // Type signature (4) + reserved (4)
)

// Limits
const (
	maxTagCount = 1000 // Sanity limit for tag count
)

// ICC Signature
var iccSignature = [4]byte{'a', 'c', 's', 'p'}

// Type Signatures
const (
	typeText = "text"
	typeDesc = "desc"
	typeMluc = "mluc"
	typeXYZ  = "XYZ "
	typeSf32 = "sf32"
	typeUf32 = "uf32"
	typeSig  = "sig "
	typeCurv = "curv"
	typePara = "para"
	typeDtim = "dtim"
	typeMeas = "meas"
	typeView = "view"
	typeChrm = "chrm"
)

// Fixed Point Conversion Constants
const (
	s15Fixed16Divisor = 65536.0
	u16Fixed16Divisor = 65536.0
	u8Fixed8Divisor   = 256.0
	curvePointMax     = 65535.0
)

// XYZ Type Constants
const (
	xyzNumberSize = 12 // 3 s15Fixed16 values
	xyzMinTagSize = 20 // 8 byte header + at least one XYZ
)

// Curve Type Constants
const (
	curvMinTagSize  = 12 // 8 byte header + 4 byte count
	curvCountOffset = 8
	curvDataOffset  = 12
	curvPointSize   = 2 // uint16
	curvGammaSize   = 2 // u8Fixed8
)

// Parametric Curve Function Types
const (
	paraFuncSimpleGamma      = 0 // Y = X^g
	paraFuncCIELab           = 1 // Y = (aX+b)^g if X >= -b/a, else 0
	paraFuncIEC61966         = 2 // Y = (aX+b)^g + c if X >= -b/a, else c
	paraFuncIEC61966Extended = 3 // Y = (aX+b)^g if X >= d, else cX
	paraFuncFull             = 4 // Y = (aX+b)^g + e if X >= d, else cX + f
)

// Measurement Type Constants
const (
	measMinTagSize     = 44
	measDataSize       = 36
	measObserverOffset = 0
	measBackingOffset  = 4
	measGeometryOffset = 16
	measFlareOffset    = 20
	measIllumOffset    = 24
)

// Observer Types
const (
	observerCIE1931 = 1
	observerCIE1964 = 2
)

// Geometry Types
const (
	geometry045or450 = 1
	geometry0dord0   = 2
)

// Illuminant Types (shared by measurement and viewing conditions)
const (
	illuminantD50     = 1
	illuminantD65     = 2
	illuminantD93     = 3
	illuminantF2      = 4
	illuminantD55     = 5
	illuminantA       = 6
	illuminantEquiPow = 7
	illuminantF8      = 8
)

// Viewing Conditions Constants
const (
	viewMinTagSize = 36
	viewDataSize   = 28
)

// Chromaticity Type Constants
const (
	chrmMinTagSize  = 12
	chrmChanOffset  = 0
	chrmPhosOffset  = 2
	chrmCoordOffset = 4
	chrmCoordSize   = 8 // 2 u16Fixed16 values per channel
)

// Phosphor Types
const (
	phosphorITURBT709   = 1
	phosphorSMPTERP145  = 2
	phosphorEBUTech3213 = 3
	phosphorP22         = 4
)

// Text Type Constants
const (
	textMinTagSize = 8
)

// Description Type Constants
const (
	descMinTagSize   = 12
	descCountOffset  = 8
	descStringOffset = 12
)

// Multi-Localized Unicode Type Constants
const (
	mlucMinTagSize   = 16
	mlucCountOffset  = 8
	mlucRecordOffset = 16
	mlucRecordSize   = 12
	mlucLengthOffset = 4
	mlucStringOffset = 8
)

// Signature Type Constants
const (
	sigMinTagSize = 12
)

// DateTime Type Constants
const (
	dtimMinTagSize = 20
	dtimDataSize   = 12
)

// S15Fixed16 Array Type Constants
const (
	sf32MinTagSize  = 8
	sf32ElementSize = 4
)

// U16Fixed16 Array Type Constants
const (
	uf32MinTagSize  = 8
	uf32ElementSize = 4
)
