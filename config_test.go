package imx

import (
	"testing"
	"time"
)

func TestWithMaxBytes(t *testing.T) {
	cfg := Config{}
	opt := WithMaxBytes(1024)
	opt(&cfg)

	if cfg.MaxBytes != 1024 {
		t.Errorf("WithMaxBytes() MaxBytes = %d, want 1024", cfg.MaxBytes)
	}
}

func TestWithBufferSize(t *testing.T) {
	cfg := Config{}
	opt := WithBufferSize(32768)
	opt(&cfg)

	if cfg.BufferSize != 32768 {
		t.Errorf("WithBufferSize() BufferSize = %d, want 32768", cfg.BufferSize)
	}
}

func TestWithStopOnFirstError(t *testing.T) {
	cfg := Config{}
	opt := WithStopOnFirstError(true)
	opt(&cfg)

	if !cfg.StopOnFirstErr {
		t.Error("WithStopOnFirstError(true) StopOnFirstErr should be true")
	}

	// Test false as well
	cfg2 := Config{StopOnFirstErr: true}
	opt2 := WithStopOnFirstError(false)
	opt2(&cfg2)

	if cfg2.StopOnFirstErr {
		t.Error("WithStopOnFirstError(false) StopOnFirstErr should be false")
	}
}

func TestWithHTTPTimeout(t *testing.T) {
	cfg := Config{}
	opt := WithHTTPTimeout(60 * time.Second)
	opt(&cfg)

	if cfg.HTTPTimeout != 60*time.Second {
		t.Errorf("WithHTTPTimeout() HTTPTimeout = %v, want 60s", cfg.HTTPTimeout)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.MaxBytes != 0 {
		t.Errorf("defaultConfig() MaxBytes = %d, want 0", cfg.MaxBytes)
	}
	if cfg.BufferSize != 64*1024 {
		t.Errorf("defaultConfig() BufferSize = %d, want %d", cfg.BufferSize, 64*1024)
	}
	if cfg.StopOnFirstErr {
		t.Errorf("defaultConfig() StopOnFirstErr = %v, want false", cfg.StopOnFirstErr)
	}
	if cfg.HTTPTimeout != 30*time.Second {
		t.Errorf("defaultConfig() HTTPTimeout = %v, want 30s", cfg.HTTPTimeout)
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}

	if cfg.MaxBytes != 0 {
		t.Errorf("Default MaxBytes = %d, want 0", cfg.MaxBytes)
	}
	if cfg.BufferSize != 0 {
		t.Errorf("Default BufferSize = %d, want 0", cfg.BufferSize)
	}
	if cfg.StopOnFirstErr {
		t.Errorf("Default StopOnFirstErr = %v, want false", cfg.StopOnFirstErr)
	}
}

func TestConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		opt   Option
		check func(t *testing.T, cfg Config)
	}{
		{
			name: "WithMaxBytes zero",
			opt:  WithMaxBytes(0),
			check: func(t *testing.T, cfg Config) {
				if cfg.MaxBytes != 0 {
					t.Errorf("MaxBytes = %d, want 0", cfg.MaxBytes)
				}
			},
		},
		{
			name: "WithBufferSize very large",
			opt:  WithBufferSize(1 << 30), // 1GB
			check: func(t *testing.T, cfg Config) {
				if cfg.BufferSize != 1<<30 {
					t.Errorf("BufferSize = %d, want %d", cfg.BufferSize, 1<<30)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{}
			tt.opt(&cfg)
			tt.check(t, cfg)
		})
	}
}

// Validation tests - panics on invalid inputs

func TestWithMaxBytes_Negative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for negative MaxBytes")
		} else if msg, ok := r.(string); ok {
			if msg != "imx: MaxBytes must be non-negative" {
				t.Errorf("Expected panic message about MaxBytes, got: %s", msg)
			}
		}
	}()
	WithMaxBytes(-1)
}

func TestWithBufferSize_Negative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for negative BufferSize")
		} else if msg, ok := r.(string); ok {
			if msg != "imx: BufferSize must be non-negative" {
				t.Errorf("Expected panic message about BufferSize non-negative, got: %s", msg)
			}
		}
	}()
	WithBufferSize(-100)
}

func TestWithBufferSize_TooSmall(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for buffer size < 1KB")
		} else if msg, ok := r.(string); ok {
			if msg != "imx: BufferSize should be at least 1KB (1024 bytes)" {
				t.Errorf("Expected panic message about BufferSize minimum, got: %s", msg)
			}
		}
	}()
	WithBufferSize(512)
}

func TestWithHTTPTimeout_Negative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for negative HTTPTimeout")
		} else if msg, ok := r.(string); ok {
			if msg != "imx: HTTPTimeout must be non-negative" {
				t.Errorf("Expected panic message about HTTPTimeout, got: %s", msg)
			}
		}
	}()
	WithHTTPTimeout(-1 * time.Second)
}

// Test that zero values are allowed

func TestWithBufferSize_Zero(t *testing.T) {
	// Zero should not panic (uses default)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Did not expect panic for BufferSize=0, got: %v", r)
		}
	}()
	cfg := Config{}
	opt := WithBufferSize(0)
	opt(&cfg)
	if cfg.BufferSize != 0 {
		t.Errorf("BufferSize = %d, want 0", cfg.BufferSize)
	}
}

func TestWithHTTPTimeout_Zero(t *testing.T) {
	// Zero should not panic (uses default)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Did not expect panic for HTTPTimeout=0, got: %v", r)
		}
	}()
	cfg := Config{}
	opt := WithHTTPTimeout(0)
	opt(&cfg)
	if cfg.HTTPTimeout != 0 {
		t.Errorf("HTTPTimeout = %v, want 0", cfg.HTTPTimeout)
	}
}
