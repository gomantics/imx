// Example: advanced - Advanced metadata extraction techniques
//
// This example demonstrates advanced usage of the imx library including:
// - Custom extractor with options
// - Filtering by metadata spec
// - Batch processing multiple files
// - Iterating tags
// - Error handling
//
// Usage:
//
//	go run main.go <image-files...>
//	go run main.go --exif-only <image-file>
//	go run main.go --all-tags <image-file>
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomantics/imx"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	args := os.Args[1:]

	// Check for flags
	switch {
	case args[0] == "--exif-only" && len(args) > 1:
		exifOnly(args[1])
	case args[0] == "--all-tags" && len(args) > 1:
		allTags(args[1])
	case args[0] == "--batch" && len(args) > 1:
		batchProcess(args[1:])
	default:
		// Process all files
		batchProcess(args)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  advanced <image-files...>        Process multiple files")
	fmt.Println("  advanced --exif-only <file>      Extract only EXIF data")
	fmt.Println("  advanced --all-tags <file>       Show all tags with details")
	fmt.Println("  advanced --batch <files...>      Batch process with summary")
}

// exifOnly demonstrates filtering metadata to only EXIF spec
func exifOnly(filename string) {
	fmt.Printf("=== EXIF Only: %s ===\n\n", filepath.Base(filename))

	// Create extractor that only extracts EXIF
	extractor := imx.New(
		imx.WithSpecs(imx.SpecEXIF),
	)

	meta, err := extractor.MetadataFromFile(filename)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Iterate all EXIF tags
	meta.EachInSpec(imx.SpecEXIF, func(tag imx.Tag) bool {
		fmt.Printf("%-30s = %v\n", tag.Name, tag.Value)
		return true
	})

	fmt.Printf("\nTotal EXIF tags: %d\n", countTags(meta))
}

// allTags demonstrates iterating all tags with full details
func allTags(filename string) {
	fmt.Printf("=== All Tags: %s ===\n\n", filepath.Base(filename))

	meta, err := imx.MetadataFromFile(filename)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Iterate all tags across all directories
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		fmt.Printf("[%s:%s] %-25s (%s) = %v\n",
			dir.Spec, dir.Name, tag.Name, tag.DataType, tag.Value)
		return true
	})

	fmt.Printf("\nTotal: %d directories, %d tags\n",
		len(meta.Directories), countTags(meta))
}

// batchProcess demonstrates processing multiple files
func batchProcess(files []string) {
	fmt.Println("=== Batch Processing ===")

	// Create a reusable extractor (safe for concurrent use)
	extractor := imx.New(
		imx.WithMaxBytes(50 << 20), // 50MB limit per file
	)

	var (
		processed int
		failed    int
		totalTags int
	)

	for _, pattern := range files {
		// Expand glob patterns
		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Printf("Invalid pattern %s: %v", pattern, err)
			continue
		}

		if len(matches) == 0 {
			// Try as literal path
			matches = []string{pattern}
		}

		for _, filename := range matches {
			// Skip non-image files
			ext := strings.ToLower(filepath.Ext(filename))
			if ext != ".jpg" && ext != ".jpeg" {
				continue
			}

			meta, err := extractor.MetadataFromFile(filename)
			if err != nil {
				fmt.Printf("✗ %s: %v\n", filepath.Base(filename), err)
				failed++
				continue
			}

			tags := countTags(meta)
			totalTags += tags

			// Print summary for each file
			make := getTagValue(meta, imx.TagMake)
			model := getTagValue(meta, imx.TagModel)
			date := getTagValue(meta, imx.TagDateTimeOriginal)

			fmt.Printf("✓ %s\n", filepath.Base(filename))
			fmt.Printf("  Camera: %s %s\n", make, model)
			fmt.Printf("  Date: %s\n", date)
			fmt.Printf("  Tags: %d\n\n", tags)

			processed++
		}
	}

	// Print summary
	fmt.Println("=== Summary ===")
	fmt.Printf("Processed: %d files\n", processed)
	fmt.Printf("Failed: %d files\n", failed)
	fmt.Printf("Total tags: %d\n", totalTags)
}

// Helper functions

func countTags(meta imx.Metadata) int {
	count := 0
	for _, dir := range meta.Directories {
		count += len(dir.Tags)
	}
	return count
}

func getTagValue(meta imx.Metadata, tagID imx.TagID) string {
	if tag, ok := meta.Tag(imx.SpecEXIF, tagID); ok {
		return fmt.Sprintf("%v", tag.Value)
	}
	return "(none)"
}
