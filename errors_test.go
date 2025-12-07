package imx

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestPartialError_Error(t *testing.T) {
	tests := []struct {
		name        string
		err         *PartialError
		wantMsg     string
		wantContain []string
	}{
		{
			name: "format error only",
			err: &PartialError{
				FormatErr: errors.New("invalid format"),
			},
			wantMsg: "imx: format: invalid format",
		},
		{
			name: "spec errors only",
			err: &PartialError{
				SpecErrs: map[Spec]error{
					SpecEXIF: errors.New("exif parse error"),
				},
			},
			wantMsg: "imx: exif: exif parse error",
		},
		{
			name:    "empty error (neither format nor spec)",
			err:     &PartialError{},
			wantMsg: "imx: partial error",
		},
		{
			name: "multiple errors",
			err: &PartialError{
				FormatErr: errors.New("format first"),
				SpecErrs: map[Spec]error{
					SpecEXIF: errors.New("exif error"),
				},
			},
			wantContain: []string{"format: format first", "exif: exif error", "multiple errors"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if tt.wantMsg != "" {
				if got != tt.wantMsg {
					t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
				}
			}
			if tt.wantContain != nil {
				for _, substr := range tt.wantContain {
					if !strings.Contains(got, substr) {
						t.Errorf("Error() = %q, want to contain %q", got, substr)
					}
				}
			}
		})
	}
}

func TestPartialError_Unwrap(t *testing.T) {
	formatErr := errors.New("format error")
	exifErr := errors.New("exif error")
	iptcErr := errors.New("iptc error")

	tests := []struct {
		name      string
		err       *PartialError
		wantErrs  []error
		wantCount int
	}{
		{
			name: "unwrap format error",
			err: &PartialError{
				FormatErr: formatErr,
			},
			wantErrs:  []error{formatErr},
			wantCount: 1,
		},
		{
			name: "unwrap spec error when no format error",
			err: &PartialError{
				SpecErrs: map[Spec]error{
					SpecEXIF: exifErr,
				},
			},
			wantErrs:  []error{exifErr},
			wantCount: 1,
		},
		{
			name:      "unwrap empty when empty",
			err:       &PartialError{},
			wantErrs:  []error{},
			wantCount: 0,
		},
		{
			name: "unwrap all errors",
			err: &PartialError{
				FormatErr: formatErr,
				SpecErrs: map[Spec]error{
					SpecEXIF: exifErr,
					SpecIPTC: iptcErr,
				},
			},
			wantErrs:  []error{formatErr, exifErr, iptcErr},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Unwrap()
			if len(got) != tt.wantCount {
				t.Errorf("Unwrap() returned %d errors, want %d", len(got), tt.wantCount)
			}

			// Verify all expected errors are present using errors.Is
			for _, wantErr := range tt.wantErrs {
				found := false
				for _, gotErr := range got {
					if errors.Is(gotErr, wantErr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unwrap() missing expected error: %v", wantErr)
				}
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors are defined correctly
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrUnknownFormat",
			err:  ErrUnknownFormat,
			want: "imx: unknown format",
		},
		{
			name: "ErrTruncatedData",
			err:  ErrTruncatedData,
			want: "imx: truncated data",
		},
		{
			name: "ErrUnsupportedMeta",
			err:  ErrUnsupportedMeta,
			want: "imx: unsupported metadata block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}

func TestPartialError_Wrapping(t *testing.T) {
	baseErr := errors.New("base error")
	partialErr := &PartialError{
		FormatErr: baseErr,
	}

	// Test wrapping with fmt.Errorf
	wrappedErr := fmt.Errorf("context: %w", partialErr)

	// Verify errors.Is works through wrapping
	if !errors.Is(wrappedErr, partialErr) {
		t.Error("errors.Is should find PartialError through wrapping")
	}

	// Verify errors.As works
	var pe *PartialError
	if !errors.As(wrappedErr, &pe) {
		t.Error("errors.As should extract PartialError through wrapping")
	}

	if pe == nil {
		t.Fatal("errors.As returned nil PartialError")
	}

	// Verify we can unwrap to find base error
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should find base error through multiple layers")
	}
}

func TestPartialError_MultipleWrapping(t *testing.T) {
	exifErr := errors.New("exif parse error")
	iptcErr := errors.New("iptc parse error")

	partialErr := &PartialError{
		SpecErrs: map[Spec]error{
			SpecEXIF: exifErr,
			SpecIPTC: iptcErr,
		},
	}

	// Wrap multiple times
	wrapped1 := fmt.Errorf("layer 1: %w", partialErr)
	wrapped2 := fmt.Errorf("layer 2: %w", wrapped1)

	// Should be able to extract PartialError through multiple layers
	var pe *PartialError
	if !errors.As(wrapped2, &pe) {
		t.Fatal("errors.As should extract PartialError through multiple wrapping layers")
	}

	// Verify the unwrapped errors are accessible
	unwrapped := pe.Unwrap()
	if len(unwrapped) != 2 {
		t.Errorf("Unwrap() returned %d errors, want 2", len(unwrapped))
	}

	// Verify we can find the original errors
	if !errors.Is(wrapped2, exifErr) {
		t.Error("errors.Is should find exifErr through multiple layers")
	}
	if !errors.Is(wrapped2, iptcErr) {
		t.Error("errors.Is should find iptcErr through multiple layers")
	}
}
