package filter

import (
	"testing"

	"github.com/gomantics/imx"
)

func TestChain_ShouldInclude(t *testing.T) {
	// Create test data
	dir := imx.Directory{Spec: imx.SpecEXIF, Name: "IFD0"}
	tag := imx.Tag{ID: "EXIF:Make", Name: "Make", Value: "Canon"}

	tests := []struct {
		name    string
		filters []Filter
		want    bool
	}{
		{
			name:    "empty chain allows all",
			filters: []Filter{},
			want:    true,
		},
		{
			name: "single filter - pass",
			filters: []Filter{
				NewSpecFilter("exif"),
			},
			want: true,
		},
		{
			name: "single filter - fail",
			filters: []Filter{
				NewSpecFilter("iptc"),
			},
			want: false,
		},
		{
			name: "multiple filters - all pass",
			filters: []Filter{
				NewSpecFilter("exif"),
				NewTagFilter("Make"),
			},
			want: true,
		},
		{
			name: "multiple filters - one fails",
			filters: []Filter{
				NewSpecFilter("exif"),
				NewTagFilter("Model"),
			},
			want: false,
		},
		{
			name: "PassThrough filter",
			filters: []Filter{
				&PassThrough{},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewChain(tt.filters...)
			got := chain.ShouldInclude(dir, tag)
			if got != tt.want {
				t.Errorf("Chain.ShouldInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChain_Add(t *testing.T) {
	chain := NewChain()
	if chain.Len() != 0 {
		t.Errorf("New chain should have length 0, got %d", chain.Len())
	}

	chain.Add(NewSpecFilter("exif"))
	if chain.Len() != 1 {
		t.Errorf("After adding 1 filter, chain should have length 1, got %d", chain.Len())
	}

	chain.Add(NewTagFilter("Make"))
	if chain.Len() != 2 {
		t.Errorf("After adding 2 filters, chain should have length 2, got %d", chain.Len())
	}
}

func TestPassThrough_ShouldInclude(t *testing.T) {
	dir := imx.Directory{Spec: imx.SpecEXIF, Name: "IFD0"}
	tag := imx.Tag{ID: "EXIF:Make", Name: "Make", Value: "Canon"}

	filter := &PassThrough{}
	if !filter.ShouldInclude(dir, tag) {
		t.Error("PassThrough should always return true")
	}
}
