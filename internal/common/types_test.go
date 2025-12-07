package common

import "testing"

func TestSpec_String(t *testing.T) {
	tests := []struct {
		name string
		spec Spec
		want string
	}{
		{
			name: "SpecEXIF returns exif",
			spec: SpecEXIF,
			want: "exif",
		},
		{
			name: "SpecIPTC returns iptc",
			spec: SpecIPTC,
			want: "iptc",
		},
		{
			name: "SpecXMP returns xmp",
			spec: SpecXMP,
			want: "xmp",
		},
		{
			name: "SpecICC returns icc",
			spec: SpecICC,
			want: "icc",
		},
		{
			name: "unknown spec returns unknown",
			spec: Spec(999),
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.String()
			if got != tt.want {
				t.Errorf("Spec.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		want   string
	}{
		{
			name:   "FormatJPEG returns jpeg",
			format: FormatJPEG,
			want:   "jpeg",
		},
		{
			name:   "FormatPNG returns png",
			format: FormatPNG,
			want:   "png",
		},
		{
			name:   "FormatWebP returns webp",
			format: FormatWebP,
			want:   "webp",
		},
		{
			name:   "FormatTIFF returns tiff",
			format: FormatTIFF,
			want:   "tiff",
		},
		{
			name:   "FormatHEIF returns heif",
			format: FormatHEIF,
			want:   "heif",
		},
		{
			name:   "unknown format returns unknown",
			format: Format(999),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.String()
			if got != tt.want {
				t.Errorf("Format.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDirectory(t *testing.T) {
	// Test Directory struct creation and fields
	dir := Directory{
		Spec: SpecEXIF,
		Name: "IFD0",
		Tags: map[TagID]Tag{
			"EXIF:Make": {
				Spec:     SpecEXIF,
				ID:       "EXIF:Make",
				Name:     "Make",
				DataType: "string",
				Value:    "Canon",
				Raw:      []byte{0x43, 0x61, 0x6E, 0x6F, 0x6E},
			},
		},
	}

	if dir.Spec != SpecEXIF {
		t.Errorf("Directory.Spec = %v, want %v", dir.Spec, SpecEXIF)
	}
	if dir.Name != "IFD0" {
		t.Errorf("Directory.Name = %q, want %q", dir.Name, "IFD0")
	}
	if len(dir.Tags) != 1 {
		t.Errorf("len(Directory.Tags) = %d, want 1", len(dir.Tags))
	}

	tag, ok := dir.Tags["EXIF:Make"]
	if !ok {
		t.Fatal("Directory.Tags[\"EXIF:Make\"] not found")
	}
	if tag.Value != "Canon" {
		t.Errorf("Tag.Value = %v, want %q", tag.Value, "Canon")
	}
}

func TestTag(t *testing.T) {
	// Test Tag struct creation and fields
	tag := Tag{
		Spec:     SpecEXIF,
		ID:       "EXIF:ISO",
		Name:     "ISO",
		DataType: "short",
		Value:    100,
		Raw:      []byte{0x64, 0x00},
	}

	if tag.Spec != SpecEXIF {
		t.Errorf("Tag.Spec = %v, want %v", tag.Spec, SpecEXIF)
	}
	if tag.ID != "EXIF:ISO" {
		t.Errorf("Tag.ID = %q, want %q", tag.ID, "EXIF:ISO")
	}
	if tag.Name != "ISO" {
		t.Errorf("Tag.Name = %q, want %q", tag.Name, "ISO")
	}
	if tag.DataType != "short" {
		t.Errorf("Tag.DataType = %q, want %q", tag.DataType, "short")
	}
	if tag.Value != 100 {
		t.Errorf("Tag.Value = %v, want %d", tag.Value, 100)
	}
	if len(tag.Raw) != 2 {
		t.Errorf("len(Tag.Raw) = %d, want 2", len(tag.Raw))
	}
}

func TestTagID(t *testing.T) {
	// Test TagID type
	var id TagID = "EXIF:DateTimeOriginal"
	if id != "EXIF:DateTimeOriginal" {
		t.Errorf("TagID = %q, want %q", id, "EXIF:DateTimeOriginal")
	}

	// Test as map key
	m := make(map[TagID]string)
	m["EXIF:Make"] = "Canon"
	m["EXIF:Model"] = "EOS 5D"

	if m["EXIF:Make"] != "Canon" {
		t.Errorf("map[TagID] lookup failed")
	}
}

func TestTagID_Spec(t *testing.T) {
	tests := []struct {
		id   TagID
		want string
	}{
		{"EXIF:Make", "EXIF"},
		{"XMP-dc:Title", "XMP"},
		{"IPTC:Byline", "IPTC"},
		{"ICC:ProfileDescription", "ICC"},
		{"Invalid", ""},
		{"NoColon", ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			got := tt.id.Spec()
			if got != tt.want {
				t.Errorf("Spec() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTagID_Name(t *testing.T) {
	tests := []struct {
		id   TagID
		want string
	}{
		{"EXIF:Make", "Make"},
		{"XMP-dc:Title", "Title"},
		{"IPTC:Byline", "Byline"},
		{"ICC:ProfileDescription", "ProfileDescription"},
		{"NoColon", "NoColon"},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			got := tt.id.Name()
			if got != tt.want {
				t.Errorf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTagID_Namespace(t *testing.T) {
	tests := []struct {
		id   TagID
		want string
	}{
		{"XMP-dc:Title", "dc"},
		{"XMP-photoshop:Credit", "photoshop"},
		{"XMP-xmp:CreateDate", "xmp"},
		{"EXIF:Make", ""},
		{"IPTC:Byline", ""},
		{"ICC:ProfileDescription", ""},
		{"NoColon", ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			got := tt.id.Namespace()
			if got != tt.want {
				t.Errorf("Namespace() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTagID_IsValid(t *testing.T) {
	tests := []struct {
		id    TagID
		valid bool
	}{
		{"EXIF:Make", true},
		{"XMP-dc:Title", true},
		{"ICC:ProfileDescription", true},
		{"IPTC:Byline", true},
		{"EXIF:Profile Description", false}, // Has space
		{"exif:Make", false},                // Lowercase spec
		{"NoColon", false},                  // No colon
		{"EXIF:", false},                    // Empty name
		{"", false},                         // Empty string
		{"XMP-DC:Title", true},              // Namespace can be uppercase
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			got := tt.id.IsValid()
			if got != tt.valid {
				t.Errorf("%q.IsValid() = %v, want %v", tt.id, got, tt.valid)
			}
		})
	}
}
