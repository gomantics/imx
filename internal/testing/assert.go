package testing

import (
	"fmt"
	"regexp"

	"github.com/gomantics/imx/internal/parser"
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string // Which field failed (e.g., "Directory.IFD0.Tag.Make")
	Message string // Error message
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult holds the outcome of a validation
type ValidationResult struct {
	Errors []ValidationError
}

// Failed returns true if there are any validation errors
func (r *ValidationResult) Failed() bool {
	return len(r.Errors) > 0
}

// AddError adds a validation error
func (r *ValidationResult) AddError(field, message string) {
	r.Errors = append(r.Errors, ValidationError{Field: field, Message: message})
}

// AddErrorf adds a formatted validation error
func (r *ValidationResult) AddErrorf(field, format string, args ...interface{}) {
	r.AddError(field, fmt.Sprintf(format, args...))
}

// AssertDirectories validates directories and returns validation result
func AssertDirectories(dirs []parser.Directory, expected []DirectoryExpectation) *ValidationResult {
	result := &ValidationResult{}

	if len(dirs) == 0 {
		result.AddError("Directories", "no directories found")
		return result
	}

	// Build maps
	gotDirMap := make(map[string]*parser.Directory)
	for i := range dirs {
		gotDirMap[dirs[i].Name] = &dirs[i]
	}

	wantDirMap := make(map[string]DirectoryExpectation)
	for _, wd := range expected {
		wantDirMap[wd.Name] = wd
	}

	// Check expected directories exist
	for _, wantDir := range expected {
		gotDir, found := gotDirMap[wantDir.Name]
		if !found {
			result.AddErrorf("Directory", "missing expected directory: %s", wantDir.Name)
			continue
		}

		// Check tag count
		if len(gotDir.Tags) != wantDir.ExactTagCount {
			result.AddErrorf("Directory."+wantDir.Name,
				"has %d tags, want exactly %d", len(gotDir.Tags), wantDir.ExactTagCount)
		}

		// Validate tags
		tagResult := AssertTags(gotDir.Tags, wantDir.Tags)
		for _, err := range tagResult.Errors {
			result.AddError("Directory."+wantDir.Name+"."+err.Field, err.Message)
		}
	}

	// Check for unexpected directories
	for _, gotDir := range dirs {
		if _, expected := wantDirMap[gotDir.Name]; !expected {
			result.AddErrorf("Directory",
				"unexpected directory: %s (with %d tags)", gotDir.Name, len(gotDir.Tags))
		}
	}

	return result
}

// AssertTags validates tags and returns validation result
func AssertTags(gotTags []parser.Tag, expected []TagExpectation) *ValidationResult {
	result := &ValidationResult{}

	// Build maps
	gotTagMap := make(map[string]parser.Tag)
	for _, tag := range gotTags {
		gotTagMap[tag.Name] = tag
	}

	wantTagMap := make(map[string]bool)
	for _, tag := range expected {
		wantTagMap[tag.Name] = true
	}

	// Check expected tags
	for _, want := range expected {
		got, found := gotTagMap[want.Name]
		if !found {
			result.AddErrorf("Tag", "missing expected tag: %s", want.Name)
			continue
		}

		// Validate tag value/type
		if err := AssertTag(got, want); err != nil {
			result.AddError("Tag."+want.Name, err.Message)
		}
	}

	// Check for unexpected tags
	for _, got := range gotTags {
		if !wantTagMap[got.Name] {
			result.AddErrorf("Tag",
				"unexpected tag: %s = %v (%T)", got.Name, got.Value, got.Value)
		}
	}

	return result
}

// AssertTag validates a single tag against expectation
func AssertTag(got parser.Tag, want TagExpectation) *ValidationError {
	// Check exact value
	if want.Value != nil {
		if !ValuesEqual(got.Value, want.Value) {
			return &ValidationError{
				Field: got.Name,
				Message: fmt.Sprintf("value = %v (%T), want %v (%T)",
					got.Value, got.Value, want.Value, want.Value),
			}
		}
		return nil
	}

	// Check type
	if want.Type != "" {
		if !TypeMatches(got.Value, want.Type) {
			return &ValidationError{
				Field:   got.Name,
				Message: fmt.Sprintf("type = %T, want %s", got.Value, want.Type),
			}
		}
		return nil
	}

	// Check pattern
	if want.Pattern != "" {
		str, ok := got.Value.(string)
		if !ok {
			return &ValidationError{
				Field:   got.Name,
				Message: fmt.Sprintf("value is %T, can't match pattern (need string)", got.Value),
			}
		}

		matched, err := regexp.MatchString(want.Pattern, str)
		if err != nil {
			return &ValidationError{
				Field:   got.Name,
				Message: fmt.Sprintf("invalid pattern %q: %v", want.Pattern, err),
			}
		}

		if !matched {
			return &ValidationError{
				Field:   got.Name,
				Message: fmt.Sprintf("value %q doesn't match pattern %q", str, want.Pattern),
			}
		}
		return nil
	}

	// No validation specified - just presence check (already passed)
	return nil
}
