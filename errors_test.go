package imx

import (
	"errors"
	"testing"

	"github.com/gomantics/imx/internal/meta"
)

func TestPartialError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *PartialError
		wantMsg string
	}{
		{
			name: "format error only",
			err: &PartialError{
				FormatErr: errors.New("invalid format"),
			},
			wantMsg: "imx: format error: invalid format",
		},
		{
			name: "spec errors only",
			err: &PartialError{
				SpecErrs: map[meta.Spec]error{
					meta.SpecEXIF: errors.New("exif parse error"),
				},
			},
			wantMsg: "imx: spec errors: map[exif:exif parse error]",
		},
		{
			name:    "empty error (neither format nor spec)",
			err:     &PartialError{},
			wantMsg: "imx: partial error",
		},
		{
			name: "format error takes precedence over spec errors",
			err: &PartialError{
				FormatErr: errors.New("format first"),
				SpecErrs: map[meta.Spec]error{
					meta.SpecEXIF: errors.New("exif error"),
				},
			},
			wantMsg: "imx: format error: format first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestPartialError_Unwrap(t *testing.T) {
	formatErr := errors.New("format error")
	exifErr := errors.New("exif error")

	tests := []struct {
		name    string
		err     *PartialError
		wantErr error
	}{
		{
			name: "unwrap format error",
			err: &PartialError{
				FormatErr: formatErr,
			},
			wantErr: formatErr,
		},
		{
			name: "unwrap spec error when no format error",
			err: &PartialError{
				SpecErrs: map[meta.Spec]error{
					meta.SpecEXIF: exifErr,
				},
			},
			wantErr: exifErr,
		},
		{
			name:    "unwrap nil when empty",
			err:     &PartialError{},
			wantErr: nil,
		},
		{
			name: "format error takes precedence",
			err: &PartialError{
				FormatErr: formatErr,
				SpecErrs: map[meta.Spec]error{
					meta.SpecEXIF: exifErr,
				},
			},
			wantErr: formatErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Unwrap()
			if got != tt.wantErr {
				t.Errorf("Unwrap() = %v, want %v", got, tt.wantErr)
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
