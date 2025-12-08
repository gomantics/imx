package filter

import (
	"testing"

	"github.com/gomantics/imx"
)

func TestSpecFilter_ShouldInclude(t *testing.T) {
	tests := []struct {
		name     string
		filterOn string
		dirSpec  imx.Spec
		want     bool
	}{
		{
			name:     "empty filter allows all",
			filterOn: "",
			dirSpec:  imx.SpecEXIF,
			want:     true,
		},
		{
			name:     "exact match lowercase",
			filterOn: "exif",
			dirSpec:  imx.SpecEXIF,
			want:     true,
		},
		{
			name:     "exact match uppercase",
			filterOn: "EXIF",
			dirSpec:  imx.SpecEXIF,
			want:     true,
		},
		{
			name:     "exact match mixed case",
			filterOn: "Exif",
			dirSpec:  imx.SpecEXIF,
			want:     true,
		},
		{
			name:     "no match",
			filterOn: "iptc",
			dirSpec:  imx.SpecEXIF,
			want:     false,
		},
		{
			name:     "filter with whitespace",
			filterOn: "  exif  ",
			dirSpec:  imx.SpecEXIF,
			want:     true,
		},
		{
			name:     "iptc match",
			filterOn: "iptc",
			dirSpec:  imx.SpecIPTC,
			want:     true,
		},
		{
			name:     "xmp match",
			filterOn: "xmp",
			dirSpec:  imx.SpecXMP,
			want:     true,
		},
		{
			name:     "icc match",
			filterOn: "icc",
			dirSpec:  imx.SpecICC,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewSpecFilter(tt.filterOn)
			dir := imx.Directory{Spec: tt.dirSpec, Name: "TestDir"}
			tag := imx.Tag{ID: "TEST:Tag", Name: "Tag", Value: "value"}

			got := filter.ShouldInclude(dir, tag)
			if got != tt.want {
				t.Errorf("SpecFilter.ShouldInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}
