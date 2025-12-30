package gif

import (
	"bytes"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name                string
		data                []byte
		wantErr             bool
		wantVersion         string
		wantWidth           int
		wantHeight          int
		wantHasGCT          bool
		wantColorResolution int
		wantBitsPerPixel    int
		wantBgColor         uint8
	}{
		{
			name:                "valid GIF89a with GCT",
			data:                []byte("GIF89a\x0A\x00\x0A\x00\xF7\x00\x00"),
			wantVersion:         "GIF89a",
			wantWidth:           10,
			wantHeight:          10,
			wantHasGCT:          true,
			wantColorResolution: 8,
			wantBitsPerPixel:    8,
			wantBgColor:         0,
		},
		{
			name:                "valid GIF87a without GCT",
			data:                []byte("GIF87a\x14\x00\x1E\x00\x00\x01\x00"),
			wantVersion:         "GIF87a",
			wantWidth:           20,
			wantHeight:          30,
			wantHasGCT:          false,
			wantColorResolution: 1,
			wantBitsPerPixel:    1,
			wantBgColor:         1,
		},
		{
			name:    "invalid signature",
			data:    []byte("PNG89a\x00\x00\x00\x00\x00\x00\x00"),
			wantErr: true,
		},
		{
			name:    "truncated header",
			data:    []byte("GIF89a"),
			wantErr: true,
		},
		{
			name:    "truncated LSD",
			data:    []byte("GIF89a\x00\x00\x00"),
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
			var buf [11]byte

			version, pos, width, height, gifDir, err := parseHeader(r, &buf)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseHeader() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseHeader() unexpected error: %v", err)
			}

			if version != tt.wantVersion {
				t.Errorf("version = %q, want %q", version, tt.wantVersion)
			}

			if width != tt.wantWidth {
				t.Errorf("width = %d, want %d", width, tt.wantWidth)
			}

			if height != tt.wantHeight {
				t.Errorf("height = %d, want %d", height, tt.wantHeight)
			}

			if gifDir == nil {
				t.Fatal("gifDir is nil")
			}

			// Verify position is correct
			expectedPos := int64(13)
			if tt.wantHasGCT {
				gctSize := 1 << ((tt.wantBitsPerPixel - 1) + 1)
				expectedPos += int64(gctSize * 3)
			}
			if pos != expectedPos {
				t.Errorf("pos = %d, want %d", pos, expectedPos)
			}

			// Verify tags
			hasColorMap := findTag(gifDir.Tags, "HasColorMap")
			if hasColorMap == nil {
				t.Fatal("HasColorMap tag not found")
			}
			if hasColorMap.Value != tt.wantHasGCT {
				t.Errorf("HasColorMap = %v, want %v", hasColorMap.Value, tt.wantHasGCT)
			}
		})
	}
}

func TestParseHeader_PixelAspectRatio(t *testing.T) {
	// Test with non-zero pixel aspect ratio
	data := []byte("GIF89a\x0A\x00\x0A\x00\x00\x00\x40") // Pixel aspect ratio = 0x40
	r := bytes.NewReader(data)
	var buf [11]byte

	_, _, _, _, gifDir, err := parseHeader(r, &buf)
	if err != nil {
		t.Fatalf("parseHeader() error: %v", err)
	}

	// Should have PixelAspectRatio tag
	parTag := findTag(gifDir.Tags, "PixelAspectRatio")
	if parTag == nil {
		t.Error("PixelAspectRatio tag not found")
	} else if parTag.Value != uint8(0x40) {
		t.Errorf("PixelAspectRatio = %v, want 64", parTag.Value)
	}
}

func TestParseHeader_GlobalColorTableSorted(t *testing.T) {
	// Test with sorted flag set
	data := []byte("GIF89a\x0A\x00\x0A\x00\xF8\x00\x00") // Packed field with sort flag
	r := bytes.NewReader(data)
	var buf [11]byte

	_, _, _, _, gifDir, err := parseHeader(r, &buf)
	if err != nil {
		t.Fatalf("parseHeader() error: %v", err)
	}

	// Should have GlobalColorTableSorted tag
	sortTag := findTag(gifDir.Tags, "GlobalColorTableSorted")
	if sortTag == nil {
		t.Error("GlobalColorTableSorted tag not found")
	} else if sortTag.Value != true {
		t.Errorf("GlobalColorTableSorted = %v, want true", sortTag.Value)
	}
}

// Helper function to find a tag by name
func findTag(tags []parser.Tag, name string) *parser.Tag {
	for i := range tags {
		if tags[i].Name == name {
			return &tags[i]
		}
	}
	return nil
}
