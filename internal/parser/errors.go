package parser

import (
	"errors"
	"fmt"
)

// ParseError holds multiple errors from parsing.
// Allows returning partial results with errors.
type ParseError struct {
	errs []error
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e == nil || len(e.errs) == 0 {
		return ""
	}

	if len(e.errs) == 1 {
		return e.errs[0].Error()
	}

	msg := fmt.Sprintf("%d errors occurred:\n", len(e.errs))
	for i, err := range e.errs {
		msg += fmt.Sprintf("  %d. %v\n", i+1, err)
	}
	return msg
}

// Unwrap returns the underlying errors.
func (e *ParseError) Unwrap() []error {
	if e == nil {
		return nil
	}
	return e.errs
}

// Is allows errors.Is to match underlying errors.
func (e *ParseError) Is(target error) bool {
	for _, err := range e.errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// Add appends an error to the ParseError.
func (e *ParseError) Add(err error) {
	if err != nil {
		e.errs = append(e.errs, err)
	}
}

// Merge merges another ParseError into this one.
func (e *ParseError) Merge(other *ParseError) {
	if other == nil {
		return
	}
	e.errs = append(e.errs, other.errs...)
}

// OrNil returns nil if there are no errors, otherwise returns the ParseError.
func (e *ParseError) OrNil() *ParseError {
	if e == nil || len(e.errs) == 0 {
		return nil
	}
	return e
}

// NewParseError creates a ParseError from multiple errors.
func NewParseError(errs ...error) *ParseError {
	return &ParseError{errs: errs}
}
