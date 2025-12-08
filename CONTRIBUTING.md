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
├── *.go                # Public API (api.go, config.go, extractor.go, types.go, tags.go)
├── cmd/imx/            # CLI tool
├── examples/           # Usage examples (basic & advanced)
├── internal/
│   ├── format/         # Container format parsers (JPEG, etc.)
│   └── meta/           # Metadata parsers (EXIF, IPTC, XMP, ICC)
├── testdata/goldens/   # Test images with expected metadata
└── Makefile            # Build automation
```

### Architecture

Three-layer pipeline:
1. **Format Layer** - Extracts raw metadata blocks from container formats
2. **Meta Layer** - Parses raw blocks into structured tags
3. **API Layer** - Provides user-facing types and functions

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

**Metadata Parser:**
1. Create package in `internal/meta/<spec>/`
2. Implement `meta.Parser` interface
3. Register in `extractor.go`
4. Add tests with 100% coverage

**Format Parser:**
1. Create package in `internal/format/<format>/`
2. Implement `format.Parser` interface
3. Register in `extractor.go`
4. Add tests with 100% coverage

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
