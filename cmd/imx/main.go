// imx - Image Metadata Extractor CLI
//
// A powerful command-line tool for extracting, querying, and analyzing
// metadata from images. Supports EXIF, IPTC, XMP, and ICC color profiles.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
