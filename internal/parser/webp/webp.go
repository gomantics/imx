package webp

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/icc"
	"github.com/gomantics/imx/internal/parser/limits"
	"github.com/gomantics/imx/internal/parser/tiff"
	"github.com/gomantics/imx/internal/parser/xmp"
)

// Parser parses WebP image files.
//
// Supported metadata:
//   - EXIF (EXIF chunk)
//   - XMP (XMP chunk)
//   - ICC Profile (ICCP chunk)
//
// WebP uses RIFF container format.
// Parser is safe for concurrent use by multiple goroutines.
//
// TODO: Extract RIFF parsing into internal/parser/riff/ when adding support
// for other RIFF-based formats (WAV, AVI, etc.). The RIFF parser should handle
// generic container parsing and return a "RIFF" directory (matching exiftool),
// while format-specific parsers (webp, wav, avi) would delegate to it.
type Parser struct {
	tiff *tiff.Parser
	xmp  *xmp.Parser
	icc  *icc.Parser
}

// New creates a new WebP parser
func New() *Parser {
	return &Parser{
		tiff: tiff.New(),
		xmp:  xmp.New(),
		icc:  icc.New(),
	}
}

// Name returns the parser name
func (p *Parser) Name() string {
	return "WebP"
}

// Detect checks if the data is a WebP file
func (p *Parser) Detect(r io.ReaderAt) bool {
	var buf [12]byte
	_, err := r.ReadAt(buf[:], 0)
	if err != nil {
		return false
	}

	// Check for RIFF signature and WEBP form type
	return string(buf[0:4]) == "RIFF" && string(buf[8:12]) == "WEBP"
}

// Parse extracts metadata from a WebP file
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()
	var dirs []parser.Directory

	// Read RIFF header
	var buf [12]byte
	_, err := r.ReadAt(buf[:], 0)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read RIFF header: %w", err))
		return nil, parseErr
	}

	// Verify RIFF signature
	if string(buf[0:4]) != "RIFF" {
		parseErr.Add(fmt.Errorf("invalid RIFF signature"))
		return nil, parseErr
	}

	// Verify WEBP form type
	if string(buf[8:12]) != "WEBP" {
		parseErr.Add(fmt.Errorf("invalid WebP signature"))
		return nil, parseErr
	}

	fileSize := binary.LittleEndian.Uint32(buf[4:8])
	if fileSize > limits.MaxWebPFileSize {
		parseErr.Add(fmt.Errorf("webp: RIFF size %d exceeds limit %d", fileSize, limits.MaxWebPFileSize))
		return nil, parseErr
	}
	pos := int64(12)

	// Create WebP technical directory
	webpDir := &parser.Directory{
		Name: "WebP",
		Tags: []parser.Tag{},
	}

	// Parse chunks
	endPos := int64(fileSize) + 8 // RIFF header is 8 bytes
	for pos < endPos {
		chunk, err := p.readChunk(r, pos)
		if err != nil {
			if err == io.EOF {
				break
			}
			parseErr.Add(err)
			break
		}

		// Process metadata chunks
		switch chunk.FourCC {
		case "VP8 ", "VP8L", "VP8X":
			// Image data chunks - extract technical metadata
			tags := p.parseImageChunk(r, chunk)
			webpDir.Tags = append(webpDir.Tags, tags...)

		case "EXIF":
			// EXIF metadata
			exifDirs := p.parseExifChunk(r, chunk)
			dirs = append(dirs, exifDirs...)

		case "XMP ":
			// XMP metadata
			xmpDirs := p.parseXMPChunk(r, chunk)
			dirs = append(dirs, xmpDirs...)

		case "ICCP":
			// ICC color profile
			iccDirs := p.parseICCPChunk(r, chunk)
			dirs = append(dirs, iccDirs...)
		}

		// Move to next chunk (account for padding)
		pos = chunk.DataOffset + int64(chunk.Size)
		if chunk.Size%2 != 0 {
			pos++ // Skip padding byte
		}

		if limits.MaxScanBytes > 0 && pos > limits.MaxScanBytes {
			break
		}
	}

	// Add WebP directory if it has tags
	if len(webpDir.Tags) > 0 {
		dirs = append([]parser.Directory{*webpDir}, dirs...)
	}

	return dirs, parseErr.OrNil()
}

// Chunk represents a WebP RIFF chunk
type Chunk struct {
	FourCC     string
	Size       uint32
	DataOffset int64
}

// readChunk reads a WebP chunk header at the given position
func (p *Parser) readChunk(r io.ReaderAt, pos int64) (*Chunk, error) {
	var buf [8]byte
	_, err := r.ReadAt(buf[:], pos)
	if err != nil {
		return nil, err
	}

	fourCC := string(buf[0:4])
	size := binary.LittleEndian.Uint32(buf[4:8])

	chunk := &Chunk{
		FourCC:     fourCC,
		Size:       size,
		DataOffset: pos + 8,
	}

	if size > limits.MaxWebPChunkSize {
		return nil, fmt.Errorf("webp: chunk %s size %d exceeds limit %d", fourCC, size, limits.MaxWebPChunkSize)
	}

	return chunk, nil
}

// parseExifChunk parses an EXIF chunk
func (p *Parser) parseExifChunk(r io.ReaderAt, chunk *Chunk) []parser.Directory {
	if chunk.Size < 6 {
		return nil
	}

	// WebP EXIF chunk format:
	// - 4 bytes: "Exif" identifier (sometimes)
	// - Followed by standard TIFF-based EXIF data

	// Check if data starts with "Exif"
	var buf [4]byte
	_, err := r.ReadAt(buf[:], chunk.DataOffset)
	if err != nil {
		return nil
	}

	offset := chunk.DataOffset
	size := int64(chunk.Size)

	// Skip "Exif\x00\x00" header if present
	if string(buf[:4]) == "Exif" {
		offset += 6 // Skip "Exif" + padding
		size -= 6
	} else if buf[0] == 0xFF && buf[1] == 0xD8 {
		// JPEG SOI (Start of Image) marker: 0xFF 0xD8
		// If EXIF chunk starts with JPEG signature, it's malformed - skip it
		return nil
	}

	if size <= 0 {
		return nil
	}

	// Parse EXIF using TIFF parser
	section := io.NewSectionReader(r, offset, size)
	dirs, _ := p.tiff.Parse(section)
	return dirs
}

// parseXMPChunk parses an XMP chunk
func (p *Parser) parseXMPChunk(r io.ReaderAt, chunk *Chunk) []parser.Directory {
	if chunk.Size == 0 {
		return nil
	}

	// XMP chunk contains XMP XML data
	// Use io.NewSectionReader to avoid loading entire chunk into memory
	section := io.NewSectionReader(r, chunk.DataOffset, int64(chunk.Size))
	dirs, _ := p.xmp.Parse(section)
	return dirs
}

// parseICCPChunk parses an ICCP chunk containing ICC profile
func (p *Parser) parseICCPChunk(r io.ReaderAt, chunk *Chunk) []parser.Directory {
	if chunk.Size == 0 {
		return nil
	}

	// ICCP chunk contains raw ICC profile data
	// Use io.NewSectionReader to avoid loading entire chunk into memory
	section := io.NewSectionReader(r, chunk.DataOffset, int64(chunk.Size))
	dirs, _ := p.icc.Parse(section)
	return dirs
}

// parseImageChunk parses VP8/VP8L/VP8X chunks for technical metadata
func (p *Parser) parseImageChunk(r io.ReaderAt, chunk *Chunk) []parser.Tag {
	var tags []parser.Tag

	switch chunk.FourCC {
	case "VP8X":
		// Extended format - contains flags and dimensions
		if chunk.Size < 10 {
			return nil
		}

		data := make([]byte, 10)
		_, err := r.ReadAt(data, chunk.DataOffset)
		if err != nil {
			return nil
		}

		flags := data[0]
		// Width and height are stored as 24-bit values minus 1
		width := (uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16) + 1
		height := (uint32(data[7]) | uint32(data[8])<<8 | uint32(data[9])<<16) + 1

		// Parse flags
		var flagStrs []string
		if flags&0x02 != 0 {
			flagStrs = append(flagStrs, "EXIF")
		}
		if flags&0x04 != 0 {
			flagStrs = append(flagStrs, "XMP")
		}
		if flags&0x08 != 0 {
			flagStrs = append(flagStrs, "ICCP")
		}
		if flags&0x10 != 0 {
			flagStrs = append(flagStrs, "Alpha")
		}
		if flags&0x20 != 0 {
			flagStrs = append(flagStrs, "Animation")
		}

		flagStr := "None"
		if len(flagStrs) > 0 {
			flagStr = flagStrs[0]
			for i := 1; i < len(flagStrs); i++ {
				flagStr += ", " + flagStrs[i]
			}
		}

		tags = append(tags,
			parser.Tag{ID: "WebP:WebPFlags", Name: "WebPFlags", Value: flagStr, DataType: "string"},
			parser.Tag{ID: "WebP:ImageWidth", Name: "ImageWidth", Value: width, DataType: "uint32"},
			parser.Tag{ID: "WebP:ImageHeight", Name: "ImageHeight", Value: height, DataType: "uint32"},
		)

	case "VP8 ":
		// Lossy format
		if chunk.Size < 10 {
			return nil
		}

		data := make([]byte, 10)
		_, err := r.ReadAt(data, chunk.DataOffset)
		if err != nil {
			return nil
		}

		// VP8 Frame Tag (3 bytes):
		// Bit 0:     show_frame flag
		// Bits 1-3:  version number (0=bicubic, 1=simple, 2=complex/normal, 3=complex/simple)
		// Bits 4-23: first_part_size (not used here)
		frameTag := uint32(data[0]) | (uint32(data[1]) << 8) | (uint32(data[2]) << 16)
		version := (frameTag >> 1) & 0x07
		showFrame := frameTag & 0x01

		// VP8 start code check (fixed 3-byte sequence: 0x9D 0x01 0x2A)
		if data[3] != 0x9D || data[4] != 0x01 || data[5] != 0x2A {
			return nil // Invalid VP8 data
		}

		// VP8 Frame Header (bytes 6-9):
		// Bytes 6-7: Width (14 bits) + horizontal scale (2 bits)
		// Bytes 8-9: Height (14 bits) + vertical scale (2 bits)
		width := uint32(data[6]) | (uint32(data[7]&0x3F) << 8)
		horizontalScale := (data[7] >> 6) & 0x03
		height := uint32(data[8]) | (uint32(data[9]&0x3F) << 8)
		verticalScale := (data[9] >> 6) & 0x03

		// VP8 version description
		versionStr := fmt.Sprintf("%d (bicubic reconstruction, normal loop)", version)
		if version == 1 {
			versionStr = fmt.Sprintf("%d (simple/no loop filter)", version)
		} else if version == 2 {
			versionStr = fmt.Sprintf("%d (complex/normal loop filter)", version)
		} else if version == 3 {
			versionStr = fmt.Sprintf("%d (complex/simple loop filter)", version)
		}

		tags = append(tags,
			parser.Tag{ID: "WebP:VP8Version", Name: "VP8Version", Value: versionStr, DataType: "string"},
			parser.Tag{ID: "WebP:ImageWidth", Name: "ImageWidth", Value: width, DataType: "uint32"},
			parser.Tag{ID: "WebP:ImageHeight", Name: "ImageHeight", Value: height, DataType: "uint32"},
			parser.Tag{ID: "WebP:HorizontalScale", Name: "HorizontalScale", Value: uint32(horizontalScale), DataType: "uint32"},
			parser.Tag{ID: "WebP:VerticalScale", Name: "VerticalScale", Value: uint32(verticalScale), DataType: "uint32"},
		)

		if showFrame == 1 {
			tags = append(tags, parser.Tag{ID: "WebP:ShowFrame", Name: "ShowFrame", Value: "Yes", DataType: "string"})
		}

	case "VP8L":
		// Lossless format
		if chunk.Size < 5 {
			return nil
		}

		data := make([]byte, 5)
		_, err := r.ReadAt(data, chunk.DataOffset)
		if err != nil {
			return nil
		}

		// VP8L signature check (0x2F = '/' character)
		if data[0] != 0x2F {
			return nil
		}

		// VP8L bitstream format (after signature byte):
		// Bits 0-13:  Image width - 1  (14 bits)
		// Bits 14-27: Image height - 1 (14 bits)
		// Bits 28-31: Alpha hint + version (not parsed here)
		//
		// Bit layout across bytes 1-4:
		// Byte 1: [width:8]
		// Byte 2: [width:6][height:2]
		// Byte 3: [height:8]
		// Byte 4: [height:4][flags:4]

		// Extract width: bits 0-13 (14 bits) + 1
		width := ((uint32(data[1]) | (uint32(data[2]) << 8) | (uint32(data[3]) << 16) | (uint32(data[4]) << 24)) & 0x3FFF) + 1

		// Extract height: bits 14-27 (14 bits) + 1
		height := ((uint32(data[2])>>6 | (uint32(data[3]) << 2) | (uint32(data[4]) << 10)) & 0x3FFF) + 1

		tags = append(tags,
			parser.Tag{ID: "WebP:ImageWidth", Name: "ImageWidth", Value: width, DataType: "uint32"},
			parser.Tag{ID: "WebP:ImageHeight", Name: "ImageHeight", Value: height, DataType: "uint32"},
			parser.Tag{ID: "WebP:Format", Name: "Format", Value: "Lossless", DataType: "string"},
		)
	}

	return tags
}
