package imx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

func TestMetadata_Directories(t *testing.T) {
	tests := []struct {
		name     string
		metadata *Metadata
		wantLen  int
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			wantLen:  0,
		},
		{
			name:     "empty directories",
			metadata: &Metadata{},
			wantLen:  0,
		},
		{
			name: "has directories",
			metadata: &Metadata{
				directories: []Directory{
					{Name: "IFD0"},
					{Name: "ExifIFD"},
				},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs := tt.metadata.Directories()
			if len(dirs) != tt.wantLen {
				t.Errorf("Directories() len = %d, want %d", len(dirs), tt.wantLen)
			}
		})
	}
}

func TestMetadata_Errors(t *testing.T) {
	tests := []struct {
		name     string
		metadata *Metadata
		wantLen  int
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			wantLen:  0,
		},
		{
			name:     "no errors",
			metadata: &Metadata{},
			wantLen:  0,
		},
		{
			name: "has errors",
			metadata: &Metadata{
				errors: []error{fmt.Errorf("error 1"), fmt.Errorf("error 2")},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.metadata.Errors()
			if len(errs) != tt.wantLen {
				t.Errorf("Errors() len = %d, want %d", len(errs), tt.wantLen)
			}
		})
	}
}

func TestMetadata_Directory(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		dirName  string
		wantOk   bool
	}{
		{
			name: "find existing directory",
			metadata: Metadata{
				directories: []Directory{
					{Name: "IFD0"},
					{Name: "ExifIFD"},
				},
			},
			dirName: "IFD0",
			wantOk:  true,
		},
		{
			name: "directory not found",
			metadata: Metadata{
				directories: []Directory{
					{Name: "IFD0"},
				},
			},
			dirName: "GPS",
			wantOk:  false,
		},
		{
			name:     "empty directories",
			metadata: Metadata{},
			dirName:  "IFD0",
			wantOk:   false,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			dir, ok := tt.metadata.Directory(tt.dirName)
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
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make", Name: "Make", Value: "Canon"},
					{ID: "EXIF:IFD0:Model", Name: "Model", Value: "EOS 5D"},
				},
			},
		},
	}

	tests := []struct {
		name   string
		id     TagID
		wantOk bool
		want   any
	}{
		{
			name:   "find existing tag",
			id:     "EXIF:IFD0:Make",
			wantOk: true,
			want:   "Canon",
		},
		{
			name:   "tag not found",
			id:     "EXIF:IFD0:ISO",
			wantOk: false,
		},
		{
			name:   "different tag ID",
			id:     "XMP:Title",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, ok := metadata.Tag(tt.id)
			if ok != tt.wantOk {
				t.Errorf("Tag() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk && tag.Value != tt.want {
				t.Errorf("Tag().Value = %v, want %v", tag.Value, tt.want)
			}
		})
	}
}

func TestMetadata_Tag_LazyIndex(t *testing.T) {
	// Test that the index is built lazily on first call
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make", Name: "Make", Value: "Canon"},
				},
			},
		},
	}

	// Index should be nil initially
	if metadata.index != nil {
		t.Error("Index should be nil initially")
	}

	// First call should build index
	tag, ok := metadata.Tag("EXIF:IFD0:Make")
	if !ok {
		t.Error("Tag() should find tag")
	}
	if tag.Value != "Canon" {
		t.Errorf("Tag().Value = %v, want %q", tag.Value, "Canon")
	}

	// Index should now be built
	if metadata.index == nil {
		t.Error("Index should be built after first Tag() call")
	}

	// Second call should use index
	tag2, ok2 := metadata.Tag("EXIF:IFD0:Make")
	if !ok2 {
		t.Error("Tag() should find tag on second call")
	}
	if tag2.Value != "Canon" {
		t.Errorf("Tag().Value = %v, want %q on second call", tag2.Value, "Canon")
	}
}

func TestMetadata_Tag_ConcurrentSafety(t *testing.T) {
	// Test that Tag() is safe to call concurrently from multiple goroutines
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make", Name: "Make", Value: "Canon"},
					{ID: "EXIF:IFD0:Model", Name: "Model", Value: "EOS 5D"},
					{ID: "EXIF:IFD0:ISO", Name: "ISO", Value: 100},
				},
			},
			{
				Name: "ExifIFD",
				Tags: []Tag{
					{ID: "EXIF:ExifIFD:FNumber", Name: "FNumber", Value: 2.8},
					{ID: "EXIF:ExifIFD:ExposureTime", Name: "ExposureTime", Value: "1/500"},
				},
			},
		},
	}

	// Launch multiple goroutines that all try to access tags simultaneously
	// This tests both the lazy index initialization and concurrent reads
	const numGoroutines = 100
	const numIterations = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numIterations; j++ {
				// Try different tags to exercise the index
				tags := []TagID{
					"EXIF:IFD0:Make",
					"EXIF:IFD0:Model",
					"EXIF:IFD0:ISO",
					"EXIF:ExifIFD:FNumber",
					"EXIF:ExifIFD:ExposureTime",
					"EXIF:NonExistent", // Also test missing tags
				}

				for _, tagID := range tags {
					tag, ok := metadata.Tag(tagID)
					if tagID == "EXIF:NonExistent" {
						if ok {
							t.Errorf("Tag() found non-existent tag")
						}
					} else {
						if !ok {
							t.Errorf("Tag() failed to find %q", tagID)
						}
						if tag.ID != tagID {
							t.Errorf("Tag().ID = %q, want %q", tag.ID, tagID)
						}
					}
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify index was built correctly
	if metadata.index == nil {
		t.Error("Index should be built after concurrent access")
	}

	// Verify all expected tags are in the index
	expectedTags := []TagID{
		"EXIF:IFD0:Make",
		"EXIF:IFD0:Model",
		"EXIF:IFD0:ISO",
		"EXIF:ExifIFD:FNumber",
		"EXIF:ExifIFD:ExposureTime",
	}

	for _, tagID := range expectedTags {
		if _, ok := metadata.index[tagID]; !ok {
			t.Errorf("Index missing expected tag %q", tagID)
		}
	}
}

func TestMetadata_GetAll(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make", Name: "Make", Value: "Canon"},
					{ID: "EXIF:IFD0:Model", Name: "Model", Value: "EOS 5D"},
					{ID: "EXIF:IFD0:ISO", Name: "ISO", Value: 100},
				},
			},
		},
	}

	tests := []struct {
		name    string
		ids     []TagID
		wantLen int
	}{
		{
			name:    "get multiple existing tags",
			ids:     []TagID{"EXIF:IFD0:Make", "EXIF:IFD0:Model"},
			wantLen: 2,
		},
		{
			name:    "get with some missing",
			ids:     []TagID{"EXIF:IFD0:Make", "EXIF:Unknown"},
			wantLen: 1,
		},
		{
			name:    "get all missing",
			ids:     []TagID{"Unknown:A", "Unknown:B"},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metadata.GetAll(tt.ids...)
			if len(result) != tt.wantLen {
				t.Errorf("GetAll() returned %d items, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestMetadata_Each(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make", Name: "Make", Value: "Canon"},
					{ID: "EXIF:IFD0:Model", Name: "Model", Value: "EOS 5D"},
				},
			},
			{
				Name: "ExifIFD",
				Tags: []Tag{
					{ID: "EXIF:ExifIFD:ISO", Name: "ISO", Value: 100},
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

func TestMetadata_EachTag(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make", Value: "Canon"},
					{ID: "EXIF:IFD0:Model", Value: "EOS 5D"},
				},
			},
		},
	}

	t.Run("iterate all tags", func(t *testing.T) {
		count := 0
		metadata.EachTag(func(tag Tag) bool {
			count++
			return true
		})
		if count != 2 {
			t.Errorf("EachTag() iterated %d times, want 2", count)
		}
	})

	t.Run("early termination", func(t *testing.T) {
		count := 0
		metadata.EachTag(func(tag Tag) bool {
			count++
			return false
		})
		if count != 1 {
			t.Errorf("EachTag() iterated %d times after early termination, want 1", count)
		}
	})
}

func TestMetadata_EachInDirectory(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make", Value: "Canon"},
					{ID: "EXIF:IFD0:Model", Value: "EOS 5D"},
				},
			},
			{
				Name: "XMP",
				Tags: []Tag{
					{ID: "XMP:Title", Value: "Test"},
				},
			},
		},
	}

	t.Run("iterate IFD0 only", func(t *testing.T) {
		count := 0
		metadata.EachInDirectory("IFD0", func(tag Tag) bool {
			count++
			return true
		})
		if count != 2 {
			t.Errorf("EachInDirectory() iterated %d times, want 2", count)
		}
	})

	t.Run("iterate XMP only", func(t *testing.T) {
		count := 0
		metadata.EachInDirectory("XMP", func(tag Tag) bool {
			count++
			return true
		})
		if count != 1 {
			t.Errorf("EachInDirectory() iterated %d times, want 1", count)
		}
	})

	t.Run("directory not found", func(t *testing.T) {
		count := 0
		metadata.EachInDirectory("NonExistent", func(tag Tag) bool {
			count++
			return true
		})
		if count != 0 {
			t.Errorf("EachInDirectory() for non-existent directory iterated %d times, want 0", count)
		}
	})
}

func TestMetadata_AllTags(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make", Value: "Canon"},
					{ID: "EXIF:IFD0:Model", Value: "EOS 5D"},
				},
			},
			{
				Name: "ExifIFD",
				Tags: []Tag{
					{ID: "EXIF:ExifIFD:ISO", Value: 100},
				},
			},
		},
	}

	tags := metadata.AllTags()
	if len(tags) != 3 {
		t.Errorf("AllTags() returned %d tags, want 3", len(tags))
	}

	// Test with empty metadata
	emptyMeta := Metadata{}
	emptyTags := emptyMeta.AllTags()
	if len(emptyTags) != 0 {
		t.Errorf("AllTags() on empty metadata returned %d tags, want 0", len(emptyTags))
	}
}

func TestMetadata_DirectoryNames(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{Name: "IFD0"},
			{Name: "ExifIFD"},
			{Name: "GPS"},
		},
	}

	names := metadata.DirectoryNames()
	if len(names) != 3 {
		t.Errorf("DirectoryNames() returned %d names, want 3", len(names))
	}

	expected := map[string]bool{"IFD0": true, "ExifIFD": true, "GPS": true}
	for _, name := range names {
		if !expected[name] {
			t.Errorf("DirectoryNames() returned unexpected name: %q", name)
		}
	}
}

func TestMetadata_TagCount(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make"},
					{ID: "EXIF:IFD0:Model"},
				},
			},
			{
				Name: "ExifIFD",
				Tags: []Tag{
					{ID: "EXIF:ExifIFD:ISO"},
				},
			},
		},
	}

	count := metadata.TagCount()
	if count != 3 {
		t.Errorf("TagCount() = %d, want 3", count)
	}

	// Test with empty metadata
	emptyMeta := Metadata{}
	emptyCount := emptyMeta.TagCount()
	if emptyCount != 0 {
		t.Errorf("TagCount() on empty metadata = %d, want 0", emptyCount)
	}
}

func TestMetadata_GetString(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Make", Value: "Canon"},
					{ID: "EXIF:IFD0:ISO", Value: 100},
					{ID: "EXIF:IFD0:Data", Value: []byte("test")},
				},
			},
		},
	}

	tests := []struct {
		name    string
		id      TagID
		want    string
		wantErr bool
	}{
		{
			name:    "string value",
			id:      "EXIF:IFD0:Make",
			want:    "Canon",
			wantErr: false,
		},
		{
			name:    "int converted to string",
			id:      "EXIF:IFD0:ISO",
			want:    "100",
			wantErr: false,
		},
		{
			name:    "bytes converted to string",
			id:      "EXIF:IFD0:Data",
			want:    "test",
			wantErr: false,
		},
		{
			name:    "tag not found",
			id:      "EXIF:NotFound",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metadata.GetString(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMetadata_GetInt(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "int", Value: int(42)},
					{ID: "int8", Value: int8(42)},
					{ID: "int16", Value: int16(42)},
					{ID: "int32", Value: int32(42)},
					{ID: "int64", Value: int64(100)},
					{ID: "uint", Value: uint(42)},
					{ID: "uint8", Value: uint8(42)},
					{ID: "uint16", Value: uint16(42)},
					{ID: "uint32", Value: uint32(1920)},
					{ID: "uint64", Value: uint64(42)},
					{ID: "uint64_overflow", Value: uint64(1 << 63)},
					{ID: "string", Value: "Canon"},
				},
			},
		},
	}

	tests := []struct {
		name    string
		id      TagID
		want    int64
		wantErr bool
	}{
		{name: "int", id: "int", want: 42, wantErr: false},
		{name: "int8", id: "int8", want: 42, wantErr: false},
		{name: "int16", id: "int16", want: 42, wantErr: false},
		{name: "int32", id: "int32", want: 42, wantErr: false},
		{name: "int64", id: "int64", want: 100, wantErr: false},
		{name: "uint", id: "uint", want: 42, wantErr: false},
		{name: "uint8", id: "uint8", want: 42, wantErr: false},
		{name: "uint16", id: "uint16", want: 42, wantErr: false},
		{name: "uint32", id: "uint32", want: 1920, wantErr: false},
		{name: "uint64", id: "uint64", want: 42, wantErr: false},
		{name: "uint64 overflow", id: "uint64_overflow", wantErr: true},
		{name: "string value", id: "string", wantErr: true},
		{name: "tag not found", id: "EXIF:NotFound", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metadata.GetInt(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMetadata_GetFloat(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "float32", Value: float32(2.5)},
					{ID: "float64", Value: float64(2.8)},
					{ID: "int", Value: int(42)},
					{ID: "int8", Value: int8(42)},
					{ID: "int16", Value: int16(42)},
					{ID: "int32", Value: int32(42)},
					{ID: "int64", Value: int64(100)},
					{ID: "uint", Value: uint(42)},
					{ID: "uint8", Value: uint8(42)},
					{ID: "uint16", Value: uint16(42)},
					{ID: "uint32", Value: uint32(42)},
					{ID: "uint64", Value: uint64(42)},
					{ID: "string", Value: "Canon"},
				},
			},
		},
	}

	tests := []struct {
		name    string
		id      TagID
		want    float64
		wantErr bool
	}{
		{name: "float32", id: "float32", want: 2.5, wantErr: false},
		{name: "float64", id: "float64", want: 2.8, wantErr: false},
		{name: "int", id: "int", want: 42.0, wantErr: false},
		{name: "int8", id: "int8", want: 42.0, wantErr: false},
		{name: "int16", id: "int16", want: 42.0, wantErr: false},
		{name: "int32", id: "int32", want: 42.0, wantErr: false},
		{name: "int64", id: "int64", want: 100.0, wantErr: false},
		{name: "uint", id: "uint", want: 42.0, wantErr: false},
		{name: "uint8", id: "uint8", want: 42.0, wantErr: false},
		{name: "uint16", id: "uint16", want: 42.0, wantErr: false},
		{name: "uint32", id: "uint32", want: 42.0, wantErr: false},
		{name: "uint64", id: "uint64", want: 42.0, wantErr: false},
		{name: "string value", id: "string", wantErr: true},
		{name: "tag not found", id: "EXIF:NotFound", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metadata.GetFloat(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetFloat() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestMetadata_GetBytes(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "EXIF:IFD0:Data", Value: []byte("test")},
					{ID: "EXIF:IFD0:Text", Value: "string"},
					{ID: "EXIF:IFD0:Number", Value: 100},
				},
			},
		},
	}

	tests := []struct {
		name    string
		id      TagID
		want    []byte
		wantErr bool
	}{
		{
			name:    "byte slice value",
			id:      "EXIF:IFD0:Data",
			want:    []byte("test"),
			wantErr: false,
		},
		{
			name:    "string converted to bytes",
			id:      "EXIF:IFD0:Text",
			want:    []byte("string"),
			wantErr: false,
		},
		{
			name:    "int value",
			id:      "EXIF:IFD0:Number",
			wantErr: true,
		},
		{
			name:    "tag not found",
			id:      "EXIF:NotFound",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metadata.GetBytes(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != string(tt.want) {
				t.Errorf("GetBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test type that implements fmt.Stringer
type testStringer struct {
	value string
}

func (ts testStringer) String() string {
	return ts.value
}

func TestMetadata_GetString_Stringer(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "stringer", Value: testStringer{value: "custom"}},
				},
			},
		},
	}

	got, err := metadata.GetString("stringer")
	if err != nil {
		t.Fatalf("GetString() error = %v", err)
	}
	if got != "custom" {
		t.Errorf("GetString() = %q, want %q", got, "custom")
	}
}

func TestMetadata_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		wantJSON string
	}{
		{
			name: "with directories and no errors",
			metadata: Metadata{
				directories: []Directory{
					{Name: "IFD0", Tags: []Tag{{ID: "make", Name: "Make", Value: "Canon"}}},
				},
			},
			wantJSON: `{"directories":[{"Name":"IFD0","Tags":[{"ID":"make","Name":"Make","Value":"Canon","DataType":""}]}]}`,
		},
		{
			name: "with errors",
			metadata: Metadata{
				directories: []Directory{},
				errors:      []error{fmt.Errorf("parse error"), fmt.Errorf("read error")},
			},
			wantJSON: `{"directories":[],"errors":["parse error","read error"]}`,
		},
		{
			name:     "empty metadata",
			metadata: Metadata{},
			wantJSON: `{"directories":null}`,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(&tt.metadata)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.wantJSON {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.wantJSON)
			}
		})
	}
}

func TestMetadata_EachInDirectory_EarlyTermination(t *testing.T) {
	metadata := Metadata{
		directories: []Directory{
			{
				Name: "IFD0",
				Tags: []Tag{
					{ID: "tag1"},
					{ID: "tag2"},
					{ID: "tag3"},
				},
			},
		},
	}

	count := 0
	metadata.EachInDirectory("IFD0", func(tag Tag) bool {
		count++
		return false // Stop after first tag
	})

	if count != 1 {
		t.Errorf("EachInDirectory() iterated %d times, want 1", count)
	}
}

func TestReaderAdapter_EdgeCases(t *testing.T) {
	t.Run("read at offset beyond EOF", func(t *testing.T) {
		data := []byte("hello")
		adapter := newReaderAdapter(bytes.NewReader(data), 0, 0)

		buf := make([]byte, 10)
		n, err := adapter.ReadAt(buf, 100)
		if err != io.EOF {
			t.Errorf("ReadAt() error = %v, want io.EOF", err)
		}
		if n != 0 {
			t.Errorf("ReadAt() n = %d, want 0", n)
		}
	})

	t.Run("partial read returns UnexpectedEOF", func(t *testing.T) {
		data := []byte("hello")
		adapter := newReaderAdapter(bytes.NewReader(data), 0, 0)

		buf := make([]byte, 10)
		n, err := adapter.ReadAt(buf, 0)
		if err != io.ErrUnexpectedEOF {
			t.Errorf("ReadAt() error = %v, want io.ErrUnexpectedEOF", err)
		}
		if n != 5 {
			t.Errorf("ReadAt() n = %d, want 5", n)
		}
	})

	t.Run("read error from underlying reader", func(t *testing.T) {
		errReader := &errorReader{err: fmt.Errorf("read error")}
		adapter := newReaderAdapter(errReader, 0, 0)

		buf := make([]byte, 10)
		_, err := adapter.ReadAt(buf, 0)
		if err == nil || err.Error() != "read error" {
			t.Errorf("ReadAt() error = %v, want 'read error'", err)
		}
	})
}

// errorReader always returns an error
type errorReader struct {
	err error
}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, er.err
}
