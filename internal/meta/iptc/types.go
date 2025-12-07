package iptc

// Record numbers in IPTC-IIM
type Record uint8

const (
	RecordEnvelope       Record = 1 // Envelope Record
	RecordApplication    Record = 2 // Application Record (most common)
	RecordNewsPhoto      Record = 3 // NewsPhoto Record (deprecated)
	RecordPreObjectData  Record = 7
	RecordObjectData     Record = 8
	RecordPostObjectData Record = 9
)

// String returns the record name
func (r Record) String() string {
	switch r {
	case RecordEnvelope:
		return "Envelope"
	case RecordApplication:
		return "Application"
	case RecordNewsPhoto:
		return "NewsPhoto"
	case RecordPreObjectData:
		return "PreObjectData"
	case RecordObjectData:
		return "ObjectData"
	case RecordPostObjectData:
		return "PostObjectData"
	default:
		return "Unknown"
	}
}

// Dataset represents an IPTC dataset (tag)
type Dataset struct {
	Record    Record
	DatasetID uint8
	Name      string
	Value     any
	Raw       []byte
}

// PhotoshopResource represents a Photoshop Image Resource Block (8BIM)
type PhotoshopResource struct {
	ID   uint16
	Name string
	Data []byte
}

// Common Photoshop resource IDs
const (
	ResourceIPTC          uint16 = 0x0404 // 1028 - IPTC-NAA record
	ResourceCaptionDigest uint16 = 0x0425 // 1061 - Caption digest
	ResourcePrintScale    uint16 = 0x0400 // 1024 - Print scale
	ResourceCopyright     uint16 = 0x040A // 1034 - Copyright flag
	ResourceURL           uint16 = 0x040B // 1035 - URL
	ResourceThumbnail     uint16 = 0x0409 // 1033 - Thumbnail (JPEG)
	ResourceGlobalAngle   uint16 = 0x040D // 1037 - Global angle
	ResourceICCProfile    uint16 = 0x040F // 1039 - ICC Profile
	ResourceXMP           uint16 = 0x0424 // 1060 - XMP
	ResourceEXIF1         uint16 = 0x0422 // 1058 - EXIF data 1
	ResourceEXIF3         uint16 = 0x0423 // 1059 - EXIF data 3
)
