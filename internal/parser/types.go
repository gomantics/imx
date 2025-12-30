package parser

// TagID uniquely identifies a tag (e.g., "EXIF:IFD0:Make").
type TagID string

// Tag represents a metadata tag.
type Tag struct {
	ID       TagID
	Name     string
	Value    any
	DataType string
}

// Directory represents a collection of tags.
type Directory struct {
	Name string
	Tags []Tag
}
