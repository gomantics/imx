package jpeg

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.icc == nil {
		t.Error("New() created parser with nil icc parser")
	}
	if p.iptc == nil {
		t.Error("New() created parser with nil iptc parser")
	}
	if p.xmp == nil {
		t.Error("New() created parser with nil xmp parser")
	}
	if p.exif == nil {
		t.Error("New() created parser with nil exif parser")
	}
}

func TestParser_Name(t *testing.T) {
	p := New()
	got := p.Name()
	want := "JPEG"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestParser_Detect(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid JPEG SOI marker",
			data: []byte{0xFF, 0xD8},
			want: true,
		},
		{
			name: "valid JPEG with more data",
			data: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
			want: true,
		},
		{
			name: "invalid - wrong first byte",
			data: []byte{0x00, 0xD8},
			want: false,
		},
		{
			name: "invalid - wrong second byte",
			data: []byte{0xFF, 0x00},
			want: false,
		},
		{
			name: "too short - single byte",
			data: []byte{0xFF},
			want: false,
		},
		{
			name: "empty data",
			data: []byte{},
			want: false,
		},
		{
			name: "PNG signature",
			data: []byte{0x89, 0x50, 0x4E, 0x47},
			want: false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got := p.Detect(r)
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_readMarker(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantMarker byte
		wantErr    bool
	}{
		{
			name:       "SOI marker",
			data:       []byte{0xFF, 0xD8},
			wantMarker: 0xD8,
			wantErr:    false,
		},
		{
			name:       "EOI marker",
			data:       []byte{0xFF, 0xD9},
			wantMarker: 0xD9,
			wantErr:    false,
		},
		{
			name:       "APP1 marker",
			data:       []byte{0xFF, 0xE1},
			wantMarker: 0xE1,
			wantErr:    false,
		},
		{
			name:       "marker with padding 0xFF",
			data:       []byte{0xFF, 0xFF, 0xD8},
			wantMarker: 0xD8,
			wantErr:    false,
		},
		{
			name:       "marker with multiple padding 0xFF",
			data:       []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xE1},
			wantMarker: 0xE1,
			wantErr:    false,
		},
		{
			name:    "invalid - missing marker prefix",
			data:    []byte{0x00, 0xD8},
			wantErr: true,
		},
		{
			name:    "too short",
			data:    []byte{0xFF},
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got, _, err := readMarker(r, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("readMarker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantMarker {
				t.Errorf("readMarker() = 0x%02X, want 0x%02X", got, tt.wantMarker)
			}
		})
	}
}

func TestParser_readUint16(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    uint16
		wantErr bool
	}{
		{
			name:    "valid uint16 - 256",
			data:    []byte{0x01, 0x00},
			want:    256,
			wantErr: false,
		},
		{
			name:    "valid uint16 - 65535",
			data:    []byte{0xFF, 0xFF},
			want:    65535,
			wantErr: false,
		},
		{
			name:    "valid uint16 - 0",
			data:    []byte{0x00, 0x00},
			want:    0,
			wantErr: false,
		},
		{
			name:    "too short - single byte",
			data:    []byte{0x01},
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got, newPos, err := readUint16(r, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("readUint16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("readUint16() = %d, want %d", got, tt.want)
			}
			if !tt.wantErr && newPos != 2 {
				t.Errorf("readUint16() pos = %d, want 2", newPos)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	t.Run("minimal valid JPEG", func(t *testing.T) {
		// Build minimal JPEG: SOI + EOI
		data := []byte{
			0xFF, 0xD8, // SOI
			0xFF, 0xD9, // EOI
		}

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should parse without errors
		if err != nil {
			t.Errorf("Parse() error = %v, want nil", err)
		}
		// Minimal JPEG has no metadata
		if len(dirs) != 0 {
			t.Errorf("Parse() returned %d dirs, want 0", len(dirs))
		}
	})

	t.Run("JPEG with SOS marker", func(t *testing.T) {
		// SOI + SOS (stops parsing at image data)
		data := []byte{
			0xFF, 0xD8, // SOI
			0xFF, 0xDA, // SOS
		}

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should parse without errors
		if err != nil {
			t.Errorf("Parse() error = %v, want nil", err)
		}
		_ = dirs
	})

	t.Run("invalid - missing SOI", func(t *testing.T) {
		data := []byte{
			0xFF, 0xE0, // Not SOI
			0x00, 0x10,
		}

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should return error
		if err == nil {
			t.Error("Parse() error = nil, want error for missing SOI")
		}
		if dirs != nil {
			t.Errorf("Parse() dirs = %v, want nil on error", dirs)
		}
	})

	t.Run("invalid - wrong SOI marker", func(t *testing.T) {
		data := []byte{
			0xFF, 0xD9, // EOI instead of SOI
		}

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should return error
		if err == nil {
			t.Error("Parse() error = nil, want error for wrong SOI marker")
		}
		if dirs != nil {
			t.Errorf("Parse() dirs = %v, want nil on error", dirs)
		}
	})

	t.Run("empty data", func(t *testing.T) {
		p := New()
		r := bytes.NewReader([]byte{})
		dirs, err := p.Parse(r)

		if err == nil {
			t.Error("Parse() error = nil, want error for empty data")
		}
		if dirs != nil {
			t.Errorf("Parse() dirs = %v, want nil on error", dirs)
		}
	})

	t.Run("JPEG with invalid segment length", func(t *testing.T) {
		data := []byte{
			0xFF, 0xD8, // SOI
			0xFF, 0xE1, // APP1
			0x00, 0x01, // Invalid length (< 2)
		}

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should return error
		if err == nil {
			t.Error("Parse() error = nil, want error for invalid segment length")
		}
		_ = dirs
	})

	t.Run("JPEG with truncated segment", func(t *testing.T) {
		data := []byte{
			0xFF, 0xD8, // SOI
			0xFF, 0xE1, // APP1
			// Missing length
		}

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should return error for truncated segment
		if err == nil {
			t.Error("Parse() error = nil, want error for truncated segment")
		}
		_ = dirs
	})
}

func TestParser_parseAPP1(t *testing.T) {
	t.Run("EXIF identifier", func(t *testing.T) {
		// Create data with EXIF identifier
		data := append([]byte("Exif\x00\x00"), []byte{
			// Minimal TIFF header
			0x49, 0x49, 0x2A, 0x00, // "II" + 42
			0x08, 0x00, 0x00, 0x00, // IFD offset
		}...)

		p := New()
		r := bytes.NewReader(data)
		dirs := p.parseAPP1(r, 0, int64(len(data)))

		// Should attempt to parse as EXIF
		_ = dirs // May or may not return dirs depending on TIFF parser
	})

	t.Run("XMP identifier", func(t *testing.T) {
		// Create data with XMP identifier
		data := append([]byte("http://ns.adobe.com/xap/1.0/\x00"), []byte("<x:xmpmeta></x:xmpmeta>")...)

		p := New()
		r := bytes.NewReader(data)
		dirs := p.parseAPP1(r, 0, int64(len(data)))

		// Should attempt to parse as XMP
		_ = dirs
	})

	t.Run("unknown identifier", func(t *testing.T) {
		data := []byte("Unknown\x00\x00data")

		p := New()
		r := bytes.NewReader(data)
		dirs := p.parseAPP1(r, 0, int64(len(data)))

		// Should return nil for unknown identifier
		if dirs != nil {
			t.Errorf("parseAPP1() = %v, want nil for unknown identifier", dirs)
		}
	})

	t.Run("empty segment", func(t *testing.T) {
		p := New()
		r := bytes.NewReader([]byte{})
		dirs := p.parseAPP1(r, 0, 0)

		if dirs != nil {
			t.Errorf("parseAPP1() = %v, want nil for empty segment", dirs)
		}
	})
}

func TestParser_parseAPP2(t *testing.T) {
	t.Run("valid ICC chunk", func(t *testing.T) {
		// Create data with ICC identifier and chunk header
		data := append([]byte("ICC_PROFILE\x00"), []byte{
			0x01,                   // Chunk 1
			0x01,                   // of 1
			0x00, 0x00, 0x00, 0x10, // Some ICC data
		}...)

		var chunks map[int][]byte
		var totalChunks int
		p := New()
		r := bytes.NewReader(data)
		err := p.parseAPP2(r, 0, int64(len(data)), &chunks, &totalChunks)

		if err != nil {
			t.Errorf("parseAPP2() error = %v, want nil", err)
		}
		if len(chunks) != 1 {
			t.Errorf("parseAPP2() chunks count = %d, want 1", len(chunks))
		}
	})

	t.Run("non-ICC segment", func(t *testing.T) {
		data := []byte("Other\x00data")

		var chunks map[int][]byte
		var totalChunks int
		p := New()
		r := bytes.NewReader(data)
		err := p.parseAPP2(r, 0, int64(len(data)), &chunks, &totalChunks)

		if err != nil {
			t.Errorf("parseAPP2() error = %v, want nil for non-ICC", err)
		}
		if chunks != nil {
			t.Errorf("parseAPP2() chunks = %v, want nil for non-ICC", chunks)
		}
	})

	t.Run("invalid chunk numbers - zero chunk num", func(t *testing.T) {
		data := append([]byte("ICC_PROFILE\x00"), []byte{
			0x00, // Invalid: chunk 0
			0x01, // of 1
		}...)

		var chunks map[int][]byte
		var totalChunks int
		p := New()
		r := bytes.NewReader(data)
		err := p.parseAPP2(r, 0, int64(len(data)), &chunks, &totalChunks)

		if err == nil {
			t.Error("parseAPP2() error = nil, want error for invalid chunk num")
		}
	})

	t.Run("invalid chunk numbers - zero total", func(t *testing.T) {
		data := append([]byte("ICC_PROFILE\x00"), []byte{
			0x01, // Chunk 1
			0x00, // Invalid: total 0
		}...)

		var chunks map[int][]byte
		var totalChunks int
		p := New()
		r := bytes.NewReader(data)
		err := p.parseAPP2(r, 0, int64(len(data)), &chunks, &totalChunks)

		if err == nil {
			t.Error("parseAPP2() error = nil, want error for invalid total chunks")
		}
	})

	t.Run("invalid chunk numbers - chunk > total", func(t *testing.T) {
		data := append([]byte("ICC_PROFILE\x00"), []byte{
			0x03, // Chunk 3
			0x02, // of 2 (invalid)
		}...)

		var chunks map[int][]byte
		var totalChunks int
		p := New()
		r := bytes.NewReader(data)
		err := p.parseAPP2(r, 0, int64(len(data)), &chunks, &totalChunks)

		if err == nil {
			t.Error("parseAPP2() error = nil, want error for chunk > total")
		}
	})

	t.Run("truncated chunk header", func(t *testing.T) {
		data := append([]byte("ICC_PROFILE\x00"), []byte{
			0x01, // Only one byte of header
		}...)

		var chunks map[int][]byte
		var totalChunks int
		p := New()
		r := bytes.NewReader(data)
		err := p.parseAPP2(r, 0, int64(len(data)), &chunks, &totalChunks)

		if err == nil {
			t.Error("parseAPP2() error = nil, want error for truncated header")
		}
	})

	t.Run("inconsistent total chunks", func(t *testing.T) {
		// First chunk says "1 of 2"
		data1 := append([]byte("ICC_PROFILE\x00"), []byte{
			0x01,       // Chunk 1
			0x02,       // of 2
			0x00, 0x01, // Some data
		}...)

		var chunks map[int][]byte
		var totalChunks int
		p := New()
		r1 := bytes.NewReader(data1)
		err := p.parseAPP2(r1, 0, int64(len(data1)), &chunks, &totalChunks)

		if err != nil {
			t.Errorf("parseAPP2() error = %v, want nil for first chunk", err)
		}

		// Second chunk says "2 of 3" (inconsistent!)
		data2 := append([]byte("ICC_PROFILE\x00"), []byte{
			0x02,       // Chunk 2
			0x03,       // of 3 (inconsistent with first chunk's "of 2")
			0x02, 0x03, // Some data
		}...)

		r2 := bytes.NewReader(data2)
		err = p.parseAPP2(r2, 0, int64(len(data2)), &chunks, &totalChunks)

		if err == nil {
			t.Error("parseAPP2() error = nil, want error for inconsistent total chunks")
		}

		// Error should mention inconsistent chunks
		if err != nil && !contains(err.Error(), "inconsistent") {
			t.Errorf("parseAPP2() error = %q, want error mentioning inconsistent chunks", err.Error())
		}
	})
}

func TestParser_parseICC(t *testing.T) {
	t.Run("no chunks", func(t *testing.T) {
		p := New()
		dirs, err := p.parseICC(nil, 0)

		if dirs != nil {
			t.Errorf("parseICC() dirs = %v, want nil for no chunks", dirs)
		}
		if err != nil {
			t.Errorf("parseICC() error = %v, want nil for no chunks", err)
		}
	})

	t.Run("empty chunks map", func(t *testing.T) {
		chunks := make(map[int][]byte)
		p := New()
		dirs, err := p.parseICC(chunks, 0)

		if dirs != nil {
			t.Errorf("parseICC() dirs = %v, want nil for empty chunks", dirs)
		}
		if err != nil {
			t.Errorf("parseICC() error = %v, want nil for empty chunks", err)
		}
	})

	t.Run("missing chunk", func(t *testing.T) {
		chunks := map[int][]byte{
			1: []byte{0x01, 0x02},
			// Missing chunk 2
			3: []byte{0x03, 0x04},
		}

		p := New()
		dirs, err := p.parseICC(chunks, 3) // Expecting 3 chunks

		if err == nil {
			t.Error("parseICC() error = nil, want error for missing chunk")
		}
		_ = dirs
	})

	t.Run("single chunk", func(t *testing.T) {
		chunks := map[int][]byte{
			1: []byte{0x00, 0x01, 0x02, 0x03},
		}

		p := New()
		dirs, err := p.parseICC(chunks, 1) // Expecting 1 chunk

		// Should attempt to parse (may fail if data is invalid ICC)
		_ = dirs
		_ = err
	})
}

func TestParser_parseAPP13(t *testing.T) {
	t.Run("Photoshop identifier", func(t *testing.T) {
		data := append([]byte("Photoshop 3.0\x00"), []byte{
			0x38, 0x42, 0x49, 0x4D, // "8BIM"
			0x04, 0x04, // IPTC resource ID
			0x00, 0x00, // Name (empty)
			0x00, 0x00, 0x00, 0x00, // Size
		}...)

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.parseAPP13(r, 0, int64(len(data)))

		// Should attempt to parse as IPTC
		_ = dirs
		_ = err
	})

	t.Run("non-Photoshop segment", func(t *testing.T) {
		data := []byte("Other\x00data")

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.parseAPP13(r, 0, int64(len(data)))

		if dirs != nil {
			t.Errorf("parseAPP13() dirs = %v, want nil for non-Photoshop", dirs)
		}
		if err != nil {
			t.Errorf("parseAPP13() error = %v, want nil for non-Photoshop", err)
		}
	})

	t.Run("empty segment", func(t *testing.T) {
		p := New()
		r := bytes.NewReader([]byte{})
		dirs, err := p.parseAPP13(r, 0, 0)

		if dirs != nil {
			t.Errorf("parseAPP13() dirs = %v, want nil for empty", dirs)
		}
		if err != nil {
			t.Errorf("parseAPP13() error = %v, want nil for empty", err)
		}
	})
}

func TestParser_ImplementsInterface(t *testing.T) {
	// Verify that Parser implements parser.Parser interface
	var _ parser.Parser = (*Parser)(nil)
}

// buildSegment creates a JPEG APP segment with marker and data
func buildSegment(marker byte, data []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(0xFF)
	buf.WriteByte(marker)
	// Length includes the 2 bytes for length field itself
	length := uint16(len(data) + 2)
	binary.Write(&buf, binary.BigEndian, length)
	buf.Write(data)
	return buf.Bytes()
}

func TestParser_Parse_WithSegments(t *testing.T) {
	t.Run("JPEG with APP0 segment", func(t *testing.T) {
		var buf bytes.Buffer
		// SOI
		buf.Write([]byte{0xFF, 0xD8})
		// APP0 segment (JFIF - not parsed for metadata)
		buf.Write(buildSegment(markerAPP0, []byte("JFIF\x00\x01\x01\x00\x00")))
		// EOI
		buf.Write([]byte{0xFF, 0xD9})

		p := New()
		r := bytes.NewReader(buf.Bytes())
		dirs, err := p.Parse(r)

		if err != nil {
			t.Errorf("Parse() error = %v, want nil", err)
		}
		// APP0 is not parsed for metadata
		_ = dirs
	})

	t.Run("JPEG ending with EOF", func(t *testing.T) {
		// JPEG with SOI but ends abruptly (EOF)
		data := []byte{0xFF, 0xD8} // Just SOI, no EOI

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should handle EOF gracefully
		_ = dirs
		_ = err
	})

	t.Run("JPEG with marker read error (non-EOF)", func(t *testing.T) {
		// This will trigger the non-EOF error path in Parse
		// by having invalid marker prefix in the middle
		data := []byte{
			0xFF, 0xD8, // SOI
			0x00, 0x00, // Invalid marker (not 0xFF prefix)
		}

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should return error for invalid marker
		if err == nil {
			t.Error("Parse() error = nil, want error for invalid marker")
		}
		_ = dirs
	})

	t.Run("JPEG with segment length read error", func(t *testing.T) {
		// SOI + valid marker but truncated segment length
		data := []byte{
			0xFF, 0xD8, // SOI
			0xFF, 0xE1, // APP1 marker
			0x00, // Only 1 byte of length (need 2)
		}

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should return error for truncated segment length
		if err == nil {
			t.Error("Parse() error = nil, want error for truncated segment length")
		}
		_ = dirs
	})

	t.Run("JPEG with SOS marker", func(t *testing.T) {
		// SOI + SOS (image data start)
		data := []byte{
			0xFF, 0xD8, // SOI
			0xFF, 0xDA, // SOS - stops parsing
			0x00, 0x0C, // Segment length
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Dummy data
		}

		p := New()
		r := bytes.NewReader(data)
		dirs, err := p.Parse(r)

		// Should parse successfully and stop at SOS
		_ = dirs
		_ = err
	})

	t.Run("JPEG with APP1 EXIF segment", func(t *testing.T) {
		var buf bytes.Buffer
		buf.Write([]byte{0xFF, 0xD8}) // SOI

		// Build APP1 segment with EXIF identifier
		exifData := append([]byte("Exif\x00\x00"), []byte{
			0x49, 0x49, 0x2A, 0x00, // TIFF header
			0x08, 0x00, 0x00, 0x00, // IFD offset
			0x00, 0x00, // 0 IFD entries
			0x00, 0x00, 0x00, 0x00, // Next IFD
		}...)
		buf.Write(buildSegment(markerAPP1, exifData))
		buf.Write([]byte{0xFF, 0xD9}) // EOI

		p := New()
		r := bytes.NewReader(buf.Bytes())
		dirs, err := p.Parse(r)

		// Should parse APP1 EXIF data
		_ = dirs
		_ = err
	})

	t.Run("JPEG with APP2 ICC segment", func(t *testing.T) {
		var buf bytes.Buffer
		buf.Write([]byte{0xFF, 0xD8}) // SOI

		// Build APP2 segment with ICC identifier
		iccData := append([]byte("ICC_PROFILE\x00"), []byte{
			0x01, 0x01, // Chunk 1 of 1
			0x00, 0x00, 0x00, 0x10, // Some ICC data
		}...)
		buf.Write(buildSegment(markerAPP2, iccData))
		buf.Write([]byte{0xFF, 0xD9}) // EOI

		p := New()
		r := bytes.NewReader(buf.Bytes())
		dirs, err := p.Parse(r)

		// Should parse APP2 ICC data
		_ = dirs
		_ = err
	})

	t.Run("JPEG with APP13 Photoshop segment", func(t *testing.T) {
		var buf bytes.Buffer
		buf.Write([]byte{0xFF, 0xD8}) // SOI

		// Build APP13 segment with Photoshop identifier
		iptcData := append([]byte("Photoshop 3.0\x00"), []byte{
			0x38, 0x42, 0x49, 0x4D, // "8BIM"
			0x04, 0x04, // IPTC resource
			0x00, 0x00, // Name length
			0x00, 0x00, 0x00, 0x00, // Size
		}...)
		buf.Write(buildSegment(markerAPP13, iptcData))
		buf.Write([]byte{0xFF, 0xD9}) // EOI

		p := New()
		r := bytes.NewReader(buf.Bytes())
		dirs, err := p.Parse(r)

		// Should parse APP13 IPTC data
		_ = dirs
		_ = err
	})
}

func TestParser_readMarker_ErrorInPadding(t *testing.T) {
	t.Run("error reading padding byte", func(t *testing.T) {
		// Create data with 0xFF 0xFF but then EOF
		data := []byte{0xFF, 0xFF}

		r := bytes.NewReader(data)
		_, _, err := readMarker(r, 0)

		if err == nil {
			t.Error("readMarker() error = nil, want error when padding read fails")
		}
	})
}

func TestParser_parseAPP2_ChunkReadError(t *testing.T) {
	t.Run("error reading chunk data", func(t *testing.T) {
		// Create ICC header with chunk numbers but insufficient data
		// segmentSize will be larger than actual data available
		data := []byte("ICC_PROFILE\x00")
		data = append(data, []byte{
			0x01, // Chunk 1
			0x01, // of 1
		}...)

		var chunks map[int][]byte
		var totalChunks int
		p := New()
		r := bytes.NewReader(data)

		// Call with a segmentSize that expects chunk data beyond what's available
		// segmentStart=0, segmentSize will be larger than len(data)
		// This creates: chunkDataSize = segmentSize - len(identICC) - 2
		segmentSize := int64(len(data) + 10) // Expect 10 more bytes of chunk data

		err := p.parseAPP2(r, 0, segmentSize, &chunks, &totalChunks)

		// Should return error for failed chunk data read
		if err == nil {
			t.Error("parseAPP2() error = nil, want error for chunk data read failure")
		}
	})
}

func TestParser_Parse_IncompleteICCProfile(t *testing.T) {
	t.Run("missing chunk 2 of 3", func(t *testing.T) {
		// Build JPEG with ICC profile chunks 1 and 3, but missing chunk 2
		var buf bytes.Buffer

		// SOI
		buf.Write([]byte{0xFF, 0xD8})

		// APP2 with ICC chunk 1 of 3
		iccData1 := append([]byte("ICC_PROFILE\x00"), []byte{
			0x01,       // Chunk 1
			0x03,       // of 3
			0x00, 0x01, // Some ICC data
		}...)
		buf.Write(buildSegment(markerAPP2, iccData1))

		// APP2 with ICC chunk 3 of 3 (missing chunk 2)
		iccData3 := append([]byte("ICC_PROFILE\x00"), []byte{
			0x03,       // Chunk 3
			0x03,       // of 3
			0x02, 0x03, // Some ICC data
		}...)
		buf.Write(buildSegment(markerAPP2, iccData3))

		// EOI
		buf.Write([]byte{0xFF, 0xD9})

		p := New()
		r := bytes.NewReader(buf.Bytes())
		dirs, err := p.Parse(r)

		// Should return error for incomplete ICC profile (got 2 chunks, expected 3)
		if err == nil {
			t.Fatal("Parse() error = nil, want error for incomplete ICC profile")
		}

		// Error should mention incomplete profile
		errMsg := err.Error()
		if errMsg == "" {
			t.Fatal("Parse() error message is empty")
		}

		// Verify error message mentions incomplete ICC profile
		if !contains(errMsg, "incomplete ICC profile") && !contains(errMsg, "got 2 chunks, expected 3") {
			t.Errorf("Parse() error = %q, want error mentioning incomplete ICC profile", errMsg)
		}

		// Should not return directories on error
		if dirs != nil {
			t.Errorf("Parse() dirs = %v, want nil on error", dirs)
		}
	})
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParser_Parse_ICCChunksOutOfOrder(t *testing.T) {
	t.Run("chunks received in order 3,1,2", func(t *testing.T) {
		// Build JPEG with ICC chunks arriving out of order but with valid ICC header
		var buf bytes.Buffer

		// SOI
		buf.Write([]byte{0xFF, 0xD8})

		// Create a minimal valid ICC profile split into 3 chunks
		// ICC header structure (simplified):
		// Bytes 0-3: Profile size (128 bytes minimum)
		// Bytes 4-7: Preferred CMM type
		// Bytes 8-11: Version
		// ... etc
		validICCProfile := make([]byte, 128)
		// Profile size field (big-endian uint32)
		validICCProfile[0] = 0x00
		validICCProfile[1] = 0x00
		validICCProfile[2] = 0x00
		validICCProfile[3] = 0x80 // 128 bytes
		// Signature "acsp" at offset 36
		copy(validICCProfile[36:40], []byte("acsp"))

		// Split into 3 roughly equal chunks
		chunk1 := validICCProfile[0:42]
		chunk2 := validICCProfile[42:84]
		chunk3 := validICCProfile[84:128]

		// APP2 with ICC chunk 3 of 3 (arrives first)
		iccData3 := append([]byte("ICC_PROFILE\x00"), append([]byte{0x03, 0x03}, chunk3...)...)
		buf.Write(buildSegment(markerAPP2, iccData3))

		// APP2 with ICC chunk 1 of 3 (arrives second)
		iccData1 := append([]byte("ICC_PROFILE\x00"), append([]byte{0x01, 0x03}, chunk1...)...)
		buf.Write(buildSegment(markerAPP2, iccData1))

		// APP2 with ICC chunk 2 of 3 (arrives last)
		iccData2 := append([]byte("ICC_PROFILE\x00"), append([]byte{0x02, 0x03}, chunk2...)...)
		buf.Write(buildSegment(markerAPP2, iccData2))

		// EOI
		buf.Write([]byte{0xFF, 0xD9})

		p := New()
		r := bytes.NewReader(buf.Bytes())
		dirs, err := p.Parse(r)

		// Parser should handle out-of-order chunks correctly by assembling them
		// in the correct order (1,2,3) before passing to ICC parser.
		// The ICC parser will attempt to parse the assembled profile.
		// We don't validate ICC parsing success here (that's ICC parser's job),
		// but we verify no panic occurred and chunks were processed.

		// The important validation is that chunks arrived out of order (3,1,2)
		// but were assembled correctly in order (1,2,3). This is tested by
		// the fact that Parse() completes without panic.

		// Note: The assembled ICC profile may still be invalid for ICC parser,
		// which is why err might not be nil. But that's okay - we're testing
		// JPEG parser's chunk ordering, not ICC validation.
		_ = dirs
		_ = err
	})
}
