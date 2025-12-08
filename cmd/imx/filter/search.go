package filter

import (
	"fmt"
	"strings"

	"github.com/gomantics/imx"
)

// SearchFilter searches for text in both tag names and values
type SearchFilter struct {
	query string // Search query (lowercase)
}

// NewSearchFilter creates a new search filter
// The query will be matched against both tag names and values (case-insensitive)
func NewSearchFilter(query string) *SearchFilter {
	return &SearchFilter{
		query: strings.ToLower(strings.TrimSpace(query)),
	}
}

// ShouldInclude returns true if the tag name or value contains the search query
func (f *SearchFilter) ShouldInclude(dir imx.Directory, tag imx.Tag) bool {
	if f.query == "" {
		return true
	}

	// Search in tag name
	nameLower := strings.ToLower(tag.Name)
	if strings.Contains(nameLower, f.query) {
		return true
	}

	// Search in tag value
	valueLower := strings.ToLower(fmt.Sprintf("%v", tag.Value))
	if strings.Contains(valueLower, f.query) {
		return true
	}

	return false
}
