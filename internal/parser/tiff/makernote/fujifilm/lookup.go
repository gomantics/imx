package fujifilm

// fujifilmTagNames maps Fujifilm MakerNote tag IDs to human-readable names.
// Based on ExifTool Fujifilm tag documentation.
var fujifilmTagNames = map[uint16]string{
	// Version and basic info
	0x0000: "Version",
	0x0010: "SerialNumber",
	0x1000: "Quality",
	0x1001: "Sharpness",
	0x1002: "WhiteBalance",
	0x1003: "Saturation",
	0x1004: "Contrast",
	0x1005: "ColorTemperature",
	0x1006: "Contrast2",
	0x100a: "WhiteBalanceFineTune",
	0x100b: "NoiseReduction",
	0x100e: "HighISONoiseReduction",

	// Focus settings
	0x1010: "FlashMode",
	0x1011: "FlashExposureComp",
	0x1020: "Macro",
	0x1021: "FocusMode",
	0x1022: "AFMode",
	0x1023: "FocusPixel",
	0x102b: "PrioritySettings",
	0x102d: "FocusSettings",
	0x102e: "AFCSettings",
	0x1030: "SlowSync",
	0x1031: "PictureMode",
	0x1032: "ExposureCount",
	0x1033: "EXRAuto",
	0x1034: "EXRMode",

	// Image processing
	0x1040: "ShadowTone",
	0x1041: "HighlightTone",
	0x1044: "DigitalZoom",
	0x1045: "LensModulationOptimizer",
	0x1047: "GrainEffect",
	0x1048: "ColorChromeEffect",
	0x1049: "BWAdjustment",
	0x104d: "CropMode",
	0x104e: "ColorChromeFXBlue",

	// Lens info
	0x1050: "ShutterType",
	0x1100: "AutoBracketing",
	0x1101: "SequenceNumber",
	0x1103: "DriveSettings",
	0x1105: "PixelShiftShots",
	0x1106: "PixelShiftOffset",

	// Advanced settings
	0x1153: "PanoramaAngle",
	0x1154: "PanoramaDirection",
	0x1201: "AdvancedFilter",

	// Color and tone
	0x1210: "ColorMode",
	0x1300: "BlurWarning",
	0x1301: "FocusWarning",
	0x1302: "ExposureWarning",
	0x1304: "GEImageSize",
	0x1400: "DynamicRange",
	0x1401: "FilmMode",
	0x1402: "DynamicRangeSetting",
	0x1403: "DevelopmentDynamicRange",
	0x1404: "MinFocalLength",
	0x1405: "MaxFocalLength",
	0x1406: "MaxApertureAtMinFocal",
	0x1407: "MaxApertureAtMaxFocal",
	0x1408: "AutoDynamicRange",
	0x1409: "ImageStabilization",
	0x140b: "SceneRecognition",
	0x1422: "Rating",
	0x1425: "ImageCount",
	0x1431: "FlickerReduction",
	0x1436: "VideoRecordingMode",
	0x1438: "PeripheralLighting",
	0x1439: "VideoCompression",
	0x1443: "DRangePriority",
	0x1444: "DRangePriorityAuto",
	0x1445: "DRangePriorityFixed",
	0x1446: "FlickerReductionLevel",

	// Face detection
	0x4100: "FacesDetected",
	0x4103: "FacePositions",
	0x4200: "NumFaceElements",
	0x4201: "FaceElementTypes",
	0x4203: "FaceElementPositions",
	0x4282: "FaceRecInfo",

	// RAW development
	0x8000: "FileSource",
	0x8002: "OrderNumber",
	0x8003: "FrameNumber",

	// Additional tags
	0xb211: "Parallax",
}
