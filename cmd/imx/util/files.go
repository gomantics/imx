package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ImageExtensions contains all supported image file extensions
var ImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".tiff": true,
	".tif":  true,
	".heic": true,
	".heif": true,
	".avif": true,
	".bmp":  true,
	// RAW formats
	".cr2": true,
	".cr3": true,
	".nef": true,
	".arw": true,
	".dng": true,
	".orf": true,
	".rw2": true,
	".pef": true,
	".srw": true,
	".raf": true,
}

// IsURL checks if the given path is an HTTP or HTTPS URL
func IsURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

// IsImageFile checks if the file has a supported image extension
func IsImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ImageExtensions[ext]
}

// ExpandFiles expands file patterns and directories into a list of image files
func ExpandFiles(paths []string, recursive bool) ([]string, error) {
	var files []string
	seen := make(map[string]bool) // Deduplicate files

	for _, path := range paths {
		// Handle URLs directly
		if IsURL(path) {
			if !seen[path] {
				files = append(files, path)
				seen[path] = true
			}
			continue
		}

		// Check if path exists
		info, err := os.Stat(path)
		if err != nil {
			// Path doesn't exist - could be a glob pattern or just invalid
			// Try glob expansion
			matches, globErr := filepath.Glob(path)
			if globErr != nil || len(matches) == 0 {
				// Not a valid glob, return original error
				return nil, NewProcessError(path, "stat", err)
			}

			// Process glob matches
			for _, match := range matches {
				matchInfo, err := os.Stat(match)
				if err != nil {
					continue
				}

				if matchInfo.IsDir() {
					if !recursive {
						return nil, fmt.Errorf("%s is a directory (use -r/--recursive to scan directories)", match)
					}
					expanded, err := expandDirectory(match, recursive)
					if err != nil {
						return nil, err
					}
					for _, f := range expanded {
						if !seen[f] {
							files = append(files, f)
							seen[f] = true
						}
					}
				} else if IsImageFile(match) {
					if !seen[match] {
						files = append(files, match)
						seen[match] = true
					}
				}
			}
			continue
		}

		// Path exists
		if info.IsDir() {
			if !recursive {
				return nil, fmt.Errorf("%s is a directory (use -r/--recursive to scan directories)", path)
			}
			expanded, err := expandDirectory(path, recursive)
			if err != nil {
				return nil, err
			}
			for _, f := range expanded {
				if !seen[f] {
					files = append(files, f)
					seen[f] = true
				}
			}
		} else {
			// Regular file
			if !seen[path] {
				files = append(files, path)
				seen[path] = true
			}
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found")
	}

	return files, nil
}

// expandDirectory recursively walks a directory and returns all image files
func expandDirectory(dir string, recursive bool) ([]string, error) {
	var files []string

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Skip files/directories we can't access
				return nil
			}

			if info.IsDir() {
				return nil
			}

			if IsImageFile(path) {
				files = append(files, path)
			}

			return nil
		})

		if err != nil {
			return nil, NewProcessError(dir, "walk", err)
		}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, NewProcessError(dir, "readdir", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			if IsImageFile(path) {
				files = append(files, path)
			}
		}
	}

	return files, nil
}
