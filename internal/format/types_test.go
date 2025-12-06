package format

import "testing"

func TestFormat_String(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		want   string
	}{
		{
			name:   "FormatJPEG returns jpeg",
			format: FormatJPEG,
			want:   "jpeg",
		},
		{
			name:   "FormatPNG returns png",
			format: FormatPNG,
			want:   "png",
		},
		{
			name:   "FormatWebP returns webp",
			format: FormatWebP,
			want:   "webp",
		},
		{
			name:   "FormatTIFF returns tiff",
			format: FormatTIFF,
			want:   "tiff",
		},
		{
			name:   "FormatHEIF returns heif",
			format: FormatHEIF,
			want:   "heif",
		},
		{
			name:   "unknown format returns unknown",
			format: Format(999),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.String()
			if got != tt.want {
				t.Errorf("Format.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
