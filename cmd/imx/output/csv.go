package output

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/gomantics/imx/cmd/imx/ui"
)

// CSVFormatter outputs results as CSV
type CSVFormatter struct {
	config *Config
}

// Format writes CSV output
func (f *CSVFormatter) Format(w io.Writer, results []*Result) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"File", "Spec", "Tag", "Value"}); err != nil {
		return err
	}

	// Write data rows
	for _, result := range results {
		if result.Error != nil {
			// Write error row
			if err := writer.Write([]string{
				result.File,
				"ERROR",
				"",
				result.Error.Error(),
			}); err != nil {
				return err
			}
			continue
		}

		// Write tag rows
		for _, tagInfo := range result.Tags {
			spec := strings.ToUpper(tagInfo.Dir.Spec.String())

			// Apply time formatting if configured and tag is a time field
			value := ui.FormatValue(tagInfo.Tag.Value, f.config.Full)
			if f.config.TimeFormat != "" && isTimeField(tagInfo.Tag.Name) {
				value = ui.FormatTime(tagInfo.Tag.Value, f.config.TimeFormat)
			}

			if err := writer.Write([]string{
				result.File,
				spec,
				tagInfo.Tag.Name,
				value,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
