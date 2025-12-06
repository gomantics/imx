package imx

import (
	"testing"
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

func TestWithSpecs(t *testing.T) {
	cfg := Config{}
	opt := WithSpecs(SpecEXIF, SpecXMP)
	opt(&cfg)

	if len(cfg.Specs) != 2 {
		t.Errorf("WithSpecs() len(Specs) = %d, want 2", len(cfg.Specs))
	}
	if cfg.Specs[0] != SpecEXIF || cfg.Specs[1] != SpecXMP {
		t.Error("WithSpecs() Specs not set correctly")
	}
}

func TestWithFormats(t *testing.T) {
	cfg := Config{}
	opt := WithFormats(FormatJPEG, FormatPNG)
	opt(&cfg)

	if len(cfg.Formats) != 2 {
		t.Errorf("WithFormats() len(Formats) = %d, want 2", len(cfg.Formats))
	}
}

func TestWithStopOnFirstError(t *testing.T) {
	cfg := Config{}
	opt := WithStopOnFirstError()
	opt(&cfg)

	if !cfg.StopOnFirstErr {
		t.Error("WithStopOnFirstError() StopOnFirstErr should be true")
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}

	if cfg.MaxBytes != 0 {
		t.Error("Default MaxBytes should be 0")
	}
	if cfg.BufferSize != 0 {
		t.Error("Default BufferSize should be 0")
	}
	if cfg.StopOnFirstErr {
		t.Error("Default StopOnFirstErr should be false")
	}
	if cfg.Specs != nil {
		t.Error("Default Specs should be nil")
	}
	if cfg.Formats != nil {
		t.Error("Default Formats should be nil")
	}
}
