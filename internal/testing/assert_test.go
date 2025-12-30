package testing

import (
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

// TestAssertTag tests single tag validation
func TestAssertTag(t *testing.T) {
	tests := []struct {
		name      string
		got       parser.Tag
		want      TagExpectation
		wantError bool
	}{
		{
			name:      "exact value match",
			got:       parser.Tag{Name: "Make", Value: "Canon"},
			want:      TagExpectation{Name: "Make", Value: "Canon"},
			wantError: false,
		},
		{
			name:      "exact value mismatch",
			got:       parser.Tag{Name: "Make", Value: "Nikon"},
			want:      TagExpectation{Name: "Make", Value: "Canon"},
			wantError: true,
		},
		{
			name:      "type match",
			got:       parser.Tag{Name: "ISO", Value: uint16(100)},
			want:      TagExpectation{Name: "ISO", Type: "uint16"},
			wantError: false,
		},
		{
			name:      "type mismatch",
			got:       parser.Tag{Name: "ISO", Value: "100"},
			want:      TagExpectation{Name: "ISO", Type: "uint16"},
			wantError: true,
		},
		{
			name:      "pattern match",
			got:       parser.Tag{Name: "DateTime", Value: "2024:01:15 10:30:45"},
			want:      TagExpectation{Name: "DateTime", Pattern: `^\d{4}:\d{2}:\d{2}`},
			wantError: false,
		},
		{
			name:      "pattern mismatch",
			got:       parser.Tag{Name: "DateTime", Value: "invalid"},
			want:      TagExpectation{Name: "DateTime", Pattern: `^\d{4}:\d{2}:\d{2}`},
			wantError: true,
		},
		{
			name:      "pattern on non-string",
			got:       parser.Tag{Name: "Value", Value: 123},
			want:      TagExpectation{Name: "Value", Pattern: `\d+`},
			wantError: true,
		},
		{
			name:      "presence only",
			got:       parser.Tag{Name: "SomeTag", Value: "anything"},
			want:      TagExpectation{Name: "SomeTag"},
			wantError: false,
		},
		{
			name:      "invalid regex pattern",
			got:       parser.Tag{Name: "Value", Value: "test"},
			want:      TagExpectation{Name: "Value", Pattern: "[invalid("},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertTag(tt.got, tt.want)
			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestAssertTags tests multiple tag validation
func TestAssertTags(t *testing.T) {
	tests := []struct {
		name      string
		gotTags   []parser.Tag
		wantTags  []TagExpectation
		wantError bool
	}{
		{
			name: "all tags match",
			gotTags: []parser.Tag{
				{Name: "Make", Value: "Canon"},
				{Name: "Model", Value: "EOS 5D"},
			},
			wantTags: []TagExpectation{
				{Name: "Make", Value: "Canon"},
				{Name: "Model", Value: "EOS 5D"},
			},
			wantError: false,
		},
		{
			name: "missing expected tag",
			gotTags: []parser.Tag{
				{Name: "Make", Value: "Canon"},
			},
			wantTags: []TagExpectation{
				{Name: "Make", Value: "Canon"},
				{Name: "Model", Value: "EOS 5D"},
			},
			wantError: true,
		},
		{
			name: "unexpected tag",
			gotTags: []parser.Tag{
				{Name: "Make", Value: "Canon"},
				{Name: "Extra", Value: "Unexpected"},
			},
			wantTags: []TagExpectation{
				{Name: "Make", Value: "Canon"},
			},
			wantError: true,
		},
		{
			name: "tag value mismatch",
			gotTags: []parser.Tag{
				{Name: "Make", Value: "Nikon"},
			},
			wantTags: []TagExpectation{
				{Name: "Make", Value: "Canon"},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AssertTags(tt.gotTags, tt.wantTags)
			if tt.wantError && !result.Failed() {
				t.Error("expected validation to fail but it passed")
			}
			if !tt.wantError && result.Failed() {
				t.Errorf("expected validation to pass but it failed: %v", result.Errors)
			}
		})
	}
}

// TestAssertDirectories tests directory validation
func TestAssertDirectories(t *testing.T) {
	tests := []struct {
		name      string
		dirs      []parser.Directory
		wantDirs  []DirectoryExpectation
		wantError bool
	}{
		{
			name: "exact match",
			dirs: []parser.Directory{
				{
					Name: "IFD0",
					Tags: []parser.Tag{
						{Name: "Make", Value: "Canon"},
						{Name: "Model", Value: "EOS 5D"},
					},
				},
			},
			wantDirs: []DirectoryExpectation{
				{
					Name:          "IFD0",
					ExactTagCount: 2,
					Tags: []TagExpectation{
						{Name: "Make", Value: "Canon"},
						{Name: "Model", Value: "EOS 5D"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "empty directories",
			dirs: []parser.Directory{},
			wantDirs: []DirectoryExpectation{
				{Name: "IFD0", ExactTagCount: 1, Tags: []TagExpectation{{Name: "Make"}}},
			},
			wantError: true,
		},
		{
			name: "missing directory",
			dirs: []parser.Directory{
				{Name: "IFD0", Tags: []parser.Tag{{Name: "Make", Value: "Canon"}}},
			},
			wantDirs: []DirectoryExpectation{
				{Name: "IFD0", ExactTagCount: 1, Tags: []TagExpectation{{Name: "Make"}}},
				{Name: "ExifIFD", ExactTagCount: 1, Tags: []TagExpectation{{Name: "ISO"}}},
			},
			wantError: true,
		},
		{
			name: "unexpected directory",
			dirs: []parser.Directory{
				{Name: "IFD0", Tags: []parser.Tag{{Name: "Make", Value: "Canon"}}},
				{Name: "Extra", Tags: []parser.Tag{{Name: "Unexpected", Value: "Tag"}}},
			},
			wantDirs: []DirectoryExpectation{
				{Name: "IFD0", ExactTagCount: 1, Tags: []TagExpectation{{Name: "Make"}}},
			},
			wantError: true,
		},
		{
			name: "wrong tag count",
			dirs: []parser.Directory{
				{
					Name: "IFD0",
					Tags: []parser.Tag{
						{Name: "Make", Value: "Canon"},
						{Name: "Model", Value: "EOS 5D"},
						{Name: "Extra", Value: "Tag"},
					},
				},
			},
			wantDirs: []DirectoryExpectation{
				{
					Name:          "IFD0",
					ExactTagCount: 2,
					Tags: []TagExpectation{
						{Name: "Make", Value: "Canon"},
						{Name: "Model", Value: "EOS 5D"},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AssertDirectories(tt.dirs, tt.wantDirs)
			if tt.wantError && !result.Failed() {
				t.Error("expected validation to fail but it passed")
			}
			if !tt.wantError && result.Failed() {
				t.Errorf("expected validation to pass but it failed: %v", result.Errors)
			}
		})
	}
}

// TestValidationResult tests the ValidationResult type
func TestValidationResult(t *testing.T) {
	result := &ValidationResult{}

	if result.Failed() {
		t.Error("empty result should not fail")
	}

	result.AddError("Field1", "error 1")
	if !result.Failed() {
		t.Error("result with errors should fail")
	}

	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}

	result.AddErrorf("Field2", "error %d", 2)
	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

// TestValidationError tests the ValidationError type
func TestValidationError(t *testing.T) {
	err := ValidationError{Field: "IFD0.Make", Message: "value mismatch"}
	expected := "IFD0.Make: value mismatch"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
