package imx

import (
	"encoding/binary"
	"fmt"
	"io"
)

var compressionNames = map[uint32]string{
	0: "BI_RGB",
	1: "BI_RLE8",
	2: "BI_RLE4",
	3: "BI_BITFIELDS",
	4: "BI_JPEG",
	5: "BI_PNG",
}

// ExtractBMPMetadata extracts metadata from a BMP file.
func ExtractBMPMetadata(r io.ReadSeeker, md *ImageMetadata) error {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Read BMP file header (14 bytes)
	var fileHeader [14]byte
	if _, err = io.ReadFull(r, fileHeader[:]); err != nil {
		return fmt.Errorf("failed to read BMP file header: %w", err)
	}

	// Verify BMP signature
	if fileHeader[0] != 0x42 || fileHeader[1] != 0x4D {
		return fmt.Errorf("invalid BMP file")
	}

	// Read file size (offset 2, 4 bytes, little-endian)
	fileSize := binary.LittleEndian.Uint32(fileHeader[2:6])
	md.setAdditional("FileSizeFromHeader", fileSize)

	// Read offset to pixel data (offset 10, 4 bytes, little-endian)
	dataOffset := binary.LittleEndian.Uint32(fileHeader[10:14])
	md.setAdditional("DataOffset", dataOffset)

	// Read DIB header size (4 bytes, little-endian)
	var dibSizeBytes [4]byte
	if _, err = io.ReadFull(r, dibSizeBytes[:]); err != nil {
		return fmt.Errorf("failed to read DIB header size: %w", err)
	}
	dibSize := binary.LittleEndian.Uint32(dibSizeBytes[:])

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
		if _, err = io.ReadFull(r, dibHeader); err != nil {
			return fmt.Errorf("failed to read DIB header: %w", err)
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

		md.Width = int(width)
		// Height can be negative (top-down DIB)
		if height < 0 {
			md.Height = int(-height)
			md.setAdditional("TopDown", true)
		} else {
			md.Height = int(height)
			md.setAdditional("TopDown", false)
		}

		md.ColorDepth = int(bitsPerPixel)
		md.setAdditional("Planes", planes)
		md.setAdditional("Compression", compression)
		md.setAdditional("ImageSize", imageSize)
		md.setAdditional("XPixelsPerMeter", xPixelsPerM)
		md.setAdditional("YPixelsPerMeter", yPixelsPerM)
		md.setAdditional("ColorsUsed", colorsUsed)
		md.setAdditional("ColorsImportant", colorsImportant)

		// Determine color space based on bits per pixel
		switch bitsPerPixel {
		case 1:
			md.ColorSpace = "Indexed"
		case 4:
			md.ColorSpace = "Indexed"
		case 8:
			md.ColorSpace = "Indexed"
		case 16:
			md.ColorSpace = "RGB"
		case 24:
			md.ColorSpace = "RGB"
		case 32:
			md.ColorSpace = "RGBA"
		default:
			md.ColorSpace = "Unknown"
		}

		// Compression types
		if name, ok := compressionNames[compression]; ok {
			md.setAdditional("CompressionName", name)
		}

	} else if dibSize == 12 {
		// BITMAPCOREHEADER (12 bytes)
		dibHeader := make([]byte, 8)
		if _, err = io.ReadFull(r, dibHeader); err != nil {
			return fmt.Errorf("failed to read DIB header: %w", err)
		}

		width = int32(int16(binary.LittleEndian.Uint16(dibHeader[0:2])))
		height = int32(int16(binary.LittleEndian.Uint16(dibHeader[2:4])))
		planes = binary.LittleEndian.Uint16(dibHeader[4:6])
		bitsPerPixel = binary.LittleEndian.Uint16(dibHeader[6:8])

		md.Width = int(width)
		md.Height = int(height)
		md.ColorDepth = int(bitsPerPixel)
		md.setAdditional("Planes", planes)
		md.ColorSpace = "RGB"
	} else {
		return fmt.Errorf("unsupported DIB header size: %d", dibSize)
	}

	return nil
}
