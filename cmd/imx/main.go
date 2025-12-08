// imx - Image Metadata Extractor CLI
//
// A powerful command-line tool for extracting, querying, and analyzing
// metadata from images. Supports EXIF, IPTC, XMP, and ICC color profiles.
package main

import (
	"os"
)

func main() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}
