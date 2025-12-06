package imx

import (
	"testing"
	"time"

	"github.com/gomantics/imx/internal/meta"
)

func TestMetadata_Directory(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		spec     Spec
		dirName  string
		wantOk   bool
	}{
		{
			name: "find existing directory",
			metadata: Metadata{
				Directories: []Directory{
					{Spec: SpecEXIF, Name: "IFD0"},
					{Spec: SpecEXIF, Name: "ExifIFD"},
				},
			},
			spec:    SpecEXIF,
			dirName: "IFD0",
			wantOk:  true,
		},
		{
			name: "directory not found - wrong name",
			metadata: Metadata{
				Directories: []Directory{
					{Spec: SpecEXIF, Name: "IFD0"},
				},
			},
			spec:    SpecEXIF,
			dirName: "GPS",
			wantOk:  false,
		},
		{
			name: "directory not found - wrong spec",
			metadata: Metadata{
				Directories: []Directory{
					{Spec: SpecEXIF, Name: "IFD0"},
				},
			},
			spec:    SpecXMP,
			dirName: "IFD0",
			wantOk:  false,
		},
		{
			name:     "empty directories",
			metadata: Metadata{},
			spec:     SpecEXIF,
			dirName:  "IFD0",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, ok := tt.metadata.Directory(tt.spec, tt.dirName)
			if ok != tt.wantOk {
				t.Errorf("Directory() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk && dir.Name != tt.dirName {
				t.Errorf("Directory().Name = %q, want %q", dir.Name, tt.dirName)
			}
		})
	}
}

func TestMetadata_Tag(t *testing.T) {
	makeTag := func(id TagID, value any) Tag {
		return Tag{
			Spec:  SpecEXIF,
			ID:    id,
			Name:  string(id),
			Value: value,
		}
	}

	metadata := Metadata{
		Directories: []Directory{
			{
				Spec: SpecEXIF,
				Name: "IFD0",
				Tags: map[TagID]Tag{
					"Exif:Make":  makeTag("Exif:Make", "Canon"),
					"Exif:Model": makeTag("Exif:Model", "EOS 5D"),
				},
			},
		},
	}

	tests := []struct {
		name   string
		spec   Spec
		id     TagID
		wantOk bool
		want   any
	}{
		{
			name:   "find existing tag",
			spec:   SpecEXIF,
			id:     "Exif:Make",
			wantOk: true,
			want:   "Canon",
		},
		{
			name:   "tag not found",
			spec:   SpecEXIF,
			id:     "Exif:ISO",
			wantOk: false,
		},
		{
			name:   "wrong spec",
			spec:   SpecXMP,
			id:     "Exif:Make",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, ok := metadata.Tag(tt.spec, tt.id)
			if ok != tt.wantOk {
				t.Errorf("Tag() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk && tag.Value != tt.want {
				t.Errorf("Tag().Value = %v, want %v", tag.Value, tt.want)
			}
		})
	}
}

func TestMetadata_Tag_WithIndex(t *testing.T) {
	metadata := Metadata{
		Directories: []Directory{
			{
				Spec: SpecEXIF,
				Name: "IFD0",
				Tags: map[TagID]Tag{
					"Exif:Make": {Spec: SpecEXIF, ID: "Exif:Make", Value: "Canon"},
				},
			},
		},
	}
	metadata.BuildIndex()

	// Test with index
	tag, ok := metadata.Tag(SpecEXIF, "Exif:Make")
	if !ok {
		t.Error("Tag() with index should find tag")
	}
	if tag.Value != "Canon" {
		t.Errorf("Tag().Value = %v, want %q", tag.Value, "Canon")
	}

	// Test not found with different spec
	_, ok = metadata.Tag(SpecXMP, "Exif:Make")
	if ok {
		t.Error("Tag() should not find tag with wrong spec even with index")
	}
}

func TestMetadata_GetAll(t *testing.T) {
	metadata := Metadata{
		Directories: []Directory{
			{
				Spec: SpecEXIF,
				Name: "IFD0",
				Tags: map[TagID]Tag{
					"Exif:Make":  {Spec: SpecEXIF, ID: "Exif:Make", Value: "Canon"},
					"Exif:Model": {Spec: SpecEXIF, ID: "Exif:Model", Value: "EOS 5D"},
					"Exif:ISO":   {Spec: SpecEXIF, ID: "Exif:ISO", Value: 100},
				},
			},
		},
	}

	tests := []struct {
		name    string
		useIdx  bool
		ids     []TagID
		wantLen int
	}{
		{
			name:    "get multiple existing tags",
			useIdx:  false,
			ids:     []TagID{"Exif:Make", "Exif:Model"},
			wantLen: 2,
		},
		{
			name:    "get with some missing",
			useIdx:  false,
			ids:     []TagID{"Exif:Make", "Exif:Unknown"},
			wantLen: 1,
		},
		{
			name:    "get all missing",
			useIdx:  false,
			ids:     []TagID{"Unknown:A", "Unknown:B"},
			wantLen: 0,
		},
		{
			name:    "get with index",
			useIdx:  true,
			ids:     []TagID{"Exif:Make", "Exif:ISO"},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := metadata
			if tt.useIdx {
				m.BuildIndex()
			}
			result := m.GetAll(tt.ids...)
			if len(result) != tt.wantLen {
				t.Errorf("GetAll() returned %d items, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestMetadata_Each(t *testing.T) {
	metadata := Metadata{
		Directories: []Directory{
			{
				Spec: SpecEXIF,
				Name: "IFD0",
				Tags: map[TagID]Tag{
					"Exif:Make":  {ID: "Exif:Make", Value: "Canon"},
					"Exif:Model": {ID: "Exif:Model", Value: "EOS 5D"},
				},
			},
			{
				Spec: SpecEXIF,
				Name: "ExifIFD",
				Tags: map[TagID]Tag{
					"Exif:ISO": {ID: "Exif:ISO", Value: 100},
				},
			},
		},
	}

	t.Run("iterate all", func(t *testing.T) {
		count := 0
		metadata.Each(func(dir Directory, tag Tag) bool {
			count++
			return true
		})
		if count != 3 {
			t.Errorf("Each() iterated %d times, want 3", count)
		}
	})

	t.Run("early termination", func(t *testing.T) {
		count := 0
		metadata.Each(func(dir Directory, tag Tag) bool {
			count++
			return count < 2
		})
		if count != 2 {
			t.Errorf("Each() iterated %d times after early termination, want 2", count)
		}
	})
}

func TestMetadata_EachInSpec(t *testing.T) {
	metadata := Metadata{
		Directories: []Directory{
			{
				Spec: SpecEXIF,
				Name: "IFD0",
				Tags: map[TagID]Tag{
					"Exif:Make":  {Spec: SpecEXIF, ID: "Exif:Make", Value: "Canon"},
					"Exif:Model": {Spec: SpecEXIF, ID: "Exif:Model", Value: "EOS 5D"},
				},
			},
			{
				Spec: SpecXMP,
				Name: "XMP",
				Tags: map[TagID]Tag{
					"XMP:Title": {Spec: SpecXMP, ID: "XMP:Title", Value: "Test"},
				},
			},
		},
	}

	t.Run("iterate EXIF only", func(t *testing.T) {
		count := 0
		metadata.EachInSpec(SpecEXIF, func(tag Tag) bool {
			count++
			return true
		})
		if count != 2 {
			t.Errorf("EachInSpec() iterated %d times, want 2", count)
		}
	})

	t.Run("iterate XMP only", func(t *testing.T) {
		count := 0
		metadata.EachInSpec(SpecXMP, func(tag Tag) bool {
			count++
			return true
		})
		if count != 1 {
			t.Errorf("EachInSpec() iterated %d times, want 1", count)
		}
	})

	t.Run("early termination", func(t *testing.T) {
		count := 0
		metadata.EachInSpec(SpecEXIF, func(tag Tag) bool {
			count++
			return false
		})
		if count != 1 {
			t.Errorf("EachInSpec() iterated %d times after early termination, want 1", count)
		}
	})
}

func TestMetadata_DateTimeOriginal(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		metadata Metadata
		wantZero bool
	}{
		{
			name: "valid datetime",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "ExifIFD",
						Tags: map[TagID]Tag{
							"Exif:DateTimeOriginal": {Spec: SpecEXIF, ID: "Exif:DateTimeOriginal", Value: now},
						},
					},
				},
			},
			wantZero: false,
		},
		{
			name:     "missing tag",
			metadata: Metadata{},
			wantZero: true,
		},
		{
			name: "wrong type",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "ExifIFD",
						Tags: map[TagID]Tag{
							"Exif:DateTimeOriginal": {Spec: SpecEXIF, ID: "Exif:DateTimeOriginal", Value: "2024:01:01 12:00:00"},
						},
					},
				},
			},
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metadata.DateTimeOriginal()
			if tt.wantZero && !result.IsZero() {
				t.Errorf("DateTimeOriginal() = %v, want zero time", result)
			}
			if !tt.wantZero && result.IsZero() {
				t.Error("DateTimeOriginal() returned zero time, want non-zero")
			}
		})
	}
}

func TestMetadata_Orientation(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		want     int
	}{
		{
			name: "valid orientation",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "IFD0",
						Tags: map[TagID]Tag{
							"Exif:Orientation": {Spec: SpecEXIF, ID: "Exif:Orientation", Value: 6},
						},
					},
				},
			},
			want: 6,
		},
		{
			name:     "missing tag",
			metadata: Metadata{},
			want:     0,
		},
		{
			name: "wrong type",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "IFD0",
						Tags: map[TagID]Tag{
							"Exif:Orientation": {Spec: SpecEXIF, ID: "Exif:Orientation", Value: "Rotate 90 CW"},
						},
					},
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metadata.Orientation()
			if result != tt.want {
				t.Errorf("Orientation() = %d, want %d", result, tt.want)
			}
		})
	}
}

func TestMetadata_Make(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		want     string
	}{
		{
			name: "valid make",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "IFD0",
						Tags: map[TagID]Tag{
							"Exif:Make": {Spec: SpecEXIF, ID: "Exif:Make", Value: "Canon"},
						},
					},
				},
			},
			want: "Canon",
		},
		{
			name:     "missing tag",
			metadata: Metadata{},
			want:     "",
		},
		{
			name: "wrong type",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "IFD0",
						Tags: map[TagID]Tag{
							"Exif:Make": {Spec: SpecEXIF, ID: "Exif:Make", Value: 12345},
						},
					},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metadata.Make()
			if result != tt.want {
				t.Errorf("Make() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestMetadata_Model(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		want     string
	}{
		{
			name: "valid model",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "IFD0",
						Tags: map[TagID]Tag{
							"Exif:Model": {Spec: SpecEXIF, ID: "Exif:Model", Value: "EOS 5D Mark IV"},
						},
					},
				},
			},
			want: "EOS 5D Mark IV",
		},
		{
			name:     "missing tag",
			metadata: Metadata{},
			want:     "",
		},
		{
			name: "wrong type",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "IFD0",
						Tags: map[TagID]Tag{
							"Exif:Model": {Spec: SpecEXIF, ID: "Exif:Model", Value: []byte{0x01}},
						},
					},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metadata.Model()
			if result != tt.want {
				t.Errorf("Model() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestMetadata_GPSCoordinates(t *testing.T) {
	gps := &GPSCoord{Lat: 37.7749, Lon: -122.4194, Altitude: 10.5}

	tests := []struct {
		name     string
		metadata Metadata
		wantNil  bool
	}{
		{
			name: "valid GPS",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "GPS",
						Tags: map[TagID]Tag{
							"Exif:GPSCoordinates": {Spec: SpecEXIF, ID: "Exif:GPSCoordinates", Value: gps},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			name:     "missing tag",
			metadata: Metadata{},
			wantNil:  true,
		},
		{
			name: "wrong type",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "GPS",
						Tags: map[TagID]Tag{
							"Exif:GPSCoordinates": {Spec: SpecEXIF, ID: "Exif:GPSCoordinates", Value: "37.7749,-122.4194"},
						},
					},
				},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metadata.GPSCoordinates()
			if tt.wantNil && result != nil {
				t.Errorf("GPSCoordinates() = %v, want nil", result)
			}
			if !tt.wantNil && result == nil {
				t.Error("GPSCoordinates() = nil, want non-nil")
			}
		})
	}
}

func TestMetadata_ISO(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		want     int
	}{
		{
			name: "valid ISO",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "ExifIFD",
						Tags: map[TagID]Tag{
							"Exif:ISO": {Spec: SpecEXIF, ID: "Exif:ISO", Value: 800},
						},
					},
				},
			},
			want: 800,
		},
		{
			name:     "missing tag",
			metadata: Metadata{},
			want:     0,
		},
		{
			name: "wrong type",
			metadata: Metadata{
				Directories: []Directory{
					{
						Spec: SpecEXIF,
						Name: "ExifIFD",
						Tags: map[TagID]Tag{
							"Exif:ISO": {Spec: SpecEXIF, ID: "Exif:ISO", Value: "800"},
						},
					},
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metadata.ISO()
			if result != tt.want {
				t.Errorf("ISO() = %d, want %d", result, tt.want)
			}
		})
	}
}

func TestMetadata_BuildIndex(t *testing.T) {
	metadata := Metadata{
		Directories: []Directory{
			{
				Spec: SpecEXIF,
				Name: "IFD0",
				Tags: map[TagID]Tag{
					"Exif:Make":  {Spec: SpecEXIF, ID: "Exif:Make", Value: "Canon"},
					"Exif:Model": {Spec: SpecEXIF, ID: "Exif:Model", Value: "EOS 5D"},
				},
			},
			{
				Spec: SpecEXIF,
				Name: "ExifIFD",
				Tags: map[TagID]Tag{
					"Exif:ISO": {Spec: SpecEXIF, ID: "Exif:ISO", Value: 100},
				},
			},
		},
	}

	// Before BuildIndex
	if metadata.index != nil {
		t.Error("index should be nil before BuildIndex()")
	}

	metadata.BuildIndex()

	// After BuildIndex
	if metadata.index == nil {
		t.Fatal("index should not be nil after BuildIndex()")
	}

	if len(metadata.index) != 3 {
		t.Errorf("index has %d entries, want 3", len(metadata.index))
	}

	// Check tags are in index
	if tag, ok := metadata.index["Exif:Make"]; !ok {
		t.Error("Exif:Make not found in index")
	} else if tag.Value != "Canon" {
		t.Errorf("index[Exif:Make].Value = %v, want %q", tag.Value, "Canon")
	}
}

func TestGPSCoord(t *testing.T) {
	coord := GPSCoord{
		Lat:      37.7749,
		Lon:      -122.4194,
		Altitude: 10.5,
	}

	if coord.Lat != 37.7749 {
		t.Errorf("Lat = %f, want %f", coord.Lat, 37.7749)
	}
	if coord.Lon != -122.4194 {
		t.Errorf("Lon = %f, want %f", coord.Lon, -122.4194)
	}
	if coord.Altitude != 10.5 {
		t.Errorf("Altitude = %f, want %f", coord.Altitude, 10.5)
	}
}

// Verify type aliases work correctly
func TestTypeAliases(t *testing.T) {
	// Spec alias
	var spec Spec = SpecEXIF
	if spec != meta.SpecEXIF {
		t.Error("Spec alias not working correctly")
	}

	// TagID alias
	var tagID TagID = "Exif:Make"
	if tagID != meta.TagID("Exif:Make") {
		t.Error("TagID alias not working correctly")
	}

	// Tag alias
	var tag Tag = meta.Tag{ID: "Exif:Make"}
	if tag.ID != "Exif:Make" {
		t.Error("Tag alias not working correctly")
	}

	// Directory alias
	var dir Directory = meta.Directory{Name: "IFD0"}
	if dir.Name != "IFD0" {
		t.Error("Directory alias not working correctly")
	}
}
