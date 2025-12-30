package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gomantics/imx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <image-file>\n", os.Args[0])
		os.Exit(1)
	}

	file := os.Args[1]

	// Extract metadata from file
	meta, err := imx.MetadataFromFile(file)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	// Print all directories and tags
	for _, dir := range meta.Directories() {
		for _, tag := range dir.Tags {
			fmt.Printf("[%s] %s = %v (%s)\n", dir.Name, tag.Name, tag.Value, tag.DataType)
		}
	}

	// Print any errors that occurred during parsing
	if len(meta.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarnings/Errors:\n")
		for _, err := range meta.Errors() {
			fmt.Fprintf(os.Stderr, "  - %v\n", err)
		}
	}
}
