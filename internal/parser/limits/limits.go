package limits

// Shared safety limits across parsers. These are conservative defaults intended
// to prevent unbounded allocations and excessive scanning while remaining large
// enough for typical real-world files.
const (
	// Generic scan caps
	MaxScanBytes = 100 * 1024 * 1024 // 100MB generic scan limit

	// JPEG
	MaxJPEGSegmentSize = 10 * 1024 * 1024  // 10MB per APP segment
	MaxJPEGScanBytes   = 100 * 1024 * 1024 // 100MB total scan

	// PNG
	MaxPNGChunkSize           = 10 * 1024 * 1024  // 10MB max chunk size
	MaxPNGDecompressedTextLen = 2 * 1024 * 1024   // 2MB decompressed text
	MaxPNGICCProfileLen       = 16 * 1024 * 1024  // 16MB ICC profile

	// WebP (RIFF-based)
	MaxWebPChunkSize = 50 * 1024 * 1024  // 50MB per chunk
	MaxWebPFileSize  = 200 * 1024 * 1024 // 200MB RIFF size cap

	// MP4
	MaxMP4AtomSize     = 100 * 1024 * 1024 // 100MB per atom
	MaxMP4MetadataSize = 1 * 1024 * 1024   // 1MB metadata payload

	// HEIC
	MaxHEICBoxSize = 100 * 1024 * 1024 // 100MB per box

	// TIFF
	MaxTIFFTagDataSize = 50 * 1024 * 1024 // 50MB per tag value

	// IPTC
	MaxIPTCDatasetSize = 10 * 1024 * 1024 // 10MB per dataset

	// XMP
	MaxXMPDepth     = 64              // Max XML nesting depth
	MaxXMPTextBytes = 2 * 1024 * 1024 // Max accumulated text per node
)
