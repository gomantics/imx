package gif

import (
	"bytes"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/xmp"
)

// parseExtension parses a GIF extension block at the given position
// Returns directories (for XMP), comment tags, new position, and whether parsing should continue
// DEPRECATED: Use parseExtensionWithLoopCount instead
func parseExtension(r io.ReaderAt, pos int64, buf *[11]byte, xmpParser *xmp.Parser) ([]parser.Directory, []parser.Tag, int64) {
	dirs, tags, _, newPos := parseExtensionWithLoopCount(r, pos, buf, xmpParser)
	return dirs, tags, newPos
}

// parseExtensionWithLoopCount parses a GIF extension block and extracts loop count if present
// Returns directories (for XMP), comment tags, loop count (-1 if not found), and new position
func parseExtensionWithLoopCount(r io.ReaderAt, pos int64, buf *[11]byte, xmpParser *xmp.Parser) ([]parser.Directory, []parser.Tag, int, int64) {
	var dirs []parser.Directory
	var tags []parser.Tag
	loopCount := -1 // -1 means not found

	// Read extension label
	_, err := r.ReadAt(buf[:1], pos)
	if err != nil {
		return nil, nil, -1, pos
	}

	label := buf[0]
	pos++

	switch label {
	case labelApplicationExt:
		xmpDirs, newLoopCount, newPos := parseApplicationExtensionWithLoopCount(r, pos, buf, xmpParser)
		dirs = append(dirs, xmpDirs...)
		if newLoopCount >= 0 {
			loopCount = newLoopCount
		}
		pos = newPos

	case labelComment:
		commentTag, newPos := parseCommentExtension(r, pos, buf)
		if commentTag != nil {
			tags = append(tags, *commentTag)
		}
		pos = newPos

	case labelGraphicControl:
		pos = skipDataSubBlocks(r, pos, buf)

	case labelPlainText:
		pos = skipDataSubBlocks(r, pos, buf)

	default:
		// Unknown extension, skip it
		pos = skipDataSubBlocks(r, pos, buf)
	}

	return dirs, tags, loopCount, pos
}

// parseApplicationExtension parses an Application Extension (may contain XMP)
// DEPRECATED: Use parseApplicationExtensionWithLoopCount instead
func parseApplicationExtension(r io.ReaderAt, pos int64, buf *[11]byte, xmpParser *xmp.Parser) ([]parser.Directory, int64) {
	dirs, _, newPos := parseApplicationExtensionWithLoopCount(r, pos, buf, xmpParser)
	return dirs, newPos
}

// parseApplicationExtensionWithLoopCount parses an Application Extension (may contain XMP or NETSCAPE loop count)
func parseApplicationExtensionWithLoopCount(r io.ReaderAt, pos int64, buf *[11]byte, xmpParser *xmp.Parser) ([]parser.Directory, int, int64) {
	// Read block size (should be 11 for Application Extension)
	_, err := r.ReadAt(buf[:1], pos)
	if err != nil {
		return nil, -1, pos
	}
	blockSize := buf[0]
	pos++

	if blockSize != applicationExtBlockSize {
		// Invalid application extension, skip it
		return nil, -1, skipDataSubBlocks(r, pos, buf)
	}

	// Read application identifier (8 bytes) + authentication code (3 bytes)
	_, err = r.ReadAt(buf[:applicationExtBlockSize], pos)
	if err != nil {
		return nil, -1, skipDataSubBlocks(r, pos, buf)
	}

	pos += applicationExtBlockSize

	appID := string(buf[0:applicationIDLength])
	authCode := string(buf[applicationIDLength:applicationExtBlockSize])

	// Check for XMP
	if appID == xmpApplicationID {
		// Check if this uses the old format (XMP directly) or standard format (sub-blocks)
		// Peek at the next byte - if it's '<' (0x3C), it's the old format
		_, err = r.ReadAt(buf[:1], pos)
		if err != nil {
			return nil, -1, pos
		}

		var xmpData []byte
		if buf[0] == xmpPacketStartChar {
			// Find XMP packet end by scanning in chunks
			xmpData, pos = readOldFormatXMP(r, pos, buf)
		} else {
			// Standard format with sub-blocks
			xmpData, pos = readDataSubBlocks(r, pos, buf)
		}

		if len(xmpData) > 0 {
			// Remove magic trailer if present (ends with 0x01 followed by 256 bytes of 0x00)
			xmpData = removeMagicTrailer(xmpData)

			// Parse XMP
			reader := bytes.NewReader(xmpData)
			dirs, _ := xmpParser.Parse(reader)
			return dirs, -1, pos
		}
	} else if appID == netscapeApplicationID && authCode == netscapeAuthCode {
		// NETSCAPE2.0 extension (animation loop count)
		// Read sub-block
		_, err := r.ReadAt(buf[:netscapeSubBlockSize], pos)
		if err == nil && buf[0] == netscapeSubBlockSize {
			// buf[1] should be 1 (sub-block ID)
			// buf[2] and next byte are loop count (little-endian uint16)
			var loopBuf [2]byte
			r.ReadAt(loopBuf[:], pos+netscapeLoopCountOffset)
			loopCount := int(loopBuf[0]) | (int(loopBuf[1]) << 8)
			pos = skipDataSubBlocks(r, pos, buf)
			return nil, loopCount, pos
		}
		pos = skipDataSubBlocks(r, pos, buf)
	} else {
		// Not XMP or NETSCAPE, skip remaining data
		pos = skipDataSubBlocks(r, pos, buf)
	}

	return nil, -1, pos
}

// parseCommentExtension parses a Comment Extension
func parseCommentExtension(r io.ReaderAt, pos int64, buf *[11]byte) (*parser.Tag, int64) {
	// Read comment data from sub-blocks
	commentData, newPos := readDataSubBlocks(r, pos, buf)
	if len(commentData) == 0 {
		return nil, newPos
	}

	tag := &parser.Tag{
		ID:       parser.TagID("GIF:Comment"),
		Name:     "Comment",
		Value:    string(commentData),
		DataType: "string",
	}

	return tag, newPos
}

// removeMagicTrailer removes the XMP magic trailer if present
func removeMagicTrailer(xmpData []byte) []byte {
	// XMP data format in GIF:
	// - Magic trailer of 257 bytes at the end (optional)
	// - Actual XMP data before the trailer

	if len(xmpData) > xmpMagicTrailerSize {
		// Check for magic trailer (0x01 followed by 256 bytes of 0x00)
		trailerStart := len(xmpData) - xmpMagicTrailerSize
		if xmpData[trailerStart] == xmpMagicTrailerMarker {
			allZeros := true
			for i := trailerStart + 1; i < len(xmpData); i++ {
				if xmpData[i] != xmpMagicTrailerFill {
					allZeros = false
					break
				}
			}
			if allZeros {
				return xmpData[:trailerStart]
			}
		}
	}

	return xmpData
}

// readOldFormatXMP reads XMP data stored directly (old format) by scanning in chunks
// It reads until it finds the block terminator (0x00), which comes after the XMP data
// and optional 257-byte magic trailer
func readOldFormatXMP(r io.ReaderAt, pos int64, buf *[11]byte) ([]byte, int64) {
	endMarker := []byte("<?xpacket end=")
	closingTag := []byte("?>")

	var accumulated []byte
	offset := pos

	for {
		// Read next chunk
		chunk := make([]byte, xmpReadChunkSize)
		n, err := r.ReadAt(chunk, offset)
		if n == 0 || (err != nil && err != io.EOF) {
			break
		}

		accumulated = append(accumulated, chunk[:n]...)
		offset += int64(n)

		// Search for end marker in accumulated data
		endIdx := bytes.Index(accumulated, endMarker)
		if endIdx != -1 {
			// Found end marker, now find closing ?>
			searchStart := endIdx + len(endMarker)
			remaining := accumulated[searchStart:]
			closeIdx := bytes.Index(remaining, closingTag)
			if closeIdx != -1 {
				// Found complete XMP packet end
				xmpEnd := searchStart + closeIdx + len(closingTag)

				// Now search for block terminator (0x00) after XMP data
				// There might be a 257-byte magic trailer between XMP and terminator
				terminatorSearch := accumulated[xmpEnd:]
				termIdx := bytes.IndexByte(terminatorSearch, separatorBlockTerminator)
				if termIdx != -1 {
					// Found block terminator
					xmpData := accumulated[:xmpEnd]
					pos += int64(xmpEnd + termIdx + 1) // +1 to skip the terminator
					return xmpData, pos
				}
				// If no terminator found yet, continue reading more chunks
			}
		}

		// If we read less than chunk size, we've hit EOF
		if n < xmpReadChunkSize {
			break
		}
	}

	return nil, pos
}

// readDataSubBlocks reads data from GIF sub-blocks
func readDataSubBlocks(r io.ReaderAt, pos int64, buf *[11]byte) ([]byte, int64) {
	var data []byte

	for {
		// Read block size
		_, err := r.ReadAt(buf[:1], pos)
		if err != nil {
			break
		}

		blockSize := buf[0]
		pos++

		// Block terminator
		if blockSize == separatorBlockTerminator {
			break
		}

		// Read block data
		blockData := make([]byte, blockSize)
		_, err = r.ReadAt(blockData, pos)
		if err != nil {
			break
		}

		data = append(data, blockData...)
		pos += int64(blockSize)
	}

	return data, pos
}

// skipDataSubBlocks skips over GIF sub-blocks
func skipDataSubBlocks(r io.ReaderAt, pos int64, buf *[11]byte) int64 {
	for {
		// Read block size
		_, err := r.ReadAt(buf[:1], pos)
		if err != nil {
			break
		}

		blockSize := buf[0]
		pos++

		// Block terminator
		if blockSize == separatorBlockTerminator {
			break
		}

		// Skip block data
		pos += int64(blockSize)
	}

	return pos
}
