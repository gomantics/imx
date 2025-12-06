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
		t.Errorf("Default MaxBytes = %d, want 0", cfg.MaxBytes)
	}
	if cfg.BufferSize != 0 {
		t.Errorf("Default BufferSize = %d, want 0", cfg.BufferSize)
	}
	if cfg.StopOnFirstErr {
		t.Errorf("Default StopOnFirstErr = %v, want false", cfg.StopOnFirstErr)
	}
	if cfg.Specs != nil {
		t.Errorf("Default Specs = %v, want nil", cfg.Specs)
	}
	if cfg.Formats != nil {
		t.Errorf("Default Formats = %v, want nil", cfg.Formats)
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
		{
			name: "WithSpecs empty slice",
			opt:  WithSpecs(),
			check: func(t *testing.T, cfg Config) {
				if len(cfg.Specs) != 0 {
					t.Errorf("len(Specs) = %d, want 0", len(cfg.Specs))
				}
			},
		},
		{
			name: "WithFormats empty slice",
			opt:  WithFormats(),
			check: func(t *testing.T, cfg Config) {
				if len(cfg.Formats) != 0 {
					t.Errorf("len(Formats) = %d, want 0", len(cfg.Formats))
				}
			},
		},
		{
			name: "WithSpecs single spec",
			opt:  WithSpecs(SpecEXIF),
			check: func(t *testing.T, cfg Config) {
				if len(cfg.Specs) != 1 {
					t.Errorf("len(Specs) = %d, want 1", len(cfg.Specs))
				}
				if len(cfg.Specs) > 0 && cfg.Specs[0] != SpecEXIF {
					t.Errorf("Specs[0] = %v, want %v", cfg.Specs[0], SpecEXIF)
				}
			},
		},
		{
			name: "WithFormats single format",
			opt:  WithFormats(FormatJPEG),
			check: func(t *testing.T, cfg Config) {
				if len(cfg.Formats) != 1 {
					t.Errorf("len(Formats) = %d, want 1", len(cfg.Formats))
				}
				if len(cfg.Formats) > 0 && cfg.Formats[0] != FormatJPEG {
					t.Errorf("Formats[0] = %v, want %v", cfg.Formats[0], FormatJPEG)
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
