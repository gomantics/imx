package util

import "fmt"

// ProcessError represents an error that occurred during file processing
type ProcessError struct {
	File    string // File path that caused the error
	Op      string // Operation that failed (e.g., "extract", "read", "parse")
	Err     error  // Underlying error
	Partial bool   // True if partial results are available despite the error
}

// Error implements the error interface
func (e *ProcessError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s: %s: %v", e.File, e.Op, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *ProcessError) Unwrap() error {
	return e.Err
}

// NewProcessError creates a new ProcessError
func NewProcessError(file, op string, err error) *ProcessError {
	return &ProcessError{
		File:    file,
		Op:      op,
		Err:     err,
		Partial: false,
	}
}

// NewPartialError creates a ProcessError that indicates partial results are available
func NewPartialError(file, op string, err error) *ProcessError {
	return &ProcessError{
		File:    file,
		Op:      op,
		Err:     err,
		Partial: true,
	}
}
