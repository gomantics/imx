# imx

[![Go Reference](https://pkg.go.dev/badge/github.com/gomantics/imx.svg)](https://pkg.go.dev/github.com/gomantics/imx)
[![CI](https://github.com/gomantics/imx/actions/workflows/ci.yml/badge.svg)](https://github.com/gomantics/imx/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/gomantics/imx/branch/main/graph/badge.svg?token=1P4DWIL8JW)](https://codecov.io/gh/gomantics/imx)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomantics/imx)](https://goreportcard.com/report/github.com/gomantics/imx)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Fast, dependency-free metadata extraction for images, audio, and video files in Go.

## Features

- **Zero dependencies** - Pure Go, stdlib only, no CGO
- **20+ formats** - JPEG, PNG, GIF, WebP, TIFF, HEIC, CR2, DNG, NEF, ARW, ORF, RAF, RW2, PEF, SRW, 3FR, MP3, FLAC, MP4, M4A
- **Multiple metadata types** - EXIF, IPTC, XMP, ICC profiles, ID3 tags, FLAC metadata
- **Streaming I/O** - Memory efficient using `io.ReaderAt`, never loads entire files
- **Safety limits** - Configurable max-bytes (default 50MB) and buffering controls to prevent unbounded reads
- **Well-tested** - Extensive unit, fuzz, and benchmark coverage across parsers
- **Thread-safe** - Stateless parsers safe for concurrent use

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
    imx.WithMaxBytes(5<<20),     // Limit total bytes (default 50MB)
    imx.WithBufferSize(64*1024), // 64KB buffer (default)
)
// Exceeding MaxBytes returns imx.ErrMaxBytesExceeded
```

### Using the Extractor

```go
extractor := imx.New(
    imx.WithMaxBytes(10<<20),              // Limit to 10MB
    imx.WithBufferSize(128*1024),          // 128KB buffer
    imx.WithHTTPTimeout(30*time.Second),   // HTTP timeout for URLs
)

meta, err := extractor.MetadataFromFile("photo.jpg")
// Default safety: 50MB max bytes; configurable via WithMaxBytes
```

### Iterating Tags

```go
// Iterate all tags
meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
    fmt.Printf("%s:%s = %v\n", dir.Name, tag.Name, tag.Value)
    return true // continue iteration
})

// Iterate tags in a specific directory
meta.EachInDirectory("IFD0", func(tag imx.Tag) bool {
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

| Type | Description |
|------|-------------|
| EXIF | Camera settings, GPS coordinates, timestamps, device information |
| IPTC | News and media metadata including captions, credits, keywords |
| XMP | Adobe's XML-based extensible metadata |
| ICC | Color profile data for accurate color reproduction |
| ID3 | Audio metadata for MP3 files (v2.2, v2.3, v2.4) |
| FLAC Metadata | StreamInfo, Vorbis Comments, Pictures, and other blocks |

## Supported Formats

### Images
- JPEG (.jpg, .jpeg) – EXIF, IPTC, XMP, ICC
- PNG (.png) – Text chunks, EXIF, XMP, ICC
- GIF (.gif) – Comments, XMP, NETSCAPE extension
- WebP (.webp) – EXIF, XMP, ICC
- TIFF (.tiff, .tif) – IFD-based metadata
- HEIC/HEIF (.heic, .heif, .hif) – EXIF, XMP, ICC

### RAW Formats
- CR2 (.cr2) – Canon RAW
- DNG (.dng) – Adobe Digital Negative
- NEF (.nef) – Nikon RAW
- ARW (.arw) – Sony RAW
- ORF (.orf) – Olympus RAW
- RAF (.raf) – Fujifilm RAW
- RW2 (.rw2) – Panasonic RAW
- PEF (.pef) – Pentax RAW
- SRW (.srw) – Samsung RAW
- 3FR (.3fr) – Hasselblad RAW
- Most TIFF-based RAW formats

### Audio/Video
- MP3 (.mp3) – ID3v2.2, v2.3, v2.4 tags
- FLAC (.flac) – All metadata blocks
- MP4 (.mp4, .m4v) – iTunes metadata, EXIF
- M4A (.m4a) – AAC audio container

## CLI Usage

```bash
# Basic extraction
imx photo.jpg

# JSON output
imx --format json photo.jpg

# CSV output
imx --format csv photo.jpg

# Filter by directory
imx --dir IFD0 photo.jpg

# Get specific tag
imx --tag Make photo.jpg

# Process multiple files
imx --recursive ./photos/

# Read from stdin
cat photo.jpg | imx

# Process audio files
imx song.mp3
imx audio.flac

# Process video files
imx video.mp4
```

## Benchmarks

Benchmarks depend on hardware and Go version. Run them locally to establish your own baselines:

```bash
make bench
```

The suite covers high-level APIs and all parsers; see `Makefile` for options.

Latest local run (darwin/arm64, Go 1.25, benchtime=2s):

```
High-Level API
BenchmarkMetadataFromFile-12        310715 ns/op   280028 B/op    2879 allocs/op
BenchmarkMetadataFromBytes-12       173898 ns/op   279574 B/op    2875 allocs/op
BenchmarkMetadataFromReader-12      191216 ns/op   425643 B/op    3105 allocs/op
BenchmarkMetadata_Tag-12                 11.10 ns/op        0 B/op       0 allocs/op
BenchmarkMetadata_GetAll-12              68.25 ns/op       48 B/op       1 allocs/op
BenchmarkMetadata_Each-12                69.36 ns/op        0 B/op       0 allocs/op

Parser Benchmarks
PNG     309 ns/op     1152 B/op     16 allocs/op
IPTC    2382 ns/op    6968 B/op    111 allocs/op
WebP    2285 ns/op    4082 B/op    133 allocs/op
ICC     2483 ns/op    8213 B/op    134 allocs/op
TIFF    2524 ns/op    4010 B/op    147 allocs/op
FLAC    2917 ns/op    9742 B/op    120 allocs/op
MP4     12576 ns/op   46324 B/op   258 allocs/op
ID3     15263 ns/op   78209 B/op   207 allocs/op
XMP     20479 ns/op   24456 B/op   373 allocs/op
GIF     41003 ns/op   160609 B/op  272 allocs/op
JPEG    42684 ns/op   45150 B/op   774 allocs/op
HEIC    57765 ns/op   87397 B/op   1822 allocs/op
CR2     69119 ns/op   114971 B/op  889 allocs/op
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
