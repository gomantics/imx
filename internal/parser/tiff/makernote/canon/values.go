package canon

// Canon MakerNote value decoding
// Reference: https://exiftool.org/TagNames/Canon.html

// CameraSettings array indices (tag 0x0001)
const (
	csQuality       = 3
	csFlashMode     = 4
	csDriveMode     = 5
	csFocusMode     = 7
	csRecordMode    = 9
	csImageSize     = 10
	csEasyMode      = 11
	csDigitalZoom   = 12
	csContrast      = 13
	csSaturation    = 14
	csSharpness     = 15
	csISOSpeed      = 16
	csMeteringMode  = 17
	csFocusType     = 18
	csAFPoint       = 19
	csExposureMode  = 20
	csLensType      = 22
	csMaxFocalLen   = 23
	csMinFocalLen   = 24
	csFlashActivity = 28
	csFlashBits     = 29
	csFocusCont     = 32
	csAEBSequence   = 33
	csImageStab     = 34
)

// ShotInfo array indices (tag 0x0004)
const (
	siAutoISO        = 1
	siBaseISO        = 2
	siMeasuredEV     = 3
	siTargetAperture = 4
	siTargetExposure = 5
	siExposureComp   = 6
	siWhiteBalance   = 7
	siSlowShutter    = 8
	siSequenceNumber = 9
	siFlashGuideNum  = 13
	siAFPointUsed    = 14
	siFlashExposComp = 15
	siAutoExposBrack = 16
	siAEBBracketVal  = 17
	siControlMode    = 18
)

// Quality values for CameraSettings[3]
var qualityValues = map[uint16]string{
	0:  "Unknown",
	1:  "Economy",
	2:  "Normal",
	3:  "Fine",
	4:  "RAW",
	5:  "Superfine",
	7:  "CRAW",
	130: "Light (RAW)",
	131: "Standard (RAW)",
}

// FlashMode values for CameraSettings[4]
var flashModeValues = map[uint16]string{
	0:  "Off",
	1:  "Auto",
	2:  "On",
	3:  "Red-eye reduction",
	4:  "Slow-sync",
	5:  "Red-eye reduction (Auto)",
	6:  "Red-eye reduction (On)",
	16: "External flash",
}

// DriveMode values for CameraSettings[5]
var driveModeValues = map[uint16]string{
	0:  "Single",
	1:  "Continuous",
	2:  "Movie",
	3:  "Continuous, Speed Priority",
	4:  "Continuous, Low",
	5:  "Continuous, High",
	6:  "Silent Single",
	9:  "Single, Silent",
	10: "Continuous, Silent",
}

// FocusMode values for CameraSettings[7]
var focusModeValues = map[uint16]string{
	0:   "One-shot AF",
	1:   "AI Servo AF",
	2:   "AI Focus AF",
	3:   "Manual Focus",
	4:   "Single",
	5:   "Continuous",
	6:   "Manual Focus",
	16:  "Pan Focus",
	256: "AF + MF",
	512: "Movie Snap Focus",
	519: "Movie Servo AF",
}

// RecordMode values for CameraSettings[9]
var recordModeValues = map[uint16]string{
	1:  "JPEG",
	2:  "CRW+THM",
	3:  "AVI+THM",
	4:  "TIF",
	5:  "TIF+JPEG",
	6:  "CR2",
	7:  "CR2+JPEG",
	9:  "MOV",
	10: "MP4",
	11: "CRM",
	12: "CR3",
	13: "CR3+JPEG",
	14: "HIF",
	15: "CR3+HIF",
}

// ImageSize values for CameraSettings[10]
var imageSizeValues = map[uint16]string{
	0:     "Large",
	1:     "Medium",
	2:     "Small",
	5:     "Medium 1",
	6:     "Medium 2",
	7:     "Medium 3",
	8:     "Postcard",
	9:     "Widescreen",
	10:    "Medium Widescreen",
	14:    "Small 1",
	15:    "Small 2",
	16:    "Small 3",
	128:   "640x480 Movie",
	129:   "Medium Movie",
	130:   "Small Movie",
	137:   "1280x720 Movie",
	142:   "1920x1080 Movie",
	65535: "Unknown",
}

// EasyMode values for CameraSettings[11]
var easyModeValues = map[uint16]string{
	0:  "Full auto",
	1:  "Manual",
	2:  "Landscape",
	3:  "Fast shutter",
	4:  "Slow shutter",
	5:  "Night",
	6:  "Gray Scale",
	7:  "Sepia",
	8:  "Portrait",
	9:  "Sports",
	10: "Macro",
	11: "Black & White",
	12: "Pan focus",
	13: "Vivid",
	14: "Neutral",
	15: "Flash Off",
	16: "Long Shutter",
	17: "Super Macro",
	18: "Foliage",
	19: "Indoor",
	20: "Fireworks",
	21: "Beach",
	22: "Underwater",
	23: "Snow",
	24: "Kids & Pets",
	25: "Night Snapshot",
	26: "Digital Macro",
	27: "My Colors",
	28: "Movie Snap",
	29: "Super Macro 2",
	30: "Color Accent",
	31: "Color Swap",
	32: "Aquarium",
	33: "ISO 3200",
	34: "ISO 6400",
	35: "Creative Light Effect",
	36: "Easy",
	37: "Quick Shot",
	38: "Creative Auto",
	39: "Zoom Blur",
	40: "Low Light",
	41: "Nostalgic",
	42: "Super Vivid",
	43: "Poster Effect",
	44: "Face Self-timer",
	45: "Smile",
	46: "Wink Self-timer",
	47: "Fisheye Effect",
	48: "Miniature Effect",
	49: "High-speed Burst",
	50: "Best Image Selection",
	51: "High Dynamic Range",
	52: "Handheld Night Scene",
	53: "Movie Digest",
	54: "Live View Control",
	55: "Discreet",
	56: "Blur Reduction",
	57: "Monochrome",
	58: "Toy Camera Effect",
	59: "Scene Intelligent Auto",
	60: "High-speed Burst HQ",
	61: "Smooth Skin",
	62: "Soft Focus",
	68: "Food",
	84: "HDR Art Standard",
	85: "HDR Art Vivid",
	93: "HDR Art Bold",
}

// MeteringMode values for CameraSettings[17]
var meteringModeValues = map[uint16]string{
	0: "Default",
	1: "Spot",
	2: "Average",
	3: "Evaluative",
	4: "Partial",
	5: "Center-weighted average",
}

// FocusType values for CameraSettings[18]
var focusTypeValues = map[uint16]string{
	0:     "Manual",
	1:     "Auto",
	2:     "Not Known",
	3:     "Macro",
	4:     "Very Close",
	5:     "Close",
	6:     "Middle Range",
	7:     "Far Range",
	8:     "Pan Focus",
	9:     "Super Macro",
	10:    "Infinity",
	65535: "Unknown",
}

// AFPoint values for CameraSettings[19]
var afPointValues = map[uint16]string{
	0x2005: "Manual AF point selection",
	0x3000: "None (MF)",
	0x3001: "Auto AF point selection",
	0x3002: "Right",
	0x3003: "Center",
	0x3004: "Left",
	0x4001: "Auto AF point selection",
	0x4006: "Face Detect",
}

// ExposureMode values for CameraSettings[20]
var exposureModeValues = map[uint16]string{
	0: "Easy",
	1: "Program AE",
	2: "Shutter speed priority AE",
	3: "Aperture-priority AE",
	4: "Manual",
	5: "Depth-of-field AE",
	6: "M-Dep",
	7: "Bulb",
	8: "Flexible-priority AE",
}

// ImageStabilization values for CameraSettings[34]
var imageStabilizationValues = map[uint16]string{
	0:     "Off",
	1:     "On",
	2:     "Shoot Only",
	3:     "Panning",
	4:     "Dynamic",
	256:   "Off (2)",
	257:   "On (2)",
	258:   "Shoot Only (2)",
	259:   "Panning (2)",
	260:   "Dynamic (2)",
	65535: "Unknown",
}

// WhiteBalance values for ShotInfo[7]
var whiteBalanceValues = map[uint16]string{
	0:   "Auto",
	1:   "Daylight",
	2:   "Cloudy",
	3:   "Tungsten",
	4:   "Fluorescent",
	5:   "Flash",
	6:   "Custom",
	7:   "Black & White",
	8:   "Shade",
	9:   "Manual Temperature (Kelvin)",
	10:  "PC Set1",
	11:  "PC Set2",
	12:  "PC Set3",
	14:  "Daylight Fluorescent",
	15:  "Custom 1",
	16:  "Custom 2",
	17:  "Underwater",
	18:  "Custom 3",
	19:  "Custom 4",
	20:  "PC Set4",
	21:  "PC Set5",
	23:  "Auto (Ambience Priority)",
}

// ModelID values (tag 0x0010)
var modelIDValues = map[uint32]string{
	0x80000001: "EOS-1D",
	0x80000167: "EOS-1DS",
	0x80000168: "EOS 10D",
	0x80000169: "EOS-1D Mark III",
	0x80000170: "EOS Digital Rebel / 300D / Kiss Digital",
	0x80000174: "EOS-1D Mark II",
	0x80000175: "EOS 20D",
	0x80000176: "EOS Digital Rebel XSi / 450D / Kiss X2",
	0x80000188: "EOS-1Ds Mark II",
	0x80000189: "EOS Digital Rebel XT / 350D / Kiss Digital N",
	0x80000190: "EOS 40D",
	0x80000213: "EOS 5D",
	0x80000215: "EOS-1Ds Mark III",
	0x80000218: "EOS 5D Mark II",
	0x80000219: "WFT-E1",
	0x80000232: "EOS-1D Mark II N",
	0x80000234: "EOS 30D",
	0x80000236: "EOS Digital Rebel XTi / 400D / Kiss Digital X",
	0x80000250: "EOS 7D",
	0x80000252: "EOS Rebel T1i / 500D / Kiss X3",
	0x80000254: "EOS Rebel XS / 1000D / Kiss F",
	0x80000261: "EOS 50D",
	0x80000269: "EOS-1D X",
	0x80000270: "EOS Rebel T2i / 550D / Kiss X4",
	0x80000271: "WFT-E2",
	0x80000273: "WFT-E3",
	0x80000281: "EOS-1D Mark IV",
	0x80000285: "EOS 5D Mark III",
	0x80000286: "EOS Rebel T3i / 600D / Kiss X5",
	0x80000287: "EOS 60D",
	0x80000288: "EOS Rebel T3 / 1100D / Kiss X50",
	0x80000289: "EOS 7D Mark II",
	0x80000297: "WFT-E4",
	0x80000298: "WFT-E5",
	0x80000301: "EOS Rebel T4i / 650D / Kiss X6i",
	0x80000302: "EOS 6D",
	0x80000324: "EOS-1D C",
	0x80000325: "EOS 70D",
	0x80000326: "EOS Rebel T5i / 700D / Kiss X7i",
	0x80000327: "EOS Rebel T5 / 1200D / Kiss X70",
	0x80000328: "EOS-1D X Mark II",
	0x80000331: "EOS M",
	0x80000346: "EOS Rebel SL1 / 100D / Kiss X7",
	0x80000347: "EOS Rebel T6s / 760D / 8000D",
	0x80000349: "EOS 5D Mark IV",
	0x80000350: "EOS 80D",
	0x80000355: "EOS M2",
	0x80000382: "EOS 5DS",
	0x80000393: "EOS Rebel T6i / 750D / Kiss X8i",
	0x80000401: "EOS 5DS R",
	0x80000404: "EOS Rebel T6 / 1300D / Kiss X80",
	0x80000405: "EOS Rebel T7i / 800D / Kiss X9i",
	0x80000406: "EOS 6D Mark II",
	0x80000408: "EOS 77D / 9000D",
	0x80000417: "EOS Rebel SL2 / 200D / Kiss X9",
	0x80000421: "EOS R5",
	0x80000422: "EOS Rebel T100 / 4000D / 3000D",
	0x80000424: "EOS R",
	0x80000428: "EOS-1D X Mark III",
	0x80000432: "EOS Rebel T7 / 2000D / 1500D / Kiss X90",
	0x80000433: "EOS RP",
	0x80000435: "EOS Rebel T8i / 850D / Kiss X10i",
	0x80000436: "EOS SL3 / 250D / Kiss X10",
	0x80000437: "EOS 90D",
	0x80000450: "EOS R3",
	0x80000453: "EOS R6",
	0x80000464: "EOS R7",
	0x80000465: "EOS R10",
	0x80000467: "EOS R6 Mark II",
	0x80000468: "EOS R8",
	0x80000480: "EOS R50",
	0x80000481: "EOS R100",
	0x80000487: "EOS R5 Mark II",
	0x80000491: "EOS R1",
}

// decodeCameraSettingsValue decodes a single value from CameraSettings array
func decodeCameraSettingsValue(index int, value uint16) (string, string) {
	switch index {
	case csQuality:
		if name, ok := qualityValues[value]; ok {
			return "Quality", name
		}
		return "Quality", ""
	case csFlashMode:
		if name, ok := flashModeValues[value]; ok {
			return "FlashMode", name
		}
		return "FlashMode", ""
	case csDriveMode:
		if name, ok := driveModeValues[value]; ok {
			return "DriveMode", name
		}
		return "DriveMode", ""
	case csFocusMode:
		if name, ok := focusModeValues[value]; ok {
			return "FocusMode", name
		}
		return "FocusMode", ""
	case csRecordMode:
		if name, ok := recordModeValues[value]; ok {
			return "RecordMode", name
		}
		return "RecordMode", ""
	case csImageSize:
		if name, ok := imageSizeValues[value]; ok {
			return "ImageSize", name
		}
		return "ImageSize", ""
	case csEasyMode:
		if name, ok := easyModeValues[value]; ok {
			return "EasyMode", name
		}
		return "EasyMode", ""
	case csMeteringMode:
		if name, ok := meteringModeValues[value]; ok {
			return "MeteringMode", name
		}
		return "MeteringMode", ""
	case csFocusType:
		if name, ok := focusTypeValues[value]; ok {
			return "FocusType", name
		}
		return "FocusType", ""
	case csAFPoint:
		if name, ok := afPointValues[value]; ok {
			return "AFPoint", name
		}
		return "AFPoint", ""
	case csExposureMode:
		if name, ok := exposureModeValues[value]; ok {
			return "ExposureMode", name
		}
		return "ExposureMode", ""
	case csImageStab:
		if name, ok := imageStabilizationValues[value]; ok {
			return "ImageStabilization", name
		}
		return "ImageStabilization", ""
	default:
		return "", ""
	}
}

// decodeShotInfoValue decodes a single value from ShotInfo array
func decodeShotInfoValue(index int, value uint16) (string, string) {
	switch index {
	case siWhiteBalance:
		if name, ok := whiteBalanceValues[value]; ok {
			return "WhiteBalance", name
		}
		return "WhiteBalance", ""
	default:
		return "", ""
	}
}

// decodeModelID returns the camera model name for a ModelID value
func decodeModelID(value uint32) string {
	if name, ok := modelIDValues[value]; ok {
		return name
	}
	return ""
}
