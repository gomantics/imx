package output

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gomantics/imx/cmd/imx/ui"
)

// JSONFormatter outputs results as JSON
type JSONFormatter struct {
	config *Config
}

// Format writes JSON output
func (f *JSONFormatter) Format(w io.Writer, results []*Result) error {
	// Single file: output as object
	if len(results) == 1 {
		return f.formatSingle(w, results[0])
	}

	// Multiple files: output as array
	return f.formatMultiple(w, results)
}

func (f *JSONFormatter) formatSingle(w io.Writer, result *Result) error {
	if result.Error != nil {
		errObj := map[string]any{
			"SourceFile": result.File,
			"Error":      result.Error.Error(),
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(errObj)
	}

	obj := f.buildObject(result)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(obj)
}

func (f *JSONFormatter) formatMultiple(w io.Writer, results []*Result) error {
	var objects []map[string]any

	for _, result := range results {
		if result.Error != nil {
			objects = append(objects, map[string]any{
				"SourceFile": result.File,
				"Error":      result.Error.Error(),
			})
			continue
		}

		objects = append(objects, f.buildObject(result))
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(objects)
}

func (f *JSONFormatter) buildObject(result *Result) map[string]any {
	obj := map[string]any{
		"SourceFile": result.File,
	}

	// Group tags by spec
	specs := make(map[string]map[string]any)

	for _, tagInfo := range result.Tags {
		specName := strings.ToUpper(tagInfo.Dir.Spec.String())
		if specs[specName] == nil {
			specs[specName] = make(map[string]any)
		}

		// Format value for JSON
		value := f.formatValue(tagInfo.Tag.Value)
		specs[specName][tagInfo.Tag.Name] = value
	}

	// Add spec data to object
	for spec, data := range specs {
		obj[spec] = data
	}

	return obj
}

func (f *JSONFormatter) formatValue(v any) any {
	switch val := v.(type) {
	case []byte:
		// For JSON, represent binary data differently
		if len(val) > 100 {
			return map[string]any{
				"type": "binary",
				"size": len(val),
			}
		}
		// Small byte arrays as hex string
		return ui.FormatValue(val, true)
	default:
		return v
	}
}
