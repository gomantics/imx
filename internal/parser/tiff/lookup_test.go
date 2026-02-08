package tiff

import (
	"testing"
)

func TestGetTIFFTagName(t *testing.T) {
	tests := []struct {
		name string
		tag  uint16
		want string
	}{
		{"ImageWidth", 0x0100, "ImageWidth"},
		{"ImageHeight", 0x0101, "ImageHeight"},
		{"BitsPerSample", 0x0102, "BitsPerSample"},
		{"Compression", 0x0103, "Compression"},
		{"Make", 0x010F, "Make"},
		{"Model", 0x0110, "Model"},
		{"Orientation", 0x0112, "Orientation"},
		{"XResolution", 0x011A, "XResolution"},
		{"YResolution", 0x011B, "YResolution"},
		{"Software", 0x0131, "Software"},
		{"DateTime", 0x0132, "DateTime"},
		{"Artist", 0x013B, "Artist"},
		{"FillOrder", 0x010A, "FillOrder"},
		{"PageNumber", 0x0129, "PageNumber"},
		{"Predictor", 0x013D, "Predictor"},
		{"ExtraSamples", 0x0152, "ExtraSamples"},
		{"SampleFormat", 0x0153, "SampleFormat"},
		{"Copyright", 0x8298, "Copyright"},
		{"ExifIFDPointer", 0x8769, "ExifIFDPointer"},
		{"GPSInfoIFDPointer", 0x8825, "GPSInfoIFDPointer"},
		{"Unknown tag", 0xFFFF, ""},
		{"Unknown tag 0", 0x0000, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTIFFTagName(tt.tag); got != tt.want {
				t.Errorf("getTIFFTagName(0x%04X) = %q, want %q", tt.tag, got, tt.want)
			}
		})
	}
}

func TestGetEXIFTagName(t *testing.T) {
	tests := []struct {
		name string
		tag  uint16
		want string
	}{
		{"ExposureTime", 0x829A, "ExposureTime"},
		{"FNumber", 0x829D, "FNumber"},
		{"ExposureProgram", 0x8822, "ExposureProgram"},
		{"ISOSpeedRatings", 0x8827, "ISOSpeedRatings"},
		{"ExifVersion", 0x9000, "ExifVersion"},
		{"DateTimeOriginal", 0x9003, "DateTimeOriginal"},
		{"DateTimeDigitized", 0x9004, "DateTimeDigitized"},
		{"ShutterSpeedValue", 0x9201, "ShutterSpeedValue"},
		{"ApertureValue", 0x9202, "ApertureValue"},
		{"Flash", 0x9209, "Flash"},
		{"FocalLength", 0x920A, "FocalLength"},
		{"MakerNote", 0x927C, "MakerNote"},
		{"ColorSpace", 0xA001, "ColorSpace"},
		{"PixelXDimension", 0xA002, "PixelXDimension"},
		{"PixelYDimension", 0xA003, "PixelYDimension"},
		{"LensModel", 0xA434, "LensModel"},
		{"Unknown EXIF tag", 0x0001, ""},
		{"Unknown EXIF tag 2", 0xFFFF, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEXIFTagName(tt.tag); got != tt.want {
				t.Errorf("getEXIFTagName(0x%04X) = %q, want %q", tt.tag, got, tt.want)
			}
		})
	}
}

func TestGetGPSTagName(t *testing.T) {
	tests := []struct {
		name string
		tag  uint16
		want string
	}{
		{"GPSVersionID", 0x0000, "GPSVersionID"},
		{"GPSLatitudeRef", 0x0001, "GPSLatitudeRef"},
		{"GPSLatitude", 0x0002, "GPSLatitude"},
		{"GPSLongitudeRef", 0x0003, "GPSLongitudeRef"},
		{"GPSLongitude", 0x0004, "GPSLongitude"},
		{"GPSAltitudeRef", 0x0005, "GPSAltitudeRef"},
		{"GPSAltitude", 0x0006, "GPSAltitude"},
		{"GPSTimeStamp", 0x0007, "GPSTimeStamp"},
		{"GPSDateStamp", 0x001D, "GPSDateStamp"},
		{"GPSDifferential", 0x001E, "GPSDifferential"},
		{"GPSHPositioningError", 0x001F, "GPSHPositioningError"},
		{"Unknown GPS tag", 0x00FF, ""},
		{"Unknown GPS tag 2", 0xFFFF, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getGPSTagName(tt.tag); got != tt.want {
				t.Errorf("getGPSTagName(0x%04X) = %q, want %q", tt.tag, got, tt.want)
			}
		})
	}
}

func TestGetInteropTagName(t *testing.T) {
	tests := []struct {
		name string
		tag  uint16
		want string
	}{
		{"InteroperabilityIndex", 0x0001, "InteroperabilityIndex"},
		{"InteroperabilityVersion", 0x0002, "InteroperabilityVersion"},
		{"RelatedImageFileFormat", 0x1000, "RelatedImageFileFormat"},
		{"RelatedImageWidth", 0x1001, "RelatedImageWidth"},
		{"RelatedImageHeight", 0x1002, "RelatedImageHeight"},
		{"Unknown Interop tag", 0x0000, ""},
		{"Unknown Interop tag 2", 0xFFFF, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getInteropTagName(tt.tag); got != tt.want {
				t.Errorf("getInteropTagName(0x%04X) = %q, want %q", tt.tag, got, tt.want)
			}
		})
	}
}

func TestTagNameMaps_NotEmpty(t *testing.T) {
	tests := []struct {
		name string
		m    map[uint16]string
	}{
		{"tiffTagNames", tiffTagNames},
		{"exifTagNames", exifTagNames},
		{"gpsTagNames", gpsTagNames},
		{"interopTagNames", interopTagNames},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.m) == 0 {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}
