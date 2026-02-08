package canon

import (
	"testing"
)

func TestDecodeCameraSettingsValue(t *testing.T) {
	tests := []struct {
		index     int
		value     uint16
		wantName  string
		wantValue string
	}{
		{csQuality, 4, "Quality", "RAW"},
		{csQuality, 3, "Quality", "Fine"},
		{csFlashMode, 0, "FlashMode", "Off"},
		{csFlashMode, 1, "FlashMode", "Auto"},
		{csDriveMode, 0, "DriveMode", "Single"},
		{csDriveMode, 1, "DriveMode", "Continuous"},
		{csFocusMode, 0, "FocusMode", "One-shot AF"},
		{csFocusMode, 1, "FocusMode", "AI Servo AF"},
		{csFocusMode, 3, "FocusMode", "Manual Focus"},
		{csRecordMode, 6, "RecordMode", "CR2"},
		{csImageSize, 0, "ImageSize", "Large"},
		{csEasyMode, 0, "EasyMode", "Full auto"},
		{csEasyMode, 1, "EasyMode", "Manual"},
		{csMeteringMode, 3, "MeteringMode", "Evaluative"},
		{csExposureMode, 1, "ExposureMode", "Program AE"},
		{csExposureMode, 4, "ExposureMode", "Manual"},
		{csImageStab, 1, "ImageStabilization", "On"},
		// Unknown values
		{csQuality, 255, "Quality", ""},
		{0, 0, "", ""},      // Index 0 - not decoded
		{1, 0, "", ""},      // Index 1 - not decoded
		{100, 0, "", ""},    // Out of range index
	}

	for _, tt := range tests {
		name, value := decodeCameraSettingsValue(tt.index, tt.value)
		if name != tt.wantName {
			t.Errorf("decodeCameraSettingsValue(%d, %d) name = %q, want %q", tt.index, tt.value, name, tt.wantName)
		}
		if value != tt.wantValue {
			t.Errorf("decodeCameraSettingsValue(%d, %d) value = %q, want %q", tt.index, tt.value, value, tt.wantValue)
		}
	}
}

func TestDecodeShotInfoValue(t *testing.T) {
	tests := []struct {
		index     int
		value     uint16
		wantName  string
		wantValue string
	}{
		{siWhiteBalance, 0, "WhiteBalance", "Auto"},
		{siWhiteBalance, 1, "WhiteBalance", "Daylight"},
		{siWhiteBalance, 5, "WhiteBalance", "Flash"},
		{siWhiteBalance, 255, "WhiteBalance", ""}, // Unknown value
		{0, 0, "", ""},                             // Index 0 - not decoded
		{100, 0, "", ""},                           // Out of range
	}

	for _, tt := range tests {
		name, value := decodeShotInfoValue(tt.index, tt.value)
		if name != tt.wantName {
			t.Errorf("decodeShotInfoValue(%d, %d) name = %q, want %q", tt.index, tt.value, name, tt.wantName)
		}
		if value != tt.wantValue {
			t.Errorf("decodeShotInfoValue(%d, %d) value = %q, want %q", tt.index, tt.value, value, tt.wantValue)
		}
	}
}

func TestDecodeModelID(t *testing.T) {
	tests := []struct {
		modelID uint32
		want    string
	}{
		{0x80000188, "EOS-1Ds Mark II"},
		{0x80000213, "EOS 5D"},
		{0x80000218, "EOS 5D Mark II"},
		{0x80000285, "EOS 5D Mark III"},
		{0x80000349, "EOS 5D Mark IV"},
		{0x80000424, "EOS R"},
		{0x80000421, "EOS R5"},
		{0x80000453, "EOS R6"},
		{0x00000000, ""}, // Unknown
		{0xFFFFFFFF, ""}, // Unknown
	}

	for _, tt := range tests {
		if got := decodeModelID(tt.modelID); got != tt.want {
			t.Errorf("decodeModelID(0x%08X) = %q, want %q", tt.modelID, got, tt.want)
		}
	}
}

func TestHandler_DecodeCameraSettings(t *testing.T) {
	h := New()

	// Test with a sample CameraSettings array
	settings := []uint16{
		0,  // 0: unused
		0,  // 1: MacroMode
		0,  // 2: SelfTimer
		4,  // 3: Quality = RAW
		0,  // 4: FlashMode = Off
		0,  // 5: DriveMode = Single
		0,  // 6: unused
		0,  // 7: FocusMode = One-shot AF
		0,  // 8: unused
		0,  // 9: RecordMode
		0,  // 10: ImageSize = Large
		0,  // 11: EasyMode = Full auto
		0,  // 12: DigitalZoom
		0,  // 13: Contrast
		0,  // 14: Saturation
		0,  // 15: Sharpness
		0,  // 16: ISOSpeed
		3,  // 17: MeteringMode = Evaluative
		0,  // 18: FocusType = Manual
		0,  // 19: AFPoint
		1,  // 20: ExposureMode = Program AE
	}

	tags := h.decodeCameraSettings(settings)

	// Build a map for easier checking
	tagMap := make(map[string]string)
	for _, tag := range tags {
		if v, ok := tag.Value.(string); ok {
			tagMap[tag.Name] = v
		}
	}

	expectedTags := map[string]string{
		"Quality":      "RAW",
		"FlashMode":    "Off",
		"DriveMode":    "Single",
		"FocusMode":    "One-shot AF",
		"ImageSize":    "Large",
		"EasyMode":     "Full auto",
		"MeteringMode": "Evaluative",
		"FocusType":    "Manual",
		"ExposureMode": "Program AE",
	}

	for name, expectedValue := range expectedTags {
		if got, ok := tagMap[name]; !ok {
			t.Errorf("Missing expected tag %q", name)
		} else if got != expectedValue {
			t.Errorf("Tag %q = %q, want %q", name, got, expectedValue)
		}
	}
}

func TestHandler_DecodeModelIDTag(t *testing.T) {
	h := New()

	// Test with known ModelID
	tag := h.decodeModelIDTag(uint32(0x80000188))
	if tag == nil {
		t.Fatal("decodeModelIDTag returned nil for known ModelID")
	}
	if tag.Name != "ModelName" {
		t.Errorf("Tag.Name = %q, want %q", tag.Name, "ModelName")
	}
	if tag.Value != "EOS-1Ds Mark II" {
		t.Errorf("Tag.Value = %v, want %q", tag.Value, "EOS-1Ds Mark II")
	}

	// Test with unknown ModelID
	tag = h.decodeModelIDTag(uint32(0x00000001))
	if tag != nil {
		t.Errorf("decodeModelIDTag returned non-nil for unknown ModelID: %v", tag)
	}

	// Test with invalid type
	tag = h.decodeModelIDTag("invalid")
	if tag != nil {
		t.Errorf("decodeModelIDTag returned non-nil for invalid type: %v", tag)
	}
}
