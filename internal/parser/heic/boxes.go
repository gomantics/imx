package heic

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser/limits"
)

// readBoxHeader reads a box header at the given offset.
func readBoxHeader(r io.ReaderAt, offset int64) (*Box, error) {
	hdr := make([]byte, boxHeaderSize)
	if _, err := r.ReadAt(hdr[:boxHeaderSize], offset); err != nil {
		return nil, err
	}

	size := uint64(binary.BigEndian.Uint32(hdr[0:4]))
	boxType := string(hdr[4:8])

	// Validate box type (printable ASCII)
	for i := 0; i < 4; i++ {
		if hdr[4+i] < 32 || hdr[4+i] > 126 {
			return nil, fmt.Errorf("invalid box type at offset %d", offset)
		}
	}

	box := &Box{
		Type:    boxType,
		Size:    size,
		Offset:  offset,
		Payload: offset + boxHeaderSize,
	}

	// Handle size == 1 (64-bit size follows)
	if size == sizeExtended {
		largeSizeBuf := make([]byte, 8)
		if _, err := r.ReadAt(largeSizeBuf[:8], offset+boxHeaderSize); err != nil {
			return nil, err
		}
		box.Size = binary.BigEndian.Uint64(largeSizeBuf[:8])
		box.Payload = offset + boxHeaderLargeSize

		// Validate extended size to prevent malicious files with unreasonable box sizes
		if box.Size > limits.MaxHEICBoxSize {
			return nil, fmt.Errorf("box size %d exceeds maximum allowed size %d at offset %d", box.Size, limits.MaxHEICBoxSize, offset)
		}
	}

	// Handle size == 0 (box extends to EOF) - not supported for safety
	if box.Size == sizeToEOF {
		return nil, fmt.Errorf("size=0 boxes not supported")
	}

	// Validate minimum box size to prevent infinite loops
	if box.Size < boxHeaderSize {
		return nil, fmt.Errorf("invalid box size %d at offset %d", box.Size, offset)
	}

	// Validate maximum box size for standard (32-bit) sizes
	if size != sizeExtended && box.Size > limits.MaxHEICBoxSize {
		return nil, fmt.Errorf("box size %d exceeds maximum allowed size %d at offset %d", box.Size, limits.MaxHEICBoxSize, offset)
	}

	return box, nil
}

// findBox finds the first box of given type, searching from offset up to maxScan bytes.
func findBox(r io.ReaderAt, boxType string, offset int64, maxScan int64) (*Box, error) {
	scanned := int64(0)

	for scanned < maxScan {
		box, err := readBoxHeader(r, offset)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil, fmt.Errorf("box %s not found", boxType)
			}
			return nil, err
		}

		if box.Type == boxType {
			return box, nil
		}

		offset += int64(box.Size)
		scanned += int64(box.Size)
	}

	return nil, fmt.Errorf("box %s not found within %d bytes", boxType, maxScan)
}

// findChildBox finds a child box within a parent box.
func findChildBox(r io.ReaderAt, parent *Box, childType string) (*Box, error) {
	offset := parent.Payload
	endOffset := parent.Offset + int64(parent.Size)

	for offset < endOffset {
		box, err := readBoxHeader(r, offset)
		if err != nil {
			return nil, err
		}

		if box.Type == childType {
			return box, nil
		}

		offset += int64(box.Size)
	}

	return nil, fmt.Errorf("child box %s not found", childType)
}

// iterateChildren calls fn for each child box in parent.
func iterateChildren(r io.ReaderAt, parent *Box, fn func(*Box) error) error {
	offset := parent.Payload
	endOffset := parent.Offset + int64(parent.Size)

	for offset < endOffset {
		if endOffset-offset < boxHeaderSize {
			break // Not enough space for header
		}

		box, err := readBoxHeader(r, offset)
		if err != nil {
			return err
		}

		if err := fn(box); err != nil {
			return err
		}

		offset += int64(box.Size)
	}

	return nil
}

// boxTypeEquals compares box type bytes to a string without allocation.
func boxTypeEquals(b []byte, expected string) bool {
	if len(b) < 4 || len(expected) != 4 {
		return false
	}
	return b[0] == expected[0] && b[1] == expected[1] &&
		b[2] == expected[2] && b[3] == expected[3]
}

// readUint reads a variable-length unsigned integer (1-8 bytes, big-endian).
func readUint(data []byte, size int) uint64 {
	if size <= 0 || size > 8 || len(data) < size {
		return 0
	}

	var val uint64
	for i := 0; i < size; i++ {
		val = (val << 8) | uint64(data[i])
	}
	return val
}
