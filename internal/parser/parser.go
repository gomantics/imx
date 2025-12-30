package parser

import (
	"io"
)

// Parser is the interface for all parsers (JPEG, EXIF, XMP, etc.).
type Parser interface {
	// Name returns the parser name (e.g., "JPEG", "XMP")
	Name() string

	// Detect returns true if this parser can handle the data.
	Detect(r io.ReaderAt) bool

	// Parse returns parsed metadata directories and any errors encountered.
	// May return partial results even when errors occur.
	Parse(r io.ReaderAt) ([]Directory, *ParseError)
}
