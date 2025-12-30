package heic

// Box represents an ISOBMFF box.
type Box struct {
	Type    string // 4-char box type
	Size    uint64 // Total size including header
	Offset  int64  // File offset to start of box
	Payload int64  // Offset to payload (after size+type+largesize)
}

// HeifIndex contains the parsed structure of HEIF metadata.
type HeifIndex struct {
	PrimaryItemID uint32
	Items         map[uint32]*HeifItem
	MdatOffset    int64 // Offset to mdat box
	MdatSize      uint64
}

// HeifItem represents an item in the HEIF meta box.
type HeifItem struct {
	ItemID       uint32
	ItemType     string
	Location     ItemLocation
	Properties   []uint32
	References   []uint32 // Item IDs this item references (cdsc)
	ReferencedBy []uint32 // Item IDs that reference this item
	ICCProperty  *Box     // Reference to colr property box containing ICC profile
}

// ItemLocation describes where an item's data is located.
type ItemLocation struct {
	BaseOffset uint64
	Extents    []Extent
}

// Extent represents a contiguous data region.
type Extent struct {
	Offset uint64 // File offset or mdat-relative
	Length uint64
}

// PropertyEntry represents a property in ipco.
type PropertyEntry struct {
	Index uint32
	Type  string
	Box   *Box
}
