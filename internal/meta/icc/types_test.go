package icc

import (
	"testing"
)

func TestProfileClass_String(t *testing.T) {
	tests := []struct {
		name  string
		class ProfileClass
		want  string
	}{
		{"Input", ClassInput, "Input Device (Scanner)"},
		{"Display", ClassDisplay, "Display Device (Monitor)"},
		{"Output", ClassOutput, "Output Device (Printer)"},
		{"Link", ClassLink, "Device Link"},
		{"Abstract", ClassAbstract, "Abstract Profile"},
		{"ColorSpace", ClassColorSpace, "Color Space Conversion"},
		{"NamedColor", ClassNamedColor, "Named Color"},
		{"Unknown", ProfileClass(0x12345678), "\x124Vx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.class.String(); got != tt.want {
				t.Errorf("ProfileClass.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorSpace_String(t *testing.T) {
	tests := []struct {
		name  string
		space ColorSpace
		want  string
	}{
		{"XYZ", SpaceXYZ, "XYZ"},
		{"Lab", SpaceLab, "Lab"},
		{"Luv", SpaceLuv, "Luv"},
		{"YCbr", SpaceYCbr, "YCbCr"},
		{"Yxy", SpaceYxy, "Yxy"},
		{"RGB", SpaceRGB, "RGB"},
		{"Gray", SpaceGray, "Grayscale"},
		{"HSV", SpaceHSV, "HSV"},
		{"HLS", SpaceHLS, "HLS"},
		{"CMYK", SpaceCMYK, "CMYK"},
		{"CMY", SpaceCMY, "CMY"},
		{"2CLR", Space2CLR, "2 Color"},
		{"3CLR", Space3CLR, "3 Color"},
		{"4CLR", Space4CLR, "4 Color"},
		{"5CLR", Space5CLR, "5 Color"},
		{"6CLR", Space6CLR, "6 Color"},
		{"7CLR", Space7CLR, "7 Color"},
		{"8CLR", Space8CLR, "8 Color"},
		{"9CLR", Space9CLR, "9 Color"},
		{"ACLR", SpaceACLR, "10 Color"},
		{"BCLR", SpaceBCLR, "11 Color"},
		{"CCLR", SpaceCCLR, "12 Color"},
		{"DCLR", SpaceDCLR, "13 Color"},
		{"ECLR", SpaceECLR, "14 Color"},
		{"FCLR", SpaceFCLR, "15 Color"},
		{"Unknown", ColorSpace(0x12345678), "\x124Vx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.space.String(); got != tt.want {
				t.Errorf("ColorSpace.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlatform_String(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{"Apple", PlatformApple, "Apple"},
		{"Microsoft", PlatformMicrosoft, "Microsoft"},
		{"SGI", PlatformSGI, "Silicon Graphics"},
		{"Sun", PlatformSun, "Sun Microsystems"},
		{"Taligent", PlatformTaligent, "Taligent"},
		{"Unspecified", Platform(0), "Unspecified"},
		{"Unknown", Platform(0x12345678), "\x124Vx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.String(); got != tt.want {
				t.Errorf("Platform.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderingIntent_String(t *testing.T) {
	tests := []struct {
		name   string
		intent RenderingIntent
		want   string
	}{
		{"Perceptual", IntentPerceptual, "Perceptual"},
		{"RelativeColorimetric", IntentRelativeColorimetric, "Media-Relative Colorimetric"},
		{"Saturation", IntentSaturation, "Saturation"},
		{"AbsoluteColorimetric", IntentAbsoluteColorimetric, "ICC-Absolute Colorimetric"},
		{"Unknown", RenderingIntent(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.intent.String(); got != tt.want {
				t.Errorf("RenderingIntent.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProfileFlags(t *testing.T) {
	tests := []struct {
		name          string
		flags         ProfileFlags
		isEmbedded    bool
		isIndependent bool
	}{
		{"None", ProfileFlags(0), false, true},
		{"Embedded", ProfileFlags(0x01), true, true},
		{"NotIndependent", ProfileFlags(0x02), false, false},
		{"Both", ProfileFlags(0x03), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.IsEmbedded(); got != tt.isEmbedded {
				t.Errorf("ProfileFlags.IsEmbedded() = %v, want %v", got, tt.isEmbedded)
			}
			if got := tt.flags.IsIndependent(); got != tt.isIndependent {
				t.Errorf("ProfileFlags.IsIndependent() = %v, want %v", got, tt.isIndependent)
			}
		})
	}
}

func TestDeviceAttributes(t *testing.T) {
	tests := []struct {
		name        string
		attrs       DeviceAttributes
		isReflective bool
		isGlossy    bool
		isPositive  bool
		isColor     bool
	}{
		{"AllDefaults", DeviceAttributes(0), true, true, true, true},
		{"Transparency", DeviceAttributes(0x01), false, true, true, true},
		{"Matte", DeviceAttributes(0x02), true, false, true, true},
		{"Negative", DeviceAttributes(0x04), true, true, false, true},
		{"BlackWhite", DeviceAttributes(0x08), true, true, true, false},
		{"AllOpposite", DeviceAttributes(0x0F), false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.attrs.IsReflective(); got != tt.isReflective {
				t.Errorf("DeviceAttributes.IsReflective() = %v, want %v", got, tt.isReflective)
			}
			if got := tt.attrs.IsGlossy(); got != tt.isGlossy {
				t.Errorf("DeviceAttributes.IsGlossy() = %v, want %v", got, tt.isGlossy)
			}
			if got := tt.attrs.IsPositive(); got != tt.isPositive {
				t.Errorf("DeviceAttributes.IsPositive() = %v, want %v", got, tt.isPositive)
			}
			if got := tt.attrs.IsColor(); got != tt.isColor {
				t.Errorf("DeviceAttributes.IsColor() = %v, want %v", got, tt.isColor)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    string
	}{
		{"v2.0.0", Version{2, 0, 0}, "2.0.0"},
		{"v4.3.0", Version{4, 3, 0}, "4.3.0"},
		{"v4.4.0", Version{4, 4, 0}, "4.4.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("Version.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSignatureToString(t *testing.T) {
	tests := []struct {
		name string
		sig  uint32
		want string
	}{
		{"APPL", 0x4150504C, "APPL"},
		{"RGB ", 0x52474220, "RGB "},
		{"acsp", 0x61637370, "acsp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := signatureToString(tt.sig); got != tt.want {
				t.Errorf("signatureToString() = %q, want %q", got, tt.want)
			}
		})
	}
}


