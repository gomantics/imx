package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/gomantics/imx/cmd/imx/ui"
)

// TableFormatter outputs results as an aligned table
type TableFormatter struct {
	config *Config
}

// Format writes table output
func (f *TableFormatter) Format(w io.Writer, results []*Result) error {
	for i, result := range results {
		// Separator between files
		if i > 0 && !f.config.Quiet {
			fmt.Fprintln(w)
		}

		if err := f.formatSingle(w, result); err != nil {
			return err
		}
	}
	return nil
}

func (f *TableFormatter) formatSingle(w io.Writer, result *Result) error {
	// Print file header
	if !f.config.Quiet {
		if f.config.NoColor {
			fmt.Fprintf(w, "%s\n", result.File)
		} else {
			ui.Bold.Fprintf(w, "%s\n", result.File)
		}
		fmt.Fprintln(w, strings.Repeat("─", min(80, len(result.File)+10)))
	}

	// Handle errors
	if result.Error != nil {
		if f.config.NoColor {
			fmt.Fprintf(w, "Error: %v\n", result.Error)
		} else {
			ui.Red.Fprintf(w, "Error: %v\n", result.Error)
		}
		return nil
	}

	// No tags
	if len(result.Tags) == 0 {
		fmt.Fprintln(w, "No metadata found")
		return nil
	}

	// Calculate column widths
	dirWidth := 4
	nameWidth := 20
	for _, tagInfo := range result.Tags {
		dir := tagInfo.Dir.Name
		if len(dir) > dirWidth {
			dirWidth = len(dir)
		}
		if len(tagInfo.Tag.Name) > nameWidth && len(tagInfo.Tag.Name) <= 30 {
			nameWidth = len(tagInfo.Tag.Name)
		}
	}

	// Print header
	if f.config.NoColor {
		fmt.Fprintf(w, "%-*s  %-*s  %s\n", dirWidth, "DIR", nameWidth, "TAG", "VALUE")
	} else {
		ui.Dim.Fprintf(w, "%-*s  %-*s  %s\n", dirWidth, "DIR", nameWidth, "TAG", "VALUE")
	}
	fmt.Fprintln(w, strings.Repeat("─", 80))

	// Print rows
	for _, tagInfo := range result.Tags {
		dir := tagInfo.Dir.Name
		name := tagInfo.Tag.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		// Apply time formatting if configured and tag is a time field
		value := ui.FormatValue(tagInfo.Tag.Value, f.config.Full)
		if f.config.TimeFormat != "" && isTimeField(name) {
			value = ui.FormatTime(tagInfo.Tag.Value, f.config.TimeFormat)
		}
		if len(value) > 45 && !f.config.Full {
			value = value[:42] + "..."
		}

		if f.config.NoColor {
			fmt.Fprintf(w, "%-*s  %-*s  %s\n", dirWidth, dir, nameWidth, name, value)
		} else {
			color := ui.SpecColor(tagInfo.Dir.Name)
			color.Fprintf(w, "%-*s", dirWidth, dir)
			fmt.Fprintf(w, "  %-*s  %s\n", nameWidth, name, value)
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
