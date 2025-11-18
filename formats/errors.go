package formats

import "errors"

var (
	// ErrInvalidData indicates malformed or incomplete format data.
	ErrInvalidData = errors.New("formats: invalid data")

	// ErrUnsupportedFormat is returned when a parser is not available.
	ErrUnsupportedFormat = errors.New("formats: unsupported format")
)

