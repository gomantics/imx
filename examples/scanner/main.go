package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomantics/imx"
)

type imageSummary struct {
	path     string
	make     string
	model    string
	datetime string
	width    int
	height   int
	hasGPS   bool
	lat      interface{}
	lon      interface{}
	tagCount int
	err      error
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: scanner <directory>")
		fmt.Println("\nRecursively scans directory for images and extracts metadata.")
		os.Exit(1)
	}

	dirPath := os.Args[1]

	// Verify directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		log.Fatalf("Error accessing directory: %v", err)
	}
	if !info.IsDir() {
		log.Fatalf("Path is not a directory: %s", dirPath)
	}

	fmt.Printf("Scanning directory: %s\n\n", dirPath)

	var summaries []imageSummary
	fileCount := 0
	successCount := 0
	errorCount := 0

	// Walk directory tree
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Warning: cannot access %s: %v\n", path, err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		fileCount++
		summary := extractMetadata(path)
		summaries = append(summaries, summary)

		if summary.err != nil {
			errorCount++
		} else {
			successCount++
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}

	// Print summaries
	fmt.Printf("Found %d files\n", fileCount)
	fmt.Printf("Successfully extracted metadata from %d files\n", successCount)
	if errorCount > 0 {
		fmt.Printf("Failed to extract metadata from %d files\n", errorCount)
	}
	fmt.Println(strings.Repeat("=", 80))

	for i, summary := range summaries {
		fmt.Printf("\n[%d] %s\n", i+1, summary.path)

		if summary.err != nil {
			fmt.Printf("    Error: %v\n", summary.err)
			continue
		}

		if summary.make != "" {
			fmt.Printf("    Camera: %s %s\n", summary.make, summary.model)
		}
		if summary.datetime != "" {
			fmt.Printf("    Date: %s\n", summary.datetime)
		}
		if summary.width > 0 && summary.height > 0 {
			fmt.Printf("    Dimensions: %dx%d\n", summary.width, summary.height)
		}
		if summary.hasGPS {
			fmt.Printf("    GPS: %v, %v\n", summary.lat, summary.lon)
		}
		fmt.Printf("    Metadata Tags: %d\n", summary.tagCount)
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("Scan complete: %d files processed\n", fileCount)
}

func extractMetadata(path string) imageSummary {
	summary := imageSummary{path: path}

	// Extract metadata
	meta, err := imx.ExtractFromFile(path)
	if err != nil {
		summary.err = err
		return summary
	}

	// Count total tags
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		summary.tagCount++
		return true
	})

	// Extract common fields
	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagMake); ok {
		if make, ok := tag.Value.(string); ok {
			summary.make = strings.TrimSpace(make)
		}
	}

	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagModel); ok {
		if model, ok := tag.Value.(string); ok {
			summary.model = strings.TrimSpace(model)
		}
	}

	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagDateTimeOriginal); ok {
		if dt, ok := tag.Value.(string); ok {
			summary.datetime = dt
		}
	}

	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagImageWidth); ok {
		if w, ok := tag.Value.(int); ok {
			summary.width = w
		} else if w, ok := tag.Value.(uint32); ok {
			summary.width = int(w)
		}
	}

	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagImageHeight); ok {
		if h, ok := tag.Value.(int); ok {
			summary.height = h
		} else if h, ok := tag.Value.(uint32); ok {
			summary.height = int(h)
		}
	}

	// Check for GPS data
	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagGPSLatitude); ok {
		summary.hasGPS = true
		summary.lat = tag.Value
	}

	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagGPSLongitude); ok {
		summary.lon = tag.Value
	}

	return summary
}
