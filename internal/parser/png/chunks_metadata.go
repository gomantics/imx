package png

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/limits"
)

// parseExifChunk parses an eXIf chunk containing EXIF data
func (p *Parser) parseExifChunk(r io.ReaderAt, chunk *Chunk) []parser.Directory {
	if chunk.Length == 0 {
		return nil
	}

	// eXIf chunk contains standard EXIF data (TIFF format)
	// starting with byte order marker
	section := io.NewSectionReader(r, chunk.DataOffset, int64(chunk.Length))
	dirs, _ := p.tiff.Parse(section)
	return dirs
}

// parseiTXtChunk parses an iTXt chunk (may contain XMP)
func (p *Parser) parseiTXtChunk(r io.ReaderAt, chunk *Chunk) ([]parser.Directory, []parser.Tag) {
	if chunk.Length < 5 {
		return nil, nil
	}

	data := make([]byte, chunk.Length)
	_, err := r.ReadAt(data, chunk.DataOffset)
	if err != nil {
		return nil, nil
	}

	// iTXt format:
	// - Null-terminated keyword
	// - Compression flag (1 byte)
	// - Compression method (1 byte)
	// - Null-terminated language tag
	// - Null-terminated translated keyword
	// - Text data

	// Find keyword
	keywordEnd := bytes.IndexByte(data, itxtKeywordEnd)
	if keywordEnd < 0 {
		return nil, nil
	}

	keyword := string(data[:keywordEnd])

	// Check for XMP
	if keyword == itxtXMPKeyword {
		// Skip to text data (after language and translated keyword)
		offset := keywordEnd + itxtCompressionFlagOffset + itxtCompressionMethodOffset + 1 // +1 for null

		// Skip language tag
		langEnd := bytes.IndexByte(data[offset:], itxtKeywordEnd)
		if langEnd < 0 {
			return nil, nil
		}
		offset += langEnd + 1

		// Skip translated keyword
		transEnd := bytes.IndexByte(data[offset:], itxtKeywordEnd)
		if transEnd < 0 {
			return nil, nil
		}
		offset += transEnd + 1

		if offset >= len(data) {
			return nil, nil
		}

		// Parse XMP
		xmpData := data[offset:]
		reader := bytes.NewReader(xmpData)
		dirs, _ := p.xmp.Parse(reader)
		return dirs, nil
	}

	// Regular text metadata
	// Extract text value (simplified - assumes no compression)
	offset := keywordEnd + 1 + itxtCompressionFlagOffset + itxtCompressionMethodOffset // Skip null, compression flag, compression method
	langEnd := bytes.IndexByte(data[offset:], 0)
	if langEnd < 0 {
		return nil, nil
	}
	offset += langEnd + 1

	transEnd := bytes.IndexByte(data[offset:], 0)
	if transEnd < 0 {
		return nil, nil
	}
	offset += transEnd + 1

	if offset < len(data) {
		value := string(data[offset:])
		tag := parser.Tag{
			ID:       parser.TagID(fmt.Sprintf("PNG:iTXt:%s", keyword)),
			Name:     keyword,
			Value:    value,
			DataType: "string",
		}
		return nil, []parser.Tag{tag}
	}

	return nil, nil
}

// parseICCPChunk parses an iCCP chunk containing ICC profile
func (p *Parser) parseICCPChunk(r io.ReaderAt, chunk *Chunk) []parser.Directory {
	if chunk.Length < 10 {
		return nil
	}

	data := make([]byte, chunk.Length)
	_, err := r.ReadAt(data, chunk.DataOffset)
	if err != nil {
		return nil
	}

	// iCCP format:
	// - Null-terminated profile name
	// - Compression method (1 byte, must be 0 for deflate)
	// - Compressed profile data

	// Find profile name
	nameEnd := bytes.IndexByte(data, 0)
	if nameEnd < 0 || nameEnd+2 >= len(data) {
		return nil
	}

	compressionMethod := data[nameEnd+1]
	if compressionMethod != iccpCompressionDeflate {
		return nil // Only deflate compression is supported
	}

	// Decompress ICC profile
	compressedData := data[nameEnd+2:]
	decompressor, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil
	}
	defer decompressor.Close()

	var decompressed bytes.Buffer
	n, err := io.Copy(&decompressed, io.LimitReader(decompressor, limits.MaxPNGICCProfileLen+1))
	if err != nil {
		return nil
	}
	if n > limits.MaxPNGICCProfileLen {
		return nil
	}

	// Parse ICC profile
	reader := bytes.NewReader(decompressed.Bytes())
	dirs, _ := p.icc.Parse(reader)
	return dirs
}

// parsetEXtChunk parses a tEXt chunk (uncompressed text)
func (p *Parser) parsetEXtChunk(r io.ReaderAt, chunk *Chunk) *parser.Tag {
	if chunk.Length == 0 {
		return nil
	}

	data := make([]byte, chunk.Length)
	_, err := r.ReadAt(data, chunk.DataOffset)
	if err != nil {
		return nil
	}

	// tEXt format:
	// - Null-terminated keyword
	// - Text string (not null-terminated)

	keywordEnd := bytes.IndexByte(data, 0)
	if keywordEnd < 0 {
		return nil
	}

	keyword := string(data[:keywordEnd])
	value := ""
	if keywordEnd+1 < len(data) {
		value = string(data[keywordEnd+1:])
	}

	return &parser.Tag{
		ID:       parser.TagID(fmt.Sprintf("PNG:tEXt:%s", keyword)),
		Name:     keyword,
		Value:    value,
		DataType: "string",
	}
}

// parsezTXtChunk parses a zTXt chunk (compressed text)
func (p *Parser) parsezTXtChunk(r io.ReaderAt, chunk *Chunk) *parser.Tag {
	if chunk.Length < 3 {
		return nil
	}

	data := make([]byte, chunk.Length)
	_, err := r.ReadAt(data, chunk.DataOffset)
	if err != nil {
		return nil
	}

	// zTXt format:
	// - Null-terminated keyword
	// - Compression method (1 byte, must be 0 for deflate)
	// - Compressed text data

	keywordEnd := bytes.IndexByte(data, 0)
	if keywordEnd < 0 || keywordEnd+2 >= len(data) {
		return nil
	}

	keyword := string(data[:keywordEnd])
	compressionMethod := data[keywordEnd+ztxtCompressionMethodOffset]

	if compressionMethod != ztxtCompressionDeflate {
		return nil // Only deflate is supported
	}

	// Decompress text
	compressedData := data[keywordEnd+2:]
	decompressor, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil
	}
	defer decompressor.Close()

	var decompressed bytes.Buffer
	n, err := io.Copy(&decompressed, io.LimitReader(decompressor, limits.MaxPNGDecompressedTextLen+1))
	if err != nil {
		return nil
	}
	if n > limits.MaxPNGDecompressedTextLen {
		return nil
	}

	return &parser.Tag{
		ID:       parser.TagID(fmt.Sprintf("PNG:zTXt:%s", keyword)),
		Name:     keyword,
		Value:    decompressed.String(),
		DataType: "string",
	}
}
