# Roadmap

This document outlines the current status and future plans for imx.

## Current Release: v0.1.0

### Supported Features

**Formats:**
- ✅ JPEG - Full support

**Metadata Specs:**
- ✅ EXIF - Full support including GPS, orientation, camera settings
- ✅ IPTC - Full support for IIM datasets
- ✅ XMP - Core namespace support (dc, photoshop, etc.)
- ✅ ICC - Color profile extraction and parsing

**API:**
- ✅ Convenience functions (MetadataFromFile, MetadataFromBytes, etc.)
- ✅ Configurable Extractor with options
- ✅ Tag iteration (Each, EachInSpec)
- ✅ Batch retrieval (GetAll)
- ✅ Tag constants for common fields

**CLI:**
- ✅ Multiple output formats (text, JSON, CSV, table)
- ✅ Filtering by spec and tag
- ✅ Recursive directory processing
- ✅ URL support
- ✅ Stdin support

**Quality:**
- ✅ 100% test coverage
- ✅ Golden tests against ExifTool
- ✅ Zero dependencies (stdlib only)
- ✅ Memory-efficient streaming I/O

---

## Planned: v0.2.0

### Additional Formats

- 🔜 **PNG** - eXIf chunk, iTXt (XMP), iCCP (ICC profile)
- 🔜 **WebP** - EXIF, XMP, ICCP chunks

### Improvements

- 🔜 Fuzz testing for all parsers
- 🔜 Benchmarks and performance optimizations

---

## Planned: v0.3.0

### Additional Formats

- 🔜 **TIFF** - IFD-based metadata extraction
- 🔜 **HEIF/HEIC** - Apple's modern image format
- 🔜 **AVIF** - AV1 Image Format

---

## Future Considerations

### Potential Features

- **Human-Readable Value Conversions** - Convert raw EXIF values to human-readable formats
  - APEX values (ShutterSpeedValue, ApertureValue, MaxApertureValue) → "1/50", "f/9.0", etc.
  - Enum values (ResolutionUnit, ExposureProgram, MeteringMode, etc.) → "inches", "Aperture-priority AE", etc.
  - GPS coordinates → Decimal degrees format
  - Fraction values (ExposureTime, ExposureCompensation) → Formatted strings
- **Maker Notes** - Decode manufacturer-specific data (Nikon, Canon, Sony, etc.)
- **Thumbnail Extraction** - Extract embedded preview images
- **Metadata Writing** - Modify and write metadata back to files
- **Streaming Export** - Export metadata as it's parsed for large files

### Out of Scope

- Video metadata (MP4, MOV, etc.)
- Audio metadata (MP3, FLAC, etc.)
- Document metadata (PDF, Office, etc.)

---

## Contributing

We welcome contributions! Priority areas:

1. PNG format support
2. WebP format support
3. Additional test images for edge cases
4. Performance improvements

See [CLAUDE.md](CLAUDE.md) for development guidelines.

