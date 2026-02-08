package tiff

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/gomantics/imx/internal/parser/tiff/makernote"
)

// TestMakerNoteRegistryDetection tests the registry detection for all supported manufacturers
func TestMakerNoteRegistryDetection(t *testing.T) {
	parser := New()

	tests := []struct {
		name         string
		data         []byte
		wantDetected bool
		wantManuf    string
	}{
		{
			name: "Nikon Type 3 little-endian",
			data: append(
				[]byte("Nikon\x00\x02\x00\x00\x00"),
				[]byte("II\x2a\x00\x08\x00\x00\x00")...,
			),
			wantDetected: true,
			wantManuf:    "Nikon",
		},
		{
			name: "Nikon Type 3 big-endian",
			data: append(
				[]byte("Nikon\x00\x02\x00\x00\x00"),
				[]byte("MM\x00\x2a\x00\x00\x00\x08")...,
			),
			wantDetected: true,
			wantManuf:    "Nikon",
		},
		{
			name:         "Nikon Type 1",
			data:         []byte("Nikon\x00\x01\x00\x00\x00\x00\x00"),
			wantDetected: true,
			wantManuf:    "Nikon",
		},
		{
			name:         "Sony DSC",
			data:         []byte("SONY DSC \x00\x00\x00"),
			wantDetected: true,
			wantManuf:    "Sony",
		},
		{
			name:         "Sony CAM",
			data:         []byte("SONY CAM \x00\x00\x00"),
			wantDetected: true,
			wantManuf:    "Sony",
		},
		{
			name:         "Fujifilm",
			data:         []byte("FUJIFILM\x0c\x00\x00\x00"),
			wantDetected: true,
			wantManuf:    "Fujifilm",
		},
		{
			name:         "Canon (IFD format)",
			data:         makeCanonIFD(5),
			wantDetected: true,
			wantManuf:    "Canon",
		},
		{
			name:         "Unknown manufacturer",
			data:         []byte("UNKNOWN HEADER DATA HERE"),
			wantDetected: false,
			wantManuf:    "",
		},
		{
			name:         "Empty data",
			data:         []byte{},
			wantDetected: false,
			wantManuf:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, cfg := parser.makernote.Detect(tt.data)

			if tt.wantDetected {
				if handler == nil {
					t.Errorf("Expected handler for %s, got nil", tt.wantManuf)
					return
				}
				if cfg == nil {
					t.Errorf("Expected config for %s, got nil", tt.wantManuf)
					return
				}
				if handler.Manufacturer() != tt.wantManuf {
					t.Errorf("Manufacturer = %s, want %s", handler.Manufacturer(), tt.wantManuf)
				}
			} else {
				if handler != nil {
					t.Errorf("Expected nil handler, got %s", handler.Manufacturer())
				}
			}
		})
	}
}

// TestMakerNoteDetectionOrder verifies handlers are checked in correct priority order
func TestMakerNoteDetectionOrder(t *testing.T) {
	parser := New()

	t.Run("Nikon takes priority over Canon fallback", func(t *testing.T) {
		data := append(
			[]byte("Nikon\x00\x02\x00\x00\x00"),
			[]byte("II\x2a\x00\x08\x00\x00\x00")...,
		)
		handler, _ := parser.makernote.Detect(data)
		if handler == nil {
			t.Fatal("Expected handler, got nil")
		}
		if handler.Manufacturer() != "Nikon" {
			t.Errorf("Expected Nikon, got %s", handler.Manufacturer())
		}
	})

	t.Run("Sony takes priority over Canon fallback", func(t *testing.T) {
		data := []byte("SONY DSC \x00\x00\x00")
		handler, _ := parser.makernote.Detect(data)
		if handler == nil {
			t.Fatal("Expected handler, got nil")
		}
		if handler.Manufacturer() != "Sony" {
			t.Errorf("Expected Sony, got %s", handler.Manufacturer())
		}
	})

	t.Run("Fujifilm takes priority over Canon fallback", func(t *testing.T) {
		data := []byte("FUJIFILM\x0c\x00\x00\x00")
		handler, _ := parser.makernote.Detect(data)
		if handler == nil {
			t.Fatal("Expected handler, got nil")
		}
		if handler.Manufacturer() != "Fujifilm" {
			t.Errorf("Expected Fujifilm, got %s", handler.Manufacturer())
		}
	})
}

// TestMakerNoteMalformedHeaders tests graceful handling of malformed data
func TestMakerNoteMalformedHeaders(t *testing.T) {
	parser := New()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"single byte", []byte{0x00}},
		{"two bytes", []byte{0x00, 0x00}},
		{"truncated Nikon", []byte("Nikon")},
		{"truncated Sony", []byte("SONY DSC")},
		{"truncated Fuji", []byte("FUJIFILM")},
		{"invalid UTF-8 sequence", []byte{0xff, 0xfe, 0xfd, 0xfc}},
		{"all zeros", make([]byte, 100)},
		{"all 0xFF", func() []byte {
			b := make([]byte, 100)
			for i := range b {
				b[i] = 0xff
			}
			return b
		}()},
		{"random garbage", []byte{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe, 0xba, 0xbe}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			handler, cfg := parser.makernote.Detect(tt.data)
			_ = handler
			_ = cfg
		})
	}
}

// TestMakerNoteBoundaryConditions tests edge cases in detection
func TestMakerNoteBoundaryConditions(t *testing.T) {
	t.Run("Canon max entries", func(t *testing.T) {
		data := makeCanonIFD(100)
		match, cfg := makernote.DetectCanon(data)
		if !match || cfg == nil {
			t.Error("100 entries should be valid")
		}
	})

	t.Run("Canon too many entries", func(t *testing.T) {
		data := make([]byte, 2)
		binary.LittleEndian.PutUint16(data, 101)
		match, _ := makernote.DetectCanon(data)
		if match {
			t.Error("101 entries should be invalid")
		}
	})

	t.Run("Fujifilm minimum valid offset", func(t *testing.T) {
		data := []byte("FUJIFILM\x0c\x00\x00\x00")
		match, cfg := makernote.DetectFujifilm(data)
		if !match {
			t.Error("Offset 12 should be valid")
		}
		if cfg.IFDOffset != 12 {
			t.Errorf("IFDOffset = %d, want 12", cfg.IFDOffset)
		}
	})
}

// TestMakerNoteConcurrentDetection tests thread safety of registry
func TestMakerNoteConcurrentDetection(t *testing.T) {
	parser := New()

	testData := [][]byte{
		append([]byte("Nikon\x00\x02\x00\x00\x00"), []byte("II\x2a\x00\x08\x00\x00\x00")...),
		[]byte("SONY DSC \x00\x00\x00"),
		[]byte("FUJIFILM\x0c\x00\x00\x00"),
		makeCanonIFD(10),
		[]byte("UNKNOWN DATA"),
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				data := testData[j%len(testData)]
				handler, cfg := parser.makernote.Detect(data)
				_ = handler
				_ = cfg
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestMakerNoteRealFile tests parsing with an actual camera file
func TestMakerNoteRealFile(t *testing.T) {
	testFiles := []struct {
		path     string
		wantDir  string
		minTags  int
	}{
		{"testdata/cr2/sample1.cr2", "Canon", 1},
	}

	for _, tf := range testFiles {
		t.Run(tf.path, func(t *testing.T) {
			if _, err := os.Stat(tf.path); os.IsNotExist(err) {
				t.Skipf("Test file not available: %s", tf.path)
			}

			f, err := os.Open(tf.path)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer f.Close()

			parser := New()
			if !parser.Detect(f) {
				t.Fatal("File not detected as TIFF/CR2")
			}

			dirs, parseErr := parser.Parse(f)
			if parseErr != nil {
				t.Logf("Parse warnings: %v", parseErr)
			}

			// Check for manufacturer directory
			var foundManufDir bool
			for _, dir := range dirs {
				if dir.Name == tf.wantDir {
					foundManufDir = true
					if len(dir.Tags) < tf.minTags {
						t.Errorf("%s directory has %d tags, want at least %d", tf.wantDir, len(dir.Tags), tf.minTags)
					}
					break
				}
			}

			if !foundManufDir {
				t.Errorf("Expected %s directory, not found in: %v", tf.wantDir, func() []string {
					names := make([]string, len(dirs))
					for i, d := range dirs {
						names[i] = d.Name
					}
					return names
				}())
			}
		})
	}
}

// Helper function to create a valid Canon IFD
func makeCanonIFD(entryCount uint16) []byte {
	size := 2 + int(entryCount)*12
	data := make([]byte, size)
	binary.LittleEndian.PutUint16(data[0:2], entryCount)
	// Fill in valid IFD entries (Canon handler validates tag types 1-12)
	for i := uint16(0); i < entryCount; i++ {
		offset := 2 + int(i)*12
		binary.LittleEndian.PutUint16(data[offset:offset+2], 0x0001+i)   // Tag ID
		binary.LittleEndian.PutUint16(data[offset+2:offset+4], 3)        // Type: SHORT (3)
		binary.LittleEndian.PutUint32(data[offset+4:offset+8], 1)        // Count
		binary.LittleEndian.PutUint32(data[offset+8:offset+12], uint32(i)) // Value
	}
	return data
}

// Benchmark tests
func BenchmarkMakerNoteDetect_Nikon(b *testing.B) {
	parser := New()
	data := append(
		[]byte("Nikon\x00\x02\x00\x00\x00"),
		[]byte("II\x2a\x00\x08\x00\x00\x00")...,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.makernote.Detect(data)
	}
}

func BenchmarkMakerNoteDetect_Sony(b *testing.B) {
	parser := New()
	data := []byte("SONY DSC \x00\x00\x00")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.makernote.Detect(data)
	}
}

func BenchmarkMakerNoteDetect_Fujifilm(b *testing.B) {
	parser := New()
	data := []byte("FUJIFILM\x0c\x00\x00\x00")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.makernote.Detect(data)
	}
}

func BenchmarkMakerNoteDetect_Canon(b *testing.B) {
	parser := New()
	data := makeCanonIFD(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.makernote.Detect(data)
	}
}

func BenchmarkMakerNoteDetect_Unknown(b *testing.B) {
	parser := New()
	data := []byte("UNKNOWN MANUFACTURER DATA")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.makernote.Detect(data)
	}
}
