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
	specWidth := 4
	nameWidth := 20
	for _, tagInfo := range result.Tags {
		spec := tagInfo.Dir.Spec.String()
		if len(spec) > specWidth {
			specWidth = len(spec)
		}
		if len(tagInfo.Tag.Name) > nameWidth && len(tagInfo.Tag.Name) <= 30 {
			nameWidth = len(tagInfo.Tag.Name)
		}
	}

	// Print header
	if f.config.NoColor {
		fmt.Fprintf(w, "%-*s  %-*s  %s\n", specWidth, "SPEC", nameWidth, "TAG", "VALUE")
	} else {
		ui.Dim.Fprintf(w, "%-*s  %-*s  %s\n", specWidth, "SPEC", nameWidth, "TAG", "VALUE")
	}
	fmt.Fprintln(w, strings.Repeat("─", 80))

	// Print rows
	for _, tagInfo := range result.Tags {
		spec := strings.ToUpper(tagInfo.Dir.Spec.String())
		name := tagInfo.Tag.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		value := ui.FormatValue(tagInfo.Tag.Value, f.config.Full)
		if len(value) > 45 && !f.config.Full {
			value = value[:42] + "..."
		}

		if f.config.NoColor {
			fmt.Fprintf(w, "%-*s  %-*s  %s\n", specWidth, spec, nameWidth, name, value)
		} else {
			color := ui.SpecColor(tagInfo.Dir.Spec)
			color.Fprintf(w, "%-*s", specWidth, spec)
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
