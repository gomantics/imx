# CLAUDE.md

Context for Claude Code when working on this repository.

## Project

`imx` - Go library for fast, dependency-free image metadata extraction (EXIF, IPTC, XMP, ICC).

**Read `docs/SPEC.md` for complete specification.**

## Workflow

- **Commit after each iteration** - Keep commits focused and incremental
- **Test before committing** - Run `go build ./...` and test with example
- **Build to bin/** - `go build -o bin/basic ./examples/basic`

## Commands

```bash
go test ./...                    # Run tests
go build ./...                   # Build all packages
go build -o bin/basic ./examples/basic  # Build example
./bin/basic ./testdata/DSC_1631.jpg     # Test extraction
```

## Architecture

**Three-Layer Pipeline:**
1. **Container** (`internal/container/`) - Extract raw metadata blocks from file format
2. **Namespace** (`internal/meta/`) - Parse raw blocks into structured tags
3. **Pipeline** (`internal/pipeline/`) - Route blocks to parsers, assemble Metadata

**Current Structure:**
```
imx/
  api.go, extractor.go, types.go, errors.go
internal/
  container/jpeg/     # JPEG marker parsing
  meta/exif/          # EXIF TIFF/IFD parsing
    exif.go           # Core parsing logic
    tags.go           # Tag name mappings (knownTags, gpsTags)
  pipeline/           # Orchestration
  types/              # Shared internal types
```

## Critical Constraints

- **No dependencies** - stdlib only
- **Never panic** - always return errors
- **Streaming only** - use bufio.Reader, never load entire files
- **Validate sizes** - before allocating (prevent attacks)
- **Concurrent-safe** - Extractor clones config per call

## API Access Patterns

1. **Convenience helpers** - `meta.Make()`, `meta.DateTimeOriginal()`, `meta.GPSCoordinates()`
2. **Batch retrieval** - `meta.GetAll(tagIDs...)` → `map[TagID]any`
3. **Iterators** - `meta.Each(func(dir Directory, tag Tag) bool)`
4. **Direct access** - `meta.Tag(namespace, tagID)`

## Key Implementation Details

**EXIF Tag Mappings:**
- `knownTags` - Main EXIF tags (IFD0, ExifIFD, InteropIFD)
- `gpsTags` - GPS-specific tags (separate due to tag ID conflicts)
- Tag lookup is context-aware based on IFD name

**Error Handling:**
- Wrap errors: `fmt.Errorf("imx: parse jpeg: %w", err)`
- Use sentinel errors: `ErrUnknownFormat`, `ErrTruncatedData`