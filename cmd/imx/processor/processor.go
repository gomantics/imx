package processor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/gomantics/imx"
	"github.com/gomantics/imx/cmd/imx/filter"
	"github.com/gomantics/imx/cmd/imx/output"
	"github.com/gomantics/imx/cmd/imx/ui"
	"github.com/gomantics/imx/cmd/imx/util"
)

// Config holds processor configuration
type Config struct {
	// Processing options
	Workers      int
	Verbose      bool
	Quiet        bool
	StopOnErr    bool
	ShowProgress bool

	// Filter options
	Filter filter.Filter
}

// Processor handles concurrent metadata extraction
type Processor struct {
	config    *Config
	extractor *imx.Extractor
}

// New creates a new processor with the given configuration
func New(config *Config) *Processor {
	if config.Workers <= 0 {
		config.Workers = runtime.NumCPU()
	}
	return &Processor{
		config:    config,
		extractor: imx.New(),
	}
}

// Process processes multiple files and returns results
func (p *Processor) Process(ctx context.Context, files []string) ([]*output.Result, error) {
	if len(files) == 0 {
		return nil, nil
	}

	// Create progress bar if needed
	var bar *ui.ProgressBar
	if p.config.ShowProgress && !p.config.Quiet && len(files) > 1 {
		bar = ui.NewProgressBarWithOutput(len(files), "Processing", os.Stderr)
	}

	// Create worker pool
	results := make([]*output.Result, len(files))
	jobs := make(chan job, len(files))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < p.config.Workers; i++ {
		wg.Add(1)
		go p.worker(ctx, &wg, jobs, results, bar)
	}

	// Send jobs
	for i, file := range files {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		case jobs <- job{index: i, file: file}:
		}
	}
	close(jobs)

	// Wait for completion
	wg.Wait()

	if bar != nil {
		bar.Finish()
		fmt.Println() // Add newline after progress bar
	}

	// Check for fatal errors if stop-on-error is enabled
	if p.config.StopOnErr {
		for _, result := range results {
			if result.Error != nil {
				return results, result.Error
			}
		}
	}

	return results, nil
}

// ProcessSingle processes a single file
func (p *Processor) ProcessSingle(ctx context.Context, file string) (*output.Result, error) {
	result := &output.Result{File: file}

	// Read file data
	data, err := p.readFile(ctx, file)
	if err != nil {
		result.Error = &util.ProcessError{
			File: file,
			Op:   "read",
			Err:  err,
		}
		return result, result.Error
	}

	// Extract metadata
	meta, err := p.extractor.MetadataFromBytes(data)
	if err != nil {
		result.Error = &util.ProcessError{
			File: file,
			Op:   "extract",
			Err:  err,
		}
		return result, result.Error
	}

	result.Meta = &meta

	// Apply filters and collect tags
	var tags []output.TagInfo
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		// Apply filter if configured
		if p.config.Filter != nil && !p.config.Filter.ShouldInclude(dir, tag) {
			return true
		}

		tags = append(tags, output.TagInfo{
			Dir: dir,
			Tag: tag,
		})
		return true
	})

	result.Tags = tags
	result.TagCount = len(tags)

	return result, nil
}

// job represents a processing job
type job struct {
	index int
	file  string
}

// worker processes jobs from the jobs channel
func (p *Processor) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan job, results []*output.Result, bar *ui.ProgressBar) {
	defer wg.Done()

	for j := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			result, _ := p.ProcessSingle(ctx, j.file)
			results[j.index] = result

			if bar != nil {
				bar.Add(1)
			}

			// Log verbose output
			if p.config.Verbose && !p.config.Quiet {
				if result.Error != nil {
					fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", j.file, result.Error)
				} else {
					fmt.Fprintf(os.Stderr, "Processed %s (%d tags)\n", j.file, result.TagCount)
				}
			}
		}
	}
}

// readFile reads file data from path or URL
func (p *Processor) readFile(ctx context.Context, path string) ([]byte, error) {
	if util.IsURL(path) {
		return p.readURL(ctx, path)
	}
	return os.ReadFile(path)
}

// readURL fetches data from a URL
func (p *Processor) readURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}
