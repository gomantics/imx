// Package binutil provides binary reading utilities for parsing binary file formats.
package binary

import (
	"encoding/binary"
	"fmt"
	"io"

)

// ReadUint16BE reads a 16-bit big-endian unsigned integer from r at the given offset.
func ReadUint16BE(r io.ReaderAt, offset int64) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := r.ReadAt(buf, offset); err != nil {
		return 0, fmt.Errorf("failed to read uint16 at offset %d: %w", offset, err)
	}
	return binary.BigEndian.Uint16(buf), nil
}

// ReadUint16LE reads a 16-bit little-endian unsigned integer from r at the given offset.
func ReadUint16LE(r io.ReaderAt, offset int64) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := r.ReadAt(buf, offset); err != nil {
		return 0, fmt.Errorf("failed to read uint16 at offset %d: %w", offset, err)
	}
	return binary.LittleEndian.Uint16(buf), nil
}

// ReadUint32BE reads a 32-bit big-endian unsigned integer from r at the given offset.
func ReadUint32BE(r io.ReaderAt, offset int64) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := r.ReadAt(buf, offset); err != nil {
		return 0, fmt.Errorf("failed to read uint32 at offset %d: %w", offset, err)
	}
	return binary.BigEndian.Uint32(buf), nil
}

// ReadUint32LE reads a 32-bit little-endian unsigned integer from r at the given offset.
func ReadUint32LE(r io.ReaderAt, offset int64) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := r.ReadAt(buf, offset); err != nil {
		return 0, fmt.Errorf("failed to read uint32 at offset %d: %w", offset, err)
	}
	return binary.LittleEndian.Uint32(buf), nil
}

// ReadUint64BE reads a 64-bit big-endian unsigned integer from r at the given offset.
func ReadUint64BE(r io.ReaderAt, offset int64) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := r.ReadAt(buf, offset); err != nil {
		return 0, fmt.Errorf("failed to read uint64 at offset %d: %w", offset, err)
	}
	return binary.BigEndian.Uint64(buf), nil
}

// ReadUint64LE reads a 64-bit little-endian unsigned integer from r at the given offset.
func ReadUint64LE(r io.ReaderAt, offset int64) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := r.ReadAt(buf, offset); err != nil {
		return 0, fmt.Errorf("failed to read uint64 at offset %d: %w", offset, err)
	}
	return binary.LittleEndian.Uint64(buf), nil
}
