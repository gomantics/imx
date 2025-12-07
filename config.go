package imx

import "time"

// Config holds configuration options for metadata extraction
type Config struct {
	MaxBytes       int64         // Maximum bytes to read (0 = no limit)
	BufferSize     int           // Buffer size for reading (0 = default 64KB)
	Specs          []Spec        // Metadata specs to extract (nil/empty = all)
	Formats        []Format      // Formats to detect (nil/empty = all registered)
	StopOnFirstErr bool          // Stop on first error vs. continue with partial results
	HTTPTimeout    time.Duration // HTTP request timeout for URL fetching (0 = 30s default)
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

// WithSpecs sets the metadata specs to extract
func WithSpecs(specs ...Spec) Option {
	return func(cfg *Config) {
		cfg.Specs = specs
	}
}

// WithFormats sets the formats to detect
func WithFormats(fs ...Format) Option {
	return func(cfg *Config) {
		cfg.Formats = fs
	}
}

// WithStopOnFirstError configures the extractor to stop on first error
func WithStopOnFirstError() Option {
	return func(cfg *Config) {
		cfg.StopOnFirstErr = true
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
