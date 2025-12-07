package iptc

import (
	"encoding/binary"
	"fmt"
	"io"
)

// DatasetReader reads IPTC datasets from a byte stream
type DatasetReader struct {
	data   []byte
	offset int
}

// NewDatasetReader creates a new dataset reader
func NewDatasetReader(data []byte) *DatasetReader {
	return &DatasetReader{
		data:   data,
		offset: 0,
	}
}

// EOF returns true if at end of data
func (r *DatasetReader) EOF() bool {
	return r.offset >= len(r.data)
}

// Skip skips n bytes
func (r *DatasetReader) Skip(n int) {
	r.offset += n
	if r.offset > len(r.data) {
		r.offset = len(r.data)
	}
}

// readByte reads a single byte
func (r *DatasetReader) readByte() (byte, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	b := r.data[r.offset]
	r.offset++
	return b, nil
}

// readBytes reads n bytes
func (r *DatasetReader) readBytes(n int) ([]byte, error) {
	if r.offset+n > len(r.data) {
		return nil, io.EOF
	}
	bytes := r.data[r.offset : r.offset+n]
	r.offset += n
	return bytes, nil
}

// expectMarker reads and validates the IPTC tag marker
func (r *DatasetReader) expectMarker() error {
	b, err := r.readByte()
	if err != nil {
		return err
	}
	if b != iptcTagMarker {
		return fmt.Errorf("invalid marker: 0x%02X, expected 0x1C", b)
	}
	return nil
}

// readSize reads dataset size (handles extended sizes)
func (r *DatasetReader) readSize() (int, error) {
	if r.offset+2 > len(r.data) {
		return 0, io.EOF
	}

	sizeBytes := binary.BigEndian.Uint16(r.data[r.offset : r.offset+2])
	r.offset += 2

	// Check for extended size
	if sizeBytes&0x8000 != 0 {
		extLen := int(sizeBytes & 0x7FFF)
		if extLen > 4 || r.offset+extLen > len(r.data) {
			return 0, fmt.Errorf("invalid extended size length: %d", extLen)
		}

		// Read extended size
		size := 0
		for i := 0; i < extLen; i++ {
			size = (size << 8) | int(r.data[r.offset])
			r.offset++
		}
		return size, nil
	}

	return int(sizeBytes), nil
}

// ReadNext reads the next dataset
func (r *DatasetReader) ReadNext() (*Dataset, error) {
	if r.EOF() {
		return nil, io.EOF
	}

	// Expect marker - skip bytes until we find one or EOF
	for {
		if r.EOF() {
			return nil, io.EOF
		}

		// Peek at current byte
		if r.data[r.offset] != iptcTagMarker {
			r.offset++
			continue
		}

		// Found marker, consume it
		r.offset++
		break
	}

	// Read record
	record, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("read record: %w", err)
	}

	// Read dataset ID
	datasetID, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("read dataset ID: %w", err)
	}

	// Read size
	size, err := r.readSize()
	if err != nil {
		return nil, fmt.Errorf("read size: %w", err)
	}

	// Read value
	value, err := r.readBytes(size)
	if err != nil {
		return nil, fmt.Errorf("read value: %w", err)
	}

	// Build dataset
	name := getDatasetName(Record(record), datasetID)
	if name == "" {
		name = fmt.Sprintf("Dataset%d:%d", record, datasetID)
	}

	return &Dataset{
		Record:    Record(record),
		DatasetID: datasetID,
		Name:      name,
		Value:     parseDatasetValue(Record(record), datasetID, value),
		Raw:       value,
	}, nil
}
