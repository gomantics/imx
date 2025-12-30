package flac

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

// errorReader simulates a reader that always returns errors
type errorReader struct{}

func (errorReader) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestParseStreamInfo(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantNil  bool
		wantTags int
		checkTag func(*testing.T, *parser.Directory)
	}{
		{
			name:     "valid STREAMINFO block",
			data:     makeValidStreamInfo(),
			wantNil:  false,
			wantTags: 10,
			checkTag: func(t *testing.T, dir *parser.Directory) {
				if dir.Name != "FLAC-StreamInfo" {
					t.Errorf("Directory name = %v, want FLAC-STREAMINFO", dir.Name)
				}
				// Check a few specific tags
				found := false
				for _, tag := range dir.Tags {
					if tag.Name == "SampleRate" {
						found = true
						if tag.Value != uint32(44100) {
							t.Errorf("SampleRate = %v, want 44100", tag.Value)
						}
					}
				}
				if !found {
					t.Error("SampleRate tag not found")
				}
			},
		},
		{
			name:     "too short - less than 34 bytes",
			data:     make([]byte, 33),
			wantNil:  true,
			wantTags: 0,
		},
		{
			name:     "empty data",
			data:     []byte{},
			wantNil:  true,
			wantTags: 0,
		},
		{
			name:     "exactly 34 bytes minimum",
			data:     make([]byte, 34),
			wantNil:  false,
			wantTags: 9, // No duration if sample rate is 0
		},
		{
			name:     "StreamInfo with zero sample rate (no duration tag)",
			data:     makeStreamInfoWithZeroSampleRate(),
			wantNil:  false,
			wantTags: 9, // Duration tag is not added when sample rate is 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			r := bytes.NewReader(tt.data)
			dir := p.parseStreamInfo(r, 0, int64(len(tt.data)))

			if tt.wantNil && dir != nil {
				t.Errorf("parseStreamInfo() = %v, want nil", dir)
				return
			}
			if !tt.wantNil && dir == nil {
				t.Error("parseStreamInfo() = nil, want non-nil")
				return
			}
			if dir != nil {
				if len(dir.Tags) != tt.wantTags {
					t.Errorf("parseStreamInfo() tags count = %d, want %d", len(dir.Tags), tt.wantTags)
				}
				if tt.checkTag != nil {
					tt.checkTag(t, dir)
				}
			}
		})
	}
}

func TestParseStreamInfo_ReadError(t *testing.T) {
	p := New()
	dir := p.parseStreamInfo(errorReader{}, 0, 100)
	if dir != nil {
		t.Error("parseStreamInfo() with read error should return nil")
	}
}

func TestParseVorbisComment_ReadError(t *testing.T) {
	p := New()
	dir := p.parseVorbisComment(errorReader{}, 0, 100)
	if dir != nil {
		t.Error("parseVorbisComment() with read error should return nil")
	}
}

func TestParsePicture_ReadError(t *testing.T) {
	p := New()
	dir := p.parsePicture(errorReader{}, 0, 100)
	if dir != nil {
		t.Error("parsePicture() with read error should return nil")
	}
}

func TestParseApplication_ReadError(t *testing.T) {
	p := New()
	dir := p.parseApplication(errorReader{}, 0, 100)
	if dir != nil {
		t.Error("parseApplication() with read error should return nil")
	}
}

func TestParseSeekTable_ReadError(t *testing.T) {
	p := New()
	dir := p.parseSeekTable(errorReader{}, 0, 36) // 36 = valid multiple of 18
	if dir != nil {
		t.Error("parseSeekTable() with read error should return nil")
	}
}

func TestParseVorbisComment(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantNil   bool
		minTags   int
		checkTags func(*testing.T, *parser.Directory)
	}{
		{
			name:    "valid Vorbis comment with vendor and tags",
			data:    makeValidVorbisComment(),
			wantNil: false,
			minTags: 1, // At least vendor
			checkTags: func(t *testing.T, dir *parser.Directory) {
				if dir.Name != "FLAC-Vorbis" {
					t.Errorf("Directory name = %v, want FLAC-VORBIS", dir.Name)
				}
				hasVendor := false
				for _, tag := range dir.Tags {
					if tag.Name == "Vendor" {
						hasVendor = true
					}
				}
				if !hasVendor {
					t.Error("Vendor tag not found")
				}
			},
		},
		{
			name:    "Vorbis comment with vendor only (no tags)",
			data:    makeVorbisCommentVendorOnly(),
			wantNil: false,
			minTags: 1, // Just vendor
		},
		{
			name:    "Vorbis comment with multiple tags",
			data:    makeVorbisCommentMultipleTags(),
			wantNil: false,
			minTags: 4, // vendor + 3 comments
			checkTags: func(t *testing.T, dir *parser.Directory) {
				found := make(map[string]bool)
				for _, tag := range dir.Tags {
					found[tag.Name] = true
				}
				if !found["ARTIST"] {
					t.Error("ARTIST tag not found")
				}
				if !found["TITLE"] {
					t.Error("TITLE tag not found")
				}
			},
		},
		{
			name:    "truncated - no vendor length",
			data:    []byte{},
			wantNil: true, // Returns nil when ReadAt fails
			minTags: 0,
		},
		{
			name:    "truncated - only 3 bytes (not enough for vendor length)",
			data:    []byte{0x01, 0x02, 0x03},
			wantNil: false,
			minTags: 0, // Early return, empty directory
		},
		{
			name:    "truncated - incomplete vendor string",
			data:    []byte{0x10, 0x00, 0x00, 0x00}, // vendor length = 16, but no data
			wantNil: false,
			minTags: 0,
		},
		{
			name:    "truncated - no comment count",
			data:    append([]byte{0x04, 0x00, 0x00, 0x00}, []byte("test")...), // vendor only
			wantNil: false,
			minTags: 1, // Just vendor
		},
		{
			name:    "comment without equals sign (invalid format)",
			data:    makeVorbisCommentInvalidFormat(),
			wantNil: false,
			minTags: 1, // vendor + skipped invalid comment
		},
		{
			name:    "truncated comment data - comment length exceeds remaining data",
			data:    makeVorbisCommentTruncatedComment(),
			wantNil: false,
			minTags: 1, // vendor only, comment skipped due to truncation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			r := bytes.NewReader(tt.data)
			dir := p.parseVorbisComment(r, 0, int64(len(tt.data)))

			if tt.wantNil && dir != nil {
				t.Errorf("parseVorbisComment() = %v, want nil", dir)
				return
			}
			if !tt.wantNil && dir == nil {
				t.Error("parseVorbisComment() = nil, want non-nil")
				return
			}
			if dir != nil {
				if len(dir.Tags) < tt.minTags {
					t.Errorf("parseVorbisComment() tags count = %d, want at least %d", len(dir.Tags), tt.minTags)
				}
				if tt.checkTags != nil {
					tt.checkTags(t, dir)
				}
			}
		})
	}
}

func TestParsePicture(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantNil   bool
		minTags   int
		checkTags func(*testing.T, *parser.Directory)
	}{
		{
			name:    "valid picture with all fields",
			data:    makeValidPicture(),
			wantNil: false,
			minTags: 6, // type, mime, width, height, depth, size (no description, colors=0 so not included)
			checkTags: func(t *testing.T, dir *parser.Directory) {
				if dir.Name != "FLAC-Picture" {
					t.Errorf("Directory name = %v, want FLAC-PICTURE", dir.Name)
				}
			},
		},
		{
			name:    "picture with description",
			data:    makeValidPictureWithDescription(),
			wantNil: false,
			minTags: 7, // includes description (colors=0 not included)
		},
		{
			name:    "picture with colors > 0",
			data:    makeValidPictureWithColors(),
			wantNil: false,
			minTags: 7, // includes colors tag (no description)
		},
		{
			name:    "picture without description (empty string)",
			data:    makeValidPictureNoDescription(),
			wantNil: false,
			minTags: 6, // no description tag, no colors tag
		},
		{
			name:    "truncated - no picture type",
			data:    []byte{},
			wantNil: true, // Returns nil when ReadAt fails
			minTags: 0,
		},
		{
			name:    "truncated - only 3 bytes (not enough for picture type)",
			data:    []byte{0x01, 0x02, 0x03},
			wantNil: false,
			minTags: 0, // Early return, empty directory
		},
		{
			name:    "truncated - no MIME length",
			data:    make([]byte, 4),
			wantNil: false,
			minTags: 1, // just picture type
		},
		{
			name:    "truncated - incomplete MIME type",
			data:    append(make([]byte, 4), []byte{0x10, 0x00, 0x00, 0x00}...),
			wantNil: false,
			minTags: 1,
		},
		{
			name:    "truncated - no description length",
			data:    makePictureTruncatedAtDescription(),
			wantNil: false,
			minTags: 2, // type and mime
		},
		{
			name:    "truncated - description length but not enough data",
			data:    makePictureTruncatedDescription(),
			wantNil: false,
			minTags: 2, // type and mime, description header present but data truncated
		},
		{
			name:    "truncated - no width/height/depth/colors",
			data:    makePictureTruncatedAtDimensions(),
			wantNil: false,
			minTags: 2, // type and mime
		},
		{
			name:    "truncated - no picture data length",
			data:    makePictureTruncatedAtDataLength(),
			wantNil: false,
			minTags: 5, // type, mime, width, height, depth (no colors tag if 0, no size tag because truncated)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			r := bytes.NewReader(tt.data)
			dir := p.parsePicture(r, 0, int64(len(tt.data)))

			if tt.wantNil && dir != nil {
				t.Errorf("parsePicture() = %v, want nil", dir)
				return
			}
			if !tt.wantNil && dir == nil {
				t.Error("parsePicture() = nil, want non-nil")
				return
			}
			if dir != nil {
				if len(dir.Tags) < tt.minTags {
					t.Errorf("parsePicture() tags count = %d, want at least %d", len(dir.Tags), tt.minTags)
				}
				if tt.checkTags != nil {
					tt.checkTags(t, dir)
				}
			}
		})
	}
}

func TestParsePadding(t *testing.T) {
	tests := []struct {
		name   string
		length int64
		want   string
	}{
		{"padding 0 bytes", 0, "0 bytes"},
		{"padding 100 bytes", 100, "100 bytes"},
		{"padding 8192 bytes", 8192, "8192 bytes"},
		{"padding 1MB", 1048576, "1048576 bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			dir := p.parsePadding(tt.length)
			if dir == nil {
				t.Fatal("parsePadding() returned nil")
			}
			if dir.Name != "FLAC-Padding" {
				t.Errorf("Directory name = %v, want FLAC-PADDING", dir.Name)
			}
			if len(dir.Tags) != 1 {
				t.Errorf("parsePadding() tags count = %d, want 1", len(dir.Tags))
			}
			if dir.Tags[0].Value != tt.want {
				t.Errorf("parsePadding() value = %v, want %v", dir.Tags[0].Value, tt.want)
			}
		})
	}
}

func TestParseApplication(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantNil   bool
		wantTags  int
		checkTags func(*testing.T, *parser.Directory)
	}{
		{
			name:     "valid application block",
			data:     []byte("TEST" + string(make([]byte, 100))),
			wantNil:  false,
			wantTags: 2, // ID and DataSize
			checkTags: func(t *testing.T, dir *parser.Directory) {
				if dir.Name != "FLAC-Application" {
					t.Errorf("Directory name = %v, want FLAC-APPLICATION", dir.Name)
				}
				if dir.Tags[0].Name == "ApplicationID" && dir.Tags[0].Value != "TEST" {
					t.Errorf("ApplicationID = %v, want TEST", dir.Tags[0].Value)
				}
			},
		},
		{
			name:     "minimum size - exactly 4 bytes",
			data:     []byte("ABCD"),
			wantNil:  false,
			wantTags: 2,
		},
		{
			name:     "too short - less than 4 bytes",
			data:     []byte("ABC"),
			wantNil:  true,
			wantTags: 0,
		},
		{
			name:     "empty data",
			data:     []byte{},
			wantNil:  true,
			wantTags: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			r := bytes.NewReader(tt.data)
			dir := p.parseApplication(r, 0, int64(len(tt.data)))

			if tt.wantNil && dir != nil {
				t.Errorf("parseApplication() = %v, want nil", dir)
				return
			}
			if !tt.wantNil && dir == nil {
				t.Error("parseApplication() = nil, want non-nil")
				return
			}
			if dir != nil {
				if len(dir.Tags) != tt.wantTags {
					t.Errorf("parseApplication() tags count = %d, want %d", len(dir.Tags), tt.wantTags)
				}
				if tt.checkTags != nil {
					tt.checkTags(t, dir)
				}
			}
		})
	}
}

func TestParseSeekTable(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantNil   bool
		wantTags  int
		checkTags func(*testing.T, *parser.Directory)
	}{
		{
			name:     "valid seek table - 2 points",
			data:     make([]byte, 36), // 2 * 18 bytes
			wantNil:  false,
			wantTags: 1,
			checkTags: func(t *testing.T, dir *parser.Directory) {
				if dir.Name != "FLAC-SeekTable" {
					t.Errorf("Directory name = %v, want FLAC-SEEKTABLE", dir.Name)
				}
				if dir.Tags[0].Value != int64(2) {
					t.Errorf("SeekPoints = %v, want 2", dir.Tags[0].Value)
				}
			},
		},
		{
			name:     "valid seek table - 1 point",
			data:     make([]byte, 18), // 1 * 18 bytes
			wantNil:  false,
			wantTags: 1,
		},
		{
			name:     "valid seek table - 10 points",
			data:     make([]byte, 180), // 10 * 18 bytes
			wantNil:  false,
			wantTags: 1,
		},
		{
			name:     "invalid - not multiple of 18",
			data:     make([]byte, 19),
			wantNil:  true,
			wantTags: 0,
		},
		{
			name:     "invalid - not multiple of 18 (17 bytes)",
			data:     make([]byte, 17),
			wantNil:  true,
			wantTags: 0,
		},
		{
			name:     "empty data (0 points, but valid)",
			data:     []byte{},
			wantNil:  false,
			wantTags: 1,
			checkTags: func(t *testing.T, dir *parser.Directory) {
				if dir.Tags[0].Value != int64(0) {
					t.Errorf("SeekPoints = %v, want 0", dir.Tags[0].Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			r := bytes.NewReader(tt.data)
			dir := p.parseSeekTable(r, 0, int64(len(tt.data)))

			if tt.wantNil && dir != nil {
				t.Errorf("parseSeekTable() = %v, want nil", dir)
				return
			}
			if !tt.wantNil && dir == nil {
				t.Error("parseSeekTable() = nil, want non-nil")
				return
			}
			if dir != nil {
				if len(dir.Tags) != tt.wantTags {
					t.Errorf("parseSeekTable() tags count = %d, want %d", len(dir.Tags), tt.wantTags)
				}
				if tt.checkTags != nil {
					tt.checkTags(t, dir)
				}
			}
		})
	}
}

func TestParseCueSheet(t *testing.T) {
	tests := []struct {
		name   string
		length int64
		want   string
	}{
		{"cue sheet 0 bytes", 0, "0 bytes"},
		{"cue sheet 100 bytes", 100, "100 bytes"},
		{"cue sheet 1024 bytes", 1024, "1024 bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			dir := p.parseCueSheet(tt.length)
			if dir == nil {
				t.Fatal("parseCueSheet() returned nil")
			}
			if dir.Name != "FLAC-CueSheet" {
				t.Errorf("Directory name = %v, want FLAC-CUESHEET", dir.Name)
			}
			if len(dir.Tags) != 1 {
				t.Errorf("parseCueSheet() tags count = %d, want 1", len(dir.Tags))
			}
			if dir.Tags[0].Value != tt.want {
				t.Errorf("parseCueSheet() value = %v, want %v", dir.Tags[0].Value, tt.want)
			}
		})
	}
}

// Helper functions to create test data

func makeValidStreamInfo() []byte {
	data := make([]byte, 34)
	// Min block size
	binary.BigEndian.PutUint16(data[0:2], 4096)
	// Max block size
	binary.BigEndian.PutUint16(data[2:4], 4096)
	// Min frame size (24-bit)
	data[4] = 0x00
	data[5] = 0x10
	data[6] = 0x00
	// Max frame size (24-bit)
	data[7] = 0x00
	data[8] = 0x20
	data[9] = 0x00
	// Sample rate (44100 Hz = 0xAC44), channels (2), bits per sample (16)
	// Sample rate is 20 bits: 44100 = 0x0AC44
	data[10] = 0x0A // high byte
	data[11] = 0xC4 // middle byte
	data[12] = 0x42 // low 4 bits (0x4) + channels (1 = stereo-1) shifted + bits (15 = 16-1) low bit
	data[13] = 0xF0 // bits per sample top 4 bits (0xF = 15, representing 16 bits)
	// Total samples
	binary.BigEndian.PutUint32(data[14:18], 100000)
	// MD5
	copy(data[18:34], []byte{0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89})
	return data
}

func makeStreamInfoWithZeroSampleRate() []byte {
	data := make([]byte, 34)
	// All zeros except MD5
	copy(data[18:34], []byte{0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89})
	return data
}

func makeValidVorbisComment() []byte {
	buf := &bytes.Buffer{}
	// Vendor string
	vendor := "TestVendor"
	binary.Write(buf, binary.LittleEndian, uint32(len(vendor)))
	buf.WriteString(vendor)
	// Number of comments
	binary.Write(buf, binary.LittleEndian, uint32(2))
	// Comment 1
	comment1 := "ARTIST=Test Artist"
	binary.Write(buf, binary.LittleEndian, uint32(len(comment1)))
	buf.WriteString(comment1)
	// Comment 2
	comment2 := "TITLE=Test Title"
	binary.Write(buf, binary.LittleEndian, uint32(len(comment2)))
	buf.WriteString(comment2)
	return buf.Bytes()
}

func makeVorbisCommentVendorOnly() []byte {
	buf := &bytes.Buffer{}
	vendor := "TestVendor"
	binary.Write(buf, binary.LittleEndian, uint32(len(vendor)))
	buf.WriteString(vendor)
	binary.Write(buf, binary.LittleEndian, uint32(0)) // 0 comments
	return buf.Bytes()
}

func makeVorbisCommentMultipleTags() []byte {
	buf := &bytes.Buffer{}
	vendor := "TestVendor"
	binary.Write(buf, binary.LittleEndian, uint32(len(vendor)))
	buf.WriteString(vendor)
	binary.Write(buf, binary.LittleEndian, uint32(3))

	comments := []string{"ARTIST=Artist", "TITLE=Title", "ALBUM=Album"}
	for _, c := range comments {
		binary.Write(buf, binary.LittleEndian, uint32(len(c)))
		buf.WriteString(c)
	}
	return buf.Bytes()
}

func makeVorbisCommentInvalidFormat() []byte {
	buf := &bytes.Buffer{}
	vendor := "TestVendor"
	binary.Write(buf, binary.LittleEndian, uint32(len(vendor)))
	buf.WriteString(vendor)
	binary.Write(buf, binary.LittleEndian, uint32(1))
	// Comment without = sign
	comment := "INVALIDCOMMENT"
	binary.Write(buf, binary.LittleEndian, uint32(len(comment)))
	buf.WriteString(comment)
	return buf.Bytes()
}

func makeVorbisCommentTruncatedComment() []byte {
	buf := &bytes.Buffer{}
	vendor := "TestVendor"
	binary.Write(buf, binary.LittleEndian, uint32(len(vendor)))
	buf.WriteString(vendor)
	binary.Write(buf, binary.LittleEndian, uint32(1)) // 1 comment
	// Comment length says 100 bytes but we only provide 5
	binary.Write(buf, binary.LittleEndian, uint32(100))
	buf.WriteString("SHORT") // Only 5 bytes, not 100
	return buf.Bytes()
}

func makeValidPicture() []byte {
	buf := &bytes.Buffer{}
	// Picture type
	binary.Write(buf, binary.BigEndian, uint32(3)) // Cover (front)
	// MIME type
	mime := "image/jpeg"
	binary.Write(buf, binary.BigEndian, uint32(len(mime)))
	buf.WriteString(mime)
	// Description (empty)
	binary.Write(buf, binary.BigEndian, uint32(0))
	// Width, height, depth, colors
	binary.Write(buf, binary.BigEndian, uint32(640))
	binary.Write(buf, binary.BigEndian, uint32(480))
	binary.Write(buf, binary.BigEndian, uint32(24))
	binary.Write(buf, binary.BigEndian, uint32(0)) // colors = 0, won't be added as tag
	// Picture data length
	binary.Write(buf, binary.BigEndian, uint32(1024))
	return buf.Bytes()
}

func makeValidPictureWithDescription() []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint32(3))
	mime := "image/png"
	binary.Write(buf, binary.BigEndian, uint32(len(mime)))
	buf.WriteString(mime)
	desc := "Album Cover"
	binary.Write(buf, binary.BigEndian, uint32(len(desc)))
	buf.WriteString(desc)
	binary.Write(buf, binary.BigEndian, uint32(640))
	binary.Write(buf, binary.BigEndian, uint32(480))
	binary.Write(buf, binary.BigEndian, uint32(24))
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(1024))
	return buf.Bytes()
}

func makeValidPictureWithColors() []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint32(3))
	mime := "image/png"
	binary.Write(buf, binary.BigEndian, uint32(len(mime)))
	buf.WriteString(mime)
	binary.Write(buf, binary.BigEndian, uint32(0)) // no description
	binary.Write(buf, binary.BigEndian, uint32(640))
	binary.Write(buf, binary.BigEndian, uint32(480))
	binary.Write(buf, binary.BigEndian, uint32(8))
	binary.Write(buf, binary.BigEndian, uint32(256)) // colors > 0
	binary.Write(buf, binary.BigEndian, uint32(1024))
	return buf.Bytes()
}

func makeValidPictureNoDescription() []byte {
	return makeValidPicture() // Already has empty description
}

func makePictureTruncatedAtDescription() []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint32(3))
	mime := "image/jpeg"
	binary.Write(buf, binary.BigEndian, uint32(len(mime)))
	buf.WriteString(mime)
	// Stop here - no description length
	return buf.Bytes()
}

func makePictureTruncatedDescription() []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint32(3))
	mime := "image/jpeg"
	binary.Write(buf, binary.BigEndian, uint32(len(mime)))
	buf.WriteString(mime)
	// Description length says 100 bytes but we only provide 5
	binary.Write(buf, binary.BigEndian, uint32(100))
	buf.WriteString("SHORT") // Only 5 bytes, not 100
	return buf.Bytes()
}

func makePictureTruncatedAtDimensions() []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint32(3))
	mime := "image/jpeg"
	binary.Write(buf, binary.BigEndian, uint32(len(mime)))
	buf.WriteString(mime)
	binary.Write(buf, binary.BigEndian, uint32(0)) // description length
	// Stop here - no dimensions
	return buf.Bytes()
}

func makePictureTruncatedAtDataLength() []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint32(3))
	mime := "image/jpeg"
	binary.Write(buf, binary.BigEndian, uint32(len(mime)))
	buf.WriteString(mime)
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(640))
	binary.Write(buf, binary.BigEndian, uint32(480))
	binary.Write(buf, binary.BigEndian, uint32(24))
	binary.Write(buf, binary.BigEndian, uint32(0))
	// Stop here - no picture data length
	return buf.Bytes()
}
