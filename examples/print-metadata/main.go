package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"imx"
)

// This example CLI reads a file path argument, extracts metadata using the
// imx.Metadata API, and prints a human-friendly summary plus the raw metadata
// maps in JSON form.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <image-path>\n", os.Args[0])
		os.Exit(1)
	}

	filepath := os.Args[1]
	md, err := imx.Metadata(filepath)
	if err != nil {
		log.Fatalf("failed to read metadata: %v", err)
	}

	// Print core metadata fields.
	fmt.Printf("File: %s\n", filepath)
	fmt.Printf("Format: %s\n", md.Format)
	fmt.Printf("Dimensions: %dx%d\n", md.Width, md.Height)
	fmt.Printf("File Size: %d bytes\n", md.FileSize)
	fmt.Printf("Color Depth: %d-bit\n", md.ColorDepth)
	fmt.Printf("Color Space: %s\n", md.ColorSpace)
	fmt.Printf("Has ICC Profile: %t\n", md.HasICCProfile)

	// Helper to pretty-print maps if data exists.
	printMap := func(title string, data map[string]interface{}) {
		if len(data) == 0 {
			return
		}

		blob, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Printf("%s: <error encoding JSON: %v>\n", title, err)
			return
		}

		fmt.Printf("\n%s:\n%s\n", title, string(blob))
	}

	printMap("EXIF Data", md.EXIF)
	printMap("Additional Metadata", md.Additional)
}


