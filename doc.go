// Package imx provides fast, dependency-free extraction of image metadata.
//
// It supports EXIF, IPTC, XMP, and ICC metadata from JPEG images
// (with more formats coming soon).
//
// Basic usage:
//
//	meta, err := imx.MetadataFromFile("photo.jpg")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Access tags using constants
//	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagMake); ok {
//		fmt.Printf("Camera: %v\n", tag.Value)
//	}
//	if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagDateTimeOriginal); ok {
//		fmt.Printf("Date: %v\n", tag.Value)
//	}
//
// For more control, use the Extractor type:
//
//	extractor := imx.New(
//		imx.WithSpecs(imx.SpecEXIF),
//		imx.WithMaxBytes(10 << 20),
//	)
//
//	meta, err := extractor.MetadataFromFile("photo.jpg")
package imx
