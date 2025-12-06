package jpeg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/gomantics/imx/internal/meta"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Error("New() returned nil")
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		peek []byte
		want bool
	}{
		{
			name: "valid JPEG signature",
			peek: []byte{0xFF, 0xD8, 0xFF, 0xE0},
			want: true,
		},
		{
			name: "valid JPEG with EXIF APP1",
			peek: []byte{0xFF, 0xD8, 0xFF, 0xE1},
			want: true,
		},
		{
			name: "minimum valid JPEG (just SOI)",
			peek: []byte{0xFF, 0xD8},
			want: true,
		},
		{
			name: "invalid - PNG signature",
			peek: []byte{0x89, 0x50, 0x4E, 0x47},
			want: false,
		},
		{
			name: "invalid - wrong first byte",
			peek: []byte{0x00, 0xD8, 0xFF, 0xE0},
			want: false,
		},
		{
			name: "invalid - wrong second byte",
			peek: []byte{0xFF, 0x00, 0xFF, 0xE0},
			want: false,
		},
		{
			name: "too short - single byte",
			peek: []byte{0xFF},
			want: false,
		},
		{
			name: "empty data",
			peek: []byte{},
			want: false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.Detect(tt.peek)
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

// buildJPEG creates a valid JPEG byte sequence for testing
func buildJPEG(segments ...[]byte) []byte {
	var buf bytes.Buffer
	// Write SOI marker
	buf.Write([]byte{0xFF, 0xD8})
	// Write each segment
	for _, seg := range segments {
		buf.Write(seg)
	}
	// Write EOI marker
	buf.Write([]byte{0xFF, 0xD9})
	return buf.Bytes()
}

// buildSegment creates an APP segment with the given marker and data
func buildSegment(marker byte, data []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(0xFF)
	buf.WriteByte(marker)
	// Length includes the 2 bytes for the length field itself
	length := uint16(len(data) + 2)
	binary.Write(&buf, binary.BigEndian, length)
	buf.Write(data)
	return buf.Bytes()
}

func TestParser_Parse(t *testing.T) {
	// EXIF magic bytes
	exifMagic := []byte("Exif\x00\x00")
	xmpMagic := []byte("http://ns.adobe.com/xap/1.0/\x00")
	iccMagic := []byte("ICC_PROFILE\x00")
	iptcMagic := []byte("Photoshop 3.0\x00")

	// Sample TIFF header (little-endian)
	sampleTIFF := []byte{
		'I', 'I', // Little-endian
		0x2A, 0x00, // TIFF magic (42)
		0x08, 0x00, 0x00, 0x00, // Offset to first IFD
		// Minimal IFD with 0 entries
		0x00, 0x00, // Entry count = 0
		0x00, 0x00, 0x00, 0x00, // Next IFD offset = 0
	}

	tests := []struct {
		name       string
		data       []byte
		wantBlocks int
		wantSpecs  []int
		wantErr    bool
	}{
		{
			name:       "valid JPEG with EXIF",
			data:       buildJPEG(buildSegment(0xE1, append(exifMagic, sampleTIFF...))),
			wantBlocks: 1,
			wantSpecs:  []int{int(meta.SpecEXIF)},
			wantErr:    false,
		},
		{
			name:       "valid JPEG with XMP",
			data:       buildJPEG(buildSegment(0xE1, append(xmpMagic, []byte("<x:xmpmeta></x:xmpmeta>")...))),
			wantBlocks: 1,
			wantSpecs:  []int{int(meta.SpecXMP)},
			wantErr:    false,
		},
		{
			name:       "valid JPEG with ICC profile",
			data:       buildJPEG(buildSegment(0xE2, append(iccMagic, []byte{0x01, 0x01, 0x00, 0x00}...))),
			wantBlocks: 1,
			wantSpecs:  []int{int(meta.SpecICC)},
			wantErr:    false,
		},
		{
			name:       "valid JPEG with IPTC",
			data:       buildJPEG(buildSegment(0xED, append(iptcMagic, []byte{0x38, 0x42, 0x49, 0x4D}...))),
			wantBlocks: 1,
			wantSpecs:  []int{int(meta.SpecIPTC)},
			wantErr:    false,
		},
		{
			name: "valid JPEG with multiple metadata blocks",
			data: buildJPEG(
				buildSegment(0xE1, append(exifMagic, sampleTIFF...)),
				buildSegment(0xE1, append(xmpMagic, []byte("<x:xmpmeta></x:xmpmeta>")...)),
				buildSegment(0xE2, append(iccMagic, []byte{0x01, 0x01}...)),
			),
			wantBlocks: 3,
			wantSpecs:  []int{int(meta.SpecEXIF), int(meta.SpecXMP), int(meta.SpecICC)},
			wantErr:    false,
		},
		{
			name:       "empty JPEG (just SOI/EOI)",
			data:       buildJPEG(),
			wantBlocks: 0,
			wantSpecs:  nil,
			wantErr:    false,
		},
		{
			name:       "JPEG with SOS marker stops parsing",
			data:       append(buildJPEG(buildSegment(0xE1, append(exifMagic, sampleTIFF...)))[:len(buildJPEG(buildSegment(0xE1, append(exifMagic, sampleTIFF...))))-2], []byte{0xFF, 0xDA, 0x00, 0x02, 0xFF, 0xD9}...),
			wantBlocks: 1,
			wantSpecs:  []int{int(meta.SpecEXIF)},
			wantErr:    false,
		},
		{
			name:       "APP1 with unknown magic (ignored)",
			data:       buildJPEG(buildSegment(0xE1, []byte("Unknown\x00\x00some data here"))),
			wantBlocks: 0,
			wantSpecs:  nil,
			wantErr:    false,
		},
		{
			name:       "APP2 with unknown magic (ignored)",
			data:       buildJPEG(buildSegment(0xE2, []byte("NotICC\x00\x00some data"))),
			wantBlocks: 0,
			wantSpecs:  nil,
			wantErr:    false,
		},
		{
			name:       "APP13 with unknown magic (ignored)",
			data:       buildJPEG(buildSegment(0xED, []byte("NotPhotoshop\x00"))),
			wantBlocks: 0,
			wantSpecs:  nil,
			wantErr:    false,
		},
		{
			name:    "invalid SOI marker",
			data:    []byte{0xFF, 0xE0, 0xFF, 0xD9}, // Starts with APP0, not SOI
			wantErr: true,
		},
		{
			name:    "truncated - missing SOI",
			data:    []byte{},
			wantErr: true,
		},
		{
			name: "invalid segment length (too small)",
			data: func() []byte {
				var buf bytes.Buffer
				buf.Write([]byte{0xFF, 0xD8})                   // SOI
				buf.Write([]byte{0xFF, 0xE1})                   // APP1
				binary.Write(&buf, binary.BigEndian, uint16(1)) // Invalid length (less than 2)
				return buf.Bytes()
			}(),
			wantErr: true,
		},
		{
			name: "truncated segment data",
			data: func() []byte {
				var buf bytes.Buffer
				buf.Write([]byte{0xFF, 0xD8})                     // SOI
				buf.Write([]byte{0xFF, 0xE1})                     // APP1
				binary.Write(&buf, binary.BigEndian, uint16(100)) // Claims 98 bytes of data
				buf.Write([]byte{0x01, 0x02, 0x03})               // Only 3 bytes
				return buf.Bytes()
			}(),
			wantErr: true,
		},
		{
			name: "multiple EXIF blocks increment index",
			data: buildJPEG(
				buildSegment(0xE1, append(exifMagic, sampleTIFF...)),
				buildSegment(0xE1, append(exifMagic, sampleTIFF...)),
			),
			wantBlocks: 2,
			wantSpecs:  []int{int(meta.SpecEXIF), int(meta.SpecEXIF)},
			wantErr:    false,
		},
		{
			name: "handles padding 0xFF bytes in markers",
			data: func() []byte {
				var buf bytes.Buffer
				buf.Write([]byte{0xFF, 0xD8})             // SOI
				buf.Write([]byte{0xFF, 0xFF, 0xFF, 0xD9}) // Padded EOI
				return buf.Bytes()
			}(),
			wantBlocks: 0,
			wantSpecs:  nil,
			wantErr:    false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bufio.NewReader(bytes.NewReader(tt.data))
			blocks, err := p.Parse(r)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if len(blocks) != tt.wantBlocks {
				t.Errorf("Parse() returned %d blocks, want %d", len(blocks), tt.wantBlocks)
				return
			}

			for i, wantSpec := range tt.wantSpecs {
				if blocks[i].Spec != wantSpec {
					t.Errorf("blocks[%d].Spec = %d, want %d", i, blocks[i].Spec, wantSpec)
				}
			}
		})
	}
}

func TestParser_Parse_EOF(t *testing.T) {
	// Test EOF handling during marker read - EOF in the main loop is acceptable
	p := New()
	data := []byte{0xFF, 0xD8, 0xFF} // SOI + incomplete marker
	r := bufio.NewReader(bytes.NewReader(data))
	blocks, err := p.Parse(r)
	// EOF during marker reading in main loop breaks gracefully (no error)
	if err != nil {
		t.Errorf("Parse() unexpected error = %v", err)
	}
	if len(blocks) != 0 {
		t.Errorf("Parse() expected 0 blocks, got %d", len(blocks))
	}
}

func TestParser_Parse_SegmentLengthReadError(t *testing.T) {
	// Test error when reading segment length fails
	p := New()
	data := []byte{0xFF, 0xD8, 0xFF, 0xE1, 0x00} // SOI + APP1 marker + incomplete length
	r := bufio.NewReader(bytes.NewReader(data))
	_, err := p.Parse(r)
	if err == nil {
		t.Error("Parse() expected error for incomplete segment length, got nil")
	}
}

// Custom reader that returns EOF immediately
type immediateEOFReader struct{}

func (r immediateEOFReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func TestParser_Parse_EmptyReader(t *testing.T) {
	p := New()
	r := bufio.NewReader(immediateEOFReader{})
	_, err := p.Parse(r)
	if err == nil {
		t.Error("Parse() expected error for empty reader, got nil")
	}
}

// Custom reader that returns an error after SOI
type errorAfterSOIReader struct {
	pos int
}

func (r *errorAfterSOIReader) Read(p []byte) (n int, err error) {
	// Return SOI first, then error
	data := []byte{0xFF, 0xD8, 0xFF}
	if r.pos >= len(data) {
		return 0, io.ErrUnexpectedEOF // Non-EOF error
	}
	n = copy(p, data[r.pos:])
	r.pos += n
	return n, nil
}

func TestParser_Parse_MarkerReadError(t *testing.T) {
	// Test error when reading marker fails with non-EOF error
	p := New()
	r := bufio.NewReader(&errorAfterSOIReader{})
	_, err := p.Parse(r)
	// Should return error for unexpected EOF (not treated as clean EOF)
	if err == nil {
		t.Error("Parse() expected error for marker read failure")
	}
}
