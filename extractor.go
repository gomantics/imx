package imx

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gomantics/imx/internal/common"
	"github.com/gomantics/imx/internal/format"
	"github.com/gomantics/imx/internal/format/jpeg"
	"github.com/gomantics/imx/internal/meta"
	"github.com/gomantics/imx/internal/meta/exif"
	"github.com/gomantics/imx/internal/meta/icc"
	"github.com/gomantics/imx/internal/meta/iptc"
	"github.com/gomantics/imx/internal/meta/xmp"
)

// Extractor is a reusable metadata extractor, safe for concurrent use
type Extractor struct {
	cfg           Config
	formatParsers []format.Parser
	metaParsers   []meta.Parser
}

// New creates a new Extractor with the given options
func New(opts ...Option) *Extractor {
	cfg := defaultConfig()
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
			xmp.New(),
			icc.New(),
			iptc.New(),
		},
	}

	return e
}

// MetadataFromReader extracts metadata from an io.Reader
func (e *Extractor) MetadataFromReader(r io.Reader, opts ...Option) (Metadata, error) {
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
	var allDirs []common.Directory
	partialErr := &PartialError{
		SpecErrs: make(map[Spec]error),
	}

	for _, metaParser := range e.metaParsers {
		spec := metaParser.Spec()

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
			// Collect error but continue parsing other specs
			partialErr.SpecErrs[spec] = err
			continue
		}

		allDirs = append(allDirs, dirs...)
	}

	// Step 4: Build result
	result := Metadata{Directories: allDirs}
	result.BuildIndex()

	// Return partial error if any specs failed
	if len(partialErr.SpecErrs) > 0 {
		return result, partialErr
	}

	return result, nil
}

// MetadataFromFile extracts metadata from a file path using this extractor
func (e *Extractor) MetadataFromFile(path string, opts ...Option) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: open file: %w", err)
	}
	defer f.Close()

	return e.MetadataFromReader(f, opts...)
}

// MetadataFromBytes extracts metadata from a byte slice using this extractor
func (e *Extractor) MetadataFromBytes(data []byte, opts ...Option) (Metadata, error) {
	return e.MetadataFromReader(bytes.NewReader(data), opts...)
}

// MetadataFromURL extracts metadata from an HTTP/HTTPS URL using this extractor
func (e *Extractor) MetadataFromURL(url string, opts ...Option) (Metadata, error) {
	// Clone config and apply per-call options
	cfg := e.cfg
	for _, opt := range opts {
		opt(&cfg)
	}

	client := &http.Client{Timeout: cfg.HTTPTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Metadata{}, fmt.Errorf("imx: http status %d", resp.StatusCode)
	}

	return e.MetadataFromReader(resp.Body, opts...)
}

// Helper functions

func filterBlocksForSpec(blocks []common.RawBlock, spec Spec) []common.RawBlock {
	var filtered []common.RawBlock
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
