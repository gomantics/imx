// Example: basic - Simple metadata extraction
//
// This example demonstrates basic usage of the imx library to extract
// and display metadata from an image file.
//
// Usage:
//
//	go run main.go <image-file>
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gomantics/imx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: basic <image-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Extract metadata using the convenience function
	meta, err := imx.MetadataFromFile(filename)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Print common EXIF tags using tag constants
	fmt.Println("=== Common EXIF Tags ===")

	if tag, ok := meta.Tag(imx.TagMake); ok {
		fmt.Printf("Make: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.TagModel); ok {
		fmt.Printf("Model: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.TagDateTimeOriginal); ok {
		fmt.Printf("Date: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.TagOrientation); ok {
		fmt.Printf("Orientation: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.TagISO); ok {
		fmt.Printf("ISO: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.TagExposureTime); ok {
		fmt.Printf("Exposure: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.TagFNumber); ok {
		fmt.Printf("Aperture: f/%v\n", tag.Value)
	}

	// Print GPS coordinates if available
	fmt.Println("\n=== GPS ===")
	if tag, ok := meta.Tag(imx.TagGPSLatitude); ok {
		fmt.Printf("Latitude: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.TagGPSLongitude); ok {
		fmt.Printf("Longitude: %v\n", tag.Value)
	}

	// Print summary of all directories
	fmt.Println("\n=== Directories ===")
	for _, dir := range meta.Directories {
		fmt.Printf("%s:%s - %d tags\n", dir.Spec, dir.Name, len(dir.Tags))
	}
}
