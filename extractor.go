package imx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/cr2"
	"github.com/gomantics/imx/internal/parser/flac"
	"github.com/gomantics/imx/internal/parser/gif"
	"github.com/gomantics/imx/internal/parser/heic"
	"github.com/gomantics/imx/internal/parser/id3"
	"github.com/gomantics/imx/internal/parser/jpeg"
	"github.com/gomantics/imx/internal/parser/mp4"
	"github.com/gomantics/imx/internal/parser/png"
	"github.com/gomantics/imx/internal/parser/tiff"
	"github.com/gomantics/imx/internal/parser/webp"
)

// ErrUnknownFormat is returned when the file format is not recognized
var ErrUnknownFormat = errors.New("imx: unknown format")

// ErrMaxBytesExceeded is returned when reading beyond the configured MaxBytes limit.
var ErrMaxBytesExceeded = errors.New("imx: max bytes exceeded")

// Extractor is a reusable metadata extractor, safe for concurrent use
type Extractor struct {
	cfg     config
	parsers []parser.Parser
}

// New creates a new Extractor with the given options
func New(opts ...Option) *Extractor {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	e := &Extractor{
		cfg: cfg,
		// Parsers are stateless and safe to reuse across calls.
		parsers: []parser.Parser{
			// Image parsers (order matters - more specific first)
			jpeg.New(), // JPEG images
			heic.New(), // HEIC/HEIF images
			png.New(),  // PNG images
			webp.New(), // WebP images
			gif.New(),  // GIF images
			cr2.New(),  // CR2 must come before TIFF (CR2 files are TIFF-based)
			tiff.New(), // TIFF images

			// Audio parsers
			id3.New(),  // MP3 files with ID3 tags
			flac.New(), // FLAC audio files
			mp4.New(),  // M4A/MP4 audio files
		},
	}

	return e
}

// cloneConfig creates a copy of the extractor's config and applies per-call options
func (e *Extractor) cloneConfig(opts ...Option) config {
	cfg := e.cfg
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// metadataFromReaderAt extracts metadata from an io.ReaderAt (primary method)
func (e *Extractor) metadataFromReaderAt(r io.ReaderAt, cfg config) (*Metadata, error) {

	// Try each parser factory until one detects the format
	var selectedParser parser.Parser
	for _, p := range e.parsers {
		if p.Detect(r) {
			selectedParser = p
			break
		}
	}

	if selectedParser == nil {
		return nil, ErrUnknownFormat
	}

	// Parse metadata - no transformation needed, types are already compatible
	dirs, parseErr := selectedParser.Parse(r)

	// Collect errors
	var errs []error
	if parseErr != nil {
		errs = parseErr.Unwrap()
	}

	if parseErr != nil && errors.Is(parseErr, ErrMaxBytesExceeded) {
		return nil, ErrMaxBytesExceeded
	}

	return &Metadata{
		directories: dirs,
		errors:      errs,
	}, nil
}

// MetadataFromFile extracts metadata from a file path
func (e *Extractor) MetadataFromFile(path string, opts ...Option) (*Metadata, error) {
	cfg := e.cloneConfig(opts...)

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("imx: open file: %w", err)
	}
	defer f.Close()

	if cfg.MaxBytes > 0 {
		info, statErr := f.Stat()
		if statErr != nil {
			return nil, fmt.Errorf("imx: stat file: %w", statErr)
		}
		if info.Size() > cfg.MaxBytes {
			return nil, ErrMaxBytesExceeded
		}
	}

	// os.File implements io.ReaderAt, parsers will handle EOF correctly
	return e.metadataFromReaderAt(&boundedReaderAt{r: f, limit: cfg.MaxBytes}, cfg)
}

// MetadataFromBytes extracts metadata from a byte slice
func (e *Extractor) MetadataFromBytes(data []byte, opts ...Option) (*Metadata, error) {
	cfg := e.cloneConfig(opts...)

	if cfg.MaxBytes > 0 && int64(len(data)) > cfg.MaxBytes {
		return nil, ErrMaxBytesExceeded
	}

	return e.metadataFromReaderAt(bytes.NewReader(data), cfg)
}

// MetadataFromReader extracts metadata from an io.Reader using a smart buffering adapter.
// This adapter implements io.ReaderAt by buffering data as it's read, avoiding the need
// to load the entire stream into memory upfront.
func (e *Extractor) MetadataFromReader(r io.Reader, opts ...Option) (*Metadata, error) {
	cfg := e.cloneConfig(opts...)

	// Create a smart reader adapter that implements ReaderAt via buffering
	adapter := newReaderAdapter(r, cfg.MaxBytes, cfg.BufferSize)

	// The adapter will handle reading on demand
	return e.metadataFromReaderAt(adapter, cfg)
}

// MetadataFromURL extracts metadata from an HTTP/HTTPS URL
func (e *Extractor) MetadataFromURL(url string, opts ...Option) (*Metadata, error) {
	// Clone config and apply per-call options
	cfg := e.cloneConfig(opts...)

	client := &http.Client{Timeout: cfg.HTTPTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("imx: fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("imx: fetch url: http status %d", resp.StatusCode)
	}

	return e.MetadataFromReader(resp.Body, opts...)
}
