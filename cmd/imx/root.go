package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/gomantics/imx/cmd/imx/filter"
	"github.com/gomantics/imx/cmd/imx/output"
	"github.com/gomantics/imx/cmd/imx/processor"
	"github.com/gomantics/imx/cmd/imx/ui"
	"github.com/gomantics/imx/cmd/imx/util"
)

const version = "1.0.0"

var (
	// Output options
	formatFlag     string
	noColorFlag    bool
	quietFlag      bool
	fullFlag       bool
	timeFormatFlag string
	gpsFormatFlag  string

	// Filter options
	specFlag    string
	tagFlag     string
	searchFlag  string
	patternFlag string

	// Processing options
	recursiveFlag bool
	workersFlag   int
	verboseFlag   bool
	stopOnErrFlag bool
	progressFlag  bool

	// Version flag
	versionFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "imx [flags] <file>...",
	Short: "Extract and analyze metadata from images, audio, and video files",
	Long: `imx - Metadata Extractor

A powerful command-line tool for extracting, querying, and analyzing
metadata from images, audio, and video files.

Supports: EXIF, IPTC, XMP, ICC profiles, ID3 tags, FLAC metadata, and more.
Formats: JPEG, PNG, GIF, WebP, TIFF, CR2, HEIC, MP3, FLAC, MP4/M4A

Examples:
  # Extract all metadata from files
  imx photo.jpg
  imx song.mp3
  imx video.mp4

  # Extract specific directory in JSON format
  imx --dir IFD0 --format json photo.jpg
  imx --dir ID3v2_4 --format json song.mp3

  # Search for GPS tags
  imx --search gps *.jpg

  # Process multiple files with progress bar
  imx --progress *.jpg

  # Filter by tag name or ID
  imx --tag Make --tag Model photo.jpg
  imx --tag Artist --tag Album song.mp3

  # Format GPS coordinates
  imx --gps-format decimal photo.jpg
  imx --gps-format dms photo.jpg
  imx --gps-format url photo.jpg

  # Format timestamps
  imx --time-format iso photo.jpg
  imx --time-format rfc3339 photo.jpg
  imx --time-format unix photo.jpg
  imx --time-format human photo.jpg
  imx --time-format "2006-01-02 15:04:05" photo.jpg

Supported options:
  - Output: text, json, csv, table, summary
  - Time: iso, rfc3339, unix, human, or custom Go layout
  - GPS: decimal, dms, url
`,
	RunE: runRoot,
	Args: func(cmd *cobra.Command, args []string) error {
		// Allow --version without files
		if versionFlag {
			return nil
		}
		// Allow no args if stdin is being piped
		if len(args) == 0 && isStdinPiped() {
			return nil
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Output flags
	rootCmd.Flags().StringVarP(&formatFlag, "format", "f", "", "Output format (text|json|csv|table|summary)")
	rootCmd.Flags().BoolVar(&noColorFlag, "no-color", false, "Disable colored output")
	rootCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "Quiet mode (minimal output)")
	rootCmd.Flags().BoolVar(&fullFlag, "full", false, "Show full values without truncation")
	rootCmd.Flags().StringVar(&timeFormatFlag, "time-format", "", "Time format (iso|rfc3339|unix|human|<layout>)")
	rootCmd.Flags().StringVar(&gpsFormatFlag, "gps-format", "dms", "GPS format (decimal|dms|url)")

	// Filter flags
	rootCmd.Flags().StringVar(&specFlag, "dir", "", "Filter by directory (IFD0|ExifIFD|GPS|ID3v2_4|FLAC-StreamInfo|etc)")
	rootCmd.Flags().StringVar(&tagFlag, "tag", "", "Filter by tag name or ID")
	rootCmd.Flags().StringVar(&searchFlag, "search", "", "Search in tag names and values")
	rootCmd.Flags().StringVar(&patternFlag, "pattern", "", "Filter by regex pattern")

	// Processing flags
	rootCmd.Flags().BoolVarP(&recursiveFlag, "recursive", "r", false, "Process directories recursively")
	rootCmd.Flags().IntVarP(&workersFlag, "workers", "w", runtime.NumCPU(), "Number of worker goroutines")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVar(&stopOnErrFlag, "stop-on-error", false, "Stop processing on first error")
	rootCmd.Flags().BoolVarP(&progressFlag, "progress", "p", false, "Show progress bar")

	// Version flag
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Print version and exit")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Handle version flag
	if versionFlag {
		fmt.Printf("imx version %s\n", version)
		return nil
	}

	// Disable colors if requested or not a TTY
	if noColorFlag || !isTerminal() {
		ui.DisableColors()
	}

	// Validate format flag
	if formatFlag != "" {
		validFormats := map[string]bool{
			"text": true, "json": true, "csv": true, "table": true, "summary": true,
		}
		if !validFormats[formatFlag] {
			return fmt.Errorf("invalid format %q, must be one of: text, json, csv, table, summary", formatFlag)
		}
	}

	// Validate gps-format flag
	validGPSFormats := map[string]bool{
		"decimal": true, "dms": true, "url": true,
	}
	if !validGPSFormats[gpsFormatFlag] {
		return fmt.Errorf("invalid gps-format %q, must be one of: decimal, dms, url", gpsFormatFlag)
	}

	var files []string
	var err error

	// Check if reading from stdin
	if len(args) == 0 && isStdinPiped() {
		// Read from stdin - use a special marker
		files = []string{"-"}
	} else {
		// Expand file paths (glob patterns, directories, etc.)
		files, err = util.ExpandFiles(args, recursiveFlag)
		if err != nil {
			return fmt.Errorf("failed to expand files: %w", err)
		}

		if len(files) == 0 {
			return fmt.Errorf("no files found")
		}
	}

	// Build filter chain
	var filters []filter.Filter
	if specFlag != "" {
		filters = append(filters, filter.NewSpecFilter(specFlag))
	}
	if tagFlag != "" {
		filters = append(filters, filter.NewTagFilter(tagFlag))
	}
	if searchFlag != "" {
		filters = append(filters, filter.NewSearchFilter(searchFlag))
	}
	if patternFlag != "" {
		f, err := filter.NewPatternFilter(patternFlag)
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}
		filters = append(filters, f)
	}

	var filterChain filter.Filter
	if len(filters) > 0 {
		filterChain = filter.NewChain(filters...)
	}

	// Create processor
	proc := processor.New(&processor.Config{
		Workers:      workersFlag,
		Verbose:      verboseFlag,
		Quiet:        quietFlag,
		StopOnErr:    stopOnErrFlag,
		ShowProgress: progressFlag && len(files) > 1,
		Filter:       filterChain,
	})

	// Process files
	ctx := context.Background()
	results, err := proc.Process(ctx, files)
	if err != nil && stopOnErrFlag {
		return err
	}

	// Auto-select format based on number of files if not explicitly set
	selectedFormat := formatFlag
	if selectedFormat == "" {
		if len(files) == 1 {
			selectedFormat = "table"
		} else {
			// For multiple files, default to JSON for machine-readable output
			selectedFormat = "json"
		}
	}

	// Create formatter
	formatter, err := output.NewFormatter(selectedFormat, &output.Config{
		NoColor:    noColorFlag,
		Quiet:      quietFlag,
		Full:       fullFlag,
		TimeFormat: timeFormatFlag,
		GPSFormat:  gpsFormatFlag,
	})
	if err != nil {
		return err
	}

	// Output results
	if err := formatter.Format(os.Stdout, results); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Report errors if any (when not in stop-on-error mode)
	if !stopOnErrFlag {
		errorCount := 0
		for _, result := range results {
			if result.Error != nil {
				errorCount++
			}
		}
		if errorCount > 0 {
			if !quietFlag {
				fmt.Fprintf(os.Stderr, "\nProcessed %d files with %d errors\n", len(results), errorCount)
			}
			// Return error to indicate processing failures
			return fmt.Errorf("%d file(s) failed to process", errorCount)
		}
	}

	return nil
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// isStdinPiped checks if stdin is being piped
func isStdinPiped() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) == 0
}
