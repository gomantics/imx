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

	// Extract metadata
	meta, err := imx.ExtractFromFile(filename)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Print common EXIF tags using tag constants
	fmt.Println("=== Common EXIF Tags ===")
	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagMake); ok {
		fmt.Printf("Make: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagModel); ok {
		fmt.Printf("Model: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagDateTimeOriginal); ok {
		fmt.Printf("DateTime: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagOrientation); ok {
		fmt.Printf("Orientation: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagISO); ok {
		fmt.Printf("ISO: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagGPSLatitude); ok {
		fmt.Printf("GPS Latitude: %v\n", tag.Value)
	}
	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagGPSLongitude); ok {
		fmt.Printf("GPS Longitude: %v\n", tag.Value)
	}

	// Print all EXIF tags using iterator
	fmt.Println("\n=== All EXIF Tags ===")
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		fmt.Printf("%s:%s = %v (%s)\n", dir.Name, tag.Name, tag.Value, tag.DataType)
		return true
	})

	// Print directories
	fmt.Println("\n=== Directories ===")
	for _, dir := range meta.Directories {
		fmt.Printf("%s:%s - %d tags\n", dir.Spec, dir.Name, len(dir.Tags))
	}
}
