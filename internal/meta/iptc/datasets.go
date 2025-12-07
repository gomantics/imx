package iptc

// DatasetInfo contains metadata about an IPTC dataset
type DatasetInfo struct {
	Name       string
	Repeatable bool
}

// Envelope Record (Record 1) datasets
// Reference: IPTC-IIM Specification 4.2
var envelopeDatasets = map[uint8]DatasetInfo{
	0:   {"RecordVersion", false},          // 1:000 - Required, version of IIM
	5:   {"Destination", true},             // 1:005 - Optional routing info
	20:  {"FileFormat", false},             // 1:020 - File format (see appendix A)
	22:  {"FileFormatVersion", false},      // 1:022 - Version of file format
	30:  {"ServiceIdentifier", false},      // 1:030 - Identifies the provider
	40:  {"EnvelopeNumber", false},         // 1:040 - 8 octet number
	50:  {"ProductID", true},               // 1:050 - Identifies subset of data
	60:  {"EnvelopePriority", false},       // 1:060 - 1=most urgent, 9=least
	70:  {"DateSent", false},               // 1:070 - CCYYMMDD
	80:  {"TimeSent", false},               // 1:080 - HHMMSS±HHMM
	90:  {"CodedCharacterSet", false},      // 1:090 - ISO 2022 escape sequences
	100: {"UniqueObjectName", false},       // 1:100 - Unique eternal identifier
	120: {"ARMIdentifier", false},          // 1:120 - Abstract Relationship Method
	122: {"ARMVersion", false},             // 1:122 - ARM version number
}

// Application Record (Record 2) datasets - the most commonly used
// Reference: IPTC-IIM Specification 4.2
var applicationDatasets = map[uint8]DatasetInfo{
	// Core identification
	0:   {"RecordVersion", false},                  // 2:000 - Required, version of IIM
	3:   {"ObjectTypeReference", false},            // 2:003 - Object type (News, Data, etc.)
	4:   {"ObjectAttributeReference", true},        // 2:004 - Object attribute (Current, Analysis, etc.)
	5:   {"ObjectName", false},                     // 2:005 - Title/shorthand reference

	// Status
	7:  {"EditStatus", false},                      // 2:007 - Status of objectdata
	8:  {"EditorialUpdate", false},                 // 2:008 - Update indicator
	10: {"Urgency", false},                         // 2:010 - 1=most urgent, 8=least, 5=normal

	// Category/Subject
	12: {"SubjectReference", true},                 // 2:012 - Structured subject reference
	15: {"Category", false},                        // 2:015 - Deprecated: 3-char category code
	20: {"SupplementalCategories", true},           // 2:020 - Deprecated: additional categories

	// Fixture/Keywords
	22: {"FixtureIdentifier", false},               // 2:022 - Identifies recurring events
	25: {"Keywords", true},                         // 2:025 - Keywords for indexing

	// Location
	26: {"ContentLocationCode", true},              // 2:026 - ISO 3166 country code
	27: {"ContentLocationName", true},              // 2:027 - Full location name

	// Temporal
	30: {"ReleaseDate", false},                     // 2:030 - Earliest release date CCYYMMDD
	35: {"ReleaseTime", false},                     // 2:035 - Earliest release time HHMMSS±HHMM
	37: {"ExpirationDate", false},                  // 2:037 - Latest use date CCYYMMDD
	38: {"ExpirationTime", false},                  // 2:038 - Latest use time HHMMSS±HHMM

	// Editorial
	40: {"SpecialInstructions", false},             // 2:040 - Editorial instructions
	42: {"ActionAdvised", false},                   // 2:042 - Type of action (01=kill, 02=replace, etc.)
	45: {"ReferenceService", true},                 // 2:045 - Service ID of prior envelope
	47: {"ReferenceDate", true},                    // 2:047 - Date of prior envelope
	50: {"ReferenceNumber", true},                  // 2:050 - Envelope number of prior envelope

	// Creation date/time
	55: {"DateCreated", false},                     // 2:055 - Intellectual content created CCYYMMDD
	60: {"TimeCreated", false},                     // 2:060 - Intellectual content created HHMMSS±HHMM
	62: {"DigitalCreationDate", false},             // 2:062 - Digital representation created CCYYMMDD
	63: {"DigitalCreationTime", false},             // 2:063 - Digital representation created HHMMSS±HHMM

	// Origin
	65: {"OriginatingProgram", false},              // 2:065 - Program used to create objectdata
	70: {"ProgramVersion", false},                  // 2:070 - Version of originating program
	75: {"ObjectCycle", false},                     // 2:075 - a=morning, p=evening, b=both

	// Creator/Author info
	80: {"Byline", true},                           // 2:080 - Creator/Author name
	85: {"BylineTitle", true},                      // 2:085 - Creator/Author title/position
	90: {"City", false},                            // 2:090 - City of origin
	92: {"Sublocation", false},                     // 2:092 - Location within city
	95: {"Province-State", false},                  // 2:095 - Province/State of origin
	100: {"Country-PrimaryLocationCode", false},    // 2:100 - ISO 3166 country code
	101: {"Country-PrimaryLocationName", false},    // 2:101 - Full country name
	103: {"OriginalTransmissionReference", false},  // 2:103 - Original owner's reference/job ID

	// Descriptive
	105: {"Headline", false},                       // 2:105 - Publishable headline
	110: {"Credit", false},                         // 2:110 - Provider credit line
	115: {"Source", false},                         // 2:115 - Original owner/creator
	116: {"CopyrightNotice", false},                // 2:116 - Copyright notice
	118: {"Contact", true},                         // 2:118 - Contact information
	120: {"Caption-Abstract", false},               // 2:120 - Description/caption
	121: {"Writer-Editor", true},                   // 2:121 - Caption writer name
	122: {"RasterizedCaption", false},              // 2:122 - B&W rasterized caption (460x128)

	// Image info
	125: {"ImageType", false},                      // 2:125 - Image type (M=monochrome, Y=yellow, etc.)
	130: {"ImageOrientation", false},               // 2:130 - L=landscape, P=portrait, S=square
	131: {"LanguageIdentifier", false},             // 2:131 - ISO 639:1988 language code

	// Audio info
	135: {"AudioType", false},                      // 2:135 - Audio type (1A, 1M, 1S, 2S, etc.)
	150: {"AudioSamplingRate", false},              // 2:150 - Hz (6 digits, leading zeros)
	151: {"AudioSamplingResolution", false},        // 2:151 - Bits per sample (2 digits)
	152: {"AudioDuration", false},                  // 2:152 - HHMMSS duration
	153: {"AudioOutcue", false},                    // 2:153 - Final words of audio

	// Preview data
	200: {"ObjectDataPreviewFileFormat", false},    // 2:200 - Preview file format (see 1:020)
	201: {"ObjectDataPreviewFileFormatVersion", false}, // 2:201 - Preview format version
	202: {"ObjectDataPreviewData", false},          // 2:202 - Preview image data

	// Extended/Custom (non-standard but commonly used)
	221: {"Prefs", false},                          // 2:221 - Photo Mechanic preferences

	// IPTC Extension (IIM 4.2)
	227: {"ContentCreator", true},                  // 2:227 - Content creator
	228: {"ContentCreatorJobTitle", true},          // 2:228 - Content creator job title
	230: {"AuthorsPosition", false},                // 2:230 - Author's position
	231: {"ExtendedCity", false},                   // 2:231 - Extended city info
	232: {"ExtendedCountry", false},                // 2:232 - Extended country info
	233: {"ExtendedProvince", false},               // 2:233 - Extended province/state
	
	// Scene/Subject codes
	240: {"SceneCode", true},                       // 2:240 - IPTC Scene codes
	241: {"SubjectCode", true},                     // 2:241 - IPTC Subject codes
}

// NewsPhoto Record (Record 3) datasets - deprecated but still encountered
// Reference: IPTC-IIM Specification (legacy)
var newsPhotoDatasets = map[uint8]DatasetInfo{
	0:   {"RecordVersion", false},                  // 3:000 - Version of record
	5:   {"PictureNumber", false},                  // 3:005 - Picture number
	10:  {"PixelsPerLine", false},                  // 3:010 - Pixels per line
	20:  {"NumberOfLines", false},                  // 3:020 - Number of lines
	30:  {"PixelSizeInScanningDirection", false},   // 3:030 - Pixel size (scanning)
	40:  {"PixelSizePerpendicularToScanning", false}, // 3:040 - Pixel size (perpendicular)
	55:  {"SupplementType", false},                 // 3:055 - Supplement type
	60:  {"ColourRepresentation", false},           // 3:060 - Colour representation
	64:  {"InterchangeColourSpace", false},         // 3:064 - Interchange colour space
	65:  {"ColourSequence", false},                 // 3:065 - Colour sequence
	66:  {"ICCInputColourProfile", false},          // 3:066 - ICC input profile
	70:  {"ColourCalibrationMatrixTable", false},   // 3:070 - Colour calibration matrix
	80:  {"LookupTable", false},                    // 3:080 - Lookup table
	84:  {"NumIndexEntries", false},                // 3:084 - Number of index entries
	85:  {"ColourPalette", false},                  // 3:085 - Colour palette
	86:  {"NumBitsPerSample", false},               // 3:086 - Bits per sample
	90:  {"SamplingStructure", false},              // 3:090 - Sampling structure
	100: {"ScanningDirection", false},              // 3:100 - Scanning direction
	102: {"ImageRotation", false},                  // 3:102 - Image rotation
	110: {"DataCompressionMethod", false},          // 3:110 - Compression method
	120: {"QuantisationMethod", false},             // 3:120 - Quantisation method
	125: {"EndPoints", false},                      // 3:125 - End points
	130: {"ExcursionTolerance", false},             // 3:130 - Excursion tolerance
	135: {"BitsPerComponent", false},               // 3:135 - Bits per component
	140: {"MaximumDensityRange", false},            // 3:140 - Maximum density range
	145: {"GammaCompensatedValue", false},          // 3:145 - Gamma compensated value
}

// Pre-ObjectData Record (Record 7) datasets
var preObjectDataDatasets = map[uint8]DatasetInfo{
	10:  {"SizeMode", false},                       // 7:010 - Size mode
	20:  {"MaxSubfileSize", false},                 // 7:020 - Maximum subfile size
	90:  {"ObjectDataSizeAnnounced", false},        // 7:090 - Object data size announced
	95:  {"MaxObjectDataSize", false},              // 7:095 - Maximum object data size
}

// ObjectData Record (Record 8) datasets
var objectDataDatasets = map[uint8]DatasetInfo{
	10: {"SubFile", true},                          // 8:010 - Subfile data
}

// Post-ObjectData Record (Record 9) datasets
var postObjectDataDatasets = map[uint8]DatasetInfo{
	10: {"ConfirmedObjectDataSize", false},         // 9:010 - Confirmed object data size
}

// getDatasetInfo returns info about a dataset
func getDatasetInfo(record Record, datasetID uint8) DatasetInfo {
	var datasets map[uint8]DatasetInfo

	switch record {
	case RecordEnvelope:
		datasets = envelopeDatasets
	case RecordApplication:
		datasets = applicationDatasets
	case RecordNewsPhoto:
		datasets = newsPhotoDatasets
	case RecordPreObjectData:
		datasets = preObjectDataDatasets
	case RecordObjectData:
		datasets = objectDataDatasets
	case RecordPostObjectData:
		datasets = postObjectDataDatasets
	default:
		return DatasetInfo{Name: "", Repeatable: false}
	}

	if info, ok := datasets[datasetID]; ok {
		return info
	}
	return DatasetInfo{Name: "", Repeatable: false}
}

// getDatasetName returns the name for a dataset
func getDatasetName(record Record, datasetID uint8) string {
	info := getDatasetInfo(record, datasetID)
	if info.Name != "" {
		return info.Name
	}
	return ""
}

// isRepeatable returns whether a dataset can appear multiple times
func isRepeatable(record Record, datasetID uint8) bool {
	return getDatasetInfo(record, datasetID).Repeatable
}
