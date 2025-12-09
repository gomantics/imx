package processor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gomantics/imx"
	"github.com/gomantics/imx/cmd/imx/filter"
)

func TestProcessor_ProcessSingle(t *testing.T) {
	// Find a test image
	testImage := findTestImage(t)

	p := New(&Config{
		Workers: 1,
	})

	ctx := context.Background()
	result, err := p.ProcessSingle(ctx, testImage)

	if err != nil {
		t.Fatalf("ProcessSingle failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.File != testImage {
		t.Errorf("Expected file %s, got %s", testImage, result.File)
	}

	if result.Meta == nil {
		t.Error("Expected metadata, got nil")
	}

	if result.TagCount == 0 {
		t.Error("Expected tags, got 0")
	}
}

func TestProcessor_ProcessSingle_WithFilter(t *testing.T) {
	testImage := findTestImage(t)

	// Filter for only EXIF tags
	p := New(&Config{
		Workers: 1,
		Filter:  filter.NewSpecFilter("exif"),
	})

	ctx := context.Background()
	result, err := p.ProcessSingle(ctx, testImage)

	if err != nil {
		t.Fatalf("ProcessSingle failed: %v", err)
	}

	// Verify all tags are EXIF
	for _, tag := range result.Tags {
		if tag.Dir.Spec != imx.SpecEXIF {
			t.Errorf("Expected only EXIF tags, got %s", tag.Dir.Spec)
		}
	}
}

func TestProcessor_Process_Multiple(t *testing.T) {
	testImages := findMultipleTestImages(t, 3)

	p := New(&Config{
		Workers: 2,
	})

	ctx := context.Background()
	results, err := p.Process(ctx, testImages)

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(results) != len(testImages) {
		t.Errorf("Expected %d results, got %d", len(testImages), len(results))
	}

	for i, result := range results {
		if result.File != testImages[i] {
			t.Errorf("Result %d: expected file %s, got %s", i, testImages[i], result.File)
		}
		if result.Meta == nil {
			if result.Error != nil {
				t.Errorf("Result %d: expected metadata, got error: %v", i, result.Error)
			} else {
				t.Errorf("Result %d: expected metadata, got nil (no error)", i)
			}
		}
	}
}

func TestProcessor_Process_Empty(t *testing.T) {
	p := New(&Config{})

	results, err := p.Process(context.Background(), []string{})

	if err != nil {
		t.Errorf("Expected no error for empty input, got %v", err)
	}

	if results != nil {
		t.Errorf("Expected nil results for empty input, got %v", results)
	}
}

func TestProcessor_Process_NonexistentFile(t *testing.T) {
	p := New(&Config{})

	results, err := p.Process(context.Background(), []string{"/nonexistent/file.jpg"})

	if err != nil {
		t.Errorf("Expected no error (continue on error), got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error == nil {
		t.Error("Expected error in result, got nil")
	}
}

func TestProcessor_ProcessSingle_URL(t *testing.T) {
	// Create test server
	testImage := findTestImage(t)
	data, err := os.ReadFile(testImage)
	if err != nil {
		t.Fatalf("Failed to read test image: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(data)
	}))
	defer server.Close()

	p := New(&Config{})

	result, err := p.ProcessSingle(context.Background(), server.URL)

	if err != nil {
		t.Fatalf("ProcessSingle URL failed: %v", err)
	}

	if result.Meta == nil {
		t.Error("Expected metadata from URL, got nil")
	}
}

func TestProcessor_Context_Cancellation(t *testing.T) {
	testImages := findMultipleTestImages(t, 10)

	p := New(&Config{
		Workers: 1, // Use single worker to make cancellation more predictable
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	results, err := p.Process(ctx, testImages)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	if results != nil {
		t.Error("Expected nil results after cancellation")
	}
}

func TestProcessor_DefaultWorkers(t *testing.T) {
	p := New(&Config{
		Workers: 0, // Should default to runtime.NumCPU()
	})

	if p.config.Workers <= 0 {
		t.Error("Expected positive worker count")
	}
}

// Helper functions

func findTestImage(t *testing.T) string {
	t.Helper()

	// Look for test images in root testdata directory
	candidates := []string{
		"../../../testdata/DSC_1631.jpg",
		"../../../testdata/jpeg/DSC_1631.jpg",
		"../../../testdata/RicohWG-6.jpg",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	t.Skip("No test image found")
	return ""
}

func findMultipleTestImages(t *testing.T, count int) []string {
	t.Helper()

	var images []string

	// Look in root testdata directory
	testdataDir := "../../../testdata"
	err := filepath.Walk(testdataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && isImageFile(path) {
			images = append(images, path)
			if len(images) >= count {
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil {
		t.Logf("Error walking testdata: %v", err)
	}

	if len(images) == 0 {
		t.Skip("No test images found")
	}

	// Return up to count images
	if len(images) > count {
		images = images[:count]
	}

	return images
}

func isImageFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		return true
	}
	return false
}
