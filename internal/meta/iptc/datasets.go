package iptc

// DatasetInfo contains metadata about an IPTC dataset
type DatasetInfo struct {
	Name       string
	Repeatable bool
}

// Envelope Record (Record 1) datasets
var envelopeDatasets = map[uint8]DatasetInfo{
	0:   {"RecordVersion", false},
	5:   {"Destination", true},
	20:  {"FileFormat", false},
	22:  {"FileFormatVersion", false},
	30:  {"ServiceIdentifier", false},
	40:  {"EnvelopeNumber", false},
	50:  {"ProductID", true},
	60:  {"EnvelopePriority", false},
	70:  {"DateSent", false},
	80:  {"TimeSent", false},
	90:  {"CodedCharacterSet", false},
	100: {"UniqueObjectName", false},
	120: {"ARMIdentifier", false},
	122: {"ARMVersion", false},
}

// Application Record (Record 2) datasets - the most commonly used
var applicationDatasets = map[uint8]DatasetInfo{
	0:   {"RecordVersion", false},
	3:   {"ObjectTypeReference", false},
	4:   {"ObjectAttributeReference", true},
	5:   {"ObjectName", false},           // Title
	7:   {"EditStatus", false},
	8:   {"EditorialUpdate", false},
	10:  {"Urgency", false},
	12:  {"SubjectReference", true},
	15:  {"Category", false},
	20:  {"SupplementalCategories", true},
	22:  {"FixtureIdentifier", false},
	25:  {"Keywords", true},
	26:  {"ContentLocationCode", true},
	27:  {"ContentLocationName", true},
	30:  {"ReleaseDate", false},
	35:  {"ReleaseTime", false},
	37:  {"ExpirationDate", false},
	38:  {"ExpirationTime", false},
	40:  {"SpecialInstructions", false},
	42:  {"ActionAdvised", false},
	45:  {"ReferenceService", true},
	47:  {"ReferenceDate", true},
	50:  {"ReferenceNumber", true},
	55:  {"DateCreated", false},
	60:  {"TimeCreated", false},
	62:  {"DigitalCreationDate", false},
	63:  {"DigitalCreationTime", false},
	65:  {"OriginatingProgram", false},
	70:  {"ProgramVersion", false},
	75:  {"ObjectCycle", false},
	80:  {"Byline", true},              // Author/Creator
	85:  {"BylineTitle", true},         // Author's Title
	90:  {"City", false},
	92:  {"Sublocation", false},
	95:  {"Province-State", false},
	100: {"Country-PrimaryLocationCode", false},
	101: {"Country-PrimaryLocationName", false},
	103: {"OriginalTransmissionReference", false},
	105: {"Headline", false},
	110: {"Credit", false},
	115: {"Source", false},
	116: {"CopyrightNotice", false},
	118: {"Contact", true},
	120: {"Caption-Abstract", false},   // Description
	121: {"Writer-Editor", true},       // Caption Writer
	122: {"RasterizedCaption", false},
	125: {"ImageType", false},
	130: {"ImageOrientation", false},
	131: {"LanguageIdentifier", false},
	135: {"AudioType", false},
	150: {"AudioSamplingRate", false},
	151: {"AudioSamplingResolution", false},
	152: {"AudioDuration", false},
	153: {"AudioOutcue", false},
	200: {"ObjectDataPreviewFileFormat", false},
	201: {"ObjectDataPreviewFileFormatVersion", false},
	202: {"ObjectDataPreviewData", false},
	221: {"Prefs", false}, // Photo Mechanic preferences
}

// getDatasetInfo returns info about a dataset
func getDatasetInfo(record Record, datasetID uint8) DatasetInfo {
	var datasets map[uint8]DatasetInfo

	switch record {
	case RecordEnvelope:
		datasets = envelopeDatasets
	case RecordApplication:
		datasets = applicationDatasets
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

