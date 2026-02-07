package canon

// Canon MakerNote tag IDs
const (
	TagCameraSettings1 uint16 = 0x0001 // Camera settings array 1
	TagFocalLength     uint16 = 0x0002 // Focal length info
	TagFlashInfo       uint16 = 0x0003 // Flash information
	TagCameraSettings2 uint16 = 0x0004 // Camera settings array 2 (ShotInfo)
	TagPanorama        uint16 = 0x0005 // Panorama info
	TagImageType       uint16 = 0x0006 // Camera model string
	TagFirmwareVersion uint16 = 0x0007 // Firmware version
	TagFileNumber      uint16 = 0x0008 // File number
	TagOwnerName       uint16 = 0x0009 // Owner name
	TagSerialNumber    uint16 = 0x000C // Camera serial number
	TagCameraInfo      uint16 = 0x000D // Camera info
	TagFileLength      uint16 = 0x000E // File length
	TagCustomFunctions uint16 = 0x000F // Custom functions
	TagModelID         uint16 = 0x0010 // Canon model ID
	TagMovieInfo       uint16 = 0x0011 // Movie info
	TagAFInfo          uint16 = 0x0012 // AF info
	TagThumbnailOffset uint16 = 0x0081 // Thumbnail image offset
	TagThumbnailLength uint16 = 0x0082 // Thumbnail image length
	TagLensModel       uint16 = 0x0095 // Lens model string
	TagInternalSerial  uint16 = 0x0096 // Internal serial number
	TagDustRemoval     uint16 = 0x0097 // Dust removal data
	TagCropInfo        uint16 = 0x0098 // Crop info
	TagAspectInfo      uint16 = 0x009A // Aspect ratio info
	TagColorInfo       uint16 = 0x00A0 // Processing info
	TagVRDOffset       uint16 = 0x00D0 // VRD recipe offset
	TagSensorInfo      uint16 = 0x00E0 // Sensor info
	TagColorData       uint16 = 0x4001 // Color data
	TagCRWParam        uint16 = 0x4002 // CRW parameters
	TagColorInfo2      uint16 = 0x4003 // Color info 2
	TagFlavor          uint16 = 0x4005 // Picture style
	TagPictureStylePC  uint16 = 0x4008 // Picture style user def
	TagVignettingCorr  uint16 = 0x4015 // Vignetting correction
	TagLensInfo        uint16 = 0x4019 // Lens info
	TagAmbienceInfo    uint16 = 0x4020 // Ambience info
	TagFilterInfo      uint16 = 0x4024 // Filter info
)

// tagNames maps Canon tag IDs to human-readable names
var tagNames = map[uint16]string{
	TagCameraSettings1: "CameraSettings1",
	TagFocalLength:     "FocalLength",
	TagFlashInfo:       "FlashInfo",
	TagCameraSettings2: "CameraSettings2",
	TagPanorama:        "Panorama",
	TagImageType:       "ImageType",
	TagFirmwareVersion: "FirmwareVersion",
	TagFileNumber:      "FileNumber",
	TagOwnerName:       "OwnerName",
	TagSerialNumber:    "SerialNumber",
	TagCameraInfo:      "CameraInfo",
	TagFileLength:      "FileLength",
	TagCustomFunctions: "CustomFunctions",
	TagModelID:         "ModelID",
	TagMovieInfo:       "MovieInfo",
	TagAFInfo:          "AFInfo",
	TagThumbnailOffset: "ThumbnailImageValidArea",
	TagThumbnailLength: "ThumbnailImageLength",
	TagLensModel:       "LensModel",
	TagInternalSerial:  "InternalSerialNumber",
	TagDustRemoval:     "DustRemovalData",
	TagCropInfo:        "CropInfo",
	TagAspectInfo:      "AspectInfo",
	TagColorInfo:       "ColorInfo",
	TagVRDOffset:       "VRDOffset",
	TagSensorInfo:      "SensorInfo",
	TagColorData:       "ColorData",
	TagCRWParam:        "CRWParam",
	TagColorInfo2:      "ColorInfo2",
	TagFlavor:          "PictureStyleUserDef",
	TagPictureStylePC:  "PictureStylePC",
	TagVignettingCorr:  "VignettingCorrection",
	TagLensInfo:        "LensInfo",
	TagAmbienceInfo:    "AmbienceInfo",
	TagFilterInfo:      "FilterInfo",
}

// GetTagName returns the human-readable name for a Canon tag ID.
// Returns empty string if tag is not recognized.
func GetTagName(tagID uint16) string {
	if name, ok := tagNames[tagID]; ok {
		return name
	}
	return ""
}
