package imx

import (
	"testing"
	"time"
)

func TestWithHTTPTimeout(t *testing.T) {
	cfg := config{}
	opt := WithHTTPTimeout(60 * time.Second)
	opt(&cfg)

	if cfg.HTTPTimeout != 60*time.Second {
		t.Errorf("WithHTTPTimeout() HTTPTimeout = %v, want 60s", cfg.HTTPTimeout)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.HTTPTimeout != 30*time.Second {
		t.Errorf("defaultConfig() HTTPTimeout = %v, want 30s", cfg.HTTPTimeout)
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := config{}

	if cfg.HTTPTimeout != 0 {
		t.Errorf("Default HTTPTimeout = %v, want 0", cfg.HTTPTimeout)
	}
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

func TestWithHTTPTimeout_Zero(t *testing.T) {
	// Zero should not panic (uses default)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Did not expect panic for HTTPTimeout=0, got: %v", r)
		}
	}()
	cfg := config{}
	opt := WithHTTPTimeout(0)
	opt(&cfg)
	if cfg.HTTPTimeout != 0 {
		t.Errorf("HTTPTimeout = %v, want 0", cfg.HTTPTimeout)
	}
}

func TestWithMaxBytes(t *testing.T) {
	cfg := config{}
	opt := WithMaxBytes(100 << 20) // 100MB
	opt(&cfg)

	if cfg.MaxBytes != 100<<20 {
		t.Errorf("WithMaxBytes() MaxBytes = %v, want 100MB", cfg.MaxBytes)
	}
}

func TestWithMaxBytes_Zero(t *testing.T) {
	cfg := config{}
	opt := WithMaxBytes(0) // unlimited
	opt(&cfg)

	if cfg.MaxBytes != 0 {
		t.Errorf("WithMaxBytes(0) MaxBytes = %v, want 0", cfg.MaxBytes)
	}
}

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

func TestWithBufferSize(t *testing.T) {
	cfg := config{}
	opt := WithBufferSize(128 << 10) // 128KB
	opt(&cfg)

	if cfg.BufferSize != 128<<10 {
		t.Errorf("WithBufferSize() BufferSize = %v, want 128KB", cfg.BufferSize)
	}
}

func TestWithBufferSize_Zero(t *testing.T) {
	cfg := config{}
	opt := WithBufferSize(0) // uses default
	opt(&cfg)

	if cfg.BufferSize != 0 {
		t.Errorf("WithBufferSize(0) BufferSize = %v, want 0", cfg.BufferSize)
	}
}

func TestWithBufferSize_Negative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for negative BufferSize")
		} else if msg, ok := r.(string); ok {
			if msg != "imx: BufferSize must be non-negative" {
				t.Errorf("Expected panic message about BufferSize, got: %s", msg)
			}
		}
	}()
	WithBufferSize(-1)
}
