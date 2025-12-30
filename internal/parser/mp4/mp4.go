package mp4

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/limits"
)

// Parser parses MP4/M4A audio container files.
type Parser struct{}

// New creates a new MP4 parser.
func New() *Parser { return &Parser{} }

// Name returns the parser name.
func (p *Parser) Name() string { return "MP4" }

// Detect checks if the data is an MP4/M4A file.
func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, 12)
	if _, err := r.ReadAt(buf, 0); err != nil {
		return false
	}
	if string(buf[4:8]) != atomFTYP {
		return false
	}
	majorBrand := string(buf[8:12])
	validBrands := []string{
		"M4A ", "M4B ", "M4P ", "M4V ",
		"mp41", "mp42",
		"isom", "iso2",
		"dash",
		"avc1",
	}
	for _, b := range validBrands {
		if majorBrand == b {
			return true
		}
	}
	return false
}

// Parse extracts metadata from an MP4/M4A file.
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()
	var dirs []parser.Directory

	pos := int64(0)
	for {
		atom, err := readAtomAt(r, pos)
		if err != nil {
			if err == io.EOF {
				break
			}
			parseErr.Add(err)
			break
		}

		switch atom.Type {
		case atomFTYP:
			if dir := p.parseFtyp(r, atom); dir != nil && len(dir.Tags) > 0 {
				dirs = append(dirs, *dir)
			}
		case atomMOOV:
			metaDirs := p.parseMoov(r, atom, parseErr)
			dirs = append(dirs, metaDirs...)
		}

		pos = atom.Offset + int64(atom.Size)
		if pos > limits.MaxScanBytes {
			break
		}
	}

	return dirs, parseErr.OrNil()
}

// readAtomAt reads an atom header.
func readAtomAt(r io.ReaderAt, offset int64) (*Atom, error) {
	header := make([]byte, atomHeaderSize)
	if _, err := r.ReadAt(header, offset); err != nil {
		return nil, err
	}
	size := uint64(binary.BigEndian.Uint32(header[0:4]))
	atomType := string(header[4:8])

	if size == 1 {
		ext := make([]byte, 8)
		if _, err := r.ReadAt(ext, offset+atomHeaderSize); err != nil {
			return nil, err
		}
		size = binary.BigEndian.Uint64(ext)
	}
	if size == 0 {
		return nil, fmt.Errorf("mp4: atom %s has zero size", atomType)
	}

	if size < atomHeaderSize {
		return nil, fmt.Errorf("mp4: atom %s size %d too small", atomType, size)
	}
	if size > limits.MaxMP4AtomSize {
		return nil, fmt.Errorf("mp4: atom %s size %d exceeds limit %d", atomType, size, limits.MaxMP4AtomSize)
	}

	return &Atom{Type: atomType, Size: size, Offset: offset}, nil
}

// parseFtyp parses the file type atom.
func (p *Parser) parseFtyp(r io.ReaderAt, atom *Atom) *parser.Directory {
	if atom.Size < 16 {
		return nil
	}
	if atom.Size > limits.MaxMP4AtomSize {
		return nil
	}
	data := make([]byte, int(atom.Size))
	if _, err := r.ReadAt(data, atom.Offset); err != nil {
		return nil
	}

	dir := &parser.Directory{Name: "MP4-File-Type", Tags: []parser.Tag{}}

	majorBrand := string(data[8:12])
	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("MP4:ftyp:MajorBrand"),
		Name:     "MajorBrand",
		Value:    majorBrand,
		DataType: "string",
	})

	minorVersion := binary.BigEndian.Uint32(data[12:16])
	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("MP4:ftyp:MinorVersion"),
		Name:     "MinorVersion",
		Value:    minorVersion,
		DataType: "uint32",
	})

	var compat []string
	for i := 16; i+4 <= int(atom.Size); i += 4 {
		brand := string(data[i : i+4])
		if brand != "\x00\x00\x00\x00" {
			compat = append(compat, brand)
		}
	}
	if len(compat) > 0 {
		dir.Tags = append(dir.Tags, parser.Tag{
			ID:       parser.TagID("MP4:ftyp:CompatibleBrands"),
			Name:     "CompatibleBrands",
			Value:    compat,
			DataType: "string[]",
		})
	}
	return dir
}

// parseMoov parses the movie atom for metadata.
func (p *Parser) parseMoov(r io.ReaderAt, moovAtom *Atom, parseErr *parser.ParseError) []parser.Directory {
	var dirs []parser.Directory

	udtaAtom := findChildAtom(r, moovAtom, atomUDTA)
	if udtaAtom == nil {
		return dirs
	}
	metaAtom := findChildAtom(r, udtaAtom, atomMETA)
	if metaAtom == nil {
		return dirs
	}

	// meta is a full box with version (1 byte) + flags (3 bytes) after the atom header
	// We need to skip these 4 bytes to get to the child atoms
	fullBoxHeader := make([]byte, fullBoxHeaderSize)
	if _, err := r.ReadAt(fullBoxHeader, metaAtom.Offset+atomHeaderSize); err != nil {
		parseErr.Add(fmt.Errorf("failed to read meta box version/flags: %w", err))
		return dirs
	}
	// version := fullBoxHeader[0] // Currently unused
	// flags := binary.BigEndian.Uint32([]byte{0, fullBoxHeader[1], fullBoxHeader[2], fullBoxHeader[3]}) // Currently unused

	// Adjust offset to skip the full box header when searching for children
	metaWithOffset := &Atom{Type: metaAtom.Type, Size: metaAtom.Size, Offset: metaAtom.Offset + fullBoxHeaderSize}
	if ilstAtom := findChildAtom(r, metaWithOffset, atomILST); ilstAtom != nil {
		if dir := p.parseIlst(r, ilstAtom); dir != nil && len(dir.Tags) > 0 {
			dirs = append(dirs, *dir)
		}
	}
	return dirs
}

// findChildAtom finds a child atom within a parent atom.
func findChildAtom(r io.ReaderAt, parent *Atom, childType string) *Atom {
	offset := parent.Offset + atomHeaderSize
	end := parent.Offset + int64(parent.Size)
	for offset < end {
		header := make([]byte, atomHeaderSize)
		if _, err := r.ReadAt(header, offset); err != nil {
			return nil
		}
		size := uint64(binary.BigEndian.Uint32(header[0:4]))
		atomType := string(header[4:8])
		if size < atomHeaderSize || size > limits.MaxMP4AtomSize {
			break
		}
		if atomType == childType {
			return &Atom{Type: atomType, Size: size, Offset: offset}
		}
		next := offset + int64(size)
		if next <= offset {
			break
		}
		offset = next
	}
	return nil
}

// parseIlst parses the iTunes metadata item list.
func (p *Parser) parseIlst(r io.ReaderAt, ilstAtom *Atom) *parser.Directory {
	dir := &parser.Directory{Name: "MP4-Metadata", Tags: []parser.Tag{}}
	offset := ilstAtom.Offset + atomHeaderSize
	end := ilstAtom.Offset + int64(ilstAtom.Size)

	for offset < end {
		header := make([]byte, atomHeaderSize)
		if _, err := r.ReadAt(header, offset); err != nil {
			break
		}

		size := uint64(binary.BigEndian.Uint32(header[0:4]))
		atomType := string(header[4:8])

		if size < atomHeaderSize || size > limits.MaxMP4MetadataSize {
			break
		}

		if offset+int64(size) > end {
			break
		}

		if tag := p.parseMetadataAtom(r, offset, size, atomType); tag != nil {
			dir.Tags = append(dir.Tags, *tag)
		}
		offset += int64(size)
	}
	return dir
}

// parseMetadataAtom parses a single metadata atom.
func (p *Parser) parseMetadataAtom(r io.ReaderAt, offset int64, size uint64, atomType string) *parser.Tag {
	dataHeader := make([]byte, minMetadataAtom)
	if _, err := r.ReadAt(dataHeader, offset+atomHeaderSize); err != nil {
		return nil
	}
	dataSize := binary.BigEndian.Uint32(dataHeader[0:4])
	dataType := string(dataHeader[4:8])
	if dataType != atomDATA || dataSize < minMetadataAtom {
		return nil
	}
	if uint64(dataSize) > size {
		return nil
	}

	dataTypeIndicator := binary.BigEndian.Uint32(dataHeader[8:12])
	valueSize := dataSize - minMetadataAtom
	if valueSize == 0 || valueSize > limits.MaxMP4MetadataSize {
		return nil
	}
	if uint64(valueSize)+uint64(minMetadataAtom) > size {
		return nil
	}
	valueData := make([]byte, valueSize)
	if _, err := r.ReadAt(valueData, offset+atomHeaderSize+minMetadataAtom); err != nil {
		return nil
	}

	var value any
	dataTypeStr := "string"
	switch dataTypeIndicator {
	case dataTypeUTF8:
		value = string(bytes.TrimRight(valueData, "\x00"))
	case dataTypeSigned:
		switch len(valueData) {
		case 1:
			value = int8(valueData[0])
			dataTypeStr = "int8"
		case 2:
			value = int16(binary.BigEndian.Uint16(valueData))
			dataTypeStr = "int16"
		case 4:
			value = int32(binary.BigEndian.Uint32(valueData))
			dataTypeStr = "int32"
		default:
			value = fmt.Sprintf("Data (%d bytes, type %d)", len(valueData), dataTypeIndicator)
		}
	case dataTypeBinary:
		if atomType == "trkn" || atomType == "disk" {
			if len(valueData) >= 6 {
				current := binary.BigEndian.Uint16(valueData[2:4])
				total := binary.BigEndian.Uint16(valueData[4:6])
				if total > 0 {
					value = fmt.Sprintf("%d/%d", current, total)
				} else {
					value = fmt.Sprintf("%d", current)
				}
				dataTypeStr = "string"
			} else {
				value = fmt.Sprintf("Binary data (%d bytes)", len(valueData))
				dataTypeStr = "binary"
			}
		} else {
			value = fmt.Sprintf("Binary data (%d bytes)", len(valueData))
			dataTypeStr = "binary"
		}
	default:
		value = fmt.Sprintf("Data (%d bytes, type %d)", len(valueData), dataTypeIndicator)
		dataTypeStr = "unknown"
	}

	tagName := getMetadataTagName(atomType)
	displayName := tagName
	if displayName == atomType && len(atomType) == 4 {
		displayName = fmt.Sprintf("Tag_%02X%02X%02X%02X", atomType[0], atomType[1], atomType[2], atomType[3])
	}

	return &parser.Tag{
		ID:       parser.TagID(fmt.Sprintf("MP4:%s", atomType)),
		Name:     displayName,
		Value:    value,
		DataType: dataTypeStr,
	}
}

// getMetadataTagName returns a human-readable name for metadata atom types.
func getMetadataTagName(atomType string) string {
	names := map[string]string{
		"\xa9nam": "Title",
		"\xa9ART": "Artist",
		"\xa9alb": "Album",
		"\xa9day": "Year",
		"\xa9gen": "Genre",
		"\xa9cmt": "Comment",
		"\xa9too": "Encoder",
		"\xa9wrt": "Composer",
		"\xa9lyr": "Lyrics",
		"\xa9grp": "Grouping",
		"trkn":    "TrackNumber",
		"disk":    "DiscNumber",
		"gnre":    "GenreID",
		"cpil":    "Compilation",
		"tmpo":    "BPM",
		"covr":    "CoverArt",
		"aART":    "AlbumArtist",
		"pgap":    "GaplessPlayback",
		"rtng":    "Rating",
		"cprt":    "Copyright",
		"desc":    "Description",
		"ldes":    "LongDescription",
		"tvsh":    "TVShowName",
		"tven":    "TVEpisode",
		"tvsn":    "TVSeason",
		"tvnn":    "TVNetwork",
		"catg":    "Category",
		"keyw":    "Keywords",
		"purd":    "PurchaseDate",
		"purl":    "PodcastURL",
		"egid":    "EpisodeGlobalID",
		"cmID":    "ContentID",
		"sfID":    "StoreFrontID",
		"atID":    "AccountTypeID",
		"cnID":    "CatalogID",
		"plID":    "PlaylistID",
		"geID":    "GenreID",
		"soal":    "SortAlbum",
		"soaa":    "SortAlbumArtist",
		"soar":    "SortArtist",
		"sonm":    "SortName",
		"soco":    "SortComposer",
	}
	if name, ok := names[atomType]; ok {
		return name
	}
	return atomType
}

// buildAtom creates a test atom for testing purposes.
func buildAtom(atomType string, payload []byte) []byte {
	size := uint32(atomHeaderSize + len(payload))
	buf := make([]byte, size)
	binary.BigEndian.PutUint32(buf[0:4], size)
	copy(buf[4:8], atomType)
	copy(buf[8:], payload)
	return buf
}
