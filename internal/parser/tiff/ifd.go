package tiff

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	imxbin "github.com/gomantics/imx/internal/binary"
	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/limits"
)

// parseIFD parses an IFD at the given offset
func (p *Parser) parseIFD(r *imxbin.Reader, fileReader io.ReaderAt, offset int64, dirName string, iccDirs, iptcDirs, xmpDirs, makernoteDirs *[]parser.Directory, sharedParseErr *parser.ParseError) (*parser.Directory, *parser.ParseError, []SubIFD, uint16) {
	var parseErr *parser.ParseError
	if sharedParseErr != nil {
		// Use shared error accumulator for multi-IFD parsing
		parseErr = sharedParseErr
	} else {
		parseErr = parser.NewParseError()
	}
	var subIFDs []SubIFD

	// Read number of entries
	numEntries, err := r.ReadUint16(offset)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read IFD entry count at offset %d: %w", offset, err))
		return nil, parseErr, nil, 0
	}

	dir := &parser.Directory{
		Name: dirName,
		Tags: make([]parser.Tag, 0),
	}

	entryOffset := offset + ifdEntryCountSize

	// Read each IFD entry
	for i := uint16(0); i < numEntries; i++ {
		entry, err := p.readIFDEntry(r, entryOffset)
		if err != nil {
			parseErr.Add(fmt.Errorf("failed to read IFD entry %d at offset %d: %w", i, entryOffset, err))
			entryOffset += ifdEntrySize
			continue
		}

		// Check for special tags
		switch entry.Tag {
		case TagExifIFD:
			subIFDs = append(subIFDs, SubIFD{Offset: int64(entry.ValueOffset), Name: "ExifIFD"})
		case TagGPSIFD:
			subIFDs = append(subIFDs, SubIFD{Offset: int64(entry.ValueOffset), Name: "GPS"})
		case TagInteropIFD:
			subIFDs = append(subIFDs, SubIFD{Offset: int64(entry.ValueOffset), Name: "Interoperability"})
		case TagSubIFDs:
			p.handleSubIFDs(r, entry, &subIFDs, parseErr)
		case TagICCProfile:
			p.handleICCProfile(r, entry, &dir.Tags, parseErr, iccDirs)
		case TagIPTC:
			p.handleIPTC(r, entry, parseErr, iptcDirs)
		case TagXMP:
			p.handleXMP(r, entry, parseErr, xmpDirs)
		case TagMakerNote:
			p.handleMakerNote(r, fileReader, entry, &dir.Tags, parseErr, makernoteDirs)
		default:
			// Regular tag
			tag, err := p.parseTag(r, entry, dirName)
			if err != nil {
				parseErr.Add(fmt.Errorf("failed to parse tag 0x%04X at offset %d: %w", entry.Tag, entryOffset, err))
			} else if tag != nil {
				dir.Tags = append(dir.Tags, *tag)
			}
		}

		entryOffset += ifdEntrySize
	}

	return dir, parseErr, subIFDs, numEntries
}

// readIFDEntry reads a single IFD entry
func (p *Parser) readIFDEntry(r *imxbin.Reader, offset int64) (*IFDEntry, error) {
	tag, err := r.ReadUint16(offset + ifdEntryTagOffset)
	if err != nil {
		return nil, err
	}

	typeVal, err := r.ReadUint16(offset + ifdEntryTypeOffset)
	if err != nil {
		return nil, err
	}

	count, err := r.ReadUint32(offset + ifdEntryCountOffset)
	if err != nil {
		return nil, err
	}

	valueOffset, err := r.ReadUint32(offset + ifdEntryValueOffset)
	if err != nil {
		return nil, err
	}

	return &IFDEntry{
		Tag:         tag,
		Type:        TagType(typeVal),
		Count:       count,
		ValueOffset: valueOffset,
	}, nil
}

// parseTag parses a tag value
func (p *Parser) parseTag(r *imxbin.Reader, entry *IFDEntry, dirName string) (*parser.Tag, error) {
	tagName := getTagName(entry.Tag, dirName)

	value, err := p.readTagValue(r, entry)
	if err != nil {
		return nil, err
	}

	// Special formatting for version tags
	switch {
	case entry.Tag == tagGPSVersionID && strings.ToLower(dirName) == "gps":
		if bytes, ok := value.([]byte); ok && len(bytes) == 4 {
			value = fmt.Sprintf("%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3])
		}
	case entry.Tag == tagExifVersion || entry.Tag == tagFlashpixVersion:
		// ExifVersion/FlashpixVersion are stored as 4 ASCII bytes (e.g., "0230" for 2.30)
		if bytes, ok := value.([]byte); ok && len(bytes) == 4 {
			value = string(bytes)
		}
	case entry.Tag == tagInteropVersion && strings.ToLower(dirName) == "interoperability":
		// InteroperabilityVersion is also stored as 4 ASCII bytes
		if bytes, ok := value.([]byte); ok && len(bytes) == 4 {
			value = string(bytes)
		}
	}

	// Decode enum values to human-readable strings
	if decoded := decodeEnumValue(entry.Tag, dirName, value); decoded != "" {
		value = decoded
	}

	return &parser.Tag{
		ID:       parser.TagID(fmt.Sprintf("%s:0x%04X", dirName, entry.Tag)),
		Name:     tagName,
		Value:    value,
		DataType: entry.Type.String(),
	}, nil
}

// readTagValue reads the actual tag value based on type and count.
//
// Inline Data Optimization:
// The TIFF specification allows tag values ≤4 bytes to be stored directly
// in the ValueOffset field instead of using it as a pointer. This optimization
// avoids extra file reads for small values like uint16, uint32, and short strings.
//
// When totalSize ≤ 4 bytes:
//   - dataOffset is set to -1 (marker for inline data)
//   - Value is extracted directly from entry.ValueOffset with correct byte order
//
// When totalSize > 4 bytes:
//   - dataOffset = entry.ValueOffset (used as file offset)
//   - Value is read from the file at that offset
func (p *Parser) readTagValue(r *imxbin.Reader, entry *IFDEntry) (interface{}, error) {
	typeSize := entry.Type.TypeSize()
	if typeSize == 0 {
		return nil, fmt.Errorf("unknown type: %d", entry.Type)
	}

	// Prevent integer overflow in size calculation
	// Use int64 to safely calculate total size and validate against limit
	totalSize64 := int64(entry.Count) * int64(typeSize)
	if totalSize64 > limits.MaxTIFFTagDataSize {
		return nil, fmt.Errorf("tag data size %d exceeds limit of %d bytes", totalSize64, limits.MaxTIFFTagDataSize)
	}

	totalSize := int(totalSize64)

	// Determine if value is inline or offset
	var dataOffset int64
	if totalSize <= inlineDataThreshold {
		// Value is stored inline in the ValueOffset field
		dataOffset = -1 // Special marker for inline data
	} else {
		dataOffset = int64(entry.ValueOffset)
	}

	switch entry.Type {
	case TypeByte, TypeUndefined:
		return p.readBytes(r, entry, dataOffset)
	case TypeASCII:
		return p.readASCII(r, entry, dataOffset)
	case TypeShort:
		return p.readShorts(r, entry, dataOffset)
	case TypeLong:
		return p.readLongs(r, entry, dataOffset)
	case TypeRational:
		return p.readRationals(r, entry, dataOffset)
	case TypeSByte:
		return p.readSBytes(r, entry, dataOffset)
	case TypeSShort:
		return p.readSShorts(r, entry, dataOffset)
	case TypeSLong:
		return p.readSLongs(r, entry, dataOffset)
	case TypeSRational:
		return p.readSRationals(r, entry, dataOffset)
	default:
		return nil, fmt.Errorf("unsupported type: %s", entry.Type.String())
	}
}

// readBytes reads byte values
func (p *Parser) readBytes(r *imxbin.Reader, entry *IFDEntry, dataOffset int64) (interface{}, error) {
	count := int(entry.Count)

	var data []byte
	if dataOffset == -1 {
		// Inline data
		buf := make([]byte, bufferSizeUint32)
		r.PutUint32(buf, entry.ValueOffset)
		data = buf[:count]
	} else {
		var err error
		data, err = r.ReadBytes(dataOffset, count)
		if err != nil {
			return nil, err
		}
	}

	if count == 1 {
		return data[0], nil
	}
	return data, nil
}

// readASCII reads ASCII string
func (p *Parser) readASCII(r *imxbin.Reader, entry *IFDEntry, dataOffset int64) (interface{}, error) {
	count := int(entry.Count)

	var data []byte
	if dataOffset == -1 {
		// Inline data
		data = make([]byte, bufferSizeUint32)
		r.PutUint32(data, entry.ValueOffset)
		data = data[:count]
	} else {
		var err error
		data, err = r.ReadBytes(dataOffset, count)
		if err != nil {
			return nil, err
		}
	}

	// Remove null terminator
	data = bytes.TrimRight(data, "\x00")
	return string(data), nil
}

// readShorts reads uint16 values
func (p *Parser) readShorts(r *imxbin.Reader, entry *IFDEntry, dataOffset int64) (interface{}, error) {
	count := int(entry.Count)
	if count == 0 {
		return []uint16{}, nil
	}

	values := make([]uint16, count)

	if dataOffset == -1 {
		// Inline data (up to 2 shorts) - stored in ValueOffset with byte order
		buf := make([]byte, bufferSizeUint32)
		r.PutUint32(buf, entry.ValueOffset)
		values[0] = r.Uint16(buf[0:2])
		if count > 1 {
			values[1] = r.Uint16(buf[2:4])
		}
	} else {
		for i := 0; i < count; i++ {
			val, err := r.ReadUint16(dataOffset + int64(i*typeSizeShort))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

// readLongs reads uint32 values
func (p *Parser) readLongs(r *imxbin.Reader, entry *IFDEntry, dataOffset int64) (interface{}, error) {
	count := int(entry.Count)
	if count == 0 {
		return []uint32{}, nil
	}

	values := make([]uint32, count)

	if dataOffset == -1 {
		// Inline data (only 1 long fits)
		values[0] = entry.ValueOffset
	} else {
		for i := 0; i < count; i++ {
			val, err := r.ReadUint32(dataOffset + int64(i*typeSizeLong))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

// readRationals reads rational values (numerator/denominator pairs)
func (p *Parser) readRationals(r *imxbin.Reader, entry *IFDEntry, dataOffset int64) (interface{}, error) {
	count := int(entry.Count)
	if count == 0 {
		return []string{}, nil
	}

	values := make([]string, count)

	for i := 0; i < count; i++ {
		offset := dataOffset + int64(i*typeSizeRational)
		num, err := r.ReadUint32(offset)
		if err != nil {
			return nil, err
		}
		denom, err := r.ReadUint32(offset + typeSizeLong)
		if err != nil {
			return nil, err
		}
		values[i] = fmt.Sprintf("%d/%d", num, denom)
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

// readSBytes reads signed byte values
func (p *Parser) readSBytes(r *imxbin.Reader, entry *IFDEntry, dataOffset int64) (interface{}, error) {
	count := int(entry.Count)
	if count == 0 {
		return []int8{}, nil
	}

	var data []byte
	if dataOffset == -1 {
		buf := make([]byte, 4)
		r.PutUint32(buf, entry.ValueOffset)
		data = buf[:count]
	} else {
		var err error
		data, err = r.ReadBytes(dataOffset, count)
		if err != nil {
			return nil, err
		}
	}

	values := make([]int8, count)
	for i, b := range data {
		values[i] = int8(b)
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

// readSShorts reads int16 values
func (p *Parser) readSShorts(r *imxbin.Reader, entry *IFDEntry, dataOffset int64) (interface{}, error) {
	count := int(entry.Count)
	if count == 0 {
		return []int16{}, nil
	}

	values := make([]int16, count)

	if dataOffset == -1 {
		// Inline data (up to 2 shorts) - stored in ValueOffset with byte order
		buf := make([]byte, bufferSizeUint32)
		r.PutUint32(buf, entry.ValueOffset)
		values[0] = int16(r.Uint16(buf[0:2]))
		if count > 1 {
			values[1] = int16(r.Uint16(buf[2:4]))
		}
	} else {
		for i := 0; i < count; i++ {
			val, err := r.ReadInt16(dataOffset + int64(i*typeSizeSShort))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

// readSLongs reads int32 values
func (p *Parser) readSLongs(r *imxbin.Reader, entry *IFDEntry, dataOffset int64) (interface{}, error) {
	count := int(entry.Count)
	if count == 0 {
		return []int32{}, nil
	}

	values := make([]int32, count)

	if dataOffset == -1 {
		values[0] = int32(entry.ValueOffset)
	} else {
		for i := 0; i < count; i++ {
			val, err := r.ReadInt32(dataOffset + int64(i*typeSizeSLong))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

// readSRationals reads signed rational values
func (p *Parser) readSRationals(r *imxbin.Reader, entry *IFDEntry, dataOffset int64) (interface{}, error) {
	count := int(entry.Count)
	if count == 0 {
		return []string{}, nil
	}

	values := make([]string, count)

	for i := 0; i < count; i++ {
		offset := dataOffset + int64(i*typeSizeSRational)
		num, err := r.ReadInt32(offset)
		if err != nil {
			return nil, err
		}
		denom, err := r.ReadInt32(offset + typeSizeSLong)
		if err != nil {
			return nil, err
		}
		values[i] = fmt.Sprintf("%d/%d", num, denom)
	}

	if count == 1 {
		return values[0], nil
	}
	return values, nil
}

// SubIFD represents a sub-IFD to be parsed
type SubIFD struct {
	Offset int64
	Name   string
}

// handleICCProfile handles ICC profile tag
func (p *Parser) handleICCProfile(r *imxbin.Reader, entry *IFDEntry, tags *[]parser.Tag, parseErr *parser.ParseError, iccDirs *[]parser.Directory) {
	// Read ICC profile data
	data, err := r.ReadBytes(int64(entry.ValueOffset), int(entry.Count))
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read ICC profile data at offset %d: %w", entry.ValueOffset, err))
		return
	}

	// Parse ICC profile using ICC parser
	reader := bytes.NewReader(data)
	if p.icc != nil {
		dirs, iccErr := p.icc.Parse(reader)
		if iccErr != nil {
			parseErr.Merge(iccErr)
		}
		*iccDirs = append(*iccDirs, dirs...)
	}
}

// handleIPTC handles IPTC tag
func (p *Parser) handleIPTC(r *imxbin.Reader, entry *IFDEntry, parseErr *parser.ParseError, iptcDirs *[]parser.Directory) {
	// Read IPTC data
	data, err := r.ReadBytes(int64(entry.ValueOffset), int(entry.Count))
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read IPTC data at offset %d: %w", entry.ValueOffset, err))
		return
	}

	// Parse IPTC using IPTC parser
	reader := bytes.NewReader(data)
	if p.iptc != nil {
		dirs, iptcErr := p.iptc.Parse(reader)
		if iptcErr != nil {
			parseErr.Merge(iptcErr)
		}
		*iptcDirs = append(*iptcDirs, dirs...)
	}
}

// handleXMP handles XMP tag
func (p *Parser) handleXMP(r *imxbin.Reader, entry *IFDEntry, parseErr *parser.ParseError, xmpDirs *[]parser.Directory) {
	// Read XMP data
	data, err := r.ReadBytes(int64(entry.ValueOffset), int(entry.Count))
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read XMP data at offset %d: %w", entry.ValueOffset, err))
		return
	}

	// XMP data should be null-terminated, trim it
	data = bytes.TrimRight(data, "\x00")

	// Parse XMP using XMP parser
	reader := bytes.NewReader(data)
	if p.xmp != nil {
		dirs, xmpErr := p.xmp.Parse(reader)
		if xmpErr != nil {
			parseErr.Merge(xmpErr)
		}
		*xmpDirs = append(*xmpDirs, dirs...)
	}
}

// handleMakerNote handles MakerNote tag (tag 0x927C)
// MakerNote contains manufacturer-specific metadata in various formats.
// When a manufacturer handler is registered and parses successfully, the
// parsed tags are returned in a separate directory.
// When no handler matches, the raw MakerNote data is returned as a tag.
func (p *Parser) handleMakerNote(r *imxbin.Reader, fileReader io.ReaderAt, entry *IFDEntry, dirTags *[]parser.Tag, parseErr *parser.ParseError, makernoteDirs *[]parser.Directory) {
	// Read MakerNote data
	makerNoteOffset := int64(entry.ValueOffset)
	data, err := r.ReadBytes(makerNoteOffset, int(entry.Count))
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to read MakerNote data at offset %d: %w", makerNoteOffset, err))
		return
	}

	// If no registry or no handler matches, return raw MakerNote as a tag
	if p.makernote == nil {
		*dirTags = append(*dirTags, parser.Tag{
			ID:       parser.TagID("ExifIFD:0x927C"),
			Name:     "MakerNote",
			Value:    data,
			DataType: "UNDEFINED",
		})
		return
	}

	handler, cfg := p.makernote.Detect(data)
	if handler == nil {
		// Unknown manufacturer - return raw data as tag
		*dirTags = append(*dirTags, parser.Tag{
			ID:       parser.TagID("ExifIFD:0x927C"),
			Name:     "MakerNote",
			Value:    data,
			DataType: "UNDEFINED",
		})
		return
	}

	// Always add raw MakerNote tag (for backward compatibility)
	*dirTags = append(*dirTags, parser.Tag{
		ID:       parser.TagID("ExifIFD:0x927C"),
		Name:     "MakerNote",
		Value:    data,
		DataType: "UNDEFINED",
	})

	// Detect manufacturer and parse
	// exifBase is 0 for standard TIFF files (TIFF header at file start)
	// TODO: For JPEG files, this would need to be the APP1 EXIF header offset
	exifBase := int64(0)

	tags, mnErr := handler.Parse(fileReader, makerNoteOffset, exifBase, cfg)
	if mnErr != nil {
		parseErr.Merge(mnErr)
	}

	// Create directory for MakerNote tags if any were parsed
	if len(tags) > 0 {
		*makernoteDirs = append(*makernoteDirs, parser.Directory{
			Name: handler.Manufacturer(),
			Tags: tags,
		})
	}
}

// handleSubIFDs handles SubIFDs tag (tag 0x014A)
// SubIFDs contain an array of offsets to sub-IFDs for preview/RAW image data
func (p *Parser) handleSubIFDs(r *imxbin.Reader, entry *IFDEntry, subIFDs *[]SubIFD, parseErr *parser.ParseError) {
	// Read array of SubIFD offsets (type LONG)
	count := int(entry.Count)

	// Determine offset to read from
	dataOffset := int64(-1)
	if count*typeSizeLong > inlineDataThreshold {
		dataOffset = int64(entry.ValueOffset)
	}

	// Read the offsets
	for i := 0; i < count; i++ {
		var offset uint32
		if dataOffset == -1 {
			// Inline data - read from ValueOffset
			if i == 0 {
				offset = entry.ValueOffset
			}
			// Only 1 offset fits inline
		} else {
			// Read from file
			val, err := r.ReadUint32(dataOffset + int64(i*typeSizeLong))
			if err != nil {
				parseErr.Add(fmt.Errorf("failed to read SubIFD offset %d at offset %d: %w", i, dataOffset+int64(i*typeSizeLong), err))
				continue
			}
			offset = val
		}

		// Create SubIFD entry with appropriate name
		name := "SubIFD"
		if i > 0 {
			name = fmt.Sprintf("SubIFD%d", i)
		}
		*subIFDs = append(*subIFDs, SubIFD{Offset: int64(offset), Name: name})
	}
}

// getTagName returns a human-readable name for a tag
func getTagName(tag uint16, dirName string) string {
	// Try directory-specific lookup first
	if name := getTagNameForDir(tag, dirName); name != "" {
		return name
	}

	// Fall back to hex representation
	return fmt.Sprintf("0x%04X", tag)
}

// getTagNameForDir returns tag name for specific directory
func getTagNameForDir(tag uint16, dirName string) string {
	dirName = strings.ToLower(dirName)

	// Check if it's a SubIFD directory (SubIFD, SubIFD1, SubIFD2, etc.)
	if strings.HasPrefix(dirName, "subifd") {
		return getTIFFTagName(tag)
	}

	switch dirName {
	case "ifd0", "ifd1", "tiff":
		return getTIFFTagName(tag)
	case "exififd":
		return getEXIFTagName(tag)
	case "gps":
		return getGPSTagName(tag)
	case "interoperability":
		return getInteropTagName(tag)
	default:
		return getTIFFTagName(tag) // Default to TIFF tags
	}
}
