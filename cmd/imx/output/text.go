package output

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gomantics/imx/cmd/imx/ui"
)

// TextFormatter outputs results as hierarchical text
type TextFormatter struct {
	config *Config
}

// Format writes text output
func (f *TextFormatter) Format(w io.Writer, results []*Result) error {
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

func (f *TextFormatter) formatSingle(w io.Writer, result *Result) error {
	// Print file header
	if !f.config.Quiet {
		if f.config.NoColor {
			fmt.Fprintf(w, "%s\n", result.File)
		} else {
			ui.Bold.Fprintf(w, "%s\n", result.File)
		}
		lineLen := min(80, len(result.File)+10)
		fmt.Fprintln(w, strings.Repeat("─", lineLen))
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

	// Group tags by directory type
	groups := f.groupTags(result.Tags)

	// Sort directory types by priority
	priority := map[string]int{"exif": 0, "iptc": 1, "xmp": 2, "icc": 3}
	var dirOrder []string
	for dirType := range groups {
		dirOrder = append(dirOrder, dirType)
	}
	sort.Slice(dirOrder, func(i, j int) bool {
		pi, oki := priority[dirOrder[i]]
		pj, okj := priority[dirOrder[j]]
		if oki && okj {
			return pi < pj
		}
		if oki {
			return true
		}
		if okj {
			return false
		}
		return dirOrder[i] < dirOrder[j]
	})

	// Output each directory type
	for _, dirTypeName := range dirOrder {
		group := groups[dirTypeName]

		// Directory type header
		fmt.Fprintln(w)
		if f.config.NoColor {
			fmt.Fprintf(w, "[%s]\n", dirTypeName)
		} else {
			dirTypeColor := ui.BoldSpecColor(group.dirTypeName)
			dirTypeColor.Fprintf(w, "[%s]\n", dirTypeName)
		}

		// Get directory names sorted
		var dirNames []string
		for name := range group.dirs {
			dirNames = append(dirNames, name)
		}
		sort.Strings(dirNames)

		// Output each directory
		for _, dirName := range dirNames {
			dirTags := group.dirs[dirName]

			// Directory subheader
			if f.config.NoColor {
				fmt.Fprintf(w, "  %s\n", dirName)
			} else {
				ui.Dim.Fprintf(w, "  %s\n", dirName)
			}

			// Find max name length for alignment
			maxLen := 0
			for _, t := range dirTags {
				if len(t.Tag.Name) > maxLen && len(t.Tag.Name) <= 35 {
					maxLen = len(t.Tag.Name)
				}
			}
			if maxLen < 15 {
				maxLen = 15
			}

			// Output tags
			for _, t := range dirTags {
				name := t.Tag.Name
				if len(name) > 35 {
					name = name[:32] + "..."
				}

				// Apply time formatting if configured and tag is a time field
				value := ui.FormatValue(t.Tag.Value, f.config.Full)
				if f.config.TimeFormat != "" && isTimeField(t.Tag.Name) {
					value = ui.FormatTime(t.Tag.Value, f.config.TimeFormat)
				}
				if len(value) > 55 && !f.config.Full {
					value = value[:52] + "..."
				}

				if f.config.NoColor {
					fmt.Fprintf(w, "    %-*s : %s\n", maxLen, name, value)
				} else {
					ui.Dim.Fprintf(w, "    %-*s", maxLen, name)
					fmt.Fprintf(w, " : %s\n", value)
				}
			}
		}
	}

	return nil
}

type dirTypeGroup struct {
	dirTypeName string
	dirs        map[string][]TagInfo
}

func (f *TextFormatter) groupTags(tags []TagInfo) map[string]*dirTypeGroup {
	groups := make(map[string]*dirTypeGroup)

	for _, t := range tags {
		dirTypeName := t.Dir.Name
		if groups[dirTypeName] == nil {
			groups[dirTypeName] = &dirTypeGroup{
				dirTypeName: dirTypeName,
				dirs:        make(map[string][]TagInfo),
			}
		}
		groups[dirTypeName].dirs[t.Dir.Name] = append(groups[dirTypeName].dirs[t.Dir.Name], t)
	}

	return groups
}
