package imx

import "time"

// Config holds configuration options for metadata extraction
type Config struct {
	MaxBytes       int64         // Maximum bytes to read (0 = no limit)
	BufferSize     int           // Buffer size for reading
	StopOnFirstErr bool          // Stop on first error vs. continue with partial results
	HTTPTimeout    time.Duration // HTTP request timeout for URL fetching

	// TODO: Add support for custom format and spec filters in a future version.
	// This would allow users to register custom parsers or filter which specs to extract.
	// Example API:
	//   - WithFormatFilter(func(Format) bool)
	//   - WithSpecFilter(func(Spec) bool)
	//   - RegisterCustomParser(Parser)
}

// defaultConfig returns a Config with reasonable defaults
func defaultConfig() Config {
	return Config{
		MaxBytes:       0,              // No limit
		BufferSize:     64 * 1024,      // 64KB
		StopOnFirstErr: false,          // Continue on errors for partial results
		HTTPTimeout:    30 * time.Second, // 30 second timeout
	}
}

// Option is a functional option for configuring an Extractor
type Option func(*Config)

// WithMaxBytes sets the maximum number of bytes to read.
// Panics if n is negative.
func WithMaxBytes(n int64) Option {
	if n < 0 {
		panic("imx: MaxBytes must be non-negative")
	}
	return func(cfg *Config) {
		cfg.MaxBytes = n
	}
}

// WithBufferSize sets the buffer size for reading.
// Panics if n is negative or if n is positive but less than 1KB.
func WithBufferSize(n int) Option {
	if n < 0 {
		panic("imx: BufferSize must be non-negative")
	}
	if n > 0 && n < 1024 {
		panic("imx: BufferSize should be at least 1KB (1024 bytes)")
	}
	return func(cfg *Config) {
		cfg.BufferSize = n
	}
}

// WithStopOnFirstError configures whether the extractor should stop on first error
func WithStopOnFirstError(stop bool) Option {
	return func(cfg *Config) {
		cfg.StopOnFirstErr = stop
	}
}

// WithHTTPTimeout sets the HTTP request timeout for URL fetching.
// Panics if d is negative.
func WithHTTPTimeout(d time.Duration) Option {
	if d < 0 {
		panic("imx: HTTPTimeout must be non-negative")
	}
	return func(cfg *Config) {
		cfg.HTTPTimeout = d
	}
}
