package flac

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
)

// parseStreamInfo parses the STREAMINFO metadata block.
// This is the only mandatory metadata block and must be the first block.
// Reference: FLAC specification, Section 4.2.1
func (p *Parser) parseStreamInfo(r io.ReaderAt, start, length int64) *parser.Directory {
	if length < streamInfoMinSize {
		return nil
	}

	data := make([]byte, length)
	_, err := r.ReadAt(data, start)
	if err != nil {
		return nil
	}

	dir := &parser.Directory{
		Name: "FLAC-StreamInfo",
		Tags: []parser.Tag{},
	}

	// Parse fields using named offsets
	minBlockSize := binary.BigEndian.Uint16(data[streamInfoMinBlockSizeOffset : streamInfoMinBlockSizeOffset+2])
	maxBlockSize := binary.BigEndian.Uint16(data[streamInfoMaxBlockSizeOffset : streamInfoMaxBlockSizeOffset+2])

	// Frame sizes are 24-bit values
	minFrameSize := uint32(data[streamInfoMinFrameSizeOffset])<<16 |
		uint32(data[streamInfoMinFrameSizeOffset+1])<<8 |
		uint32(data[streamInfoMinFrameSizeOffset+2])
	maxFrameSize := uint32(data[streamInfoMaxFrameSizeOffset])<<16 |
		uint32(data[streamInfoMaxFrameSizeOffset+1])<<8 |
		uint32(data[streamInfoMaxFrameSizeOffset+2])

	// Sample rate (20 bits), channels (3 bits), bits per sample (5 bits)
	sampleRateHigh := uint32(data[streamInfoSampleRateOffset])<<12 |
		uint32(data[streamInfoSampleRateOffset+1])<<4 |
		uint32(data[streamInfoChannelsOffset])>>4
	channels := ((data[streamInfoChannelsOffset] >> 1) & 0x07) + 1
	bitsPerSample := ((data[streamInfoBitsPerSampleStart] & 0x01) << 4) |
		(data[streamInfoBitsPerSampleEnd] >> 4) + 1

	// Total samples (36 bits)
	totalSamples := (uint64(data[streamInfoTotalSamplesStart]&0x0F) << 32) |
		(uint64(data[streamInfoTotalSamplesStart+1]) << 24) |
		(uint64(data[streamInfoTotalSamplesStart+2]) << 16) |
		(uint64(data[streamInfoTotalSamplesStart+3]) << 8) |
		uint64(data[streamInfoTotalSamplesStart+4])

	// MD5 signature
	md5 := data[streamInfoMD5Offset : streamInfoMD5Offset+streamInfoMD5Size]

	// Add tags
	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:StreamInfo:MinBlockSize"),
		Name:     "MinimumBlockSize",
		Value:    minBlockSize,
		DataType: "uint16",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:StreamInfo:MaxBlockSize"),
		Name:     "MaximumBlockSize",
		Value:    maxBlockSize,
		DataType: "uint16",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:StreamInfo:MinFrameSize"),
		Name:     "MinimumFrameSize",
		Value:    minFrameSize,
		DataType: "uint32",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:StreamInfo:MaxFrameSize"),
		Name:     "MaximumFrameSize",
		Value:    maxFrameSize,
		DataType: "uint32",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:StreamInfo:SampleRate"),
		Name:     "SampleRate",
		Value:    sampleRateHigh,
		DataType: "uint32",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:StreamInfo:Channels"),
		Name:     "Channels",
		Value:    channels,
		DataType: "uint8",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:StreamInfo:BitsPerSample"),
		Name:     "BitsPerSample",
		Value:    bitsPerSample,
		DataType: "uint8",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:StreamInfo:TotalSamples"),
		Name:     "TotalSamples",
		Value:    totalSamples,
		DataType: "uint64",
	})

	// Calculate duration in seconds
	if sampleRateHigh > 0 {
		duration := float64(totalSamples) / float64(sampleRateHigh)
		dir.Tags = append(dir.Tags, parser.Tag{
			ID:       parser.TagID("FLAC:StreamInfo:Duration"),
			Name:     "Duration",
			Value:    fmt.Sprintf("%.2f seconds", duration),
			DataType: "string",
		})
	}

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:StreamInfo:MD5"),
		Name:     "MD5Signature",
		Value:    fmt.Sprintf("%x", md5),
		DataType: "string",
	})

	return dir
}

// parseVorbisComment parses Vorbis Comment metadata (tags like artist, title, etc.).
// Reference: https://www.xiph.org/vorbis/doc/v-comment.html
func (p *Parser) parseVorbisComment(r io.ReaderAt, start, length int64) *parser.Directory {
	data := make([]byte, length)
	_, err := r.ReadAt(data, start)
	if err != nil {
		return nil
	}

	dir := &parser.Directory{
		Name: "FLAC-Vorbis",
		Tags: []parser.Tag{},
	}

	offset := 0

	// Read vendor string length (32-bit little-endian)
	if offset+4 > len(data) {
		return dir
	}
	vendorLength := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read vendor string
	if offset+int(vendorLength) > len(data) {
		return dir
	}
	vendorString := string(data[offset : offset+int(vendorLength)])
	offset += int(vendorLength)

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:VorbisComment:Vendor"),
		Name:     "Vendor",
		Value:    vendorString,
		DataType: "string",
	})

	// Read number of comments
	if offset+4 > len(data) {
		return dir
	}
	numComments := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read each comment
	for i := uint32(0); i < numComments && offset < len(data); i++ {
		if offset+4 > len(data) {
			break
		}

		commentLength := binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4

		if offset+int(commentLength) > len(data) {
			break
		}

		comment := string(data[offset : offset+int(commentLength)])
		offset += int(commentLength)

		// Parse "KEY=VALUE" format
		if idx := bytes.IndexByte([]byte(comment), '='); idx > 0 {
			key := comment[:idx]
			value := comment[idx+1:]

			dir.Tags = append(dir.Tags, parser.Tag{
				ID:       parser.TagID(fmt.Sprintf("FLAC:VorbisComment:%s", key)),
				Name:     key,
				Value:    value,
				DataType: "string",
			})
		}
	}

	return dir
}

// parsePicture parses embedded picture metadata.
// Reference: FLAC specification, Section 4.6
func (p *Parser) parsePicture(r io.ReaderAt, start, length int64) *parser.Directory {
	data := make([]byte, length)
	_, err := r.ReadAt(data, start)
	if err != nil {
		return nil
	}

	dir := &parser.Directory{
		Name: "FLAC-Picture",
		Tags: []parser.Tag{},
	}

	offset := 0

	// Read picture type (32-bit big-endian)
	if offset+4 > len(data) {
		return dir
	}
	pictureType := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	pictureTypeStr := getPictureType(pictureType)
	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:Picture:Type"),
		Name:     "PictureType",
		Value:    pictureTypeStr,
		DataType: "string",
	})

	// Read MIME type length
	if offset+4 > len(data) {
		return dir
	}
	mimeLength := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read MIME type
	if offset+int(mimeLength) > len(data) {
		return dir
	}
	mimeType := string(data[offset : offset+int(mimeLength)])
	offset += int(mimeLength)

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:Picture:MIMEType"),
		Name:     "MIMEType",
		Value:    mimeType,
		DataType: "string",
	})

	// Read description length
	if offset+4 > len(data) {
		return dir
	}
	descLength := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read description
	if offset+int(descLength) > len(data) {
		return dir
	}
	description := string(data[offset : offset+int(descLength)])
	offset += int(descLength)

	if len(description) > 0 {
		dir.Tags = append(dir.Tags, parser.Tag{
			ID:       parser.TagID("FLAC:Picture:Description"),
			Name:     "Description",
			Value:    description,
			DataType: "string",
		})
	}

	// Read width, height, depth, colors (4 bytes each)
	if offset+16 > len(data) {
		return dir
	}
	width := binary.BigEndian.Uint32(data[offset : offset+4])
	height := binary.BigEndian.Uint32(data[offset+4 : offset+8])
	depth := binary.BigEndian.Uint32(data[offset+8 : offset+12])
	colors := binary.BigEndian.Uint32(data[offset+12 : offset+16])
	offset += 16

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:Picture:Width"),
		Name:     "Width",
		Value:    width,
		DataType: "uint32",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:Picture:Height"),
		Name:     "Height",
		Value:    height,
		DataType: "uint32",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:Picture:ColorDepth"),
		Name:     "ColorDepth",
		Value:    depth,
		DataType: "uint32",
	})

	if colors > 0 {
		dir.Tags = append(dir.Tags, parser.Tag{
			ID:       parser.TagID("FLAC:Picture:Colors"),
			Name:     "Colors",
			Value:    colors,
			DataType: "uint32",
		})
	}

	// Read picture data length
	if offset+4 > len(data) {
		return dir
	}
	pictureLength := binary.BigEndian.Uint32(data[offset : offset+4])

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:Picture:Size"),
		Name:     "PictureSize",
		Value:    fmt.Sprintf("%d bytes", pictureLength),
		DataType: "string",
	})

	return dir
}

// parsePadding returns information about a padding block.
// Padding blocks contain only null bytes and are used to reserve space for future metadata.
func (p *Parser) parsePadding(length int64) *parser.Directory {
	return &parser.Directory{
		Name: "FLAC-Padding",
		Tags: []parser.Tag{
			{
				ID:       parser.TagID("FLAC:Padding:Size"),
				Name:     "PaddingSize",
				Value:    fmt.Sprintf("%d bytes", length),
				DataType: "string",
			},
		},
	}
}

// parseApplication parses application-specific data.
// Applications can register their own block type for custom metadata.
func (p *Parser) parseApplication(r io.ReaderAt, start, length int64) *parser.Directory {
	if length < applicationIDSize {
		return nil
	}

	data := make([]byte, length)
	_, err := r.ReadAt(data, start)
	if err != nil {
		return nil
	}

	dir := &parser.Directory{
		Name: "FLAC-Application",
		Tags: []parser.Tag{},
	}

	// First 4 bytes are application ID
	appID := string(data[0:applicationIDSize])
	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:Application:ID"),
		Name:     "ApplicationID",
		Value:    appID,
		DataType: "string",
	})

	dir.Tags = append(dir.Tags, parser.Tag{
		ID:       parser.TagID("FLAC:Application:DataSize"),
		Name:     "DataSize",
		Value:    fmt.Sprintf("%d bytes", length-applicationIDSize),
		DataType: "string",
	})

	return dir
}

// parseSeekTable parses seek point information.
// Seek tables enable fast seeking to arbitrary sample positions.
func (p *Parser) parseSeekTable(r io.ReaderAt, start, length int64) *parser.Directory {
	if length%seekPointSize != 0 {
		return nil
	}

	numPoints := length / seekPointSize

	// Handle empty seek table (valid case)
	if length == 0 {
		return &parser.Directory{
			Name: "FLAC-SeekTable",
			Tags: []parser.Tag{
				{
					ID:       parser.TagID("FLAC:SeekTable:Points"),
					Name:     "SeekPoints",
					Value:    numPoints,
					DataType: "int64",
				},
			},
		}
	}

	data := make([]byte, length)
	_, err := r.ReadAt(data, start)
	if err != nil {
		return nil
	}

	dir := &parser.Directory{
		Name: "FLAC-SeekTable",
		Tags: []parser.Tag{
			{
				ID:       parser.TagID("FLAC:SeekTable:Points"),
				Name:     "SeekPoints",
				Value:    numPoints,
				DataType: "int64",
			},
		},
	}

	return dir
}

// parseCueSheet returns information about a cue sheet block.
// Cue sheets store track and index point information for CD media.
func (p *Parser) parseCueSheet(length int64) *parser.Directory {
	return &parser.Directory{
		Name: "FLAC-CueSheet",
		Tags: []parser.Tag{
			{
				ID:       parser.TagID("FLAC:CueSheet:Size"),
				Name:     "CueSheetSize",
				Value:    fmt.Sprintf("%d bytes", length),
				DataType: "string",
			},
		},
	}
}
