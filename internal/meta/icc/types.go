package icc

import "time"

// ProfileClass represents the ICC profile device class
type ProfileClass uint32

const (
	ClassInput      ProfileClass = 0x73636E72 // 'scnr' - Input device (scanner)
	ClassDisplay    ProfileClass = 0x6D6E7472 // 'mntr' - Display device (monitor)
	ClassOutput     ProfileClass = 0x70727472 // 'prtr' - Output device (printer)
	ClassLink       ProfileClass = 0x6C696E6B // 'link' - Device link
	ClassAbstract   ProfileClass = 0x61627374 // 'abst' - Abstract profile
	ClassColorSpace ProfileClass = 0x73706163 // 'spac' - Color space conversion
	ClassNamedColor ProfileClass = 0x6E6D636C // 'nmcl' - Named color
)

// String returns a human-readable name for the profile class
func (c ProfileClass) String() string {
	switch c {
	case ClassInput:
		return "Input Device (Scanner)"
	case ClassDisplay:
		return "Display Device (Monitor)"
	case ClassOutput:
		return "Output Device (Printer)"
	case ClassLink:
		return "Device Link"
	case ClassAbstract:
		return "Abstract Profile"
	case ClassColorSpace:
		return "Color Space Conversion"
	case ClassNamedColor:
		return "Named Color"
	default:
		return signatureToString(uint32(c))
	}
}

// ColorSpace represents the ICC color space signature
type ColorSpace uint32

const (
	SpaceXYZ  ColorSpace = 0x58595A20 // 'XYZ '
	SpaceLab  ColorSpace = 0x4C616220 // 'Lab '
	SpaceLuv  ColorSpace = 0x4C757620 // 'Luv '
	SpaceYCbr ColorSpace = 0x59436272 // 'YCbr'
	SpaceYxy  ColorSpace = 0x59787920 // 'Yxy '
	SpaceRGB  ColorSpace = 0x52474220 // 'RGB '
	SpaceGray ColorSpace = 0x47524159 // 'GRAY'
	SpaceHSV  ColorSpace = 0x48535620 // 'HSV '
	SpaceHLS  ColorSpace = 0x484C5320 // 'HLS '
	SpaceCMYK ColorSpace = 0x434D594B // 'CMYK'
	SpaceCMY  ColorSpace = 0x434D5920 // 'CMY '
	Space2CLR ColorSpace = 0x32434C52 // '2CLR'
	Space3CLR ColorSpace = 0x33434C52 // '3CLR'
	Space4CLR ColorSpace = 0x34434C52 // '4CLR'
	Space5CLR ColorSpace = 0x35434C52 // '5CLR'
	Space6CLR ColorSpace = 0x36434C52 // '6CLR'
	Space7CLR ColorSpace = 0x37434C52 // '7CLR'
	Space8CLR ColorSpace = 0x38434C52 // '8CLR'
	Space9CLR ColorSpace = 0x39434C52 // '9CLR'
	SpaceACLR ColorSpace = 0x41434C52 // 'ACLR' (10 color)
	SpaceBCLR ColorSpace = 0x42434C52 // 'BCLR' (11 color)
	SpaceCCLR ColorSpace = 0x43434C52 // 'CCLR' (12 color)
	SpaceDCLR ColorSpace = 0x44434C52 // 'DCLR' (13 color)
	SpaceECLR ColorSpace = 0x45434C52 // 'ECLR' (14 color)
	SpaceFCLR ColorSpace = 0x46434C52 // 'FCLR' (15 color)
)

// String returns a human-readable name for the color space
func (s ColorSpace) String() string {
	switch s {
	case SpaceXYZ:
		return "XYZ"
	case SpaceLab:
		return "Lab"
	case SpaceLuv:
		return "Luv"
	case SpaceYCbr:
		return "YCbCr"
	case SpaceYxy:
		return "Yxy"
	case SpaceRGB:
		return "RGB"
	case SpaceGray:
		return "Grayscale"
	case SpaceHSV:
		return "HSV"
	case SpaceHLS:
		return "HLS"
	case SpaceCMYK:
		return "CMYK"
	case SpaceCMY:
		return "CMY"
	case Space2CLR:
		return "2 Color"
	case Space3CLR:
		return "3 Color"
	case Space4CLR:
		return "4 Color"
	case Space5CLR:
		return "5 Color"
	case Space6CLR:
		return "6 Color"
	case Space7CLR:
		return "7 Color"
	case Space8CLR:
		return "8 Color"
	case Space9CLR:
		return "9 Color"
	case SpaceACLR:
		return "10 Color"
	case SpaceBCLR:
		return "11 Color"
	case SpaceCCLR:
		return "12 Color"
	case SpaceDCLR:
		return "13 Color"
	case SpaceECLR:
		return "14 Color"
	case SpaceFCLR:
		return "15 Color"
	default:
		return signatureToString(uint32(s))
	}
}

// Platform represents the primary platform/OS signature
type Platform uint32

const (
	PlatformApple     Platform = 0x4150504C // 'APPL'
	PlatformMicrosoft Platform = 0x4D534654 // 'MSFT'
	PlatformSGI       Platform = 0x53474920 // 'SGI '
	PlatformSun       Platform = 0x53554E57 // 'SUNW'
	PlatformTaligent  Platform = 0x54474E54 // 'TGNT'
)

// String returns a human-readable name for the platform
func (p Platform) String() string {
	switch p {
	case PlatformApple:
		return "Apple"
	case PlatformMicrosoft:
		return "Microsoft"
	case PlatformSGI:
		return "Silicon Graphics"
	case PlatformSun:
		return "Sun Microsystems"
	case PlatformTaligent:
		return "Taligent"
	default:
		if p == 0 {
			return "Unspecified"
		}
		return signatureToString(uint32(p))
	}
}

// RenderingIntent represents the rendering intent
type RenderingIntent uint32

const (
	IntentPerceptual           RenderingIntent = 0
	IntentRelativeColorimetric RenderingIntent = 1
	IntentSaturation           RenderingIntent = 2
	IntentAbsoluteColorimetric RenderingIntent = 3
)

// String returns a human-readable name for the rendering intent
func (i RenderingIntent) String() string {
	switch i {
	case IntentPerceptual:
		return "Perceptual"
	case IntentRelativeColorimetric:
		return "Media-Relative Colorimetric"
	case IntentSaturation:
		return "Saturation"
	case IntentAbsoluteColorimetric:
		return "ICC-Absolute Colorimetric"
	default:
		return "Unknown"
	}
}

// ProfileFlags represents profile flags (embedded profile, use with embedded data only)
type ProfileFlags uint32

// IsEmbedded returns true if the profile is embedded
func (f ProfileFlags) IsEmbedded() bool {
	return f&0x01 != 0
}

// IsIndependent returns true if the profile can be used independently
func (f ProfileFlags) IsIndependent() bool {
	return f&0x02 == 0
}

// DeviceAttributes represents device attributes (reflective/transparency, glossy/matte, etc.)
type DeviceAttributes uint64

// IsReflective returns true if the media is reflective (vs transmissive)
func (a DeviceAttributes) IsReflective() bool {
	return a&0x01 == 0
}

// IsGlossy returns true if the media is glossy (vs matte)
func (a DeviceAttributes) IsGlossy() bool {
	return a&0x02 == 0
}

// IsPositive returns true for positive media (vs negative)
func (a DeviceAttributes) IsPositive() bool {
	return a&0x04 == 0
}

// IsColor returns true for color media (vs black & white)
func (a DeviceAttributes) IsColor() bool {
	return a&0x08 == 0
}

// XYZNumber represents a CIE XYZ color value (s15Fixed16Number format)
type XYZNumber struct {
	X float64
	Y float64
	Z float64
}

// Version represents an ICC profile version
type Version struct {
	Major  uint8
	Minor  uint8
	BugFix uint8
}

// String returns the version as a string (e.g., "4.3.0")
func (v Version) String() string {
	return string('0'+v.Major) + "." + string('0'+v.Minor) + "." + string('0'+v.BugFix)
}

// Header represents the 128-byte ICC profile header
type Header struct {
	ProfileSize        uint32
	PreferredCMM       string // 4-char signature
	Version            Version
	ProfileClass       ProfileClass
	DataColorSpace     ColorSpace
	PCS                ColorSpace // Profile Connection Space
	Created            time.Time
	Signature          string // Should always be 'acsp'
	Platform           Platform
	Flags              ProfileFlags
	DeviceManufacturer string // 4-char signature
	DeviceModel        string // 4-char signature
	DeviceAttributes   DeviceAttributes
	RenderingIntent    RenderingIntent
	PCSIlluminant      XYZNumber
	Creator            string // 4-char signature
	ProfileID          []byte // 16-byte MD5 hash (v4+)
}

// TagEntry represents an entry in the tag table
type TagEntry struct {
	Signature string // 4-char tag signature
	Offset    uint32
	Size      uint32
}

// Profile represents a parsed ICC profile
type Profile struct {
	Header Header
	Tags   []TagEntry
	Data   []byte // Full profile data for tag extraction
}

// signatureToString converts a 4-byte signature to a string
func signatureToString(sig uint32) string {
	b := make([]byte, 4)
	b[0] = byte(sig >> 24)
	b[1] = byte(sig >> 16)
	b[2] = byte(sig >> 8)
	b[3] = byte(sig)
	return string(b)
}
