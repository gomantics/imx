# imx

[![Go Reference](https://pkg.go.dev/badge/github.com/gomantics/imx.svg)](https://pkg.go.dev/github.com/gomantics/imx)
[![CI](https://github.com/gomantics/imx/actions/workflows/ci.yml/badge.svg)](https://github.com/gomantics/imx/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomantics/imx)](https://goreportcard.com/report/github.com/gomantics/imx)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Fast, dependency-free image metadata extraction for Go. Extract EXIF, IPTC, XMP, and ICC color profile data from images.

## Features

- **Zero dependencies** - Pure Go, stdlib only, no CGO
- **Streaming I/O** - Memory efficient, never loads entire files
- **Multiple formats** - JPEG (more formats coming soon)
- **Multiple metadata types** - EXIF, IPTC, XMP, ICC profiles

## Installation

```bash
go get github.com/gomantics/imx
```

### CLI Tool

```bash
go install github.com/gomantics/imx/cmd/imx@latest
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/gomantics/imx"
)

func main() {
    // Extract metadata from a file
    meta, err := imx.MetadataFromFile("photo.jpg")
    if err != nil {
        log.Fatal(err)
    }

    // Access common EXIF tags using constants
    if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagMake); ok {
        fmt.Printf("Camera: %v\n", tag.Value)
    }
    if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagModel); ok {
        fmt.Printf("Model: %v\n", tag.Value)
    }
    if tag, ok := meta.Tag(imx.SpecEXIF, imx.TagDateTimeOriginal); ok {
        fmt.Printf("Date: %v\n", tag.Value)
    }
}
```

## API Overview

### Convenience Functions

These package-level functions use a shared default extractor and are safe for concurrent use. All functions accept optional configuration:

```go
// From file path
meta, err := imx.MetadataFromFile("photo.jpg")

// From io.Reader
meta, err := imx.MetadataFromReader(reader)

// From byte slice
meta, err := imx.MetadataFromBytes(data)

// From URL
meta, err := imx.MetadataFromURL("https://example.com/photo.jpg")

// With options
meta, err := imx.MetadataFromFile("photo.jpg", imx.WithSpecs(imx.SpecEXIF))
```

### Using the Extractor

For more control or when processing many files, create a reusable `Extractor`. Once created, an Extractor is safe for concurrent use across goroutines:

```go
extractor := imx.New(
    imx.WithSpecs(imx.SpecEXIF, imx.SpecXMP),  // Only extract EXIF and XMP
    imx.WithMaxBytes(10 << 20),                 // Limit to 10MB
)

meta, err := extractor.MetadataFromFile("photo.jpg")
```

### Iterating Tags

```go
// Iterate all tags
meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
    fmt.Printf("%s:%s = %v\n", dir.Name, tag.Name, tag.Value)
    return true // continue iteration
})

// Iterate tags in a specific spec
meta.EachInSpec(imx.SpecEXIF, func(tag imx.Tag) bool {
    fmt.Printf("%s = %v\n", tag.Name, tag.Value)
    return true
})
```

### Batch Retrieval

```go
// Get multiple tags at once
values := meta.GetAll(imx.TagMake, imx.TagModel, imx.TagISO)
for id, value := range values {
    fmt.Printf("%s: %v\n", id, value)
}
```

## Supported Metadata

| Spec | Description | Status |
|------|-------------|--------|
| EXIF | Exchangeable Image File Format - camera settings, GPS coordinates, timestamps, and device information embedded by cameras and smartphones | ✅ Full support |
| IPTC | International Press Telecommunications Council - industry standard for news and media metadata including captions, credits, and keywords | ✅ Full support |
| XMP | Extensible Metadata Platform - Adobe's XML-based format for extensible metadata used by creative applications | ✅ Full support |
| ICC | International Color Consortium - color profile data describing color space characteristics for accurate color reproduction | ✅ Full support |

## Supported Formats

| Format | Status |
|--------|--------|
| JPEG | ✅ Full support |
| PNG | 🔜 Planned |
| WebP | 🔜 Planned |
| TIFF | 🔜 Planned |
| HEIF/HEIC | 🔜 Planned |
| AVIF | 🔜 Planned |

## CLI Usage

```bash
# Basic extraction
imx photo.jpg

# JSON output
imx --format json photo.jpg

# Filter by spec
imx --spec exif photo.jpg

# Get specific tag
imx --tag Make photo.jpg

# Process multiple files
imx --recursive ./photos/

# Read from stdin
cat photo.jpg | imx --stdin
```

## Performance

imx is designed for high performance:

- **Streaming** - Uses `bufio.Reader`, never loads entire files into memory
- **Minimal allocations** - Reuses buffers where possible
- **Early termination** - Stops parsing after metadata segments
- **Concurrent safe** - Extractor can be shared across goroutines

## Contributing

Contributions are welcome! Please see [CLAUDE.md](CLAUDE.md) for development guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

