# Contributing to imx

Thank you for considering contributing to imx! This document provides guidelines and information for contributors.

## Project Overview

`imx` is a zero-dependency Go library for fast image metadata extraction. It supports EXIF, IPTC, XMP, and ICC color profiles from JPEG images (with more formats planned).


## Development Setup

### Prerequisites

- Go 1.22 or later
- Make (optional, for convenience commands)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/gomantics/imx.git
cd imx

# Run all checks (lint, test, build)
make check

# Or without make:
go vet ./...
go test -race ./...
go build ./...
```

## Project Structure

```
imx/
├── api.go              # Package-level convenience functions
├── config.go           # Configuration and options
├── errors.go           # Sentinel errors
├── extractor.go        # Core Extractor type
├── types.go            # Metadata, Directory, Tag types
├── tags.go             # Tag ID constants (TagMake, TagModel, etc.)
│
├── cmd/imx/            # CLI tool
│   └── main.go         # Rich command-line interface
│
├── examples/           # Usage examples
│   ├── basic/          # Simple metadata extraction
│   └── advanced/       # Advanced features demo
│
├── internal/
│   ├── format/         # Container format parsers
│   │   └── jpeg/       # JPEG marker parsing
│   │
│   └── meta/           # Metadata specification parsers
│       ├── exif/       # EXIF TIFF/IFD parsing
│       ├── iptc/       # IPTC-IIM parsing
│       ├── xmp/        # XMP XML parsing
│       └── icc/        # ICC color profile parsing
│
├── testdata/goldens/   # Test images with expected metadata
│
└── Makefile            # Build automation
```

## Architecture

### Three-Layer Pipeline

1. **Format Layer** (`internal/format/`) - Extracts raw metadata blocks from container formats (JPEG APP markers, etc.)
2. **Meta Layer** (`internal/meta/`) - Parses raw blocks into structured tags (EXIF IFDs, XMP XML, etc.)
3. **API Layer** (root package) - Provides user-facing types and functions

### Data Flow

```
File/Reader → Format Parser → RawBlocks → Meta Parsers → Directories → Metadata
     ↓              ↓              ↓            ↓              ↓
   JPEG         APP1/APP2/     EXIF/IPTC/    Parsed      User-facing
               APP13 data      XMP/ICC       Tags         result
```

## Development Guidelines

### Code Style

- Follow standard Go conventions (`go fmt`, `go vet`)
- Use meaningful variable and function names
- Add comments for exported types and functions
- Keep functions focused and small

### Testing

```bash
# Run all tests
make test

# Run library tests with coverage
make test-lib

# Generate coverage report
make coverage

# Generate HTML coverage report
make coverage-html
```

**Coverage Target: 100% for library code** (internal/ and root packages)

### Commit Messages

**IMPORTANT**: Commit messages serve as the project's changelog. Write detailed, informative messages that explain the what, why, and impact of changes.

#### Format

```
type(scope): short description (max 72 chars)

Detailed explanation of changes:
- What was changed and why
- Any breaking changes (BREAKING:)
- Performance impact (if any)
- Related issues (#123)

[optional footer with breaking changes, deprecations]
```

#### Types
- `feat` - New feature (adds functionality)
- `fix` - Bug fix (fixes broken functionality)
- `refactor` - Code restructuring (no functional changes)
- `perf` - Performance improvement
- `test` - Add/update tests
- `docs` - Documentation only
- `chore` - Maintenance (deps, build, etc.)
- `style` - Code style/formatting

#### Breaking Changes

Prefix breaking changes with `BREAKING:` in the commit body:

```
feat(api): simplify Tag() method signature

BREAKING: Removed spec parameter from Metadata.Tag()

Old: meta.Tag(imx.SpecEXIF, imx.TagMake)
New: meta.Tag(imx.TagMake)

The spec is now automatically extracted from the TagID.
This simplifies the API and makes tag lookup more intuitive.
```

#### Good Examples

```
feat(bench): add comprehensive benchmark suite

Added 38 benchmarks covering all major code paths:
- High-level API benchmarks (MetadataFromFile, etc.)
- Parser-specific benchmarks (EXIF, IPTC, XMP, ICC)
- Metadata operation benchmarks (Tag lookup, iteration)

Also added:
- GitHub Actions workflow for continuous benchmarking
- Makefile targets: bench, bench-all, bench-report, bench-compare
- Scripts for benchmark comparison using benchstat

Performance on Apple M4 Pro:
- MetadataFromFile: 279μs/op
- Tag lookup: 7ns/op
- EXIF parsing: 6.9μs/op
```

```
refactor(exif): reduce parseValue() complexity using strategy pattern

Reduced parseValue() from 107 lines to 32 lines (-70%) by:
- Creating TIFFTypeParser interface
- Implementing 10 type-specific parsers
- Moving parsers to internal/common/tiff_types.go

Benefits:
- Easier to test (each parser isolated)
- More maintainable (single responsibility)
- Reusable for future TIFF format support

No functional changes, all tests pass.
```

```
fix(iptc): handle extended dataset sizes correctly

Fixed parsing of IPTC datasets with sizes >32KB by properly
handling the extended size format (high bit set).

Added DatasetReader to encapsulate byte-level parsing and
reduce parseIPTCIIM() complexity from 75 to 20 lines (-73%).

Fixes #123
```

#### Bad Examples

```
❌ fix bug
❌ update code
❌ refactor
❌ wip
❌ fixes
```

These don't explain what was changed or why.

### Adding a New Metadata Parser

1. Create package in `internal/meta/<spec>/`
2. Implement `meta.Parser` interface:
   ```go
   type Parser interface {
       Spec() meta.Spec
       Parse(blocks []format.RawBlock) ([]meta.Directory, error)
   }
   ```
3. Register in `extractor.go` metaParsers slice
4. Add tests with 100% coverage
5. Update documentation

### Adding a New Format Parser

1. Create package in `internal/format/<format>/`
2. Implement `format.Parser` interface:
   ```go
   type Parser interface {
       Detect(peek []byte) bool
       Parse(r *bufio.Reader) ([]RawBlock, error)
   }
   ```
3. Register in `extractor.go` formatParsers slice
4. Add tests with 100% coverage
5. Update documentation

## Critical Constraints

### No External Dependencies

The library uses only the Go standard library. This is a core design principle.

### Never Panic

Always return errors instead of panicking. Use sentinel errors where appropriate:

```go
var ErrTruncatedData = errors.New("imx: truncated data")

func parse(data []byte) error {
    if len(data) < minSize {
        return ErrTruncatedData
    }
    // ...
}
```

### Streaming Only

Use `bufio.Reader` for parsing. Never load entire files into memory:

```go
// Good
func Parse(r *bufio.Reader) ([]Block, error)

// Bad
func Parse(data []byte) ([]Block, error)
```

### Validate Sizes

Always validate sizes before allocating to prevent attacks:

```go
if size > MaxAllowedSize {
    return nil, fmt.Errorf("size %d exceeds maximum %d", size, MaxAllowedSize)
}
data := make([]byte, size)
```

### Concurrent Safety

The `Extractor` type clones configuration per call and is safe for concurrent use.

## Benchmarking

### Running Benchmarks

```bash
make bench              # Quick benchmark run
make bench-all          # Detailed output (3s timing, saved to bench.txt)
make bench-report       # Generate formatted report
make bench-compare OLD=commit1 NEW=commit2  # Compare commits
```

### Continuous Benchmarking

Every commit to `main` and every PR automatically runs benchmarks via GitHub Actions:
- Results stored in `gh-pages` branch
- Visualized at `https://gomantics.github.io/imx/benchmarks`
- Regression detection (>120% slowdown alerts on PRs)

### Comparing Performance

```bash
# Compare your changes vs main
make bench-compare OLD=main NEW=your-branch

# Or manually with benchstat
go test -bench=. -benchmem ./... > new.txt
git stash
git checkout main
go test -bench=. -benchmem ./... > old.txt
benchstat old.txt new.txt
```

### Benchmark Coverage

38 benchmarks covering:
- High-level operations (MetadataFromFile, Tag lookup, etc.)
- Parser performance (EXIF, IPTC, XMP, ICC)
- Format parsing (JPEG)
- Memory allocations

### Performance Goals

| Operation | Target | Current |
|-----------|--------|---------|
| MetadataFromFile | <300μs | 279μs ✅ |
| Tag lookup | <10ns | 7ns ✅ |
| EXIF parsing | <10μs | 6.9μs ✅ |

## Make Targets

```bash
make help          # Show all targets

make check         # Run lint, test, and build (default)
make build         # Build library, CLI, and examples
make test          # Run all tests with race detector
make test-lib      # Run library tests with coverage
make lint          # Run go vet
make fmt           # Format code
make clean         # Remove build artifacts
make install       # Install CLI to GOPATH/bin
make coverage      # Show coverage report
make coverage-html # Generate HTML coverage report
make bench         # Run benchmarks
make bench-all     # Run benchmarks with detailed output
make bench-report  # Generate benchmark report
make bench-compare # Compare benchmarks between commits
make example       # Build and run example
```

## CLI Tool

The `imx` CLI provides a rich interface for metadata extraction:

```bash
# Build CLI
make build

# Or install globally
make install

# Usage examples
imx photo.jpg                    # Show all metadata
imx -S photo.jpg                 # Quick summary
imx --json photo.jpg             # JSON output
imx --get=Make photo.jpg         # Get specific tag
imx --spec=exif photo.jpg        # Filter by spec
imx https://example.com/img.jpg  # From URL
imx -r --stats ./photos/         # Batch with stats
```

Run `imx --help` for full documentation.

## Questions?

- Open an issue for bugs or feature requests
- Submit a PR for contributions

## License

MIT License - see [LICENSE](LICENSE) for details.

