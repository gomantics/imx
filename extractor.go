package imx

import (
	"bufio"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/format"
	"github.com/gomantics/imx/internal/format/jpeg"
	"github.com/gomantics/imx/internal/meta"
	"github.com/gomantics/imx/internal/meta/exif"
)

// Format represents an image container format (JPEG, PNG, WebP, etc.)
type Format = format.Format

const (
	FormatJPEG = format.FormatJPEG
	FormatPNG  = format.FormatPNG
	FormatWebP = format.FormatWebP
	FormatTIFF = format.FormatTIFF
	FormatHEIF = format.FormatHEIF
)

// ExtractorConfig holds configuration options for metadata extraction
type ExtractorConfig struct {
	MaxBytes       int64        // Maximum bytes to read (0 = no limit)
	BufferSize     int          // Buffer size for reading (0 = default 64KB)
	Specs          []Spec       // Metadata specs to extract (nil/empty = all)
	Formats        []Format     // Formats to detect (nil/empty = all registered)
	StopOnFirstErr bool         // Stop on first error vs. continue with partial results
}

// Option is a functional option for configuring an Extractor
type Option func(*ExtractorConfig)

// WithMaxBytes sets the maximum number of bytes to read
func WithMaxBytes(n int64) Option {
	return func(cfg *ExtractorConfig) {
		cfg.MaxBytes = n
	}
}

// WithBufferSize sets the buffer size for reading
func WithBufferSize(n int) Option {
	return func(cfg *ExtractorConfig) {
		cfg.BufferSize = n
	}
}

// WithSpecs sets the metadata specs to extract
func WithSpecs(specs ...Spec) Option {
	return func(cfg *ExtractorConfig) {
		cfg.Specs = specs
	}
}

// WithFormats sets the formats to detect
func WithFormats(fs ...Format) Option {
	return func(cfg *ExtractorConfig) {
		cfg.Formats = fs
	}
}

// WithStopOnFirstError configures the extractor to stop on first error
func WithStopOnFirstError() Option {
	return func(cfg *ExtractorConfig) {
		cfg.StopOnFirstErr = true
	}
}

// Extractor is a reusable metadata extractor, safe for concurrent use
type Extractor struct {
	cfg           ExtractorConfig
	formatParsers []format.Parser
	metaParsers   []meta.Parser
}

// New creates a new Extractor with the given options
func New(opts ...Option) *Extractor {
	cfg := ExtractorConfig{
		BufferSize: 64 * 1024, // 64KB default
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	e := &Extractor{
		cfg: cfg,
		formatParsers: []format.Parser{
			jpeg.New(),
		},
		metaParsers: []meta.Parser{
			exif.New(),
		},
	}

	return e
}

// Metadata extracts metadata from an io.Reader
func (e *Extractor) Metadata(r io.Reader, opts ...Option) (Metadata, error) {
	// Clone config and apply per-call options
	cfg := e.cfg
	for _, opt := range opts {
		opt(&cfg)
	}

	// Wrap reader with limit if MaxBytes is set
	if cfg.MaxBytes > 0 {
		r = io.LimitReader(r, cfg.MaxBytes)
	}

	// Wrap with buffered reader
	br := bufio.NewReaderSize(r, cfg.BufferSize)

	// Step 1: Format detection
	peek, err := br.Peek(64)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: peek failed: %w", err)
	}

	var formatParser format.Parser
	for _, p := range e.formatParsers {
		if p.Detect(peek) {
			formatParser = p
			break
		}
	}
	if formatParser == nil {
		return Metadata{}, ErrUnknownFormat
	}

	// Step 2: Extract raw blocks from format
	blocks, err := formatParser.Parse(br)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: parse format: %w", err)
	}

	// Step 3: Parse metadata from blocks
	var allDirs []meta.Directory

	for _, metaParser := range e.metaParsers {
		spec := metaParser.Spec()

		// Apply spec filter
		if len(cfg.Specs) > 0 && !contains(cfg.Specs, spec) {
			continue
		}

		// Filter blocks for this spec
		relevantBlocks := filterBlocksForSpec(blocks, spec)
		if len(relevantBlocks) == 0 {
			continue
		}

		// Parse
		dirs, err := metaParser.Parse(relevantBlocks)
		if err != nil {
			if cfg.StopOnFirstErr {
				return Metadata{}, fmt.Errorf("imx: parse %s: %w", spec, err)
			}
			continue
		}

		allDirs = append(allDirs, dirs...)
	}

	// Step 4: Build result
	result := Metadata{Directories: allDirs}
	result.BuildIndex()

	return result, nil
}

// Helper functions

func filterBlocksForSpec(blocks []format.RawBlock, spec Spec) []format.RawBlock {
	var filtered []format.RawBlock
	for _, b := range blocks {
		if Spec(b.Spec) == spec {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

func contains(slice []Spec, item Spec) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
