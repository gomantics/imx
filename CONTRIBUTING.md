# Contributing to imx

Thank you for considering contributing to imx! This document provides guidelines for contributors.

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
```

## Project Structure

```
imx/
├── *.go                    # Public API (api.go, config.go, extractor.go, types.go, tags.go)
├── cmd/imx/                # CLI tool
│   ├── filter/             # Tag filtering logic
│   ├── output/             # Output formatters (JSON, CSV, Table, Text, Summary)
│   ├── processor/          # File processing
│   ├── ui/                 # CLI interface
│   └── util/               # Utilities
├── examples/               # Usage examples
├── internal/
│   ├── binary/             # Binary reading helpers
│   ├── bufpool/            # Buffer pool for performance
│   ├── parser/             # Unified parser architecture
│   │   ├── cr2/            # Canon RAW parser
│   │   ├── flac/           # FLAC audio parser
│   │   ├── gif/            # GIF parser
│   │   ├── heic/           # HEIC/HEIF parser
│   │   ├── icc/            # ICC profile parser
│   │   ├── id3/            # ID3/MP3 parser
│   │   ├── iptc/           # IPTC metadata parser
│   │   ├── jpeg/           # JPEG parser
│   │   ├── mp4/            # MP4/M4A parser
│   │   ├── png/            # PNG parser
│   │   ├── tiff/           # TIFF parser
│   │   ├── webp/           # WebP parser
│   │   └── xmp/            # XMP parser
│   └── testing/            # Shared test utilities
├── testdata/               # Test files for all formats
│   ├── jpeg/, png/, gif/   # Image formats
│   ├── flac/, mp3/, mp4/   # Audio/video formats
│   └── goldens/            # Expected metadata outputs
└── Makefile                # Build automation
```

### Architecture

**Unified Parser Model**:
- All parsers implement `parser.Parser` interface
- Each parser is stateless and thread-safe
- Uses `io.ReaderAt` for efficient random access
- Returns `[]parser.Directory` with structured tags
- 100% test coverage for all parsers

## Development Guidelines

### Code Style

- Follow standard Go conventions (`go fmt`, `go vet`)
- Add comments for exported types and functions
- Keep functions focused and small

### Testing

```bash
make test          # Run all tests
make coverage      # Show coverage report (target: 100%)
```

### Commit Messages

Commit messages serve as the project's changelog. Write clear, informative messages.

**Format:**
```
type(scope): short description

Detailed explanation:
- What was changed and why
- Breaking changes (BREAKING:)
- Related issues (#123)
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `refactor` - Code restructuring
- `perf` - Performance improvement
- `test` - Add/update tests
- `docs` - Documentation
- `chore` - Maintenance

**Example:**
```
feat(exif): add GPS coordinate parsing

Added support for extracting GPS latitude/longitude from EXIF tags.
Handles both degrees/minutes/seconds and decimal formats.

Closes #45
```

### Adding a New Parser

1. Create package in `internal/parser/<format>/`
2. Implement the `parser.Parser` interface:
   ```go
   type Parser interface {
       Name() string
       Detect(r io.ReaderAt) bool
       Parse(r io.ReaderAt) ([]Directory, *ParseError)
   }
   ```
3. Make parser stateless and thread-safe (no struct fields that store state)
4. Use `io.ReaderAt` for efficient random access
5. Add comprehensive tests:
   - Unit tests for all functions
   - Fuzz tests (`FuzzParser`)
   - Benchmark tests
   - Concurrent access tests
   - Target: 100% test coverage
6. Add constants file if you have 10+ magic numbers
7. Document the format structure in package comments
8. Register parser in the main extractor

## Core Principles

### No External Dependencies

Use only the Go standard library. This is a core design principle.

### Never Panic

Always return errors instead of panicking:

```go
var ErrTruncatedData = errors.New("imx: truncated data")

func parse(data []byte) error {
    if len(data) < minSize {
        return ErrTruncatedData
    }
    // ...
}
```

### Efficient I/O

Use `io.ReaderAt` for parsing. Never load entire files into memory:

```go
// Good - Random access without loading entire file
func Parse(r io.ReaderAt) ([]Directory, *ParseError)

// Bad - Loads entire file into memory
func Parse(data []byte) ([]Directory, *ParseError)
```

**Benefits of `io.ReaderAt`**:
- Random access to any file position
- No memory copying
- Thread-safe for concurrent reads
- Works with files, byte slices, and network streams

### Validate Sizes

Always validate sizes before allocating to prevent attacks:

```go
if size > MaxAllowedSize {
    return nil, fmt.Errorf("size %d exceeds maximum %d", size, MaxAllowedSize)
}
data := make([]byte, size)
```

## Benchmarking

### Running Benchmarks

```bash
make bench
```

### Continuous Benchmarking

Performance is automatically tracked on every commit to `main`:

- **Dashboard**: [https://gomantics.github.io/imx/dev/bench/](https://gomantics.github.io/imx/dev/bench/)
- **Benchmarks run** on every push to main
- **Regression alerts**: Warns if performance degrades >20%
- **Historical trends**: Interactive charts showing performance over time

## Make Targets

```bash
make help          # Show all targets
make check         # Run lint, test, and build
make build         # Build library, CLI, and examples
make test          # Run all tests
make lint          # Run go vet
make clean         # Remove build artifacts
make coverage      # Show coverage report
make bench         # Run benchmarks
```

## Questions?

- Open an issue for bugs or feature requests
- Submit a PR for contributions

## License

MIT License - see [LICENSE](LICENSE) for details.
