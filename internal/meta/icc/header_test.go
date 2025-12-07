package icc

import (
	"encoding/binary"
	"testing"
	"time"
)

// buildValidHeader creates a valid 128-byte ICC profile header
func buildValidHeader() []byte {
	data := make([]byte, 128)

	// Profile size (bytes 0-3)
	binary.BigEndian.PutUint32(data[0:4], 596)

	// Preferred CMM type (bytes 4-7)
	copy(data[4:8], "APPL")

	// Profile version (bytes 8-11) - v4.3.0
	data[8] = 4
	data[9] = 0x30 // 3 << 4

	// Profile class (bytes 12-15) - Display
	binary.BigEndian.PutUint32(data[12:16], uint32(ClassDisplay))

	// Data color space (bytes 16-19) - RGB
	binary.BigEndian.PutUint32(data[16:20], uint32(SpaceRGB))

	// PCS (bytes 20-23) - XYZ
	binary.BigEndian.PutUint32(data[20:24], uint32(SpaceXYZ))

	// Creation date/time (bytes 24-35)
	binary.BigEndian.PutUint16(data[24:26], 2023) // year
	binary.BigEndian.PutUint16(data[26:28], 3)    // month
	binary.BigEndian.PutUint16(data[28:30], 9)    // day
	binary.BigEndian.PutUint16(data[30:32], 10)   // hour
	binary.BigEndian.PutUint16(data[32:34], 57)   // minute
	binary.BigEndian.PutUint16(data[34:36], 0)    // second

	// Profile signature (bytes 36-39) - 'acsp'
	binary.BigEndian.PutUint32(data[36:40], ICCSignature)

	// Platform (bytes 40-43)
	binary.BigEndian.PutUint32(data[40:44], uint32(PlatformApple))

	// Flags (bytes 44-47)
	binary.BigEndian.PutUint32(data[44:48], 0)

	// Device manufacturer (bytes 48-51)
	copy(data[48:52], "GOOG")

	// Device model (bytes 52-55)
	copy(data[52:56], "test")

	// Device attributes (bytes 56-63)
	binary.BigEndian.PutUint64(data[56:64], 0)

	// Rendering intent (bytes 64-67)
	binary.BigEndian.PutUint32(data[64:68], uint32(IntentPerceptual))

	// PCS illuminant (bytes 68-79) - D50
	binary.BigEndian.PutUint32(data[68:72], 0x0000F6D6) // X = 0.9642
	binary.BigEndian.PutUint32(data[72:76], 0x00010000) // Y = 1.0
	binary.BigEndian.PutUint32(data[76:80], 0x0000D32D) // Z = 0.8249

	// Profile creator (bytes 80-83)
	copy(data[80:84], "GOOG")

	// Profile ID (bytes 84-99) - non-zero MD5
	for i := 84; i < 100; i++ {
		data[i] = byte(i - 84 + 1)
	}

	return data
}

func TestParseHeader(t *testing.T) {
	data := buildValidHeader()

	h, err := parseHeader(data)
	if err != nil {
		t.Fatalf("parseHeader() error = %v", err)
	}

	// Verify fields
	if h.ProfileSize != 596 {
		t.Errorf("ProfileSize = %d, want 596", h.ProfileSize)
	}
	if h.PreferredCMM != "APPL" {
		t.Errorf("PreferredCMM = %q, want %q", h.PreferredCMM, "APPL")
	}
	if h.Version.Major != 4 || h.Version.Minor != 3 || h.Version.BugFix != 0 {
		t.Errorf("Version = %v, want 4.3.0", h.Version)
	}
	if h.ProfileClass != ClassDisplay {
		t.Errorf("ProfileClass = %v, want ClassDisplay", h.ProfileClass)
	}
	if h.DataColorSpace != SpaceRGB {
		t.Errorf("DataColorSpace = %v, want SpaceRGB", h.DataColorSpace)
	}
	if h.PCS != SpaceXYZ {
		t.Errorf("PCS = %v, want SpaceXYZ", h.PCS)
	}
	if h.Platform != PlatformApple {
		t.Errorf("Platform = %v, want PlatformApple", h.Platform)
	}
	if h.RenderingIntent != IntentPerceptual {
		t.Errorf("RenderingIntent = %v, want IntentPerceptual", h.RenderingIntent)
	}
	if h.DeviceManufacturer != "GOOG" {
		t.Errorf("DeviceManufacturer = %q, want %q", h.DeviceManufacturer, "GOOG")
	}
	if h.Creator != "GOOG" {
		t.Errorf("Creator = %q, want %q", h.Creator, "GOOG")
	}

	// Check date
	wantDate := time.Date(2023, 3, 9, 10, 57, 0, 0, time.UTC)
	if !h.Created.Equal(wantDate) {
		t.Errorf("Created = %v, want %v", h.Created, wantDate)
	}
}

func TestParseHeader_TooShort(t *testing.T) {
	data := make([]byte, 64) // Less than 128 bytes
	_, err := parseHeader(data)
	if err == nil {
		t.Error("parseHeader() expected error for short data")
	}
}

func TestParseHeader_InvalidSignature(t *testing.T) {
	data := buildValidHeader()
	// Corrupt the 'acsp' signature
	copy(data[36:40], "xxxx")

	_, err := parseHeader(data)
	if err == nil {
		t.Error("parseHeader() expected error for invalid signature")
	}
}

func TestParseDateTimeNumber(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want time.Time
	}{
		{
			name: "valid date",
			data: func() []byte {
				d := make([]byte, 12)
				binary.BigEndian.PutUint16(d[0:2], 2023)
				binary.BigEndian.PutUint16(d[2:4], 3)
				binary.BigEndian.PutUint16(d[4:6], 9)
				binary.BigEndian.PutUint16(d[6:8], 10)
				binary.BigEndian.PutUint16(d[8:10], 57)
				binary.BigEndian.PutUint16(d[10:12], 30)
				return d
			}(),
			want: time.Date(2023, 3, 9, 10, 57, 30, 0, time.UTC),
		},
		{
			name: "short data",
			data: make([]byte, 6),
			want: time.Time{},
		},
		{
			name: "invalid year",
			data: make([]byte, 12), // all zeros
			want: time.Time{},
		},
		{
			name: "invalid month",
			data: func() []byte {
				d := make([]byte, 12)
				binary.BigEndian.PutUint16(d[0:2], 2023)
				binary.BigEndian.PutUint16(d[2:4], 13) // invalid month
				return d
			}(),
			want: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDateTimeNumber(tt.data)
			if !got.Equal(tt.want) {
				t.Errorf("parseDateTimeNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseXYZNumber(t *testing.T) {
	data := make([]byte, 12)
	// X = 1.0 (0x00010000 in s15Fixed16)
	binary.BigEndian.PutUint32(data[0:4], 0x00010000)
	// Y = 0.5 (0x00008000 in s15Fixed16)
	binary.BigEndian.PutUint32(data[4:8], 0x00008000)
	// Z = -0.5 (0xFFFF8000 in s15Fixed16)
	binary.BigEndian.PutUint32(data[8:12], 0xFFFF8000)

	xyz := parseXYZNumber(data)

	if xyz.X != 1.0 {
		t.Errorf("X = %f, want 1.0", xyz.X)
	}
	if xyz.Y != 0.5 {
		t.Errorf("Y = %f, want 0.5", xyz.Y)
	}
	if xyz.Z != -0.5 {
		t.Errorf("Z = %f, want -0.5", xyz.Z)
	}
}

func TestParseXYZNumber_Short(t *testing.T) {
	data := make([]byte, 8) // Less than 12 bytes
	xyz := parseXYZNumber(data)
	if xyz.X != 0 || xyz.Y != 0 || xyz.Z != 0 {
		t.Error("parseXYZNumber() should return zero XYZ for short data")
	}
}

func TestParseS15Fixed16(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want float64
	}{
		{
			name: "1.0",
			data: func() []byte {
				d := make([]byte, 4)
				binary.BigEndian.PutUint32(d, 0x00010000)
				return d
			}(),
			want: 1.0,
		},
		{
			name: "0.5",
			data: func() []byte {
				d := make([]byte, 4)
				binary.BigEndian.PutUint32(d, 0x00008000)
				return d
			}(),
			want: 0.5,
		},
		{
			name: "-1.0",
			data: func() []byte {
				d := make([]byte, 4)
				binary.BigEndian.PutUint32(d, 0xFFFF0000)
				return d
			}(),
			want: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseS15Fixed16(tt.data)
			if got != tt.want {
				t.Errorf("parseS15Fixed16() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestParseU16Fixed16(t *testing.T) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, 0x00020000) // 2.0

	got := parseU16Fixed16(data)
	if got != 2.0 {
		t.Errorf("parseU16Fixed16() = %f, want 2.0", got)
	}
}

func TestParseU8Fixed8(t *testing.T) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, 0x0180) // 1.5 (1 + 128/256)

	got := parseU8Fixed8(data)
	if got != 1.5 {
		t.Errorf("parseU8Fixed8() = %f, want 1.5", got)
	}
}
