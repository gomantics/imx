package tiff

import (
	"encoding/binary"
	"fmt"
	"io"

	imxbin "github.com/gomantics/imx/internal/binary"
	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/icc"
	"github.com/gomantics/imx/internal/parser/iptc"
	"github.com/gomantics/imx/internal/parser/tiff/makernote"
	"github.com/gomantics/imx/internal/parser/tiff/makernote/canon"
	"github.com/gomantics/imx/internal/parser/tiff/makernote/fujifilm"
	"github.com/gomantics/imx/internal/parser/tiff/makernote/nikon"
	"github.com/gomantics/imx/internal/parser/tiff/makernote/sony"
	"github.com/gomantics/imx/internal/parser/xmp"
)

// Parser parses TIFF files and TIFF-based raw formats.
//
// Supported formats:
//   - TIFF (Tagged Image File Format) - standard image format
//   - DNG (Digital Negative) - Adobe's open raw format
//   - NEF, NRW (Nikon Electronic Format) - Nikon raw files
//   - ARW, SRF, SR2 (Sony Alpha Raw) - Sony raw files
//   - ORF (Olympus Raw Format) - Olympus raw files
//   - PEF (Pentax Electronic Format) - Pentax raw files
//   - RW2 (Panasonic Raw 2) - Panasonic raw files
//   - SRW (Samsung Raw) - Samsung raw files
//   - RWL (Leica Raw) - Leica raw files
//   - ERF (Epson Raw File) - Epson raw files
//   - 3FR (Hasselblad 3F Raw) - Hasselblad raw files
//   - DCR, KDC, K25 (Kodak Digital Camera Raw) - Kodak raw files
//   - MRW (Minolta Raw) - Minolta raw files
//   - IIQ (Phase One Intelligent Image Quality) - Phase One raw files
//   - MEF (Mamiya Raw Format) - Mamiya raw files
//   - MOS (Leaf Raw) - Leaf raw files
type Parser struct {
	icc       *icc.Parser
	iptc      *iptc.Parser
	xmp       *xmp.Parser
	makernote *makernote.Registry
}

// New creates a new TIFF parser
func New() *Parser {
	registry := makernote.NewRegistry()
	// Register MakerNote handlers in priority order (most specific first)
	// Canon must be last (has no header, is fallback detection)
	registry.Register(nikon.New())
	registry.Register(sony.New())
	registry.Register(fujifilm.New())
	registry.Register(canon.New()) // Must be last - no header, fallback

	return &Parser{
		icc:       icc.New(),
		iptc:      iptc.New(),
		xmp:       xmp.New(),
		makernote: registry,
	}
}

// Name returns the parser name
func (p *Parser) Name() string {
	return "TIFF"
}

// Detect checks if the data is a TIFF file
func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, tiffHeaderPrefixSize)
	_, err := r.ReadAt(buf[:tiffHeaderPrefixSize], 0)
	if err != nil {
		return false
	}

	// Check for TIFF byte order markers and magic number
	// II (little-endian) or MM (big-endian) followed by 42
	return (buf[0] == byteOrderLittleEndian && buf[1] == byteOrderLittleEndian && buf[2] == tiffMagicNumber && buf[3] == 0) ||
		(buf[0] == byteOrderBigEndian && buf[1] == byteOrderBigEndian && buf[2] == 0 && buf[3] == tiffMagicNumber)
}

// Parse extracts metadata from a TIFF file
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()
	var dirs []parser.Directory

	// Embedded metadata directories (collected locally for thread safety)
	var iccDirs, iptcDirs, xmpDirs, makernoteDirs []parser.Directory

	// Read header to determine byte order
	headerBuf := make([]byte, tiffHeaderSize)
	_, err := r.ReadAt(headerBuf[:tiffHeaderSize], 0)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read TIFF header: %w", err))
		return nil, parseErr
	}

	// Determine byte order
	var order binary.ByteOrder
	if headerBuf[0] == byteOrderLittleEndian && headerBuf[1] == byteOrderLittleEndian {
		order = binary.LittleEndian
	} else if headerBuf[0] == byteOrderBigEndian && headerBuf[1] == byteOrderBigEndian {
		order = binary.BigEndian
	} else {
		parseErr.Add(fmt.Errorf("invalid TIFF byte order: %c%c", headerBuf[0], headerBuf[1]))
		return nil, parseErr
	}

	// Verify magic number
	magic := order.Uint16(headerBuf[2:4])
	if magic != tiffMagicNumber {
		parseErr.Add(fmt.Errorf("invalid TIFF magic number: %d (expected %d)", magic, tiffMagicNumber))
		return nil, parseErr
	}

	// Get offset to first IFD
	ifd0Offset := int64(order.Uint32(headerBuf[4:8]))

	// Create reader with byte order
	reader := imxbin.NewReader(r, order)

	// Parse IFD0
	ifd0Dir, ifd0Err, subIFDs, numEntries := p.parseIFD(reader, r, ifd0Offset, "IFD0", &iccDirs, &iptcDirs, &xmpDirs, &makernoteDirs, parseErr)
	if ifd0Err != nil {
		parseErr.Merge(ifd0Err)
	}
	if ifd0Dir != nil && len(ifd0Dir.Tags) > 0 {
		dirs = append(dirs, *ifd0Dir)
	}

	// Parse sub-IFDs (EXIF, GPS, Interoperability, SubIFDs for RAW previews)
	for _, sub := range subIFDs {
		subDir, subErr, _, _ := p.parseIFD(reader, r, sub.Offset, sub.Name, &iccDirs, &iptcDirs, &xmpDirs, &makernoteDirs, parseErr)
		if subErr != nil {
			parseErr.Merge(subErr)
		}
		if subDir != nil && len(subDir.Tags) > 0 {
			dirs = append(dirs, *subDir)
		}
	}

	// Read next IFD offset from IFD0 (for IFD1, typically thumbnail)
	// Offset is after: entry count + numEntries * entry size
	nextIFDOffsetPos := ifd0Offset + ifdEntryCountSize + int64(numEntries)*ifdEntrySize
	nextIFDOffset, err := reader.ReadUint32(nextIFDOffsetPos)
	if err == nil && nextIFDOffset != 0 {
		ifd1Dir, ifd1Err, _, _ := p.parseIFD(reader, r, int64(nextIFDOffset), "IFD1", &iccDirs, &iptcDirs, &xmpDirs, &makernoteDirs, parseErr)
		if ifd1Err != nil {
			parseErr.Merge(ifd1Err)
		}
		if ifd1Dir != nil && len(ifd1Dir.Tags) > 0 {
			dirs = append(dirs, *ifd1Dir)
		}
	}

	// Add embedded metadata directories
	dirs = append(dirs, iccDirs...)
	dirs = append(dirs, iptcDirs...)
	dirs = append(dirs, xmpDirs...)
	dirs = append(dirs, makernoteDirs...)

	return dirs, parseErr.OrNil()
}
