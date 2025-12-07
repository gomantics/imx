package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gomantics/imx"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.extractor == nil {
		t.Error("extractor is nil")
	}
}

func TestParseArgs_Help(t *testing.T) {
	app := NewApp()

	tests := []struct {
		args []string
		want bool
	}{
		{[]string{"-h"}, true},
		{[]string{"--help"}, true},
		{[]string{"file.jpg"}, false},
	}

	for _, tt := range tests {
		opts := app.parseArgs(tt.args)
		if opts.Help != tt.want {
			t.Errorf("parseArgs(%v).Help = %v, want %v", tt.args, opts.Help, tt.want)
		}
	}
}

func TestParseArgs_Version(t *testing.T) {
	app := NewApp()

	tests := []struct {
		args []string
		want bool
	}{
		{[]string{"-V"}, true},
		{[]string{"--version"}, true},
		{[]string{"file.jpg"}, false},
	}

	for _, tt := range tests {
		opts := app.parseArgs(tt.args)
		if opts.Version != tt.want {
			t.Errorf("parseArgs(%v).Version = %v, want %v", tt.args, opts.Version, tt.want)
		}
	}
}

func TestParseArgs_Format(t *testing.T) {
	app := NewApp()

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"-j"}, "json"},
		{[]string{"--json"}, "json"},
		{[]string{"-t"}, "table"},
		{[]string{"--table"}, "table"},
		{[]string{"--csv"}, "csv"},
		{[]string{"-S"}, "summary"},
		{[]string{"--summary"}, "summary"},
		{[]string{"file.jpg"}, "text"},
	}

	for _, tt := range tests {
		opts := app.parseArgs(tt.args)
		if opts.Format != tt.want {
			t.Errorf("parseArgs(%v).Format = %v, want %v", tt.args, opts.Format, tt.want)
		}
	}
}

func TestParseArgs_Filtering(t *testing.T) {
	app := NewApp()

	// Test --spec
	opts := app.parseArgs([]string{"--spec", "exif"})
	if opts.Spec != "exif" {
		t.Errorf("--spec exif: got %q", opts.Spec)
	}

	opts = app.parseArgs([]string{"-s", "iptc"})
	if opts.Spec != "iptc" {
		t.Errorf("-s iptc: got %q", opts.Spec)
	}

	opts = app.parseArgs([]string{"--spec=xmp"})
	if opts.Spec != "xmp" {
		t.Errorf("--spec=xmp: got %q", opts.Spec)
	}

	// Test --get
	opts = app.parseArgs([]string{"--get", "Make"})
	if opts.Tag != "Make" {
		t.Errorf("--get Make: got %q", opts.Tag)
	}

	opts = app.parseArgs([]string{"-g=Model"})
	if opts.Tag != "Model" {
		t.Errorf("-g=Model: got %q", opts.Tag)
	}

	// Test --search
	opts = app.parseArgs([]string{"--search", "date"})
	if opts.Search != "date" {
		t.Errorf("--search date: got %q", opts.Search)
	}

	opts = app.parseArgs([]string{"--search=time"})
	if opts.Search != "time" {
		t.Errorf("--search=time: got %q", opts.Search)
	}

	// Test --pattern
	opts = app.parseArgs([]string{"--pattern", "GPS.*"})
	if opts.Pattern != "GPS.*" {
		t.Errorf("--pattern GPS.*: got %q", opts.Pattern)
	}

	opts = app.parseArgs([]string{"-p=Date.*"})
	if opts.Pattern != "Date.*" {
		t.Errorf("-p=Date.*: got %q", opts.Pattern)
	}
}

func TestParseArgs_Features(t *testing.T) {
	app := NewApp()

	// Test --recursive
	opts := app.parseArgs([]string{"-r"})
	if !opts.Recursive {
		t.Error("-r should set Recursive")
	}

	opts = app.parseArgs([]string{"--recursive"})
	if !opts.Recursive {
		t.Error("--recursive should set Recursive")
	}

	// Test --stdin
	opts = app.parseArgs([]string{"-"})
	if !opts.Stdin {
		t.Error("- should set Stdin")
	}

	opts = app.parseArgs([]string{"--stdin"})
	if !opts.Stdin {
		t.Error("--stdin should set Stdin")
	}

	// Test --timeout
	opts = app.parseArgs([]string{"--timeout", "60"})
	if opts.Timeout != 60 {
		t.Errorf("--timeout 60: got %d", opts.Timeout)
	}

	opts = app.parseArgs([]string{"--timeout=45"})
	if opts.Timeout != 45 {
		t.Errorf("--timeout=45: got %d", opts.Timeout)
	}

	// Test --gps
	opts = app.parseArgs([]string{"--gps", "url"})
	if opts.GPS != "url" {
		t.Errorf("--gps url: got %q", opts.GPS)
	}

	opts = app.parseArgs([]string{"--gps=decimal"})
	if opts.GPS != "decimal" {
		t.Errorf("--gps=decimal: got %q", opts.GPS)
	}

	// Test --stats
	opts = app.parseArgs([]string{"--stats"})
	if !opts.Stats {
		t.Error("--stats should set Stats")
	}

	// Test --export
	opts = app.parseArgs([]string{"--export", "json"})
	if opts.Export != "json" {
		t.Errorf("--export json: got %q", opts.Export)
	}

	opts = app.parseArgs([]string{"-e=xmp"})
	if opts.Export != "xmp" {
		t.Errorf("-e=xmp: got %q", opts.Export)
	}
}

func TestParseArgs_DisplayOptions(t *testing.T) {
	app := NewApp()

	opts := app.parseArgs([]string{"-q"})
	if !opts.Quiet {
		t.Error("-q should set Quiet")
	}

	opts = app.parseArgs([]string{"--quiet"})
	if !opts.Quiet {
		t.Error("--quiet should set Quiet")
	}

	opts = app.parseArgs([]string{"-f"})
	if !opts.Full {
		t.Error("-f should set Full")
	}

	opts = app.parseArgs([]string{"--full"})
	if !opts.Full {
		t.Error("--full should set Full")
	}

	opts = app.parseArgs([]string{"--no-color"})
	if !opts.NoColor {
		t.Error("--no-color should set NoColor")
	}
}

func TestParseArgs_Files(t *testing.T) {
	app := NewApp()

	opts := app.parseArgs([]string{"file1.jpg", "file2.png"})
	if len(opts.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(opts.Files))
	}
	if opts.Files[0] != "file1.jpg" || opts.Files[1] != "file2.png" {
		t.Errorf("files = %v", opts.Files)
	}
}

func TestParseArgs_DefaultTimeout(t *testing.T) {
	app := NewApp()
	opts := app.parseArgs([]string{"file.jpg"})
	if opts.Timeout != 30 {
		t.Errorf("default timeout = %d, want 30", opts.Timeout)
	}
}

func TestRun_Help(t *testing.T) {
	app := NewApp()
	app.opts.NoColor = true

	code := app.Run([]string{"--help"})
	if code != 0 {
		t.Errorf("--help returned %d, want 0", code)
	}
}

func TestRun_Version(t *testing.T) {
	app := NewApp()
	app.opts.NoColor = true

	code := app.Run([]string{"--version"})
	if code != 0 {
		t.Errorf("--version returned %d, want 0", code)
	}
}

func TestRun_NoFiles(t *testing.T) {
	app := NewApp()
	app.opts.NoColor = true

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := app.Run([]string{})

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 1 {
		t.Errorf("no files returned %d, want 1", code)
	}
	if !strings.Contains(buf.String(), "no input files") {
		t.Error("expected 'no input files' error")
	}
}

func TestRun_FileNotFound(t *testing.T) {
	app := NewApp()
	app.opts.NoColor = true

	code := app.Run([]string{"-q", "nonexistent.jpg"})
	if code != 1 {
		t.Errorf("nonexistent file returned %d, want 1", code)
	}
}

func TestRun_RealFile(t *testing.T) {
	// Use a test file from testdata
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	app := NewApp()
	app.opts.NoColor = true

	code := app.Run([]string{"-q", testFile})
	if code != 0 {
		t.Errorf("real file returned %d, want 0", code)
	}
}

func TestRun_JSON(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"--json", testFile})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("json output returned %d, want 0", code)
	}

	// Verify it's valid JSON
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("invalid JSON output: %v", err)
	}

	// Check for expected fields
	if result["SourceFile"] == nil {
		t.Error("missing SourceFile in JSON")
	}
	if result["EXIF"] == nil {
		t.Error("missing EXIF in JSON")
	}
}

func TestRun_Summary(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"-S", testFile})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("summary returned %d, want 0", code)
	}

	output := buf.String()
	if !strings.Contains(output, "Camera:") {
		t.Error("summary missing Camera")
	}
}

func TestRun_SpecFilter(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"--spec=iptc", "--json", testFile})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("spec filter returned %d, want 0", code)
	}

	var result map[string]any
	json.Unmarshal(buf.Bytes(), &result)

	// Should have IPTC but not EXIF
	if result["IPTC"] == nil {
		t.Error("missing IPTC in filtered output")
	}
	if result["EXIF"] != nil {
		t.Error("EXIF should be filtered out")
	}
}

func TestRun_GetTag(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"--get=Make", testFile})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("get tag returned %d, want 0", code)
	}

	output := strings.TrimSpace(buf.String())
	if output != "Google" {
		t.Errorf("Make = %q, want 'Google'", output)
	}
}

func TestRun_GetTagNotFound(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"-q", "--get=NonExistentTag", testFile})

	if code != 1 {
		t.Errorf("get nonexistent tag returned %d, want 1", code)
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"http://example.com/image.jpg", true},
		{"https://example.com/image.jpg", true},
		{"HTTP://EXAMPLE.COM/image.jpg", false}, // Case sensitive
		{"file.jpg", false},
		{"/path/to/file.jpg", false},
		{"ftp://example.com/image.jpg", false},
	}

	for _, tt := range tests {
		got := isURL(tt.path)
		if got != tt.want {
			t.Errorf("isURL(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestGetArgValue(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{"--spec=exif", "exif"},
		{"-s=iptc", "iptc"},
		{"--timeout=30", "30"},
		{"noequals", ""},
	}

	for _, tt := range tests {
		got := getArgValue(tt.arg)
		if got != tt.want {
			t.Errorf("getArgValue(%q) = %q, want %q", tt.arg, got, tt.want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{1.0, "1"},
		{1.5, "1.5"},
		{1.123456, "1.123456"},
		{1.100000, "1.1"},
		{0.0, "0"},
		{123.456789, "123.456789"},
	}

	for _, tt := range tests {
		got := formatFloat(tt.input)
		if got != tt.want {
			t.Errorf("formatFloat(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatValue(t *testing.T) {
	app := NewApp()

	tests := []struct {
		input any
		want  string
	}{
		{nil, ""},
		{"hello", "hello"},
		{" hello ", "hello"},
		{123, "123"},
		{1.5, "1.5"},
		{float32(1.5), "1.5"},
		{[]float64{1.0, 2.5}, "1, 2.5"},
		{[]uint16{1, 2, 3}, "1, 2, 3"},
	}

	for _, tt := range tests {
		got := app.formatValue(tt.input, true)
		if got != tt.want {
			t.Errorf("formatValue(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}

	// Test binary data
	smallBinary := []byte{0x01, 0x02, 0x03}
	got := app.formatValue(smallBinary, false)
	if got != "010203" {
		t.Errorf("formatValue(small binary) = %q, want '010203'", got)
	}

	largeBinary := make([]byte, 100)
	got = app.formatValue(largeBinary, false)
	if !strings.Contains(got, "binary") {
		t.Errorf("formatValue(large binary) = %q, should contain 'binary'", got)
	}
}

func TestColorizer(t *testing.T) {
	enabled := Colorizer{enabled: true}
	disabled := Colorizer{enabled: false}

	if enabled.Reset() != ColorReset {
		t.Error("enabled.Reset() should return color code")
	}
	if disabled.Reset() != "" {
		t.Error("disabled.Reset() should return empty string")
	}

	if enabled.Bold() != ColorBold {
		t.Error("enabled.Bold() should return color code")
	}
	if disabled.Bold() != "" {
		t.Error("disabled.Bold() should return empty string")
	}

	if enabled.Red() != ColorRed {
		t.Error("enabled.Red() should return color code")
	}
	if disabled.Red() != "" {
		t.Error("disabled.Red() should return empty string")
	}

	if enabled.Green() != ColorGreen {
		t.Error("enabled.Green() should return color code")
	}
	if enabled.Yellow() != ColorYellow {
		t.Error("enabled.Yellow() should return color code")
	}
	if enabled.Blue() != ColorBlue {
		t.Error("enabled.Blue() should return color code")
	}
	if enabled.Cyan() != ColorCyan {
		t.Error("enabled.Cyan() should return color code")
	}
	if enabled.Dim() != ColorDim {
		t.Error("enabled.Dim() should return color code")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 1, -1},
		{0, 0, 0},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFilterTag(t *testing.T) {
	app := NewApp()
	app.opts.NoColor = true

	dir := imx.Directory{Spec: imx.SpecEXIF, Name: "IFD0"}
	tag := imx.Tag{Name: "Make", Value: "Google"}

	// No filters - should pass
	if !app.filterTag(dir, tag) {
		t.Error("tag should pass with no filters")
	}

	// Spec filter - match
	app.opts.Spec = "exif"
	if !app.filterTag(dir, tag) {
		t.Error("tag should pass with matching spec filter")
	}

	// Spec filter - no match
	app.opts.Spec = "iptc"
	if app.filterTag(dir, tag) {
		t.Error("tag should fail with non-matching spec filter")
	}
	app.opts.Spec = ""

	// Search filter - match in name
	app.opts.Search = "make"
	if !app.filterTag(dir, tag) {
		t.Error("tag should pass with matching search in name")
	}

	// Search filter - match in value
	app.opts.Search = "google"
	if !app.filterTag(dir, tag) {
		t.Error("tag should pass with matching search in value")
	}

	// Search filter - no match
	app.opts.Search = "canon"
	if app.filterTag(dir, tag) {
		t.Error("tag should fail with non-matching search")
	}
	app.opts.Search = ""

	// Pattern filter - match
	app.opts.Pattern = "^Make$"
	if !app.filterTag(dir, tag) {
		t.Error("tag should pass with matching pattern")
	}

	// Pattern filter - no match
	app.opts.Pattern = "^Model$"
	if app.filterTag(dir, tag) {
		t.Error("tag should fail with non-matching pattern")
	}
	app.opts.Pattern = ""

	// Binary filter
	binaryTag := imx.Tag{Name: "Thumbnail", Value: make([]byte, 200)}
	if app.filterTag(dir, binaryTag) {
		t.Error("large binary should be filtered by default")
	}

	app.opts.Full = true
	if !app.filterTag(dir, binaryTag) {
		t.Error("large binary should pass with --full")
	}
}

func TestExpandFiles(t *testing.T) {
	app := NewApp()
	app.opts.NoColor = true

	// Test URL handling
	app.opts.Files = []string{"http://example.com/image.jpg"}
	files := app.expandFiles()
	if len(files) != 1 || files[0] != "http://example.com/image.jpg" {
		t.Error("URL should be passed through directly")
	}

	// Test file handling
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); err == nil {
		app.opts.Files = []string{testFile}
		files = app.expandFiles()
		if len(files) != 1 {
			t.Errorf("expected 1 file, got %d", len(files))
		}
	}

	// Test nonexistent file (should still be returned, error handled later)
	app.opts.Files = []string{"nonexistent.jpg"}
	files = app.expandFiles()
	if len(files) != 1 {
		t.Error("nonexistent file should be passed through")
	}
}

func TestExpandFiles_Recursive(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "imx-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "test.jpg"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte{}, 0644)
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "nested.jpg"), []byte{}, 0644)

	app := NewApp()
	app.opts.NoColor = true
	app.opts.Recursive = true
	app.opts.Files = []string{tmpDir}

	files := app.expandFiles()

	// Should find 2 jpg files
	jpgCount := 0
	for _, f := range files {
		if strings.HasSuffix(f, ".jpg") {
			jpgCount++
		}
	}

	if jpgCount != 2 {
		t.Errorf("expected 2 jpg files, got %d", jpgCount)
	}
}

func TestFormatGPS(t *testing.T) {
	app := NewApp()

	// Test DMS format (default)
	lat := []float64{37.0, 46.0, 29.88}
	lon := []float64{122.0, 25.0, 9.72}

	gps := app.formatGPS(lat, lon, "N", "W")
	if !strings.Contains(gps, "37°") || !strings.Contains(gps, "122°") {
		t.Errorf("DMS format incorrect: %s", gps)
	}

	// Test decimal format
	app.opts.GPS = "decimal"
	gps = app.formatGPS(lat, lon, "N", "W")
	if !strings.Contains(gps, "37.") && !strings.Contains(gps, "-122.") {
		t.Errorf("decimal format incorrect: %s", gps)
	}

	// Test URL format
	app.opts.GPS = "url"
	gps = app.formatGPS(lat, lon, "N", "W")
	if !strings.Contains(gps, "maps.google.com") {
		t.Errorf("URL format incorrect: %s", gps)
	}
}

func TestToDMS(t *testing.T) {
	app := NewApp()

	// Test North latitude
	dms := app.toDMS(37.7749, false, true)
	if !strings.HasSuffix(dms, "N") {
		t.Errorf("North latitude should end with N: %s", dms)
	}

	// Test South latitude
	dms = app.toDMS(-37.7749, true, true)
	if !strings.HasSuffix(dms, "S") {
		t.Errorf("South latitude should end with S: %s", dms)
	}

	// Test East longitude
	dms = app.toDMS(122.4194, false, false)
	if !strings.HasSuffix(dms, "E") {
		t.Errorf("East longitude should end with E: %s", dms)
	}

	// Test West longitude
	dms = app.toDMS(-122.4194, true, false)
	if !strings.HasSuffix(dms, "W") {
		t.Errorf("West longitude should end with W: %s", dms)
	}
}

func TestToDecimalDegrees(t *testing.T) {
	app := NewApp()

	// DMS array
	coord := []float64{37.0, 46.0, 29.88}
	dec := app.toDecimalDegrees(coord, "N")
	if dec < 37.77 || dec > 37.78 {
		t.Errorf("decimal degrees incorrect: %f", dec)
	}

	// South reference should be negative
	dec = app.toDecimalDegrees(coord, "S")
	if dec > -37.77 || dec < -37.78 {
		t.Errorf("south decimal degrees should be negative: %f", dec)
	}

	// Already decimal
	dec = app.toDecimalDegrees(37.7749, "")
	if dec != 37.7749 {
		t.Errorf("decimal passthrough incorrect: %f", dec)
	}
}

func TestSpecColor(t *testing.T) {
	app := NewApp()
	app.colors = Colorizer{enabled: true}

	if app.specColor(imx.SpecEXIF) != ColorGreen {
		t.Error("EXIF should be green")
	}
	if app.specColor(imx.SpecIPTC) != ColorBlue {
		t.Error("IPTC should be blue")
	}
	if app.specColor(imx.SpecXMP) != ColorCyan {
		t.Error("XMP should be cyan")
	}
	if app.specColor(imx.SpecICC) != ColorYellow {
		t.Error("ICC should be yellow")
	}

	// Disabled colors
	app.colors = Colorizer{enabled: false}
	if app.specColor(imx.SpecEXIF) != "" {
		t.Error("disabled colors should return empty")
	}
}

func TestExitCode(t *testing.T) {
	app := NewApp()

	if app.exitCode(nil) != 0 {
		t.Error("nil error should return 0")
	}

	if app.exitCode(io.EOF) != 1 {
		t.Error("error should return 1")
	}
}

func TestOutputCSV(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"--csv", testFile})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("csv output returned %d, want 0", code)
	}

	output := buf.String()
	// Should have CSV header
	if !strings.Contains(output, "File,Spec,Tag,Value") {
		t.Error("CSV should have header")
	}
	// Should have data rows
	if !strings.Contains(output, "EXIF") {
		t.Error("CSV should contain EXIF data")
	}
}

func TestOutputTable(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"--table", testFile})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("table output returned %d, want 0", code)
	}

	output := buf.String()
	// Should have table header
	if !strings.Contains(output, "SPEC") || !strings.Contains(output, "TAG") {
		t.Error("table should have header")
	}
}

func TestSearchFilter(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"--search=date", "--json", testFile})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("search filter returned %d, want 0", code)
	}

	var result map[string]any
	json.Unmarshal(buf.Bytes(), &result)

	// Check that results contain date-related tags
	if exif, ok := result["EXIF"].(map[string]any); ok {
		hasDate := false
		for key := range exif {
			if strings.Contains(strings.ToLower(key), "date") {
				hasDate = true
				break
			}
		}
		if !hasDate {
			t.Error("search for 'date' should return date-related tags")
		}
	}
}

func TestDirectoryNotRecursive(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "imx-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	app := NewApp()
	app.opts.NoColor = true
	app.opts.Files = []string{tmpDir}

	// Without -r, should print error
	files := app.expandFiles()
	if len(files) != 0 {
		t.Error("directory without -r should return no files")
	}
}

func TestRun_Stats(t *testing.T) {
	testDir := "../../testdata/goldens/jpeg"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("test directory not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"-r", "--stats", "-q", testDir})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("stats output returned %d, want 0", code)
	}

	output := buf.String()
	if !strings.Contains(output, "Statistics:") {
		t.Error("stats output should contain 'Statistics:'")
	}
}

func TestExportSidecar_JSON(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Create temp directory for export
	tmpDir, err := os.MkdirTemp("", "imx-export-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy test file to temp dir
	destFile := filepath.Join(tmpDir, "test.jpg")
	data, _ := os.ReadFile(testFile)
	os.WriteFile(destFile, data, 0644)

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"-q", "--export=json", destFile})

	if code != 0 {
		t.Errorf("export json returned %d, want 0", code)
	}

	// Check sidecar file exists
	sidecarPath := destFile + ".json"
	if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
		t.Error("sidecar JSON file should exist")
	}

	// Verify it's valid JSON
	sidecarData, _ := os.ReadFile(sidecarPath)
	var result map[string]any
	if err := json.Unmarshal(sidecarData, &result); err != nil {
		t.Errorf("sidecar should be valid JSON: %v", err)
	}
}

func TestExportSidecar_XMP(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Create temp directory for export
	tmpDir, err := os.MkdirTemp("", "imx-export-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy test file to temp dir
	destFile := filepath.Join(tmpDir, "test.jpg")
	data, _ := os.ReadFile(testFile)
	os.WriteFile(destFile, data, 0644)

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"-q", "--export=xmp", destFile})

	if code != 0 {
		t.Errorf("export xmp returned %d, want 0", code)
	}

	// Check sidecar file exists
	sidecarPath := destFile + ".xmp"
	if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
		t.Error("sidecar XMP file should exist")
	}

	// Verify it contains XMP structure
	sidecarData, _ := os.ReadFile(sidecarPath)
	if !strings.Contains(string(sidecarData), "x:xmpmeta") {
		t.Error("sidecar should contain XMP structure")
	}
}

func TestExportSidecar_InvalidFormat(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	app := NewApp()
	app.opts.NoColor = true
	// This will process file successfully but fail to export
	code := app.Run([]string{"-q", "--export=invalid", testFile})

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// File processing succeeds, only export fails (which prints error but doesn't fail)
	if code != 0 {
		t.Errorf("export invalid format returned %d, want 0", code)
	}

	// But error should be printed
	if !strings.Contains(buf.String(), "unknown export format") {
		t.Error("should print error about unknown format")
	}
}

func TestFormatJSONValue(t *testing.T) {
	app := NewApp()

	// Regular values should pass through
	if app.formatJSONValue("test") != "test" {
		t.Error("string should pass through")
	}
	if app.formatJSONValue(123) != 123 {
		t.Error("int should pass through")
	}

	// Large binary should be formatted
	largeBinary := make([]byte, 200)
	result := app.formatJSONValue(largeBinary)
	if s, ok := result.(string); !ok || !strings.Contains(s, "binary") {
		t.Error("large binary should be formatted")
	}

	// Small binary should be hex
	smallBinary := []byte{0x01, 0x02, 0x03}
	result = app.formatJSONValue(smallBinary)
	if s, ok := result.(string); !ok || s != "010203" {
		t.Errorf("small binary should be hex: %v", result)
	}
}

func TestPatternFilter(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	code := app.Run([]string{"--pattern=^Make$", "--json", testFile})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("pattern filter returned %d, want 0", code)
	}

	var result map[string]any
	json.Unmarshal(buf.Bytes(), &result)

	// Should have only Make tag
	if exif, ok := result["EXIF"].(map[string]any); ok {
		if len(exif) != 1 || exif["Make"] == nil {
			t.Error("pattern filter should only return Make tag")
		}
	}
}

func TestFormatValue_Arrays(t *testing.T) {
	app := NewApp()

	// Test []any
	anySlice := []any{"a", "b", "c"}
	got := app.formatValue(anySlice, true)
	if got != "a, b, c" {
		t.Errorf("[]any formatting = %q", got)
	}

	// Test nested values
	nested := []any{1, "two", 3.0}
	got = app.formatValue(nested, true)
	if !strings.Contains(got, "1") || !strings.Contains(got, "two") {
		t.Errorf("nested []any formatting = %q", got)
	}
}

func TestGetTagValue(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	app := NewApp()
	meta, err := app.extractor.ExtractFromFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Get existing tag
	value := app.getTagValue(&meta, imx.SpecEXIF, "Make")
	if value != "Google" {
		t.Errorf("getTagValue(Make) = %q, want 'Google'", value)
	}

	// Get non-existing tag
	value = app.getTagValue(&meta, imx.SpecEXIF, "NonExistent")
	if value != "" {
		t.Errorf("getTagValue(NonExistent) = %q, want ''", value)
	}
}

func TestGetRawTagValue(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	app := NewApp()
	meta, err := app.extractor.ExtractFromFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Get existing tag
	value := app.getRawTagValue(&meta, imx.SpecEXIF, "Make")
	if value == nil {
		t.Error("getRawTagValue(Make) should not be nil")
	}

	// Get non-existing tag
	value = app.getRawTagValue(&meta, imx.SpecEXIF, "NonExistent")
	if value != nil {
		t.Error("getRawTagValue(NonExistent) should be nil")
	}
}

func TestOutputText_MultipleFiles(t *testing.T) {
	testFile := "../../testdata/goldens/jpeg/google_iptc.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.opts.NoColor = true
	// Process same file twice
	code := app.Run([]string{testFile, testFile})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if code != 0 {
		t.Errorf("multiple files returned %d, want 0", code)
	}

	output := buf.String()
	// Should have output for both files
	count := strings.Count(output, "google_iptc.jpg")
	if count < 2 {
		t.Error("should have output for multiple files")
	}
}

func TestPrintError(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	app := NewApp()
	app.colors = Colorizer{enabled: false}
	app.printError("test error message")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if !strings.Contains(buf.String(), "Error:") {
		t.Error("printError should contain 'Error:'")
	}
	if !strings.Contains(buf.String(), "test error message") {
		t.Error("printError should contain the message")
	}
}

func TestPrintStats(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := NewApp()
	app.colors = Colorizer{enabled: false}
	stats := Stats{
		Start:   time.Now().Add(-time.Second),
		Total:   10,
		Success: 8,
		Errors:  2,
		Tags:    500,
	}
	app.printStats(stats)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	if !strings.Contains(output, "Statistics:") {
		t.Error("should contain 'Statistics:'")
	}
	if !strings.Contains(output, "10 total") {
		t.Error("should contain total count")
	}
	if !strings.Contains(output, "8 success") {
		t.Error("should contain success count")
	}
	if !strings.Contains(output, "2 errors") {
		t.Error("should contain error count")
	}
	if !strings.Contains(output, "500") {
		t.Error("should contain tag count")
	}
}

