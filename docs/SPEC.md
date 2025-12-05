Project Specification – github.com/gomantics/imx

1. Overview

imx is a Go library for fast, dependency-free extraction of image metadata. It provides:
	•	Simple high-level functions for reading metadata from files, readers, byte slices, and URLs.
	•	An extensible, composable architecture with:
	•	Container parsers (JPEG/PNG/WebP/TIFF/HEIF/…)
	•	Namespace parsers (EXIF/IPTC/XMP/ICC/custom)
	•	A clean, Go-idiomatic API designed for production use in backend services, CLIs, and tools.

The project emphasizes:
	•	High performance and low allocations.
	•	Streaming I/O (do not load whole files unnecessarily).
	•	Strong test coverage (unit, integration, golden, fuzz, benchmarks).
	•	Adherence to Go best practices and SOLID-ish design within Go’s idioms.

⸻

2. Scope

2.1 In-Scope (v1–v1.x)
	•	Read-only metadata extraction for:
	•	EXIF (incl. GPS, basic orientation; maker notes handled as opaque blobs initially).
	•	IPTC IIM.
	•	XMP (core namespaces: dc, photoshop, basic mapping).
	•	ICC profiles (opaque or minimally parsed).
	•	Supported container formats:
	•	v1:
	•	JPEG
	•	v1.1:
	•	PNG
	•	WebP
	•	v1.2+:
	•	TIFF
	•	HEIF/HEIC
	•	AVIF
Public API:
imx.Metadata(r io.Reader, opts ...Option) (Metadata, error)
imx.MetadataFromFile(path string, opts ...Option) (Metadata, error)
imx.MetadataFromBytes(data []byte, opts ...Option) (Metadata, error)
imx.MetadataFromURL(url string, opts ...Option) (Metadata, error)

	•	Advanced API:
	•	Extractor type with configurable options and extensibility hooks.
	•	Public registration for custom parsers.

2.2 Out-of-Scope (for early versions)
	•	Full writer APIs for all formats (metadata modification).
	•	Complete maker-note decoding for all camera brands.
	•	Video/audio/document metadata (focus is images; future versions may expand).

⸻

3. High-Level Requirements

3.1 Functional Requirements
	1.	The library MUST provide functions to read metadata from:
	•	io.Reader (Metadata)
	•	File path (MetadataFromFile)
	•	Byte slice (MetadataFromBytes)
	•	HTTP/HTTPS URL (MetadataFromURL)
	2.	The library MUST:
	•	Detect supported image formats from initial bytes or container signatures.
	•	Extract raw metadata blocks per format.
	•	Decode those blocks into a unified Metadata model.
	•	Support filtering by metadata namespace (e.g. EXIF-only).
	3.	The library MUST be extensible:
	•	Allow registration of additional container parsers.
	•	Allow registration of additional namespace parsers.
	4.	The library MUST handle:
	•	Corrupted or truncated files without panicking.
	•	Multiple metadata blocks of the same type per image (e.g. multiple XMP packets, ICC segments).
	5.	The library SHOULD:
	•	Provide helper methods for convenient tag lookup.
	•	Allow configuration of limits (e.g. max bytes read, enabled formats, enabled namespaces).

3.2 Non-Functional Requirements
	•	Performance:
	•	MUST avoid loading the entire file unless strictly necessary.
	•	SHOULD use buffered IO (bufio.Reader) and minimal copies.
	•	SHOULD provide benchmarks and keep regressions detectable.
	•	Dependencies:
	•	MUST have no non-stdlib dependencies (no cgo).
	•	Error Handling:
	•	MUST never panic on user input.
	•	Errors MUST be descriptive and wrap internal causes where helpful.
	•	Concurrency:
	•	Extractor SHOULD be safe for concurrent use across goroutines once constructed.
	•	Registration of parsers (formats/namespaces) MUST be done before concurrent use.
	•	Testing:
	•	MUST have unit tests for core logic and edge cases.
	•	MUST have golden tests comparing against known tools (e.g. ExifTool).
	•	MUST have fuzz tests for parsers.
	•	SHOULD target very high coverage (~100% where feasible).
	•	Documentation:
	•	MUST provide README with basic usage and examples.
	•	SHOULD provide internal docs for contributors, including coding style and project layout.

⸻

4. Public Data Model

4.1 Core Types
package imx

type Namespace string

const (
    NamespaceEXIF Namespace = "exif"
    NamespaceIPTC Namespace = "iptc"
    NamespaceXMP  Namespace = "xmp"
    NamespaceICC  Namespace = "icc"
    // Future: NamespaceCustom or user-defined strings.
)

type TagID string

// Tag represents a single metadata attribute.
type Tag struct {
    Namespace Namespace
    ID        TagID  // e.g. "Exif:DateTimeOriginal", "IPTC:Caption", "XMP:dc:title"
    Name      string // human-readable, stable within this library ("DateTimeOriginal")
    Type      string // short descriptor: "string", "int", "rational", "gpscoord", "time", etc.
    Value     any    // parsed Go value (string, int, time.Time, GPSCoord, []string, etc.)
    Raw       []byte // raw encoded bytes for potential round-trip or custom parsing
}

// Directory is a logical collection of tags for a given namespace and grouping.
type Directory struct {
    Namespace Namespace
    Name      string             // e.g. "IFD0", "ExifIFD", "GPS", "IPTC-Application2", "XMP-dc"
    Tags      map[TagID]Tag      // tags by ID
}

// Metadata is the top-level container for all parsed metadata.
type Metadata struct {
    Directories []Directory
    // Optionally: an internal index (unexported) for fast lookup.
}

// Convenience methods.
func (m Metadata) Directory(namespace Namespace, name string) (Directory, bool)
func (m Metadata) Tag(namespace Namespace, id TagID) (Tag, bool)

// Convenience helpers for common EXIF fields (returns zero value if missing)
func (m Metadata) DateTimeOriginal() time.Time
func (m Metadata) Orientation() int
func (m Metadata) Make() string
func (m Metadata) Model() string
func (m Metadata) GPSCoordinates() *GPSCoord
func (m Metadata) ISO() int
// ... additional common field helpers

// Batch retrieval for multiple fields
func (m Metadata) GetAll(ids ...TagID) map[TagID]any

// Iterators for processing tags
func (m Metadata) Each(fn func(Tag) bool)
func (m Metadata) EachInNamespace(namespace Namespace, fn func(Tag) bool)

4.2 Future Helper Types (optional but recommended later)
// GPSCoord represents a parsed GPS coordinate.
type GPSCoord struct {
    Lat      float64
    Lon      float64
    Altitude float64
    // Maybe: Reference (N/S/E/W), Accuracy, etc.
}
5. Public API Specification

5.1 Extractor Configuration
type Format string

const (
    FormatJPEG Format = "jpeg"
    FormatPNG  Format = "png"
    FormatWebP Format = "webp"
    FormatTIFF Format = "tiff"
    FormatHEIF Format = "heif"
    // ...
)

// ExtractorConfig holds configurable options. Zero value MUST be valid and useful.
type ExtractorConfig struct {
    MaxBytes       int64       // 0 == no explicit limit; callers may set for safety.
    BufferSize     int         // 0 == default (e.g. 64KiB).
    Namespaces     []Namespace // nil/empty == all supported.
    Formats        []Format    // nil/empty == auto-detect among all registered formats.
    StopOnFirstErr bool        // false == try to continue and return partial metadata.
    // Future: StrictnessLevel, Logger, etc.
}

// Option is a functional option modifying ExtractorConfig.
type Option func(*ExtractorConfig)

Example options (to implement):
func WithMaxBytes(n int64) Option
func WithBufferSize(n int) Option
func WithNamespaces(ns ...Namespace) Option
func WithFormats(fs ...Format) Option
func WithStopOnFirstError() Option

5.2 Extractor Type

// Extractor is a reusable metadata extractor, safe for concurrent use once constructed.
type Extractor struct {
    cfg          ExtractorConfig
    containerReg *containerRegistry
    nsReg        *namespaceRegistry
}

// New returns a new Extractor with default configuration overridden by options.
func New(opts ...Option) *Extractor

The Extractor methods mirror the top-level helpers:

func (e *Extractor) Metadata(r io.Reader, opts ...Option) (Metadata, error)
func (e *Extractor) MetadataFromFile(path string, opts ...Option) (Metadata, error)
func (e *Extractor) MetadataFromBytes(data []byte, opts ...Option) (Metadata, error)
func (e *Extractor) MetadataFromURL(url string, opts ...Option) (Metadata, error)

Behavior:
	•	Each call:
	•	Clones e.cfg into a local config.
	•	Applies per-call opts to that clone.
	•	Executes the extraction pipeline with the effective config.

5.3 Top-Level Convenience Functions

These are thin wrappers around a package-level default Extractor:

func Metadata(r io.Reader, opts ...Option) (Metadata, error)
func MetadataFromFile(path string, opts ...Option) (Metadata, error)
func MetadataFromBytes(data []byte, opts ...Option) (Metadata, error)
func MetadataFromURL(url string, opts ...Option) (Metadata, error)

Implementation detail:
	•	A package-level defaultExtractor *Extractor is initialized lazily with default config and built-in parsers.

⸻

6. Extensibility Interfaces

6.1 Container Parsers (Formats)

Internal representation types (exported only via constructors/registrations where necessary):
type MetaKind int

const (
    MetaKindEXIF MetaKind = iota
    MetaKindIPTC
    MetaKindXMP
    MetaKindICC
    // Future: MetaKindCustom
)

// RawBlock is a raw metadata payload extracted from a container.
type RawBlock struct {
    Kind    MetaKind
    Payload []byte
    Origin  string // e.g. "APP1 Exif", "APP13 IPTC", "eXIf chunk", "EXIF box"
    Format  Format
    Index   int    // sequence number for multiple blocks of same type
}

Container parser interface:

// ContainerParser parses a specific container format and emits raw metadata blocks.
type ContainerParser interface {
    // Detect returns true if this parser supports the given initial bytes.
    // 'peek' is a prefix of the stream (e.g. 16–64 bytes).
    Detect(peek []byte) bool

    // Parse reads from r (typically a *bufio.Reader wrapping original reader)
    // and returns all metadata blocks found in this container.
    Parse(r *bufio.Reader, cfg ExtractorConfig) ([]RawBlock, error)
}
6.2 Namespace Parsers (Metadata Families)
// NamespaceParser parses raw metadata blocks for a given namespace into directories/tags.
type NamespaceParser interface {
    Namespace() Namespace

    // Parse consumes all relevant RawBlocks and returns Directories for this namespace.
    Parse(blocks []RawBlock, cfg ExtractorConfig) ([]Directory, error)
}

6.3 Registries

Internal registries:

type containerRegistry struct {
    parsers []ContainerParser
}

type namespaceRegistry struct {
    parsers map[Namespace]NamespaceParser
}

type containerRegistry struct {
    parsers []ContainerParser
}

type namespaceRegistry struct {
    parsers map[Namespace]NamespaceParser
}

Container detection:
	•	Given a *bufio.Reader:
	•	Peek some bytes.
	•	Iterate parsers and call Detect.
	•	First parser that returns true is used.
	•	If none match, return ErrUnknownFormat.

6.4 Public Registration Functions
// RegisterFormat registers a custom container parser.
// Intended to be called from init() before any concurrent use of imx.
func RegisterFormat(p ContainerParser)

// RegisterNamespace registers a custom namespace parser.
func RegisterNamespace(p NamespaceParser)

Constraints:
	•	Registration functions are not goroutine-safe at runtime; they are intended for early initialization.
	•	Once extraction begins concurrently, no further registrations should occur.

⸻

7. Extraction Pipeline (End-to-End)

7.1 Steps

For Extractor.Metadata(r, opts...):
	1.	Config Resolution
	•	Clone base ExtractorConfig.
	•	Apply per-call Options.
	2.	Reader Setup
	•	Wrap r with io.LimitedReader if MaxBytes > 0.
	•	Wrap with bufio.Reader using BufferSize (or default).
	3.	Format Detection
	•	Peek N bytes (e.g. peek, err := br.Peek(64)).
	•	For each ContainerParser:
	•	Call Detect(peek).
	•	Use first parser that returns true.
	•	If none match → return ErrUnknownFormat.
	4.	Container Parsing
	•	Call selected parser’s Parse(br, cfg):
	•	Traverse segments/chunks/boxes depending on format:
	•	JPEG: markers (SOI, APPn, SOF, SOS, EOI).
	•	PNG: chunks (IHDR, eXIf, iTXt, iCCP, IDAT, etc.).
	•	WebP: RIFF chunks (VP8, EXIF, XMP, ICCP).
	•	TIFF: IFDs.
	•	HEIF: MP4-like boxes (meta, Exif, etc.).
	•	For each metadata-containing piece:
	•	Identify MetaKind.
	•	Extract payload bytes into RawBlock.
	•	Return []RawBlock.
	5.	Namespace Parsing
	•	Group RawBlocks by MetaKind.
	•	For each registered NamespaceParser:
	•	Filter blocks that correspond to that namespace (via MetaKind).
	•	Call Parse(blocks, cfg) to produce []Directory.
	•	Collect directories.
	6.	Namespace Filtering
	•	If cfg.Namespaces is non-empty, filter directories to those namespaces only.
	7.	Metadata Assembly
	•	Create Metadata{Directories: allDirs}.
	•	Optionally build an internal lookup index (unexported) for Tag/Directory helpers.
	8.	Error Aggregation
	•	If container parse failed → return error, no metadata.
	•	If a namespace parser fails:
	•	If StopOnFirstErr:
	•	Return error and possibly partial metadata.
	•	Else:
	•	Accumulate PartialError, return metadata + aggregated error.

7.2 Error Types

Define sentinel errors:
var (
    ErrUnknownFormat   = errors.New("imx: unknown format")
    ErrTruncatedData   = errors.New("imx: truncated data")
    ErrUnsupportedMeta = errors.New("imx: unsupported metadata block")
)

Optional composite error:
type PartialError struct {
    FormatErr     error
    NamespaceErrs map[Namespace]error
}

func (e *PartialError) Error() string

8. Internal Package Layout

Proposed structure:
imx/
    extractor.go        // Extractor, default extractor, options
    api.go              // Public functions MetadataFromFile, etc.
    types.go            // Metadata, Directory, Tag, Namespace, TagID
    registry.go         // Registration functions, registry construction
    errors.go           // Public error values and types

internal/
  container/
      sniff.go         // format detection helper
      jpeg/
          jpeg.go      // JPEG marker parsing, APPn handling
      png/
          png.go       // PNG chunk parsing, eXIf, iTXt, iCCP
      webp/
          webp.go      // WebP RIFF chunk parsing
      tiff/
          tiff.go      // TIFF container parsing
      heif/
          heif.go      // HEIF/AVIF boxes parsing

  meta/
      exif/
          exif.go      // EXIF TIFF parsing (IFDs, tags)
          tiff_ifd.go  // shared IFD logic
      iptc/
          iptc.go      // IPTC IIM datasets parsing
      xmp/
          xmp.go       // XMP XML parsing + namespace mapping
      icc/
          icc.go       // ICC profile handling (basic or opaque)

  pipeline/
      pipeline.go      // Orchestration glue: RawBlock routing, parser invocations

  testdata/
      jpeg/
      png/
      webp/
      tiff/
      heif/
      golden/          // JSON goldens from ExifTool/Exiv2

imxtest/
    helpers.go         // Helper functions for external tests (optional)

All packages under internal/ are not imported by external modules, allowing internal refactoring without breaking API.

⸻

9. Performance & Resource Usage

9.1 IO & Buffering
	•	Use bufio.Reader for container parsing.
	•	Avoid full-file reads:
	•	For JPEG, scan markers sequentially.
	•	For PNG, read each chunk header and decide whether to skip or read payload.
	•	For WebP, read RIFF chunk headers, skip non-metadata chunks.
	•	Respect MaxBytes:
	•	Wrap the original reader in io.LimitedReader when configured.
	•	Handle large declared sizes with caution; protect against attempts to allocate huge slices.

9.2 Allocation Strategy
	•	Use local buffers sized appropriately for segment lengths.
	•	Reuse temporary buffers where appropriate.
	•	For Tag.Raw, decide whether to:
	•	Always copy (simple, safe).
	•	Or have an option to avoid copies (document lifetime assumptions clearly).

9.3 Concurrency
	•	Extractor MUST be safe for concurrent calls:
	•	Its config is copied on each invocation.
	•	Registries are immutable after construction.
	•	Registration (RegisterFormat, RegisterNamespace) MUST be done in init() or before concurrent use.

⸻

10. Coding Guidelines & Style

The codebase should follow common Go guidelines:
	1.	Effective Go & Go Proverbs
	•	Small, behavior-focused interfaces.
	•	“The bigger the interface, the weaker the abstraction.”
	•	“Make the zero value useful.”
	•	“Errors are values.”
	2.	API Design
	•	Accept interfaces (io.Reader), return concrete types (Metadata, Extractor).
	•	Avoid stutter (imx.Metadata, not imx.ImageMetadata if the context is clear).
	•	Keep exported API minimal and well-documented.
	3.	Error Handling
	•	Do not panic on user input.
	•	Wrap errors with context: fmt.Errorf("imx: parse jpeg: %w", err).
	•	Use sentinel errors for common conditions (ErrUnknownFormat etc.).
	4.	No External Runtime Dependencies
	•	Only use Go stdlib.
	•	No cgo, no external C libraries.
	5.	Testing & Tooling
	•	Use go test ./... as primary testing mechanism.
	•	Optional: add a small Makefile or task file for developer convenience, but not required.

⸻

11. Testing Strategy

11.1 Unit Tests
	•	Per container parser:
	•	JPEG: verify correct detection, extraction of APP1/APP13 etc.
	•	PNG: verify handling of eXIf, iTXt with XMP keyword, iCCP.
	•	WebP: verify EXIF/XMP/ICCP chunks.
	•	Per namespace parser:
	•	EXIF: parse synthetic TIFF IFD structures.
	•	IPTC: parse sample IIM records with known datasets.
	•	XMP: parse small XML fragments with known dc/photoshop fields.
	•	ICC: basic recognition or tag extraction.

11.2 Integration & Golden Tests
	•	Maintain a corpus in internal/testdata/:
	•	Real images from different cameras/phones/software.
	•	Edited variants (photo editors, social media exports).
	•	Corrupted/truncated files.
	•	For each image:
	•	Generate golden metadata via ExifTool/Exiv2 and store as JSON.
	•	In tests:
	•	Call imx.MetadataFromFile.
	•	Normalize Metadata to a comparable representation.
	•	Compare against golden (allowing some differences where not implemented).

11.3 Benchmarks
	•	Benchmarks for:
	•	JPEG EXIF-heavy image.
	•	JPEG with XMP/ICC.
	•	PNG/WebP when implemented.
	•	Measure allocations and runtime; use -benchmem.

11.4 Fuzz Tests
	•	Fuzz JPEG segment walker:
	•	Feed random byte slices to JPEG parser and ensure no panics.
	•	Fuzz EXIF TIFF parser:
	•	Feed random byte slices as EXIF blocks.
	•	Fuzz PNG/WebP parsers similarly.

⸻

12. Example Usage

12.1 Basic Usage with Convenience Helpers
package main

import (
    "fmt"
    "log"

    "github.com/gomantics/imx"
)

func main() {
    meta, err := imx.MetadataFromFile("photo.jpg")
    if err != nil {
        log.Fatal(err)
    }

    // Convenience helpers for common fields
    fmt.Printf("Taken at: %v\n", meta.DateTimeOriginal())
    fmt.Printf("Orientation: %v\n", meta.Orientation())
    fmt.Printf("Camera: %s %s\n", meta.Make(), meta.Model())

    if gps := meta.GPSCoordinates(); gps != nil {
        fmt.Printf("Location: %f, %f\n", gps.Lat, gps.Lon)
    }
}

12.2 Batch Retrieval for Multiple Fields
func main() {
    meta, err := imx.MetadataFromFile("photo.jpg")
    if err != nil {
        log.Fatal(err)
    }

    // Get multiple fields at once
    fields := meta.GetAll(
        "Exif:DateTimeOriginal",
        "Exif:Orientation",
        "Exif:Make",
        "Exif:Model",
        "Exif:ISO",
        "XMP:dc:title",
    )

    for id, value := range fields {
        fmt.Printf("%s: %v\n", id, value)
    }
}

12.3 Iterator Pattern for Processing All Tags
func main() {
    meta, err := imx.MetadataFromFile("photo.jpg")
    if err != nil {
        log.Fatal(err)
    }

    // Process all EXIF tags
    meta.EachInNamespace(imx.NamespaceEXIF, func(tag imx.Tag) bool {
        fmt.Printf("%s = %v (%s)\n", tag.Name, tag.Value, tag.Type)
        return true // continue iteration
    })
}

12.4 Power User: Direct Tag Access
func main() {
    meta, err := imx.MetadataFromFile("photo.jpg")
    if err != nil {
        log.Fatal(err)
    }

    // Explicit access for rare/custom fields
    if tag, ok := meta.Tag(imx.NamespaceEXIF, "Exif:MakerNote"); ok {
        fmt.Printf("MakerNote (raw): %x\n", tag.Raw)
        fmt.Printf("Type: %s, Value: %v\n", tag.Type, tag.Value)
    }
}
12.5 Advanced Usage with Reusable Extractor

extractor := imx.New(
    imx.WithNamespaces(imx.NamespaceEXIF, imx.NamespaceXMP),
    imx.WithMaxBytes(10<<20), // 10 MiB safety limit
)

func handleUpload(r io.Reader) (imx.Metadata, error) {
    return extractor.Metadata(r)
}

13. Implementation Roadmap (Suggested)
	1.	Phase 1: Core API Skeleton
	•	Implement Metadata types, Extractor, options, and stub registries.
	•	Implement Metadata/MetadataFromFile/… that return ErrUnknownFormat.
	2.	Phase 2: JPEG + EXIF
	•	Implement JPEG container parser (SOI, APPn parsing).
	•	Implement basic EXIF TIFF IFD parser.
	•	Unit tests for JPEG + EXIF.
	3.	Phase 3: XMP & IPTC for JPEG
	•	APP1 XMP and APP13 IPTC support.
	•	Namespace parsers for XMP & IPTC.
	•	Golden tests against ExifTool.
	4.	Phase 4: PNG & WebP
	•	PNG: eXIf, iTXt (XMP), iCCP.
	•	WebP: EXIF, XMP, ICCP chunks.
	5.	Phase 5: ICC & Additional Formats
	•	ICC profile handling (opaque or minimal parser).
	•	TIFF, HEIF, AVIF container parsers.
	6.	Phase 6: Fuzzing & Performance
	•	Add fuzz tests.
	•	Profile and optimize hot paths.
	7.	Phase 7: Docs & Polish
	•	README, docs, examples.
	•	API review and any renames before v1.0.0.
