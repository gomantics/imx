package filter

import (
	"testing"

	"github.com/gomantics/imx"
)

func TestTagFilter_ShouldInclude(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		tagID   imx.TagID
		tagName string
		want    bool
	}{
		{
			name:    "empty query allows all",
			query:   "",
			tagID:   "EXIF:Make",
			tagName: "Make",
			want:    true,
		},
		{
			name:    "exact tag name match",
			query:   "Make",
			tagID:   "EXIF:Make",
			tagName: "Make",
			want:    true,
		},
		{
			name:    "case insensitive tag name",
			query:   "make",
			tagID:   "EXIF:Make",
			tagName: "Make",
			want:    true,
		},
		{
			name:    "exact tag ID match",
			query:   "EXIF:Make",
			tagID:   "EXIF:Make",
			tagName: "Make",
			want:    true,
		},
		{
			name:    "case insensitive tag ID",
			query:   "exif:make",
			tagID:   "EXIF:Make",
			tagName: "Make",
			want:    true,
		},
		{
			name:    "partial ID with colon",
			query:   ":Make",
			tagID:   "EXIF:Make",
			tagName: "Make",
			want:    true,
		},
		{
			name:    "partial ID matches any spec",
			query:   ":Make",
			tagID:   "XMP:Make",
			tagName: "Make",
			want:    true,
		},
		{
			name:    "no match",
			query:   "Model",
			tagID:   "EXIF:Make",
			tagName: "Make",
			want:    false,
		},
		{
			name:    "query with whitespace",
			query:   "  Make  ",
			tagID:   "EXIF:Make",
			tagName: "Make",
			want:    true,
		},
		{
			name:    "XMP namespace tag",
			query:   "Title",
			tagID:   "XMP-dc:Title",
			tagName: "Title",
			want:    true,
		},
		{
			name:    "full XMP tag ID",
			query:   "XMP-dc:Title",
			tagID:   "XMP-dc:Title",
			tagName: "Title",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewTagFilter(tt.query)
			dir := imx.Directory{Name: "IFD0"}
			tag := imx.Tag{ID: tt.tagID, Name: tt.tagName, Value: "test"}

			got := filter.ShouldInclude(dir, tag)
			if got != tt.want {
				t.Errorf("TagFilter.ShouldInclude() = %v, want %v (query=%q, tagID=%q, tagName=%q)",
					got, tt.want, tt.query, tt.tagID, tt.tagName)
			}
		})
	}
}
