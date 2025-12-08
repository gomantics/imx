package filter

import (
	"testing"

	"github.com/gomantics/imx"
)

func TestNewPatternFilter(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{
			name:    "valid pattern",
			pattern: "^Make$",
			wantErr: false,
		},
		{
			name:    "empty pattern",
			pattern: "",
			wantErr: false,
		},
		{
			name:    "invalid pattern",
			pattern: "[invalid",
			wantErr: true,
		},
		{
			name:    "complex pattern",
			pattern: "Date.*Original",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPatternFilter(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPatternFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPatternFilter_ShouldInclude(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		tagName  string
		tagValue any
		want     bool
	}{
		{
			name:     "empty pattern allows all",
			pattern:  "",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "exact match name",
			pattern:  "^Make$",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "partial match name",
			pattern:  "Date",
			tagName:  "DateTimeOriginal",
			tagValue: "2024:01:15",
			want:     true,
		},
		{
			name:     "match in value",
			pattern:  "Canon",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "regex in value",
			pattern:  "202[0-9]",
			tagName:  "Year",
			tagValue: "2024",
			want:     true,
		},
		{
			name:     "no match",
			pattern:  "Nikon",
			tagName:  "Make",
			tagValue: "Canon",
			want:     false,
		},
		{
			name:     "case sensitive match",
			pattern:  "canon",
			tagName:  "Make",
			tagValue: "Canon",
			want:     false,
		},
		{
			name:     "case insensitive pattern",
			pattern:  "(?i)canon",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "match numeric value",
			pattern:  "^100$",
			tagName:  "ISOSpeedRatings",
			tagValue: 100,
			want:     true,
		},
		{
			name:     "wildcard pattern",
			pattern:  ".*Time.*",
			tagName:  "DateTimeOriginal",
			tagValue: "2024:01:15",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewPatternFilter(tt.pattern)
			if err != nil {
				t.Fatalf("NewPatternFilter() error = %v", err)
			}

			dir := imx.Directory{Spec: imx.SpecEXIF, Name: "IFD0"}
			tag := imx.Tag{ID: imx.TagID("EXIF:" + tt.tagName), Name: tt.tagName, Value: tt.tagValue}

			got := filter.ShouldInclude(dir, tag)
			if got != tt.want {
				t.Errorf("PatternFilter.ShouldInclude() = %v, want %v (pattern=%q, name=%q, value=%v)",
					got, tt.want, tt.pattern, tt.tagName, tt.tagValue)
			}
		})
	}
}
