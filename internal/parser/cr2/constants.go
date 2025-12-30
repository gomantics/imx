package cr2

// CR2 format constants
// Reference: Canon CR2 (Canon Raw 2) specification

const (
	// CR2 magic bytes offset in file (after TIFF header)
	cr2MagicOffset = 8

	// CR2 magic bytes "CR" (0x43 0x52)
	cr2MagicByte1 = 0x43 // 'C'
	cr2MagicByte2 = 0x52 // 'R'

	// CR2 major version (always 0x02 for CR2 format)
	cr2MajorVersion = 0x02

	// CR2 version offset in file
	cr2VersionOffset = 10
)
