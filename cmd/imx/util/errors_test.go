package util

import (
	"errors"
	"testing"
)

func TestProcessError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ProcessError
		want string
	}{
		{
			name: "with file",
			err: &ProcessError{
				File: "photo.jpg",
				Op:   "extract",
				Err:  errors.New("invalid format"),
			},
			want: "photo.jpg: extract: invalid format",
		},
		{
			name: "without file",
			err: &ProcessError{
				File: "",
				Op:   "read",
				Err:  errors.New("permission denied"),
			},
			want: "read: permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProcessError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	procErr := &ProcessError{
		File: "test.jpg",
		Op:   "parse",
		Err:  baseErr,
	}

	unwrapped := procErr.Unwrap()
	if unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}

	if !errors.Is(procErr, baseErr) {
		t.Error("errors.Is() should return true for wrapped error")
	}
}

func TestNewProcessError(t *testing.T) {
	err := errors.New("test error")
	procErr := NewProcessError("test.jpg", "extract", err)

	if procErr.File != "test.jpg" {
		t.Errorf("File = %q, want %q", procErr.File, "test.jpg")
	}
	if procErr.Op != "extract" {
		t.Errorf("Op = %q, want %q", procErr.Op, "extract")
	}
	if procErr.Err != err {
		t.Errorf("Err = %v, want %v", procErr.Err, err)
	}
	if procErr.Partial {
		t.Error("Partial should be false")
	}
}

func TestNewPartialError(t *testing.T) {
	err := errors.New("test error")
	procErr := NewPartialError("test.jpg", "extract", err)

	if !procErr.Partial {
		t.Error("Partial should be true")
	}
	if procErr.File != "test.jpg" {
		t.Errorf("File = %q, want %q", procErr.File, "test.jpg")
	}
}
