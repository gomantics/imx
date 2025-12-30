package mp4

// Atom represents an MP4 atom/box
type Atom struct {
	Type   string
	Size   uint64
	Offset int64
}
