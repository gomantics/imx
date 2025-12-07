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
//	if tag, ok := meta.Tag(imx.TagMake); ok {
//		fmt.Printf("Camera: %v\n", tag.Value)
//	}
//	if tag, ok := meta.Tag(imx.TagDateTimeOriginal); ok {
//		fmt.Printf("Date: %v\n", tag.Value)
//	}
//
// For more control, use the Extractor type:
//
//	extractor := imx.New(
//		imx.WithMaxBytes(10<<20),           // Limit to 10MB
//		imx.WithBufferSize(128*1024),       // 128KB buffer
//		imx.WithStopOnFirstError(true),     // Stop on first error
//	)
//
//	meta, err := extractor.MetadataFromFile("photo.jpg")
//
// Iterate over tags:
//
//	// All tags
//	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
//		fmt.Printf("%s = %v\n", tag.Name, tag.Value)
//		return true // continue
//	})
//
//	// Tags in a specific spec
//	meta.EachInSpec(imx.SpecEXIF, func(tag imx.Tag) bool {
//		fmt.Printf("%s = %v\n", tag.Name, tag.Value)
//		return true
//	})
//
// Error handling:
//
//	meta, err := imx.MetadataFromFile("photo.jpg")
//	if err != nil {
//		// Check for partial errors
//		var partialErr *imx.PartialError
//		if errors.As(err, &partialErr) {
//			// Some parsers failed, but we got partial results
//			fmt.Printf("Got partial data with errors: %v\n", partialErr)
//			// meta still contains successfully parsed data
//		} else {
//			// Complete failure
//			log.Fatal(err)
//		}
//	}
package imx
