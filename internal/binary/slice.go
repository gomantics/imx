package binary

import "encoding/binary"

// Uint16BE reads a big-endian uint16 from a byte slice at the given offset.
func Uint16BE(b []byte, offset int) uint16 {
	return binary.BigEndian.Uint16(b[offset:])
}

// Uint16LE reads a little-endian uint16 from a byte slice at the given offset.
func Uint16LE(b []byte, offset int) uint16 {
	return binary.LittleEndian.Uint16(b[offset:])
}

// Uint32BE reads a big-endian uint32 from a byte slice at the given offset.
func Uint32BE(b []byte, offset int) uint32 {
	return binary.BigEndian.Uint32(b[offset:])
}

// Uint32LE reads a little-endian uint32 from a byte slice at the given offset.
func Uint32LE(b []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(b[offset:])
}

// Uint64BE reads a big-endian uint64 from a byte slice at the given offset.
func Uint64BE(b []byte, offset int) uint64 {
	return binary.BigEndian.Uint64(b[offset:])
}

// Uint64LE reads a little-endian uint64 from a byte slice at the given offset.
func Uint64LE(b []byte, offset int) uint64 {
	return binary.LittleEndian.Uint64(b[offset:])
}

// PutUint16BE writes a big-endian uint16 to a byte slice at the given offset.
func PutUint16BE(b []byte, offset int, v uint16) {
	binary.BigEndian.PutUint16(b[offset:], v)
}

// PutUint16LE writes a little-endian uint16 to a byte slice at the given offset.
func PutUint16LE(b []byte, offset int, v uint16) {
	binary.LittleEndian.PutUint16(b[offset:], v)
}

// PutUint32BE writes a big-endian uint32 to a byte slice at the given offset.
func PutUint32BE(b []byte, offset int, v uint32) {
	binary.BigEndian.PutUint32(b[offset:], v)
}

// PutUint32LE writes a little-endian uint32 to a byte slice at the given offset.
func PutUint32LE(b []byte, offset int, v uint32) {
	binary.LittleEndian.PutUint32(b[offset:], v)
}

// PutUint64BE writes a big-endian uint64 to a byte slice at the given offset.
func PutUint64BE(b []byte, offset int, v uint64) {
	binary.BigEndian.PutUint64(b[offset:], v)
}

// PutUint64LE writes a little-endian uint64 to a byte slice at the given offset.
func PutUint64LE(b []byte, offset int, v uint64) {
	binary.LittleEndian.PutUint64(b[offset:], v)
}
