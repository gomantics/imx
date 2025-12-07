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
BenchmarkMetadataFromFile-12         4308      279μs/op     583KB/op     3616 allocs/op
BenchmarkMetadataFromBytes-12        4707      250μs/op     583KB/op     3614 allocs/op
BenchmarkMetadata_Tag-12        168282322        7ns/op        0B/op        0 allocs/op
BenchmarkParser_EXIF-12           154510       6.9μs/op      16KB/op      219 allocs/op
BenchmarkParser_IPTC-12           729804       1.6μs/op     4.6KB/op       54 allocs/op
BenchmarkParser_XMP-12             47097        25μs/op      34KB/op      462 allocs/op
BenchmarkParser_ICC-12         263156931       4.6ns/op        0B/op        0 allocs/op
BenchmarkParser_JPEG-12           481448       2.5μs/op      48KB/op       24 allocs/op
```

**Continuous Benchmarking**:
- View [historical performance graphs](../../benchmarks) (auto-updated on every commit)
- Automatic regression detection with 120% threshold
- Performance alerts on PRs

**Run Benchmarks Locally**:
```bash
make bench              # Quick benchmark run
make bench-all          # Detailed output with 3s timing
make bench-report       # Generate formatted report
make bench-compare OLD=commit1 NEW=commit2  # Compare commits
make bench-viz N=50     # Generate performance graphs across last 50 commits
```

**Performance Visualization**:

Generate performance graphs showing how benchmarks evolved over time:

```bash
# Generate graphs for last 50 commits
make bench-viz N=50
```

This creates visualizations in `out/` directory showing:
- **ns/op** - Nanoseconds per operation over time
- **B/op** - Bytes allocated per operation over time
- **allocs/op** - Number of allocations per operation over time

Example graphs (run `make bench-viz` to generate):

![High-Level API Performance](out/summary_HighLevelAPI.png)
![Parser Performance](out/summary_ParserEXIF.png)

**Requirements**: Python 3 with matplotlib (`pip3 install matplotlib`)

**Tools**:
- Results tracked via [github-action-benchmark](https://github.com/benchmark-action/github-action-benchmark)
- Statistical comparison using [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- Historical visualization using Python + matplotlib

## Contributing

Contributions are welcome!

- **Development guidelines**: See [CONTRIBUTING.md](CONTRIBUTING.md)
- **AI assistance setup**: See [CLAUDE.md](CLAUDE.md)

Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting PRs - it contains important information about:
- Commit message format (serves as changelog)
- Benchmarking guidelines
- Code style and testing requirements

## License

MIT License - see [LICENSE](LICENSE) for details.

