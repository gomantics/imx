package icc

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestGetTagName(t *testing.T) {
	tests := []struct {
		sig  string
		want string
	}{
		{"desc", "ProfileDescription"},
		{"cprt", "ProfileCopyright"},
		{"wtpt", "MediaWhitePoint"},
		{"rXYZ", "RedMatrixColumn"},
		{"gXYZ", "GreenMatrixColumn"},
		{"bXYZ", "BlueMatrixColumn"},
		{"rTRC", "RedToneReproductionCurve"},
		{"gTRC", "GreenToneReproductionCurve"},
		{"bTRC", "BlueToneReproductionCurve"},
		{"chad", "ChromaticAdaptation"},
		{"meas", "Measurement"},
		{"view", "ViewingConditions"},
		{"tech", "Technology"},
		{"A2B0", "AToB0Perceptual"},
		{"B2A0", "BToA0Perceptual"},
		{"xxxx", "xxxx"}, // Unknown returns signature
	}

	for _, tt := range tests {
		t.Run(tt.sig, func(t *testing.T) {
			if got := getTagName(tt.sig); got != tt.want {
				t.Errorf("getTagName(%q) = %q, want %q", tt.sig, got, tt.want)
			}
		})
	}
}

func TestGetTypeConverter(t *testing.T) {
	// Test known converters exist and can be called
	knownTypes := []string{"text", "desc", "mluc", "XYZ ", "sf32", "uf32", "sig ", "curv", "para", "dtim", "meas", "view", "chrm"}
	for _, typ := range knownTypes {
		t.Run(typ, func(t *testing.T) {
			converter := getTypeConverter(typ)
			if converter == nil {
				t.Errorf("getTypeConverter(%q) returned nil", typ)
			}
		})
	}

	// Test that converters can be called through the map (for coverage)
	t.Run("call para converter", func(t *testing.T) {
		d := make([]byte, 12)
		copy(d[0:4], "para")
		converter := getTypeConverter("para")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 12})
	})

	t.Run("call dtim converter", func(t *testing.T) {
		d := make([]byte, 20)
		copy(d[0:4], "dtim")
		converter := getTypeConverter("dtim")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 20})
	})

	t.Run("call meas converter", func(t *testing.T) {
		d := make([]byte, 44)
		copy(d[0:4], "meas")
		converter := getTypeConverter("meas")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 44})
	})

	t.Run("call view converter", func(t *testing.T) {
		d := make([]byte, 36)
		copy(d[0:4], "view")
		converter := getTypeConverter("view")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 36})
	})

	t.Run("call chrm converter", func(t *testing.T) {
		d := make([]byte, 12)
		copy(d[0:4], "chrm")
		converter := getTypeConverter("chrm")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 12})
	})

	// Test remaining converters through the map
	t.Run("call desc converter", func(t *testing.T) {
		d := make([]byte, 16)
		copy(d[0:4], "desc")
		binary.BigEndian.PutUint32(d[8:12], 0) // count = 0
		converter := getTypeConverter("desc")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 16})
	})

	t.Run("call mluc converter", func(t *testing.T) {
		d := make([]byte, 16)
		copy(d[0:4], "mluc")
		binary.BigEndian.PutUint32(d[8:12], 0) // numRecords = 0
		converter := getTypeConverter("mluc")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 16})
	})

	t.Run("call XYZ converter", func(t *testing.T) {
		d := make([]byte, 20)
		copy(d[0:4], "XYZ ")
		converter := getTypeConverter("XYZ ")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 20})
	})

	t.Run("call sf32 converter", func(t *testing.T) {
		d := make([]byte, 12)
		copy(d[0:4], "sf32")
		converter := getTypeConverter("sf32")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 12})
	})

	t.Run("call uf32 converter", func(t *testing.T) {
		d := make([]byte, 12)
		copy(d[0:4], "uf32")
		converter := getTypeConverter("uf32")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 12})
	})

	t.Run("call sig converter", func(t *testing.T) {
		d := make([]byte, 12)
		copy(d[0:4], "sig ")
		converter := getTypeConverter("sig ")
		_, _ = converter(bytes.NewReader(d), TagRecord{Offset: 0, Size: 12})
	})

	// Test unknown type returns default converter
	t.Run("unknown type", func(t *testing.T) {
		converter := getTypeConverter("unkn")
		if converter == nil {
			t.Fatal("getTypeConverter('unkn') returned nil")
		}

		// Test default converter returns raw bytes
		data := []byte("test data here")
		r := bytes.NewReader(data)
		tag := TagRecord{Offset: 0, Size: uint32(len(data))}
		result, err := converter(r, tag)
		if err != nil {
			t.Errorf("default converter error = %v", err)
		}
		if b, ok := result.([]byte); !ok || string(b) != "test data here" {
			t.Errorf("default converter result = %v", result)
		}
	})

	// Test default converter read error
	t.Run("unknown type read error", func(t *testing.T) {
		converter := getTypeConverter("unkn")
		r := bytes.NewReader(make([]byte, 5))
		tag := TagRecord{Offset: 0, Size: 20}
		_, err := converter(r, tag)
		if err == nil {
			t.Error("expected error for read failure")
		}
	})
}

func TestGetProfileClassName(t *testing.T) {
	tests := []struct {
		sig  string
		want string
	}{
		{"scnr", "Input Device Profile"},
		{"mntr", "Display Device Profile"},
		{"prtr", "Output Device Profile"},
		{"link", "DeviceLink Profile"},
		{"abst", "Abstract Profile"},
		{"spac", "ColorSpace Conversion Profile"},
		{"nmcl", "Named Color Profile"},
		{"unkn", "unkn"}, // Unknown returns signature
	}

	for _, tt := range tests {
		t.Run(tt.sig, func(t *testing.T) {
			if got := getProfileClassName(tt.sig); got != tt.want {
				t.Errorf("getProfileClassName(%q) = %q, want %q", tt.sig, got, tt.want)
			}
		})
	}
}

func TestGetColorSpaceName(t *testing.T) {
	tests := []struct {
		sig  string
		want string
	}{
		{"XYZ ", "XYZ"},
		{"Lab ", "Lab"},
		{"Luv ", "Luv"},
		{"YCbr", "YCbCr"},
		{"Yxy ", "Yxy"},
		{"RGB ", "RGB"},
		{"GRAY", "Grayscale"},
		{"HSV ", "HSV"},
		{"HLS ", "HLS"},
		{"CMYK", "CMYK"},
		{"CMY ", "CMY"},
		{"2CLR", "2Color"},
		{"3CLR", "3Color"},
		{"4CLR", "4Color"},
		{"5CLR", "5Color"},
		{"6CLR", "6Color"},
		{"7CLR", "7Color"},
		{"8CLR", "8Color"},
		{"9CLR", "9Color"},
		{"ACLR", "10Color"},
		{"BCLR", "11Color"},
		{"CCLR", "12Color"},
		{"DCLR", "13Color"},
		{"ECLR", "14Color"},
		{"FCLR", "15Color"},
		{"unkn", "unkn"}, // Unknown returns signature
	}

	for _, tt := range tests {
		t.Run(tt.sig, func(t *testing.T) {
			if got := getColorSpaceName(tt.sig); got != tt.want {
				t.Errorf("getColorSpaceName(%q) = %q, want %q", tt.sig, got, tt.want)
			}
		})
	}
}

func TestGetPlatformName(t *testing.T) {
	tests := []struct {
		sig  string
		want string
	}{
		{"APPL", "Apple"},
		{"MSFT", "Microsoft"},
		{"SGI ", "SiliconGraphics"},
		{"SUNW", "SunMicrosystems"},
		{"TGNT", "Taligent"},
		{"\x00\x00\x00\x00", "Unspecified"},
		{"unkn", "unkn"}, // Unknown returns signature
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := getPlatformName(tt.sig); got != tt.want {
				t.Errorf("getPlatformName(%q) = %q, want %q", tt.sig, got, tt.want)
			}
		})
	}
}

func TestGetRenderingIntentName(t *testing.T) {
	tests := []struct {
		intent uint32
		want   string
	}{
		{0, "Perceptual"},
		{1, "MediaRelativeColorimetric"},
		{2, "Saturation"},
		{3, "ICCAbsoluteColorimetric"},
		{99, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := getRenderingIntentName(tt.intent); got != tt.want {
				t.Errorf("getRenderingIntentName(%d) = %q, want %q", tt.intent, got, tt.want)
			}
		})
	}
}

func TestGetTechnologyName(t *testing.T) {
	tests := []struct {
		sig  uint32
		want string
	}{
		{0x66736E20, "FilmScanner"},
		{0x64636D20, "DigitalCamera"},
		{0x7273636E, "ReflectiveScanner"},
		{0x696A6574, "InkJetPrinter"},
		{0x74776178, "ThermalWaxPrinter"},
		{0x65706879, "ElectrophotographicPrinter"},
		{0x65737461, "ElectrostaticPrinter"},
		{0x64737562, "DyeSublimationPrinter"},
		{0x7270686F, "PhotographicPaperPrinter"},
		{0x6670726E, "FilmWriter"},
		{0x7669646C, "VideoMonitor"},
		{0x76696463, "VideoCamera"},
		{0x706A7476, "ProjectionTelevision"},
		{0x43525420, "CathodeRayTubeDisplay"},
		{0x504D4420, "PassiveMatrixDisplay"},
		{0x414D4420, "ActiveMatrixDisplay"},
		{0x4C434420, "LCDDisplay"},
		{0x4F4C4544, "OLEDDisplay"},
		{0x4C454420, "LEDDisplay"},
		{0x6770686F, "Gravure"},
		{0x6F666673, "OffsetLithography"},
		{0x73696C6B, "Silkscreen"},
		{0x666C6578, "Flexography"},
		{0x6D706673, "MotionPictureFilmScanner"},
		{0x6D706672, "MotionPictureFilmRecorder"},
		{0x646D7063, "DigitalMotionPictureCamera"},
		{0x64637067, "DigitalCinemaProjector"},
		{0x74657374, "test"}, // Unknown returns signature as string
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := getTechnologyName(tt.sig); got != tt.want {
				t.Errorf("getTechnologyName(0x%08X) = %q, want %q", tt.sig, got, tt.want)
			}
		})
	}
}

func TestGetProfileFlagsName(t *testing.T) {
	tests := []struct {
		flags uint32
		want  string
	}{
		{0x00, "Not Embedded, Independent"},
		{0x01, "Embedded, Independent"},
		{0x02, "Not Embedded, Cannot be used independently"},
		{0x03, "Embedded, Cannot be used independently"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := getProfileFlagsName(tt.flags); got != tt.want {
				t.Errorf("getProfileFlagsName(0x%02X) = %q, want %q", tt.flags, got, tt.want)
			}
		})
	}
}

func TestGetDeviceAttributesName(t *testing.T) {
	tests := []struct {
		attrs uint64
		want  string
	}{
		{0x00, "Reflective, Glossy, Positive, Color"},
		{0x01, "Transmissive, Glossy, Positive, Color"},
		{0x02, "Reflective, Matte, Positive, Color"},
		{0x03, "Transmissive, Matte, Positive, Color"},
		{0x04, "Reflective, Glossy, Negative, Color"},
		{0x08, "Reflective, Glossy, Positive, Black & White"},
		{0x0F, "Transmissive, Matte, Negative, Black & White"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := getDeviceAttributesName(tt.attrs); got != tt.want {
				t.Errorf("getDeviceAttributesName(0x%02X) = %q, want %q", tt.attrs, got, tt.want)
			}
		})
	}
}

func TestTagNamesMapNotEmpty(t *testing.T) {
	if len(tagNames) == 0 {
		t.Error("tagNames map is empty")
	}
}

func TestTypeConvertersMapNotEmpty(t *testing.T) {
	if len(typeConverters) == 0 {
		t.Error("typeConverters map is empty")
	}
}

func TestProfileClassNamesMapNotEmpty(t *testing.T) {
	if len(profileClassNames) == 0 {
		t.Error("profileClassNames map is empty")
	}
}

func TestColorSpaceNamesMapNotEmpty(t *testing.T) {
	if len(colorSpaceNames) == 0 {
		t.Error("colorSpaceNames map is empty")
	}
}

func TestPlatformNamesMapNotEmpty(t *testing.T) {
	if len(platformNames) == 0 {
		t.Error("platformNames map is empty")
	}
}

func TestRenderingIntentNamesMapNotEmpty(t *testing.T) {
	if len(renderingIntentNames) == 0 {
		t.Error("renderingIntentNames map is empty")
	}
}

func TestTechnologySignaturesMapNotEmpty(t *testing.T) {
	if len(technologySignatures) == 0 {
		t.Error("technologySignatures map is empty")
	}
}
