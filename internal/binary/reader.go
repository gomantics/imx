package binary

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Reader provides byte-order aware reading from io.ReaderAt.
// It wraps an io.ReaderAt and applies a consistent byte order for all read operations.
//
// Usage:
//
//	reader := binary.NewReader(r, binary.BigEndian)
//	value16, err := reader.ReadUint16(offset)
//	value32, err := reader.ReadUint32(offset)
type Reader struct {
	r     io.ReaderAt
	order binary.ByteOrder
}

// NewReader creates a new Reader with the specified byte order.
func NewReader(r io.ReaderAt, order binary.ByteOrder) *Reader {
	return &Reader{
		r:     r,
		order: order,
	}
}

// ReadUint16 reads a uint16 at the given offset using the Reader's byte order.
func (r *Reader) ReadUint16(offset int64) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := r.r.ReadAt(buf, offset); err != nil {
		return 0, fmt.Errorf("failed to read uint16 at offset %d: %w", offset, err)
	}
	if r.order == binary.BigEndian {
		return Uint16BE(buf, 0), nil
	}
	return Uint16LE(buf, 0), nil
}

// ReadUint32 reads a uint32 at the given offset using the Reader's byte order.
func (r *Reader) ReadUint32(offset int64) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := r.r.ReadAt(buf, offset); err != nil {
		return 0, fmt.Errorf("failed to read uint32 at offset %d: %w", offset, err)
	}
	if r.order == binary.BigEndian {
		return Uint32BE(buf, 0), nil
	}
	return Uint32LE(buf, 0), nil
}

// ReadUint64 reads a uint64 at the given offset using the Reader's byte order.
func (r *Reader) ReadUint64(offset int64) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := r.r.ReadAt(buf, offset); err != nil {
		return 0, fmt.Errorf("failed to read uint64 at offset %d: %w", offset, err)
	}
	if r.order == binary.BigEndian {
		return Uint64BE(buf, 0), nil
	}
	return Uint64LE(buf, 0), nil
}

// ReadInt16 reads an int16 at the given offset using the Reader's byte order.
func (r *Reader) ReadInt16(offset int64) (int16, error) {
	v, err := r.ReadUint16(offset)
	return int16(v), err
}

// ReadInt32 reads an int32 at the given offset using the Reader's byte order.
func (r *Reader) ReadInt32(offset int64) (int32, error) {
	v, err := r.ReadUint32(offset)
	return int32(v), err
}

// ReadBytes reads n bytes at the given offset.
func (r *Reader) ReadBytes(offset int64, n int) ([]byte, error) {
	result := make([]byte, n)
	if _, err := r.r.ReadAt(result, offset); err != nil {
		return nil, fmt.Errorf("failed to read %d bytes at offset %d: %w", n, offset, err)
	}
	return result, nil
}

// PutUint16 writes a uint16 to a byte slice using the Reader's byte order.
func (r *Reader) PutUint16(b []byte, v uint16) {
	if r.order == binary.BigEndian {
		PutUint16BE(b, 0, v)
	} else {
		PutUint16LE(b, 0, v)
	}
}

// PutUint32 writes a uint32 to a byte slice using the Reader's byte order.
func (r *Reader) PutUint32(b []byte, v uint32) {
	if r.order == binary.BigEndian {
		PutUint32BE(b, 0, v)
	} else {
		PutUint32LE(b, 0, v)
	}
}

// Uint16 reads a uint16 from a byte slice using the Reader's byte order.
func (r *Reader) Uint16(b []byte) uint16 {
	if r.order == binary.BigEndian {
		return Uint16BE(b, 0)
	}
	return Uint16LE(b, 0)
}
