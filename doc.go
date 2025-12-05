// Package imx provides fast, dependency-free extraction of image metadata.
//
// It supports EXIF, IPTC, XMP, and ICC metadata from JPEG, PNG, WebP, TIFF,
// HEIF/HEIC, and AVIF formats.
//
// Basic usage:
//
//	meta, err := imx.MetadataFromFile("photo.jpg")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Taken at: %v\n", meta.DateTimeOriginal())
//	fmt.Printf("Camera: %s %s\n", meta.Make(), meta.Model())
//
// For more control, use the Extractor type:
//
//	extractor := imx.New(
//		imx.WithNamespaces(imx.NamespaceEXIF),
//		imx.WithMaxBytes(10 << 20),
//	)
//
//	meta, err := extractor.MetadataFromFile("photo.jpg")
package imx
