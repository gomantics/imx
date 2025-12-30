package filter

import (
	"testing"

	"github.com/gomantics/imx"
)

func TestSearchFilter_ShouldInclude(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		tagName  string
		tagValue any
		want     bool
	}{
		{
			name:     "empty query allows all",
			query:    "",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "match in tag name",
			query:    "make",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "match in tag value",
			query:    "canon",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "partial match in name",
			query:    "date",
			tagName:  "DateTimeOriginal",
			tagValue: "2024:01:15 10:30:00",
			want:     true,
		},
		{
			name:     "partial match in value",
			query:    "2024",
			tagName:  "DateTimeOriginal",
			tagValue: "2024:01:15 10:30:00",
			want:     true,
		},
		{
			name:     "case insensitive name match",
			query:    "MAKE",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "case insensitive value match",
			query:    "CANON",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "no match",
			query:    "nikon",
			tagName:  "Make",
			tagValue: "Canon",
			want:     false,
		},
		{
			name:     "query with whitespace",
			query:    "  canon  ",
			tagName:  "Make",
			tagValue: "Canon",
			want:     true,
		},
		{
			name:     "match in numeric value",
			query:    "100",
			tagName:  "ISOSpeedRatings",
			tagValue: 100,
			want:     true,
		},
		{
			name:     "match in float value",
			query:    "3.14",
			tagName:  "FNumber",
			tagValue: 3.14,
			want:     true,
		},
		{
			name:     "match in tag name with array value",
			query:    "latitude",
			tagName:  "GPSLatitude",
			tagValue: []float64{40, 26, 46},
			want:     true, // Matches in tag name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewSearchFilter(tt.query)
			dir := imx.Directory{Name: "IFD0"}
			tag := imx.Tag{ID: imx.TagID("EXIF:" + tt.tagName), Name: tt.tagName, Value: tt.tagValue}

			got := filter.ShouldInclude(dir, tag)
			if got != tt.want {
				t.Errorf("SearchFilter.ShouldInclude() = %v, want %v (query=%q, name=%q, value=%v)",
					got, tt.want, tt.query, tt.tagName, tt.tagValue)
			}
		})
	}
}
