package filter

import "github.com/gomantics/imx"

// Filter determines whether a tag should be included in the output
type Filter interface {
	ShouldInclude(dir imx.Directory, tag imx.Tag) bool
}

// Chain combines multiple filters with AND logic
// A tag must pass all filters to be included
type Chain struct {
	filters []Filter
}

// NewChain creates a new filter chain
func NewChain(filters ...Filter) *Chain {
	return &Chain{filters: filters}
}

// ShouldInclude returns true if the tag passes all filters in the chain
func (c *Chain) ShouldInclude(dir imx.Directory, tag imx.Tag) bool {
	for _, f := range c.filters {
		if !f.ShouldInclude(dir, tag) {
			return false
		}
	}
	return true
}

// Add adds a filter to the chain
func (c *Chain) Add(f Filter) {
	c.filters = append(c.filters, f)
}

// Len returns the number of filters in the chain
func (c *Chain) Len() int {
	return len(c.filters)
}

// PassThrough is a filter that allows all tags through
type PassThrough struct{}

// ShouldInclude always returns true
func (p *PassThrough) ShouldInclude(dir imx.Directory, tag imx.Tag) bool {
	return true
}
