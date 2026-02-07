package tiff

import "fmt"

// decodeEnumValue returns a human-readable string for enum tag values.
// Returns empty string if the tag is not an enum or value is unknown.
func decodeEnumValue(tag uint16, _ string, value any) string {
	// Handle uint16 values (most common for enums)
	var v uint16
	switch val := value.(type) {
	case uint16:
		v = val
	case uint32:
		// Some tags may be stored as uint32
		v = uint16(val)
	case int:
		v = uint16(val)
	default:
		return ""
	}

	// Check TIFF tags first
	if lookup := tiffEnumValues[tag]; lookup != nil {
		if name, ok := lookup[v]; ok {
			return name
		}
	}

	// Check EXIF tags
	if lookup := exifEnumValues[tag]; lookup != nil {
		if name, ok := lookup[v]; ok {
			return name
		}
	}

	return ""
}

// TIFF enum value mappings
var tiffEnumValues = map[uint16]map[uint16]string{
	// Compression (0x0103)
	0x0103: {
		1:     "Uncompressed",
		2:     "CCITT 1D",
		3:     "T4/Group 3 Fax",
		4:     "T6/Group 4 Fax",
		5:     "LZW",
		6:     "JPEG (old-style)",
		7:     "JPEG",
		8:     "Adobe Deflate",
		9:     "JBIG B&W",
		10:    "JBIG Color",
		99:    "JPEG",
		262:   "Kodak 262",
		32766: "Next",
		32767: "Sony ARW Compressed",
		32769: "Packed RAW",
		32770: "Samsung SRW Compressed",
		32771: "CCIRLEW",
		32772: "Samsung SRW Compressed 2",
		32773: "PackBits",
		32809: "Thunderscan",
		32867: "Kodak KDC Compressed",
		32895: "IT8CTPAD",
		32896: "IT8LW",
		32897: "IT8MP",
		32898: "IT8BL",
		32908: "PixarFilm",
		32909: "PixarLog",
		32946: "Deflate",
		32947: "DCS",
		33003: "Aperio JPEG 2000 YCbCr",
		33005: "Aperio JPEG 2000 RGB",
		34661: "JBIG",
		34676: "SGILog",
		34677: "SGILog24",
		34712: "JPEG 2000",
		34713: "Nikon NEF Compressed",
		34715: "JBIG2 TIFF FX",
		34718: "Microsoft Document Imaging (MDI) Binary Level Codec",
		34719: "Microsoft Document Imaging (MDI) Progressive Transform Codec",
		34720: "Microsoft Document Imaging (MDI) Vector",
		34887: "ESRI Lerc",
		34892: "Lossy JPEG",
		34925: "LZMA2",
		34926: "Zstd",
		34927: "WebP",
		34933: "PNG",
		34934: "JPEG XR",
		65000: "Kodak DCR Compressed",
		65535: "Pentax PEF Compressed",
	},

	// PhotometricInterpretation (0x0106)
	0x0106: {
		0:     "WhiteIsZero",
		1:     "BlackIsZero",
		2:     "RGB",
		3:     "RGB Palette",
		4:     "Transparency Mask",
		5:     "CMYK",
		6:     "YCbCr",
		8:     "CIELab",
		9:     "ICCLab",
		10:    "ITULab",
		32803: "Color Filter Array",
		32844: "Pixar LogL",
		32845: "Pixar LogLuv",
		32892: "Sequential Color Filter",
		34892: "Linear Raw",
		51177: "Depth Map",
	},

	// Orientation (0x0112)
	0x0112: {
		1: "Horizontal (normal)",
		2: "Mirror horizontal",
		3: "Rotate 180",
		4: "Mirror vertical",
		5: "Mirror horizontal and rotate 270 CW",
		6: "Rotate 90 CW",
		7: "Mirror horizontal and rotate 90 CW",
		8: "Rotate 270 CW",
	},

	// PlanarConfiguration (0x011C)
	0x011C: {
		1: "Chunky",
		2: "Planar",
	},

	// ResolutionUnit (0x0128)
	0x0128: {
		1: "None",
		2: "inches",
		3: "centimeters",
	},

	// YCbCrPositioning (0x0213)
	0x0213: {
		1: "Centered",
		2: "Co-sited",
	},
}

// EXIF enum value mappings
var exifEnumValues = map[uint16]map[uint16]string{
	// ExposureProgram (0x8822)
	0x8822: {
		0: "Not Defined",
		1: "Manual",
		2: "Program AE",
		3: "Aperture-priority AE",
		4: "Shutter speed priority AE",
		5: "Creative (Slow speed)",
		6: "Action (High speed)",
		7: "Portrait",
		8: "Landscape",
		9: "Bulb",
	},

	// SensitivityType (0x8830)
	0x8830: {
		0: "Unknown",
		1: "Standard Output Sensitivity",
		2: "Recommended Exposure Index",
		3: "ISO Speed",
		4: "Standard Output Sensitivity and Recommended Exposure Index",
		5: "Standard Output Sensitivity and ISO Speed",
		6: "Recommended Exposure Index and ISO Speed",
		7: "Standard Output Sensitivity, Recommended Exposure Index and ISO Speed",
	},

	// MeteringMode (0x9207)
	0x9207: {
		0:   "Unknown",
		1:   "Average",
		2:   "Center-weighted average",
		3:   "Spot",
		4:   "Multi-spot",
		5:   "Multi-segment",
		6:   "Partial",
		255: "Other",
	},

	// LightSource (0x9208)
	0x9208: {
		0:   "Unknown",
		1:   "Daylight",
		2:   "Fluorescent",
		3:   "Tungsten (Incandescent)",
		4:   "Flash",
		9:   "Fine Weather",
		10:  "Cloudy",
		11:  "Shade",
		12:  "Daylight Fluorescent",
		13:  "Day White Fluorescent",
		14:  "Cool White Fluorescent",
		15:  "White Fluorescent",
		16:  "Warm White Fluorescent",
		17:  "Standard Light A",
		18:  "Standard Light B",
		19:  "Standard Light C",
		20:  "D55",
		21:  "D65",
		22:  "D75",
		23:  "D50",
		24:  "ISO Studio Tungsten",
		255: "Other",
	},

	// Flash (0x9209) - bit field, handled specially
	0x9209: {
		0x00: "No Flash",
		0x01: "Fired",
		0x05: "Fired, Return not detected",
		0x07: "Fired, Return detected",
		0x08: "On, Did not fire",
		0x09: "On, Fired",
		0x0D: "On, Return not detected",
		0x0F: "On, Return detected",
		0x10: "Off, Did not fire",
		0x14: "Off, Did not fire, Return not detected",
		0x18: "Auto, Did not fire",
		0x19: "Auto, Fired",
		0x1D: "Auto, Fired, Return not detected",
		0x1F: "Auto, Fired, Return detected",
		0x20: "No flash function",
		0x30: "Off, No flash function",
		0x41: "Fired, Red-eye reduction",
		0x45: "Fired, Red-eye reduction, Return not detected",
		0x47: "Fired, Red-eye reduction, Return detected",
		0x49: "On, Red-eye reduction",
		0x4D: "On, Red-eye reduction, Return not detected",
		0x4F: "On, Red-eye reduction, Return detected",
		0x50: "Off, Red-eye reduction",
		0x58: "Auto, Did not fire, Red-eye reduction",
		0x59: "Auto, Fired, Red-eye reduction",
		0x5D: "Auto, Fired, Red-eye reduction, Return not detected",
		0x5F: "Auto, Fired, Red-eye reduction, Return detected",
	},

	// ColorSpace (0xA001)
	0xA001: {
		1:     "sRGB",
		2:     "Adobe RGB",
		65535: "Uncalibrated",
	},

	// FocalPlaneResolutionUnit (0xA210)
	0xA210: {
		1: "None",
		2: "inches",
		3: "centimeters",
		4: "millimeters",
		5: "micrometers",
	},

	// SensingMethod (0xA217)
	0xA217: {
		1: "Not defined",
		2: "One-chip color area",
		3: "Two-chip color area",
		4: "Three-chip color area",
		5: "Color sequential area",
		7: "Trilinear",
		8: "Color sequential linear",
	},

	// FileSource (0xA300)
	0xA300: {
		1: "Film Scanner",
		2: "Reflection Print Scanner",
		3: "Digital Camera",
	},

	// SceneType (0xA301)
	0xA301: {
		1: "Directly photographed",
	},

	// CustomRendered (0xA401)
	0xA401: {
		0: "Normal",
		1: "Custom",
		2: "HDR (no original saved)",
		3: "HDR (original saved)",
		4: "Original (for HDR)",
		6: "Panorama",
		7: "Portrait HDR",
		8: "Portrait",
	},

	// ExposureMode (0xA402)
	0xA402: {
		0: "Auto",
		1: "Manual",
		2: "Auto bracket",
	},

	// WhiteBalance (0xA403)
	0xA403: {
		0: "Auto",
		1: "Manual",
	},

	// SceneCaptureType (0xA406)
	0xA406: {
		0: "Standard",
		1: "Landscape",
		2: "Portrait",
		3: "Night",
		4: "Other",
	},

	// GainControl (0xA407)
	0xA407: {
		0: "None",
		1: "Low gain up",
		2: "High gain up",
		3: "Low gain down",
		4: "High gain down",
	},

	// Contrast (0xA408)
	0xA408: {
		0: "Normal",
		1: "Low",
		2: "High",
	},

	// Saturation (0xA409)
	0xA409: {
		0: "Normal",
		1: "Low",
		2: "High",
	},

	// Sharpness (0xA40A)
	0xA40A: {
		0: "Normal",
		1: "Soft",
		2: "Hard",
	},

	// SubjectDistanceRange (0xA40C)
	0xA40C: {
		0: "Unknown",
		1: "Macro",
		2: "Close",
		3: "Distant",
	},

	// CompositeImage (0xA460)
	0xA460: {
		0: "Unknown",
		1: "Not a Composite Image",
		2: "General Composite Image",
		3: "Composite Image Captured While Shooting",
	},
}

// decodeFlashValue provides detailed Flash decoding using bit fields
func decodeFlashValue(value uint16) string {
	// First check if we have an exact match
	if lookup := exifEnumValues[0x9209]; lookup != nil {
		if name, ok := lookup[value]; ok {
			return name
		}
	}

	// Otherwise decode bit by bit
	var parts []string

	// Bit 0: Flash fired
	if value&0x01 != 0 {
		parts = append(parts, "Fired")
	} else {
		parts = append(parts, "Did not fire")
	}

	// Bits 1-2: Return detection
	switch (value >> 1) & 0x03 {
	case 2:
		parts = append(parts, "Return not detected")
	case 3:
		parts = append(parts, "Return detected")
	}

	// Bits 3-4: Flash mode
	switch (value >> 3) & 0x03 {
	case 1:
		parts = append(parts, "On")
	case 2:
		parts = append(parts, "Off")
	case 3:
		parts = append(parts, "Auto")
	}

	// Bit 5: Flash function
	if value&0x20 != 0 {
		parts = append(parts, "No flash function")
	}

	// Bit 6: Red-eye reduction
	if value&0x40 != 0 {
		parts = append(parts, "Red-eye reduction")
	}

	if len(parts) == 0 {
		return fmt.Sprintf("Unknown (%d)", value)
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ", " + parts[i]
	}
	return result
}
