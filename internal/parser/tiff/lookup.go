package tiff

// getTIFFTagName returns the name for a TIFF tag
func getTIFFTagName(tag uint16) string {
	name, ok := tiffTagNames[tag]
	if !ok {
		return ""
	}
	return name
}

// getEXIFTagName returns the name for an EXIF tag
func getEXIFTagName(tag uint16) string {
	name, ok := exifTagNames[tag]
	if !ok {
		return ""
	}
	return name
}

// getGPSTagName returns the name for a GPS tag
func getGPSTagName(tag uint16) string {
	name, ok := gpsTagNames[tag]
	if !ok {
		return ""
	}
	return name
}

// getInteropTagName returns the name for an Interoperability tag
func getInteropTagName(tag uint16) string {
	name, ok := interopTagNames[tag]
	if !ok {
		return ""
	}
	return name
}

// TIFF tag names (IFD0/IFD1)
var tiffTagNames = map[uint16]string{
	0x00FE: "NewSubfileType",
	0x0100: "ImageWidth",
	0x0101: "ImageHeight",
	0x0102: "BitsPerSample",
	0x0103: "Compression",
	0x0106: "PhotometricInterpretation",
	0x010A: "FillOrder",
	0x010E: "ImageDescription",
	0x010F: "Make",
	0x0110: "Model",
	0x0111: "StripOffsets",
	0x0112: "Orientation",
	0x0115: "SamplesPerPixel",
	0x0116: "RowsPerStrip",
	0x0117: "StripByteCounts",
	0x011A: "XResolution",
	0x011B: "YResolution",
	0x011C: "PlanarConfiguration",
	0x0128: "ResolutionUnit",
	0x0129: "PageNumber",
	0x0131: "Software",
	0x0132: "DateTime",
	0x013B: "Artist",
	0x013C: "HostComputer",
	0x013D: "Predictor",
	0x013E: "WhitePoint",
	0x013F: "PrimaryChromaticities",
	0x0142: "TileWidth",
	0x0143: "TileLength",
	0x0152: "ExtraSamples",
	0x0153: "SampleFormat",
	0x0201: "JPEGInterchangeFormat",
	0x0202: "JPEGInterchangeFormatLength",
	0x0211: "YCbCrCoefficients",
	0x0212: "YCbCrSubSampling",
	0x0213: "YCbCrPositioning",
	0x0214: "ReferenceBlackWhite",
	0x828D: "CFARepeatPatternDim",
	0x828E: "CFAPattern2",
	0x8298: "Copyright",
	0x8769: "ExifIFDPointer",
	0x8825: "GPSInfoIFDPointer",
	0x9003: "DateTimeOriginal",
	0x9216: "TIFF/EPStandardID",
	0x9217: "SensingMethod",
}

// EXIF tag names
var exifTagNames = map[uint16]string{
	0x829A: "ExposureTime",
	0x829D: "FNumber",
	0x8822: "ExposureProgram",
	0x8824: "SpectralSensitivity",
	0x8827: "ISOSpeedRatings",
	0x8828: "OECF",
	0x8830: "SensitivityType",
	0x8831: "StandardOutputSensitivity",
	0x8832: "RecommendedExposureIndex",
	0x8833: "ISOSpeed",
	0x8834: "ISOSpeedLatitudeyyy",
	0x8835: "ISOSpeedLatitudezzz",
	0x9000: "ExifVersion",
	0x9003: "DateTimeOriginal",
	0x9004: "DateTimeDigitized",
	0x9010: "OffsetTime",
	0x9011: "OffsetTimeOriginal",
	0x9012: "OffsetTimeDigitized",
	0x9101: "ComponentsConfiguration",
	0x9102: "CompressedBitsPerPixel",
	0x9201: "ShutterSpeedValue",
	0x9202: "ApertureValue",
	0x9203: "BrightnessValue",
	0x9204: "ExposureBiasValue",
	0x9205: "MaxApertureValue",
	0x9206: "SubjectDistance",
	0x9207: "MeteringMode",
	0x9208: "LightSource",
	0x9209: "Flash",
	0x920A: "FocalLength",
	0x9214: "SubjectArea",
	0x927C: "MakerNote",
	0x9286: "UserComment",
	0x9290: "SubSecTime",
	0x9291: "SubSecTimeOriginal",
	0x9292: "SubSecTimeDigitized",
	0xA000: "FlashpixVersion",
	0xA001: "ColorSpace",
	0xA002: "PixelXDimension",
	0xA003: "PixelYDimension",
	0xA004: "RelatedSoundFile",
	0xA005: "InteroperabilityIFDPointer",
	0xA20B: "FlashEnergy",
	0xA20C: "SpatialFrequencyResponse",
	0xA20E: "FocalPlaneXResolution",
	0xA20F: "FocalPlaneYResolution",
	0xA210: "FocalPlaneResolutionUnit",
	0xA214: "SubjectLocation",
	0xA215: "ExposureIndex",
	0xA217: "SensingMethod",
	0xA300: "FileSource",
	0xA301: "SceneType",
	0xA302: "CFAPattern",
	0xA401: "CustomRendered",
	0xA402: "ExposureMode",
	0xA403: "WhiteBalance",
	0xA404: "DigitalZoomRatio",
	0xA405: "FocalLengthIn35mmFilm",
	0xA406: "SceneCaptureType",
	0xA407: "GainControl",
	0xA408: "Contrast",
	0xA409: "Saturation",
	0xA40A: "Sharpness",
	0xA40B: "DeviceSettingDescription",
	0xA40C: "SubjectDistanceRange",
	0xA420: "ImageUniqueID",
	0xA430: "CameraOwnerName",
	0xA431: "BodySerialNumber",
	0xA432: "LensSpecification",
	0xA433: "LensMake",
	0xA434: "LensModel",
	0xA435: "LensSerialNumber",
	0xA460: "CompositeImage",
}

// GPS tag names
var gpsTagNames = map[uint16]string{
	0x0000: "GPSVersionID",
	0x0001: "GPSLatitudeRef",
	0x0002: "GPSLatitude",
	0x0003: "GPSLongitudeRef",
	0x0004: "GPSLongitude",
	0x0005: "GPSAltitudeRef",
	0x0006: "GPSAltitude",
	0x0007: "GPSTimeStamp",
	0x0008: "GPSSatellites",
	0x0009: "GPSStatus",
	0x000A: "GPSMeasureMode",
	0x000B: "GPSDOP",
	0x000C: "GPSSpeedRef",
	0x000D: "GPSSpeed",
	0x000E: "GPSTrackRef",
	0x000F: "GPSTrack",
	0x0010: "GPSImgDirectionRef",
	0x0011: "GPSImgDirection",
	0x0012: "GPSMapDatum",
	0x0013: "GPSDestLatitudeRef",
	0x0014: "GPSDestLatitude",
	0x0015: "GPSDestLongitudeRef",
	0x0016: "GPSDestLongitude",
	0x0017: "GPSDestBearingRef",
	0x0018: "GPSDestBearing",
	0x0019: "GPSDestDistanceRef",
	0x001A: "GPSDestDistance",
	0x001B: "GPSProcessingMethod",
	0x001C: "GPSAreaInformation",
	0x001D: "GPSDateStamp",
	0x001E: "GPSDifferential",
	0x001F: "GPSHPositioningError",
}

// Interoperability tag names
var interopTagNames = map[uint16]string{
	0x0001: "InteroperabilityIndex",
	0x0002: "InteroperabilityVersion",
	0x1000: "RelatedImageFileFormat",
	0x1001: "RelatedImageWidth",
	0x1002: "RelatedImageHeight",
}
