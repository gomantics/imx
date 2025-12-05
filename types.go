package imx

import (
	"time"

	"github.com/gomantics/imx/internal/meta"
	"github.com/gomantics/imx/internal/types"
)

// Namespace represents a metadata namespace (EXIF, IPTC, XMP, ICC, etc.)
type Namespace = types.Namespace

const (
	NamespaceEXIF = types.NamespaceEXIF
	NamespaceIPTC = types.NamespaceIPTC
	NamespaceXMP  = types.NamespaceXMP
	NamespaceICC  = types.NamespaceICC
)

// TagID is a unique identifier for a metadata tag (e.g. "Exif:DateTimeOriginal")
type TagID = meta.TagID

// Tag represents a single metadata attribute
type Tag = meta.Tag

// Directory is a logical collection of tags for a given namespace and grouping
type Directory = meta.Directory

// Metadata is the top-level container for all parsed metadata
type Metadata struct {
	Directories []Directory
	index       map[TagID]*Tag // Internal index for fast lookup
}

// Directory returns the directory with the given namespace and name
func (m *Metadata) Directory(namespace Namespace, name string) (Directory, bool) {
	for _, dir := range m.Directories {
		if dir.Namespace == namespace && dir.Name == name {
			return dir, true
		}
	}
	return Directory{}, false
}

// Tag returns the tag with the given namespace and ID
func (m *Metadata) Tag(namespace Namespace, id TagID) (Tag, bool) {
	// Use index if available
	if m.index != nil {
		if tag, ok := m.index[id]; ok && tag.Namespace == namespace {
			return *tag, true
		}
	}

	// Fallback: scan directories
	for _, dir := range m.Directories {
		if dir.Namespace == namespace {
			if tag, ok := dir.Tags[id]; ok {
				return tag, true
			}
		}
	}
	return Tag{}, false
}

// GetAll returns a map of values for the given tag IDs
func (m *Metadata) GetAll(ids ...TagID) map[TagID]any {
	result := make(map[TagID]any, len(ids))
	for _, id := range ids {
		if m.index != nil {
			if tag, ok := m.index[id]; ok {
				result[id] = tag.Value
			}
		} else {
			// Fallback: scan directories
			for _, dir := range m.Directories {
				if tag, ok := dir.Tags[id]; ok {
					result[id] = tag.Value
					break
				}
			}
		}
	}
	return result
}

// Each iterates over all tags, calling fn for each tag.
// If fn returns false, iteration stops.
func (m *Metadata) Each(fn func(Directory, Tag) bool) {
	for _, dir := range m.Directories {
		for _, tag := range dir.Tags {
			if !fn(dir, tag) {
				return
			}
		}
	}
}

// EachInNamespace iterates over tags in the given namespace.
// If fn returns false, iteration stops.
func (m *Metadata) EachInNamespace(namespace Namespace, fn func(Tag) bool) {
	for _, dir := range m.Directories {
		if dir.Namespace == namespace {
			for _, tag := range dir.Tags {
				if !fn(tag) {
					return
				}
			}
		}
	}
}

// Convenience helpers for common EXIF fields

// DateTimeOriginal returns the EXIF DateTimeOriginal as time.Time (zero value if missing)
func (m *Metadata) DateTimeOriginal() time.Time {
	if tag, ok := m.Tag(NamespaceEXIF, "Exif:DateTimeOriginal"); ok {
		if t, ok := tag.Value.(time.Time); ok {
			return t
		}
	}
	return time.Time{}
}

// Orientation returns the EXIF Orientation (0 if missing)
func (m *Metadata) Orientation() int {
	if tag, ok := m.Tag(NamespaceEXIF, "Exif:Orientation"); ok {
		if i, ok := tag.Value.(int); ok {
			return i
		}
	}
	return 0
}

// Make returns the camera make (empty string if missing)
func (m *Metadata) Make() string {
	if tag, ok := m.Tag(NamespaceEXIF, "Exif:Make"); ok {
		if s, ok := tag.Value.(string); ok {
			return s
		}
	}
	return ""
}

// Model returns the camera model (empty string if missing)
func (m *Metadata) Model() string {
	if tag, ok := m.Tag(NamespaceEXIF, "Exif:Model"); ok {
		if s, ok := tag.Value.(string); ok {
			return s
		}
	}
	return ""
}

// GPSCoord represents a GPS coordinate
type GPSCoord struct {
	Lat      float64
	Lon      float64
	Altitude float64
}

// GPSCoordinates returns the GPS coordinates (nil if missing)
func (m *Metadata) GPSCoordinates() *GPSCoord {
	if tag, ok := m.Tag(NamespaceEXIF, "Exif:GPSCoordinates"); ok {
		if gps, ok := tag.Value.(*GPSCoord); ok {
			return gps
		}
	}
	return nil
}

// ISO returns the ISO speed (0 if missing)
func (m *Metadata) ISO() int {
	if tag, ok := m.Tag(NamespaceEXIF, "Exif:ISO"); ok {
		if i, ok := tag.Value.(int); ok {
			return i
		}
	}
	return 0
}

// BuildIndex builds an internal index for fast tag lookup
func (m *Metadata) BuildIndex() {
	m.index = make(map[TagID]*Tag)
	for i := range m.Directories {
		dir := &m.Directories[i]
		for id := range dir.Tags {
			tag := dir.Tags[id]
			m.index[id] = &tag
		}
	}
}
