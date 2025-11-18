package formats

import (
	"fmt"
	"io"
)

// Extract dispatches to the appropriate format parser based on the format string.
func Extract(format string, r io.ReadSeeker) (*Result, error) {
	switch format {
	case "JPEG":
		return ExtractJPEG(r)
	case "PNG":
		return ExtractPNG(r)
	case "GIF":
		return ExtractGIF(r)
	case "WebP":
		return ExtractWebP(r)
	case "BMP":
		return ExtractBMP(r)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}
}
