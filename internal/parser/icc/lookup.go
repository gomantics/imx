package icc

import "io"

// tagNames maps ICC tag signatures to their descriptive names.
// Reference: ICC.1:2022 specification, Section 9 (Tag definitions)
var tagNames = map[string]string{
	// Profile information tags
	"desc": "ProfileDescription",
	"cprt": "ProfileCopyright",
	"dmnd": "DeviceManufacturerDescription",
	"dmdd": "DeviceModelDescription",

	// Color space tags
	"wtpt": "MediaWhitePoint",
	"bkpt": "MediaBlackPoint",
	"rXYZ": "RedMatrixColumn",
	"gXYZ": "GreenMatrixColumn",
	"bXYZ": "BlueMatrixColumn",
	"kTRC": "GrayToneReproductionCurve",
	"rTRC": "RedToneReproductionCurve",
	"gTRC": "GreenToneReproductionCurve",
	"bTRC": "BlueToneReproductionCurve",

	// Rendering tags
	"chad": "ChromaticAdaptation",
	"chrm": "Chromaticity",
	"clro": "ColorantOrder",
	"clrt": "ColorantTable",
	"clot": "ColorantTableOut",

	// Measurement and viewing conditions
	"meas": "Measurement",
	"view": "ViewingConditions",
	"vued": "ViewingConditionsDescription",
	"lumi": "Luminance",

	// Device settings
	"tech": "Technology",
	"devs": "DeviceSettings",

	// Profile connection space transforms
	"A2B0": "AToB0Perceptual",
	"A2B1": "AToB1Colorimetric",
	"A2B2": "AToB2Saturation",
	"B2A0": "BToA0Perceptual",
	"B2A1": "BToA1Colorimetric",
	"B2A2": "BToA2Saturation",
	"gamt": "Gamut",
	"pre0": "Preview0",
	"pre1": "Preview1",
	"pre2": "Preview2",

	// Color rendering dictionary (CRD)
	"ps2s": "PostScript2CSA",
	"ps2i": "PostScript2CRD",

	// Calibration tags
	"calt": "CalibrationDateTime",
	"targ": "CharacterizationTarget",

	// Screen tags
	"scrd": "ScreeningDescription",
	"scrn": "Screening",

	// Other tags
	"bfd ": "UCRBG",
	"pseq": "ProfileSequenceDescription",
	"psid": "ProfileSequenceIdentifier",

	// Named color tags
	"ncol": "NamedColor",
	"ncl2": "NamedColor2",

	// Metadata tags
	"meta": "Metadata",

	// Output response
	"resp": "OutputResponse",

	// Colorimetric intent image state
	"ciis": "ColorimetricIntentImageState",
	"ciin": "ColorimetricIntentImageName",

	// Rendering intent gamut
	"rig0": "PerceptualRenderingIntentGamut",
	"rig2": "SaturationRenderingIntentGamut",

	// Coding independent code points
	"cicp": "CodingIndependentCodePoints",
}

// getTagName returns the descriptive name for a tag signature.
// If the signature is not found, returns the raw signature (4-character code).
// This ensures unknown tags are still identifiable by their signature.
func getTagName(signature string) string {
	if name, ok := tagNames[signature]; ok {
		return name
	}
	return signature
}

// TypeConverter is a function that converts raw tag data to a meaningful value.
type TypeConverter func(r io.ReaderAt, tag TagRecord) (any, error)

// typeConverters maps ICC type signatures to their conversion functions.
var typeConverters = map[string]TypeConverter{
	"text": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseTextType(r, tag) },
	"desc": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseDescType(r, tag) },
	"mluc": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseMlucType(r, tag) },
	"XYZ ": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseXYZType(r, tag) },
	"sf32": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseS15Fixed16Type(r, tag) },
	"uf32": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseU16Fixed16ArrayType(r, tag) },
	"sig ": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseSigType(r, tag) },
	"curv": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseCurvType(r, tag) },
	"para": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseParametricCurveType(r, tag) },
	"dtim": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseDateTimeType(r, tag) },
	"meas": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseMeasurementType(r, tag) },
	"view": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseViewingConditionsType(r, tag) },
	"chrm": func(r io.ReaderAt, tag TagRecord) (any, error) { return parseChromaticityType(r, tag) },
}

// getTypeConverter returns the converter function for a type signature.
// If the type is unknown, returns a default converter that returns raw bytes.
//
// Supported type signatures (13 total):
//   - text, desc, mluc: Text variants
//   - XYZ, sf32, uf32: Numeric arrays
//   - sig: Signature/technology
//   - curv, para: Curves
//   - dtim: Date/time
//   - meas, view, chrm: Measurement/viewing/chromaticity
//
// Unknown types return their complete raw bytes for forward compatibility.
func getTypeConverter(typeSignature string) TypeConverter {
	if converter, ok := typeConverters[typeSignature]; ok {
		return converter
	}
	// Return default converter for unknown types that returns raw bytes
	return func(r io.ReaderAt, tag TagRecord) (any, error) {
		buf := make([]byte, tag.Size)
		_, err := r.ReadAt(buf, int64(tag.Offset))
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
}

// profileClassNames maps profile class signatures to human-readable names.
var profileClassNames = map[string]string{
	"scnr": "Input Device Profile",          // Input device (scanner)
	"mntr": "Display Device Profile",        // Display device (monitor)
	"prtr": "Output Device Profile",         // Output device (printer)
	"link": "DeviceLink Profile",            // Device link
	"abst": "Abstract Profile",              // Abstract profile
	"spac": "ColorSpace Conversion Profile", // Color space conversion
	"nmcl": "Named Color Profile",           // Named color
}

// getProfileClassName returns the human-readable name for a profile class signature.
func getProfileClassName(sig string) string {
	if name, ok := profileClassNames[sig]; ok {
		return name
	}
	return sig
}

// colorSpaceNames maps color space signatures to human-readable names.
var colorSpaceNames = map[string]string{
	"XYZ ": "XYZ",
	"Lab ": "Lab",
	"Luv ": "Luv",
	"YCbr": "YCbCr",
	"Yxy ": "Yxy",
	"RGB ": "RGB",
	"GRAY": "Grayscale",
	"HSV ": "HSV",
	"HLS ": "HLS",
	"CMYK": "CMYK",
	"CMY ": "CMY",
	"2CLR": "2Color",
	"3CLR": "3Color",
	"4CLR": "4Color",
	"5CLR": "5Color",
	"6CLR": "6Color",
	"7CLR": "7Color",
	"8CLR": "8Color",
	"9CLR": "9Color",
	"ACLR": "10Color",
	"BCLR": "11Color",
	"CCLR": "12Color",
	"DCLR": "13Color",
	"ECLR": "14Color",
	"FCLR": "15Color",
}

// getColorSpaceName returns the human-readable name for a color space signature.
func getColorSpaceName(sig string) string {
	if name, ok := colorSpaceNames[sig]; ok {
		return name
	}
	return sig
}

// platformNames maps platform signatures to human-readable names.
var platformNames = map[string]string{
	"APPL": "Apple",
	"MSFT": "Microsoft",
	"SGI ": "SiliconGraphics",
	"SUNW": "SunMicrosystems",
	"TGNT": "Taligent",
}

// getPlatformName returns the human-readable name for a platform signature.
func getPlatformName(sig string) string {
	if sig == "\x00\x00\x00\x00" {
		return "Unspecified"
	}
	if name, ok := platformNames[sig]; ok {
		return name
	}
	return sig
}

// renderingIntentNames maps rendering intent values to human-readable names.
var renderingIntentNames = map[uint32]string{
	0: "Perceptual",
	1: "MediaRelativeColorimetric",
	2: "Saturation",
	3: "ICCAbsoluteColorimetric",
}

// getRenderingIntentName returns the human-readable name for a rendering intent value.
func getRenderingIntentName(intent uint32) string {
	if name, ok := renderingIntentNames[intent]; ok {
		return name
	}
	return "Unknown"
}

// technologySignatures maps technology signature values to names.
var technologySignatures = map[uint32]string{
	0x66736E20: "FilmScanner",
	0x64636D20: "DigitalCamera",
	0x7273636E: "ReflectiveScanner",
	0x696A6574: "InkJetPrinter",
	0x74776178: "ThermalWaxPrinter",
	0x65706879: "ElectrophotographicPrinter",
	0x65737461: "ElectrostaticPrinter",
	0x64737562: "DyeSublimationPrinter",
	0x7270686F: "PhotographicPaperPrinter",
	0x6670726E: "FilmWriter",
	0x7669646C: "VideoMonitor",
	0x76696463: "VideoCamera",
	0x706A7476: "ProjectionTelevision",
	0x43525420: "CathodeRayTubeDisplay",
	0x504D4420: "PassiveMatrixDisplay",
	0x414D4420: "ActiveMatrixDisplay",
	0x4C434420: "LCDDisplay",
	0x4F4C4544: "OLEDDisplay",
	0x4C454420: "LEDDisplay",
	0x6770686F: "Gravure",
	0x6F666673: "OffsetLithography",
	0x73696C6B: "Silkscreen",
	0x666C6578: "Flexography",
	0x6D706673: "MotionPictureFilmScanner",
	0x6D706672: "MotionPictureFilmRecorder",
	0x646D7063: "DigitalMotionPictureCamera",
	0x64637067: "DigitalCinemaProjector",
}

// getTechnologyName returns the name for a technology signature.
func getTechnologyName(sig uint32) string {
	if name, ok := technologySignatures[sig]; ok {
		return name
	}
	// Return signature as string
	return string([]byte{byte(sig >> 24), byte(sig >> 16), byte(sig >> 8), byte(sig)})
}

// getProfileFlagsName returns human-readable profile flags.
func getProfileFlagsName(flags uint32) string {
	var parts []string

	// Bit 0: Embedded profile
	if flags&0x01 != 0 {
		parts = append(parts, "Embedded")
	} else {
		parts = append(parts, "Not Embedded")
	}

	// Bit 1: Profile can be used independently
	if flags&0x02 == 0 {
		parts = append(parts, "Independent")
	} else {
		parts = append(parts, "Cannot be used independently")
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ", " + parts[i]
	}
	return result
}

// getDeviceAttributesName returns human-readable device attributes.
func getDeviceAttributesName(attrs uint64) string {
	var parts []string

	// Bit 0: Reflective (0) or Transmissive (1)
	if attrs&0x01 == 0 {
		parts = append(parts, "Reflective")
	} else {
		parts = append(parts, "Transmissive")
	}

	// Bit 1: Glossy (0) or Matte (1)
	if attrs&0x02 == 0 {
		parts = append(parts, "Glossy")
	} else {
		parts = append(parts, "Matte")
	}

	// Bit 2: Positive (0) or Negative (1)
	if attrs&0x04 == 0 {
		parts = append(parts, "Positive")
	} else {
		parts = append(parts, "Negative")
	}

	// Bit 3: Color (0) or Black & White (1)
	if attrs&0x08 == 0 {
		parts = append(parts, "Color")
	} else {
		parts = append(parts, "Black & White")
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ", " + parts[i]
	}
	return result
}
