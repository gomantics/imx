package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"imx"
)

// This example CLI reads a file path argument, extracts metadata using the
// imx.MetadataFromFile API, and prints a human-friendly summary plus the raw metadata
// maps in JSON form.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <image-path>\n", os.Args[0])
		os.Exit(1)
	}

	filepath := os.Args[1]
	md, err := imx.MetadataFromURL("https://substackcdn.com/image/fetch/$s_!jKmW!,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fbucketeer-e05bbc84-baa3-437e-9518-adb32be77984.s3.amazonaws.com%2Fpublic%2Fimages%2Ff6c16ac1-bd59-4746-a705-457ffb913af5_896x552.png")
	if err != nil {
		log.Fatalf("failed to read metadata: %v", err)
	}

	// Print core metadata fields.
	fmt.Printf("File: %s\n", filepath)
	fmt.Printf("Format: %s\n", md.Format)
	fmt.Printf("Dimensions: %dx%d\n", md.Width, md.Height)
	fmt.Printf("File Size: %f mb\n", float64(md.FileSize)/1024.0/1024.0)
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
