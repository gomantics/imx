package imx

import (
	"errors"
	"fmt"
)

// Sentinel errors
var (
	ErrUnknownFormat   = errors.New("imx: unknown format")
	ErrTruncatedData   = errors.New("imx: truncated data")
	ErrUnsupportedMeta = errors.New("imx: unsupported metadata block")
)

// PartialError represents errors that occurred during metadata extraction
// while still producing partial results
type PartialError struct {
	FormatErr     error
	NamespaceErrs map[Namespace]error
}

func (e *PartialError) Error() string {
	if e.FormatErr != nil {
		return fmt.Sprintf("imx: format error: %v", e.FormatErr)
	}
	if len(e.NamespaceErrs) > 0 {
		return fmt.Sprintf("imx: namespace errors: %v", e.NamespaceErrs)
	}
	return "imx: partial error"
}

func (e *PartialError) Unwrap() error {
	if e.FormatErr != nil {
		return e.FormatErr
	}
	// Return first namespace error
	for _, err := range e.NamespaceErrs {
		return err
	}
	return nil
}
