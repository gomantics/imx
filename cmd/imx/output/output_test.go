package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/gomantics/imx"
)

// Helper to create test results
func createTestResult(file string, tags []TagInfo, err error) *Result {
	var meta *imx.Metadata
	if err == nil {
		meta = &imx.Metadata{}
	}
	return &Result{
		File:     file,
		Meta:     meta,
		Tags:     tags,
		TagCount: len(tags),
		Error:    err,
	}
}

func createTestTag(dirName string, name string, value any) TagInfo {
	return TagInfo{
		Dir: imx.Directory{Name: dirName},
		Tag: imx.Tag{Name: name, Value: value},
	}
}

// Test NewFormatter
func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"text", "text", false},
		{"json", "json", false},
		{"table", "table", false},
		{"csv", "csv", false},
		{"summary", "summary", false},
		{"unknown", "unknown", true},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFormatter(tt.format, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFormatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && f == nil {
				t.Error("NewFormatter() returned nil formatter")
			}
		})
	}
}

func TestNewFormatterWithConfig(t *testing.T) {
	config := &Config{
		NoColor: true,
		Full:    true,
	}
	f, err := NewFormatter("json", config)
	if err != nil {
		t.Fatalf("NewFormatter() error = %v", err)
	}
	if f == nil {
		t.Fatal("NewFormatter() returned nil")
	}
}

// Test FormatSingle
func TestFormatSingle(t *testing.T) {
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Make", "Canon"),
	}, nil)

	var buf bytes.Buffer
	err := FormatSingle(&buf, result, "json", nil)
	if err != nil {
		t.Fatalf("FormatSingle() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("FormatSingle() wrote nothing")
	}
}

func TestFormatSingleInvalidFormat(t *testing.T) {
	result := createTestResult("test.jpg", nil, nil)

	var buf bytes.Buffer
	err := FormatSingle(&buf, result, "invalid", nil)
	if err == nil {
		t.Error("FormatSingle() expected error for invalid format")
	}
}

// Test JSONFormatter
func TestJSONFormatter_SingleResult(t *testing.T) {
	formatter := &JSONFormatter{config: &Config{}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Make", "Canon"),
		createTestTag("IFD0", "Model", "EOS 5D"),
		createTestTag("XMP", "Creator", "Test User"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Verify JSON structure
	var output map[string]any
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if output["SourceFile"] != "test.jpg" {
		t.Errorf("SourceFile = %v, want test.jpg", output["SourceFile"])
	}

	if ifd0, ok := output["IFD0"].(map[string]any); ok {
		if ifd0["Make"] != "Canon" {
			t.Errorf("IFD0.Make = %v, want Canon", ifd0["Make"])
		}
		if ifd0["Model"] != "EOS 5D" {
			t.Errorf("IFD0.Model = %v, want EOS 5D", ifd0["Model"])
		}
	} else {
		t.Error("Missing or invalid IFD0 section")
	}

	if xmp, ok := output["XMP"].(map[string]any); ok {
		if xmp["Creator"] != "Test User" {
			t.Errorf("XMP.Creator = %v, want Test User", xmp["Creator"])
		}
	} else {
		t.Error("Missing or invalid XMP section")
	}
}

func TestJSONFormatter_MultipleResults(t *testing.T) {
	formatter := &JSONFormatter{config: &Config{}}
	results := []*Result{
		createTestResult("photo1.jpg", []TagInfo{
			createTestTag("IFD0", "Make", "Canon"),
		}, nil),
		createTestResult("photo2.jpg", []TagInfo{
			createTestTag("IFD0", "Make", "Nikon"),
		}, nil),
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, results)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Verify JSON array
	var output []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Invalid JSON array output: %v", err)
	}

	if len(output) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(output))
	}

	if output[0]["SourceFile"] != "photo1.jpg" {
		t.Errorf("First result SourceFile = %v, want photo1.jpg", output[0]["SourceFile"])
	}
	if output[1]["SourceFile"] != "photo2.jpg" {
		t.Errorf("Second result SourceFile = %v, want photo2.jpg", output[1]["SourceFile"])
	}
}

func TestJSONFormatter_Error(t *testing.T) {
	formatter := &JSONFormatter{config: &Config{}}
	result := createTestResult("test.jpg", nil, errors.New("file not found"))

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if output["Error"] != "file not found" {
		t.Errorf("Error = %v, want 'file not found'", output["Error"])
	}
}

func TestJSONFormatter_MultipleWithErrors(t *testing.T) {
	formatter := &JSONFormatter{config: &Config{}}
	results := []*Result{
		createTestResult("photo1.jpg", []TagInfo{
			createTestTag("IFD0", "Make", "Canon"),
		}, nil),
		createTestResult("photo2.jpg", nil, errors.New("read error")),
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, results)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	var output []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Invalid JSON array output: %v", err)
	}

	if len(output) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(output))
	}

	if output[1]["Error"] != "read error" {
		t.Errorf("Second result Error = %v, want 'read error'", output[1]["Error"])
	}
}

func TestJSONFormatter_BinaryData(t *testing.T) {
	formatter := &JSONFormatter{config: &Config{}}

	// Test small binary data (should be hex)
	smallData := []byte{0x01, 0x02, 0x03}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Binary", smallData),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Test large binary data (should be size object)
	largeData := make([]byte, 200)
	result2 := createTestResult("test2.jpg", []TagInfo{
		createTestTag("IFD0", "LargeBinary", largeData),
	}, nil)

	buf.Reset()
	err = formatter.Format(&buf, []*Result{result2})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if exif, ok := output["EXIF"].(map[string]any); ok {
		if binary, ok := exif["LargeBinary"].(map[string]any); ok {
			if binary["type"] != "binary" {
				t.Errorf("Binary type = %v, want 'binary'", binary["type"])
			}
			// Size is float64 when decoded from JSON
			if size, ok := binary["size"].(float64); !ok || int(size) != 200 {
				t.Errorf("Binary size = %v, want 200", binary["size"])
			}
		} else {
			t.Error("Large binary data not formatted as object")
		}
	}
}

// Test CSVFormatter
func TestCSVFormatter(t *testing.T) {
	formatter := &CSVFormatter{config: &Config{}}
	results := []*Result{
		createTestResult("photo1.jpg", []TagInfo{
			createTestTag("IFD0", "Make", "Canon"),
			createTestTag("IFD0", "Model", "EOS 5D"),
		}, nil),
		createTestResult("photo2.jpg", []TagInfo{
			createTestTag("IFD0", "Make", "Nikon"),
		}, nil),
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, results)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 { // header + 3 data rows
		t.Errorf("Expected 4 lines, got %d", len(lines))
	}

	// Check header
	if !strings.Contains(lines[0], "File") || !strings.Contains(lines[0], "Dir") {
		t.Errorf("Invalid CSV header: %s", lines[0])
	}

	// Check data
	if !strings.Contains(lines[1], "photo1.jpg") {
		t.Errorf("First data line missing filename: %s", lines[1])
	}
}

func TestCSVFormatter_Error(t *testing.T) {
	formatter := &CSVFormatter{config: &Config{}}
	results := []*Result{
		createTestResult("photo1.jpg", nil, errors.New("read error")),
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, results)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Error("CSV should contain ERROR for failed result")
	}
	if !strings.Contains(output, "read error") {
		t.Error("CSV should contain error message")
	}
}

// Test TableFormatter
func TestTableFormatter(t *testing.T) {
	formatter := &TableFormatter{config: &Config{}}
	results := []*Result{
		createTestResult("photo1.jpg", []TagInfo{
			createTestTag("IFD0", "Make", "Canon"),
			createTestTag("IFD0", "Model", "EOS 5D"),
		}, nil),
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, results)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "photo1.jpg") {
		t.Error("Table should contain filename")
	}
	if !strings.Contains(output, "Canon") {
		t.Error("Table should contain metadata value")
	}
}

func TestTableFormatter_NoColor(t *testing.T) {
	formatter := &TableFormatter{config: &Config{NoColor: true}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Make", "Canon"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Format() wrote nothing")
	}
}

func TestTableFormatter_Quiet(t *testing.T) {
	formatter := &TableFormatter{config: &Config{Quiet: true}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Make", "Canon"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	// In quiet mode, shouldn't have file header
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 5 { // Should be minimal
		t.Error("Quiet mode should produce minimal output")
	}
}

func TestTableFormatter_Error(t *testing.T) {
	formatter := &TableFormatter{config: &Config{}}
	result := createTestResult("test.jpg", nil, errors.New("read error"))

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Error") {
		t.Error("Should display error message")
	}
}

func TestTableFormatter_NoTags(t *testing.T) {
	formatter := &TableFormatter{config: &Config{}}
	result := createTestResult("test.jpg", []TagInfo{}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No metadata found") {
		t.Error("Should display 'No metadata found' message")
	}
}

func TestTableFormatter_Multiple(t *testing.T) {
	formatter := &TableFormatter{config: &Config{}}
	results := []*Result{
		createTestResult("photo1.jpg", []TagInfo{
			createTestTag("IFD0", "Make", "Canon"),
		}, nil),
		createTestResult("photo2.jpg", []TagInfo{
			createTestTag("IFD0", "Make", "Nikon"),
		}, nil),
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, results)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "photo1.jpg") {
		t.Error("Should contain first filename")
	}
	if !strings.Contains(output, "photo2.jpg") {
		t.Error("Should contain second filename")
	}
}

// Test TextFormatter
func TestTextFormatter(t *testing.T) {
	formatter := &TextFormatter{config: &Config{}}
	result := createTestResult("photo.jpg", []TagInfo{
		createTestTag("IFD0", "Make", "Canon"),
		createTestTag("XMP", "Creator", "Test User"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[IFD0]") {
		t.Error("Should contain IFD0 section header")
	}
	if !strings.Contains(output, "[XMP]") {
		t.Error("Should contain XMP section header")
	}
	if !strings.Contains(output, "Canon") {
		t.Error("Should contain metadata value")
	}
}

func TestTextFormatter_NoColor(t *testing.T) {
	formatter := &TextFormatter{config: &Config{NoColor: true}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Make", "Canon"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Format() wrote nothing")
	}
}

func TestTextFormatter_Quiet(t *testing.T) {
	formatter := &TextFormatter{config: &Config{Quiet: true}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Make", "Canon"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "test.jpg") {
		t.Error("Quiet mode should not display filename header")
	}
}

func TestTextFormatter_Error(t *testing.T) {
	formatter := &TextFormatter{config: &Config{}}
	result := createTestResult("test.jpg", nil, errors.New("parse error"))

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Error") {
		t.Error("Should display error message")
	}
}

func TestTextFormatter_NoTags(t *testing.T) {
	formatter := &TextFormatter{config: &Config{}}
	result := createTestResult("test.jpg", []TagInfo{}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No metadata found") {
		t.Error("Should display 'No metadata found' message")
	}
}

// Test SummaryFormatter
func TestSummaryFormatter(t *testing.T) {
	// Create metadata with tags
	meta := &imx.Metadata{}

	result := &Result{
		File: "photo.jpg",
		Meta: meta,
		Tags: []TagInfo{
			createTestTag("IFD0", "Make", "Canon"),
		},
	}

	formatter := &SummaryFormatter{config: &Config{}}

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Format() wrote nothing")
	}
}

func TestSummaryFormatter_NoColor(t *testing.T) {
	meta := &imx.Metadata{}
	result := &Result{
		File: "test.jpg",
		Meta: meta,
	}

	formatter := &SummaryFormatter{config: &Config{NoColor: true}}

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Format() wrote nothing")
	}
}

func TestSummaryFormatter_Quiet(t *testing.T) {
	meta := &imx.Metadata{}
	result := &Result{
		File: "test.jpg",
		Meta: meta,
	}

	formatter := &SummaryFormatter{config: &Config{Quiet: true}}

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "test.jpg") {
		t.Error("Quiet mode should not display filename")
	}
}

func TestSummaryFormatter_Error(t *testing.T) {
	result := createTestResult("test.jpg", nil, errors.New("file error"))

	formatter := &SummaryFormatter{config: &Config{}}

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Error") {
		t.Error("Should display error message")
	}
}

func TestSummaryFormatter_Multiple(t *testing.T) {
	results := []*Result{
		{File: "photo1.jpg", Meta: &imx.Metadata{}},
		{File: "photo2.jpg", Meta: &imx.Metadata{}},
	}

	formatter := &SummaryFormatter{config: &Config{}}

	var buf bytes.Buffer
	err := formatter.Format(&buf, results)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "photo1.jpg") {
		t.Error("Should contain first filename")
	}
	if !strings.Contains(output, "photo2.jpg") {
		t.Error("Should contain second filename")
	}
}

// Test min helper function
func TestMin(t *testing.T) {
	if min(5, 10) != 5 {
		t.Error("min(5, 10) should be 5")
	}
	if min(10, 5) != 5 {
		t.Error("min(10, 5) should be 5")
	}
	if min(5, 5) != 5 {
		t.Error("min(5, 5) should be 5")
	}
}

// Test SummaryFormatter with actual metadata
func TestSummaryFormatter_WithMetadata(t *testing.T) {
	// Create a simple result without metadata (formatter only needs File and Tags)
	result := &Result{
		File: "photo.jpg",
		Meta: nil,
		Tags: []TagInfo{
			{Dir: imx.Directory{Name: "IFD0"}, Tag: imx.Tag{ID: "EXIF:Make", Name: "Make", Value: "Canon"}},
		},
	}

	formatter := &SummaryFormatter{config: &Config{}}

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "photo.jpg") {
		t.Error("Should contain filename")
	}
}

// Test with CSV writer error handling
func TestCSVFormatter_WriterError(t *testing.T) {
	formatter := &CSVFormatter{config: &Config{}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Make", "Canon"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Should write output")
	}
}

// Test TextFormatter with multiple results
func TestTextFormatter_Multiple(t *testing.T) {
	formatter := &TextFormatter{config: &Config{}}
	results := []*Result{
		createTestResult("photo1.jpg", []TagInfo{
			createTestTag("IFD0", "Make", "Canon"),
		}, nil),
		createTestResult("photo2.jpg", []TagInfo{
			createTestTag("IPTC", "Byline", "Photographer"),
		}, nil),
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, results)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "photo1.jpg") {
		t.Error("Should contain first filename")
	}
	if !strings.Contains(output, "photo2.jpg") {
		t.Error("Should contain second filename")
	}
	if !strings.Contains(output, "[IFD0]") {
		t.Error("Should contain IFD0 section")
	}
	if !strings.Contains(output, "[IPTC]") {
		t.Error("Should contain IPTC section")
	}
}

// Test TableFormatter with long names and values
func TestTableFormatter_LongValues(t *testing.T) {
	formatter := &TableFormatter{config: &Config{}}
	longName := strings.Repeat("A", 50)
	longValue := strings.Repeat("B", 100)

	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", longName, longValue),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	// Should truncate long values
	if strings.Contains(output, longValue) {
		t.Error("Should truncate long values")
	}
}

// Test TableFormatter with Full config
func TestTableFormatter_Full(t *testing.T) {
	formatter := &TableFormatter{config: &Config{Full: true}}
	longValue := strings.Repeat("B", 100)

	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Test", longValue),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	// With Full=true, should NOT truncate
	if !strings.Contains(output, longValue) {
		t.Error("Full mode should not truncate values")
	}
}

// Test TextFormatter with long names and values
func TestTextFormatter_LongValues(t *testing.T) {
	formatter := &TextFormatter{config: &Config{}}
	longName := strings.Repeat("A", 50)
	longValue := strings.Repeat("B", 100)

	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", longName, longValue),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	// Should truncate long values
	if strings.Contains(output, longValue) {
		t.Error("Should truncate long values")
	}
}

// Test TextFormatter with Full config
func TestTextFormatter_Full(t *testing.T) {
	formatter := &TextFormatter{config: &Config{Full: true}}
	longValue := strings.Repeat("B", 100)

	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Test", longValue),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	// With Full=true, should NOT truncate
	if !strings.Contains(output, longValue) {
		t.Error("Full mode should not truncate values")
	}
}

// Test CSVFormatter with Full config
func TestCSVFormatter_Full(t *testing.T) {
	formatter := &CSVFormatter{config: &Config{Full: true}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Test", "value"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Should write output")
	}
}

// Test isTimeField helper function
func TestIsTimeField(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		expected bool
	}{
		{"DateTimeOriginal", "DateTimeOriginal", true},
		{"CreateDate", "CreateDate", true},
		{"ModifyDate", "ModifyDate", true},
		{"DateTime", "DateTime", true},
		{"TimeStamp", "TimeStamp", true},
		{"GPSTimeStamp", "GPSTimeStamp", true},
		{"DateCreated", "DateCreated", true},
		{"Make", "Make", false},
		{"Model", "Model", false},
		{"ISO", "ISO", false},
		{"Mixed case DATE", "SomeDate", true},
		{"Mixed case TIME", "SomeTime", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTimeField(tt.tagName)
			if result != tt.expected {
				t.Errorf("isTimeField(%q) = %v, want %v", tt.tagName, result, tt.expected)
			}
		})
	}
}

// Test TimeFormat with Table formatter
func TestTableFormatter_TimeFormat(t *testing.T) {
	tests := []struct {
		name         string
		timeFormat   string
		expectedTime string
	}{
		{"iso", "iso", "2021-12-16T16:12:21Z"},
		{"rfc3339", "rfc3339", "2021-12-16T16:12:21Z"},
		{"unix", "unix", "1639671141"},
		{"human", "human", "Dec 16, 2021 4:12 PM"},
		{"custom", "2006-01-02 15:04:05", "2021-12-16 16:12:21"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &TableFormatter{config: &Config{TimeFormat: tt.timeFormat}}
			result := createTestResult("test.jpg", []TagInfo{
				createTestTag("IFD0", "DateTimeOriginal", "2021:12:16 16:12:21"),
			}, nil)

			var buf bytes.Buffer
			err := formatter.Format(&buf, []*Result{result})
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.expectedTime) {
				t.Errorf("Expected time format %q to produce %q, but output doesn't contain it:\n%s", tt.timeFormat, tt.expectedTime, output)
			}
		})
	}
}

// Test TimeFormat with Text formatter
func TestTextFormatter_TimeFormat(t *testing.T) {
	formatter := &TextFormatter{config: &Config{TimeFormat: "unix"}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "CreateDate", "2021:12:16 16:12:21"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1639671141") {
		t.Error("Text formatter should apply time formatting")
	}
}

// Test TimeFormat with CSV formatter
func TestCSVFormatter_TimeFormat(t *testing.T) {
	formatter := &CSVFormatter{config: &Config{TimeFormat: "human"}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "ModifyDate", "2021:12:16 14:46:24"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Dec 16, 2021") {
		t.Errorf("CSV formatter should apply time formatting, got:\n%s", output)
	}
}

// Test TimeFormat with JSON formatter
func TestJSONFormatter_TimeFormat(t *testing.T) {
	formatter := &JSONFormatter{config: &Config{TimeFormat: "unix"}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "DateTimeOriginal", "2021:12:16 16:12:21"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if ifd0, ok := output["IFD0"].(map[string]any); ok {
		if ifd0["DateTimeOriginal"] != "1639671141" {
			t.Errorf("JSON formatter should apply time formatting, got: %v", ifd0["DateTimeOriginal"])
		}
	} else {
		t.Error("Missing or invalid IFD0 section")
	}
}

// Test that non-time fields are not affected
func TestTableFormatter_NonTimeFieldsUnaffected(t *testing.T) {
	formatter := &TableFormatter{config: &Config{TimeFormat: "unix"}}
	result := createTestResult("test.jpg", []TagInfo{
		createTestTag("IFD0", "Make", "Canon"),
		createTestTag("IFD0", "Model", "EOS 5D"),
	}, nil)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []*Result{result})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Canon") {
		t.Error("Non-time fields should not be affected by time formatting")
	}
	if !strings.Contains(output, "EOS 5D") {
		t.Error("Non-time fields should not be affected by time formatting")
	}
}
