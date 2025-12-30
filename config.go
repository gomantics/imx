package imx

import "time"

// config holds configuration options for metadata extraction.
// This type is unexported; users configure via Option functions.
type config struct {
	HTTPTimeout    time.Duration // HTTP request timeout for URL fetching
	MaxBytes       int64         // Maximum bytes to read from any source (0 = unlimited)
	BufferSize     int           // Read buffer size for streaming sources
}

// defaultConfig returns a config with reasonable defaults
func defaultConfig() config {
	return config{
		HTTPTimeout:    30 * time.Second, // 30 second timeout
		MaxBytes:       1 << 30,          // 1GB limit to handle large RAW files
		BufferSize:     64 << 10,         // 64KB streaming buffer
	}
}

// Option is a functional option for configuring an Extractor
type Option func(*config)

// WithHTTPTimeout sets the HTTP request timeout for URL fetching.
// The timeout applies only to MetadataFromURL operations.
//
// Panics if d is negative. A timeout of 0 means no timeout (unlimited).
func WithHTTPTimeout(d time.Duration) Option {
	if d < 0 {
		panic("imx: HTTPTimeout must be non-negative")
	}
	return func(cfg *config) {
		cfg.HTTPTimeout = d
	}
}

// WithMaxBytes sets an upper bound on the total bytes that can be read from
// any source (file, reader, or URL). A value of 0 means no limit.
func WithMaxBytes(n int64) Option {
	if n < 0 {
		panic("imx: MaxBytes must be non-negative")
	}
	return func(cfg *config) {
		cfg.MaxBytes = n
	}
}

// WithBufferSize sets the streaming read buffer size used for reader/URL inputs.
// A value of 0 falls back to the default buffer size.
func WithBufferSize(n int) Option {
	if n < 0 {
		panic("imx: BufferSize must be non-negative")
	}
	return func(cfg *config) {
		cfg.BufferSize = n
	}
}
