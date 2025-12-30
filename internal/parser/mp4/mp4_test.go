package mp4

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/limits"
)

func TestParser_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "MP4" {
		t.Errorf("Name() = %v, want %v", got, "MP4")
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid M4A ftyp",
			data: createFtypBox("M4A ", 0, []string{}),
			want: true,
		},
		{
			name: "valid mp41",
			data: createFtypBox("mp41", 0, []string{}),
			want: true,
		},
		{
			name: "valid isom",
			data: createFtypBox("isom", 0, []string{}),
			want: true,
		},
		{
			name: "invalid marker",
			data: []byte("NOT_MP4_DATA"),
			want: false,
		},
		{
			name: "too short",
			data: []byte("ftyp"),
			want: false,
		},
		{
			name: "empty",
			data: []byte{},
			want: false,
		},
		{
			name: "unknown brand",
			data: createFtypBox("unkn", 0, []string{}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			r := bytes.NewReader(tt.data)
			if got := p.Detect(r); got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Parse_MinimalFile(t *testing.T) {
	// Create minimal MP4 with just ftyp
	var buf bytes.Buffer

	ftypData := createFtypBox("M4A ", 0, []string{"mp42", "isom"})
	buf.Write(ftypData)
	buf.Write(buildMoovWithMeta())

	p := New()
	r := bytes.NewReader(buf.Bytes())

	dirs, err := p.Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) < 2 {
		t.Fatalf("Parse() got %d directories, want at least 2 (ftyp + metadata)", len(dirs))
	}
}

func TestParser_Parse_InvalidData(t *testing.T) {
	data := []byte("INVALID_DATA_NOT_MP4")
	p := New()
	r := bytes.NewReader(data)

	dirs, _ := p.Parse(r)
	if len(dirs) != 0 {
		t.Errorf("Parse() with invalid data returned %d directories, want 0", len(dirs))
	}
}

func TestParseMetadataAtom_TrackDisc(t *testing.T) {
	// Build metadata atom for trkn with data atom
	value := []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x02} // padding(2), current=1, total=2
	tag := buildDataAtom(value, dataTypeBinary)
	atom := buildAtom("trkn", tag)

	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "trkn")
	if res == nil || res.Value != "1/2" {
		t.Fatalf("parseMetadataAtom trkn got %+v", res)
	}
}

func TestParseMetadataAtom_String(t *testing.T) {
	value := []byte("Hello\x00")
	tag := buildDataAtom(value, dataTypeUTF8)
	atom := buildAtom("\xa9nam", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "\xa9nam")
	if res == nil || res.Value != "Hello" {
		t.Fatalf("parseMetadataAtom string got %+v", res)
	}
}

func TestParseMetadataAtom_Int(t *testing.T) {
	value := []byte{0x00, 0x02}
	tag := buildDataAtom(value, dataTypeSigned)
	atom := buildAtom("tmpo", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "tmpo")
	if res == nil || res.Value != int16(2) {
		t.Fatalf("parseMetadataAtom int got %+v", res)
	}
}

func TestParseMetadataAtom_Int8(t *testing.T) {
	value := []byte{0x7F}
	tag := buildDataAtom(value, dataTypeSigned)
	atom := buildAtom("tmpo", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "tmpo")
	if res == nil || res.Value != int8(0x7F) {
		t.Fatalf("parseMetadataAtom int8 got %+v", res)
	}
}

func TestParseMetadataAtom_Int32(t *testing.T) {
	value := make([]byte, 4)
	binary.BigEndian.PutUint32(value, 0x01020304)
	tag := buildDataAtom(value, dataTypeSigned)
	atom := buildAtom("tmpo", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "tmpo")
	if res == nil || res.Value != int32(0x01020304) {
		t.Fatalf("parseMetadataAtom int32 got %+v", res)
	}
}

func TestParseMetadataAtom_UnknownType(t *testing.T) {
	value := []byte{0x01}
	tag := buildDataAtom(value, 999)
	atom := buildAtom("xxxx", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "xxxx")
	if res == nil || res.DataType != "unknown" {
		t.Fatalf("expected unknown type, got %+v", res)
	}
}

func TestParseMetadataAtom_Binary(t *testing.T) {
	value := []byte{0xAA, 0xBB}
	tag := buildDataAtom(value, dataTypeBinary)
	atom := buildAtom("covr", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "covr")
	if res == nil || res.DataType != "binary" {
		t.Fatalf("expected binary type, got %+v", res)
	}
}

func TestReadAtomAt_EOF(t *testing.T) {
	_, err := readAtomAt(bytes.NewReader([]byte{0x00}), 0)
	if err == nil {
		t.Fatal("expected EOF error")
	}
}

func TestReadAtomAt_ReadError(t *testing.T) {
	r := errorReaderAt{err: io.ErrUnexpectedEOF}
	_, err := readAtomAt(r, 0)
	if err == nil {
		t.Fatal("expected read error")
	}
}

func TestReadAtomAt_SizeZero(t *testing.T) {
	buf := make([]byte, atomHeaderSize)
	// size=0 => to EOF sentinel
	binary.BigEndian.PutUint32(buf[0:4], 0)
	copy(buf[4:8], atomFTYP)
	r := bytes.NewReader(buf)
	if _, err := readAtomAt(r, 0); err == nil {
		t.Fatalf("expected error for zero atom size")
	}
}

func TestReadAtomAt_Offset(t *testing.T) {
	first := buildAtom("skip", []byte("1234"))
	second := buildAtom(atomFTYP, []byte("data"))
	buf := append(first, second...)
	a, err := readAtomAt(bytes.NewReader(buf), int64(len(first)))
	if err != nil {
		t.Fatalf("readAtomAt offset err: %v", err)
	}
	if a.Type != atomFTYP || a.Offset != int64(len(first)) {
		t.Fatalf("unexpected atom %+v", a)
	}
}

func TestReadAtomAt_Regular(t *testing.T) {
	atom := buildAtom("abcd", []byte("payload"))
	a, err := readAtomAt(bytes.NewReader(atom), 0)
	if err != nil {
		t.Fatalf("readAtomAt err: %v", err)
	}
	if a.Type != "abcd" || a.Size != uint64(len(atom)) {
		t.Fatalf("unexpected atom %+v", a)
	}
}

func TestReadAtomAt_ExtendedSize(t *testing.T) {
	// size=1 indicates extended 64-bit size follows
	payload := []byte("abcd")
	totalSize := uint64(atomHeaderSize + 8 + len(payload))
	buf := make([]byte, atomHeaderSize+8+len(payload))
	binary.BigEndian.PutUint32(buf[0:4], 1)
	copy(buf[4:8], atomFTYP)
	binary.BigEndian.PutUint64(buf[8:16], totalSize)
	copy(buf[16:], payload)

	r := bytes.NewReader(buf)
	a, err := readAtomAt(r, 0)
	if err != nil {
		t.Fatalf("readAtomAt extended error: %v", err)
	}
	if a.Size != totalSize {
		t.Fatalf("extended size = %d, want %d", a.Size, totalSize)
	}
}

func TestReadAtomAt_ExtendedReadError(t *testing.T) {
	// size=1 but extended read fails
	buf := make([]byte, atomHeaderSize)
	binary.BigEndian.PutUint32(buf[0:4], 1)
	copy(buf[4:8], atomFTYP)
	r := &partialErrReader{data: buf, failAfter: 1, err: io.ErrUnexpectedEOF}
	if _, err := readAtomAt(r, 0); err == nil {
		t.Fatalf("expected error on extended read")
	}
}

type partialErrReader struct {
	data      []byte
	failAfter int
	calls     int
	err       error
}

func (p *partialErrReader) ReadAt(b []byte, off int64) (int, error) {
	p.calls++
	if p.calls > p.failAfter {
		return 0, p.err
	}
	if off >= int64(len(p.data)) {
		return 0, io.EOF
	}
	n := copy(b, p.data[off:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

func TestParser_Parse_ReadAtomError(t *testing.T) {
	p := New()
	r := errorReaderAt{err: io.ErrUnexpectedEOF}
	dirs, parseErr := p.Parse(r)
	if parseErr == nil {
		t.Fatal("expected parse error")
	}
	if len(dirs) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(dirs))
	}
}

func TestParser_Parse_MaxScanLimit(t *testing.T) {
	// Atom with size=0 (to EOF) should advance pos beyond maxScanBytes and stop
	buf := make([]byte, atomHeaderSize)
	binary.BigEndian.PutUint32(buf[0:4], 0) // size=0
	copy(buf[4:8], "skip")
	p := New()
	dirs, _ := p.Parse(bytes.NewReader(buf))
	if len(dirs) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(dirs))
	}
}

type errorReaderAt struct{ err error }

func (e errorReaderAt) ReadAt(p []byte, off int64) (int, error) {
	return 0, e.err
}

func TestParseMoov_NoUdta(t *testing.T) {
	moov := buildAtom(atomMOOV, []byte("no udta"))
	p := New()
	res := p.parseMoov(bytes.NewReader(moov), &Atom{Type: atomMOOV, Size: uint64(len(moov)), Offset: 0}, parser.NewParseError())
	if len(res) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(res))
	}
}

func TestParseMoov_MetaNoIlst(t *testing.T) {
	metaPayload := make([]byte, fullBoxHeaderSize) // meta with no ilst
	meta := buildAtom(atomMETA, metaPayload)
	udta := buildAtom(atomUDTA, meta)
	moov := buildAtom(atomMOOV, udta)
	p := New()
	res := p.parseMoov(bytes.NewReader(moov), &Atom{Type: atomMOOV, Size: uint64(len(moov)), Offset: 0}, parser.NewParseError())
	if len(res) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(res))
	}
}

func TestParseMoov_UdtaNoMeta(t *testing.T) {
	udta := buildAtom(atomUDTA, buildAtom("xxxx", []byte("payload")))
	moov := buildAtom(atomMOOV, udta)
	p := New()
	res := p.parseMoov(bytes.NewReader(moov), &Atom{Type: atomMOOV, Size: uint64(len(moov)), Offset: 0}, parser.NewParseError())
	if len(res) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(res))
	}
}
func TestParseMoov_ReadError(t *testing.T) {
	moov := buildAtom(atomMOOV, []byte{0x00}) // payload too small for child header
	p := New()
	res := p.parseMoov(bytes.NewReader(moov), &Atom{Type: atomMOOV, Size: uint64(len(moov)), Offset: 0}, parser.NewParseError())
	if len(res) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(res))
	}
}

func TestParseMoov_ReadErrorFindChild(t *testing.T) {
	// reader returns error to findChildAtom
	p := New()
	moov := buildAtom(atomMOOV, []byte("payload"))
	r := errorReaderAt{err: io.ErrUnexpectedEOF}
	res := p.parseMoov(r, &Atom{Type: atomMOOV, Size: uint64(len(moov)), Offset: 0}, parser.NewParseError())
	if len(res) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(res))
	}
}

func TestParseMoov_MetaVersionReadError(t *testing.T) {
	// Build udta with meta, but meta is too short to read version/flags
	meta := buildAtom(atomMETA, []byte{}) // meta with no payload - can't read version/flags
	udta := buildAtom(atomUDTA, meta)
	moov := buildAtom(atomMOOV, udta)

	p := New()
	parseErr := parser.NewParseError()
	res := p.parseMoov(bytes.NewReader(moov), &Atom{Type: atomMOOV, Size: uint64(len(moov)), Offset: 0}, parseErr)

	if len(res) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(res))
	}
	if parseErr.OrNil() == nil {
		t.Fatal("expected parse error for failed meta version/flags read")
	}
}

func TestFindChildAtom_InvalidSize(t *testing.T) {
	// Child atom with size < atomHeaderSize should break the loop
	parentData := make([]byte, atomHeaderSize+atomHeaderSize)
	binary.BigEndian.PutUint32(parentData[0:4], uint32(len(parentData))) // parent size
	copy(parentData[4:8], "test")
	// child with invalid size (3 < 8)
	binary.BigEndian.PutUint32(parentData[8:12], 3)
	copy(parentData[12:16], "bad!")

	parent := &Atom{Type: "test", Size: uint64(len(parentData)), Offset: 0}
	r := bytes.NewReader(parentData)
	result := findChildAtom(r, parent, "bad!")
	if result != nil {
		t.Fatalf("expected nil for invalid child size, got %+v", result)
	}
}

func TestParseIlst_StopOnBadSize(t *testing.T) {
	// ilst with invalid child size
	var buf bytes.Buffer
	// child with size 4 (too small)
	binary.Write(&buf, binary.BigEndian, uint32(4))
	buf.WriteString("bad!")
	ilst := buildAtom(atomILST, buf.Bytes())
	p := New()
	r := bytes.NewReader(ilst)
	dir := p.parseIlst(r, &Atom{Type: atomILST, Size: uint64(len(ilst)), Offset: 0})
	if len(dir.Tags) != 0 {
		t.Fatalf("expected no tags, got %d", len(dir.Tags))
	}
}

func TestParseIlst_ReadError(t *testing.T) {
	// ilst but reader fails
	ilst := buildAtom(atomILST, []byte("payload"))
	p := New()
	r := errorReaderAt{err: io.ErrUnexpectedEOF}
	dir := p.parseIlst(r, &Atom{Type: atomILST, Size: uint64(len(ilst)), Offset: 0})
	if len(dir.Tags) != 0 {
		t.Fatalf("expected no tags, got %d", len(dir.Tags))
	}
}

func TestParseIlst_SizeTooLarge(t *testing.T) {
	// child size > maxMetadataSize
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(limits.MaxMP4MetadataSize+atomHeaderSize+1))
	buf.WriteString("abcd")
	ilst := buildAtom(atomILST, buf.Bytes())
	p := New()
	r := bytes.NewReader(ilst)
	dir := p.parseIlst(r, &Atom{Type: atomILST, Size: uint64(len(ilst)), Offset: 0})
	if len(dir.Tags) != 0 {
		t.Fatalf("expected no tags, got %d", len(dir.Tags))
	}
}

func TestParseMetadataAtom_InvalidType(t *testing.T) {
	// dataType != "data"
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(minMetadataAtom))
	buf.WriteString("xxxx") // not data
	binary.Write(&buf, binary.BigEndian, uint32(dataTypeUTF8))
	binary.Write(&buf, binary.BigEndian, uint32(0))
	buf.Write([]byte("abc"))
	atom := buildAtom("xxxx", buf.Bytes())

	p := New()
	r := bytes.NewReader(atom)
	if tag := p.parseMetadataAtom(r, 0, uint64(len(atom)), "xxxx"); tag != nil {
		t.Fatalf("expected nil tag for invalid data type")
	}
}

func TestParseMetadataAtom_ValueTooBig(t *testing.T) {
	// valueSize too large
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(limits.MaxMP4MetadataSize+minMetadataAtom+1))
	buf.WriteString(atomDATA)
	binary.Write(&buf, binary.BigEndian, uint32(dataTypeUTF8))
	binary.Write(&buf, binary.BigEndian, uint32(0))
	atom := buildAtom("xxxx", buf.Bytes())
	p := New()
	r := bytes.NewReader(atom)
	if tag := p.parseMetadataAtom(r, 0, uint64(len(atom)), "xxxx"); tag != nil {
		t.Fatalf("expected nil tag for oversized value")
	}
}

func TestParseMetadataAtom_ReadError(t *testing.T) {
	// Truncated data atom
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(minMetadataAtom))
	buf.WriteString(atomDATA)
	binary.Write(&buf, binary.BigEndian, uint32(dataTypeUTF8))
	binary.Write(&buf, binary.BigEndian, uint32(0))
	// no value bytes
	atom := buildAtom("xxxx", buf.Bytes())
	p := New()
	r := bytes.NewReader(atom[:len(atom)-2]) // truncate
	if tag := p.parseMetadataAtom(r, 0, uint64(len(atom)-2), "xxxx"); tag != nil {
		t.Fatalf("expected nil tag for read error")
	}
}

func TestParseMetadataAtom_TrackNoTotal(t *testing.T) {
	value := []byte{0x00, 0x00, 0x00, 0x02, 0x00, 0x00} // current=2, total=0
	tag := buildDataAtom(value, dataTypeBinary)
	atom := buildAtom("trkn", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "trkn")
	if res == nil || res.Value != "2" {
		t.Fatalf("expected track value '2', got %+v", res)
	}
}

func TestParseMetadataAtom_ValueZero(t *testing.T) {
	// dataSize equals header; valueSize =0 -> nil
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(minMetadataAtom))
	buf.WriteString(atomDATA)
	binary.Write(&buf, binary.BigEndian, uint32(dataTypeUTF8))
	binary.Write(&buf, binary.BigEndian, uint32(0))
	atom := buildAtom("xxxx", buf.Bytes())
	p := New()
	r := bytes.NewReader(atom)
	if tag := p.parseMetadataAtom(r, 0, uint64(len(atom)), "xxxx"); tag != nil {
		t.Fatalf("expected nil tag for zero value size")
	}
}

func TestParseMetadataAtom_DataSizeTooSmall(t *testing.T) {
	// dataSize < minMetadataAtom
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(minMetadataAtom))
	buf.WriteString(atomDATA)
	binary.Write(&buf, binary.BigEndian, uint32(dataTypeUTF8))
	binary.Write(&buf, binary.BigEndian, uint32(0))
	atom := buildAtom("xxxx", buf.Bytes())
	// force size to 8 (too small) when calling
	p := New()
	r := bytes.NewReader(atom)
	if tag := p.parseMetadataAtom(r, 0, uint64(atomHeaderSize), "xxxx"); tag != nil {
		t.Fatalf("expected nil for small data size")
	}
}

func TestParseMetadataAtom_SignedLengthMismatch(t *testing.T) {
	value := []byte{0x01, 0x02, 0x03} // len 3 not handled -> default unknown
	tag := buildDataAtom(value, dataTypeSigned)
	atom := buildAtom("tmpo", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "tmpo")
	if res == nil || res.DataType != "string" {
		t.Fatalf("expected string fallback for mismatch len, got %+v", res)
	}
}

func TestParseMetadataAtom_BinaryNonTrack(t *testing.T) {
	value := []byte{0x01, 0x02}
	tag := buildDataAtom(value, dataTypeBinary)
	atom := buildAtom("abcd", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "abcd")
	if res == nil || res.DataType != "binary" {
		t.Fatalf("expected binary fallback, got %+v", res)
	}
}

func TestParseMetadataAtom_TrackTooShort(t *testing.T) {
	value := []byte{0x00, 0x01} // too short for track format
	tag := buildDataAtom(value, dataTypeBinary)
	atom := buildAtom("trkn", tag)
	p := New()
	r := bytes.NewReader(atom)
	res := p.parseMetadataAtom(r, 0, uint64(len(atom)), "trkn")
	if res == nil || res.DataType != "binary" {
		t.Fatalf("expected binary fallback for short track, got %+v", res)
	}
}

func TestParseMetadataAtom_ValueReadError(t *testing.T) {
	// First ReadAt succeeds, second (value read) fails
	value := []byte("abc")
	tag := buildDataAtom(value, dataTypeUTF8)
	atom := buildAtom("xxxx", tag)
	r := &partialErrReader{data: atom, failAfter: 1, err: io.ErrUnexpectedEOF}
	p := New()
	if tagRes := p.parseMetadataAtom(r, 0, uint64(len(atom)), "xxxx"); tagRes != nil {
		t.Fatalf("expected nil on value read error")
	}
}

func TestParseFtyp_ZeroBrandSkipped(t *testing.T) {
	// compat brand of zeros should be skipped
	buf := make([]byte, 20)
	binary.BigEndian.PutUint32(buf[0:4], 20)
	copy(buf[4:8], atomFTYP)
	copy(buf[8:12], "M4A ")
	binary.BigEndian.PutUint32(buf[12:16], 0)
	// compat brand all zeros
	copy(buf[16:20], []byte{0, 0, 0, 0})

	p := New()
	r := bytes.NewReader(buf)
	dirs, _ := p.Parse(r)
	if len(dirs) == 0 {
		t.Fatalf("expected ftyp dir")
	}
	for _, d := range dirs {
		for _, tag := range d.Tags {
			if tag.Name == "Compatible Brands" {
				t.Fatalf("expected zero brand to be skipped")
			}
		}
	}
}

func TestParseFtyp_SmallSize(t *testing.T) {
	atom := buildAtom(atomFTYP, []byte{0x00}) // size <16
	p := New()
	if dir := p.parseFtyp(bytes.NewReader(atom), &Atom{Type: atomFTYP, Size: uint64(len(atom)), Offset: 0}); dir != nil {
		t.Fatalf("expected nil dir for small size")
	}
}

func TestParseFtyp_ReadError(t *testing.T) {
	atom := buildAtom(atomFTYP, []byte("payloadpayload"))
	r := errorReaderAt{err: io.ErrUnexpectedEOF}
	p := New()
	if dir := p.parseFtyp(r, &Atom{Type: atomFTYP, Size: uint64(len(atom)), Offset: 0}); dir != nil {
		t.Fatalf("expected nil dir on read error")
	}
}

func TestGetMetadataTagName(t *testing.T) {
	tests := []struct {
		atomType string
		want     string
	}{
		{"\xa9nam", "Title"},  // © is 0xA9 in MP4 files
		{"\xa9ART", "Artist"}, // © is 0xA9 in MP4 files
		{"\xa9alb", "Album"},  // © is 0xA9 in MP4 files
		{"trkn", "TrackNumber"},
		{"aART", "AlbumArtist"},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.atomType, func(t *testing.T) {
			if got := getMetadataTagName(tt.atomType); got != tt.want {
				t.Errorf("getMetadataTagName(%v) = %v, want %v", tt.atomType, got, tt.want)
			}
		})
	}
}

// Helper function to create ftyp box
func createFtypBox(majorBrand string, minorVersion uint32, compatibleBrands []string) []byte {
	size := uint32(16 + len(compatibleBrands)*4)
	buf := make([]byte, size)

	// Size
	binary.BigEndian.PutUint32(buf[0:4], size)
	// Type
	copy(buf[4:8], "ftyp")
	// Major brand
	copy(buf[8:12], majorBrand)
	// Minor version
	binary.BigEndian.PutUint32(buf[12:16], minorVersion)

	// Compatible brands
	offset := 16
	for _, brand := range compatibleBrands {
		copy(buf[offset:offset+4], brand)
		offset += 4
	}

	return buf
}

// buildMoovWithMeta builds a minimal moov/udta/meta/ilst hierarchy with one string tag
func buildMoovWithMeta() []byte {
	dataAtom := buildDataAtom([]byte("Sample\x00"), dataTypeUTF8)
	titleAtom := buildAtom("\xa9nam", dataAtom)
	ilst := buildAtom(atomILST, titleAtom)

	// meta with version/flags (4 bytes) then ilst
	metaPayload := append(make([]byte, fullBoxHeaderSize), ilst...)
	meta := buildAtom(atomMETA, metaPayload)
	udta := buildAtom(atomUDTA, meta)
	moov := buildAtom(atomMOOV, udta)
	return moov
}

func buildDataAtom(value []byte, typ uint32) []byte {
	var buf bytes.Buffer
	// data atom header
	binary.Write(&buf, binary.BigEndian, uint32(minMetadataAtom+len(value)))
	buf.WriteString(atomDATA)
	binary.Write(&buf, binary.BigEndian, typ) // type/flags
	binary.Write(&buf, binary.BigEndian, uint32(0))
	buf.Write(value)
	return buf.Bytes()
}

// Ensure Parser implements parser.Parser interface
func TestParser_ImplementsInterface(t *testing.T) {
	var _ parser.Parser = (*Parser)(nil)
}

func TestParser_ConcurrentParse(t *testing.T) {
	p := New()
	data := append(createFtypBox("M4A ", 0, []string{"mp42"}), buildMoovWithMeta()...)
	r := bytes.NewReader(data)

	const goroutines = 10
	done := make(chan bool, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			p.Parse(r)
			done <- true
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}
