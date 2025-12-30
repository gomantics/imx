// Package imx provides fast, dependency-free extraction of image metadata.
//
// It supports EXIF, IPTC, XMP, and ICC metadata from JPEG, PNG, GIF, WebP,
// TIFF-based formats (including CR2/DNG), HEIC, plus ID3/FLAC/MP4 audio/video tags.
//
// Version: 1.0.0
//
// Basic usage:
//
//	meta, err := imx.MetadataFromFile("photo.jpg")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Access tags by ID
//	if tag, ok := meta.Tag("EXIF:IFD0:Make"); ok {
//		fmt.Printf("Camera: %v\n", tag.Value)
//	}
//
//	// Or use type-safe getters
//	make, err := meta.GetString("EXIF:IFD0:Make")
//	if err == nil {
//		fmt.Printf("Camera: %s\n", make)
//	}
//
// For more control, use the Extractor type:
//
//	extractor := imx.New(
//		imx.WithHTTPTimeout(60 * time.Second), // Set HTTP timeout for URL fetching
//	)
//
//	meta, err := extractor.MetadataFromFile("photo.jpg")
//
// Iterate over tags:
//
//	// All tags across all directories
//	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
//		fmt.Printf("[%s] %s = %v\n", dir.Name, tag.Name, tag.Value)
//		return true // continue
//	})
//
//	// Tags in a specific directory
//	meta.EachInDirectory("IFD0", func(tag imx.Tag) bool {
//		fmt.Printf("%s = %v\n", tag.Name, tag.Value)
//		return true
//	})
//
// Error handling:
//
//	meta, err := imx.MetadataFromFile("photo.jpg")
//	if err != nil {
//		if errors.Is(err, imx.ErrUnknownFormat) {
//			fmt.Println("Unsupported file format")
//		} else {
//			log.Fatal(err)
//		}
//	}
//
//	// Check for parsing errors
//	if len(meta.Errors()) > 0 {
//		fmt.Printf("Parsing errors: %v\n", meta.Errors())
//		// meta still contains successfully parsed data
//	}
//
// Multiple input sources:
//
//	// From file with safety limit
//	meta, err := imx.MetadataFromFile("photo.jpg", imx.WithMaxBytes(50<<20))
//
//	// From byte slice
//	data, _ := os.ReadFile("photo.jpg")
//	meta, err = imx.MetadataFromBytes(data)
//
//	// From io.Reader (buffered on-demand)
//	file, _ := os.Open("photo.jpg")
//	meta, err = imx.MetadataFromReader(file)
//
//	// From URL
//	meta, err = imx.MetadataFromURL("https://example.com/photo.jpg")
package imx

// Version is the semantic version of the imx package
const Version = "1.0.0"
