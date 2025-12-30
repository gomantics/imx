package imx

// =============================================================================
// EXIF Tags
// =============================================================================

// Common EXIF tag constants for easy access

// Camera and Device Tags
const (
	TagMake         TagID = "EXIF:Make"
	TagModel        TagID = "EXIF:Model"
	TagSoftware     TagID = "EXIF:Software"
	TagOrientation  TagID = "EXIF:Orientation"
	TagSerialNumber TagID = "EXIF:SerialNumber"
	TagOwnerName    TagID = "EXIF:OwnerName"
)

// Lens Tags
const (
	TagLensMake         TagID = "EXIF:LensMake"
	TagLensModel        TagID = "EXIF:LensModel"
	TagLensSerialNumber TagID = "EXIF:LensSerialNumber"
	TagLensInfo         TagID = "EXIF:LensInfo"
)

// Image Description Tags
const (
	TagImageDescription TagID = "EXIF:ImageDescription"
	TagImageTitle       TagID = "EXIF:ImageTitle"
	TagArtist           TagID = "EXIF:Artist"
	TagPhotographer     TagID = "EXIF:Photographer"
	TagCopyright        TagID = "EXIF:Copyright"
	TagUserComment      TagID = "EXIF:UserComment"
)

// Date and Time Tags
const (
	TagDateTimeOriginal    TagID = "EXIF:DateTimeOriginal"
	TagCreateDate          TagID = "EXIF:CreateDate"
	TagModifyDate          TagID = "EXIF:ModifyDate"
	TagOffsetTime          TagID = "EXIF:OffsetTime"
	TagOffsetTimeOriginal  TagID = "EXIF:OffsetTimeOriginal"
	TagSubSecTime          TagID = "EXIF:SubSecTime"
	TagSubSecTimeOriginal  TagID = "EXIF:SubSecTimeOriginal"
	TagSubSecTimeDigitized TagID = "EXIF:SubSecTimeDigitized"
)

// Image Dimensions Tags
const (
	TagImageWidth      TagID = "EXIF:ImageWidth"
	TagImageHeight     TagID = "EXIF:ImageHeight"
	TagExifImageWidth  TagID = "EXIF:ExifImageWidth"
	TagExifImageHeight TagID = "EXIF:ExifImageHeight"
)

// Exposure Tags
const (
	TagExposureTime         TagID = "EXIF:ExposureTime"
	TagShutterSpeedValue    TagID = "EXIF:ShutterSpeedValue"
	TagFNumber              TagID = "EXIF:FNumber"
	TagApertureValue        TagID = "EXIF:ApertureValue"
	TagExposureProgram      TagID = "EXIF:ExposureProgram"
	TagExposureMode         TagID = "EXIF:ExposureMode"
	TagExposureCompensation TagID = "EXIF:ExposureCompensation"
	TagBrightnessValue      TagID = "EXIF:BrightnessValue"
)

// ISO Tags
const (
	TagISO                       TagID = "EXIF:ISO"
	TagISOSpeed                  TagID = "EXIF:ISOSpeed"
	TagSensitivityType           TagID = "EXIF:SensitivityType"
	TagStandardOutputSensitivity TagID = "EXIF:StandardOutputSensitivity"
	TagRecommendedExposureIndex  TagID = "EXIF:RecommendedExposureIndex"
)

// Focus and Lens Settings Tags
const (
	TagFocalLength             TagID = "EXIF:FocalLength"
	TagFocalLengthIn35mmFormat TagID = "EXIF:FocalLengthIn35mmFormat"
	TagMaxApertureValue        TagID = "EXIF:MaxApertureValue"
	TagSubjectDistance         TagID = "EXIF:SubjectDistance"
	TagSubjectDistanceRange    TagID = "EXIF:SubjectDistanceRange"
)

// Flash Tags
const (
	TagFlash       TagID = "EXIF:Flash"
	TagFlashEnergy TagID = "EXIF:FlashEnergy"
)

// Metering and Lighting Tags
const (
	TagMeteringMode TagID = "EXIF:MeteringMode"
	TagLightSource  TagID = "EXIF:LightSource"
	TagWhiteBalance TagID = "EXIF:WhiteBalance"
)

// Image Quality Tags
const (
	TagColorSpace       TagID = "EXIF:ColorSpace"
	TagContrast         TagID = "EXIF:Contrast"
	TagSaturation       TagID = "EXIF:Saturation"
	TagSharpness        TagID = "EXIF:Sharpness"
	TagDigitalZoomRatio TagID = "EXIF:DigitalZoomRatio"
)

// Scene Tags
const (
	TagSceneCaptureType TagID = "EXIF:SceneCaptureType"
	TagSceneType        TagID = "EXIF:SceneType"
)

// GPS Tags
const (
	TagGPSVersionID         TagID = "EXIF:GPSVersionID"
	TagGPSLatitudeRef       TagID = "EXIF:GPSLatitudeRef"
	TagGPSLatitude          TagID = "EXIF:GPSLatitude"
	TagGPSLongitudeRef      TagID = "EXIF:GPSLongitudeRef"
	TagGPSLongitude         TagID = "EXIF:GPSLongitude"
	TagGPSAltitudeRef       TagID = "EXIF:GPSAltitudeRef"
	TagGPSAltitude          TagID = "EXIF:GPSAltitude"
	TagGPSTimeStamp         TagID = "EXIF:GPSTimeStamp"
	TagGPSDateStamp         TagID = "EXIF:GPSDateStamp"
	TagGPSSatellites        TagID = "EXIF:GPSSatellites"
	TagGPSStatus            TagID = "EXIF:GPSStatus"
	TagGPSMeasureMode       TagID = "EXIF:GPSMeasureMode"
	TagGPSDOP               TagID = "EXIF:GPSDOP"
	TagGPSSpeed             TagID = "EXIF:GPSSpeed"
	TagGPSSpeedRef          TagID = "EXIF:GPSSpeedRef"
	TagGPSTrack             TagID = "EXIF:GPSTrack"
	TagGPSTrackRef          TagID = "EXIF:GPSTrackRef"
	TagGPSImgDirection      TagID = "EXIF:GPSImgDirection"
	TagGPSImgDirectionRef   TagID = "EXIF:GPSImgDirectionRef"
	TagGPSMapDatum          TagID = "EXIF:GPSMapDatum"
	TagGPSDestLatitude      TagID = "EXIF:GPSDestLatitude"
	TagGPSDestLatitudeRef   TagID = "EXIF:GPSDestLatitudeRef"
	TagGPSDestLongitude     TagID = "EXIF:GPSDestLongitude"
	TagGPSDestLongitudeRef  TagID = "EXIF:GPSDestLongitudeRef"
	TagGPSDestBearing       TagID = "EXIF:GPSDestBearing"
	TagGPSDestBearingRef    TagID = "EXIF:GPSDestBearingRef"
	TagGPSDestDistance      TagID = "EXIF:GPSDestDistance"
	TagGPSDestDistanceRef   TagID = "EXIF:GPSDestDistanceRef"
	TagGPSProcessingMethod  TagID = "EXIF:GPSProcessingMethod"
	TagGPSAreaInformation   TagID = "EXIF:GPSAreaInformation"
	TagGPSDifferential      TagID = "EXIF:GPSDifferential"
	TagGPSHPositioningError TagID = "EXIF:GPSHPositioningError"
)

// Version Tags
const (
	TagExifVersion     TagID = "EXIF:ExifVersion"
	TagFlashpixVersion TagID = "EXIF:FlashpixVersion"
)

// Other Common Tags
const (
	TagCompression               TagID = "EXIF:Compression"
	TagPhotometricInterpretation TagID = "EXIF:PhotometricInterpretation"
	TagXResolution               TagID = "EXIF:XResolution"
	TagYResolution               TagID = "EXIF:YResolution"
	TagResolutionUnit            TagID = "EXIF:ResolutionUnit"
	TagYCbCrPositioning          TagID = "EXIF:YCbCrPositioning"
	TagRating                    TagID = "EXIF:Rating"
	TagRatingPercent             TagID = "EXIF:RatingPercent"
)

// =============================================================================
// IPTC Tags (from datasets.go)
// =============================================================================

// IPTC Core Identification Tags
const (
	TagIPTCObjectName             TagID = "IPTC:ObjectName"             // Title/shorthand reference
	TagIPTCUrgency                TagID = "IPTC:Urgency"                // 1=most urgent, 8=least
	TagIPTCCategory               TagID = "IPTC:Category"               // Subject category code
	TagIPTCSupplementalCategories TagID = "IPTC:SupplementalCategories" // Additional categories
	TagIPTCKeywords               TagID = "IPTC:Keywords"               // Keywords for indexing
	TagIPTCFixtureIdentifier      TagID = "IPTC:FixtureIdentifier"      // Identifies recurring events
	TagIPTCEditStatus             TagID = "IPTC:EditStatus"             // Status of object
	TagIPTCSpecialInstructions    TagID = "IPTC:SpecialInstructions"    // Special instructions
	TagIPTCSubjectReference       TagID = "IPTC:SubjectReference"       // Structured subject reference
)

// IPTC Date/Time Tags
const (
	TagIPTCDateCreated         TagID = "IPTC:DateCreated"         // Intellectual content created
	TagIPTCTimeCreated         TagID = "IPTC:TimeCreated"         // Time content created
	TagIPTCDigitalCreationDate TagID = "IPTC:DigitalCreationDate" // Digital file created
	TagIPTCDigitalCreationTime TagID = "IPTC:DigitalCreationTime" // Digital file time
	TagIPTCReleaseDate         TagID = "IPTC:ReleaseDate"         // Earliest release date
	TagIPTCReleaseTime         TagID = "IPTC:ReleaseTime"         // Earliest release time
	TagIPTCExpirationDate      TagID = "IPTC:ExpirationDate"      // Latest use date
	TagIPTCExpirationTime      TagID = "IPTC:ExpirationTime"      // Latest use time
)

// IPTC Creator/Author Tags
const (
	TagIPTCByline          TagID = "IPTC:Byline"          // Creator/author name
	TagIPTCBylineTitle     TagID = "IPTC:BylineTitle"     // Creator's title/position
	TagIPTCCredit          TagID = "IPTC:Credit"          // Provider credit line
	TagIPTCSource          TagID = "IPTC:Source"          // Original owner/creator
	TagIPTCCopyrightNotice TagID = "IPTC:CopyrightNotice" // Copyright notice
	TagIPTCContact         TagID = "IPTC:Contact"         // Contact information
	TagIPTCWriterEditor    TagID = "IPTC:Writer-Editor"   // Caption writer name
)

// IPTC Location Tags
const (
	TagIPTCCity                TagID = "IPTC:City"                        // City of origin
	TagIPTCSublocation         TagID = "IPTC:Sublocation"                 // Location within city
	TagIPTCProvinceState       TagID = "IPTC:Province-State"              // Province/State of origin
	TagIPTCCountryCode         TagID = "IPTC:Country-PrimaryLocationCode" // ISO 3166 country code
	TagIPTCCountryName         TagID = "IPTC:Country-PrimaryLocationName" // Full country name
	TagIPTCContentLocationCode TagID = "IPTC:ContentLocationCode"         // Content location code
	TagIPTCContentLocationName TagID = "IPTC:ContentLocationName"         // Content location name
)

// IPTC Description Tags
const (
	TagIPTCHeadline                      TagID = "IPTC:Headline"                      // Publishable headline
	TagIPTCCaptionAbstract               TagID = "IPTC:Caption-Abstract"              // Description/caption
	TagIPTCOriginatingProgram            TagID = "IPTC:OriginatingProgram"            // Program that created file
	TagIPTCProgramVersion                TagID = "IPTC:ProgramVersion"                // Version of program
	TagIPTCOriginalTransmissionReference TagID = "IPTC:OriginalTransmissionReference" // Original reference/job ID
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
	TagXMPCertificate  TagID = "XMP-xmpRights:Certificate"  // Rights certificate
	TagXMPMarked       TagID = "XMP-xmpRights:Marked"       // Copyright marked
	TagXMPOwner        TagID = "XMP-xmpRights:Owner"        // Rights owner
	TagXMPUsageTerms   TagID = "XMP-xmpRights:UsageTerms"   // Usage terms
	TagXMPWebStatement TagID = "XMP-xmpRights:WebStatement" // Web rights statement
)

// XMP Photoshop Tags
const (
	TagXMPPhotoshopCity          TagID = "XMP-photoshop:City"            // City
	TagXMPPhotoshopState         TagID = "XMP-photoshop:State"           // State/Province
	TagXMPPhotoshopCountry       TagID = "XMP-photoshop:Country"         // Country
	TagXMPPhotoshopCredit        TagID = "XMP-photoshop:Credit"          // Credit line
	TagXMPPhotoshopSource        TagID = "XMP-photoshop:Source"          // Source
	TagXMPPhotoshopHeadline      TagID = "XMP-photoshop:Headline"        // Headline
	TagXMPPhotoshopInstructions  TagID = "XMP-photoshop:Instructions"    // Instructions
	TagXMPPhotoshopDateCreated   TagID = "XMP-photoshop:DateCreated"     // Date created
	TagXMPPhotoshopAuthorsPos    TagID = "XMP-photoshop:AuthorsPosition" // Author's position
	TagXMPPhotoshopCaptionWriter TagID = "XMP-photoshop:CaptionWriter"   // Caption writer
	TagXMPPhotoshopCategory      TagID = "XMP-photoshop:Category"        // Category
	TagXMPPhotoshopColorMode     TagID = "XMP-photoshop:ColorMode"       // Color mode
	TagXMPPhotoshopICCProfile    TagID = "XMP-photoshop:ICCProfile"      // ICC profile name
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
	TagXMPExifDateTimeOriginal TagID = "XMP-exif:DateTimeOriginal" // Date/time taken
	TagXMPExifExposureTime     TagID = "XMP-exif:ExposureTime"     // Exposure time
	TagXMPExifFNumber          TagID = "XMP-exif:FNumber"          // Aperture
	TagXMPExifISOSpeedRatings  TagID = "XMP-exif:ISOSpeedRatings"  // ISO
	TagXMPExifFocalLength      TagID = "XMP-exif:FocalLength"      // Focal length
	TagXMPExifFlash            TagID = "XMP-exif:Flash"            // Flash info
	TagXMPExifGPSLatitude      TagID = "XMP-exif:GPSLatitude"      // GPS latitude
	TagXMPExifGPSLongitude     TagID = "XMP-exif:GPSLongitude"     // GPS longitude
	TagXMPExifGPSAltitude      TagID = "XMP-exif:GPSAltitude"      // GPS altitude
)

// =============================================================================
// ICC Tags (from icc.go header fields and tags.go knownTags)
// =============================================================================

// ICC Profile Header Tags (from icc.go addTag calls)
const (
	TagICCProfileSize        TagID = "ICC:ProfileSize"        // Profile size in bytes
	TagICCPreferredCMM       TagID = "ICC:PreferredCMM"       // Preferred CMM
	TagICCVersion            TagID = "ICC:Version"            // Profile version
	TagICCProfileClass       TagID = "ICC:ProfileClass"       // Profile class
	TagICCColorSpace         TagID = "ICC:ColorSpace"         // Color space
	TagICCPCS                TagID = "ICC:PCS"                // Profile connection space
	TagICCCreateDate         TagID = "ICC:CreateDate"         // Profile creation date
	TagICCPlatform           TagID = "ICC:Platform"           // Primary platform
	TagICCRenderingIntent    TagID = "ICC:RenderingIntent"    // Rendering intent
	TagICCDeviceManufacturer TagID = "ICC:DeviceManufacturer" // Device manufacturer
	TagICCDeviceModel        TagID = "ICC:DeviceModel"        // Device model
	TagICCCreator            TagID = "ICC:Creator"            // Profile creator
	TagICCPCSIlluminant      TagID = "ICC:PCSIlluminant"      // PCS illuminant
	TagICCProfileFlags       TagID = "ICC:ProfileFlags"       // Profile flags
	TagICCDeviceAttributes   TagID = "ICC:DeviceAttributes"   // Device attributes
	TagICCProfileID          TagID = "ICC:ProfileID"          // Profile ID (MD5 hash)
)

// ICC Parsed Tag Values (human-readable names from knownTags)
const (
	TagICCProfileDescription  TagID = "ICC:ProfileDescription"  // Profile description
	TagICCProfileCopyright    TagID = "ICC:ProfileCopyright"    // Copyright notice
	TagICCMediaWhitePoint     TagID = "ICC:MediaWhitePoint"     // Media white point XYZ
	TagICCMediaBlackPoint     TagID = "ICC:MediaBlackPoint"     // Media black point XYZ
	TagICCChromaticAdaptation TagID = "ICC:ChromaticAdaptation" // Chromatic adaptation matrix
)

// ICC Color Matrix Tags
const (
	TagICCRedMatrixColumn   TagID = "ICC:RedMatrixColumn"   // Red matrix column XYZ
	TagICCGreenMatrixColumn TagID = "ICC:GreenMatrixColumn" // Green matrix column XYZ
	TagICCBlueMatrixColumn  TagID = "ICC:BlueMatrixColumn"  // Blue matrix column XYZ
)

// ICC Tone Reproduction Curve Tags
const (
	TagICCRedTRC   TagID = "ICC:RedToneReproductionCurve"   // Red TRC
	TagICCGreenTRC TagID = "ICC:GreenToneReproductionCurve" // Green TRC
	TagICCBlueTRC  TagID = "ICC:BlueToneReproductionCurve"  // Blue TRC
	TagICCGrayTRC  TagID = "ICC:GrayToneReproductionCurve"  // Gray TRC
)

// ICC Device Description Tags
const (
	TagICCDeviceMfgDesc         TagID = "ICC:DeviceManufacturerDescription" // Device mfg description
	TagICCDeviceModelDesc       TagID = "ICC:DeviceModelDescription"        // Device model description
	TagICCTechnology            TagID = "ICC:Technology"                    // Device technology
	TagICCViewingConditionsDesc TagID = "ICC:ViewingConditionsDescription"  // Viewing conditions
	TagICCLuminance             TagID = "ICC:Luminance"                     // Luminance value
	TagICCMeasurement           TagID = "ICC:Measurement"                   // Measurement info
)
