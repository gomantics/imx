# imx - Image Metadata Extraction Library

A Go library for detecting image formats and extracting comprehensive metadata from image files.

## Features

- **Format Detection**: Automatically detects image format by reading magic bytes
- **Comprehensive Metadata**: Extracts dimensions, color depth, color space, EXIF data, and ICC profiles
- **Multiple Formats**: Supports JPEG, PNG, GIF, WebP, and BMP
- **Zero Dependencies**: Uses only the Go standard library
- **Type-Safe**: Well-defined struct types for metadata

## Installation

```bash
go get imx
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "imx"
)

func main() {
    md, err := imx.Metadata("image.jpg")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Format: %s\n", md.Format)
    fmt.Printf("Dimensions: %dx%d\n", md.Width, md.Height)
    fmt.Printf("File Size: %d bytes\n", md.FileSize)
    fmt.Printf("Color Depth: %d bits\n", md.ColorDepth)
    fmt.Printf("Color Space: %s\n", md.ColorSpace)
    fmt.Printf("Has ICC Profile: %v\n", md.HasICCProfile)
    
    // Print EXIF data
    if len(md.EXIF) > 0 {
        fmt.Println("\nEXIF Data:")
        for key, value := range md.EXIF {
            fmt.Printf("  %s: %v\n", key, value)
        }
    }
    
    // Print additional format-specific metadata
    if len(md.Additional) > 0 {
        fmt.Println("\nAdditional Metadata:")
        for key, value := range md.Additional {
            fmt.Printf("  %s: %v\n", key, value)
        }
    }
}
```

### Reusing Existing Readers

When an image is already available as an `io.ReadSeeker`, you can avoid another
open/stat/seek cycle by calling `MetadataFromReader` directly:

```go
f, err := os.Open("image.jpg")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

info, _ := f.Stat()
md, err := imx.MetadataFromReader(f, info.Size())
```

### ImageMetadata Structure

The `ImageMetadata` struct contains the following fields:

```go
type ImageMetadata struct {
    Format        string                 // Image format (JPEG, PNG, GIF, WebP, BMP)
    Width         int                    // Image width in pixels
    Height        int                    // Image height in pixels
    FileSize      int64                  // File size in bytes
    ColorDepth    int                    // Color depth in bits
    ColorSpace    string                 // Color space (RGB, RGBA, CMYK, Grayscale, etc.)
    HasICCProfile bool                   // Whether the image contains an ICC profile
    EXIF          map[string]interface{} // EXIF metadata tags
    Additional    map[string]interface{} // Format-specific additional metadata
}
```

### Supported Formats

#### JPEG
- Dimensions from SOF segments
- Color space detection (RGB, Grayscale, CMYK)
- EXIF data extraction from APP1 segments
- ICC profile detection from APP2 segments
- Additional metadata: bits per sample, components

#### PNG
- Dimensions from IHDR chunk
- Bit depth and color type
- Color space detection
- EXIF data from eXIf chunk
- ICC profile detection from iCCP chunk
- Additional metadata: compression method, filter method, interlace

#### GIF
- Dimensions from Logical Screen Descriptor
- Color table information
- Animation detection
- Transparency detection
- Additional metadata: version, color resolution, frame count

#### WebP
- Dimensions from VP8/VP8L/VP8X chunks
- Animation detection
- Alpha channel detection
- ICC profile detection
- Additional metadata: format variant, flags

#### BMP
- Dimensions from DIB header
- Bit depth and color space
- Compression type
- Additional metadata: planes, resolution, color table info

### EXIF Data

The library extracts common EXIF tags including:
- `DateTime`: Image creation date/time
- `Make`: Camera manufacturer
- `Model`: Camera model
- `Orientation`: Image orientation
- `Software`: Software used to create the image
- `Artist`: Artist/photographer name
- `Copyright`: Copyright information
- And more...

### Error Handling

The library returns descriptive errors for:
- File not found
- Unsupported image formats
- Corrupted or invalid image files
- I/O errors

Example error handling:

```go
md, err := imx.Metadata("image.jpg")
if err != nil {
    if os.IsNotExist(err) {
        log.Fatal("File does not exist")
    }
    log.Fatalf("Failed to extract metadata: %v", err)
}
```

## Performance Notes

- Metadata maps are allocated lazily, so parsing lightweight assets avoids
  unnecessary garbage collection work.
- Large JPEG/PNG chunk buffers are recycled through a shared `sync.Pool`,
  reducing per-file allocations when scanning directories.
- `MetadataFromReader` lets you reuse an existing `io.ReadSeeker` to skip extra
  `open/stat/seek` operations.
- Benchmarks (e.g. `BenchmarkMetadataJPEG` in `metadata_test.go`) can be run via
  `go test -bench=. ./...` to validate improvements on your machine.

## Testing

Run tests with:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## Examples

A runnable CLI example is provided under `examples/print-metadata`. It accepts a
path to an image file and prints all metadata extracted by the `imx` library.

```bash
cd examples/print-metadata
go run . /path/to/image.jpg
```

## License

This library is provided as-is for use in your projects.

## Contributing

Contributions are welcome! Please ensure that:
- Code follows Go best practices
- Tests are included for new features
- Documentation is updated

## Implementation Details

The library works by:
1. Reading the first few bytes of the file to detect the format via magic numbers
2. Parsing format-specific headers and chunks to extract metadata
3. Extracting EXIF data from JPEG APP1 segments or PNG eXIf chunks
4. Detecting ICC profiles in format-specific locations
5. Returning a comprehensive `ImageMetadata` struct with all extracted information

All format parsers are located in the `formats/` subdirectory for better code organization.

