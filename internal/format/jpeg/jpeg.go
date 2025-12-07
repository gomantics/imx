package jpeg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/common"
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
func (p *Parser) Parse(r *bufio.Reader) ([]common.RawBlock, error) {
	var blocks []common.RawBlock
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
			if bytes.HasPrefix(data, common.MagicEXIF) {
				blocks = append(blocks, common.RawBlock{
					Spec:    common.SpecEXIF,
					Payload: data[len(common.MagicEXIF):], // Skip "Exif\x00\x00"
					Origin:  "APP1 Exif",
					Format:  common.FormatJPEG,
					Index:   exifIndex,
				})
				exifIndex++
			} else if bytes.HasPrefix(data, common.MagicXMP) {
				blocks = append(blocks, common.RawBlock{
					Spec:    common.SpecXMP,
					Payload: data[len(common.MagicXMP):], // Skip XMP namespace
					Origin:  "APP1 XMP",
					Format:  common.FormatJPEG,
					Index:   xmpIndex,
				})
				xmpIndex++
			}

		case markerAPP2:
			// APP2 can contain ICC profiles
			if bytes.HasPrefix(data, common.MagicICC) {
				blocks = append(blocks, common.RawBlock{
					Spec:    common.SpecICC,
					Payload: data[len(common.MagicICC):], // Skip "ICC_PROFILE\x00"
					Origin:  "APP2 ICC",
					Format:  common.FormatJPEG,
					Index:   iccIndex,
				})
				iccIndex++
			}

		case markerAPP13:
			// APP13 contains IPTC/Photoshop data
			if bytes.HasPrefix(data, common.MagicIPTC) {
				blocks = append(blocks, common.RawBlock{
					Spec:    common.SpecIPTC,
					Payload: data[len(common.MagicIPTC):], // Skip "Photoshop 3.0\x00"
					Origin:  "APP13 IPTC",
					Format:  common.FormatJPEG,
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
