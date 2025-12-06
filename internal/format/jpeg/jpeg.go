package jpeg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/format"
	"github.com/gomantics/imx/internal/types"
)

// JPEG marker constants
const (
	markerSOI   = 0xD8 // Start of Image
	markerEOI   = 0xD9 // End of Image
	markerSOS   = 0xDA // Start of Scan
	markerAPP0  = 0xE0 // APP0
	markerAPP1  = 0xE1 // APP1 (EXIF, XMP)
	markerAPP2  = 0xE2 // APP2 (ICC, FlashPix)
	markerAPP13 = 0xED // APP13 (IPTC, Photoshop)
)

// Magic bytes for identifying metadata types
var (
	exifMagic = []byte("Exif\x00\x00")
	xmpMagic  = []byte("http://ns.adobe.com/xap/1.0/\x00")
	iccMagic  = []byte("ICC_PROFILE\x00")
	iptcMagic = []byte("Photoshop 3.0\x00")
)

// Parser implements format.Parser for JPEG
type Parser struct{}

// New creates a JPEG parser
func New() *Parser {
	return &Parser{}
}

// Detect checks if the data is a JPEG file
func (p *Parser) Detect(peek []byte) bool {
	// JPEG starts with SOI marker: 0xFF 0xD8
	return len(peek) >= 2 && peek[0] == 0xFF && peek[1] == markerSOI
}

// Parse extracts metadata blocks from a JPEG file
func (p *Parser) Parse(r *bufio.Reader) ([]format.RawBlock, error) {
	var blocks []format.RawBlock
	exifIndex := 0
	xmpIndex := 0
	iccIndex := 0
	iptcIndex := 0

	// Read SOI marker
	marker, err := readMarker(r)
	if err != nil {
		return nil, fmt.Errorf("read SOI marker: %w", err)
	}
	if marker != markerSOI {
		return nil, fmt.Errorf("expected SOI marker, got 0x%02X", marker)
	}

	// Process markers until SOS (Start of Scan) or EOI
	for {
		marker, err := readMarker(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read marker: %w", err)
		}

		// SOS means we've reached image data - no more metadata
		if marker == markerSOS || marker == markerEOI {
			break
		}

		// Read segment length (2 bytes, big-endian, includes length itself)
		var length uint16
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, fmt.Errorf("read segment length: %w", err)
		}

		if length < 2 {
			return nil, fmt.Errorf("invalid segment length: %d", length)
		}

		// Read segment data (length includes the 2 bytes for length itself)
		dataLen := int(length) - 2
		data := make([]byte, dataLen)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, fmt.Errorf("read segment data: %w", err)
		}

		// Parse APP markers for metadata
		switch marker {
		case markerAPP1:
			// APP1 can contain EXIF or XMP
			if bytes.HasPrefix(data, exifMagic) {
				blocks = append(blocks, format.RawBlock{
					Spec:    types.SpecEXIF,
					Payload: data[len(exifMagic):], // Skip "Exif\x00\x00"
					Origin:  "APP1 Exif",
					Format:  types.FormatJPEG,
					Index:   exifIndex,
				})
				exifIndex++
			} else if bytes.HasPrefix(data, xmpMagic) {
				blocks = append(blocks, format.RawBlock{
					Spec:    types.SpecXMP,
					Payload: data[len(xmpMagic):], // Skip XMP namespace
					Origin:  "APP1 XMP",
					Format:  types.FormatJPEG,
					Index:   xmpIndex,
				})
				xmpIndex++
			}

		case markerAPP2:
			// APP2 can contain ICC profiles
			if bytes.HasPrefix(data, iccMagic) {
				blocks = append(blocks, format.RawBlock{
					Spec:    types.SpecICC,
					Payload: data[len(iccMagic):], // Skip "ICC_PROFILE\x00"
					Origin:  "APP2 ICC",
					Format:  types.FormatJPEG,
					Index:   iccIndex,
				})
				iccIndex++
			}

		case markerAPP13:
			// APP13 contains IPTC/Photoshop data
			if bytes.HasPrefix(data, iptcMagic) {
				blocks = append(blocks, format.RawBlock{
					Spec:    types.SpecIPTC,
					Payload: data[len(iptcMagic):], // Skip "Photoshop 3.0\x00"
					Origin:  "APP13 IPTC",
					Format:  types.FormatJPEG,
					Index:   iptcIndex,
				})
				iptcIndex++
			}
		}
	}

	return blocks, nil
}

// readMarker reads a JPEG marker (0xFF followed by marker byte)
func readMarker(r *bufio.Reader) (byte, error) {
	// Read 0xFF
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	// Skip any padding 0xFF bytes
	for b == 0xFF {
		b, err = r.ReadByte()
		if err != nil {
			return 0, err
		}
	}

	return b, nil
}
