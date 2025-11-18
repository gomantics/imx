package imx

import "errors"

var (
	// ErrUnsupportedFormat is returned when the image format cannot be detected.
	ErrUnsupportedFormat = errors.New("imx: unsupported format")

	// ErrInvalidSource is returned when the provided data source cannot be read.
	ErrInvalidSource = errors.New("imx: invalid source")

	// ErrFetchFailed indicates that fetching a remote resource failed.
	ErrFetchFailed = errors.New("imx: fetch failed")
)
