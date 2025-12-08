package filter

import (
	"fmt"
	"regexp"

	"github.com/gomantics/imx"
)

// PatternFilter filters tags using a regular expression pattern
type PatternFilter struct {
	pattern *regexp.Regexp
}

// NewPatternFilter creates a new pattern filter
// Returns an error if the pattern is invalid
func NewPatternFilter(pattern string) (*PatternFilter, error) {
	if pattern == "" {
		return &PatternFilter{pattern: nil}, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	return &PatternFilter{pattern: re}, nil
}

// ShouldInclude returns true if the tag name or value matches the regex pattern
func (f *PatternFilter) ShouldInclude(dir imx.Directory, tag imx.Tag) bool {
	if f.pattern == nil {
		return true
	}

	// Match against tag name
	if f.pattern.MatchString(tag.Name) {
		return true
	}

	// Match against tag value
	valueStr := fmt.Sprintf("%v", tag.Value)
	if f.pattern.MatchString(valueStr) {
		return true
	}

	return false
}
