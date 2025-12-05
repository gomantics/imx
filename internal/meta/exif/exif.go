package exif

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gomantics/imx/internal/container"
	"github.com/gomantics/imx/internal/meta"
	"github.com/gomantics/imx/internal/types"
)

// Parser implements meta.Parser for EXIF
type Parser struct{}

// Namespace returns the EXIF namespace
func (p *Parser) Namespace() types.Namespace {
	return types.NamespaceEXIF
}

// Parse extracts EXIF data from raw blocks
func (p *Parser) Parse(blocks []container.RawBlock, cfg types.ExtractorConfig) ([]meta.Directory, error) {
	var dirs []meta.Directory

	for _, block := range blocks {
		if block.Kind != container.MetaKindEXIF {
			continue
		}

		// Parse TIFF structure
		blockDirs, err := p.parseTIFF(block.Payload)
		if err != nil {
			return nil, fmt.Errorf("parse TIFF: %w", err)
		}

		dirs = append(dirs, blockDirs...)
	}

	return dirs, nil
}

// parseTIFF parses a TIFF-formatted EXIF block
func (p *Parser) parseTIFF(data []byte) ([]meta.Directory, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("TIFF header too short")
	}

	// Read byte order (first 2 bytes)
	var byteOrder binary.ByteOrder
	if data[0] == 'I' && data[1] == 'I' {
		byteOrder = binary.LittleEndian // Intel
	} else if data[0] == 'M' && data[1] == 'M' {
		byteOrder = binary.BigEndian // Motorola
	} else {
		return nil, fmt.Errorf("invalid TIFF byte order: %02X %02X", data[0], data[1])
	}

	// Verify TIFF magic number (should be 42)
	magic := byteOrder.Uint16(data[2:4])
	if magic != 42 {
		return nil, fmt.Errorf("invalid TIFF magic number: %d", magic)
	}

	// Read offset to first IFD
	ifd0Offset := byteOrder.Uint32(data[4:8])

	var dirs []meta.Directory

	// Parse IFD0
	if ifd0Offset > 0 && int(ifd0Offset) < len(data) {
		ifd0, nextOffset, err := p.parseIFD(data, int(ifd0Offset), byteOrder, "IFD0")
		if err != nil {
			return nil, fmt.Errorf("parse IFD0: %w", err)
		}
		dirs = append(dirs, ifd0)

		// Check for EXIF sub-IFD pointer
		if exifOffset, ok := ifd0.Tags["Exif:ExifOffset"]; ok {
			if offset, ok := exifOffset.Value.(int); ok && offset > 0 && offset < len(data) {
				exifIFD, _, err := p.parseIFD(data, offset, byteOrder, "ExifIFD")
				if err == nil {
					dirs = append(dirs, exifIFD)
				}
			}
		}

		// Check for GPS sub-IFD pointer
		if gpsOffset, ok := ifd0.Tags["Exif:GPSInfo"]; ok {
			if offset, ok := gpsOffset.Value.(int); ok && offset > 0 && offset < len(data) {
				gpsIFD, _, err := p.parseIFD(data, offset, byteOrder, "GPS")
				if err == nil {
					dirs = append(dirs, gpsIFD)
				}
			}
		}

		// Parse IFD1 (thumbnail) if present
		if nextOffset > 0 && int(nextOffset) < len(data) {
			ifd1, _, err := p.parseIFD(data, int(nextOffset), byteOrder, "IFD1")
			if err == nil {
				dirs = append(dirs, ifd1)
			}
		}
	}

	return dirs, nil
}

// parseIFD parses a single IFD (Image File Directory)
func (p *Parser) parseIFD(data []byte, offset int, byteOrder binary.ByteOrder, name string) (meta.Directory, uint32, error) {
	if offset+2 > len(data) {
		return meta.Directory{}, 0, fmt.Errorf("IFD offset out of bounds")
	}

	// Read number of entries
	entryCount := byteOrder.Uint16(data[offset : offset+2])
	offset += 2

	dir := meta.Directory{
		Namespace: types.NamespaceEXIF,
		Name:      name,
		Tags:      make(map[meta.TagID]meta.Tag),
	}

	// Parse each entry (12 bytes each)
	for i := 0; i < int(entryCount); i++ {
		if offset+12 > len(data) {
			break
		}

		tag := p.parseEntry(data, offset, byteOrder, name)
		if tag.ID != "" {
			dir.Tags[tag.ID] = tag
		}

		offset += 12
	}

	// Read offset to next IFD
	var nextOffset uint32
	if offset+4 <= len(data) {
		nextOffset = byteOrder.Uint32(data[offset : offset+4])
	}

	return dir, nextOffset, nil
}

// parseEntry parses a single IFD entry (tag)
func (p *Parser) parseEntry(data []byte, offset int, byteOrder binary.ByteOrder, ifdName string) meta.Tag {
	tagID := byteOrder.Uint16(data[offset : offset+2])
	tagType := byteOrder.Uint16(data[offset+2 : offset+4])
	count := byteOrder.Uint32(data[offset+4 : offset+8])
	valueOffset := offset + 8 // Last 4 bytes contain value or offset

	tag := meta.Tag{
		Namespace: types.NamespaceEXIF,
	}

	// Get tag name and ID based on IFD
	var tagName string
	var ok bool

	if ifdName == "GPS" {
		// GPS tags have their own namespace because they conflict with main EXIF tags
		tagName, ok = gpsTags[tagID]
	} else {
		// All other tags (IFD0, ExifIFD, InteropIFD, IFD1) use the main tag map
		tagName, ok = knownTags[tagID]
	}

	if !ok {
		tagName = fmt.Sprintf("Tag%04X", tagID)
	}
	tag.ID = meta.TagID(fmt.Sprintf("Exif:%s", tagName))
	tag.Name = tagName

	// Parse value based on type
	value, typeName := p.parseValue(data, tagType, count, valueOffset, byteOrder)
	tag.Value = value
	tag.Type = typeName

	// Store raw bytes (4 bytes of value/offset)
	tag.Raw = make([]byte, 4)
	copy(tag.Raw, data[valueOffset:valueOffset+4])

	return tag
}

// parseValue parses a tag value based on its type
func (p *Parser) parseValue(data []byte, tagType uint16, count uint32, offset int, byteOrder binary.ByteOrder) (any, string) {
	// Type sizes in bytes
	typeSizes := map[uint16]int{
		1:  1, // BYTE
		2:  1, // ASCII
		3:  2, // SHORT
		4:  4, // LONG
		5:  8, // RATIONAL
		6:  1, // SBYTE
		7:  1, // UNDEFINED
		8:  2, // SSHORT
		9:  4, // SLONG
		10: 8, // SRATIONAL
	}

	typeSize, ok := typeSizes[tagType]
	if !ok {
		return nil, "unknown"
	}

	totalSize := int(count) * typeSize

	// If value fits in 4 bytes, it's stored directly in the offset field
	// Otherwise, the offset field points to the actual data
	var valueData []byte
	if totalSize <= 4 {
		valueData = data[offset : offset+4]
	} else {
		// Read offset to actual value
		valueOffset := int(byteOrder.Uint32(data[offset : offset+4]))
		if valueOffset+totalSize > len(data) {
			return nil, "invalid_offset"
		}
		valueData = data[valueOffset : valueOffset+totalSize]
	}

	switch tagType {
	case 1: // BYTE
		if count == 1 {
			return int(valueData[0]), "byte"
		}
		return valueData[:count], "bytes"

	case 2: // ASCII string
		// Remove trailing null bytes
		str := string(bytes.TrimRight(valueData[:count], "\x00"))
		return str, "string"

	case 3: // SHORT (uint16)
		if count == 1 {
			return int(byteOrder.Uint16(valueData)), "short"
		}
		vals := make([]int, count)
		for i := uint32(0); i < count; i++ {
			vals[i] = int(byteOrder.Uint16(valueData[i*2:]))
		}
		return vals, "shorts"

	case 4: // LONG (uint32)
		if count == 1 {
			return int(byteOrder.Uint32(valueData)), "long"
		}
		vals := make([]int, count)
		for i := uint32(0); i < count; i++ {
			vals[i] = int(byteOrder.Uint32(valueData[i*4:]))
		}
		return vals, "longs"

	case 5: // RATIONAL (num/denom as uint32)
		if count == 1 {
			num := byteOrder.Uint32(valueData[0:4])
			denom := byteOrder.Uint32(valueData[4:8])
			if denom == 0 {
				return 0.0, "rational"
			}
			return float64(num) / float64(denom), "rational"
		}
		vals := make([]float64, count)
		for i := uint32(0); i < count; i++ {
			num := byteOrder.Uint32(valueData[i*8:])
			denom := byteOrder.Uint32(valueData[i*8+4:])
			if denom == 0 {
				vals[i] = 0
			} else {
				vals[i] = float64(num) / float64(denom)
			}
		}
		return vals, "rationals"

	case 7: // UNDEFINED (raw bytes)
		return valueData[:count], "undefined"

	default:
		return valueData[:count], fmt.Sprintf("type_%d", tagType)
	}
}

// knownTags maps EXIF tag IDs to names
// Based on ExifTool tag specification
var knownTags = map[uint16]string{
	// IFD0 tags
	0x000B: "ProcessingSoftware",
	0x00FE: "SubfileType",
	0x00FF: "OldSubfileType",
	0x0100: "ImageWidth",
	0x0101: "ImageHeight",
	0x0102: "BitsPerSample",
	0x0103: "Compression",
	0x0106: "PhotometricInterpretation",
	0x0107: "Thresholding",
	0x0108: "CellWidth",
	0x0109: "CellLength",
	0x010A: "FillOrder",
	0x010D: "DocumentName",
	0x010E: "ImageDescription",
	0x010F: "Make",
	0x0110: "Model",
	0x0111: "StripOffsets",
	0x0112: "Orientation",
	0x0115: "SamplesPerPixel",
	0x0116: "RowsPerStrip",
	0x0117: "StripByteCounts",
	0x0118: "MinSampleValue",
	0x0119: "MaxSampleValue",
	0x011A: "XResolution",
	0x011B: "YResolution",
	0x011C: "PlanarConfiguration",
	0x011D: "PageName",
	0x011E: "XPosition",
	0x011F: "YPosition",
	0x0122: "GrayResponseUnit",
	0x0128: "ResolutionUnit",
	0x0129: "PageNumber",
	0x012D: "TransferFunction",
	0x0131: "Software",
	0x0132: "ModifyDate",
	0x013B: "Artist",
	0x013C: "HostComputer",
	0x013D: "Predictor",
	0x013E: "WhitePoint",
	0x013F: "PrimaryChromaticities",
	0x0141: "HalftoneHints",
	0x0142: "TileWidth",
	0x0143: "TileLength",
	0x014A: "SubIFD",
	0x014C: "InkSet",
	0x0150: "DotRange",
	0x0151: "TargetPrinter",
	0x0152: "ExtraSamples",
	0x0153: "SampleFormat",
	0x015B: "JPEGTables",
	0x0201: "ThumbnailOffset",
	0x0202: "ThumbnailLength",
	0x0211: "YCbCrCoefficients",
	0x0212: "YCbCrSubSampling",
	0x0213: "YCbCrPositioning",
	0x0214: "ReferenceBlackWhite",
	0x02BC: "ApplicationNotes",
	0x4746: "Rating",
	0x4749: "RatingPercent",
	0x828D: "CFARepeatPatternDim",
	0x828E: "CFAPattern2",
	0x828F: "BatteryLevel",
	0x8298: "Copyright",
	0x829A: "ExposureTime",
	0x829D: "FNumber",
	0x83BB: "IPTC-NAA",
	0x8649: "PhotoshopSettings",
	0x8769: "ExifOffset",
	0x8773: "ICC_Profile",
	0x8822: "ExposureProgram",
	0x8824: "SpectralSensitivity",
	0x8825: "GPSInfo",
	0x8827: "ISO",
	0x8828: "OECF",
	0x882A: "TimeZoneOffset",
	0x882B: "SelfTimerMode",
	0x8830: "SensitivityType",
	0x8831: "StandardOutputSensitivity",
	0x8832: "RecommendedExposureIndex",
	0x8833: "ISOSpeed",
	0x8834: "ISOSpeedLatitudeyyy",
	0x8835: "ISOSpeedLatitudezzz",
	0x9000: "ExifVersion",
	0x9003: "DateTimeOriginal",
	0x9004: "CreateDate",
	0x9010: "OffsetTime",
	0x9011: "OffsetTimeOriginal",
	0x9012: "OffsetTimeDigitized",
	0x9101: "ComponentsConfiguration",
	0x9102: "CompressedBitsPerPixel",
	0x9201: "ShutterSpeedValue",
	0x9202: "ApertureValue",
	0x9203: "BrightnessValue",
	0x9204: "ExposureCompensation",
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
	0xA002: "ExifImageWidth",
	0xA003: "ExifImageHeight",
	0xA004: "RelatedSoundFile",
	0xA005: "InteropOffset",
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
	0xA405: "FocalLengthIn35mmFormat",
	0xA406: "SceneCaptureType",
	0xA407: "GainControl",
	0xA408: "Contrast",
	0xA409: "Saturation",
	0xA40A: "Sharpness",
	0xA40B: "DeviceSettingDescription",
	0xA40C: "SubjectDistanceRange",
	0xA420: "ImageUniqueID",
	0xA430: "OwnerName",
	0xA431: "SerialNumber",
	0xA432: "LensInfo",
	0xA433: "LensMake",
	0xA434: "LensModel",
	0xA435: "LensSerialNumber",
	0xA436: "ImageTitle",
	0xA437: "Photographer",
	0xA438: "ImageEditor",
	0xA439: "CameraFirmware",
	0xA43A: "RAWDevelopingSoftware",
	0xA43B: "ImageEditingSoftware",
	0xA43C: "MetadataEditingSoftware",
	0xA460: "CompositeImage",
	0xA461: "CompositeImageCount",
	0xA462: "CompositeImageExposureTimes",
	0xA500: "Gamma",
	0x9C9B: "XPTitle",
	0x9C9C: "XPComment",
	0x9C9D: "XPAuthor",
	0x9C9E: "XPKeywords",
	0x9C9F: "XPSubject",

	// InteropIFD tags (part of EXIF, no conflict)
	0x0001: "InteropIndex",
	0x0002: "InteropVersion",
	0x1000: "RelatedImageFileFormat",
	0x1001: "RelatedImageWidth",
	0x1002: "RelatedImageHeight",
}

// GPS-specific tags (separate because they conflict with main EXIF tag IDs)
var gpsTags = map[uint16]string{
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

// Post-process specific tags (called after all parsing is done)
func postProcessTags(dirs []meta.Directory) {
	for i := range dirs {
		dir := &dirs[i]

		// Parse DateTimeOriginal into time.Time
		if tag, ok := dir.Tags["Exif:DateTimeOriginal"]; ok {
			if str, ok := tag.Value.(string); ok {
				// EXIF datetime format: "YYYY:MM:DD HH:MM:SS"
				t, err := time.Parse("2006:01:02 15:04:05", str)
				if err == nil {
					tag.Value = t
					tag.Type = "time"
					dir.Tags["Exif:DateTimeOriginal"] = tag
				}
			}
		}

		// Parse DateTime into time.Time
		if tag, ok := dir.Tags["Exif:DateTime"]; ok {
			if str, ok := tag.Value.(string); ok {
				t, err := time.Parse("2006:01:02 15:04:05", str)
				if err == nil {
					tag.Value = t
					tag.Type = "time"
					dir.Tags["Exif:DateTime"] = tag
				}
			}
		}
	}
}

// Parse with post-processing
func (p *Parser) ParseWithPostProcessing(blocks []container.RawBlock, cfg types.ExtractorConfig) ([]meta.Directory, error) {
	dirs, err := p.Parse(blocks, cfg)
	if err != nil {
		return nil, err
	}

	postProcessTags(dirs)
	return dirs, nil
}

func init() {
	parser := &Parser{}
	meta.Register(parser)
}
