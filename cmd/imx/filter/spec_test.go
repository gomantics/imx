package filter

import (
	"testing"

	"github.com/gomantics/imx"
)

func TestSpecFilter_ShouldInclude(t *testing.T) {
	tests := []struct {
		name     string
		filterOn string
		dirName  string
		want     bool
	}{
		{
			name:     "empty filter allows all",
			filterOn: "",
			dirName:  "IFD0",
			want:     true,
		},
		{
			name:     "exact match lowercase - exif directory",
			filterOn: "ifd0",
			dirName:  "IFD0",
			want:     true,
		},
		{
			name:     "exact match uppercase - exif directory",
			filterOn: "IFD0",
			dirName:  "IFD0",
			want:     true,
		},
		{
			name:     "exact match mixed case - exif directory",
			filterOn: "Ifd0",
			dirName:  "IFD0",
			want:     true,
		},
		{
			name:     "no match",
			filterOn: "iptc",
			dirName:  "IFD0",
			want:     false,
		},
		{
			name:     "filter with whitespace",
			filterOn: "  ifd0  ",
			dirName:  "IFD0",
			want:     true,
		},
		{
			name:     "iptc match",
			filterOn: "iptc",
			dirName:  "IPTC",
			want:     true,
		},
		{
			name:     "xmp match",
			filterOn: "xmp",
			dirName:  "XMP",
			want:     true,
		},
		{
			name:     "icc match",
			filterOn: "icc",
			dirName:  "ICC",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewSpecFilter(tt.filterOn)
			dir := imx.Directory{Name: tt.dirName}
			tag := imx.Tag{ID: "TEST:Tag", Name: "Tag", Value: "value"}

			got := filter.ShouldInclude(dir, tag)
			if got != tt.want {
				t.Errorf("SpecFilter.ShouldInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}
