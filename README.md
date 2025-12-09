# imx

[![Go Reference](https://pkg.go.dev/badge/github.com/gomantics/imx.svg)](https://pkg.go.dev/github.com/gomantics/imx)
[![CI](https://github.com/gomantics/imx/actions/workflows/ci.yml/badge.svg)](https://github.com/gomantics/imx/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/gomantics/imx/branch/main/graph/badge.svg?token=1P4DWIL8JW)](https://codecov.io/gh/gomantics/imx)
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
    if tag, ok := meta.Tag(imx.TagMake); ok {
        fmt.Printf("Camera: %v\n", tag.Value)
    }
    if tag, ok := meta.Tag(imx.TagModel); ok {
        fmt.Printf("Model: %v\n", tag.Value)
    }
    if tag, ok := meta.Tag(imx.TagDateTimeOriginal); ok {
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
meta, err := imx.MetadataFromFile("photo.jpg",
    imx.WithMaxBytes(5<<20),     // Limit to 5MB
    imx.WithBufferSize(64*1024), // 64KB buffer
)
```

### Using the Extractor

For more control or when processing many files, create a reusable `Extractor`.

```go
extractor := imx.New(
    imx.WithMaxBytes(10<<20),              // Limit to 10MB
    imx.WithBufferSize(128*1024),          // 128KB buffer
    imx.WithStopOnFirstError(true),        // Stop on first parser error
    imx.WithHTTPTimeout(30*time.Second),   // HTTP timeout for URLs
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

### Benchmarks

**Latest Results** *(Apple M4 Pro, Go 1.23)*:

```
High-Level API
BenchmarkMetadataFromFile-12      12282     195814 ns/op   446581 B/op     2390 allocs/op
BenchmarkMetadataFromBytes-12     14323     167753 ns/op   446282 B/op     2388 allocs/op
BenchmarkMetadataFromReader-12    14336     168602 ns/op   446283 B/op     2388 allocs/op
BenchmarkMetadata_Tag-12       321816780      7.207 ns/op        0 B/op        0 allocs/op
BenchmarkMetadata_GetAll-12     35377401      68.13 ns/op        0 B/op        0 allocs/op
BenchmarkMetadata_Each-12        1884883       1294 ns/op        0 B/op        0 allocs/op

Parser Benchmarks
BenchmarkEXIFParse-12            2799916       870.9 ns/op     1552 B/op       26 allocs/op
BenchmarkIPTCParse-12            1767890       1325 ns/op     4506 B/op       46 allocs/op
BenchmarkXMPParse-12              119239      18799 ns/op    21888 B/op      355 allocs/op
BenchmarkICCParse-12           502302393      4.574 ns/op        0 B/op        0 allocs/op
BenchmarkJPEGParse-12             764920       2631 ns/op    47920 B/op       24 allocs/op
```

**Continuous Benchmarking**: Performance is automatically tracked on every commit to main. View historical trends and charts at:
- 📊 **[Performance Dashboard](https://gomantics.github.io/imx/dev/bench/)**

**Run locally**: `make bench`

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed benchmarking information

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines

## License

MIT License - see [LICENSE](LICENSE) for details.

