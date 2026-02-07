package makernote

import "encoding/binary"

// OffsetBase determines how tag value offsets are calculated.
type OffsetBase int

const (
	// OffsetAbsolute means offsets are relative to EXIF TIFF header.
	// Used by: Canon, Nikon Type 1/2, Sony
	OffsetAbsolute OffsetBase = iota

	// OffsetRelativeToMakerNote means offsets are relative to MakerNote start.
	// Used by: Fujifilm, Nikon Type 3
	OffsetRelativeToMakerNote
)

// Config holds manufacturer-specific parsing configuration.
// Returned by Handler.Detect() to configure parsing behavior.
type Config struct {
	// IFDOffset is where the IFD starts within the MakerNote data.
	// For Canon: 0 (no header)
	// For Nikon Type 1: 8 (after 8-byte header)
	// For Nikon Type 3: 18 (after 10-byte header + 8-byte TIFF header)
	// For Sony: 12 (after 12-byte header)
	// For Fujifilm: 12 (after 8-byte header + 4-byte offset)
	IFDOffset int64

	// OffsetBase determines how tag value offsets are calculated.
	OffsetBase OffsetBase

	// ByteOrder for parsing. If nil, inherits from parent EXIF.
	// Nikon Type 3 and Fujifilm have embedded byte order.
	ByteOrder binary.ByteOrder

	// HasNextIFD indicates if next-IFD pointer should be followed.
	// Canon CR2 files have non-zero next-IFD that should be ignored.
	HasNextIFD bool

	// Variant identifies the specific format variant (e.g., "Type1", "Type3").
	Variant string
}
