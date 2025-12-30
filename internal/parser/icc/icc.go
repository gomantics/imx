package icc

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/gomantics/imx/internal/parser"
)

// Parser parses ICC color profiles.
//
// The parser is stateless and safe for concurrent use.
//
// Supported Tag Types:
//   - text: ASCII text
//   - desc: Text description (legacy)
//   - mluc: Multi-localized Unicode text
//   - XYZ : CIE XYZ color values
//   - sf32: S15Fixed16 number array
//   - uf32: U16Fixed16 number array
//   - sig : Technology signature
//   - curv: Tone reproduction curve
//   - para: Parametric curve
//   - dtim: Date and time
//   - meas: Measurement conditions
//   - view: Viewing conditions
//   - chrm: Chromaticity
//
// Unknown tag types return raw bytes. Unknown tag signatures return the
// raw 4-character signature code.
type Parser struct{}

// New creates a new ICC parser.
func New() *Parser {
	return &Parser{}
}

// Name returns the parser name.
func (p *Parser) Name() string {
	return "ICC"
}

// Detect checks if the data is an ICC profile by looking for the 'acsp' signature.
func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, signatureSize)
	_, err := r.ReadAt(buf, offsetSignature)
	return err == nil && buf[0] == iccSignature[0] && buf[1] == iccSignature[1] &&
		buf[2] == iccSignature[2] && buf[3] == iccSignature[3]
}

// Parse extracts metadata from an ICC profile.
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()
	var dirs []parser.Directory

	// Parse header
	headerDir, err := p.parseHeader(r)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to parse header: %w", err))
		return nil, parseErr
	}

	// Add header directory
	dirs = append(dirs, *headerDir)

	// Parse tag table
	tags, err := p.parseTagTable(r)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to parse tag table: %w", err))
		return dirs, parseErr
	}

	// Build ICC Profile directory with tag data
	profileDir := parser.Directory{
		Name: "ICC-Profile",
		Tags: make([]parser.Tag, 0),
	}

	// Parse each tag
	for _, tagRecord := range tags {
		tagData, err := p.parseTagData(r, tagRecord)
		if err != nil {
			// Skip tags that fail to parse
			continue
		}

		profileDir.Tags = append(profileDir.Tags, parser.Tag{
			ID:       parser.TagID("ICC:" + tagData.Signature),
			Name:     tagData.Signature,
			Value:    tagData.Value,
			DataType: tagData.Type,
		})
	}

	dirs = append(dirs, profileDir)

	return dirs, parseErr.OrNil()
}

// parseTagTable reads the tag table from the ICC profile.
func (p *Parser) parseTagTable(r io.ReaderAt) ([]TagRecord, error) {
	// Read tag count at offset 128
	buf := make([]byte, 4)
	_, err := r.ReadAt(buf, offsetTagTableCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read tag count: %w", err)
	}

	tagCount := binary.BigEndian.Uint32(buf)
	if tagCount == 0 {
		return nil, nil
	}

	// Sanity check
	if tagCount > maxTagCount {
		return nil, fmt.Errorf("unreasonable tag count: %d", tagCount)
	}

	// Read tag records (each 12 bytes)
	recordSize := tagCount * tagRecordSize
	records := make([]byte, recordSize)
	_, err = r.ReadAt(records, offsetTagTableEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to read tag records: %w", err)
	}

	tags := make([]TagRecord, tagCount)
	for i := uint32(0); i < tagCount; i++ {
		offset := i * tagRecordSize
		copy(tags[i].Signature[:], records[offset:offset+signatureSize])
		tags[i].Offset = binary.BigEndian.Uint32(records[offset+4 : offset+8])
		tags[i].Size = binary.BigEndian.Uint32(records[offset+8 : offset+12])
	}

	return tags, nil
}

// parseTagData reads and parses tag data based on its type.
func (p *Parser) parseTagData(r io.ReaderAt, tag TagRecord) (*TagData, error) {
	if tag.Size < minTagDataSize {
		return nil, fmt.Errorf("tag data too small: %d bytes", tag.Size)
	}

	// Read type signature (first 4 bytes at tag offset)
	typeBuf := make([]byte, minTagDataSize)
	_, err := r.ReadAt(typeBuf, int64(tag.Offset))
	if err != nil {
		return nil, fmt.Errorf("failed to read tag type: %w", err)
	}

	typeSignature := string(typeBuf[0:signatureSize])
	tagSig := string(tag.Signature[:])

	data := &TagData{
		Signature: getTagName(tagSig),
		Type:      typeSignature,
	}

	// Get converter from lookup table (returns default raw bytes converter for unknown types)
	converter := getTypeConverter(typeSignature)
	data.Value, err = converter(r, tag)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// parseHeader reads and parses the ICC profile header, returning it as a Directory.
func (p *Parser) parseHeader(r io.ReaderAt) (*parser.Directory, error) {
	buf := make([]byte, headerSize)
	_, err := r.ReadAt(buf, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Verify signature at offset 36
	if buf[offsetSignature] != iccSignature[0] || buf[offsetSignature+1] != iccSignature[1] ||
		buf[offsetSignature+2] != iccSignature[2] || buf[offsetSignature+3] != iccSignature[3] {
		return nil, fmt.Errorf("invalid ICC signature")
	}

	// Build directory
	dir := &parser.Directory{
		Name: "ICC-Header",
		Tags: make([]parser.Tag, 0, 17),
	}

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:ProfileSize",
		Name:     "ProfileSize",
		Value:    binary.BigEndian.Uint32(buf[0:4]),
		DataType: "uint32",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:CMMType",
		Name:     "CMMType",
		Value:    string(buf[4:8]),
		DataType: "string",
	})

	// Format version as major.minor.bugfix
	profileVersion := binary.BigEndian.Uint32(buf[8:12])
	major := (profileVersion >> 24) & 0xFF
	minor := (profileVersion >> 20) & 0x0F
	bugfix := (profileVersion >> 16) & 0x0F
	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:ProfileVersion",
		Name:     "ProfileVersion",
		Value:    fmt.Sprintf("%d.%d.%d", major, minor, bugfix),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:ProfileClass",
		Name:     "ProfileClass",
		Value:    getProfileClassName(string(buf[12:16])),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:ColorSpace",
		Name:     "ColorSpace",
		Value:    getColorSpaceName(string(buf[16:20])),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:ProfileConnectionSpace",
		Name:     "ProfileConnectionSpace",
		Value:    getColorSpaceName(string(buf[20:24])),
		DataType: "string",
	})

	// Parse datetime (12 bytes: year, month, day, hour, minute, second)
	dateTimeCreated := time.Date(
		int(binary.BigEndian.Uint16(buf[24:26])),        // year
		time.Month(binary.BigEndian.Uint16(buf[26:28])), // month
		int(binary.BigEndian.Uint16(buf[28:30])),        // day
		int(binary.BigEndian.Uint16(buf[30:32])),        // hour
		int(binary.BigEndian.Uint16(buf[32:34])),        // minute
		int(binary.BigEndian.Uint16(buf[34:36])),        // second
		0, time.UTC)
	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:DateTimeCreated",
		Name:     "DateTimeCreated",
		Value:    dateTimeCreated,
		DataType: "time.Time",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:ProfileSignature",
		Name:     "ProfileSignature",
		Value:    string(buf[36:40]),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:PrimaryPlatform",
		Name:     "PrimaryPlatform",
		Value:    getPlatformName(string(buf[40:44])),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:ProfileFlags",
		Name:     "ProfileFlags",
		Value:    getProfileFlagsName(binary.BigEndian.Uint32(buf[44:48])),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:DeviceManufacturer",
		Name:     "DeviceManufacturer",
		Value:    string(buf[48:52]),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:DeviceModel",
		Name:     "DeviceModel",
		Value:    string(buf[52:56]),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:DeviceAttributes",
		Name:     "DeviceAttributes",
		Value:    getDeviceAttributesName(binary.BigEndian.Uint64(buf[56:64])),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:RenderingIntent",
		Name:     "RenderingIntent",
		Value:    getRenderingIntentName(binary.BigEndian.Uint32(buf[64:68])),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:IlluminantX",
		Name:     "IlluminantX",
		Value:    float64(int32(binary.BigEndian.Uint32(buf[68:72]))) / 65536.0,
		DataType: "float64",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:IlluminantY",
		Name:     "IlluminantY",
		Value:    float64(int32(binary.BigEndian.Uint32(buf[72:76]))) / 65536.0,
		DataType: "float64",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:IlluminantZ",
		Name:     "IlluminantZ",
		Value:    float64(int32(binary.BigEndian.Uint32(buf[76:80]))) / 65536.0,
		DataType: "float64",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:ProfileCreator",
		Name:     "ProfileCreator",
		Value:    string(buf[80:84]),
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       "ICC:ProfileID",
		Name:     "ProfileID",
		Value:    fmt.Sprintf("%X", buf[84:100]),
		DataType: "string",
	})

	return dir, nil
}
