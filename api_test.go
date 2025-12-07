package imx

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// testJPEGPathAPI is the path to the test JPEG file
const testJPEGPathAPI = "testdata/goldens/jpeg/canon_xmp.jpg"

// loadTestJPEGAPI loads the test JPEG file for API testing
func loadTestJPEGAPI(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(testJPEGPathAPI)
	if err != nil {
		t.Fatalf("Failed to load test JPEG: %v", err)
	}
	return data
}

func TestMetadataFromReader(t *testing.T) {
	validJPEG := loadTestJPEGAPI(t)

	tests := []struct {
		name    string
		data    []byte
		opts    []Option
		wantErr bool
	}{
		{
			name:    "valid JPEG",
			data:    validJPEG,
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid JPEG with options",
			data:    validJPEG,
			opts:    []Option{WithMaxBytes(20000000)}, // Large enough for the test file
			wantErr: false,
		},
		{
			name:    "invalid data",
			data:    []byte{0x00, 0x01, 0x02, 0x03},
			opts:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			_, err := MetadataFromReader(r, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("MetadataFromReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetadataFromFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "valid file",
			path:    testJPEGPathAPI,
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid file with options",
			path:    testJPEGPathAPI,
			opts:    []Option{WithMaxBytes(20000000)},
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    "testdata/nonexistent.jpg",
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "directory instead of file",
			path:    "testdata",
			opts:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MetadataFromFile(tt.path, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("MetadataFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetadataFromBytes(t *testing.T) {
	validJPEG := loadTestJPEGAPI(t)

	tests := []struct {
		name    string
		data    []byte
		opts    []Option
		wantErr bool
	}{
		{
			name:    "valid JPEG bytes",
			data:    validJPEG,
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid JPEG with options",
			data:    validJPEG,
			opts:    []Option{WithSpecs(SpecEXIF)},
			wantErr: false,
		},
		{
			name:    "empty bytes",
			data:    []byte{},
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "invalid bytes",
			data:    []byte{0x00, 0x01, 0x02},
			opts:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MetadataFromBytes(tt.data, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("MetadataFromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetadataFromURL(t *testing.T) {
	validJPEG := loadTestJPEGAPI(t)

	// Create test server
	mux := http.NewServeMux()

	mux.HandleFunc("/valid.jpg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(validJPEG)
	})

	mux.HandleFunc("/invalid.jpg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{0x00, 0x01, 0x02, 0x03})
	})

	mux.HandleFunc("/notfound", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/servererror", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "valid JPEG URL",
			url:     server.URL + "/valid.jpg",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid URL with options",
			url:     server.URL + "/valid.jpg",
			opts:    []Option{WithMaxBytes(20000000)},
			wantErr: false,
		},
		{
			name:    "invalid JPEG data",
			url:     server.URL + "/invalid.jpg",
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "404 not found",
			url:     server.URL + "/notfound",
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "500 server error",
			url:     server.URL + "/servererror",
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			url:     "://invalid-url",
			opts:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MetadataFromURL(tt.url, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("MetadataFromURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractor_MetadataFromFile(t *testing.T) {
	e := New()

	t.Run("valid file", func(t *testing.T) {
		_, err := e.MetadataFromFile(testJPEGPathAPI)
		if err != nil {
			t.Errorf("MetadataFromFile() error = %v", err)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := e.MetadataFromFile("/nonexistent/path/image.jpg")
		if err == nil {
			t.Error("MetadataFromFile() expected error for non-existent file")
		}
	})
}

func TestExtractor_MetadataFromBytes(t *testing.T) {
	e := New()
	validJPEG := loadTestJPEGAPI(t)

	t.Run("valid bytes", func(t *testing.T) {
		_, err := e.MetadataFromBytes(validJPEG)
		if err != nil {
			t.Errorf("MetadataFromBytes() error = %v", err)
		}
	})

	t.Run("invalid bytes", func(t *testing.T) {
		_, err := e.MetadataFromBytes([]byte{0x00, 0x01})
		if err == nil {
			t.Error("MetadataFromBytes() expected error for invalid data")
		}
	})
}

func TestExtractor_MetadataFromURL(t *testing.T) {
	e := New()
	validJPEG := loadTestJPEGAPI(t)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid.jpg":
			w.WriteHeader(http.StatusOK)
			w.Write(validJPEG)
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("valid URL", func(t *testing.T) {
		_, err := e.MetadataFromURL(server.URL + "/valid.jpg")
		if err != nil {
			t.Errorf("MetadataFromURL() error = %v", err)
		}
	})

	t.Run("non-200 status", func(t *testing.T) {
		_, err := e.MetadataFromURL(server.URL + "/error")
		if err == nil {
			t.Error("MetadataFromURL() expected error for non-200 status")
		}
	})

	t.Run("invalid URL format", func(t *testing.T) {
		_, err := e.MetadataFromURL("://invalid-url")
		if err == nil {
			t.Error("MetadataFromURL() expected error for invalid URL format")
		}
	})
}

// TestDefaultExtractor verifies the default extractor is initialized
func TestDefaultExtractor(t *testing.T) {
	if defaultExtractor == nil {
		t.Error("defaultExtractor should not be nil")
	}
}

// TestMetadataContent verifies that real metadata is extracted from the test file
func TestMetadataContent(t *testing.T) {
	// Use the real test file to validate actual metadata extraction
	metadata, err := MetadataFromFile(testJPEGPathAPI)
	if err != nil {
		t.Fatalf("MetadataFromFile() error = %v", err)
	}

	// Verify we got some directories
	if len(metadata.Directories) == 0 {
		t.Error("Expected at least one directory from real JPEG file")
	}

	// Verify we can extract common EXIF tags
	// Test that Tag() method works and returns valid data
	if tag, ok := metadata.Tag(SpecEXIF, TagMake); ok {
		if make, ok := tag.Value.(string); ok && make != "" {
			t.Logf("Successfully extracted Camera Make: %s", make)
		} else {
			t.Errorf("Make tag value = %v (type %T), want non-empty string", tag.Value, tag.Value)
		}
	}

	if tag, ok := metadata.Tag(SpecEXIF, TagModel); ok {
		if model, ok := tag.Value.(string); ok && model != "" {
			t.Logf("Successfully extracted Camera Model: %s", model)
		} else {
			t.Errorf("Model tag value = %v (type %T), want non-empty string", tag.Value, tag.Value)
		}
	}

	// At minimum, we should have extracted SOME tags from this real file
	totalTags := 0
	for _, dir := range metadata.Directories {
		totalTags += len(dir.Tags)
	}
	if totalTags == 0 {
		t.Error("Expected to extract at least some tags from real JPEG file")
	}
}
