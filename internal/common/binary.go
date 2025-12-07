package common

import (
	"encoding/binary"
	"fmt"
)

// ReadUint16 safely reads a uint16 from data with bounds checking
func ReadUint16(data []byte, offset int, order binary.ByteOrder) (uint16, error) {
	if offset+2 > len(data) {
		return 0, fmt.Errorf("offset %d+2 out of bounds (len=%d)", offset, len(data))
	}
	return order.Uint16(data[offset : offset+2]), nil
}

// ReadUint32 safely reads a uint32 from data with bounds checking
func ReadUint32(data []byte, offset int, order binary.ByteOrder) (uint32, error) {
	if offset+4 > len(data) {
		return 0, fmt.Errorf("offset %d+4 out of bounds (len=%d)", offset, len(data))
	}
	return order.Uint32(data[offset : offset+4]), nil
}

// ReadUint64 safely reads a uint64 from data with bounds checking
func ReadUint64(data []byte, offset int, order binary.ByteOrder) (uint64, error) {
	if offset+8 > len(data) {
		return 0, fmt.Errorf("offset %d+8 out of bounds (len=%d)", offset, len(data))
	}
	return order.Uint64(data[offset : offset+8]), nil
}

// SafeSlice safely slices data with bounds checking
func SafeSlice(data []byte, offset, size int) ([]byte, error) {
	if offset < 0 || size < 0 {
		return nil, fmt.Errorf("negative offset or size")
	}
	if offset+size > len(data) {
		return nil, fmt.Errorf("slice [%d:%d] out of bounds (len=%d)", offset, offset+size, len(data))
	}
	return data[offset : offset+size], nil
}

// ParseS15Fixed16 parses a signed 15.16 fixed-point number
func ParseS15Fixed16(data []byte) (float64, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("insufficient data for s15Fixed16")
	}
	val := int32(binary.BigEndian.Uint32(data))
	return float64(val) / 65536.0, nil
}

// ParseU16Fixed16 parses an unsigned 16.16 fixed-point number
func ParseU16Fixed16(data []byte) (float64, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("insufficient data for u16Fixed16")
	}
	val := binary.BigEndian.Uint32(data)
	return float64(val) / 65536.0, nil
}

// ParseU8Fixed8 parses an unsigned 8.8 fixed-point number
func ParseU8Fixed8(data []byte) (float64, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("insufficient data for u8Fixed8")
	}
	val := binary.BigEndian.Uint16(data)
	return float64(val) / 256.0, nil
}

// TrimNullBytes removes null bytes from a string
func TrimNullBytes(s string) string {
	for i, c := range s {
		if c == 0 {
			return s[:i]
		}
	}
	return s
}

// TrimNullBytesFromSlice converts byte slice to string, trimming nulls
func TrimNullBytesFromSlice(data []byte) string {
	return TrimNullBytes(string(data))
}
