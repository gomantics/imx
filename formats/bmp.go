package formats

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ExtractBMP extracts metadata from a BMP file.
func ExtractBMP(r io.ReadSeeker) (*Result, error) {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Read BMP file header (14 bytes)
	fileHeader := make([]byte, 14)
	_, err = r.Read(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to read BMP file header: %w", err)
	}

	// Verify BMP signature
	if fileHeader[0] != 0x42 || fileHeader[1] != 0x4D {
		return nil, fmt.Errorf("%w: invalid BMP file", ErrInvalidData)
	}

	result := newResult()

	// Read file size (offset 2, 4 bytes, little-endian)
	fileSize := binary.LittleEndian.Uint32(fileHeader[2:6])
	result.Additional["FileSizeFromHeader"] = fileSize

	// Read offset to pixel data (offset 10, 4 bytes, little-endian)
	dataOffset := binary.LittleEndian.Uint32(fileHeader[10:14])
	result.Additional["DataOffset"] = dataOffset

	// Read DIB header size (4 bytes, little-endian)
	dibSizeBytes := make([]byte, 4)
	_, err = r.Read(dibSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read DIB header size: %w", err)
	}
	dibSize := binary.LittleEndian.Uint32(dibSizeBytes)

	// Read DIB header based on size
	var width, height int32
	var planes, bitsPerPixel uint16
	var compression uint32
	var imageSize uint32
	var xPixelsPerM, yPixelsPerM uint32
	var colorsUsed, colorsImportant uint32

	if dibSize >= 40 {
		// BITMAPINFOHEADER (40 bytes) or larger
		dibHeader := make([]byte, 36) // Read remaining 36 bytes of 40-byte header
		_, err = r.Read(dibHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to read DIB header: %w", err)
		}

		width = int32(binary.LittleEndian.Uint32(dibHeader[0:4]))
		height = int32(binary.LittleEndian.Uint32(dibHeader[4:8]))
		planes = binary.LittleEndian.Uint16(dibHeader[8:10])
		bitsPerPixel = binary.LittleEndian.Uint16(dibHeader[10:12])
		compression = binary.LittleEndian.Uint32(dibHeader[12:16])
		imageSize = binary.LittleEndian.Uint32(dibHeader[16:20])
		xPixelsPerM = binary.LittleEndian.Uint32(dibHeader[20:24])
		yPixelsPerM = binary.LittleEndian.Uint32(dibHeader[24:28])
		colorsUsed = binary.LittleEndian.Uint32(dibHeader[28:32])
		colorsImportant = binary.LittleEndian.Uint32(dibHeader[32:36])

		result.Width = int(width)
		// Height can be negative (top-down DIB)
		if height < 0 {
			result.Height = int(-height)
			result.Additional["TopDown"] = true
		} else {
			result.Height = int(height)
			result.Additional["TopDown"] = false
		}

		result.ColorDepth = int(bitsPerPixel)
		result.Additional["Planes"] = planes
		result.Additional["Compression"] = compression
		result.Additional["ImageSize"] = imageSize
		result.Additional["XPixelsPerMeter"] = xPixelsPerM
		result.Additional["YPixelsPerMeter"] = yPixelsPerM
		result.Additional["ColorsUsed"] = colorsUsed
		result.Additional["ColorsImportant"] = colorsImportant

		// Determine color space based on bits per pixel
		switch bitsPerPixel {
		case 1:
			result.ColorSpace = "Indexed"
		case 4:
			result.ColorSpace = "Indexed"
		case 8:
			result.ColorSpace = "Indexed"
		case 16:
			result.ColorSpace = "RGB"
		case 24:
			result.ColorSpace = "RGB"
		case 32:
			result.ColorSpace = "RGBA"
		default:
			result.ColorSpace = "Unknown"
		}

		// Compression types
		compressionNames := map[uint32]string{
			0: "BI_RGB",
			1: "BI_RLE8",
			2: "BI_RLE4",
			3: "BI_BITFIELDS",
			4: "BI_JPEG",
			5: "BI_PNG",
		}
		if name, ok := compressionNames[compression]; ok {
			result.Additional["CompressionName"] = name
		}

	} else if dibSize == 12 {
		// BITMAPCOREHEADER (12 bytes)
		dibHeader := make([]byte, 8)
		_, err = r.Read(dibHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to read DIB header: %w", err)
		}

		width = int32(int16(binary.LittleEndian.Uint16(dibHeader[0:2])))
		height = int32(int16(binary.LittleEndian.Uint16(dibHeader[2:4])))
		planes = binary.LittleEndian.Uint16(dibHeader[4:6])
		bitsPerPixel = binary.LittleEndian.Uint16(dibHeader[6:8])

		result.Width = int(width)
		result.Height = int(height)
		result.ColorDepth = int(bitsPerPixel)
		result.Additional["Planes"] = planes
		result.ColorSpace = "RGB"
	} else {
		return nil, fmt.Errorf("%w: unsupported DIB header size %d", ErrInvalidData, dibSize)
	}

	return result, nil
}
