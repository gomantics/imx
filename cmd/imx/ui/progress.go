package ui

import (
	"io"

	"github.com/schollz/progressbar/v3"
)

// ProgressBar wraps schollz/progressbar for consistent styling
type ProgressBar struct {
	bar *progressbar.ProgressBar
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, description string) *ProgressBar {
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(io.Discard), // Don't show by default
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65),
		progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	return &ProgressBar{bar: bar}
}

// NewProgressBarWithOutput creates a progress bar that writes to the given writer
func NewProgressBarWithOutput(total int, description string, w io.Writer) *ProgressBar {
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(w),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65),
		progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	return &ProgressBar{bar: bar}
}

// Increment increments the progress bar by 1
func (p *ProgressBar) Increment() error {
	return p.bar.Add(1)
}

// Add adds n to the progress bar
func (p *ProgressBar) Add(n int) error {
	return p.bar.Add(n)
}

// Set sets the progress bar to a specific value
func (p *ProgressBar) Set(n int) error {
	return p.bar.Set(n)
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() error {
	return p.bar.Finish()
}

// Clear clears the progress bar
func (p *ProgressBar) Clear() error {
	return p.bar.Clear()
}
