package pipeline

import (
	"bufio"
	"fmt"

	"github.com/gomantics/imx/internal/container"
	"github.com/gomantics/imx/internal/meta"
	"github.com/gomantics/imx/internal/types"
)

// Pipeline orchestrates the extraction process
type Pipeline struct{}

// New creates a new pipeline
func New() *Pipeline {
	return &Pipeline{}
}

// Metadata is the public metadata type
type Metadata struct {
	Directories []meta.Directory
}

// Extract executes the full extraction pipeline
func (p *Pipeline) Extract(r *bufio.Reader, cfg types.ExtractorConfig) (Metadata, error) {
	// Step 1: Format detection
	peek, err := r.Peek(64)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: peek failed: %w", err)
	}

	parser := container.Detect(peek)
	if parser == nil {
		return Metadata{}, fmt.Errorf("imx: unknown format")
	}

	// Step 2: Container parsing - extract raw blocks
	blocks, err := parser.Parse(r, cfg)
	if err != nil {
		return Metadata{}, fmt.Errorf("imx: parse container: %w", err)
	}

	// Step 3: Namespace parsing - decode blocks into directories
	var allDirs []meta.Directory
	nsRegistry := meta.GetRegistry()
	nsErrs := make(map[types.Namespace]error)

	for ns, nsParser := range nsRegistry.All() {
		// Filter blocks for this namespace
		var relevantBlocks []container.RawBlock
		for _, block := range blocks {
			if p.isRelevant(block.Kind, ns) {
				relevantBlocks = append(relevantBlocks, block)
			}
		}

		if len(relevantBlocks) == 0 {
			continue
		}

		// Parse blocks
		dirs, err := nsParser.Parse(relevantBlocks, cfg)
		if err != nil {
			nsErrs[ns] = err
			if cfg.StopOnFirstErr {
				return Metadata{}, fmt.Errorf("imx: parse namespace %s: %w", ns, err)
			}
			continue
		}

		allDirs = append(allDirs, dirs...)
	}

	// Step 4: Filter by requested namespaces
	if len(cfg.Namespaces) > 0 {
		filtered := make([]meta.Directory, 0)
		for _, dir := range allDirs {
			for _, ns := range cfg.Namespaces {
				if dir.Namespace == ns {
					filtered = append(filtered, dir)
					break
				}
			}
		}
		allDirs = filtered
	}

	// Step 5: Build metadata
	result := Metadata{
		Directories: allDirs,
	}

	return result, nil
}

// isRelevant checks if a MetaKind is relevant for a namespace
func (p *Pipeline) isRelevant(kind container.MetaKind, ns types.Namespace) bool {
	switch ns {
	case types.NamespaceEXIF:
		return kind == container.MetaKindEXIF
	case types.NamespaceIPTC:
		return kind == container.MetaKindIPTC
	case types.NamespaceXMP:
		return kind == container.MetaKindXMP
	case types.NamespaceICC:
		return kind == container.MetaKindICC
	default:
		return false
	}
}
