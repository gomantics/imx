package imx

import (
	"errors"
	"fmt"

	"github.com/gomantics/imx/internal/meta"
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
	FormatErr error
	SpecErrs  map[meta.Spec]error
}

func (e *PartialError) Error() string {
	var msgs []string
	if e.FormatErr != nil {
		msgs = append(msgs, fmt.Sprintf("format: %v", e.FormatErr))
	}
	for spec, err := range e.SpecErrs {
		msgs = append(msgs, fmt.Sprintf("%s: %v", spec, err))
	}
	if len(msgs) == 0 {
		return "imx: partial error"
	}
	if len(msgs) == 1 {
		return fmt.Sprintf("imx: %s", msgs[0])
	}
	return fmt.Sprintf("imx: multiple errors: %v", msgs)
}

func (e *PartialError) Unwrap() []error {
	var errs []error
	if e.FormatErr != nil {
		errs = append(errs, e.FormatErr)
	}
	for _, err := range e.SpecErrs {
		errs = append(errs, err)
	}
	return errs
}
