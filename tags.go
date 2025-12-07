package imx

// Common EXIF tag constants for easy access

// Camera and Device Tags
const (
	TagMake         TagID = "Exif:Make"
	TagModel        TagID = "Exif:Model"
	TagSoftware     TagID = "Exif:Software"
	TagOrientation  TagID = "Exif:Orientation"
	TagSerialNumber TagID = "Exif:SerialNumber"
	TagOwnerName    TagID = "Exif:OwnerName"
)

// Lens Tags
const (
	TagLensMake         TagID = "Exif:LensMake"
	TagLensModel        TagID = "Exif:LensModel"
	TagLensSerialNumber TagID = "Exif:LensSerialNumber"
	TagLensInfo         TagID = "Exif:LensInfo"
)

// Image Description Tags
const (
	TagImageDescription TagID = "Exif:ImageDescription"
	TagImageTitle       TagID = "Exif:ImageTitle"
	TagArtist           TagID = "Exif:Artist"
	TagPhotographer     TagID = "Exif:Photographer"
	TagCopyright        TagID = "Exif:Copyright"
	TagUserComment      TagID = "Exif:UserComment"
)

// Date and Time Tags
const (
	TagDateTimeOriginal    TagID = "Exif:DateTimeOriginal"
	TagCreateDate          TagID = "Exif:CreateDate"
	TagModifyDate          TagID = "Exif:ModifyDate"
	TagOffsetTime          TagID = "Exif:OffsetTime"
	TagOffsetTimeOriginal  TagID = "Exif:OffsetTimeOriginal"
	TagSubSecTime          TagID = "Exif:SubSecTime"
	TagSubSecTimeOriginal  TagID = "Exif:SubSecTimeOriginal"
	TagSubSecTimeDigitized TagID = "Exif:SubSecTimeDigitized"
)

// Image Dimensions Tags
const (
	TagImageWidth      TagID = "Exif:ImageWidth"
	TagImageHeight     TagID = "Exif:ImageHeight"
	TagExifImageWidth  TagID = "Exif:ExifImageWidth"
	TagExifImageHeight TagID = "Exif:ExifImageHeight"
)

// Exposure Tags
const (
	TagExposureTime         TagID = "Exif:ExposureTime"
	TagShutterSpeedValue    TagID = "Exif:ShutterSpeedValue"
	TagFNumber              TagID = "Exif:FNumber"
	TagApertureValue        TagID = "Exif:ApertureValue"
	TagExposureProgram      TagID = "Exif:ExposureProgram"
	TagExposureMode         TagID = "Exif:ExposureMode"
	TagExposureCompensation TagID = "Exif:ExposureCompensation"
	TagBrightnessValue      TagID = "Exif:BrightnessValue"
)

// ISO Tags
const (
	TagISO                       TagID = "Exif:ISO"
	TagISOSpeed                  TagID = "Exif:ISOSpeed"
	TagSensitivityType           TagID = "Exif:SensitivityType"
	TagStandardOutputSensitivity TagID = "Exif:StandardOutputSensitivity"
	TagRecommendedExposureIndex  TagID = "Exif:RecommendedExposureIndex"
)

// Focus and Lens Settings Tags
const (
	TagFocalLength             TagID = "Exif:FocalLength"
	TagFocalLengthIn35mmFormat TagID = "Exif:FocalLengthIn35mmFormat"
	TagMaxApertureValue        TagID = "Exif:MaxApertureValue"
	TagSubjectDistance         TagID = "Exif:SubjectDistance"
	TagSubjectDistanceRange    TagID = "Exif:SubjectDistanceRange"
)

// Flash Tags
const (
	TagFlash       TagID = "Exif:Flash"
	TagFlashEnergy TagID = "Exif:FlashEnergy"
)

// Metering and Lighting Tags
const (
	TagMeteringMode TagID = "Exif:MeteringMode"
	TagLightSource  TagID = "Exif:LightSource"
	TagWhiteBalance TagID = "Exif:WhiteBalance"
)

// Image Quality Tags
const (
	TagColorSpace       TagID = "Exif:ColorSpace"
	TagContrast         TagID = "Exif:Contrast"
	TagSaturation       TagID = "Exif:Saturation"
	TagSharpness        TagID = "Exif:Sharpness"
	TagDigitalZoomRatio TagID = "Exif:DigitalZoomRatio"
)

// Scene Tags
const (
	TagSceneCaptureType TagID = "Exif:SceneCaptureType"
	TagSceneType        TagID = "Exif:SceneType"
)

// GPS Tags
const (
	TagGPSVersionID         TagID = "Exif:GPSVersionID"
	TagGPSLatitudeRef       TagID = "Exif:GPSLatitudeRef"
	TagGPSLatitude          TagID = "Exif:GPSLatitude"
	TagGPSLongitudeRef      TagID = "Exif:GPSLongitudeRef"
	TagGPSLongitude         TagID = "Exif:GPSLongitude"
	TagGPSAltitudeRef       TagID = "Exif:GPSAltitudeRef"
	TagGPSAltitude          TagID = "Exif:GPSAltitude"
	TagGPSTimeStamp         TagID = "Exif:GPSTimeStamp"
	TagGPSDateStamp         TagID = "Exif:GPSDateStamp"
	TagGPSSatellites        TagID = "Exif:GPSSatellites"
	TagGPSStatus            TagID = "Exif:GPSStatus"
	TagGPSMeasureMode       TagID = "Exif:GPSMeasureMode"
	TagGPSDOP               TagID = "Exif:GPSDOP"
	TagGPSSpeed             TagID = "Exif:GPSSpeed"
	TagGPSSpeedRef          TagID = "Exif:GPSSpeedRef"
	TagGPSTrack             TagID = "Exif:GPSTrack"
	TagGPSTrackRef          TagID = "Exif:GPSTrackRef"
	TagGPSImgDirection      TagID = "Exif:GPSImgDirection"
	TagGPSImgDirectionRef   TagID = "Exif:GPSImgDirectionRef"
	TagGPSMapDatum          TagID = "Exif:GPSMapDatum"
	TagGPSDestLatitude      TagID = "Exif:GPSDestLatitude"
	TagGPSDestLatitudeRef   TagID = "Exif:GPSDestLatitudeRef"
	TagGPSDestLongitude     TagID = "Exif:GPSDestLongitude"
	TagGPSDestLongitudeRef  TagID = "Exif:GPSDestLongitudeRef"
	TagGPSDestBearing       TagID = "Exif:GPSDestBearing"
	TagGPSDestBearingRef    TagID = "Exif:GPSDestBearingRef"
	TagGPSDestDistance      TagID = "Exif:GPSDestDistance"
	TagGPSDestDistanceRef   TagID = "Exif:GPSDestDistanceRef"
	TagGPSProcessingMethod  TagID = "Exif:GPSProcessingMethod"
	TagGPSAreaInformation   TagID = "Exif:GPSAreaInformation"
	TagGPSDifferential      TagID = "Exif:GPSDifferential"
	TagGPSHPositioningError TagID = "Exif:GPSHPositioningError"
)

// Version Tags
const (
	TagExifVersion     TagID = "Exif:ExifVersion"
	TagFlashpixVersion TagID = "Exif:FlashpixVersion"
)

// Other Common Tags
const (
	TagCompression               TagID = "Exif:Compression"
	TagPhotometricInterpretation TagID = "Exif:PhotometricInterpretation"
	TagXResolution               TagID = "Exif:XResolution"
	TagYResolution               TagID = "Exif:YResolution"
	TagResolutionUnit            TagID = "Exif:ResolutionUnit"
	TagYCbCrPositioning          TagID = "Exif:YCbCrPositioning"
	TagRating                    TagID = "Exif:Rating"
	TagRatingPercent             TagID = "Exif:RatingPercent"
)
