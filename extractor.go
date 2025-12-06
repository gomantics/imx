package imx

import (
	"bufio"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/format"
	"github.com/gomantics/imx/internal/format/jpeg"
	"github.com/gomantics/imx/internal/meta"
	"github.com/gomantics/imx/internal/meta/exif"
	"github.com/gomantics/imx/internal/types"
)

// Format represents an image container format (JPEG, PNG, WebP, etc.)
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
	cfg           types.ExtractorConfig
	formatParsers []format.Parser
	metaParsers   []meta.Parser
}

// New creates a new Extractor with the given options
func New(opts ...Option) *Extractor {
	cfg := types.ExtractorConfig{
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
		ns := metaParser.Namespace()

		// Apply namespace filter
		if len(cfg.Namespaces) > 0 && !contains(cfg.Namespaces, ns) {
			continue
		}

		// Filter blocks for this namespace
		relevantBlocks := filterBlocksForNamespace(blocks, ns)
		if len(relevantBlocks) == 0 {
			continue
		}

		// Parse
		dirs, err := metaParser.Parse(relevantBlocks)
		if err != nil {
			if cfg.StopOnFirstErr {
				return Metadata{}, fmt.Errorf("imx: parse %s: %w", ns, err)
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

func filterBlocksForNamespace(blocks []format.RawBlock, ns types.Namespace) []format.RawBlock {
	kind := namespaceToMetaKind(ns)
	var filtered []format.RawBlock
	for _, b := range blocks {
		if b.Kind == kind {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

func namespaceToMetaKind(ns types.Namespace) types.MetaKind {
	switch ns {
	case types.NamespaceEXIF:
		return types.MetaKindEXIF
	case types.NamespaceIPTC:
		return types.MetaKindIPTC
	case types.NamespaceXMP:
		return types.MetaKindXMP
	case types.NamespaceICC:
		return types.MetaKindICC
	default:
		return -1
	}
}

func contains(slice []types.Namespace, item types.Namespace) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Default extractor instance
var defaultExtractor = New()

// Extract extracts metadata from an io.Reader using the default extractor
func Extract(r io.Reader, opts ...Option) (Metadata, error) {
	return defaultExtractor.Metadata(r, opts...)
}
