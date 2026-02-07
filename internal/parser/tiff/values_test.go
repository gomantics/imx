package tiff

import "testing"

func TestDecodeEnumValue(t *testing.T) {
	tests := []struct {
		name     string
		tag      uint16
		dirName  string
		value    any
		expected string
	}{
		// TIFF tags
		{
			name:     "Orientation Horizontal",
			tag:      0x0112,
			dirName:  "IFD0",
			value:    uint16(1),
			expected: "Horizontal (normal)",
		},
		{
			name:     "Orientation Rotate 90 CW",
			tag:      0x0112,
			dirName:  "IFD0",
			value:    uint16(6),
			expected: "Rotate 90 CW",
		},
		{
			name:     "ResolutionUnit None",
			tag:      0x0128,
			dirName:  "IFD0",
			value:    uint16(1),
			expected: "None",
		},
		{
			name:     "ResolutionUnit inches",
			tag:      0x0128,
			dirName:  "IFD0",
			value:    uint16(2),
			expected: "inches",
		},
		{
			name:     "Compression Uncompressed",
			tag:      0x0103,
			dirName:  "IFD0",
			value:    uint16(1),
			expected: "Uncompressed",
		},
		{
			name:     "Compression JPEG",
			tag:      0x0103,
			dirName:  "IFD0",
			value:    uint16(7),
			expected: "JPEG",
		},
		{
			name:     "YCbCrPositioning Centered",
			tag:      0x0213,
			dirName:  "IFD0",
			value:    uint16(1),
			expected: "Centered",
		},

		// EXIF tags
		{
			name:     "ColorSpace sRGB",
			tag:      0xA001,
			dirName:  "ExifIFD",
			value:    uint16(1),
			expected: "sRGB",
		},
		{
			name:     "ColorSpace Adobe RGB",
			tag:      0xA001,
			dirName:  "ExifIFD",
			value:    uint16(2),
			expected: "Adobe RGB",
		},
		{
			name:     "ColorSpace Uncalibrated",
			tag:      0xA001,
			dirName:  "ExifIFD",
			value:    uint16(65535),
			expected: "Uncalibrated",
		},
		{
			name:     "ExposureMode Auto",
			tag:      0xA402,
			dirName:  "ExifIFD",
			value:    uint16(0),
			expected: "Auto",
		},
		{
			name:     "ExposureMode Manual",
			tag:      0xA402,
			dirName:  "ExifIFD",
			value:    uint16(1),
			expected: "Manual",
		},
		{
			name:     "ExposureMode Auto bracket",
			tag:      0xA402,
			dirName:  "ExifIFD",
			value:    uint16(2),
			expected: "Auto bracket",
		},
		{
			name:     "WhiteBalance Auto",
			tag:      0xA403,
			dirName:  "ExifIFD",
			value:    uint16(0),
			expected: "Auto",
		},
		{
			name:     "WhiteBalance Manual",
			tag:      0xA403,
			dirName:  "ExifIFD",
			value:    uint16(1),
			expected: "Manual",
		},
		{
			name:     "SceneCaptureType Standard",
			tag:      0xA406,
			dirName:  "ExifIFD",
			value:    uint16(0),
			expected: "Standard",
		},
		{
			name:     "SceneCaptureType Landscape",
			tag:      0xA406,
			dirName:  "ExifIFD",
			value:    uint16(1),
			expected: "Landscape",
		},
		{
			name:     "SceneCaptureType Portrait",
			tag:      0xA406,
			dirName:  "ExifIFD",
			value:    uint16(2),
			expected: "Portrait",
		},
		{
			name:     "SceneCaptureType Night",
			tag:      0xA406,
			dirName:  "ExifIFD",
			value:    uint16(3),
			expected: "Night",
		},
		{
			name:     "ExposureProgram Manual",
			tag:      0x8822,
			dirName:  "ExifIFD",
			value:    uint16(1),
			expected: "Manual",
		},
		{
			name:     "ExposureProgram Aperture-priority",
			tag:      0x8822,
			dirName:  "ExifIFD",
			value:    uint16(3),
			expected: "Aperture-priority AE",
		},
		{
			name:     "MeteringMode Multi-segment",
			tag:      0x9207,
			dirName:  "ExifIFD",
			value:    uint16(5),
			expected: "Multi-segment",
		},
		{
			name:     "MeteringMode Spot",
			tag:      0x9207,
			dirName:  "ExifIFD",
			value:    uint16(3),
			expected: "Spot",
		},
		{
			name:     "Flash No Flash",
			tag:      0x9209,
			dirName:  "ExifIFD",
			value:    uint16(0x00),
			expected: "No Flash",
		},
		{
			name:     "Flash Fired",
			tag:      0x9209,
			dirName:  "ExifIFD",
			value:    uint16(0x01),
			expected: "Fired",
		},
		{
			name:     "Flash Off Did not fire",
			tag:      0x9209,
			dirName:  "ExifIFD",
			value:    uint16(0x10),
			expected: "Off, Did not fire",
		},
		{
			name:     "Flash Auto Fired",
			tag:      0x9209,
			dirName:  "ExifIFD",
			value:    uint16(0x19),
			expected: "Auto, Fired",
		},
		{
			name:     "LightSource Daylight",
			tag:      0x9208,
			dirName:  "ExifIFD",
			value:    uint16(1),
			expected: "Daylight",
		},
		{
			name:     "LightSource Fluorescent",
			tag:      0x9208,
			dirName:  "ExifIFD",
			value:    uint16(2),
			expected: "Fluorescent",
		},
		{
			name:     "Contrast Normal",
			tag:      0xA408,
			dirName:  "ExifIFD",
			value:    uint16(0),
			expected: "Normal",
		},
		{
			name:     "Saturation Low",
			tag:      0xA409,
			dirName:  "ExifIFD",
			value:    uint16(1),
			expected: "Low",
		},
		{
			name:     "Sharpness Hard",
			tag:      0xA40A,
			dirName:  "ExifIFD",
			value:    uint16(2),
			expected: "Hard",
		},

		// Edge cases
		{
			name:     "uint32 value",
			tag:      0xA001,
			dirName:  "ExifIFD",
			value:    uint32(1),
			expected: "sRGB",
		},
		{
			name:     "int value",
			tag:      0xA001,
			dirName:  "ExifIFD",
			value:    int(1),
			expected: "sRGB",
		},
		{
			name:     "unknown value returns empty",
			tag:      0xA001,
			dirName:  "ExifIFD",
			value:    uint16(999),
			expected: "",
		},
		{
			name:     "unknown tag returns empty",
			tag:      0x9999,
			dirName:  "ExifIFD",
			value:    uint16(1),
			expected: "",
		},
		{
			name:     "string value returns empty",
			tag:      0xA001,
			dirName:  "ExifIFD",
			value:    "not a number",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeEnumValue(tt.tag, tt.dirName, tt.value)
			if result != tt.expected {
				t.Errorf("decodeEnumValue(0x%04X, %q, %v) = %q, want %q",
					tt.tag, tt.dirName, tt.value, result, tt.expected)
			}
		})
	}
}

func TestDecodeFlashValue(t *testing.T) {
	tests := []struct {
		value    uint16
		expected string
	}{
		{0x00, "No Flash"},
		{0x01, "Fired"},
		{0x05, "Fired, Return not detected"},
		{0x07, "Fired, Return detected"},
		{0x10, "Off, Did not fire"},
		{0x18, "Auto, Did not fire"},
		{0x19, "Auto, Fired"},
		{0x41, "Fired, Red-eye reduction"},
		{0x59, "Auto, Fired, Red-eye reduction"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := decodeFlashValue(tt.value)
			if result != tt.expected {
				t.Errorf("decodeFlashValue(0x%02X) = %q, want %q",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestEnumMappingsComplete ensures all documented enums are present
func TestEnumMappingsComplete(t *testing.T) {
	// EXIF ColorSpace must have sRGB
	if exifEnumValues[0xA001][1] != "sRGB" {
		t.Error("ColorSpace 1 should be sRGB")
	}

	// EXIF ExposureMode must have Auto, Manual, Auto bracket
	if len(exifEnumValues[0xA402]) != 3 {
		t.Errorf("ExposureMode should have 3 values, got %d", len(exifEnumValues[0xA402]))
	}

	// TIFF Orientation must have all 8 values
	if len(tiffEnumValues[0x0112]) != 8 {
		t.Errorf("Orientation should have 8 values, got %d", len(tiffEnumValues[0x0112]))
	}

	// TIFF ResolutionUnit must have None, inches, centimeters
	if len(tiffEnumValues[0x0128]) != 3 {
		t.Errorf("ResolutionUnit should have 3 values, got %d", len(tiffEnumValues[0x0128]))
	}
}

func TestDecodeComponentsConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{
			name:     "YCbCr standard",
			value:    []byte{1, 2, 3, 0},
			expected: "Y, Cb, Cr, -",
		},
		{
			name:     "RGB",
			value:    []byte{4, 5, 6, 0},
			expected: "R, G, B, -",
		},
		{
			name:     "hex string YCbCr",
			value:    "01020300",
			expected: "Y, Cb, Cr, -",
		},
		{
			name:     "hex string RGB",
			value:    "04050600",
			expected: "R, G, B, -",
		},
		{
			name:     "empty bytes",
			value:    []byte{0, 0, 0, 0},
			expected: "-, -, -, -",
		},
		{
			name:     "too short",
			value:    []byte{1, 2},
			expected: "",
		},
		{
			name:     "invalid type",
			value:    123,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeComponentsConfiguration(tt.value)
			if result != tt.expected {
				t.Errorf("decodeComponentsConfiguration(%v) = %q, want %q",
					tt.value, result, tt.expected)
			}
		})
	}
}

func TestDecodeEnumValue_ComponentsConfiguration(t *testing.T) {
	// Test that decodeEnumValue handles ComponentsConfiguration (0x9101)
	result := decodeEnumValue(0x9101, "ExifIFD", []byte{1, 2, 3, 0})
	expected := "Y, Cb, Cr, -"
	if result != expected {
		t.Errorf("decodeEnumValue for ComponentsConfiguration = %q, want %q", result, expected)
	}
}
