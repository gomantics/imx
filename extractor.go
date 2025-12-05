package imx

import (
	"bufio"
	"io"

	"github.com/gomantics/imx/internal/pipeline"
	"github.com/gomantics/imx/internal/types"
)

// Format represents a container format (JPEG, PNG, WebP, etc.)
type Format = types.Format

const (
	FormatJPEG = types.FormatJPEG
	FormatPNG  = types.FormatPNG
	FormatWebP = types.FormatWebP
	FormatTIFF = types.FormatTIFF
	FormatHEIF = types.FormatHEIF
)

// ExtractorConfig holds configuration options for metadata extraction
type ExtractorConfig = types.ExtractorConfig

// Option is a functional option for configuring an Extractor
type Option func(*types.ExtractorConfig)

// WithMaxBytes sets the maximum number of bytes to read
func WithMaxBytes(n int64) Option {
	return func(cfg *types.ExtractorConfig) {
		cfg.MaxBytes = n
	}
}

// WithBufferSize sets the buffer size for reading
func WithBufferSize(n int) Option {
	return func(cfg *types.ExtractorConfig) {
		cfg.BufferSize = n
	}
}

// WithNamespaces sets the namespaces to extract
func WithNamespaces(ns ...Namespace) Option {
	return func(cfg *types.ExtractorConfig) {
		cfg.Namespaces = ns
	}
}

// WithFormats sets the formats to detect
func WithFormats(fs ...Format) Option {
	return func(cfg *types.ExtractorConfig) {
		cfg.Formats = fs
	}
}

// WithStopOnFirstError configures the extractor to stop on first error
func WithStopOnFirstError() Option {
	return func(cfg *types.ExtractorConfig) {
		cfg.StopOnFirstErr = true
	}
}

// Extractor is a reusable metadata extractor, safe for concurrent use
type Extractor struct {
	cfg      types.ExtractorConfig
	pipeline *pipeline.Pipeline
}

// New creates a new Extractor with the given options
func New(opts ...Option) *Extractor {
	cfg := types.ExtractorConfig{
		BufferSize: 64 * 1024, // 64KB default
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Extractor{
		cfg:      cfg,
		pipeline: pipeline.New(),
	}
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

	// Execute pipeline
	pipelineResult, err := e.pipeline.Extract(br, cfg)
	if err != nil {
		return Metadata{}, err
	}

	// Convert pipeline.Metadata to imx.Metadata
	result := Metadata{
		Directories: pipelineResult.Directories,
	}
	result.BuildIndex()

	return result, nil
}

// Default extractor instance
var defaultExtractor = New()

// Extract extracts metadata from an io.Reader using the default extractor
func Extract(r io.Reader, opts ...Option) (Metadata, error) {
	return defaultExtractor.Metadata(r, opts...)
}
