package iptc

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Photoshop 8BIM signature
var signature8BIM = []byte("8BIM")

// IPTC tag marker
const iptcTagMarker = 0x1C

// parsePhotoshopIRB parses Photoshop Image Resource Blocks
// Returns the IPTC-IIM data if found
func parsePhotoshopIRB(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short for IRB")
	}

	offset := 0
	for offset+12 <= len(data) {
		// Check for 8BIM signature
		if !bytes.Equal(data[offset:offset+4], signature8BIM) {
			// Try to find next 8BIM
			idx := bytes.Index(data[offset:], signature8BIM)
			if idx < 0 {
				break
			}
			offset += idx
			continue
		}
		offset += 4

		// Resource ID (2 bytes) - loop guard ensures at least 12 bytes from offset
		resourceID := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		// Pascal string (name) - first byte is length
		nameLen := int(data[offset])
		offset++
		
		// Name is padded to even length (including length byte)
		// If nameLen is even, we need 1 byte padding; if odd, no padding
		namePadded := nameLen
		if (nameLen+1)%2 != 0 {
			namePadded++
		}
		offset += namePadded

		// Resource data size (4 bytes)
		if offset+4 > len(data) {
			break
		}
		dataSize := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4

		// Resource data
		if offset+dataSize > len(data) {
			break
		}
		resourceData := data[offset : offset+dataSize]

		// Check if this is IPTC resource
		if resourceID == ResourceIPTC {
			return resourceData, nil
		}

		// Move to next resource (padded to even)
		offset += dataSize
		if dataSize%2 != 0 {
			offset++
		}
	}

	return nil, nil
}

// parseIPTCIIM parses IPTC-IIM (Information Interchange Model) data
func parseIPTCIIM(data []byte) ([]Dataset, error) {
	if len(data) < 5 {
		return nil, nil
	}

	var datasets []Dataset
	offset := 0

	for offset+5 <= len(data) {
		// Tag marker (0x1C)
		if data[offset] != iptcTagMarker {
			offset++
			continue
		}
		offset++

		// Record number (1 byte)
		record := Record(data[offset])
		offset++

		// Dataset number (1 byte)
		datasetID := data[offset]
		offset++

		// Data size (2 bytes for standard, 4 bytes for extended)
		// Loop guard ensures at least 5 bytes from offset; we've read 3, so 2 remain
		var dataSize int
		sizeBytes := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		if sizeBytes&0x8000 != 0 {
			// Extended size (bit 15 set means extended)
			// Next bytes indicate actual size
			extLen := int(sizeBytes & 0x7FFF)
			if extLen > 4 || offset+extLen > len(data) {
				break
			}
			// Read extended size
			dataSize = 0
			for i := 0; i < extLen; i++ {
				dataSize = (dataSize << 8) | int(data[offset])
				offset++
			}
		} else {
			dataSize = int(sizeBytes)
		}

		// Read dataset value
		if offset+dataSize > len(data) {
			break
		}
		value := data[offset : offset+dataSize]
		offset += dataSize

		// Get dataset name
		name := getDatasetName(record, datasetID)
		if name == "" {
			name = fmt.Sprintf("Dataset%d:%d", record, datasetID)
		}

		// Parse value based on dataset
		parsedValue := parseDatasetValue(record, datasetID, value)

		datasets = append(datasets, Dataset{
			Record:    record,
			DatasetID: datasetID,
			Name:      name,
			Value:     parsedValue,
			Raw:       value,
		})
	}

	return datasets, nil
}

// parseDatasetValue parses the value based on dataset type
func parseDatasetValue(record Record, datasetID uint8, data []byte) any {
	// Most IPTC values are text strings
	// Some are binary or have special formats

	if record == RecordApplication {
		switch datasetID {
		case 0: // RecordVersion
			if len(data) >= 2 {
				return int(binary.BigEndian.Uint16(data))
			}
		case 10: // Urgency
			if len(data) >= 1 {
				return int(data[0] - '0')
			}
		case 55, 62: // DateCreated, DigitalCreationDate (CCYYMMDD)
			return parseDateString(data)
		case 60, 63: // TimeCreated, DigitalCreationTime (HHMMSS±HHMM)
			return parseTimeString(data)
		case 30, 37: // ReleaseDate, ExpirationDate
			return parseDateString(data)
		case 35, 38: // ReleaseTime, ExpirationTime
			return parseTimeString(data)
		}
	}

	if record == RecordEnvelope {
		switch datasetID {
		case 0: // RecordVersion
			if len(data) >= 2 {
				return int(binary.BigEndian.Uint16(data))
			}
		case 70: // DateSent
			return parseDateString(data)
		case 80: // TimeSent
			return parseTimeString(data)
		}
	}

	// Default: treat as string
	return string(bytes.TrimRight(data, "\x00"))
}

// parseDateString parses IPTC date format (CCYYMMDD or YYYYMMDD)
func parseDateString(data []byte) string {
	s := string(data)
	if len(s) == 8 {
		// Format as YYYY-MM-DD
		return s[0:4] + "-" + s[4:6] + "-" + s[6:8]
	}
	return s
}

// parseTimeString parses IPTC time format (HHMMSS±HHMM)
func parseTimeString(data []byte) string {
	s := string(data)
	if len(s) >= 6 {
		result := s[0:2] + ":" + s[2:4] + ":" + s[4:6]
		if len(s) >= 11 {
			// Include timezone
			result += " " + s[6:7] + s[7:9] + ":" + s[9:11]
		}
		return result
	}
	return s
}

