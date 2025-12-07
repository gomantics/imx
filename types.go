package imx

import (
	"github.com/gomantics/imx/internal/common"
)

// Format represents an image container format (JPEG, PNG, WebP, etc.)
type Format = common.Format

const (
	FormatJPEG = common.FormatJPEG
	FormatPNG  = common.FormatPNG
	FormatWebP = common.FormatWebP
	FormatTIFF = common.FormatTIFF
	FormatHEIF = common.FormatHEIF
)

// Spec represents a metadata specification (EXIF, IPTC, XMP, ICC, etc.)
type Spec = common.Spec

const (
	SpecEXIF = common.SpecEXIF
	SpecIPTC = common.SpecIPTC
	SpecXMP  = common.SpecXMP
	SpecICC  = common.SpecICC
)

// TagID is a unique identifier for a metadata tag (e.g. "EXIF:DateTimeOriginal")
type TagID = common.TagID

// Tag represents a single metadata attribute
type Tag = common.Tag

// Directory is a logical collection of tags for a given kind and grouping
type Directory = common.Directory

// Metadata is the top-level container for all parsed metadata
type Metadata struct {
	Directories []Directory
	index       map[TagID]*Tag // Internal index for fast lookup
}

// Directory returns the directory with the given spec and name
func (m *Metadata) Directory(spec Spec, name string) (Directory, bool) {
	for _, dir := range m.Directories {
		if dir.Spec == spec && dir.Name == name {
			return dir, true
		}
	}
	return Directory{}, false
}

// Tag returns the tag with the given spec and ID
func (m *Metadata) Tag(spec Spec, id TagID) (Tag, bool) {
	// Use index if available
	if m.index != nil {
		if tag, ok := m.index[id]; ok && tag.Spec == spec {
			return *tag, true
		}
	}

	// Fallback: scan directories
	for _, dir := range m.Directories {
		if dir.Spec == spec {
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

// EachInSpec iterates over tags in the given spec.
// If fn returns false, iteration stops.
func (m *Metadata) EachInSpec(spec Spec, fn func(Tag) bool) {
	for _, dir := range m.Directories {
		if dir.Spec == spec {
			for _, tag := range dir.Tags {
				if !fn(tag) {
					return
				}
			}
		}
	}
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
