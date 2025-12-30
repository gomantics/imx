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
