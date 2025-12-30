package icc

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
}

func TestParser_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "ICC" {
		t.Errorf("Name() = %q, want %q", got, "ICC")
	}
}

func TestParser_Detect(t *testing.T) {
	// Build valid ICC header with 'acsp' signature at offset 36
	makeICCHeader := func() []byte {
		data := make([]byte, 128)
		binary.BigEndian.PutUint32(data[0:4], 128) // Profile size
		copy(data[36:40], []byte("acsp"))          // Signature
		return data
	}

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid ICC profile",
			data: makeICCHeader(),
			want: true,
		},
		{
			name: "invalid signature",
			data: func() []byte {
				data := make([]byte, 128)
				copy(data[36:40], []byte("xxxx"))
				return data
			}(),
			want: false,
		},
		{
			name: "too short to read signature",
			data: make([]byte, 30),
			want: false,
		},
		{
			name: "empty data",
			data: []byte{},
			want: false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			if got := p.Detect(r); got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

// buildMinimalICCProfile creates a minimal valid ICC profile for testing
func buildMinimalICCProfile() []byte {
	data := make([]byte, 256)

	// Header (128 bytes)
	binary.BigEndian.PutUint32(data[0:4], 256)         // Profile size
	copy(data[4:8], "test")                            // CMM Type
	binary.BigEndian.PutUint32(data[8:12], 0x02400000) // Version 2.4.0
	copy(data[12:16], "mntr")                          // Profile class: display
	copy(data[16:20], "RGB ")                          // Color space
	copy(data[20:24], "XYZ ")                          // PCS
	// DateTime (offset 24-36)
	binary.BigEndian.PutUint16(data[24:26], 2024) // Year
	binary.BigEndian.PutUint16(data[26:28], 1)    // Month
	binary.BigEndian.PutUint16(data[28:30], 15)   // Day
	binary.BigEndian.PutUint16(data[30:32], 12)   // Hour
	binary.BigEndian.PutUint16(data[32:34], 30)   // Minute
	binary.BigEndian.PutUint16(data[34:36], 45)   // Second
	copy(data[36:40], "acsp")                     // Signature
	copy(data[40:44], "APPL")                     // Platform
	binary.BigEndian.PutUint32(data[44:48], 0)    // Flags
	copy(data[48:52], "manu")                     // Device manufacturer
	copy(data[52:56], "modl")                     // Device model
	binary.BigEndian.PutUint64(data[56:64], 0)    // Device attributes
	binary.BigEndian.PutUint32(data[64:68], 0)    // Rendering intent
	// D50 illuminant (s15Fixed16)
	binary.BigEndian.PutUint32(data[68:72], 0x0000F6D6) // X
	binary.BigEndian.PutUint32(data[72:76], 0x00010000) // Y
	binary.BigEndian.PutUint32(data[76:80], 0x0000D32D) // Z
	copy(data[80:84], "crtr")                           // Creator
	// Profile ID (84-100) - zeros

	// Tag table at offset 128
	binary.BigEndian.PutUint32(data[128:132], 1) // Tag count: 1

	// Tag record (offset 132)
	copy(data[132:136], "desc")                    // Signature
	binary.BigEndian.PutUint32(data[136:140], 144) // Offset
	binary.BigEndian.PutUint32(data[140:144], 20)  // Size

	// Tag data at offset 144 (text type)
	copy(data[144:148], "text")                  // Type signature
	binary.BigEndian.PutUint32(data[148:152], 0) // Reserved
	copy(data[152:164], "Test Profile")          // Text data

	return data
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantDirs int
		wantErr  bool
	}{
		{
			name:     "valid minimal ICC profile",
			data:     buildMinimalICCProfile(),
			wantDirs: 2, // Header + Profile
			wantErr:  false,
		},
		{
			name:     "invalid header - wrong signature",
			data:     make([]byte, 128),
			wantDirs: 0,
			wantErr:  true,
		},
		{
			name:     "too short",
			data:     make([]byte, 50),
			wantDirs: 0,
			wantErr:  true,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			dirs, parseErr := p.Parse(r)

			hasErr := parseErr != nil && parseErr.Error() != ""
			if hasErr != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", parseErr, tt.wantErr)
			}
			if len(dirs) != tt.wantDirs {
				t.Errorf("Parse() returned %d dirs, want %d", len(dirs), tt.wantDirs)
			}
		})
	}
}

func TestParser_parseHeader(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid header",
			data:    buildMinimalICCProfile(),
			wantErr: false,
		},
		{
			name:    "too short",
			data:    make([]byte, 50),
			wantErr: true,
		},
		{
			name: "invalid signature",
			data: func() []byte {
				d := buildMinimalICCProfile()
				copy(d[36:40], "xxxx")
				return d
			}(),
			wantErr: true,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			dir, err := p.parseHeader(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && dir == nil {
				t.Error("parseHeader() returned nil directory")
			}
			if !tt.wantErr && dir.Name != "ICC-Header" {
				t.Errorf("parseHeader() dir name = %q, want 'ICC-Header'", dir.Name)
			}
		})
	}
}

func TestParser_parseTagTable(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantCount int
		wantErr   bool
	}{
		{
			name:      "valid tag table with 1 tag",
			data:      buildMinimalICCProfile(),
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "zero tags",
			data: func() []byte {
				d := buildMinimalICCProfile()
				binary.BigEndian.PutUint32(d[128:132], 0)
				return d
			}(),
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "unreasonable tag count",
			data: func() []byte {
				d := buildMinimalICCProfile()
				binary.BigEndian.PutUint32(d[128:132], 5000)
				return d
			}(),
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "too short to read tag count",
			data:      make([]byte, 130),
			wantCount: 0,
			wantErr:   true,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			tags, err := p.parseTagTable(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTagTable() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(tags) != tt.wantCount {
				t.Errorf("parseTagTable() count = %d, want %d", len(tags), tt.wantCount)
			}
		})
	}
}

func TestParser_parseTagTable_ReadRecordsError(t *testing.T) {
	// Create data with tag count but insufficient data for records
	data := make([]byte, 140)
	copy(data[36:40], "acsp")
	binary.BigEndian.PutUint32(data[128:132], 5) // 5 tags but not enough data

	p := New()
	r := bytes.NewReader(data)
	_, err := p.parseTagTable(r)
	if err == nil {
		t.Error("parseTagTable() expected error for truncated records")
	}
}

func TestParser_parseTagData(t *testing.T) {
	// Build a profile with a text tag
	profile := buildMinimalICCProfile()

	tests := []struct {
		name    string
		tag     TagRecord
		wantErr bool
	}{
		{
			name: "valid text tag",
			tag: TagRecord{
				Signature: [4]byte{'d', 'e', 's', 'c'},
				Offset:    144,
				Size:      20,
			},
			wantErr: false,
		},
		{
			name: "tag too small",
			tag: TagRecord{
				Signature: [4]byte{'d', 'e', 's', 'c'},
				Offset:    144,
				Size:      4, // < 8
			},
			wantErr: true,
		},
		{
			name: "read error",
			tag: TagRecord{
				Signature: [4]byte{'d', 'e', 's', 'c'},
				Offset:    10000, // Beyond data
				Size:      20,
			},
			wantErr: true,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(profile)
			data, err := p.parseTagData(r, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTagData() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && data == nil {
				t.Error("parseTagData() returned nil")
			}
		})
	}
}

func TestParser_Parse_WithTagParseError(t *testing.T) {
	// Build profile with a tag that will fail to parse
	profile := buildMinimalICCProfile()
	// Set tag offset to invalid location
	binary.BigEndian.PutUint32(profile[136:140], 10000)

	p := New()
	r := bytes.NewReader(profile)
	dirs, _ := p.Parse(r)

	// Should still return header directory even if tag parsing fails
	if len(dirs) < 1 {
		t.Error("Parse() should return at least header directory")
	}
}

func TestParser_Parse_TagTableError(t *testing.T) {
	// Build profile with valid header but broken tag table
	profile := buildMinimalICCProfile()
	// Set unreasonable tag count
	binary.BigEndian.PutUint32(profile[128:132], 5000)

	p := New()
	r := bytes.NewReader(profile)
	dirs, parseErr := p.Parse(r)

	// Should return header directory and error
	if len(dirs) != 1 {
		t.Errorf("Parse() should return 1 dir (header), got %d", len(dirs))
	}
	if parseErr == nil || parseErr.Error() == "" {
		t.Error("Parse() should return error for broken tag table")
	}
}

func TestParser_ImplementsInterface(t *testing.T) {
	var _ parser.Parser = (*Parser)(nil)
}

func TestParser_ConcurrentParse(t *testing.T) {
	profile := buildMinimalICCProfile()

	p := New()
	r := bytes.NewReader(profile)

	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, _ = p.Parse(r)
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestParser_parseTagData_ConverterError(t *testing.T) {
	// Build tag data where the converter will return an error
	// Use curv type with count=5 but insufficient point data
	data := make([]byte, 20)
	copy(data[0:4], "curv")                   // type signature
	binary.BigEndian.PutUint32(data[8:12], 5) // count = 5 points expected

	p := New()
	r := bytes.NewReader(data)
	tag := TagRecord{
		Signature: [4]byte{'r', 'T', 'R', 'C'},
		Offset:    0,
		Size:      22, // 8 header + 4 count + 10 bytes (not enough for 5 points = 10 bytes)
	}
	_, err := p.parseTagData(r, tag)
	// This triggers the converter returning an error path
	if err == nil {
		t.Error("expected error from converter")
	}
}

func TestParser_parseHeader_AllFields(t *testing.T) {
	profile := buildMinimalICCProfile()

	p := New()
	r := bytes.NewReader(profile)
	dir, err := p.parseHeader(r)
	if err != nil {
		t.Fatalf("parseHeader() error = %v", err)
	}

	// Verify expected tags exist
	expectedTags := []string{
		"ProfileSize", "CMMType", "ProfileVersion", "ProfileClass",
		"ColorSpace", "ProfileConnectionSpace", "DateTimeCreated",
		"ProfileSignature", "PrimaryPlatform", "ProfileFlags",
		"DeviceManufacturer", "DeviceModel", "DeviceAttributes",
		"RenderingIntent", "IlluminantX", "IlluminantY", "IlluminantZ",
		"ProfileCreator", "ProfileID",
	}

	tagMap := make(map[string]bool)
	for _, tag := range dir.Tags {
		tagMap[tag.Name] = true
	}

	for _, name := range expectedTags {
		if !tagMap[name] {
			t.Errorf("Missing tag: %s", name)
		}
	}
}
