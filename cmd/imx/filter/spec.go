package filter

import (
	"strings"

	"github.com/gomantics/imx"
)

// SpecFilter filters tags by metadata specification type
type SpecFilter struct {
	spec string // exif, iptc, xmp, icc (lowercase)
}

// NewSpecFilter creates a new spec filter
// spec should be one of: exif, iptc, xmp, icc (case-insensitive)
func NewSpecFilter(spec string) *SpecFilter {
	return &SpecFilter{
		spec: strings.ToLower(strings.TrimSpace(spec)),
	}
}

// ShouldInclude returns true if the tag's spec matches the filter
func (f *SpecFilter) ShouldInclude(dir imx.Directory, tag imx.Tag) bool {
	if f.spec == "" {
		return true
	}
	return strings.EqualFold(dir.Spec.String(), f.spec)
}
