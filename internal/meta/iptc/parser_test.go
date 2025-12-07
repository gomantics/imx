package iptc

import (
	"encoding/binary"
	"testing"
)

// buildPhotoshopIRB creates a Photoshop Image Resource Block
func buildPhotoshopIRB(resourceID uint16, data []byte) []byte {
	result := make([]byte, 0, 12+len(data))
	result = append(result, signature8BIM...)                      // 8BIM
	result = append(result, byte(resourceID>>8), byte(resourceID)) // Resource ID
	result = append(result, 0)                                     // Pascal string length (0 = no name)
	result = append(result, 0)                                     // Padding to even
	// Data size (4 bytes)
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(data)))
	result = append(result, size...)
	result = append(result, data...)
	// Pad to even if needed
	if len(data)%2 != 0 {
		result = append(result, 0)
	}
	return result
}

// buildIPTCDataset creates an IPTC-IIM dataset
func buildIPTCDataset(record Record, datasetID uint8, value []byte) []byte {
	result := make([]byte, 0, 5+len(value))
	result = append(result, iptcTagMarker) // Tag marker
	result = append(result, byte(record))  // Record
	result = append(result, datasetID)     // Dataset ID
	// Size (2 bytes)
	result = append(result, byte(len(value)>>8), byte(len(value)))
	result = append(result, value...)
	return result
}

func TestParsePhotoshopIRB(t *testing.T) {
	// Create IPTC data
	iptcData := buildIPTCDataset(RecordApplication, 5, []byte("Test Title"))

	// Wrap in Photoshop IRB
	irbData := buildPhotoshopIRB(ResourceIPTC, iptcData)

	result, err := parsePhotoshopIRB(irbData)
	if err != nil {
		t.Fatalf("parsePhotoshopIRB() error = %v", err)
	}

	if len(result) != len(iptcData) {
		t.Errorf("parsePhotoshopIRB() returned %d bytes, want %d", len(result), len(iptcData))
	}
}

func TestParsePhotoshopIRB_NonIPTC(t *testing.T) {
	// Create a non-IPTC resource
	data := buildPhotoshopIRB(ResourceXMP, []byte("<xmp>test</xmp>"))

	result, err := parsePhotoshopIRB(data)
	if err != nil {
		t.Fatalf("parsePhotoshopIRB() error = %v", err)
	}

	if result != nil {
		t.Error("parsePhotoshopIRB() should return nil for non-IPTC resource")
	}
}

func TestParsePhotoshopIRB_TooShort(t *testing.T) {
	_, err := parsePhotoshopIRB([]byte{1, 2, 3})
	if err == nil {
		t.Error("parsePhotoshopIRB() should error on short data")
	}
}

func TestParsePhotoshopIRB_InvalidSignature(t *testing.T) {
	data := []byte("XXXX\x04\x04\x00\x00\x00\x00\x00\x05hello")
	result, _ := parsePhotoshopIRB(data)
	if result != nil {
		t.Error("parsePhotoshopIRB() should return nil for invalid signature")
	}
}

func TestParsePhotoshopIRB_FindNext8BIM(t *testing.T) {
	// Some garbage followed by valid 8BIM
	garbage := []byte{0, 0, 0, 0}
	iptcData := buildIPTCDataset(RecordApplication, 5, []byte("Test"))
	irb := buildPhotoshopIRB(ResourceIPTC, iptcData)

	data := append(garbage, irb...)

	result, err := parsePhotoshopIRB(data)
	if err != nil {
		t.Fatalf("parsePhotoshopIRB() error = %v", err)
	}
	if len(result) == 0 {
		t.Error("parsePhotoshopIRB() should find IPTC after garbage")
	}
}

func TestParsePhotoshopIRB_MultipleResources(t *testing.T) {
	// First a non-IPTC resource
	xmpIRB := buildPhotoshopIRB(ResourceXMP, []byte("<xmp/>"))
	// Then IPTC
	iptcData := buildIPTCDataset(RecordApplication, 5, []byte("Title"))
	iptcIRB := buildPhotoshopIRB(ResourceIPTC, iptcData)

	data := append(xmpIRB, iptcIRB...)

	result, err := parsePhotoshopIRB(data)
	if err != nil {
		t.Fatalf("parsePhotoshopIRB() error = %v", err)
	}
	if len(result) == 0 {
		t.Error("parsePhotoshopIRB() should find IPTC in multiple resources")
	}
}

func TestParsePhotoshopIRB_OddNameLength(t *testing.T) {
	// Build manually with odd name length
	data := make([]byte, 0, 30)
	data = append(data, signature8BIM...)
	data = append(data, 0x04, 0x04)    // IPTC resource ID
	data = append(data, 3)             // Name length = 3 (odd)
	data = append(data, 'a', 'b', 'c') // Name
	// Padding not needed for odd length (3+1 = 4, already even)
	// Data size
	data = append(data, 0, 0, 0, 5)
	data = append(data, 'h', 'e', 'l', 'l', 'o')
	data = append(data, 0) // Pad to even

	result, err := parsePhotoshopIRB(data)
	if err != nil {
		t.Fatalf("parsePhotoshopIRB() error = %v", err)
	}
	if len(result) != 5 {
		t.Errorf("parsePhotoshopIRB() returned %d bytes, want 5", len(result))
	}
}

func TestParseIPTCIIM(t *testing.T) {
	data := buildIPTCDataset(RecordApplication, 5, []byte("Test Title"))
	data = append(data, buildIPTCDataset(RecordApplication, 80, []byte("John Doe"))...)

	datasets, err := parseIPTCIIM(data)
	if err != nil {
		t.Fatalf("parseIPTCIIM() error = %v", err)
	}

	if len(datasets) != 2 {
		t.Fatalf("parseIPTCIIM() returned %d datasets, want 2", len(datasets))
	}

	if datasets[0].Name != "ObjectName" {
		t.Errorf("datasets[0].Name = %q, want %q", datasets[0].Name, "ObjectName")
	}
	if datasets[0].Value != "Test Title" {
		t.Errorf("datasets[0].Value = %v, want %q", datasets[0].Value, "Test Title")
	}

	if datasets[1].Name != "Byline" {
		t.Errorf("datasets[1].Name = %q, want %q", datasets[1].Name, "Byline")
	}
}

func TestParseIPTCIIM_TooShort(t *testing.T) {
	datasets, _ := parseIPTCIIM([]byte{1, 2})
	if datasets != nil {
		t.Error("parseIPTCIIM() should return nil for short data")
	}
}

func TestParseIPTCIIM_SkipNonMarker(t *testing.T) {
	// Some garbage followed by valid dataset
	data := []byte{0, 0, 0}
	data = append(data, buildIPTCDataset(RecordApplication, 5, []byte("Title"))...)

	datasets, _ := parseIPTCIIM(data)
	if len(datasets) != 1 {
		t.Errorf("parseIPTCIIM() should skip non-marker bytes, got %d datasets", len(datasets))
	}
}

func TestParseIPTCIIM_UnknownDataset(t *testing.T) {
	data := buildIPTCDataset(RecordApplication, 255, []byte("Unknown"))

	datasets, _ := parseIPTCIIM(data)
	if len(datasets) != 1 {
		t.Fatalf("parseIPTCIIM() returned %d datasets, want 1", len(datasets))
	}

	if datasets[0].Name != "Dataset2:255" {
		t.Errorf("datasets[0].Name = %q, want %q", datasets[0].Name, "Dataset2:255")
	}
}

func TestParseIPTCIIM_ExtendedSize(t *testing.T) {
	// Build dataset with extended size indicator
	data := []byte{
		iptcTagMarker,
		byte(RecordApplication),
		5,          // ObjectName
		0x80, 0x04, // Extended size flag + 4 bytes follow
		0x00, 0x00, 0x00, 0x05, // Size = 5
		'H', 'e', 'l', 'l', 'o',
	}

	datasets, _ := parseIPTCIIM(data)
	if len(datasets) != 1 {
		t.Fatalf("parseIPTCIIM() returned %d datasets, want 1", len(datasets))
	}
	if datasets[0].Value != "Hello" {
		t.Errorf("datasets[0].Value = %v, want %q", datasets[0].Value, "Hello")
	}
}

func TestParseDatasetValue_RecordVersion(t *testing.T) {
	data := []byte{0x00, 0x04} // Version 4
	val := parseDatasetValue(RecordApplication, 0, data)
	if val != 4 {
		t.Errorf("parseDatasetValue() = %v, want 4", val)
	}
}

func TestParseDatasetValue_Urgency(t *testing.T) {
	data := []byte{'5'} // Urgency 5
	val := parseDatasetValue(RecordApplication, 10, data)
	if val != 5 {
		t.Errorf("parseDatasetValue() = %v, want 5", val)
	}
}

func TestParseDatasetValue_DateCreated(t *testing.T) {
	data := []byte("20231215")
	val := parseDatasetValue(RecordApplication, 55, data)
	if val != "2023-12-15" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "2023-12-15")
	}
}

func TestParseDatasetValue_TimeCreated(t *testing.T) {
	data := []byte("143052+0530")
	val := parseDatasetValue(RecordApplication, 60, data)
	if val != "14:30:52+05:30" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "14:30:52+05:30")
	}
}

func TestParseDatasetValue_DigitalCreationDate(t *testing.T) {
	data := []byte("20231215")
	val := parseDatasetValue(RecordApplication, 62, data)
	if val != "2023-12-15" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "2023-12-15")
	}
}

func TestParseDatasetValue_DigitalCreationTime(t *testing.T) {
	data := []byte("143052")
	val := parseDatasetValue(RecordApplication, 63, data)
	if val != "14:30:52" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "14:30:52")
	}
}

func TestParseDatasetValue_ReleaseDate(t *testing.T) {
	data := []byte("20231225")
	val := parseDatasetValue(RecordApplication, 30, data)
	if val != "2023-12-25" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "2023-12-25")
	}
}

func TestParseDatasetValue_ExpirationDate(t *testing.T) {
	data := []byte("20241231")
	val := parseDatasetValue(RecordApplication, 37, data)
	if val != "2024-12-31" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "2024-12-31")
	}
}

func TestParseDatasetValue_ReleaseTime(t *testing.T) {
	data := []byte("120000")
	val := parseDatasetValue(RecordApplication, 35, data)
	if val != "12:00:00" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "12:00:00")
	}
}

func TestParseDatasetValue_ExpirationTime(t *testing.T) {
	data := []byte("235959")
	val := parseDatasetValue(RecordApplication, 38, data)
	if val != "23:59:59" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "23:59:59")
	}
}

func TestParseDatasetValue_EnvelopeRecordVersion(t *testing.T) {
	data := []byte{0x00, 0x04}
	val := parseDatasetValue(RecordEnvelope, 0, data)
	if val != 4 {
		t.Errorf("parseDatasetValue() = %v, want 4", val)
	}
}

func TestParseDatasetValue_EnvelopeDateSent(t *testing.T) {
	data := []byte("20231201")
	val := parseDatasetValue(RecordEnvelope, 70, data)
	if val != "2023-12-01" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "2023-12-01")
	}
}

func TestParseDatasetValue_EnvelopeTimeSent(t *testing.T) {
	data := []byte("100000")
	val := parseDatasetValue(RecordEnvelope, 80, data)
	if val != "10:00:00" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "10:00:00")
	}
}

func TestParseDatasetValue_String(t *testing.T) {
	data := []byte("Hello World\x00")
	val := parseDatasetValue(RecordApplication, 5, data)
	if val != "Hello World" {
		t.Errorf("parseDatasetValue() = %v, want %q", val, "Hello World")
	}
}

func TestParseDateString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"20231215", "2023-12-15"},
		{"2023", "2023"},
		{"", ""},
	}

	for _, tt := range tests {
		got := parseDateString([]byte(tt.input))
		if got != tt.want {
			t.Errorf("parseDateString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseTimeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"143052+0530", "14:30:52+05:30"},
		{"143052-0600", "14:30:52-06:00"},
		{"143052", "14:30:52"},
		{"1430", "1430"},
		{"", ""},
	}

	for _, tt := range tests {
		got := parseTimeString([]byte(tt.input))
		if got != tt.want {
			t.Errorf("parseTimeString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseDatasetValue_ShortRecordVersion(t *testing.T) {
	// Too short for uint16
	data := []byte{0x04}
	val := parseDatasetValue(RecordApplication, 0, data)
	// Should fall through to string
	if val != "\x04" {
		t.Errorf("parseDatasetValue() = %v, want string", val)
	}
}

func TestParseDatasetValue_ShortUrgency(t *testing.T) {
	data := []byte{}
	val := parseDatasetValue(RecordApplication, 10, data)
	if val != "" {
		t.Errorf("parseDatasetValue() = %v, want empty string", val)
	}
}

func TestParseDatasetValue_Prefs(t *testing.T) {
	data := []byte("1:0:0:-00001")
	val := parseDatasetValue(RecordApplication, 221, data)
	want := "Tagged:1, ColorClass:0, Rating:0, FrameNum:-00001"
	if val != want {
		t.Errorf("parseDatasetValue() = %v, want %q", val, want)
	}
}

func TestParsePrefs(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1:0:0:-00001", "Tagged:1, ColorClass:0, Rating:0, FrameNum:-00001"},
		{"0:2:5:00123", "Tagged:0, ColorClass:2, Rating:5, FrameNum:00123"},
		{"simple", "simple"}, // Not enough parts
	}

	for _, tt := range tests {
		got := parsePrefs([]byte(tt.input))
		if got != tt.want {
			t.Errorf("parsePrefs(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// Edge case tests for truncated data in parsePhotoshopIRB

func TestParsePhotoshopIRB_TruncatedBeforeDataSize(t *testing.T) {
	// Test hitting line 59 where offset+4 > len(data) after Pascal string
	// We need valid signature + resourceID + Pascal string, but no room for data size
	data := make([]byte, 13)
	copy(data[0:4], signature8BIM)
	data[4], data[5] = 0x04, 0x04 // Resource ID (offset = 6)
	data[6] = 5                   // Pascal string len = 5
	// namePadded = 5, (5+1)%2 = 0, condition false, namePadded stays 5
	// offset = 7 + 5 = 12 (after len byte and name)
	// Check: 12+4 > 13 = 16 > 13 = true, breaks!
	result, _ := parsePhotoshopIRB(data)
	if result != nil {
		t.Error("parsePhotoshopIRB() should return nil for truncated data")
	}
}

func TestParsePhotoshopIRB_TruncatedBeforeResourceData(t *testing.T) {
	// Test hitting line 66 where offset+dataSize > len(data)
	// We need valid header but dataSize claims more than available
	data := make([]byte, 16)
	copy(data[0:4], signature8BIM)
	data[4], data[5] = 0x04, 0x04 // Resource ID
	data[6] = 0                   // Pascal string len = 0
	// namePadded = 0, (0+1)%2 = 1 != 0, so namePadded = 1
	// offset after len byte = 7, offset after namePadded = 8
	data[7] = 0                                         // Padding byte
	data[8], data[9], data[10], data[11] = 0, 0, 0, 100 // Data size = 100
	// offset = 12, check: 12 + 100 > 16 = true, breaks!

	result, _ := parsePhotoshopIRB(data)
	if result != nil {
		t.Error("parsePhotoshopIRB() should return nil when dataSize exceeds buffer")
	}
}

// Edge case tests for truncated data in parseIPTCIIM

func TestParseIPTCIIM_TruncatedValue(t *testing.T) {
	// Valid header but value truncated
	data := []byte{
		iptcTagMarker,
		byte(RecordApplication),
		5,     // ObjectName
		0, 20, // Size = 20
		'H', 'e', 'l', 'l', 'o', // Only 5 bytes
	}
	datasets, _ := parseIPTCIIM(data)
	if len(datasets) != 0 {
		t.Errorf("parseIPTCIIM() should break on truncated value, got %d", len(datasets))
	}
}

func TestParseIPTCIIM_ExtendedSizeInvalid(t *testing.T) {
	// Extended size with extLen > 4 (invalid)
	data := []byte{
		iptcTagMarker,
		byte(RecordApplication),
		5,          // ObjectName
		0x80, 0x05, // Extended size flag + 5 bytes (invalid, max is 4)
		0, 0, 0, 0, 0, // 5 bytes of size
	}
	datasets, _ := parseIPTCIIM(data)
	if len(datasets) != 0 {
		t.Errorf("parseIPTCIIM() should break on invalid extended size, got %d", len(datasets))
	}
}

func TestParseIPTCIIM_ExtendedSizeTruncated(t *testing.T) {
	// Extended size but not enough bytes for it
	data := []byte{
		iptcTagMarker,
		byte(RecordApplication),
		5,          // ObjectName
		0x80, 0x04, // Extended size flag + 4 bytes follow
		0, 0, // Only 2 bytes (truncated)
	}
	datasets, _ := parseIPTCIIM(data)
	if len(datasets) != 0 {
		t.Errorf("parseIPTCIIM() should break on truncated extended size, got %d", len(datasets))
	}
}
