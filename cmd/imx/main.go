// imx - Image Metadata Extractor CLI
//
// A powerful command-line tool for extracting, querying, and analyzing
// metadata from images. Supports EXIF, IPTC, XMP, and ICC color profiles.
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gomantics/imx"
)

const version = "0.1.0"

func main() {
	app := NewApp()
	os.Exit(app.Run(os.Args[1:]))
}

// App holds CLI state
type App struct {
	opts      Options
	colors    Colorizer
	extractor *imx.Extractor
}

// Options holds all CLI options
type Options struct {
	// Input
	Files     []string
	Recursive bool
	Stdin     bool // Read from stdin
	Timeout   int  // HTTP timeout in seconds for URLs

	// Output format
	Format  string // text, json, table, csv, summary
	NoColor bool
	Quiet   bool
	Full    bool // Show full values without truncation

	// Filtering
	Spec    string // Filter by spec
	Tag     string // Get specific tag
	Search  string // Search tag names/values
	Pattern string // Regex pattern for filtering

	// Features
	GPS    string // GPS format: url, dms, decimal
	Stats  bool   // Show statistics after batch
	Export string // Export sidecar file

	// Info
	Help    bool
	Version bool
}

func NewApp() *App {
	return &App{
		extractor: imx.New(),
	}
}

func (a *App) Run(args []string) int {
	a.opts = a.parseArgs(args)
	a.colors = Colorizer{enabled: !a.opts.NoColor && isTerminal()}

	if a.opts.Help {
		a.printHelp()
		return 0
	}

	if a.opts.Version {
		fmt.Printf("imx version %s\n", version)
		return 0
	}

	// Handle stdin input
	if a.opts.Stdin {
		return a.runStdin()
	}

	if len(a.opts.Files) == 0 {
		a.printError("no input files specified")
		fmt.Fprintln(os.Stderr, "Run 'imx --help' for usage")
		return 1
	}

	// Process files
	return a.runExtract()
}

func (a *App) parseArgs(args []string) Options {
	opts := Options{Format: "text", Timeout: 30}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		// Help & Version
		case arg == "-h" || arg == "--help":
			opts.Help = true
		case arg == "-V" || arg == "--version":
			opts.Version = true

		// Output Format
		case arg == "-j" || arg == "--json":
			opts.Format = "json"
		case arg == "-t" || arg == "--table":
			opts.Format = "table"
		case arg == "--csv":
			opts.Format = "csv"
		case arg == "-S" || arg == "--summary":
			opts.Format = "summary"
		case arg == "-q" || arg == "--quiet":
			opts.Quiet = true
		case arg == "-f" || arg == "--full":
			opts.Full = true
		case arg == "--no-color":
			opts.NoColor = true

		// Filtering
		case arg == "-s" || arg == "--spec":
			i++
			if i < len(args) {
				opts.Spec = strings.ToLower(args[i])
			}
		case strings.HasPrefix(arg, "-s=") || strings.HasPrefix(arg, "--spec="):
			opts.Spec = strings.ToLower(getArgValue(arg))
		case arg == "-g" || arg == "--get":
			i++
			if i < len(args) {
				opts.Tag = args[i]
			}
		case strings.HasPrefix(arg, "-g=") || strings.HasPrefix(arg, "--get="):
			opts.Tag = getArgValue(arg)
		case arg == "--search":
			i++
			if i < len(args) {
				opts.Search = strings.ToLower(args[i])
			}
		case strings.HasPrefix(arg, "--search="):
			opts.Search = strings.ToLower(getArgValue(arg))
		case arg == "-p" || arg == "--pattern":
			i++
			if i < len(args) {
				opts.Pattern = args[i]
			}
		case strings.HasPrefix(arg, "-p=") || strings.HasPrefix(arg, "--pattern="):
			opts.Pattern = getArgValue(arg)

		// Features
		case arg == "-r" || arg == "--recursive":
			opts.Recursive = true
		case arg == "-" || arg == "--stdin":
			opts.Stdin = true
		case arg == "--timeout":
			i++
			if i < len(args) {
				fmt.Sscanf(args[i], "%d", &opts.Timeout)
			}
		case strings.HasPrefix(arg, "--timeout="):
			fmt.Sscanf(getArgValue(arg), "%d", &opts.Timeout)
		case arg == "--gps":
			i++
			if i < len(args) {
				opts.GPS = args[i]
			}
		case strings.HasPrefix(arg, "--gps="):
			opts.GPS = getArgValue(arg)
		case arg == "--stats":
			opts.Stats = true
		case arg == "-e" || arg == "--export":
			i++
			if i < len(args) {
				opts.Export = args[i]
			}
		case strings.HasPrefix(arg, "-e=") || strings.HasPrefix(arg, "--export="):
			opts.Export = getArgValue(arg)

		// Files (including URLs)
		case !strings.HasPrefix(arg, "-"):
			opts.Files = append(opts.Files, arg)

		default:
			fmt.Fprintf(os.Stderr, "Unknown option: %s\n", arg)
		}
	}

	return opts
}

func getArgValue(arg string) string {
	if idx := strings.Index(arg, "="); idx != -1 {
		return arg[idx+1:]
	}
	return ""
}

func (a *App) printHelp() {
	c := a.colors
	fmt.Printf(`%s%simx%s - Image Metadata Extractor

%s%sUSAGE%s
    imx [OPTIONS] <FILE|URL>...
    imx [OPTIONS] - < image.jpg

%s%sDESCRIPTION%s
    Extract and display metadata from image files or URLs. Supports EXIF,
    IPTC, XMP, and ICC color profiles from JPEG, PNG, WebP, TIFF, and HEIF.

%s%sINPUT%s
    <FILE>              Local file path
    <URL>               Remote image URL (http:// or https://)
    -, --stdin          Read image data from stdin
        --timeout=SEC   HTTP timeout for URLs (default: 30s)

%s%sOUTPUT FORMATS%s
    -j, --json          Output as JSON (useful for scripting)
    -t, --table         Output as aligned table
        --csv           Output as CSV (for spreadsheets)
    -S, --summary       Quick summary (camera, date, GPS, dimensions)
    -q, --quiet         Suppress headers and decorations
    -f, --full          Show full values without truncation
        --no-color      Disable colored output

%s%sFILTERING%s
    -s, --spec=SPEC     Filter by spec: exif, iptc, xmp, icc
    -g, --get=TAG       Get specific tag value (e.g., "Make" or "EXIF:Make")
        --search=TEXT   Search tags containing text
    -p, --pattern=REGEX Filter tags matching regex pattern

%s%sFEATURES%s
    -r, --recursive     Scan directories recursively
        --gps=FORMAT    GPS output: url, dms, decimal (default: dms)
        --stats         Show statistics after batch processing
    -e, --export=FMT    Export to sidecar: json, xmp

%s%sEXAMPLES%s
    # Show all metadata from a file
    %simx photo.jpg%s

    # Extract from URL
    %simx https://example.com/photo.jpg%s

    # Read from stdin (piped data)
    %scurl -s https://example.com/img.jpg | imx -%s

    # Quick summary
    %simx -S photo.jpg%s

    # Output as JSON
    %simx --json photo.jpg%s

    # Get camera make
    %simx --get=Make photo.jpg%s

    # Show only EXIF data
    %simx --spec=exif photo.jpg%s

    # Search for "date" in tags
    %simx --search=date photo.jpg%s

    # Scan directory with stats
    %simx -r --stats ./photos/%s

    # GPS as Google Maps URL
    %simx --gps=url photo.jpg%s

    # Export metadata sidecar
    %simx --export=json photo.jpg%s

`,
		c.Bold(), c.Cyan(), c.Reset(),
		c.Bold(), c.Yellow(), c.Reset(),
		c.Bold(), c.Yellow(), c.Reset(),
		c.Bold(), c.Yellow(), c.Reset(),
		c.Bold(), c.Yellow(), c.Reset(),
		c.Bold(), c.Yellow(), c.Reset(),
		c.Bold(), c.Yellow(), c.Reset(),
		c.Bold(), c.Yellow(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
		c.Dim(), c.Reset(),
	)
}

func (a *App) runStdin() int {
	// Use MetadataFromReader directly with stdin
	meta, err := a.extractor.MetadataFromReader(os.Stdin)
	if err != nil {
		a.printError(fmt.Sprintf("parsing stdin: %v", err))
		return 1
	}

	result := &ProcessResult{Meta: &meta}

	// Count and filter tags
	var tags []TagInfo
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		if a.filterTag(dir, tag) {
			tags = append(tags, TagInfo{Dir: dir, Tag: tag})
			result.TagCount++
		}
		return true
	})

	// Handle specific tag query
	if a.opts.Tag != "" {
		return a.exitCode(a.outputTag(tags))
	}

	// Output based on format
	switch a.opts.Format {
	case "json":
		return a.exitCode(a.outputJSON("<stdin>", tags))
	case "table":
		return a.exitCode(a.outputTable("<stdin>", tags))
	case "csv":
		return a.exitCode(a.outputCSV("<stdin>", tags))
	case "summary":
		return a.exitCode(a.outputSummary("<stdin>", &meta))
	default:
		return a.exitCode(a.outputText("<stdin>", tags))
	}
}

func (a *App) exitCode(err error) int {
	if err != nil {
		a.printError(err.Error())
		return 1
	}
	return 0
}

func (a *App) runExtract() int {
	files := a.expandFiles()
	if len(files) == 0 {
		a.printError("no matching files found")
		return 1
	}

	var stats Stats
	stats.Start = time.Now()
	exitCode := 0

	for i, file := range files {
		stats.Total++

		// Separator between files
		if i > 0 && a.opts.Format == "text" && !a.opts.Quiet {
			fmt.Println()
		}

		result, err := a.processFile(file)
		if err != nil {
			stats.Errors++
			if !a.opts.Quiet {
				a.printError(fmt.Sprintf("%s: %v", file, err))
			}
			exitCode = 1
			continue
		}

		stats.Success++
		stats.Tags += result.TagCount

		// Handle export (not for URLs)
		if a.opts.Export != "" && !isURL(file) {
			if err := a.exportSidecar(file, result.Meta); err != nil {
				a.printError(fmt.Sprintf("export %s: %v", file, err))
			}
		}
	}

	// Print stats if requested
	if a.opts.Stats && stats.Total > 1 {
		a.printStats(stats)
	}

	return exitCode
}

func (a *App) expandFiles() []string {
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".webp": true, ".tiff": true, ".tif": true,
		".heic": true, ".heif": true, ".avif": true,
		".cr2": true, ".nef": true, ".arw": true, ".dng": true,
	}

	var files []string
	for _, path := range a.opts.Files {
		// Handle URLs directly
		if isURL(path) {
			files = append(files, path)
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			files = append(files, path) // Let processFile handle error
			continue
		}

		if !info.IsDir() {
			files = append(files, path)
			continue
		}

		if !a.opts.Recursive {
			a.printError(fmt.Sprintf("%s is a directory (use -r to scan)", path))
			continue
		}

		filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(p))
			if imageExts[ext] {
				files = append(files, p)
			}
			return nil
		})
	}

	return files
}

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

type ProcessResult struct {
	Meta     *imx.Metadata
	TagCount int
}

func (a *App) processFile(path string) (*ProcessResult, error) {
	var m imx.Metadata
	var err error

	// Use library's built-in URL and file handling with timeout option
	timeout := imx.WithHTTPTimeout(time.Duration(a.opts.Timeout) * time.Second)

	if isURL(path) {
		m, err = a.extractor.MetadataFromURL(path, timeout)
	} else {
		m, err = a.extractor.MetadataFromFile(path)
	}
	if err != nil {
		return nil, err
	}

	meta := &m

	result := &ProcessResult{Meta: meta}

	// Count and filter tags
	var tags []TagInfo
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		if a.filterTag(dir, tag) {
			tags = append(tags, TagInfo{Dir: dir, Tag: tag})
			result.TagCount++
		}
		return true
	})

	// Handle specific tag query
	if a.opts.Tag != "" {
		return result, a.outputTag(tags)
	}

	// Output based on format
	switch a.opts.Format {
	case "json":
		return result, a.outputJSON(path, tags)
	case "table":
		return result, a.outputTable(path, tags)
	case "csv":
		return result, a.outputCSV(path, tags)
	case "summary":
		return result, a.outputSummary(path, meta)
	default:
		return result, a.outputText(path, tags)
	}
}

type TagInfo struct {
	Dir imx.Directory
	Tag imx.Tag
}

func (a *App) filterTag(dir imx.Directory, tag imx.Tag) bool {
	// Filter by spec
	if a.opts.Spec != "" && !strings.EqualFold(dir.Spec.String(), a.opts.Spec) {
		return false
	}

	// Filter by search text
	if a.opts.Search != "" {
		name := strings.ToLower(tag.Name)
		value := strings.ToLower(fmt.Sprintf("%v", tag.Value))
		if !strings.Contains(name, a.opts.Search) && !strings.Contains(value, a.opts.Search) {
			return false
		}
	}

	// Filter by regex pattern
	if a.opts.Pattern != "" {
		re, err := regexp.Compile(a.opts.Pattern)
		if err == nil {
			if !re.MatchString(tag.Name) && !re.MatchString(fmt.Sprintf("%v", tag.Value)) {
				return false
			}
		}
	}

	// Skip large binary data unless --full
	if !a.opts.Full {
		if b, ok := tag.Value.([]byte); ok && len(b) > 100 {
			return false
		}
	}

	return true
}

func (a *App) outputTag(tags []TagInfo) error {
	query := a.opts.Tag
	queryLower := strings.ToLower(query)

	for _, t := range tags {
		tagID := string(t.Tag.ID)
		nameLower := strings.ToLower(t.Tag.Name)

		// Match by name, ID, or spec:name
		if nameLower == queryLower ||
			strings.EqualFold(tagID, query) ||
			strings.HasSuffix(strings.ToLower(tagID), ":"+queryLower) {
			fmt.Println(a.formatValue(t.Tag.Value, true))
			return nil
		}
	}

	return fmt.Errorf("tag not found: %s", query)
}

func (a *App) outputJSON(path string, tags []TagInfo) error {
	result := map[string]any{"SourceFile": path}

	// Group by spec
	specs := map[string]map[string]any{}
	for _, t := range tags {
		spec := strings.ToUpper(t.Dir.Spec.String())
		if specs[spec] == nil {
			specs[spec] = map[string]any{}
		}
		specs[spec][t.Tag.Name] = a.formatJSONValue(t.Tag.Value)
	}

	for spec, data := range specs {
		result[spec] = data
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func (a *App) outputTable(path string, tags []TagInfo) error {
	c := a.colors

	if !a.opts.Quiet {
		fmt.Printf("%s%s%s\n", c.Bold(), path, c.Reset())
		fmt.Println(strings.Repeat("─", 80))
	}

	// Find column widths
	specWidth := 4
	nameWidth := 20
	for _, t := range tags {
		if len(t.Dir.Spec.String()) > specWidth {
			specWidth = len(t.Dir.Spec.String())
		}
		if len(t.Tag.Name) > nameWidth && len(t.Tag.Name) <= 30 {
			nameWidth = len(t.Tag.Name)
		}
	}

	// Header
	fmt.Printf("%s%-*s  %-*s  %s%s\n",
		c.Dim(),
		specWidth, "SPEC",
		nameWidth, "TAG",
		"VALUE",
		c.Reset())
	fmt.Println(strings.Repeat("─", 80))

	// Rows
	for _, t := range tags {
		spec := strings.ToUpper(t.Dir.Spec.String())
		name := t.Tag.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		value := a.formatValue(t.Tag.Value, false)
		if len(value) > 45 && !a.opts.Full {
			value = value[:42] + "..."
		}

		fmt.Printf("%s%-*s%s  %-*s  %s\n",
			a.specColor(t.Dir.Spec), specWidth, spec, c.Reset(),
			nameWidth, name,
			value)
	}

	return nil
}

func (a *App) outputCSV(path string, tags []TagInfo) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Header (only for first file)
	w.Write([]string{"File", "Spec", "Tag", "Value"})

	for _, t := range tags {
		w.Write([]string{
			path,
			strings.ToUpper(t.Dir.Spec.String()),
			t.Tag.Name,
			a.formatValue(t.Tag.Value, true),
		})
	}

	return nil
}

func (a *App) outputSummary(path string, meta *imx.Metadata) error {
	c := a.colors

	if !a.opts.Quiet {
		fmt.Printf("%s%s%s\n", c.Bold(), path, c.Reset())
	}

	// Camera info
	make := a.getTagValue(meta, imx.SpecEXIF, "Make")
	model := a.getTagValue(meta, imx.SpecEXIF, "Model")
	if make != "" || model != "" {
		fmt.Printf("  %sCamera:%s      %s %s\n", c.Dim(), c.Reset(), make, model)
	}

	// Date
	date := a.getTagValue(meta, imx.SpecEXIF, "DateTimeOriginal")
	if date == "" {
		date = a.getTagValue(meta, imx.SpecEXIF, "DateTime")
	}
	if date != "" {
		fmt.Printf("  %sDate:%s        %s\n", c.Dim(), c.Reset(), date)
	}

	// Dimensions
	width := a.getTagValue(meta, imx.SpecEXIF, "ImageWidth")
	height := a.getTagValue(meta, imx.SpecEXIF, "ImageHeight")
	if width == "" {
		width = a.getTagValue(meta, imx.SpecEXIF, "PixelXDimension")
	}
	if height == "" {
		height = a.getTagValue(meta, imx.SpecEXIF, "PixelYDimension")
	}
	if width != "" && height != "" {
		fmt.Printf("  %sDimensions:%s  %s × %s\n", c.Dim(), c.Reset(), width, height)
	}

	// GPS
	lat := a.getRawTagValue(meta, imx.SpecEXIF, "GPSLatitude")
	lon := a.getRawTagValue(meta, imx.SpecEXIF, "GPSLongitude")
	latRef := a.getTagValue(meta, imx.SpecEXIF, "GPSLatitudeRef")
	lonRef := a.getTagValue(meta, imx.SpecEXIF, "GPSLongitudeRef")
	if lat != nil && lon != nil {
		gpsStr := a.formatGPS(lat, lon, latRef, lonRef)
		fmt.Printf("  %sGPS:%s         %s\n", c.Dim(), c.Reset(), gpsStr)
	}

	// Exposure
	exposure := a.getTagValue(meta, imx.SpecEXIF, "ExposureTime")
	fNumber := a.getTagValue(meta, imx.SpecEXIF, "FNumber")
	iso := a.getTagValue(meta, imx.SpecEXIF, "ISOSpeedRatings")
	if exposure != "" || fNumber != "" || iso != "" {
		parts := []string{}
		if exposure != "" {
			parts = append(parts, exposure+"s")
		}
		if fNumber != "" {
			parts = append(parts, "f/"+fNumber)
		}
		if iso != "" {
			parts = append(parts, "ISO "+iso)
		}
		fmt.Printf("  %sExposure:%s    %s\n", c.Dim(), c.Reset(), strings.Join(parts, "  "))
	}

	// Lens
	lens := a.getTagValue(meta, imx.SpecEXIF, "LensModel")
	if lens == "" {
		lens = a.getTagValue(meta, imx.SpecEXIF, "Lens")
	}
	if lens != "" {
		fmt.Printf("  %sLens:%s        %s\n", c.Dim(), c.Reset(), lens)
	}

	// Copyright
	copyright := a.getTagValue(meta, imx.SpecEXIF, "Copyright")
	if copyright == "" {
		copyright = a.getTagValue(meta, imx.SpecIPTC, "CopyrightNotice")
	}
	if copyright != "" {
		fmt.Printf("  %sCopyright:%s   %s\n", c.Dim(), c.Reset(), copyright)
	}

	// Tag count by spec
	counts := map[string]int{}
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		counts[dir.Spec.String()]++
		return true
	})

	var specParts []string
	for _, spec := range []string{"exif", "iptc", "xmp", "icc"} {
		if c := counts[spec]; c > 0 {
			specParts = append(specParts, fmt.Sprintf("%s%s:%s%d", a.specColor(imx.Spec(0)), strings.ToUpper(spec), a.colors.Reset(), c))
		}
	}
	if len(specParts) > 0 {
		fmt.Printf("  %sTags:%s        %s\n", c.Dim(), c.Reset(), strings.Join(specParts, "  "))
	}

	return nil
}

func (a *App) outputText(path string, tags []TagInfo) error {
	c := a.colors

	if !a.opts.Quiet {
		fmt.Printf("%s%s%s\n", c.Bold(), path, c.Reset())
		fmt.Println(strings.Repeat("─", min(80, len(path)+10)))
	}

	// Group by spec
	type specGroup struct {
		spec imx.Spec
		dirs map[string][]TagInfo
	}
	groups := map[string]*specGroup{}
	var specOrder []string

	for _, t := range tags {
		specName := t.Dir.Spec.String()
		if groups[specName] == nil {
			groups[specName] = &specGroup{spec: t.Dir.Spec, dirs: map[string][]TagInfo{}}
			specOrder = append(specOrder, specName)
		}
		groups[specName].dirs[t.Dir.Name] = append(groups[specName].dirs[t.Dir.Name], t)
	}

	// Sort by priority
	priority := map[string]int{"exif": 0, "iptc": 1, "xmp": 2, "icc": 3}
	sort.Slice(specOrder, func(i, j int) bool {
		return priority[specOrder[i]] < priority[specOrder[j]]
	})

	// Output each spec
	for _, specName := range specOrder {
		group := groups[specName]

		// Spec header
		fmt.Printf("\n%s%s[%s]%s\n",
			a.specColor(group.spec), c.Bold(),
			strings.ToUpper(specName),
			c.Reset())

		// Get directory names sorted
		var dirNames []string
		for name := range group.dirs {
			dirNames = append(dirNames, name)
		}
		sort.Strings(dirNames)

		for _, dirName := range dirNames {
			dirTags := group.dirs[dirName]

			// Directory subheader
			fmt.Printf("%s  %s%s\n", c.Dim(), dirName, c.Reset())

			// Find max name length
			maxLen := 0
			for _, t := range dirTags {
				if len(t.Tag.Name) > maxLen && len(t.Tag.Name) <= 35 {
					maxLen = len(t.Tag.Name)
				}
			}
			if maxLen < 15 {
				maxLen = 15
			}

			// Tags
			for _, t := range dirTags {
				name := t.Tag.Name
				if len(name) > 35 {
					name = name[:32] + "..."
				}

				value := a.formatValue(t.Tag.Value, false)
				if len(value) > 55 && !a.opts.Full {
					value = value[:52] + "..."
				}

				fmt.Printf("    %s%-*s%s : %s\n",
					c.Dim(), maxLen, name, c.Reset(), value)
			}
		}
	}

	return nil
}

func (a *App) exportSidecar(path string, meta *imx.Metadata) error {
	var sidecarPath string
	var data []byte
	var err error

	switch a.opts.Export {
	case "json":
		sidecarPath = path + ".json"
		result := map[string]any{"SourceFile": path}
		specs := map[string]map[string]any{}

		meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
			spec := strings.ToUpper(dir.Spec.String())
			if specs[spec] == nil {
				specs[spec] = map[string]any{}
			}
			specs[spec][tag.Name] = a.formatJSONValue(tag.Value)
			return true
		})

		for spec, d := range specs {
			result[spec] = d
		}

		data, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}

	case "xmp":
		sidecarPath = path + ".xmp"
		// Basic XMP sidecar
		var xmpTags []string
		meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
			if dir.Spec.String() == "xmp" {
				xmpTags = append(xmpTags, fmt.Sprintf("    <%s>%v</%s>",
					tag.Name, a.formatValue(tag.Value, true), tag.Name))
			}
			return true
		})

		xmp := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about="">
%s
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>`, strings.Join(xmpTags, "\n"))
		data = []byte(xmp)

	default:
		return fmt.Errorf("unknown export format: %s (use: json, xmp)", a.opts.Export)
	}

	if err := os.WriteFile(sidecarPath, data, 0644); err != nil {
		return err
	}

	if !a.opts.Quiet {
		fmt.Printf("Exported: %s\n", sidecarPath)
	}
	return nil
}

// Helper functions

func (a *App) getTagValue(meta *imx.Metadata, spec imx.Spec, name string) string {
	var result string
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		if dir.Spec == spec && tag.Name == name {
			result = a.formatValue(tag.Value, true)
			return false
		}
		return true
	})
	return result
}

func (a *App) getRawTagValue(meta *imx.Metadata, spec imx.Spec, name string) any {
	var result any
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		if dir.Spec == spec && tag.Name == name {
			result = tag.Value
			return false
		}
		return true
	})
	return result
}

func (a *App) formatGPS(lat, lon any, latRef, lonRef string) string {
	// Try to extract decimal coordinates
	latDec := a.toDecimalDegrees(lat, latRef)
	lonDec := a.toDecimalDegrees(lon, lonRef)

	if latDec == 0 && lonDec == 0 {
		return fmt.Sprintf("%v, %v", lat, lon)
	}

	switch a.opts.GPS {
	case "url":
		return fmt.Sprintf("https://maps.google.com/maps?q=%f,%f", latDec, lonDec)
	case "decimal":
		return fmt.Sprintf("%.6f, %.6f", latDec, lonDec)
	default: // dms
		return fmt.Sprintf("%s, %s",
			a.toDMS(latDec, latRef == "S" || latDec < 0, true),
			a.toDMS(lonDec, lonRef == "W" || lonDec < 0, false))
	}
}

func (a *App) toDecimalDegrees(coord any, ref string) float64 {
	switch v := coord.(type) {
	case []float64:
		if len(v) >= 3 {
			dec := v[0] + v[1]/60 + v[2]/3600
			if ref == "S" || ref == "W" {
				dec = -dec
			}
			return dec
		}
	case float64:
		return v
	}
	return 0
}

func (a *App) toDMS(decimal float64, isNegative, isLat bool) string {
	if decimal < 0 {
		decimal = -decimal
	}

	d := int(decimal)
	m := int((decimal - float64(d)) * 60)
	s := (decimal - float64(d) - float64(m)/60) * 3600

	dir := ""
	if isLat {
		if isNegative {
			dir = "S"
		} else {
			dir = "N"
		}
	} else {
		if isNegative {
			dir = "W"
		} else {
			dir = "E"
		}
	}

	return fmt.Sprintf("%d°%d'%.2f\"%s", d, m, s, dir)
}

func (a *App) formatValue(v any, full bool) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(val)
	case []byte:
		if len(val) > 20 && !full {
			return fmt.Sprintf("(binary, %d bytes)", len(val))
		}
		return fmt.Sprintf("%x", val)
	case []any:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = a.formatValue(item, full)
		}
		return strings.Join(parts, ", ")
	case []float64:
		parts := make([]string, len(val))
		for i, f := range val {
			parts[i] = formatFloat(f)
		}
		return strings.Join(parts, ", ")
	case []uint16:
		parts := make([]string, len(val))
		for i, n := range val {
			parts[i] = fmt.Sprintf("%d", n)
		}
		return strings.Join(parts, ", ")
	case float64:
		return formatFloat(val)
	case float32:
		return formatFloat(float64(val))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatFloat(f float64) string {
	s := fmt.Sprintf("%.6f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func (a *App) formatJSONValue(v any) any {
	switch val := v.(type) {
	case []byte:
		if len(val) > 100 {
			return fmt.Sprintf("(binary, %d bytes)", len(val))
		}
		return fmt.Sprintf("%x", val)
	default:
		return v
	}
}

func (a *App) specColor(spec imx.Spec) string {
	if !a.colors.enabled {
		return ""
	}
	switch spec.String() {
	case "exif":
		return ColorGreen
	case "iptc":
		return ColorBlue
	case "xmp":
		return ColorCyan
	case "icc":
		return ColorYellow
	default:
		return ColorWhite
	}
}

func (a *App) printError(msg string) {
	fmt.Fprintf(os.Stderr, "%s%sError:%s %s\n",
		a.colors.Bold(), a.colors.Red(), a.colors.Reset(), msg)
}

func (a *App) printStats(stats Stats) {
	c := a.colors
	elapsed := time.Since(stats.Start)

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("%sStatistics:%s\n", c.Bold(), c.Reset())
	fmt.Printf("  Files:    %d total, %s%d success%s, %s%d errors%s\n",
		stats.Total,
		c.Green(), stats.Success, c.Reset(),
		c.Red(), stats.Errors, c.Reset())
	fmt.Printf("  Tags:     %d extracted\n", stats.Tags)
	fmt.Printf("  Time:     %v\n", elapsed.Round(time.Millisecond))
}

type Stats struct {
	Start   time.Time
	Total   int
	Success int
	Errors  int
	Tags    int
}

// Colorizer handles terminal colors
type Colorizer struct {
	enabled bool
}

const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func (c Colorizer) Reset() string {
	if c.enabled {
		return ColorReset
	}
	return ""
}
func (c Colorizer) Bold() string {
	if c.enabled {
		return ColorBold
	}
	return ""
}
func (c Colorizer) Dim() string {
	if c.enabled {
		return ColorDim
	}
	return ""
}
func (c Colorizer) Red() string {
	if c.enabled {
		return ColorRed
	}
	return ""
}
func (c Colorizer) Green() string {
	if c.enabled {
		return ColorGreen
	}
	return ""
}
func (c Colorizer) Yellow() string {
	if c.enabled {
		return ColorYellow
	}
	return ""
}
func (c Colorizer) Blue() string {
	if c.enabled {
		return ColorBlue
	}
	return ""
}
func (c Colorizer) Cyan() string {
	if c.enabled {
		return ColorCyan
	}
	return ""
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
