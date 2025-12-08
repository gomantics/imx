package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"http://example.com/image.jpg", true},
		{"https://example.com/image.jpg", true},
		{"HTTP://example.com/image.jpg", false}, // Case sensitive
		{"/path/to/file.jpg", false},
		{"file.jpg", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsURL(tt.path)
			if got != tt.want {
				t.Errorf("IsURL(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"photo.jpg", true},
		{"photo.JPG", true},
		{"photo.jpeg", true},
		{"photo.png", true},
		{"photo.gif", true},
		{"photo.webp", true},
		{"photo.tiff", true},
		{"photo.tif", true},
		{"photo.heic", true},
		{"photo.heif", true},
		{"photo.avif", true},
		{"photo.cr2", true},
		{"photo.nef", true},
		{"photo.arw", true},
		{"photo.dng", true},
		{"document.pdf", false},
		{"script.sh", false},
		{"noextension", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsImageFile(tt.path)
			if got != tt.want {
				t.Errorf("IsImageFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestExpandFiles(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test files
	files := []string{
		"photo1.jpg",
		"photo2.png",
		"document.pdf",
		"subdir/photo3.jpg",
		"subdir/photo4.webp",
		"subdir/deep/photo5.tiff",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		paths     []string
		recursive bool
		wantCount int
		wantErr   bool
	}{
		{
			name:      "single file",
			paths:     []string{filepath.Join(tmpDir, "photo1.jpg")},
			recursive: false,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "multiple files",
			paths:     []string{filepath.Join(tmpDir, "photo1.jpg"), filepath.Join(tmpDir, "photo2.png")},
			recursive: false,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "glob pattern",
			paths:     []string{filepath.Join(tmpDir, "*.jpg")},
			recursive: false,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "directory without recursive",
			paths:     []string{tmpDir},
			recursive: false,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "directory with recursive - root only has 2 images",
			paths:     []string{tmpDir},
			recursive: true,
			wantCount: 5, // All image files in all subdirectories
			wantErr:   false,
		},
		{
			name:      "URL",
			paths:     []string{"https://example.com/photo.jpg"},
			recursive: false,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "mixed files and URLs",
			paths:     []string{filepath.Join(tmpDir, "photo1.jpg"), "https://example.com/photo.jpg"},
			recursive: false,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "nonexistent file",
			paths:     []string{filepath.Join(tmpDir, "nonexistent.jpg")},
			recursive: false,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "duplicate files",
			paths:     []string{filepath.Join(tmpDir, "photo1.jpg"), filepath.Join(tmpDir, "photo1.jpg")},
			recursive: false,
			wantCount: 1, // Should deduplicate
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandFiles(tt.paths, tt.recursive)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("ExpandFiles() returned %d files, want %d files", len(got), tt.wantCount)
				t.Logf("Files: %v", got)
			}
		})
	}
}

func TestExpandFiles_NoFiles(t *testing.T) {
	_, err := ExpandFiles([]string{}, false)
	if err == nil {
		t.Error("ExpandFiles() with empty paths should return error")
	}
}
