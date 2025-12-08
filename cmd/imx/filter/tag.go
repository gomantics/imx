package filter

import (
	"strings"

	"github.com/gomantics/imx"
)

// TagFilter filters for a specific tag by name or ID
type TagFilter struct {
	query string // Tag name or ID to search for (lowercase)
}

// NewTagFilter creates a new tag filter
// query can be:
//   - Tag name: "Make", "Model"
//   - Full tag ID: "EXIF:Make", "XMP-dc:Title"
//   - Partial ID: ":Make" (matches any spec)
func NewTagFilter(query string) *TagFilter {
	return &TagFilter{
		query: strings.ToLower(strings.TrimSpace(query)),
	}
}

// ShouldInclude returns true if the tag matches the query
func (f *TagFilter) ShouldInclude(dir imx.Directory, tag imx.Tag) bool {
	if f.query == "" {
		return true
	}

	queryLower := f.query
	tagIDLower := strings.ToLower(string(tag.ID))
	nameLower := strings.ToLower(tag.Name)

	// Match by exact tag ID
	if tagIDLower == queryLower {
		return true
	}

	// Match by tag name
	if nameLower == queryLower {
		return true
	}

	// Match by "SPEC:Name" format
	if strings.HasSuffix(tagIDLower, ":"+queryLower) {
		return true
	}

	// Match by ":Name" (any spec)
	if strings.HasPrefix(queryLower, ":") {
		targetName := strings.TrimPrefix(queryLower, ":")
		if nameLower == targetName {
			return true
		}
	}

	return false
}
