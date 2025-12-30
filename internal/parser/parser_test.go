package parser

import (
	"errors"
	"testing"
)

// TestParseError_Error tests the Error method.
func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name string
		pe   *ParseError
		want string
	}{
		{
			name: "nil ParseError",
			pe:   nil,
			want: "",
		},
		{
			name: "empty ParseError",
			pe:   &ParseError{},
			want: "",
		},
		{
			name: "single error",
			pe:   &ParseError{errs: []error{errors.New("test error")}},
			want: "test error",
		},
		{
			name: "multiple errors",
			pe:   &ParseError{errs: []error{errors.New("error 1"), errors.New("error 2")}},
			want: "2 errors occurred:\n  1. error 1\n  2. error 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pe.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestParseError_Unwrap tests the Unwrap method.
func TestParseError_Unwrap(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	tests := []struct {
		name    string
		pe      *ParseError
		wantLen int
	}{
		{
			name:    "nil ParseError",
			pe:      nil,
			wantLen: 0,
		},
		{
			name:    "empty ParseError",
			pe:      &ParseError{},
			wantLen: 0,
		},
		{
			name:    "single error",
			pe:      &ParseError{errs: []error{err1}},
			wantLen: 1,
		},
		{
			name:    "multiple errors",
			pe:      &ParseError{errs: []error{err1, err2}},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pe.Unwrap()
			if len(got) != tt.wantLen {
				t.Errorf("Unwrap() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// TestParseError_Add tests the Add method.
func TestParseError_Add(t *testing.T) {
	tests := []struct {
		name    string
		initial []error
		add     error
		wantLen int
	}{
		{
			name:    "add nil error",
			initial: nil,
			add:     nil,
			wantLen: 0,
		},
		{
			name:    "add error to empty",
			initial: nil,
			add:     errors.New("new error"),
			wantLen: 1,
		},
		{
			name:    "add error to existing",
			initial: []error{errors.New("existing")},
			add:     errors.New("new error"),
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &ParseError{errs: tt.initial}
			pe.Add(tt.add)
			if len(pe.errs) != tt.wantLen {
				t.Errorf("Add() resulted in len = %d, want %d", len(pe.errs), tt.wantLen)
			}
		})
	}
}

// TestParseError_Merge tests the Merge method.
func TestParseError_Merge(t *testing.T) {
	tests := []struct {
		name    string
		pe      *ParseError
		other   *ParseError
		wantLen int
	}{
		{
			name:    "merge nil",
			pe:      &ParseError{},
			other:   nil,
			wantLen: 0,
		},
		{
			name:    "merge empty",
			pe:      &ParseError{},
			other:   &ParseError{},
			wantLen: 0,
		},
		{
			name:    "merge into empty",
			pe:      &ParseError{},
			other:   &ParseError{errs: []error{errors.New("error 1")}},
			wantLen: 1,
		},
		{
			name:    "merge with existing",
			pe:      &ParseError{errs: []error{errors.New("existing")}},
			other:   &ParseError{errs: []error{errors.New("error 1"), errors.New("error 2")}},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pe.Merge(tt.other)
			if len(tt.pe.errs) != tt.wantLen {
				t.Errorf("Merge() resulted in len = %d, want %d", len(tt.pe.errs), tt.wantLen)
			}
		})
	}
}

// TestParseError_OrNil tests the OrNil method.
func TestParseError_OrNil(t *testing.T) {
	tests := []struct {
		name    string
		pe      *ParseError
		wantNil bool
	}{
		{
			name:    "nil ParseError",
			pe:      nil,
			wantNil: true,
		},
		{
			name:    "empty ParseError",
			pe:      &ParseError{},
			wantNil: true,
		},
		{
			name:    "ParseError with errors",
			pe:      &ParseError{errs: []error{errors.New("error")}},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pe.OrNil()
			if (got == nil) != tt.wantNil {
				t.Errorf("OrNil() = %v, wantNil = %v", got, tt.wantNil)
			}
		})
	}
}

// TestNewParseError tests the NewParseError constructor.
func TestNewParseError(t *testing.T) {
	tests := []struct {
		name    string
		errs    []error
		wantLen int
	}{
		{
			name:    "no errors",
			errs:    nil,
			wantLen: 0,
		},
		{
			name:    "single error",
			errs:    []error{errors.New("error 1")},
			wantLen: 1,
		},
		{
			name:    "multiple errors",
			errs:    []error{errors.New("error 1"), errors.New("error 2")},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := NewParseError(tt.errs...)
			if pe == nil {
				t.Fatal("NewParseError() returned nil")
			}
			if len(pe.errs) != tt.wantLen {
				t.Errorf("NewParseError() len = %d, want %d", len(pe.errs), tt.wantLen)
			}
		})
	}
}
