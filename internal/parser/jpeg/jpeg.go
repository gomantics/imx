package jpeg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/icc"
	"github.com/gomantics/imx/internal/parser/iptc"
	"github.com/gomantics/imx/internal/parser/limits"
	"github.com/gomantics/imx/internal/parser/tiff"
	"github.com/gomantics/imx/internal/parser/xmp"
)

// Parser parses JPEG files.
//
// The parser is stateless and safe for concurrent use.
type Parser struct {
	icc  *icc.Parser
	iptc *iptc.Parser
	xmp  *xmp.Parser
	exif *tiff.Parser
}

// New creates a new JPEG parser.
func New() *Parser {
	return &Parser{
		icc:  icc.New(),
		iptc: iptc.New(),
		xmp:  xmp.New(),
		exif: tiff.New(),
	}
}

// Name returns the parser name.
func (p *Parser) Name() string {
	return "JPEG"
}

// Detect checks if the data is a JPEG file by looking for SOI marker.
func (p *Parser) Detect(r io.ReaderAt) bool {
	var buf [2]byte
	_, err := r.ReadAt(buf[:], 0)
	return err == nil && buf[0] == markerPrefix && buf[1] == markerSOI
}

// Parse extracts metadata directories from a JPEG file.
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()
	var dirs []parser.Directory
	var pos int64
	var iccChunks map[int][]byte
	var totalChunks int // Expected total ICC chunks from header

	// Read and verify SOI marker
	marker, newPos, err := readMarker(r, pos)
	pos = newPos
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read SOI marker: %w", err))
		return nil, parseErr
	}
	if marker != markerSOI {
		parseErr.Add(fmt.Errorf("expected SOI marker (0xFF 0x%02X), got 0xFF 0x%02X", markerSOI, marker))
		return nil, parseErr
	}

	// Process segments until we hit SOS (image data) or EOI
	for {
		if limits.MaxJPEGScanBytes > 0 && pos > int64(limits.MaxJPEGScanBytes) {
			break
		}

		marker, newPos, err := readMarker(r, pos)
		pos = newPos
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			parseErr.Add(fmt.Errorf("failed to read marker at offset %d: %w", pos-2, err))
			break
		}

		// Stop at image data or end of image
		if marker == markerSOS || marker == markerEOI {
			break
		}

		// Read segment length (2 bytes, big-endian, includes length field itself)
		length, newPos, err := readUint16(r, pos)
		pos = newPos
		if err != nil {
			parseErr.Add(fmt.Errorf("failed to read segment length at offset %d: %w", pos-2, err))
			break
		}

		if length < 2 {
			parseErr.Add(fmt.Errorf("invalid segment length %d at offset %d", length, pos-2))
			break
		}

		// Extract metadata based on marker type
		segmentStart := pos
		segmentSize := int64(length - 2)
		if limits.MaxJPEGSegmentSize > 0 && segmentSize > int64(limits.MaxJPEGSegmentSize) {
			parseErr.Add(fmt.Errorf("segment length %d exceeds limit %d", segmentSize, limits.MaxJPEGSegmentSize))
			break
		}

		pos += segmentSize

		switch marker {
		case markerAPP1:
			dirs = append(dirs, p.parseAPP1(r, segmentStart, segmentSize)...)
		case markerAPP2:
			parseErr.Merge(p.parseAPP2(r, segmentStart, segmentSize, &iccChunks, &totalChunks))
		case markerAPP13:
			appDirs, appErr := p.parseAPP13(r, segmentStart, segmentSize)
			dirs = append(dirs, appDirs...)
			parseErr.Merge(appErr)
		}
	}

	// Parse ICC profile if we collected chunks
	iccDirs, iccErr := p.parseICC(iccChunks, totalChunks)
	dirs = append(dirs, iccDirs...)
	parseErr.Merge(iccErr)

	// Return results
	return dirs, parseErr.OrNil()
}

// readMarker reads a JPEG marker (0xFF followed by marker byte).
// Returns the marker byte and new position.
func readMarker(r io.ReaderAt, pos int64) (byte, int64, error) {
	buf := make([]byte, 2)
	_, err := r.ReadAt(buf, pos)
	if err != nil {
		return 0, pos, err
	}

	// First byte must be 0xFF
	if buf[0] != markerPrefix {
		return 0, pos, fmt.Errorf("expected marker prefix 0xFF, got 0x%02X", buf[0])
	}

	// Skip padding 0xFF bytes (some encoders add extra 0xFF)
	marker := buf[1]
	pos += 2

	for marker == markerPrefix {
		_, err := r.ReadAt(buf[:1], pos)
		if err != nil {
			return 0, pos, err
		}
		marker = buf[0]
		pos++
	}

	return marker, pos, nil
}

// readUint16 reads a big-endian uint16.
// Returns the value and new position.
func readUint16(r io.ReaderAt, pos int64) (uint16, int64, error) {
	buf := make([]byte, 2)
	_, err := r.ReadAt(buf, pos)
	if err != nil {
		return 0, pos, err
	}
	pos += 2
	return binary.BigEndian.Uint16(buf), pos, nil
}

// parseAPP1 extracts EXIF or XMP data from APP1 segment.
func (p *Parser) parseAPP1(r io.ReaderAt, segmentStart, segmentSize int64) []parser.Directory {
	// Check for EXIF identifier
	buf := make([]byte, len(identEXIF))
	_, err := r.ReadAt(buf, segmentStart)
	if err == nil && bytes.Equal(buf, identEXIF) {
		// Create section reader for data after the identifier
		dataStart := segmentStart + int64(len(identEXIF))
		dataSize := segmentSize - int64(len(identEXIF))
		section := io.NewSectionReader(r, dataStart, dataSize)

		// Parse EXIF using TIFF parser (EXIF is TIFF format)
		dirs, _ := p.exif.Parse(section)
		return dirs
	}

	// Check for XMP identifier
	buf = make([]byte, len(identXMP))
	_, err = r.ReadAt(buf, segmentStart)
	if err == nil && bytes.Equal(buf, identXMP) {
		dataStart := segmentStart + int64(len(identXMP))
		dataSize := segmentSize - int64(len(identXMP))
		section := io.NewSectionReader(r, dataStart, dataSize)
		dirs, _ := p.xmp.Parse(section)
		return dirs
	}

	return nil
}

// parseAPP2 extracts ICC profile chunks from APP2 segment.
func (p *Parser) parseAPP2(r io.ReaderAt, segmentStart, segmentSize int64, iccChunks *map[int][]byte, totalChunks *int) *parser.ParseError {
	// Check for ICC identifier
	buf := make([]byte, len(identICC))
	_, err := r.ReadAt(buf, segmentStart)
	if err != nil || !bytes.Equal(buf, identICC) {
		return nil
	}

	// Move past the identifier
	dataStart := segmentStart + int64(len(identICC))
	dataSize := segmentSize - int64(len(identICC))

	// Read chunk header (2 bytes: chunk number, total chunks)
	chunkHeader := make([]byte, 2)
	_, err = r.ReadAt(chunkHeader, dataStart)
	if err != nil {
		return parser.NewParseError(fmt.Errorf("failed to read ICC chunk header at offset %d: %w", dataStart, err))
	}

	chunkNum := int(chunkHeader[0])
	chunkTotal := int(chunkHeader[1])

	// Validate chunk numbers
	if chunkNum == 0 || chunkTotal == 0 || chunkNum > chunkTotal {
		return parser.NewParseError(fmt.Errorf("invalid ICC chunk numbers: %d/%d", chunkNum, chunkTotal))
	}

	// Store expected total chunks (validate all chunks match)
	if *totalChunks == 0 {
		*totalChunks = chunkTotal
	} else if *totalChunks != chunkTotal {
		return parser.NewParseError(fmt.Errorf("inconsistent ICC total chunks: expected %d, got %d", *totalChunks, chunkTotal))
	}

	// Initialize chunks map if needed
	if *iccChunks == nil {
		*iccChunks = make(map[int][]byte, chunkTotal)
	}

	// Read chunk data (after identifier and 2-byte header)
	chunkDataStart := dataStart + 2
	chunkDataSize := dataSize - 2

	if chunkDataSize > 0 {
		chunkData := make([]byte, chunkDataSize)
		_, err = r.ReadAt(chunkData, chunkDataStart)
		if err != nil {
			return parser.NewParseError(fmt.Errorf("failed to read ICC chunk data at offset %d: %w", chunkDataStart, err))
		}
		(*iccChunks)[chunkNum] = chunkData
	}

	return nil
}

// parseICC assembles ICC chunks and parses the complete profile.
func (p *Parser) parseICC(iccChunks map[int][]byte, totalChunks int) ([]parser.Directory, *parser.ParseError) {
	if len(iccChunks) == 0 {
		return nil, nil
	}

	// Validate that we have all expected chunks
	if len(iccChunks) != totalChunks {
		parseErr := parser.NewParseError()
		parseErr.Add(fmt.Errorf("incomplete ICC profile: got %d chunks, expected %d", len(iccChunks), totalChunks))
		return nil, parseErr
	}

	// Assemble chunks in order
	var assembled []byte
	for i := 1; i <= totalChunks; i++ {
		chunkData := iccChunks[i]
		assembled = append(assembled, chunkData...)
	}

	// Create a ReaderAt from the assembled data
	reader := bytes.NewReader(assembled)

	// Parse the complete ICC profile
	return p.icc.Parse(reader)
}

// parseAPP13 extracts IPTC/Photoshop data from APP13 segment.
func (p *Parser) parseAPP13(r io.ReaderAt, segmentStart, segmentSize int64) ([]parser.Directory, *parser.ParseError) {
	// Check for Photoshop identifier
	buf := make([]byte, len(identPhotoshop))
	_, err := r.ReadAt(buf, segmentStart)
	if err == nil && bytes.Equal(buf, identPhotoshop) {
		// Create section reader for data after the identifier
		dataStart := segmentStart + int64(len(identPhotoshop))
		dataSize := segmentSize - int64(len(identPhotoshop))
		section := io.NewSectionReader(r, dataStart, dataSize)
		return p.iptc.Parse(section)
	}

	return nil, nil
}
