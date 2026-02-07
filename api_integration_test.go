package imx

import (
	"testing"

	imxtest "github.com/gomantics/imx/internal/testing"
)

// TestIntegration_JPEG tests JPEG format end-to-end with full metadata validation
func TestIntegration_JPEG(t *testing.T) {
	meta, err := MetadataFromFile("testdata/jpeg/apple_xmp.jpg")
	if err != nil {
		t.Skipf("Test file not found: %v", err)
	}

	result := imxtest.AssertDirectories(meta.Directories(), []imxtest.DirectoryExpectation{
		{
			Name:          "ExifIFD",
			ExactTagCount: 36,
			Tags: []imxtest.TagExpectation{
				{Name: "ApertureValue", Value: "54823/32325"},
				{Name: "BrightnessValue", Value: "40874/4739"},
				{Name: "ColorSpace", Value: "Uncalibrated"},
				{Name: "ComponentsConfiguration", Value: []byte{1, 2, 3, 0}},
				{Name: "CustomRendered", Value: "Portrait HDR"},
				{Name: "DateTimeDigitized", Value: "2019:09:21 14:43:51"},
				{Name: "DateTimeOriginal", Value: "2019:09:21 14:43:51"},
				{Name: "ExifVersion", Value: "0221"},
				{Name: "ExposureBiasValue", Value: "0/1"},
				{Name: "ExposureMode", Value: "Auto"},
				{Name: "ExposureProgram", Value: "Program AE"},
				{Name: "ExposureTime", Value: "1/758"},
				{Name: "FNumber", Value: "9/5"},
				{Name: "Flash", Value: "Auto, Did not fire"},
				{Name: "FlashpixVersion", Value: "0100"},
				{Name: "FocalLength", Value: "17/4"},
				{Name: "FocalLengthIn35mmFilm", Value: uint16(26)},
				{Name: "ISOSpeedRatings", Value: uint16(32)},
				{Name: "LensMake", Value: "Apple"},
				{Name: "LensModel", Value: "iPhone 11 back dual wide camera 4.25mm f/1.8"},
				{Name: "MakerNote", Type: "[]byte"}, // Binary data - check presence only
				{Name: "MeteringMode", Value: "Multi-segment"},
				{Name: "OffsetTime", Value: "-07:00"},
				{Name: "OffsetTimeDigitized", Value: "-07:00"},
				{Name: "OffsetTimeOriginal", Value: "-07:00"},
				{Name: "PixelXDimension", Value: uint32(3024)},
				{Name: "PixelYDimension", Value: uint32(4032)},
				{Name: "SceneCaptureType", Value: "Standard"},
				{Name: "SceneType", Value: uint8(1)},
				{Name: "SensingMethod", Value: "One-chip color area"},
				{Name: "ShutterSpeedValue", Value: "373322/39029"},
				{Name: "SubSecTimeDigitized", Value: "705"},
				{Name: "SubSecTimeOriginal", Value: "705"},
				{Name: "WhiteBalance", Value: "Auto"},
				{Name: "LensSpecification"}, // Array - check presence only
				{Name: "SubjectArea"},        // Array - check presence only
			},
		},
		{
			Name:          "GPS",
			ExactTagCount: 13,
			Tags: []imxtest.TagExpectation{
				{Name: "GPSAltitude", Value: "55895/11923"},
				{Name: "GPSAltitudeRef", Value: uint8(0)},
				{Name: "GPSDestBearing", Value: "412255/1278"},
				{Name: "GPSDestBearingRef", Value: "T"},
				{Name: "GPSHPositioningError", Value: "149275/5238"},
				{Name: "GPSImgDirection", Value: "412255/1278"},
				{Name: "GPSImgDirectionRef", Value: "T"},
				{Name: "GPSLatitudeRef", Value: "N"},
				{Name: "GPSLongitudeRef", Value: "W"},
				{Name: "GPSSpeed", Value: "0/1"},
				{Name: "GPSSpeedRef", Value: "K"},
				{Name: "GPSLatitude", Type: "[]string"},  // Array - check presence only
				{Name: "GPSLongitude", Type: "[]string"}, // Array - check presence only
			},
		},
		{
			Name:          "ICC-Header",
			ExactTagCount: 19,
			Tags: []imxtest.TagExpectation{
				{Name: "CMMType", Value: "appl"},
				{Name: "ColorSpace", Value: "RGB"},
				{Name: "DateTimeCreated"}, // time.Time - check presence only
				{Name: "DeviceAttributes", Value: "Reflective, Glossy, Positive, Color"},
				{Name: "DeviceManufacturer", Value: "APPL"},
				{Name: "DeviceModel", Value: "\x00\x00\x00\x00"},
				{Name: "IlluminantX"}, // float64 - check presence only
				{Name: "IlluminantY"}, // float64 - check presence only
				{Name: "IlluminantZ"}, // float64 - check presence only
				{Name: "PrimaryPlatform", Value: "Apple"},
				{Name: "ProfileClass", Value: "Display Device Profile"},
				{Name: "ProfileConnectionSpace", Value: "XYZ"},
				{Name: "ProfileCreator", Value: "appl"},
				{Name: "ProfileFlags", Value: "Not Embedded, Independent"},
				{Name: "ProfileID", Value: "CA1A9582257F104D389913D5D1EA1582"},
				{Name: "ProfileSignature", Value: "acsp"},
				{Name: "ProfileSize", Value: uint32(548)},
				{Name: "ProfileVersion", Value: "4.0.0"},
				{Name: "RenderingIntent", Value: "Perceptual"},
			},
		},
		{
			Name:          "ICC-Profile",
			ExactTagCount: 10,
			Tags: []imxtest.TagExpectation{
				{Name: "ProfileCopyright", Value: "Copyright Apple Inc., 2017"},
				{Name: "ProfileDescription", Value: "Display P3"},
				{Name: "BlueMatrixColumn"},              // Nested object - check presence only
				{Name: "BlueToneReproductionCurve"},     // Nested object - check presence only
				{Name: "ChromaticAdaptation"},           // Array - check presence only
				{Name: "GreenMatrixColumn"},             // Nested object - check presence only
				{Name: "GreenToneReproductionCurve"},    // Nested object - check presence only
				{Name: "MediaWhitePoint"},               // Nested object - check presence only
				{Name: "RedMatrixColumn"},               // Nested object - check presence only
				{Name: "RedToneReproductionCurve"},      // Nested object - check presence only
			},
		},
		{
			Name:          "IFD0",
			ExactTagCount: 9,
			Tags: []imxtest.TagExpectation{
				{Name: "DateTime", Value: "2019:09:21 14:43:51"},
				{Name: "Make", Value: "Apple"},
				{Name: "Model", Value: "iPhone 11"},
				{Name: "Orientation", Value: "Horizontal (normal)"},
				{Name: "ResolutionUnit", Value: "inches"},
				{Name: "Software", Value: "13.0"},
				{Name: "XResolution", Value: "72/1"},
				{Name: "YCbCrPositioning", Value: "Centered"},
				{Name: "YResolution", Value: "72/1"},
			},
		},
		{
			Name:          "IFD1",
			ExactTagCount: 6,
			Tags: []imxtest.TagExpectation{
				{Name: "Compression", Value: "JPEG (old-style)"},
				{Name: "JPEGInterchangeFormat", Value: uint32(2364)},
				{Name: "JPEGInterchangeFormatLength", Value: uint32(10216)},
				{Name: "ResolutionUnit", Value: "inches"},
				{Name: "XResolution", Value: "72/1"},
				{Name: "YResolution", Value: "72/1"},
			},
		},
		{
			Name:          "IPTC-Application",
			ExactTagCount: 3,
			Tags: []imxtest.TagExpectation{
				{Name: "DateCreated", Value: "2019-09-21"},
				{Name: "RecordVersion", Value: int(4)},
				{Name: "TimeCreated", Value: "14:43:51-07:00"},
			},
		},
		{
			Name:          "IPTC-Envelope",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "CodedCharacterSet", Value: "\x1b%G"},
			},
		},
		{
			Name:          "XMP-aux",
			ExactTagCount: 2,
			Tags: []imxtest.TagExpectation{
				{Name: "Lens", Value: "iPhone 11 back dual wide camera 4.25mm f/1.8"},
				{Name: "LensInfo", Value: "807365/524263 17/4 9/5 12/5"},
			},
		},
		{
			Name:          "XMP-dc",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "format", Value: "image/jpeg"},
			},
		},
		{
			Name:          "XMP-exifEX",
			ExactTagCount: 4,
			Tags: []imxtest.TagExpectation{
				{Name: "LensMake", Value: "Apple"},
				{Name: "LensModel", Value: "iPhone 11 back dual wide camera 4.25mm f/1.8"},
				{Name: "PhotographicSensitivity", Value: int(32)},
				{Name: "LensSpecification"}, // Array - check presence only
			},
		},
		{
			Name:          "XMP-mwg-rs",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "Regions"}, // Nested object - check presence only
			},
		},
		{
			Name:          "XMP-photoshop",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "DateCreated", Value: "2019-09-21T14:43:51.705-07:00"},
			},
		},
		{
			Name:          "XMP-x",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "XMPToolkit", Value: "Adobe XMP Core 5.6-c140 79.160451, 2017/05/06-01:08:21        "},
			},
		},
		{
			Name:          "XMP-xmp",
			ExactTagCount: 5,
			Tags: []imxtest.TagExpectation{
				{Name: "CreateDate", Value: "2019-09-21T14:43:51.705-07:00"},
				{Name: "CreatorTool", Value: float64(13)},
				{Name: "MetadataDate", Value: "2019-09-25T16:54:55-07:00"},
				{Name: "ModifyDate", Value: "2019-09-21T14:43:51-07:00"},
				{Name: "Rating", Value: int(5)},
			},
		},
		{
			Name:          "XMP-xmpMM",
			ExactTagCount: 4,
			Tags: []imxtest.TagExpectation{
				{Name: "DocumentID", Value: "7DC2C86492E3FC7EE0661F2F0F6E0F35"},
				{Name: "InstanceID", Value: "xmp.iid:5deb5869-7884-4705-874c-e9cb27f507b7"},
				{Name: "OriginalDocumentID", Value: "7DC2C86492E3FC7EE0661F2F0F6E0F35"},
				{Name: "History"}, // Array of objects - check presence only
			},
		},
	})

	if result.Failed() {
		for _, err := range result.Errors {
			t.Error(err.Error())
		}
	}
}

// TestIntegration_CR2 tests CR2 format end-to-end with full metadata validation
func TestIntegration_CR2(t *testing.T) {
	meta, err := MetadataFromFile("testdata/cr2/sample1.cr2")
	if err != nil {
		t.Skipf("Test file not found: %v", err)
	}

	result := imxtest.AssertDirectories(meta.Directories(), []imxtest.DirectoryExpectation{
		{
			Name:          "IFD0",
			ExactTagCount: 13,
			Tags: []imxtest.TagExpectation{
				// All tags with exact values where possible
				{Name: "Make", Value: "Canon"},
				{Name: "Model", Value: "Canon EOS-1Ds Mark II"},
				{Name: "ImageWidth", Value: uint16(1536)},
				{Name: "ImageHeight", Value: uint16(1024)},
				{Name: "BitsPerSample", Value: []uint16{8, 8, 8}},
				{Name: "Compression", Value: "JPEG (old-style)"},
				{Name: "Orientation", Value: "Rotate 270 CW"},
				{Name: "XResolution", Value: "72/1"},
				{Name: "YResolution", Value: "72/1"},
				{Name: "ResolutionUnit", Value: "inches"},
				{Name: "DateTime", Value: "2004:11:13 23:02:21"},
				{Name: "StripOffsets", Value: uint32(10084)},
				{Name: "StripByteCounts", Value: uint32(401596)},
			},
		},
		{
			Name:          "ExifIFD",
			ExactTagCount: 27,
			Tags: []imxtest.TagExpectation{
				// All tags with exact values where possible, binary data checked for presence only
				{Name: "ExposureTime", Value: "1/100"},
				{Name: "FNumber", Value: "14/10"},
				{Name: "ExposureProgram", Value: "Program AE"},
				{Name: "ISOSpeedRatings", Value: uint16(640)},
				{Name: "ExifVersion", Value: "0221"},
				{Name: "DateTimeOriginal", Value: "2004:11:13 23:02:21"},
				{Name: "DateTimeDigitized", Value: "2004:11:13 23:02:21"},
				{Name: "ComponentsConfiguration", Value: []byte{1, 2, 3, 0}}, // Y, Cb, Cr, -
				{Name: "ShutterSpeedValue", Value: "434176/65536"},
				{Name: "ApertureValue", Value: "65536/65536"},
				{Name: "ExposureBiasValue", Value: "0/1"},
				{Name: "MeteringMode", Value: "Multi-segment"},
				{Name: "Flash", Value: "Off, Did not fire"},
				{Name: "FocalLength", Value: "85/1"},
				{Name: "MakerNote", Type: "[]byte"},   // Binary data - check presence only
				{Name: "UserComment", Type: "[]byte"}, // Binary data - check presence only
				{Name: "FlashpixVersion", Value: "0100"},
				{Name: "ColorSpace", Value: "sRGB"},
				{Name: "PixelXDimension", Value: uint16(4992)},
				{Name: "PixelYDimension", Value: uint16(3328)},
				{Name: "FocalPlaneXResolution", Value: "5008000/1420"},
				{Name: "FocalPlaneYResolution", Value: "3334000/945"},
				{Name: "FocalPlaneResolutionUnit", Value: "inches"},
				{Name: "CustomRendered", Value: "Normal"},
				{Name: "ExposureMode", Value: "Auto"},
				{Name: "WhiteBalance", Value: "Auto"},
				{Name: "SceneCaptureType", Value: "Standard"},
			},
		},
		{
			Name:          "IFD1",
			ExactTagCount: 2,
			Tags: []imxtest.TagExpectation{
				{Name: "JPEGInterchangeFormat", Value: uint32(411710)},
				{Name: "JPEGInterchangeFormatLength", Value: uint32(13120)},
			},
		},
	})
	if result.Failed() {
		for _, err := range result.Errors {
			t.Error(err.Error())
		}
	}
}

// TestIntegration_FLAC tests FLAC format end-to-end with full metadata validation
func TestIntegration_FLAC(t *testing.T) {
	meta, err := MetadataFromFile("testdata/flac/sample3_hires.flac")
	if err != nil {
		t.Skipf("Test file not found: %v", err)
	}

	result := imxtest.AssertDirectories(meta.Directories(), []imxtest.DirectoryExpectation{
		{
			Name:          "FLAC-StreamInfo",
			ExactTagCount: 10,
			Tags: []imxtest.TagExpectation{
				{Name: "MinimumBlockSize", Value: uint16(4608)},
				{Name: "MaximumBlockSize", Value: uint16(4608)},
				{Name: "MinimumFrameSize", Value: uint32(1011)},
				{Name: "MaximumFrameSize", Value: uint32(1425)},
				{Name: "SampleRate", Value: uint32(48000)},
				{Name: "Channels", Value: uint8(1)},
				{Name: "BitsPerSample", Value: uint8(16)},
				{Name: "TotalSamples", Value: uint64(192000)},
				{Name: "Duration", Value: "4.00 seconds"},
				{Name: "MD5Signature", Value: "4451532732537b635de6390990774131"},
			},
		},
		{
			Name:          "FLAC-Vorbis",
			ExactTagCount: 20,
			Tags: []imxtest.TagExpectation{
				{Name: "Vendor", Value: "Lavf60.16.100"},
				{Name: "TITLE", Value: "High Resolution Audio Test"},
				{Name: "ARTIST", Value: "FLAC Test Artist"},
				{Name: "ALBUM", Value: "Lossless Collection"},
				{Name: "ALBUMARTIST", Value: "Studio Masters"},
				{Name: "TRACKNUMBER", Value: "3"},
				{Name: "TRACKTOTAL", Value: "8"},
				{Name: "DISCNUMBER", Value: "1"},
				{Name: "DATE", Value: "2024-01-20"},
				{Name: "GENRE", Value: "Ambient"},
				{Name: "COMPOSER", Value: "Digital Composer"},
				{Name: "PERFORMER", Value: "Sine Wave Generator"},
				{Name: "DESCRIPTION", Value: "Test FLAC file for metadata parsing"},
				{Name: "ORGANIZATION", Value: "Test Organization"},
				{Name: "LOCATION", Value: "Virtual Studio"},
				{Name: "COPYRIGHT", Value: "CC0 Public Domain"},
				{Name: "LICENSE", Type: "string"},
				{Name: "ISRC", Value: "TEST00000001"},
				{Name: "REPLAYGAIN_TRACK_GAIN", Value: "-6.5 dB"},
				{Name: "encoder", Value: "Lavf60.16.100"},
			},
		},
		{
			Name:          "FLAC-Padding",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "PaddingSize", Value: "8192 bytes"},
			},
		},
	})
	if result.Failed() {
		for _, err := range result.Errors {
			t.Error(err.Error())
		}
	}
}

// TestIntegration_GIF tests GIF format end-to-end with full metadata validation
func TestIntegration_GIF(t *testing.T) {
	meta, err := MetadataFromFile("testdata/gif/animated_art.gif")
	if err != nil {
		t.Skipf("Test file not found: %v", err)
	}

	result := imxtest.AssertDirectories(meta.Directories(), []imxtest.DirectoryExpectation{
		{
			Name:          "GIF",
			ExactTagCount: 7, // Version, Width, Height, ColorMap, ColorResolution, BitsPerPixel, Background
			Tags: []imxtest.TagExpectation{
				{Name: "GIFVersion", Value: "89a"},
				{Name: "ImageWidth", Value: uint16(400)},
				{Name: "ImageHeight", Value: uint16(400)},
				{Name: "HasColorMap", Value: true},
				{Name: "ColorResolutionDepth", Value: uint8(8)},
				{Name: "BitsPerPixel", Value: uint8(8)},
				{Name: "BackgroundColor", Value: uint8(0)},
			},
		},
		{
			Name:          "XMP-tiff",
			ExactTagCount: 3,
			Tags: []imxtest.TagExpectation{
				{Name: "Artist"}, // Just check existence, XMP types vary
				{Name: "Copyright"},
				{Name: "ImageDescription"},
			},
		},
		{
			Name:          "XMP-pdf",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "Keywords"}, // Just check existence
			},
		},
		{
			Name:          "XMP-x",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "XMPToolkit"}, // XMP toolkit version
			},
		},
	})
	if result.Failed() {
		for _, err := range result.Errors {
			t.Error(err.Error())
		}
	}
}

// TestIntegration_HEIC tests HEIC format end-to-end with full metadata validation
func TestIntegration_HEIC(t *testing.T) {
	meta, err := MetadataFromFile("testdata/heic/apple_icc.HEIC")
	if err != nil {
		t.Skipf("Test file not found: %v", err)
	}

	result := imxtest.AssertDirectories(meta.Directories(), []imxtest.DirectoryExpectation{
		{
			Name:          "XMP-xmp",
			ExactTagCount: 3,
			Tags: []imxtest.TagExpectation{
				{Name: "ModifyDate"},
				{Name: "CreateDate"},
				{Name: "CreatorTool"},
			},
		},
		{
			Name:          "XMP-photoshop",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "DateCreated"},
			},
		},
		{
			Name:          "XMP-x",
			ExactTagCount: 1,
			Tags: []imxtest.TagExpectation{
				{Name: "XMPToolkit"},
			},
		},
		{
			Name:          "IFD0",
			ExactTagCount: 10,
			Tags: []imxtest.TagExpectation{
				{Name: "Make", Value: "Apple"},
				{Name: "Model", Value: "iPhone 11 Pro Max"},
				{Name: "Software", Value: "14.4.2"},
				{Name: "HostComputer", Value: "iPhone 11 Pro Max"},
				{Name: "XResolution"},
				{Name: "YResolution"},
				{Name: "ResolutionUnit"},
				{Name: "DateTime"},
				{Name: "TileWidth"},
				{Name: "TileLength"},
			},
		},
		{
			Name:          "ExifIFD",
			ExactTagCount: 36,
			Tags: []imxtest.TagExpectation{
				{Name: "ExposureTime"},
				{Name: "FNumber"},
				{Name: "ExposureProgram"},
				{Name: "ISOSpeedRatings"},
				{Name: "ExifVersion"},
				{Name: "DateTimeOriginal"},
				{Name: "DateTimeDigitized"},
				{Name: "OffsetTime"},
				{Name: "OffsetTimeOriginal"},
				{Name: "OffsetTimeDigitized"},
				{Name: "ComponentsConfiguration"},
				{Name: "ShutterSpeedValue"},
				{Name: "ApertureValue"},
				{Name: "BrightnessValue"},
				{Name: "ExposureBiasValue"},
				{Name: "MeteringMode"},
				{Name: "Flash"},
				{Name: "FocalLength"},
				{Name: "SubjectArea"},
				{Name: "MakerNote"},
				{Name: "SubSecTimeOriginal"},
				{Name: "SubSecTimeDigitized"},
				{Name: "FlashpixVersion"},
				{Name: "ColorSpace"},
				{Name: "PixelXDimension"},
				{Name: "PixelYDimension"},
				{Name: "SensingMethod"},
				{Name: "SceneType"},
				{Name: "ExposureMode"},
				{Name: "WhiteBalance"},
				{Name: "FocalLengthIn35mmFilm"},
				{Name: "SceneCaptureType"},
				{Name: "LensSpecification"},
				{Name: "LensMake"},
				{Name: "LensModel"},
				{Name: "CompositeImage"},
			},
		},
		{
			Name:          "GPS",
			ExactTagCount: 13,
			Tags: []imxtest.TagExpectation{
				{Name: "GPSLatitudeRef"},
				{Name: "GPSLatitude"},
				{Name: "GPSLongitudeRef"},
				{Name: "GPSLongitude"},
				{Name: "GPSAltitudeRef"},
				{Name: "GPSAltitude"},
				{Name: "GPSSpeedRef"},
				{Name: "GPSSpeed"},
				{Name: "GPSImgDirectionRef"},
				{Name: "GPSImgDirection"},
				{Name: "GPSDestBearingRef"},
				{Name: "GPSDestBearing"},
				{Name: "GPSHPositioningError"},
			},
		},
		{
			Name:          "ICC-Header",
			ExactTagCount: 19,
			Tags: []imxtest.TagExpectation{
				{Name: "ProfileSize"},
				{Name: "CMMType"},
				{Name: "ProfileVersion"},
				{Name: "ProfileClass"},
				{Name: "ColorSpace"},
				{Name: "ProfileConnectionSpace"},
				{Name: "DateTimeCreated"},
				{Name: "ProfileSignature"},
				{Name: "PrimaryPlatform"},
				{Name: "ProfileFlags"},
				{Name: "DeviceManufacturer"},
				{Name: "DeviceModel"},
				{Name: "DeviceAttributes"},
				{Name: "RenderingIntent"},
				{Name: "IlluminantX"},
				{Name: "IlluminantY"},
				{Name: "IlluminantZ"},
				{Name: "ProfileCreator"},
				{Name: "ProfileID"},
			},
		},
		{
			Name:          "ICC-Profile",
			ExactTagCount: 10,
			Tags: []imxtest.TagExpectation{
				{Name: "ProfileDescription"},
				{Name: "ProfileCopyright"},
				{Name: "MediaWhitePoint"},
				{Name: "RedMatrixColumn"},
				{Name: "GreenMatrixColumn"},
				{Name: "BlueMatrixColumn"},
				{Name: "RedToneReproductionCurve"},
				{Name: "ChromaticAdaptation"},
				{Name: "BlueToneReproductionCurve"},
				{Name: "GreenToneReproductionCurve"},
			},
		},
	})
	if result.Failed() {
		for _, err := range result.Errors {
			t.Error(err.Error())
		}
	}
}
