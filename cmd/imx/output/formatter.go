package output

import (
	"fmt"
	"io"
)

// NewFormatter creates a formatter based on the format name
func NewFormatter(format string, config *Config) (Formatter, error) {
	if config == nil {
		config = &Config{}
	}

	switch format {
	case "text":
		return &TextFormatter{config: config}, nil
	case "json":
		return &JSONFormatter{config: config}, nil
	case "table":
		return &TableFormatter{config: config}, nil
	case "csv":
		return &CSVFormatter{config: config}, nil
	case "summary":
		return &SummaryFormatter{config: config}, nil
	default:
		return nil, fmt.Errorf("unknown format: %s (supported: text, json, table, csv, summary)", format)
	}
}

// FormatSingle is a helper to format a single result
func FormatSingle(w io.Writer, result *Result, format string, config *Config) error {
	formatter, err := NewFormatter(format, config)
	if err != nil {
		return err
	}
	return formatter.Format(w, []*Result{result})
}
