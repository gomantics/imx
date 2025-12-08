package output

import (
	"io"

	"github.com/gomantics/imx"
)

// Result represents the processing result for a single file
type Result struct {
	File     string          // File path or URL
	Meta     *imx.Metadata   // Extracted metadata (nil if error)
	Tags     []TagInfo       // Filtered tags
	TagCount int             // Number of tags that passed filters
	Error    error           // Error if processing failed
}

// TagInfo holds a tag with its directory context
type TagInfo struct {
	Dir imx.Directory
	Tag imx.Tag
}

// Formatter formats processing results for output
type Formatter interface {
	// Format writes formatted output to the writer
	Format(w io.Writer, results []*Result) error
}

// Config holds configuration for formatters
type Config struct {
	// Display options
	NoColor bool   // Disable colored output
	Quiet   bool   // Suppress headers and decorations
	Full    bool   // Show full values without truncation

	// Format options
	GPSFormat  string // GPS coordinate format: url, dms, decimal
	TimeFormat string // Time format: iso, rfc3339, unix, human, or Go layout
}
