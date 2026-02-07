// Package makernote provides MakerNote parsing for camera manufacturer-specific metadata.
//
// MakerNote is an EXIF tag (0x927C) containing manufacturer-specific metadata.
// Each manufacturer uses a different format, requiring specialized parsers.
//
// Supported manufacturers:
//   - Canon: No header, IFD at offset 0, absolute offsets
//   - Nikon: Type 1 (8-byte header), Type 3 (embedded TIFF)
//   - Sony: 12-byte header, absolute offsets, little-endian
//   - Fujifilm: 8-byte header + offset, relative offsets, little-endian
package makernote

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/gomantics/imx/internal/parser"
)

// Handler defines the interface for manufacturer-specific MakerNote parsers.
type Handler interface {
	// Manufacturer returns the manufacturer name (e.g., "Canon", "Nikon").
	Manufacturer() string

	// Detect checks if this handler can parse the given MakerNote data.
	// Returns true and a Config if the data matches this manufacturer's format.
	Detect(data []byte) (bool, *Config)

	// Parse extracts metadata from the MakerNote.
	// r: reader for the entire file (needed for offset-based reads)
	// makerNoteOffset: absolute offset of MakerNote in file
	// exifBase: absolute offset of EXIF TIFF header (for absolute offset calculation)
	// cfg: parsing configuration from Detect()
	Parse(r io.ReaderAt, makerNoteOffset, exifBase int64, cfg *Config) ([]parser.Tag, *parser.ParseError)

	// TagName returns the human-readable name for a tag ID.
	TagName(tagID uint16) string
}

// Registry manages manufacturer handlers and routes MakerNote data to the appropriate parser.
type Registry struct {
	handlers []Handler
}

// NewRegistry creates a new Registry with no registered handlers.
func NewRegistry() *Registry {
	return &Registry{
		handlers: make([]Handler, 0),
	}
}

// Register adds a handler to the registry.
// Handlers are checked in registration order during detection.
// Register handlers in priority order: most specific first.
//
// Recommended order:
//  1. Nikon Type 3 ('Nikon' + 0x02)
//  2. Nikon Type 1 ('Nikon' + 0x01)
//  3. Sony ('SONY DSC' or 'SONY CAM')
//  4. Fujifilm ('FUJIFILM')
//  5. Canon (no header - validated by IFD structure, must be last)
func (r *Registry) Register(h Handler) {
	r.handlers = append(r.handlers, h)
}

// Detect finds the appropriate handler for the given MakerNote data.
// Returns the handler and its configuration, or nil if no handler matches.
func (r *Registry) Detect(data []byte) (Handler, *Config) {
	for _, h := range r.handlers {
		if ok, cfg := h.Detect(data); ok {
			return h, cfg
		}
	}
	return nil, nil
}

// Parse parses MakerNote data using the appropriate manufacturer handler.
// Returns nil tags (not error) for unknown manufacturers.
func (r *Registry) Parse(reader io.ReaderAt, makerNoteOffset, exifBase int64, data []byte) ([]parser.Tag, *parser.ParseError) {
	handler, cfg := r.Detect(data)
	if handler == nil {
		// Unknown manufacturer - return nil without error
		return nil, nil
	}
	return handler.Parse(reader, makerNoteOffset, exifBase, cfg)
}

// Header detection constants
var (
	// Nikon headers
	nikonHeader     = []byte("Nikon")
	nikonType3Magic = byte(0x02)
	nikonType1Magic = byte(0x01)

	// Sony headers
	sonyDSCHeader = []byte("SONY DSC ")
	sonyCAMHeader = []byte("SONY CAM ")

	// Fujifilm header
	fujifilmHeader = []byte("FUJIFILM")
)

// DetectNikonType3 checks for Nikon Type 3 format.
// Header: 'Nikon' + 0x00 + 0x02 + 0x00 + 0x00 + embedded TIFF header
func DetectNikonType3(data []byte) (bool, *Config) {
	if len(data) < 18 {
		return false, nil
	}

	// Check 'Nikon' + 0x00 + 0x02
	if !bytes.HasPrefix(data, nikonHeader) {
		return false, nil
	}
	if data[5] != 0x00 || data[6] != nikonType3Magic {
		return false, nil
	}

	// Read embedded TIFF header byte order at offset 10
	var order binary.ByteOrder
	if data[10] == 'I' && data[11] == 'I' {
		order = binary.LittleEndian
	} else if data[10] == 'M' && data[11] == 'M' {
		order = binary.BigEndian
	} else {
		return false, nil
	}

	return true, &Config{
		IFDOffset:  18, // 10-byte header + 8-byte TIFF header
		OffsetBase: OffsetRelativeToMakerNote,
		ByteOrder:  order,
		HasNextIFD: false,
		Variant:    "Type3",
	}
}

// DetectNikonType1 checks for Nikon Type 1 format.
// Header: 'Nikon' + 0x00 + 0x01 + 0x00
func DetectNikonType1(data []byte) (bool, *Config) {
	if len(data) < 8 {
		return false, nil
	}

	// Check 'Nikon' + 0x00 + 0x01
	if !bytes.HasPrefix(data, nikonHeader) {
		return false, nil
	}
	if data[5] != 0x00 || data[6] != nikonType1Magic {
		return false, nil
	}

	return true, &Config{
		IFDOffset:  8,
		OffsetBase: OffsetAbsolute,
		ByteOrder:  nil, // Inherit from parent
		HasNextIFD: false,
		Variant:    "Type1",
	}
}

// DetectSony checks for Sony format.
// Header: 'SONY DSC ' or 'SONY CAM ' (12 bytes)
func DetectSony(data []byte) (bool, *Config) {
	if len(data) < 12 {
		return false, nil
	}

	if !bytes.HasPrefix(data, sonyDSCHeader) && !bytes.HasPrefix(data, sonyCAMHeader) {
		return false, nil
	}

	return true, &Config{
		IFDOffset:  12,
		OffsetBase: OffsetAbsolute,
		ByteOrder:  binary.LittleEndian,
		HasNextIFD: false,
		Variant:    "Standard",
	}
}

// DetectFujifilm checks for Fujifilm format.
// Header: 'FUJIFILM' (8 bytes) + 4-byte IFD offset (little-endian)
func DetectFujifilm(data []byte) (bool, *Config) {
	if len(data) < 12 {
		return false, nil
	}

	if !bytes.HasPrefix(data, fujifilmHeader) {
		return false, nil
	}

	// Read IFD offset at bytes 8-11 (little-endian)
	ifdOffset := int64(binary.LittleEndian.Uint32(data[8:12]))

	return true, &Config{
		IFDOffset:  ifdOffset,
		OffsetBase: OffsetRelativeToMakerNote,
		ByteOrder:  binary.LittleEndian,
		HasNextIFD: false,
		Variant:    "Standard",
	}
}

// DetectCanon checks for Canon format.
// Canon has no header - the MakerNote starts directly with an IFD.
// Detection: first 2 bytes form a valid entry count (1-100).
// This should be called LAST as a fallback.
func DetectCanon(data []byte) (bool, *Config) {
	if len(data) < 14 { // Minimum: 2-byte count + one 12-byte entry
		return false, nil
	}

	// First 2 bytes are entry count (little-endian assumed, validated later)
	entryCount := binary.LittleEndian.Uint16(data[0:2])

	// Sanity check: reasonable entry count (1-100)
	if entryCount < 1 || entryCount > 100 {
		return false, nil
	}

	// Additional validation: check if data is large enough for entries
	requiredSize := 2 + int(entryCount)*12
	if len(data) < requiredSize {
		return false, nil
	}

	return true, &Config{
		IFDOffset:  0,
		OffsetBase: OffsetAbsolute,
		ByteOrder:  nil, // Inherit from parent
		HasNextIFD: false,
		Variant:    "Standard",
	}
}
