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

	// Print using convenience helpers
	fmt.Println("=== Convenience Helpers ===")
	fmt.Printf("Make: %s\n", meta.Make())
	fmt.Printf("Model: %s\n", meta.Model())
	fmt.Printf("DateTime: %v\n", meta.DateTimeOriginal())
	fmt.Printf("Orientation: %d\n", meta.Orientation())
	fmt.Printf("ISO: %d\n", meta.ISO())

	if gps := meta.GPSCoordinates(); gps != nil {
		fmt.Printf("GPS: %.6f, %.6f (altitude: %.2fm)\n", gps.Lat, gps.Lon, gps.Altitude)
	}

	// Print all EXIF tags using iterator
	fmt.Println("\n=== All EXIF Tags ===")
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		fmt.Printf("%s:%s = %v (%s)\n", dir.Name, tag.Name, tag.Value, tag.Type)
		return true
	})

	// Print directories
	fmt.Println("\n=== Directories ===")
	for _, dir := range meta.Directories {
		fmt.Printf("%s:%s - %d tags\n", dir.Namespace, dir.Name, len(dir.Tags))
	}
}
