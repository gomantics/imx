package heic

import (
	"encoding/binary"
	"fmt"
	"io"
)

// buildHeifIndex parses the meta box and builds the HEIF index.
func buildHeifIndex(r io.ReaderAt, metaBox *Box, maxScan int64) (*HeifIndex, error) {
	index := &HeifIndex{
		Items: make(map[uint32]*HeifItem),
	}

	// Meta box has version/flags (4 bytes) before children
	metaPayload := metaBox.Payload + fullBoxHeaderSize

	// Find mdat box for offset calculations
	mdatBox, err := findBox(r, boxTypeMdat, 0, maxScan)
	if err == nil {
		index.MdatOffset = mdatBox.Offset
		index.MdatSize = mdatBox.Size
	}

	// Create a virtual box for iterating meta children
	metaChildren := &Box{
		Type:    metaBox.Type,
		Size:    metaBox.Size,
		Offset:  metaBox.Offset,
		Payload: metaPayload,
	}

	// Parse hdlr (handler) - required but we don't need data
	_, err = findChildBox(r, metaChildren, boxTypeHdlr)
	if err != nil {
		return nil, fmt.Errorf("hdlr box required in meta: %w", err)
	}

	// Parse pitm (primary item)
	pitmBox, err := findChildBox(r, metaChildren, boxTypePitm)
	if err == nil {
		if err := parsePitm(r, pitmBox, index); err != nil {
			return nil, err
		}
	}

	// Parse iinf (item info)
	iinfBox, err := findChildBox(r, metaChildren, boxTypeIinf)
	if err != nil {
		return nil, fmt.Errorf("iinf box required: %w", err)
	}
	if err := parseIinf(r, iinfBox, index); err != nil {
		return nil, err
	}

	// Parse iloc (item locations)
	ilocBox, err := findChildBox(r, metaChildren, boxTypeIloc)
	if err != nil {
		return nil, fmt.Errorf("iloc box required: %w", err)
	}
	if err := parseIloc(r, ilocBox, index); err != nil {
		return nil, err
	}

	// Parse iref (item references) - optional
	irefBox, err := findChildBox(r, metaChildren, boxTypeIref)
	if err == nil {
		if err := parseIref(r, irefBox, index); err != nil {
			return nil, err
		}
	}

	// Parse iprp (item properties) - optional
	iprpBox, err := findChildBox(r, metaChildren, boxTypeIprp)
	if err == nil {
		if err := parseIprp(r, iprpBox, index); err != nil {
			return nil, err
		}
	}

	return index, nil
}

// parsePitm parses the primary item box.
func parsePitm(r io.ReaderAt, box *Box, index *HeifIndex) error {
	data := make([]byte, 8)
	if _, err := r.ReadAt(data, box.Payload); err != nil {
		return err
	}

	version := data[0]
	if version == 0 {
		index.PrimaryItemID = uint32(binary.BigEndian.Uint16(data[4:6]))
	} else {
		index.PrimaryItemID = binary.BigEndian.Uint32(data[4:8])
	}

	return nil
}

// parseIinf parses the item information box.
func parseIinf(r io.ReaderAt, box *Box, index *HeifIndex) error {
	data := make([]byte, 8)
	if _, err := r.ReadAt(data, box.Payload); err != nil {
		return err
	}

	version := data[0]
	var count uint32
	if version == 0 {
		count = uint32(binary.BigEndian.Uint16(data[4:6]))
	} else {
		count = binary.BigEndian.Uint32(data[4:8])
	}

	// Offset to first infe box
	offset := box.Payload + 8
	if version == 0 {
		offset = box.Payload + 6
	}

	for i := uint32(0); i < count; i++ {
		infeBox, err := readBoxHeader(r, offset)
		if err != nil {
			return err
		}

		if infeBox.Type != boxTypeInfe {
			return fmt.Errorf("expected infe box, got %s", infeBox.Type)
		}

		if err := parseInfe(r, infeBox, box, index); err != nil {
			return err
		}

		offset += int64(infeBox.Size)
	}

	return nil
}

// parseInfe parses a single item info entry.
func parseInfe(r io.ReaderAt, infeBox *Box, parentBox *Box, index *HeifIndex) error {
	infeData := make([]byte, 16)
	if _, err := r.ReadAt(infeData, infeBox.Payload); err != nil {
		return err
	}

	infeVersion := infeData[0]
	var itemID uint32
	var itemType string

	if infeVersion == 2 || infeVersion == 3 {
		itemID = uint32(binary.BigEndian.Uint16(infeData[4:6]))
		// Item type at offset 8 (4 bytes)
		if int(infeBox.Payload)+12 <= int(parentBox.Offset)+int(parentBox.Size) {
			typeData := make([]byte, 4)
			if _, err := r.ReadAt(typeData, infeBox.Payload+8); err == nil {
				itemType = string(typeData)
			}
		}
	}

	// Create or update item
	if _, exists := index.Items[itemID]; !exists {
		index.Items[itemID] = &HeifItem{
			ItemID:   itemID,
			ItemType: itemType,
		}
	} else {
		index.Items[itemID].ItemType = itemType
	}

	return nil
}

// parseIloc parses the item location box.
func parseIloc(r io.ReaderAt, box *Box, index *HeifIndex) error {
	data := make([]byte, 12)
	if _, err := r.ReadAt(data, box.Payload); err != nil {
		return err
	}

	version := data[0]
	offsetSize := (data[4] >> 4) & 0x0F
	lengthSize := data[4] & 0x0F
	baseOffsetSize := (data[5] >> 4) & 0x0F
	indexSize := uint8(0)
	if version == 1 || version == 2 {
		indexSize = data[5] & 0x0F
	}

	var itemCount uint32
	if version < 2 {
		itemCount = uint32(binary.BigEndian.Uint16(data[6:8]))
	} else {
		itemCount = binary.BigEndian.Uint32(data[8:12])
	}

	offset := box.Payload + 8
	if version >= 2 {
		offset = box.Payload + 12
	}

	for i := uint32(0); i < itemCount; i++ {
		entryData := make([]byte, 64)
		if _, err := r.ReadAt(entryData, offset); err != nil {
			return err
		}

		pos, itemID := parseIlocItemID(entryData, version)
		if version >= 1 {
			pos += 2 // Skip construction_method
		}
		pos += 2 // Skip data_reference_index

		baseOffset := uint64(0)
		if baseOffsetSize > 0 {
			baseOffset = readUint(entryData[pos:], int(baseOffsetSize))
			pos += int(baseOffsetSize)
		}

		extentCount := binary.BigEndian.Uint16(entryData[pos : pos+2])
		pos += 2

		extents := parseIlocExtents(entryData[pos:], extentCount, version, indexSize, offsetSize, lengthSize)

		// Create or update item
		if _, exists := index.Items[itemID]; !exists {
			index.Items[itemID] = &HeifItem{ItemID: itemID}
		}

		index.Items[itemID].Location = ItemLocation{
			BaseOffset: baseOffset,
			Extents:    extents,
		}

		offset += int64(pos) + int64(len(extents))*(int64(indexSize)+int64(offsetSize)+int64(lengthSize))
	}

	return nil
}

// parseIlocItemID extracts item ID from iloc entry data.
func parseIlocItemID(data []byte, version uint8) (pos int, itemID uint32) {
	if version < 2 {
		itemID = uint32(binary.BigEndian.Uint16(data[0:2]))
		return 2, itemID
	}
	itemID = binary.BigEndian.Uint32(data[0:4])
	return 4, itemID
}

// parseIlocExtents parses extents from iloc entry.
func parseIlocExtents(data []byte, count uint16, version uint8, indexSize, offsetSize, lengthSize uint8) []Extent {
	var extents []Extent
	pos := 0

	for j := uint16(0); j < count && j < 1000; j++ {
		if version >= 1 && indexSize > 0 {
			pos += int(indexSize)
		}

		extentOffset := uint64(0)
		if offsetSize > 0 {
			extentOffset = readUint(data[pos:], int(offsetSize))
			pos += int(offsetSize)
		}

		extentLength := readUint(data[pos:], int(lengthSize))
		pos += int(lengthSize)

		extents = append(extents, Extent{
			Offset: extentOffset,
			Length: extentLength,
		})
	}

	return extents
}

// parseIref parses the item reference box.
func parseIref(r io.ReaderAt, box *Box, index *HeifIndex) error {
	data := make([]byte, 4)
	if _, err := r.ReadAt(data, box.Payload); err != nil {
		return err
	}

	version := data[0]
	offset := box.Payload + 4
	endOffset := box.Offset + int64(box.Size)

	for offset < endOffset {
		refBox, err := readBoxHeader(r, offset)
		if err != nil {
			break
		}

		// Only care about cdsc (content describes) references
		if refBox.Type == boxTypeCdsc {
			parseIrefCdsc(r, refBox, version, index)
		}

		offset += int64(refBox.Size)
	}

	return nil
}

// parseIrefCdsc parses a cdsc reference entry.
func parseIrefCdsc(r io.ReaderAt, refBox *Box, version uint8, index *HeifIndex) {
	refData := make([]byte, 32)
	if _, err := r.ReadAt(refData, refBox.Payload); err != nil {
		return
	}

	pos := 0
	var fromID uint32
	if version == 0 {
		fromID = uint32(binary.BigEndian.Uint16(refData[pos : pos+2]))
		pos += 2
	} else {
		fromID = binary.BigEndian.Uint32(refData[pos : pos+4])
		pos += 4
	}

	refCount := binary.BigEndian.Uint16(refData[pos : pos+2])
	pos += 2

	for i := uint16(0); i < refCount && i < 100; i++ {
		var toID uint32
		if version == 0 {
			toID = uint32(binary.BigEndian.Uint16(refData[pos : pos+2]))
			pos += 2
		} else {
			toID = binary.BigEndian.Uint32(refData[pos : pos+4])
			pos += 4
		}

		if fromItem, exists := index.Items[fromID]; exists {
			fromItem.References = append(fromItem.References, toID)
		}
		if toItem, exists := index.Items[toID]; exists {
			toItem.ReferencedBy = append(toItem.ReferencedBy, fromID)
		}
	}
}

// parseIprp parses the item properties box.
func parseIprp(r io.ReaderAt, box *Box, index *HeifIndex) error {
	// Find ipco (property container)
	ipcoBox, err := findChildBox(r, box, boxTypeIpco)
	if err != nil {
		return nil // Optional
	}

	// Parse properties in ipco
	var properties []PropertyEntry
	propIndex := uint32(1) // Properties are 1-indexed

	err = iterateChildren(r, ipcoBox, func(propBox *Box) error {
		properties = append(properties, PropertyEntry{
			Index: propIndex,
			Type:  propBox.Type,
			Box:   propBox,
		})
		propIndex++
		return nil
	})
	if err != nil {
		return err
	}

	// Find ipma (property association)
	ipmaBox, err := findChildBox(r, box, boxTypeIpma)
	if err != nil {
		return nil // Optional
	}

	return parseIpma(r, ipmaBox, index, properties)
}

// parseIpma parses item property associations.
func parseIpma(r io.ReaderAt, box *Box, index *HeifIndex, properties []PropertyEntry) error {
	data := make([]byte, 8)
	if _, err := r.ReadAt(data, box.Payload); err != nil {
		return err
	}

	version := data[0]
	flags := binary.BigEndian.Uint32([]byte{0, data[1], data[2], data[3]})
	entryCount := binary.BigEndian.Uint32(data[4:8])

	offset := box.Payload + 8

	for i := uint32(0); i < entryCount && i < 10000; i++ {
		assocData := make([]byte, 32)
		if _, err := r.ReadAt(assocData, offset); err != nil {
			break
		}

		pos, itemID := parseIpmaItemID(assocData, version)
		assocCount := assocData[pos]
		pos++

		for j := uint8(0); j < assocCount && j < 100; j++ {
			propIndex, bytesRead := parseIpmaProperty(assocData[pos:], flags)
			pos += bytesRead

			if item, exists := index.Items[itemID]; exists {
				item.Properties = append(item.Properties, propIndex)

				// Store colr property reference for ICC extraction
				if int(propIndex) > 0 && int(propIndex) <= len(properties) {
					prop := properties[propIndex-1]
					if prop.Type == boxTypeColr && item.ICCProperty == nil {
						item.ICCProperty = prop.Box
					}
				}
			}
		}

		offset += int64(pos)
	}

	return nil
}

// parseIpmaItemID extracts item ID from ipma entry.
func parseIpmaItemID(data []byte, version uint8) (pos int, itemID uint32) {
	if version < 1 {
		itemID = uint32(binary.BigEndian.Uint16(data[0:2]))
		return 2, itemID
	}
	itemID = binary.BigEndian.Uint32(data[0:4])
	return 4, itemID
}

// parseIpmaProperty extracts property index from ipma association.
func parseIpmaProperty(data []byte, flags uint32) (propIndex uint32, bytesRead int) {
	if (flags & 1) != 0 {
		// 15-bit index + 1-bit essential flag
		val := binary.BigEndian.Uint16(data[0:2])
		propIndex = uint32(val & maskPropertyIndex15)
		return propIndex, 2
	}
	// 7-bit index + 1-bit essential flag
	propIndex = uint32(data[0] & maskPropertyIndex7)
	return propIndex, 1
}
