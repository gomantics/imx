package iptc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/limits"
)

// Parser parses IPTC metadata from Photoshop IRB.
type Parser struct{}

// New creates a new IPTC parser.
func New() *Parser {
	return &Parser{}
}

// Name returns the parser name.
func (p *Parser) Name() string {
	return "IPTC"
}

// Detect checks if the data contains Photoshop 8BIM signature.
func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, 4)
	_, err := r.ReadAt(buf, 0)
	return err == nil && bytes.Equal(buf, signature8BIM)
}

// Parse extracts IPTC metadata from Photoshop Image Resource Blocks.
func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()

	// Find IPTC resource in Photoshop IRB structure
	iptcOffset, iptcSize, err := p.findIPTCResource(r)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to find IPTC resource: %w", err))
		return nil, parseErr
	}

	if iptcSize == 0 {
		// No IPTC data found, try parsing as raw IPTC-IIM
		iptcOffset = 0
		// Try to determine size by reading until we hit an error
		iptcSize = 64 * 1024 // reasonable max for IPTC data
	}

	// Parse IPTC-IIM data
	datasets, err := p.parseIPTCIIM(r, iptcOffset, iptcSize)
	if err != nil {
		parseErr.Add(fmt.Errorf("failed to parse IPTC-IIM: %w", err))
	}

	if len(datasets) == 0 {
		return nil, parseErr.OrNil()
	}

	// Build directories from datasets
	dirs := p.buildDirectories(datasets)
	return dirs, parseErr.OrNil()
}

// findIPTCResource scans Photoshop IRB structure to find IPTC resource.
// Returns offset and size of IPTC data, or (0, 0, nil) if not found.
func (p *Parser) findIPTCResource(r io.ReaderAt) (int64, int64, error) {
	var offset int64 = 0
	headerBuf := make([]byte, 7) // 4 (sig) + 2 (ID) + 1 (nameLen)

	for {
		// Read IRB header
		_, err := r.ReadAt(headerBuf, offset)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return 0, 0, nil
		}
		if err != nil {
			return 0, 0, err
		}

		// Check for 8BIM signature
		if !bytes.Equal(headerBuf[0:4], signature8BIM) {
			offset++
			continue
		}

		// Found 8BIM, parse the block structure
		resourceID := binary.BigEndian.Uint16(headerBuf[4:6])
		nameLen := int(headerBuf[6])

		// Name is padded to even length (including length byte)
		namePadded := nameLen
		if (nameLen+1)%2 != 0 {
			namePadded++
		}

		// Read data size
		dataSizeOffset := offset + 7 + int64(namePadded)
		sizeBuf := make([]byte, 4)
		_, err = r.ReadAt(sizeBuf, dataSizeOffset)
		if err != nil {
			// Can't read size, treat as invalid and continue byte search
			offset++
			continue
		}
		dataSize := int64(binary.BigEndian.Uint32(sizeBuf))

		// Check if this is IPTC resource
		if resourceID == ResourceIPTC {
			dataOffset := dataSizeOffset + 4
			return dataOffset, dataSize, nil
		}

		// Not IPTC, skip entire block to next resource
		offset = dataSizeOffset + 4 + dataSize
		if dataSize%2 != 0 {
			offset++
		}
	}
}

func (p *Parser) parseIPTCIIM(r io.ReaderAt, offset int64, maxSize int64) ([]Dataset, error) {
	var datasets []Dataset
	pos := offset
	end := offset + maxSize

	for pos < end {
		if err := p.readDataset(r, &pos, end, &datasets); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return datasets, err
		}
	}

	return datasets, nil
}

func (p *Parser) readDataset(r io.ReaderAt, pos *int64, end int64, datasets *[]Dataset) error {
	if *pos >= end {
		return io.EOF
	}

	marker, err := p.readByte(r, pos)
	if err != nil {
		return err
	}

	if marker != iptcTagMarker {
		return nil
	}

	record, err := p.readByte(r, pos)
	if err != nil {
		return err
	}

	datasetID, err := p.readByte(r, pos)
	if err != nil {
		return err
	}

	dataSize, err := p.readSize(r, pos)
	if err != nil {
		return err
	}

	data := make([]byte, dataSize)
	if dataSize > 0 {
		_, err = r.ReadAt(data, *pos)
		if err != nil {
			return err
		}
		*pos += int64(dataSize)
	}

	*datasets = append(*datasets, Dataset{
		Record:    Record(record),
		DatasetID: datasetID,
		Name:      getDatasetName(Record(record), datasetID),
		Value:     parseDatasetValue(Record(record), datasetID, data),
		Raw:       data,
	})

	return nil
}

func (p *Parser) readByte(r io.ReaderAt, pos *int64) (byte, error) {
	buf := make([]byte, 1)
	_, err := r.ReadAt(buf, *pos)
	if err != nil {
		return 0, err
	}
	*pos++
	return buf[0], nil
}

func (p *Parser) readSize(r io.ReaderAt, pos *int64) (int, error) {
	buf := make([]byte, 2)
	_, err := r.ReadAt(buf, *pos)
	if err != nil {
		return 0, err
	}
	*pos += 2

	size := binary.BigEndian.Uint16(buf)

	if size&sizeExtendedFlag == 0 {
		return int(size), nil
	}

	extLen := int(size & sizeExtendedMask)
	if extLen == 0 || extLen > maxExtendedSizeLen {
		return 0, fmt.Errorf("invalid extended size length: %d", extLen)
	}

	extBuf := make([]byte, extLen)
	_, err = r.ReadAt(extBuf, *pos)
	if err != nil {
		return 0, err
	}
	*pos += int64(extLen)

	// Use int64 to prevent overflow in size calculation
	var extSize int64
	for i := 0; i < extLen; i++ {
		extSize = (extSize << 8) | int64(extBuf[i])
	}

	// Validate against limit
	if extSize > limits.MaxIPTCDatasetSize {
		return 0, fmt.Errorf("extended size %d exceeds limit of %d bytes", extSize, limits.MaxIPTCDatasetSize)
	}

	return int(extSize), nil
}

func (p *Parser) buildDirectories(datasets []Dataset) []parser.Directory {
	byRecord := make(map[Record][]Dataset)
	for _, ds := range datasets {
		byRecord[ds.Record] = append(byRecord[ds.Record], ds)
	}

	var dirs []parser.Directory

	for record, recordDatasets := range byRecord {
		dir := parser.Directory{
			Name: "IPTC-" + record.String(),
			Tags: make([]parser.Tag, 0),
		}

		tagValues := make(map[string][]any)

		for _, ds := range recordDatasets {
			tagValues[ds.Name] = append(tagValues[ds.Name], ds.Value)
		}

		for name, values := range tagValues {
			var value any
			var dataType string

			if len(values) == 1 {
				value = values[0]
				switch value.(type) {
				case int:
					dataType = "int"
				default:
					dataType = "string"
				}
			} else {
				value = values
				dataType = "array"
			}

			dir.Tags = append(dir.Tags, parser.Tag{
				ID:       parser.TagID("IPTC:" + name),
				Name:     name,
				Value:    value,
				DataType: dataType,
			})
		}

		if len(dir.Tags) > 0 {
			dirs = append(dirs, dir)
		}
	}

	return dirs
}
