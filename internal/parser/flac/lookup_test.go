package flac

import (
	"testing"
)

func TestGetPictureType(t *testing.T) {
	tests := []struct {
		name     string
		typeCode uint32
		want     string
	}{
		{"Other", 0, "Other"},
		{"32x32 PNG File Icon", 1, "32x32 PNG File Icon"},
		{"Other File Icon", 2, "Other File Icon"},
		{"Cover (front)", 3, "Cover (front)"},
		{"Cover (back)", 4, "Cover (back)"},
		{"Leaflet Page", 5, "Leaflet Page"},
		{"Media", 6, "Media"},
		{"Lead Artist/Lead Performer/Soloist", 7, "Lead Artist/Lead Performer/Soloist"},
		{"Artist/Performer", 8, "Artist/Performer"},
		{"Conductor", 9, "Conductor"},
		{"Band/Orchestra", 10, "Band/Orchestra"},
		{"Composer", 11, "Composer"},
		{"Lyricist/Text Writer", 12, "Lyricist/Text Writer"},
		{"Recording Location", 13, "Recording Location"},
		{"During Recording", 14, "During Recording"},
		{"During Performance", 15, "During Performance"},
		{"Movie/Video Screen Capture", 16, "Movie/Video Screen Capture"},
		{"A Bright Colored Fish", 17, "A Bright Colored Fish"},
		{"Illustration", 18, "Illustration"},
		{"Band/Artist Logotype", 19, "Band/Artist Logotype"},
		{"Publisher/Studio Logotype", 20, "Publisher/Studio Logotype"},
		{"Unknown type - low value", 21, "Unknown (21)"},
		{"Unknown type - high value", 999, "Unknown (999)"},
		{"Unknown type - max uint32", 4294967295, "Unknown (4294967295)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPictureType(tt.typeCode); got != tt.want {
				t.Errorf("getPictureType(%d) = %q, want %q", tt.typeCode, got, tt.want)
			}
		})
	}
}
