package icc

import (
	"encoding/binary"
	"fmt"
)

// Tag signatures - commonly used ICC tags
const (
	// Required tags
	TagProfileDescription = "desc" // profileDescriptionTag
	TagCopyright          = "cprt" // copyrightTag
	TagMediaWhitePoint    = "wtpt" // mediaWhitePointTag
	TagChromAdaptation    = "chad" // chromaticAdaptationTag

	// Display profile tags
	TagRedColorant   = "rXYZ" // redMatrixColumnTag
	TagGreenColorant = "gXYZ" // greenMatrixColumnTag
	TagBlueColorant  = "bXYZ" // blueMatrixColumnTag
	TagRedTRC        = "rTRC" // redTRCTag
	TagGreenTRC      = "gTRC" // greenTRCTag
	TagBlueTRC       = "bTRC" // blueTRCTag

	// Grayscale profile tags
	TagGrayTRC = "kTRC" // grayTRCTag

	// Lookup table tags
	TagAToB0 = "A2B0" // AToB0Tag
	TagAToB1 = "A2B1" // AToB1Tag
	TagAToB2 = "A2B2" // AToB2Tag
	TagBToA0 = "B2A0" // BToA0Tag
	TagBToA1 = "B2A1" // BToA1Tag
	TagBToA2 = "B2A2" // BToA2Tag
	TagGamut = "gamt" // gamutTag

	// Profile information tags
	TagCalibrationDateTime = "calt" // calibrationDateTimeTag
	TagCharTarget          = "targ" // charTargetTag
	TagDeviceMfgDesc       = "dmnd" // deviceMfgDescTag
	TagDeviceModelDesc     = "dmdd" // deviceModelDescTag
	TagLuminance           = "lumi" // luminanceTag
	TagMeasurement         = "meas" // measurementTag
	TagTechnology          = "tech" // technologyTag
	TagViewingCondDesc     = "vued" // viewingCondDescTag
	TagViewingConditions   = "view" // viewingConditionsTag

	// Named color tags
	TagNamedColor2 = "ncl2" // namedColor2Tag

	// Output profile tags
	TagOutputResponse = "resp" // outputResponseTag
	TagPreview0       = "pre0" // preview0Tag
	TagPreview1       = "pre1" // preview1Tag
	TagPreview2       = "pre2" // preview2Tag

	// Colorant tags
	TagColorantOrder    = "clro" // colorantOrderTag
	TagColorantTable    = "clrt" // colorantTableTag
	TagColorantTableOut = "clot" // colorantTableOutTag

	// Metadata tags (v4+)
	TagMetadata            = "meta" // metadataTag
	TagProfileSequenceDesc = "pseq" // profileSequenceDescTag
	TagProfileSequenceId   = "psid" // profileSequenceIdentifierTag
)

// Tag type signatures
const (
	TypeText                = "text" // textType
	TypeDesc                = "desc" // textDescriptionType
	TypeMLUC                = "mluc" // multiLocalizedUnicodeType
	TypeXYZ                 = "XYZ " // XYZType
	TypeCurve               = "curv" // curveType
	TypeParametricCurve     = "para" // parametricCurveType
	TypeSignature           = "sig " // signatureType
	TypeDateTime            = "dtim" // dateTimeType
	TypeMeasurement         = "meas" // measurementType
	TypeViewingConditions   = "view" // viewingConditionsType
	TypeLUT8                = "mft1" // lut8Type
	TypeLUT16               = "mft2" // lut16Type
	TypeLUTAToB             = "mAB " // lutAtoBType
	TypeLUTBToA             = "mBA " // lutBtoAType
	TypeNamedColor2         = "ncl2" // namedColor2Type
	TypeColorantOrder       = "clro" // colorantOrderType
	TypeColorantTable       = "clrt" // colorantTableType
	TypeS15Fixed16Array     = "sf32" // s15Fixed16ArrayType
	TypeU16Fixed16Array     = "uf32" // u16Fixed16ArrayType
	TypeChromaticity        = "chrm" // chromaticityType
	TypeCIEXYZ              = "XYZ " // Same as TypeXYZ
	TypeResponseCurveSet16  = "rcs2" // responseCurveSet16Type
	TypeDict                = "dict" // dictType (v5)
	TypeMultiProcessElement = "mpet" // multiProcessElementsType
)

// TagInfo contains parsed tag information
type TagInfo struct {
	Signature string
	Offset    uint32
	Size      uint32
	TypeSig   string
	Value     any
}

// parseTagTable parses the tag table from profile data
func parseTagTable(data []byte) ([]TagEntry, error) {
	if len(data) < HeaderSize+4 {
		return nil, fmt.Errorf("data too short for tag table")
	}

	// Tag count is at offset 128
	tagCount := binary.BigEndian.Uint32(data[128:132])

	// Sanity check
	if tagCount > 1000 {
		return nil, fmt.Errorf("unreasonable tag count: %d", tagCount)
	}

	// Each tag entry is 12 bytes: signature (4) + offset (4) + size (4)
	tableSize := 4 + int(tagCount)*12
	if len(data) < HeaderSize+tableSize {
		return nil, fmt.Errorf("data too short for %d tag entries", tagCount)
	}

	tags := make([]TagEntry, tagCount)
	offset := 132 // Start of first tag entry

	for i := uint32(0); i < tagCount; i++ {
		tags[i] = TagEntry{
			Signature: string(data[offset : offset+4]),
			Offset:    binary.BigEndian.Uint32(data[offset+4 : offset+8]),
			Size:      binary.BigEndian.Uint32(data[offset+8 : offset+12]),
		}
		offset += 12
	}

	return tags, nil
}

// knownTags maps tag signatures to human-readable names
var knownTags = map[string]string{
	"desc": "Profile Description",
	"cprt": "Profile Copyright",
	"wtpt": "Media White Point",
	"bkpt": "Media Black Point",
	"chad": "Chromatic Adaptation",
	"rXYZ": "Red Matrix Column",
	"gXYZ": "Green Matrix Column",
	"bXYZ": "Blue Matrix Column",
	"rTRC": "Red Tone Reproduction Curve",
	"gTRC": "Green Tone Reproduction Curve",
	"bTRC": "Blue Tone Reproduction Curve",
	"kTRC": "Gray Tone Reproduction Curve",
	"A2B0": "AToB0 (Perceptual)",
	"A2B1": "AToB1 (Colorimetric)",
	"A2B2": "AToB2 (Saturation)",
	"B2A0": "BToA0 (Perceptual)",
	"B2A1": "BToA1 (Colorimetric)",
	"B2A2": "BToA2 (Saturation)",
	"gamt": "Gamut",
	"calt": "Calibration Date/Time",
	"targ": "Characterization Target",
	"dmnd": "Device Manufacturer Description",
	"dmdd": "Device Model Description",
	"lumi": "Luminance",
	"meas": "Measurement",
	"tech": "Technology",
	"vued": "Viewing Conditions Description",
	"view": "Viewing Conditions",
	"ncl2": "Named Color 2",
	"resp": "Output Response",
	"pre0": "Preview 0",
	"pre1": "Preview 1",
	"pre2": "Preview 2",
	"clro": "Colorant Order",
	"clrt": "Colorant Table",
	"clot": "Colorant Table Out",
	"meta": "Metadata",
	"pseq": "Profile Sequence Description",
	"psid": "Profile Sequence Identifier",
	"cicp": "Coding-Independent Code Points",
	"ciis": "Colorimetric Intent Image State",
	"ciin": "Colorimetric Intent Image Name",
}

// getTagName returns the human-readable name for a tag signature
func getTagName(sig string) string {
	if name, ok := knownTags[sig]; ok {
		return name
	}
	return sig
}

// technologySignatures maps technology signature values to names
var technologySignatures = map[uint32]string{
	0x66736E20: "Film Scanner",
	0x64636D20: "Digital Camera",
	0x7273636E: "Reflective Scanner",
	0x696A6574: "Ink Jet Printer",
	0x74776178: "Thermal Wax Printer",
	0x65706879: "Electrophotographic Printer",
	0x65737461: "Electrostatic Printer",
	0x64737562: "Dye Sublimation Printer",
	0x7270686F: "Photographic Paper Printer",
	0x6670726E: "Film Writer",
	0x766964C6: "Video Monitor",
	0x76696463: "Video Camera",
	0x706A7476: "Projection Television",
	0x43525420: "Cathode Ray Tube Display",
	0x504D4420: "Passive Matrix Display",
	0x414D4420: "Active Matrix Display",
	0x4C434420: "LCD Display",
	0x4F4C4544: "OLED Display",
	0x4C454420: "LED Display",
	0x6770686F: "Gravure",
	0x6F666673: "Offset Lithography",
	0x73696C6B: "Silkscreen",
	0x666C6578: "Flexography",
	0x6D706673: "Motion Picture Film Scanner",
	0x6D706672: "Motion Picture Film Recorder",
	0x646D7063: "Digital Motion Picture Camera",
	0x64637067: "Digital Cinema Projector",
}

// getTechnologyName returns the name for a technology signature
func getTechnologyName(sig uint32) string {
	if name, ok := technologySignatures[sig]; ok {
		return name
	}
	return signatureToString(sig)
}
