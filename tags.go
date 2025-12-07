package imx

// =============================================================================
// EXIF Tags
// =============================================================================

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

// =============================================================================
// IPTC Tags
// =============================================================================

// IPTC Core Identification Tags
const (
	TagIPTCObjectName    TagID = "IPTC:ObjectName"    // Title/headline reference
	TagIPTCUrgency       TagID = "IPTC:Urgency"       // 1=most urgent, 8=least
	TagIPTCCategory      TagID = "IPTC:Category"      // Subject category code
	TagIPTCKeywords      TagID = "IPTC:Keywords"      // Keywords for indexing
	TagIPTCFixtureID     TagID = "IPTC:FixtureID"     // Identifies recurring events
	TagIPTCEditStatus    TagID = "IPTC:EditStatus"    // Status of object
	TagIPTCSpecialInstr  TagID = "IPTC:SpecialInstr"  // Special instructions
)

// IPTC Date/Time Tags
const (
	TagIPTCDateCreated        TagID = "IPTC:DateCreated"        // Intellectual content created
	TagIPTCTimeCreated        TagID = "IPTC:TimeCreated"        // Time content created
	TagIPTCDigitalCreationDate TagID = "IPTC:DigitalCreationDate" // Digital file created
	TagIPTCDigitalCreationTime TagID = "IPTC:DigitalCreationTime" // Digital file time
	TagIPTCReleaseDate        TagID = "IPTC:ReleaseDate"        // Earliest release date
	TagIPTCReleaseTime        TagID = "IPTC:ReleaseTime"        // Earliest release time
	TagIPTCExpirationDate     TagID = "IPTC:ExpirationDate"     // Latest use date
)

// IPTC Creator/Author Tags
const (
	TagIPTCByline         TagID = "IPTC:Byline"         // Creator/author name
	TagIPTCBylineTitle    TagID = "IPTC:BylineTitle"    // Creator's title/position
	TagIPTCCredit         TagID = "IPTC:Credit"         // Provider credit line
	TagIPTCSource         TagID = "IPTC:Source"         // Original owner/creator
	TagIPTCCopyrightNotice TagID = "IPTC:CopyrightNotice" // Copyright notice
	TagIPTCContact        TagID = "IPTC:Contact"        // Contact information
	TagIPTCWriterEditor   TagID = "IPTC:WriterEditor"   // Caption writer
)

// IPTC Location Tags
const (
	TagIPTCCity            TagID = "IPTC:City"            // City of origin
	TagIPTCSublocation     TagID = "IPTC:Sublocation"     // Location within city
	TagIPTCProvinceState   TagID = "IPTC:ProvinceState"   // Province/State
	TagIPTCCountryCode     TagID = "IPTC:CountryCode"     // ISO 3166 country code
	TagIPTCCountryName     TagID = "IPTC:CountryName"     // Full country name
	TagIPTCContentLocCode  TagID = "IPTC:ContentLocCode"  // Content location code
	TagIPTCContentLocName  TagID = "IPTC:ContentLocName"  // Content location name
)

// IPTC Description Tags
const (
	TagIPTCHeadline       TagID = "IPTC:Headline"       // Publishable headline
	TagIPTCCaptionAbstract TagID = "IPTC:CaptionAbstract" // Description/caption
	TagIPTCOriginProgram  TagID = "IPTC:OriginProgram"  // Program that created file
	TagIPTCProgramVersion TagID = "IPTC:ProgramVersion" // Version of program
	TagIPTCTransmissionRef TagID = "IPTC:TransmissionRef" // Original reference/job ID
)

// =============================================================================
// XMP Tags
// =============================================================================

// XMP Dublin Core (dc) Tags
const (
	TagXMPTitle       TagID = "XMP-dc:title"       // Title of the work
	TagXMPCreator     TagID = "XMP-dc:creator"     // Creator/author
	TagXMPDescription TagID = "XMP-dc:description" // Description/caption
	TagXMPSubject     TagID = "XMP-dc:subject"     // Keywords/subjects
	TagXMPRights      TagID = "XMP-dc:rights"      // Copyright/rights info
	TagXMPDate        TagID = "XMP-dc:date"        // Date
	TagXMPFormat      TagID = "XMP-dc:format"      // MIME type
	TagXMPIdentifier  TagID = "XMP-dc:identifier"  // Unique identifier
	TagXMPLanguage    TagID = "XMP-dc:language"    // Language
	TagXMPPublisher   TagID = "XMP-dc:publisher"   // Publisher
	TagXMPRelation    TagID = "XMP-dc:relation"    // Related resources
	TagXMPSource      TagID = "XMP-dc:source"      // Source
	TagXMPType        TagID = "XMP-dc:type"        // Type/genre
)

// XMP Basic (xmp) Tags
const (
	TagXMPCreateDate   TagID = "XMP-xmp:CreateDate"   // Date created
	TagXMPModifyDate   TagID = "XMP-xmp:ModifyDate"   // Date modified
	TagXMPMetadataDate TagID = "XMP-xmp:MetadataDate" // Metadata last modified
	TagXMPCreatorTool  TagID = "XMP-xmp:CreatorTool"  // Application that created file
	TagXMPRating       TagID = "XMP-xmp:Rating"       // User rating (0-5)
	TagXMPLabel        TagID = "XMP-xmp:Label"        // Color label
	TagXMPBaseURL      TagID = "XMP-xmp:BaseURL"      // Base URL for relative URLs
)

// XMP Rights (xmpRights) Tags
const (
	TagXMPCertificate   TagID = "XMP-xmpRights:Certificate"   // Rights certificate
	TagXMPMarked        TagID = "XMP-xmpRights:Marked"        // Copyright marked
	TagXMPOwner         TagID = "XMP-xmpRights:Owner"         // Rights owner
	TagXMPUsageTerms    TagID = "XMP-xmpRights:UsageTerms"    // Usage terms
	TagXMPWebStatement  TagID = "XMP-xmpRights:WebStatement"  // Web rights statement
)

// XMP Photoshop Tags
const (
	TagXMPPhotoshopCity        TagID = "XMP-photoshop:City"            // City
	TagXMPPhotoshopState       TagID = "XMP-photoshop:State"           // State/Province
	TagXMPPhotoshopCountry     TagID = "XMP-photoshop:Country"         // Country
	TagXMPPhotoshopCredit      TagID = "XMP-photoshop:Credit"          // Credit line
	TagXMPPhotoshopSource      TagID = "XMP-photoshop:Source"          // Source
	TagXMPPhotoshopHeadline    TagID = "XMP-photoshop:Headline"        // Headline
	TagXMPPhotoshopInstructions TagID = "XMP-photoshop:Instructions"    // Instructions
	TagXMPPhotoshopDateCreated TagID = "XMP-photoshop:DateCreated"     // Date created
	TagXMPPhotoshopAuthorsPos  TagID = "XMP-photoshop:AuthorsPosition" // Author's position
	TagXMPPhotoshopCaptionWriter TagID = "XMP-photoshop:CaptionWriter" // Caption writer
	TagXMPPhotoshopCategory    TagID = "XMP-photoshop:Category"        // Category
	TagXMPPhotoshopColorMode   TagID = "XMP-photoshop:ColorMode"       // Color mode
	TagXMPPhotoshopICCProfile  TagID = "XMP-photoshop:ICCProfile"      // ICC profile name
)

// XMP TIFF Tags (via XMP)
const (
	TagXMPTIFFMake        TagID = "XMP-tiff:Make"        // Camera make
	TagXMPTIFFModel       TagID = "XMP-tiff:Model"       // Camera model
	TagXMPTIFFOrientation TagID = "XMP-tiff:Orientation" // Image orientation
	TagXMPTIFFXResolution TagID = "XMP-tiff:XResolution" // X resolution
	TagXMPTIFFYResolution TagID = "XMP-tiff:YResolution" // Y resolution
	TagXMPTIFFImageWidth  TagID = "XMP-tiff:ImageWidth"  // Image width
	TagXMPTIFFImageLength TagID = "XMP-tiff:ImageLength" // Image height
)

// XMP EXIF Tags (via XMP)
const (
	TagXMPExifDateTimeOriginal  TagID = "XMP-exif:DateTimeOriginal"  // Date/time taken
	TagXMPExifExposureTime      TagID = "XMP-exif:ExposureTime"      // Exposure time
	TagXMPExifFNumber           TagID = "XMP-exif:FNumber"           // Aperture
	TagXMPExifISOSpeedRatings   TagID = "XMP-exif:ISOSpeedRatings"   // ISO
	TagXMPExifFocalLength       TagID = "XMP-exif:FocalLength"       // Focal length
	TagXMPExifFlash             TagID = "XMP-exif:Flash"             // Flash info
	TagXMPExifGPSLatitude       TagID = "XMP-exif:GPSLatitude"       // GPS latitude
	TagXMPExifGPSLongitude      TagID = "XMP-exif:GPSLongitude"      // GPS longitude
	TagXMPExifGPSAltitude       TagID = "XMP-exif:GPSAltitude"       // GPS altitude
)

// =============================================================================
// ICC Tags
// =============================================================================

// ICC Profile Header Tags
const (
	TagICCProfileDescription TagID = "ICC:ProfileDescription" // Profile description
	TagICCCopyright          TagID = "ICC:Copyright"          // Copyright notice
	TagICCProfileClass       TagID = "ICC:ProfileClass"       // Profile class
	TagICCColorSpace         TagID = "ICC:ColorSpace"         // Color space
	TagICCPCS                TagID = "ICC:PCS"                // Profile connection space
	TagICCProfileVersion     TagID = "ICC:ProfileVersion"     // Profile version
	TagICCDeviceManufacturer TagID = "ICC:DeviceManufacturer" // Device manufacturer
	TagICCDeviceModel        TagID = "ICC:DeviceModel"        // Device model
	TagICCRenderingIntent    TagID = "ICC:RenderingIntent"    // Rendering intent
	TagICCCreationDate       TagID = "ICC:CreationDate"       // Profile creation date
	TagICCPlatform           TagID = "ICC:Platform"           // Primary platform
	TagICCCMMType            TagID = "ICC:CMMType"            // CMM type
)

// ICC Color Tags
const (
	TagICCMediaWhitePoint    TagID = "ICC:MediaWhitePoint"    // Media white point XYZ
	TagICCMediaBlackPoint    TagID = "ICC:MediaBlackPoint"    // Media black point XYZ
	TagICCRedColorant        TagID = "ICC:RedColorant"        // Red matrix column
	TagICCGreenColorant      TagID = "ICC:GreenColorant"      // Green matrix column
	TagICCBlueColorant       TagID = "ICC:BlueColorant"       // Blue matrix column
	TagICCRedTRC             TagID = "ICC:RedTRC"             // Red tone curve
	TagICCGreenTRC           TagID = "ICC:GreenTRC"           // Green tone curve
	TagICCBlueTRC            TagID = "ICC:BlueTRC"            // Blue tone curve
	TagICCGrayTRC            TagID = "ICC:GrayTRC"            // Gray tone curve
	TagICCLuminance          TagID = "ICC:Luminance"          // Luminance value
	TagICCChromaticAdaptation TagID = "ICC:ChromaticAdaptation" // Chromatic adaptation
)

// ICC Device Tags
const (
	TagICCDeviceMfgDesc   TagID = "ICC:DeviceMfgDesc"   // Device manufacturer description
	TagICCDeviceModelDesc TagID = "ICC:DeviceModelDesc" // Device model description
	TagICCTechnology      TagID = "ICC:Technology"      // Device technology
	TagICCViewingCondDesc TagID = "ICC:ViewingCondDesc" // Viewing conditions description
	TagICCMeasurement     TagID = "ICC:Measurement"     // Measurement info
)
