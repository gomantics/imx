# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`imx` is a Go library for fast, dependency-free extraction of image metadata. It extracts EXIF, IPTC, XMP, and ICC metadata from JPEG, PNG, WebP, TIFF, HEIF/HEIC, and AVIF formats.

**Read `docs/SPEC.md` for complete API specification and requirements.**

## Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run fuzz tests (when implemented)
go test -fuzz=FuzzJPEG ./internal/container/jpeg
go test -fuzz=FuzzEXIF ./internal/meta/exif

# Format and vet
go fmt ./...
go vet ./...
```

## Architecture

### Three-Layer Pipeline

1. **Container Parsers** (`internal/container/`) - Format-specific parsers that extract raw metadata blocks
   - JPEG: Parse markers (SOI, APPn, SOF, SOS, EOI)
   - PNG: Parse chunks (IHDR, eXIf, iTXt, iCCP, IDAT)
   - WebP: Parse RIFF chunks (VP8, EXIF, XMP, ICCP)
   - TIFF: Parse IFDs
   - HEIF: Parse MP4-like boxes

2. **Namespace Parsers** (`internal/meta/`) - Decode raw blocks into structured metadata
   - EXIF: TIFF IFD parsing (tags, types, values)
   - IPTC: IIM dataset parsing
   - XMP: XML parsing and namespace mapping
   - ICC: Profile handling

3. **Pipeline Orchestration** (`internal/pipeline/`) - Routes RawBlocks to parsers and assembles final Metadata

### Data Flow

```
io.Reader
  → bufio.Reader
  → ContainerParser.Detect() [format detection]
  → ContainerParser.Parse() [extract raw blocks]
  → []RawBlock
  → NamespaceParser.Parse() [decode to tags/directories]
  → []Directory
  → Metadata
```

### Package Structure

```
imx/
  extractor.go        # Extractor type, New(), options
  api.go              # MetadataFromFile/Bytes/URL convenience functions
  types.go            # Metadata, Directory, Tag, Namespace, TagID
  registry.go         # RegisterFormat, RegisterNamespace
  errors.go           # ErrUnknownFormat, ErrTruncatedData, etc.

internal/
  container/
    sniff.go          # Format detection helpers
    jpeg/jpeg.go      # JPEG marker parsing
    png/png.go        # PNG chunk parsing
    webp/webp.go      # WebP RIFF parsing
    tiff/tiff.go      # TIFF container parsing
    heif/heif.go      # HEIF/AVIF boxes parsing

  meta/
    exif/
      exif.go         # EXIF namespace parser
      tiff_ifd.go     # Shared IFD parsing logic
    iptc/iptc.go      # IPTC IIM parser
    xmp/xmp.go        # XMP XML parser
    icc/icc.go        # ICC profile parser

  pipeline/
    pipeline.go       # RawBlock routing and orchestration

  testdata/
    jpeg/, png/, webp/, tiff/, heif/
    golden/           # JSON goldens from ExifTool/Exiv2
```

## Critical Constraints

**MUST:**
- No external dependencies (stdlib only, no cgo)
- Never panic on user input - always return errors
- Never load entire files - use streaming IO with bufio.Reader
- Wrap reader with io.LimitedReader when MaxBytes is set
- Handle corrupted/truncated files gracefully
- Make Extractor safe for concurrent use (config cloned per call)
- Call RegisterFormat/RegisterNamespace in init() before concurrent use

**SHOULD:**
- Use buffered IO (bufio.Reader) and minimize copies
- Provide benchmarks and track allocations
- Protect against malicious size declarations (validate before allocating)

## API Design Patterns

The library provides multiple access patterns for different use cases:

**1. Convenience helpers** - For simple cases (1-3 common fields):
```go
dt := meta.DateTimeOriginal()      // returns time.Time
orientation := meta.Orientation()   // returns int
make := meta.Make()                 // returns string
gps := meta.GPSCoordinates()        // returns *GPSCoord
```

**2. Batch retrieval** - For extracting many fields without repetitive if-statements:
```go
fields := meta.GetAll(
    "Exif:DateTimeOriginal",
    "Exif:Orientation",
    "Exif:Make",
    "Exif:Model",
    "XMP:dc:title",
)
// Returns map[TagID]any
```

**3. Iterators** - For processing/filtering tags:
```go
meta.Each(func(tag Tag) bool {
    fmt.Printf("%s: %v\n", tag.Name, tag.Value)
    return true  // continue
})

meta.EachInNamespace(imx.NamespaceEXIF, func(tag Tag) bool {
    // process only EXIF tags
    return true
})
```

**4. Direct access** - For power users and custom namespaces:
```go
tag, ok := meta.Tag(imx.NamespaceEXIF, "Exif:RareField")
if ok {
    fmt.Printf("Raw: %x, Value: %v\n", tag.Raw, tag.Value)
}
```

## Go Design Principles

From Effective Go and Go Proverbs:

1. **Make the zero value useful** - `ExtractorConfig{}` must work with sensible defaults
2. **Accept interfaces, return structs** - Accept `io.Reader`, return `Metadata`/`Extractor`
3. **Small interfaces** - `ContainerParser` and `NamespaceParser` are minimal and focused
4. **Errors are values** - No panics, return descriptive errors with context
5. **Avoid stutter** - Use `imx.Metadata` not `imx.ImageMetadata`
6. **Keep exported API minimal** - Only export what's necessary for users

## Error Handling

- Never panic on user input
- Wrap errors with context: `fmt.Errorf("imx: parse jpeg: %w", err)`
- Use sentinel errors: `ErrUnknownFormat`, `ErrTruncatedData`, `ErrUnsupportedMeta`
- Support partial results: If `StopOnFirstErr` is false, return partial Metadata + aggregated error via `PartialError` type

## Performance Requirements

**IO Strategy:**
- Use `bufio.Reader` for all container parsing
- Scan sequentially (JPEG markers, PNG chunks) - never read full file
- Use `Peek()` for format detection, `Read()` for payloads
- Respect `MaxBytes` limit via `io.LimitedReader` wrapper

**Memory:**
- Size buffers based on segment/chunk length from headers
- Validate sizes before allocating to prevent attacks
- For `Tag.Raw`, decide: always copy (safe) or reference (document lifetime)

**Concurrency:**
- `Extractor` is concurrent-safe - config is cloned on each call
- Registries are immutable after construction
- Registration must happen before concurrent use

## Core Interfaces

```go
// ContainerParser extracts raw metadata blocks from a container format
type ContainerParser interface {
    Detect(peek []byte) bool
    Parse(r *bufio.Reader, cfg ExtractorConfig) ([]RawBlock, error)
}

// NamespaceParser decodes raw blocks into structured directories/tags
type NamespaceParser interface {
    Namespace() Namespace
    Parse(blocks []RawBlock, cfg ExtractorConfig) ([]Directory, error)
}
```

## Testing Strategy

**Unit Tests:**
- Test each container parser with synthetic inputs
- Test each namespace parser with known metadata blocks
- Verify error handling for corrupted/truncated data

**Golden Tests:**
- Use real-world images in `internal/testdata/`
- Generate expected output via ExifTool/Exiv2 (JSON format)
- Compare imx output against golden files

**Fuzz Tests:**
- Fuzz container parsers with random bytes - ensure no panics
- Fuzz namespace parsers with random payloads

**Benchmarks:**
- Benchmark JPEG EXIF extraction, XMP/ICC parsing
- Use `-benchmem` to track allocations
- Monitor for regressions

## Implementation Phases

1. **Core API Skeleton** - Types, Extractor, options, stub registries
2. **JPEG + EXIF** - JPEG container parser, basic EXIF TIFF IFD parser
3. **XMP & IPTC for JPEG** - APP1 XMP, APP13 IPTC support
4. **PNG & WebP** - eXIf, iTXt, iCCP chunks
5. **ICC & Additional Formats** - ICC profiles, TIFF, HEIF, AVIF
6. **Fuzzing & Performance** - Fuzz tests, profiling, optimization
7. **Docs & Polish** - README, examples, API review before v1.0
