package flac

import "fmt"

// Metadata block types as defined in the FLAC specification.
// Reference: https://xiph.org/flac/format.html#metadata_block
const (
	blockTypeStreamInfo    = 0
	blockTypePadding       = 1
	blockTypeApplication   = 2
	blockTypeSeekTable     = 3
	blockTypeVorbisComment = 4
	blockTypeCueSheet      = 5
	blockTypePicture       = 6
)

// pictureTypes maps FLAC picture type codes to their descriptive names.
// Reference: FLAC specification, Section 4.6 (PICTURE block)
// These types are based on ID3v2 APIC frame picture types.
var pictureTypes = map[uint32]string{
	0:  "Other",
	1:  "32x32 PNG File Icon",
	2:  "Other File Icon",
	3:  "Cover (front)",
	4:  "Cover (back)",
	5:  "Leaflet Page",
	6:  "Media",
	7:  "Lead Artist/Lead Performer/Soloist",
	8:  "Artist/Performer",
	9:  "Conductor",
	10: "Band/Orchestra",
	11: "Composer",
	12: "Lyricist/Text Writer",
	13: "Recording Location",
	14: "During Recording",
	15: "During Performance",
	16: "Movie/Video Screen Capture",
	17: "A Bright Colored Fish",
	18: "Illustration",
	19: "Band/Artist Logotype",
	20: "Publisher/Studio Logotype",
}

// getPictureType returns a human-readable picture type description.
func getPictureType(t uint32) string {
	if str, ok := pictureTypes[t]; ok {
		return str
	}
	return fmt.Sprintf("Unknown (%d)", t)
}
