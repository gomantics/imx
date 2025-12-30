package png

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/icc"
	"github.com/gomantics/imx/internal/parser/limits"
	"github.com/gomantics/imx/internal/parser/tiff"
	"github.com/gomantics/imx/internal/parser/xmp"
)

// Parser parses PNG image files.
//
// Supported metadata:
//   - EXIF (eXIf chunk)
//   - XMP (iTXt chunk with keyword "XML:com.adobe.xmp")
//   - ICC Profile (iCCP chunk)
//   - Text metadata (tEXt, zTXt, iTXt chunks)
//
// PNG uses a chunk-based format.
type Parser struct {
	tiff *tiff.Parser
	xmp  *xmp.Parser
	icc  *icc.Parser
}

// New creates a new PNG parser
func New() *Parser {
	return &Parser{
		tiff: tiff.New(),
		xmp:  xmp.New(),
		icc:  icc.New(),
	}
}

// Name returns the parser name
func (p *Parser) Name() string {
	return "PNG"
}

// PNG signature is defined in constants.go

// Detect checks if the data is a PNG file
func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, len(pngSignature))
	_, err := r.ReadAt(buf, 0)
	return err == nil && bytes.Equal(buf, pngSignature)
}

// Parse extracts metadata from a PNG file
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()
	var dirs []parser.Directory

	// Verify PNG signature
	buf := make([]byte, len(pngSignature))
	_, err := r.ReadAt(buf, 0)
	if err != nil || !bytes.Equal(buf, pngSignature) {
		parseErr.Add(fmt.Errorf("invalid PNG signature"))
		return nil, parseErr
	}

	pos := int64(len(pngSignature))

	// Create text metadata directory
	textDir := &parser.Directory{
		Name: "PNG-Text",
		Tags: []parser.Tag{},
	}

	// Create PNG technical directory
	pngDir := &parser.Directory{
		Name: "PNG",
		Tags: []parser.Tag{},
	}

	// Parse chunks
	for {
		var chunk *Chunk
		chunk, pos, err = p.readChunk(r, pos)
		if err != nil {
			if err == io.EOF {
				break
			}
			parseErr.Add(err)
			break
		}

		// Process metadata chunks
		switch chunk.Type {
		case chunkTypeIHDR:
			// Image header
			tags := p.parseIHDRChunk(r, chunk)
			pngDir.Tags = append(pngDir.Tags, tags...)

		case chunkTypecHRM:
			// Chromaticity
			tags := p.parsecHRMChunk(r, chunk)
			pngDir.Tags = append(pngDir.Tags, tags...)

		case chunkTypegAMA:
			// Gamma
			tag := p.parsegAMAChunk(r, chunk)
			if tag != nil {
				pngDir.Tags = append(pngDir.Tags, *tag)
			}

		case chunkTypepHYs:
			// Physical dimensions
			tags := p.parsepHYsChunk(r, chunk)
			pngDir.Tags = append(pngDir.Tags, tags...)

		case chunkTypetIME:
			// Modification time
			tag := p.parsetIMEChunk(r, chunk)
			if tag != nil {
				pngDir.Tags = append(pngDir.Tags, *tag)
			}

		case chunkTypebKGD:
			// Background color
			tag := p.parsebKGDChunk(r, chunk)
			if tag != nil {
				pngDir.Tags = append(pngDir.Tags, *tag)
			}

		case chunkTypeeXIf:
			// EXIF metadata
			exifDirs := p.parseExifChunk(r, chunk)
			dirs = append(dirs, exifDirs...)

		case chunkTypeiTXt:
			// International text (may contain XMP)
			xmpDirs, textTags := p.parseiTXtChunk(r, chunk)
			dirs = append(dirs, xmpDirs...)
			textDir.Tags = append(textDir.Tags, textTags...)

		case chunkTypeiCCP:
			// ICC color profile
			iccDirs := p.parseICCPChunk(r, chunk)
			dirs = append(dirs, iccDirs...)

		case chunkTypetEXt:
			// Uncompressed text
			textTag := p.parsetEXtChunk(r, chunk)
			if textTag != nil {
				textDir.Tags = append(textDir.Tags, *textTag)
			}

		case chunkTypezTXt:
			// Compressed text
			textTag := p.parsezTXtChunk(r, chunk)
			if textTag != nil {
				textDir.Tags = append(textDir.Tags, *textTag)
			}

		case chunkTypeIEND:
			// End of PNG
			goto done
		}
	}

done:
	// Add PNG directory if it has tags
	if len(pngDir.Tags) > 0 {
		dirs = append(dirs, *pngDir)
	}

	// Add text directory if it has tags
	if len(textDir.Tags) > 0 {
		dirs = append(dirs, *textDir)
	}

	return dirs, parseErr.OrNil()
}

// Chunk represents a PNG chunk
type Chunk struct {
	Length     uint32
	Type       string
	DataOffset int64
}

// readChunk reads a PNG chunk header
func (p *Parser) readChunk(r io.ReaderAt, pos int64) (*Chunk, int64, error) {
	buf := make([]byte, chunkHeaderSize)
	_, err := r.ReadAt(buf, pos)
	if err != nil {
		return nil, pos, err
	}

	length := binary.BigEndian.Uint32(buf[0:4])
	if length > limits.MaxPNGChunkSize {
		return nil, pos, fmt.Errorf("png: chunk length %d exceeds limit %d", length, limits.MaxPNGChunkSize)
	}
	chunkType := string(buf[4:8])

	chunk := &Chunk{
		Length:     length,
		Type:       chunkType,
		DataOffset: pos + chunkHeaderSize,
	}

	// Update position to after this chunk (data + CRC)
	newPos := pos + chunkHeaderSize + int64(length) + crcSize
	if newPos < pos {
		return nil, pos, fmt.Errorf("png: chunk position overflow for %s", chunkType)
	}
	return chunk, newPos, nil
}
