package iptc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

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
	5:   {"ObjectName", false},
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
	80:  {"Byline", true},
	85:  {"BylineTitle", true},
	90:  {"City", false},
	92:  {"Sublocation", false},
	95:  {"ProvinceState", false},
	100: {"CountryPrimaryLocationCode", false},
	101: {"CountryPrimaryLocationName", false},
	103: {"OriginalTransmissionReference", false},
	105: {"Headline", false},
	110: {"Credit", false},
	115: {"Source", false},
	116: {"CopyrightNotice", false},
	118: {"Contact", true},
	120: {"CaptionAbstract", false},
	121: {"WriterEditor", true},
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
	221: {"Prefs", false},
	227: {"ContentCreator", true},
	228: {"ContentCreatorJobTitle", true},
	230: {"AuthorsPosition", false},
	231: {"ExtendedCity", false},
	232: {"ExtendedCountry", false},
	233: {"ExtendedProvince", false},
	240: {"SceneCode", true},
	241: {"SubjectCode", true},
}

// NewsPhoto Record (Record 3) datasets - deprecated but still encountered
var newsPhotoDatasets = map[uint8]DatasetInfo{
	0:   {"RecordVersion", false},
	5:   {"PictureNumber", false},
	10:  {"PixelsPerLine", false},
	20:  {"NumberOfLines", false},
	30:  {"PixelSizeInScanningDirection", false},
	40:  {"PixelSizePerpendicularToScanning", false},
	55:  {"SupplementType", false},
	60:  {"ColourRepresentation", false},
	64:  {"InterchangeColourSpace", false},
	65:  {"ColourSequence", false},
	66:  {"ICCInputColourProfile", false},
	70:  {"ColourCalibrationMatrixTable", false},
	80:  {"LookupTable", false},
	84:  {"NumIndexEntries", false},
	85:  {"ColourPalette", false},
	86:  {"NumBitsPerSample", false},
	90:  {"SamplingStructure", false},
	100: {"ScanningDirection", false},
	102: {"ImageRotation", false},
	110: {"DataCompressionMethod", false},
	120: {"QuantisationMethod", false},
	125: {"EndPoints", false},
	130: {"ExcursionTolerance", false},
	135: {"BitsPerComponent", false},
	140: {"MaximumDensityRange", false},
	145: {"GammaCompensatedValue", false},
}

// Pre-ObjectData Record (Record 7) datasets
var preObjectDataDatasets = map[uint8]DatasetInfo{
	10: {"SizeMode", false},
	20: {"MaxSubfileSize", false},
	90: {"ObjectDataSizeAnnounced", false},
	95: {"MaxObjectDataSize", false},
}

// ObjectData Record (Record 8) datasets
var objectDataDatasets = map[uint8]DatasetInfo{
	10: {"SubFile", true},
}

// Post-ObjectData Record (Record 9) datasets
var postObjectDataDatasets = map[uint8]DatasetInfo{
	10: {"ConfirmedObjectDataSize", false},
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
	return fmt.Sprintf("Dataset%d:%d", record, datasetID)
}

// isRepeatable returns whether a dataset can appear multiple times
func isRepeatable(record Record, datasetID uint8) bool {
	return getDatasetInfo(record, datasetID).Repeatable
}

// parseDatasetValue parses the value based on dataset type
func parseDatasetValue(record Record, datasetID uint8, data []byte) any {
	if record == RecordApplication {
		switch datasetID {
		case 0: // RecordVersion
			if len(data) >= 2 {
				return int(binary.BigEndian.Uint16(data[0:2]))
			}
		case 10: // Urgency
			if len(data) >= 1 && data[0] >= '0' && data[0] <= '9' {
				return int(data[0] - '0')
			}
		case 55, 62: // DateCreated, DigitalCreationDate (CCYYMMDD)
			return parseDateString(data)
		case 60, 63: // TimeCreated, DigitalCreationTime (HHMMSS±HHMM)
			return parseTimeString(data)
		case 30, 37: // ReleaseDate, ExpirationDate
			return parseDateString(data)
		case 35, 38: // ReleaseTime, ExpirationTime
			return parseTimeString(data)
		case 221: // Prefs (Photo Mechanic format: Tagged:ColorClass:Rating:FrameNum)
			return parsePrefs(data)
		}
	}

	if record == RecordEnvelope {
		switch datasetID {
		case 0: // RecordVersion
			if len(data) >= 2 {
				return int(binary.BigEndian.Uint16(data[0:2]))
			}
		case 70: // DateSent
			return parseDateString(data)
		case 80: // TimeSent
			return parseTimeString(data)
		}
	}

	// Default: treat as string, trim null bytes
	return trimNullBytes(data)
}

// parseDateString parses IPTC date format (CCYYMMDD or YYYYMMDD)
func parseDateString(data []byte) string {
	s := string(data)
	if len(s) == 8 {
		// Format as YYYY-MM-DD
		return s[0:4] + "-" + s[4:6] + "-" + s[6:8]
	}
	return s
}

// parseTimeString parses IPTC time format (HHMMSS±HHMM)
func parseTimeString(data []byte) string {
	s := string(data)
	if len(s) >= 6 {
		result := s[0:2] + ":" + s[2:4] + ":" + s[4:6]
		if len(s) >= 11 {
			// Include timezone (format: ±HH:MM)
			result += s[6:7] + s[7:9] + ":" + s[9:11]
		}
		return result
	}
	return s
}

// parsePrefs parses Photo Mechanic Prefs field (format: Tagged:ColorClass:Rating:FrameNum)
func parsePrefs(data []byte) string {
	s := trimNullBytes(data)
	parts := bytes.Split(data, []byte(":"))
	if len(parts) >= 4 {
		return fmt.Sprintf("Tagged:%s, ColorClass:%s, Rating:%s, FrameNum:%s",
			trimNullBytes(parts[0]),
			trimNullBytes(parts[1]),
			trimNullBytes(parts[2]),
			trimNullBytes(parts[3]))
	}
	return s
}

// trimNullBytes removes trailing null bytes and converts to string
func trimNullBytes(data []byte) string {
	// Trim trailing nulls
	for len(data) > 0 && data[len(data)-1] == 0 {
		data = data[:len(data)-1]
	}
	return strings.TrimSpace(string(data))
}
