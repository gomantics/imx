# AI Development Assistant Guide

This document provides guidance for AI assistants (like Claude) working on this project.

## Important Files to Always Reference

Before making any changes, always read and follow:

1. **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development guidelines, commit message format, benchmarking, testing requirements
2. This file (CLAUDE.md) - AI-specific instructions

## Key Principles

### 1. Commit Messages as Changelog
- Commit messages serve as the project's changelog
- Follow the detailed format in CONTRIBUTING.md
- Include what, why, and impact of changes
- Mark breaking changes with `BREAKING:` prefix

### 2. No Standalone Documentation Files
- **DO NOT** create separate CHANGELOG.md, docs/BENCHMARKING.md, etc.
- Documentation belongs in: README.md, CONTRIBUTING.md, or Makefile comments
- Usage instructions go in CONTRIBUTING.md
- Quick reference goes in Makefile help

### 3. Testing & Coverage
- Maintain 100% test coverage for library code
- Run `make test-lib` before committing
- All benchmarks must pass: `make bench`

### 4. Code Quality
- Zero external dependencies (stdlib only)
- Never panic, return errors
- Use streaming I/O (`bufio.Reader`)
- Validate sizes before allocating

## Workflow

### Before Starting Work
1. Read CONTRIBUTING.md for current guidelines
2. Check existing code patterns
3. Understand the three-layer architecture (Format → Meta → API)

### When Making Changes
1. Write detailed commit messages (see CONTRIBUTING.md examples)
2. Run `make check` (lint, test, build)
3. Run `make bench` if performance-related
4. Update README.md only if API changes

### When Adding Features
1. Add tests first (TDD when possible)
2. Maintain 100% coverage
3. Add benchmarks for performance-critical paths
4. Update CONTRIBUTING.md if it adds new development workflows

## Project Structure Reference

```
imx/
├── api.go, config.go, extractor.go, types.go  # Public API
├── internal/
│   ├── common/           # Shared types (Spec, TagID, etc.)
│   ├── format/           # Container format parsers
│   │   └── jpeg/
│   └── meta/             # Metadata parsers
│       ├── exif/
│       ├── iptc/
│       ├── xmp/
│       └── icc/
├── cmd/imx/              # CLI tool
├── examples/             # Usage examples
├── scripts/              # Utility scripts (benchmarking, etc.)
├── testdata/goldens/     # Test images
├── CONTRIBUTING.md       # **READ THIS FIRST**
├── README.md             # User-facing documentation
└── Makefile              # Build automation + usage help
```

## Common Tasks

### Running Tests
```bash
make test       # All tests with race detector
make test-lib   # Library tests with coverage
make coverage   # Check coverage (must be 100%)
```

### Benchmarking
```bash
make bench                              # Quick run
make bench-compare OLD=main NEW=HEAD    # Compare changes
```

### Before Committing
```bash
make check      # Lint, test, and build
```

## Anti-Patterns to Avoid

❌ Creating CHANGELOG.md (use detailed commit messages instead)
❌ Creating docs/ directory with separate documentation files
❌ Adding external dependencies
❌ Panicking instead of returning errors
❌ Loading entire files into memory
❌ Vague commit messages ("fix bug", "update code")
❌ Breaking changes without `BREAKING:` in commit message

## Questions?

- Check CONTRIBUTING.md first
- Look at existing code for patterns
- When in doubt, ask the user
