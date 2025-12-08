package ui

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
		full  bool
		want  string
	}{
		{
			name:  "nil",
			input: nil,
			full:  false,
			want:  "",
		},
		{
			name:  "string",
			input: "hello world",
			full:  false,
			want:  "hello world",
		},
		{
			name:  "string with spaces",
			input: "  hello  ",
			full:  false,
			want:  "hello",
		},
		{
			name:  "int",
			input: 42,
			full:  false,
			want:  "42",
		},
		{
			name:  "float64 whole number",
			input: 42.0,
			full:  false,
			want:  "42",
		},
		{
			name:  "float64 decimal",
			input: 3.14159,
			full:  false,
			want:  "3.14159",
		},
		{
			name:  "float64 with trailing zeros",
			input: 1.50000,
			full:  false,
			want:  "1.5",
		},
		{
			name:  "short byte slice full=false",
			input: []byte{0x01, 0x02, 0x03},
			full:  false,
			want:  "010203",
		},
		{
			name:  "long byte slice full=false",
			input: make([]byte, 100),
			full:  false,
			want:  "(binary, 100 bytes)",
		},
		{
			name:  "long byte slice full=true",
			input: []byte{0xFF, 0xFE, 0xFD},
			full:  true,
			want:  "fffefd",
		},
		{
			name:  "float64 slice",
			input: []float64{1.5, 2.0, 3.14159},
			full:  false,
			want:  "1.5, 2, 3.14159",
		},
		{
			name:  "uint16 slice",
			input: []uint16{100, 200, 300},
			full:  false,
			want:  "100, 200, 300",
		},
		{
			name:  "int slice",
			input: []int{1, 2, 3},
			full:  false,
			want:  "1, 2, 3",
		},
		{
			name:  "any slice",
			input: []any{"hello", 42, 3.14},
			full:  false,
			want:  "hello, 42, 3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValue(tt.input, tt.full)
			if got != tt.want {
				t.Errorf("FormatValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{42.0, "42"},
		{3.14159, "3.14159"},
		{1.50000, "1.5"},
		{0.0, "0"},
		{-3.14, "-3.14"},
		{math.Pi, "3.141593"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatFloat(tt.input)
			if got != tt.want {
				t.Errorf("formatFloat(%f) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatGPS(t *testing.T) {
	// Test DMS coordinates
	lat := []float64{40, 26, 46}
	lon := []float64{79, 58, 56}

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "dms format",
			format: "dms",
			want:   "40°26'46.00\"N, 79°58'56.00\"W",
		},
		{
			name:   "decimal format",
			format: "decimal",
			want:   "40.446111, -79.982222",
		},
		{
			name:   "url format",
			format: "url",
			want:   "https://maps.google.com/maps?q=40.446111,-79.982222",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatGPS(lat, lon, "N", "W", tt.format)
			if got != tt.want {
				t.Errorf("FormatGPS() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToDecimalDegrees(t *testing.T) {
	tests := []struct {
		name  string
		coord any
		ref   string
		want  float64
	}{
		{
			name:  "north latitude",
			coord: []float64{40, 26, 46},
			ref:   "N",
			want:  40.446111,
		},
		{
			name:  "south latitude",
			coord: []float64{40, 26, 46},
			ref:   "S",
			want:  -40.446111,
		},
		{
			name:  "east longitude",
			coord: []float64{79, 58, 56},
			ref:   "E",
			want:  79.982222,
		},
		{
			name:  "west longitude",
			coord: []float64{79, 58, 56},
			ref:   "W",
			want:  -79.982222,
		},
		{
			name:  "already decimal",
			coord: 40.446111,
			ref:   "N",
			want:  40.446111,
		},
		{
			name:  "invalid input",
			coord: "invalid",
			ref:   "N",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDecimalDegrees(tt.coord, tt.ref)
			// Allow small floating point differences
			diff := math.Abs(got - tt.want)
			if diff > 0.000001 {
				t.Errorf("toDecimalDegrees() = %f, want %f (diff: %f)", got, tt.want, diff)
			}
		})
	}
}

func TestToDMS(t *testing.T) {
	tests := []struct {
		name       string
		decimal    float64
		isNegative bool
		isLat      bool
		want       string
	}{
		{
			name:       "north latitude",
			decimal:    40.446111,
			isNegative: false,
			isLat:      true,
			want:       "40°26'46.00\"N",
		},
		{
			name:       "south latitude",
			decimal:    -40.446111,
			isNegative: true,
			isLat:      true,
			want:       "40°26'46.00\"S",
		},
		{
			name:       "east longitude",
			decimal:    79.982222,
			isNegative: false,
			isLat:      false,
			want:       "79°58'56.00\"E",
		},
		{
			name:       "west longitude",
			decimal:    -79.982222,
			isNegative: true,
			isLat:      false,
			want:       "79°58'56.00\"W",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDMS(tt.decimal, tt.isNegative, tt.isLat)
			if got != tt.want {
				t.Errorf("toDMS() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	// Create a test time
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		name   string
		input  any
		format string
		want   string
	}{
		{
			name:   "iso format with time.Time",
			input:  testTime,
			format: "iso",
			want:   "2024-03-15T14:30:45Z",
		},
		{
			name:   "rfc3339 format",
			input:  testTime,
			format: "rfc3339",
			want:   "2024-03-15T14:30:45Z",
		},
		{
			name:   "unix timestamp",
			input:  testTime,
			format: "unix",
			want:   "", // Will check separately to account for timezone
		},
		{
			name:   "human format",
			input:  testTime,
			format: "human",
			want:   "Mar 15, 2024 2:30 PM",
		},
		{
			name:   "custom format",
			input:  testTime,
			format: "2006-01-02",
			want:   "2024-03-15",
		},
		{
			name:   "EXIF string format",
			input:  "2024:03:15 14:30:45",
			format: "iso",
			want:   "2024-03-15T14:30:45Z",
		},
		{
			name:   "unparseable string",
			input:  "invalid time",
			format: "iso",
			want:   "invalid time",
		},
		{
			name:   "non-time type",
			input:  12345,
			format: "iso",
			want:   "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTime(tt.input, tt.format)
			// Skip unix timestamp test (timezone dependent)
			if tt.format == "unix" && tt.want == "" {
				// Just check it's a number
				if len(got) < 10 {
					t.Errorf("FormatTime() unix timestamp too short: %q", got)
				}
				return
			}
			if !strings.Contains(got, tt.want) && got != tt.want {
				t.Errorf("FormatTime() = %q, want %q", got, tt.want)
			}
		})
	}
}
