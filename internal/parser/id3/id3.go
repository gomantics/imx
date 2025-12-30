package id3

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"

	"github.com/gomantics/imx/internal/parser"
)

// Parser parses ID3v2 and ID3v1 metadata from MP3 files.
//
// Supported formats:
//   - ID3v2.4 (preferred, released 2000)
//   - ID3v2.3 (common, released 1999)
//   - ID3v2.2 (legacy, released 1998)
//   - ID3v1 (fallback, at end of file, released 1996)
//
// The parser uses io.ReaderAt for efficient random access without
// loading the entire file into memory.
type Parser struct{}

// New creates a new ID3 parser
func New() *Parser {
	return &Parser{}
}

// Name returns the parser name
func (p *Parser) Name() string {
	return "ID3"
}

// Detect checks if the data contains ID3v2 tags at the beginning
func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, 3)
	_, err := r.ReadAt(buf, 0)
	return err == nil && buf[0] == id3v2Signature[0] &&
		buf[1] == id3v2Signature[1] && buf[2] == id3v2Signature[2]
}

// Parse extracts ID3 metadata from MP3 file
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()
	var dirs []parser.Directory

	// Try to parse ID3v2 (at beginning of file)
	if v2Dir := p.parseID3v2(r, parseErr); v2Dir != nil {
		dirs = append(dirs, *v2Dir)
	}

	return dirs, parseErr.OrNil()
}

// parseID3v2 parses ID3v2 tags at the beginning of the file
func (p *Parser) parseID3v2(r io.ReaderAt, parseErr *parser.ParseError) *parser.Directory {
	// Read 10-byte header
	header := make([]byte, id3v2HeaderSize)
	_, err := r.ReadAt(header, 0)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read ID3v2 header: %w", err))
		return nil
	}

	// Verify "ID3" identifier
	if header[0] != id3v2Signature[0] || header[1] != id3v2Signature[1] || header[2] != id3v2Signature[2] {
		return nil // Not an ID3v2 tag
	}

	version := header[3]
	revision := header[4]
	flags := header[5]
	tagSize := decodeSynchsafeInt(header[6:10])

	dir := &parser.Directory{
		Name: fmt.Sprintf("ID3v2_%d", version),
		Tags: []parser.Tag{},
	}

	// Add header information
	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("ID3:Version"),
		Name:     "Version",
		Value:    fmt.Sprintf("2.%d.%d", version, revision),
		DataType: "string",
	})

	// Parse header flags
	hasExtHeader := (flags & flagExtendedHeader) != 0
	isExperimental := (flags & flagExperimental) != 0
	hasFooter := (flags & flagFooter) != 0

	if hasExtHeader || isExperimental || hasFooter {
		flagStr := ""
		if hasExtHeader {
			flagStr += "ExtHeader "
		}
		if isExperimental {
			flagStr += "Experimental "
		}
		if hasFooter {
			flagStr += "Footer "
		}
		dir.Tags = append(dir.Tags, parser.Tag{
			ID:       parser.TagID("ID3:Flags"),
			Name:     "Flags",
			Value:    flagStr,
			DataType: "string",
		})
	}

	// Start after 10-byte header
	pos := int64(id3v2HeaderSize)

	// Skip extended header if present
	if hasExtHeader {
		extSize, err := readSynchsafeInt(r, pos, 4)
		if err != nil {
			parseErr.Add(fmt.Errorf("failed to read extended header size: %w", err))
			return dir
		}
		pos += 4 + int64(extSize)
	}

	// Parse frames until end of tag
	tagEnd := int64(id3v2HeaderSize) + int64(tagSize)
	frameCount := 0

	for pos < tagEnd-int64(id3v2HeaderSize) { // Need at least 10 bytes for frame header
		frame, newPos, err := parseFrame(r, pos, version)
		if err != nil {
			if err == io.EOF {
				break // Hit padding
			}
			parseErr.Add(err)
			break
		}

		pos = newPos
		dir.Tags = append(dir.Tags, *frame)
		frameCount++

		// Safety check to prevent infinite loop
		if frameCount > maxFrameCount {
			parseErr.Add(fmt.Errorf("too many frames (>%d), stopping parse", maxFrameCount))
			break
		}
	}

	return dir
}

// parseFrame parses a single ID3v2 frame and returns the tag and new position
func parseFrame(r io.ReaderAt, pos int64, version byte) (*parser.Tag, int64, error) {
	// Frame header size depends on version
	var frameIDSize int
	if version == versionID3v22 {
		frameIDSize = frameIDSizeV2
	} else {
		frameIDSize = frameIDSizeV3
	}

	// Read frame ID
	frameIDBytes := make([]byte, frameIDSize)
	_, err := r.ReadAt(frameIDBytes, pos)
	if err != nil {
		return nil, pos, err
	}
	pos += int64(frameIDSize)

	// Check for padding (all zeros means end of frames)
	if frameIDBytes[0] == 0 {
		return nil, pos, io.EOF
	}

	frameID := string(frameIDBytes)

	// Read frame size
	var frameSize uint32
	if version == versionID3v22 {
		// ID3v2.2 uses 24-bit big-endian size
		sizeBuf := make([]byte, frameSizeV2)
		_, err := r.ReadAt(sizeBuf, pos)
		if err != nil {
			return nil, pos, err
		}
		pos += frameSizeV2
		frameSize = uint32(sizeBuf[0])<<16 | uint32(sizeBuf[1])<<8 | uint32(sizeBuf[2])
	} else if version == versionID3v24 {
		// ID3v2.4 uses synchsafe integer
		frameSize, err = readSynchsafeInt(r, pos, frameSizeV3)
		if err != nil {
			return nil, pos, err
		}
		pos += frameSizeV3
	} else {
		// ID3v2.3 uses regular 32-bit big-endian
		sizeBuf := make([]byte, frameSizeV3)
		_, err := r.ReadAt(sizeBuf, pos)
		if err != nil {
			return nil, pos, err
		}
		pos += frameSizeV3
		frameSize = binary.BigEndian.Uint32(sizeBuf)
	}

	// Validate frame size
	if frameSize == 0 || frameSize > maxFrameSize {
		return nil, pos, fmt.Errorf("invalid frame size: %d", frameSize)
	}

	// Read frame flags (v2.3 and v2.4 only)
	if version >= versionID3v23 {
		flagsBuf := make([]byte, frameFlagsSize)
		_, err := r.ReadAt(flagsBuf, pos)
		if err != nil {
			return nil, pos, err
		}
		pos += frameFlagsSize
		// TODO: Parse flags if needed (compression, encryption, etc.)
	}

	// Read frame data
	frameData := make([]byte, frameSize)
	_, err = r.ReadAt(frameData, pos)
	if err != nil {
		return nil, pos, err
	}
	pos += int64(frameSize)

	// Create tag from frame
	tag := &parser.Tag{
		ID:       parser.TagID(fmt.Sprintf("ID3:%s", frameID)),
		Name:     getFrameDescription(frameID),
		DataType: "string",
	}

	// Parse frame content based on type
	if isTextFrame(frameID) {
		tag.Value = decodeTextFrame(frameData)
	} else if frameID == framePicture || frameID == frameV2Picture {
		// Attached picture - store metadata about it
		tag.Value = fmt.Sprintf("Picture (%d bytes)", frameSize)
		tag.DataType = "binary"
	} else if frameID == frameComment {
		// Comment frame
		tag.Value = decodeCommentFrame(frameData)
	} else {
		// Generic binary frame
		tag.Value = fmt.Sprintf("Binary data (%d bytes)", frameSize)
		tag.DataType = "binary"
	}

	return tag, pos, nil
}

// readSynchsafeInt reads a synchsafe integer at the given position
func readSynchsafeInt(r io.ReaderAt, pos int64, numBytes int) (uint32, error) {
	buf := make([]byte, numBytes)
	_, err := r.ReadAt(buf, pos)
	if err != nil {
		return 0, err
	}
	return decodeSynchsafeInt(buf), nil
}

// decodeSynchsafeInt decodes a synchsafe integer
// Each byte uses only the lower 7 bits, MSB is always 0
func decodeSynchsafeInt(data []byte) uint32 {
	var result uint32
	for _, b := range data {
		result = (result << synchsafeBits) | uint32(b&synchsafeMask)
	}
	return result
}

// decodeTextFrame decodes text frame content
func decodeTextFrame(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// First byte is text encoding
	encoding := data[0]
	text := data[1:]

	switch encoding {
	case encodingISO88591:
		return string(trimNull(text))
	case encodingUTF16BOM:
		return decodeUTF16WithBOM(text)
	case encodingUTF16BE:
		return decodeUTF16BE(text)
	case encodingUTF8:
		return string(trimNull(text))
	default:
		return string(trimNull(text))
	}
}

// decodeCommentFrame decodes comment frame (COMM)
func decodeCommentFrame(data []byte) string {
	if len(data) < 4 {
		return ""
	}

	encoding := data[0]
	// Skip language (3 bytes) and short description
	// For simplicity, just decode the entire content
	text := data[4:]

	switch encoding {
	case encodingISO88591:
		return string(trimNull(text))
	case encodingUTF16BOM:
		return decodeUTF16WithBOM(text)
	case encodingUTF16BE:
		return decodeUTF16BE(text)
	case encodingUTF8:
		return string(trimNull(text))
	default:
		return string(trimNull(text))
	}
}

// decodeUTF16WithBOM decodes UTF-16 with byte order mark
func decodeUTF16WithBOM(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	// Check BOM
	if data[0] == bomUTF16LE[0] && data[1] == bomUTF16LE[1] {
		return decodeUTF16LE(data[2:])
	} else if data[0] == bomUTF16BE[0] && data[1] == bomUTF16BE[1] {
		return decodeUTF16BE(data[2:])
	}

	// No BOM, assume little-endian (most common)
	return decodeUTF16LE(data)
}

// decodeUTF16LE decodes UTF-16 little-endian
func decodeUTF16LE(data []byte) string {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}

	u16s := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		u16 := binary.LittleEndian.Uint16(data[i : i+2])
		if u16 == 0 {
			break
		}
		u16s = append(u16s, u16)
	}

	return string(utf16.Decode(u16s))
}

// decodeUTF16BE decodes UTF-16 big-endian
func decodeUTF16BE(data []byte) string {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}

	u16s := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		u16 := binary.BigEndian.Uint16(data[i : i+2])
		if u16 == 0 {
			break
		}
		u16s = append(u16s, u16)
	}

	return string(utf16.Decode(u16s))
}

// trimNull removes trailing null bytes
func trimNull(data []byte) []byte {
	for len(data) > 0 && data[len(data)-1] == 0 {
		data = data[:len(data)-1]
	}
	return data
}

// isTextFrame returns true if the frame ID represents a text frame
func isTextFrame(frameID string) bool {
	if len(frameID) == 0 {
		return false
	}
	// Text frames start with 'T' (ID3v2.3/2.4) or are text-like
	return frameID[0] == 'T' && frameID != frameUserText
}

// getFrameDescription returns a human-readable description for a frame ID
func getFrameDescription(frameID string) string {
	// ID3v2.3/2.4 frame IDs (4 characters)
	descriptions := map[string]string{
		"TIT2": "Title",
		"TPE1": "Artist",
		"TALB": "Album",
		"TDRC": "Recording Time",
		"TYER": "Year",
		"TDAT": "Date",
		"TIME": "Time",
		"TRCK": "Track Number",
		"TPOS": "Disc Number",
		"TCON": "Genre",
		"TPE2": "Album Artist",
		"TPE3": "Conductor",
		"TPE4": "Remixer",
		"TCOM": "Composer",
		"TEXT": "Lyricist",
		"TPUB": "Publisher",
		"TCOP": "Copyright",
		"TENC": "Encoded By",
		"TBPM": "BPM",
		"TKEY": "Initial Key",
		"TLAN": "Language",
		"TLEN": "Length",
		"TSRC": "ISRC",
		"TXXX": "User Defined Text",
		"COMM": "Comment",
		"USLT": "Unsynchronized Lyrics",
		"APIC": "Attached Picture",
		"GEOB": "General Encapsulated Object",
		"PCNT": "Play Counter",
		"POPM": "Popularimeter",
		"PRIV": "Private",
		"UFID": "Unique File Identifier",
		"USER": "Terms of Use",
		"WCOM": "Commercial Information URL",
		"WCOP": "Copyright URL",
		"WOAF": "Official Audio File URL",
		"WOAR": "Official Artist URL",
		"WOAS": "Official Source URL",
		"WORS": "Official Radio Station URL",
		"WPAY": "Payment URL",
		"WPUB": "Publisher URL",

		// ID3v2.2 frame IDs (3 characters)
		"TT2": "Title",
		"TP1": "Artist",
		"TAL": "Album",
		"TYE": "Year",
		"TRK": "Track Number",
		"TPA": "Disc Number",
		"TCO": "Genre",
		"TP2": "Album Artist",
		"TCM": "Composer",
		"TXT": "Lyricist",
		"TEN": "Encoded By",
		"TBP": "BPM",
		"TCR": "Copyright",
		"COM": "Comment",
		"PIC": "Attached Picture",
		"CNT": "Play Counter",
		"POP": "Popularimeter",
	}

	if desc, ok := descriptions[frameID]; ok {
		return desc
	}
	return frameID // Return frame ID as fallback
}
