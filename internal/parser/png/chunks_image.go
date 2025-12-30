package png

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
)

// parseIHDRChunk parses an IHDR chunk (image header)
func (p *Parser) parseIHDRChunk(r io.ReaderAt, chunk *Chunk) []parser.Tag {
	if chunk.Length != ihdrChunkSize {
		return nil
	}

	buf := make([]byte, ihdrChunkSize)
	_, err := r.ReadAt(buf, chunk.DataOffset)
	if err != nil {
		return nil
	}

	// IHDR structure:
	// - Width (4 bytes)
	// - Height (4 bytes)
	// - Bit depth (1 byte)
	// - Color type (1 byte)
	// - Compression method (1 byte)
	// - Filter method (1 byte)
	// - Interlace method (1 byte)

	width := binary.BigEndian.Uint32(buf[ihdrWidthOffset : ihdrWidthOffset+4])
	height := binary.BigEndian.Uint32(buf[ihdrHeightOffset : ihdrHeightOffset+4])
	bitDepth := buf[ihdrBitDepthOffset]
	colorType := buf[ihdrColorTypeOffset]
	compression := buf[ihdrCompressionOffset]
	filter := buf[ihdrFilterOffset]
	interlace := buf[ihdrInterlaceOffset]

	colorTypeStr := map[byte]string{
		colorTypeGrayscale:      "Grayscale",
		colorTypeRGB:            "RGB",
		colorTypePalette:        "Palette",
		colorTypeGrayscaleAlpha: "Grayscale with Alpha",
		colorTypeRGBA:           "RGB with Alpha",
	}

	colorTypeVal, ok := colorTypeStr[colorType]
	if !ok {
		colorTypeVal = fmt.Sprintf("Unknown (%d)", colorType)
	}

	compressionStr := "Deflate/Inflate"
	if compression != compressionDeflate {
		compressionStr = fmt.Sprintf("Unknown (%d)", compression)
	}

	filterStr := "Adaptive"
	if filter != filterAdaptive {
		filterStr = fmt.Sprintf("Unknown (%d)", filter)
	}

	interlaceStr := "Noninterlaced"
	if interlace == interlaceAdam7 {
		interlaceStr = "Adam7 Interlace"
	} else if interlace != interlaceNone {
		interlaceStr = fmt.Sprintf("Unknown (%d)", interlace)
	}

	return []parser.Tag{
		{
			ID:       "PNG:ImageWidth",
			Name:     "ImageWidth",
			Value:    width,
			DataType: "uint32",
		},
		{
			ID:       "PNG:ImageHeight",
			Name:     "ImageHeight",
			Value:    height,
			DataType: "uint32",
		},
		{
			ID:       "PNG:BitDepth",
			Name:     "BitDepth",
			Value:    bitDepth,
			DataType: "uint8",
		},
		{
			ID:       "PNG:ColorType",
			Name:     "ColorType",
			Value:    colorTypeVal,
			DataType: "string",
		},
		{
			ID:       "PNG:Compression",
			Name:     "Compression",
			Value:    compressionStr,
			DataType: "string",
		},
		{
			ID:       "PNG:Filter",
			Name:     "Filter",
			Value:    filterStr,
			DataType: "string",
		},
		{
			ID:       "PNG:Interlace",
			Name:     "Interlace",
			Value:    interlaceStr,
			DataType: "string",
		},
	}
}

// parsecHRMChunk parses a cHRM chunk (chromaticity)
func (p *Parser) parsecHRMChunk(r io.ReaderAt, chunk *Chunk) []parser.Tag {
	if chunk.Length != chrmChunkSize {
		return nil
	}

	data := make([]byte, chrmChunkSize)
	_, err := r.ReadAt(data, chunk.DataOffset)
	if err != nil {
		return nil
	}

	// cHRM stores values as integers that need to be divided by 100000
	whiteX := float64(binary.BigEndian.Uint32(data[0:4])) / chrmScale
	whiteY := float64(binary.BigEndian.Uint32(data[4:8])) / chrmScale
	redX := float64(binary.BigEndian.Uint32(data[8:12])) / chrmScale
	redY := float64(binary.BigEndian.Uint32(data[12:16])) / chrmScale
	greenX := float64(binary.BigEndian.Uint32(data[16:20])) / chrmScale
	greenY := float64(binary.BigEndian.Uint32(data[20:24])) / chrmScale
	blueX := float64(binary.BigEndian.Uint32(data[24:28])) / chrmScale
	blueY := float64(binary.BigEndian.Uint32(data[28:32])) / chrmScale

	return []parser.Tag{
		{ID: "PNG:WhitePointX", Name: "WhitePointX", Value: whiteX, DataType: "float64"},
		{ID: "PNG:WhitePointY", Name: "WhitePointY", Value: whiteY, DataType: "float64"},
		{ID: "PNG:RedX", Name: "RedX", Value: redX, DataType: "float64"},
		{ID: "PNG:RedY", Name: "RedY", Value: redY, DataType: "float64"},
		{ID: "PNG:GreenX", Name: "GreenX", Value: greenX, DataType: "float64"},
		{ID: "PNG:GreenY", Name: "GreenY", Value: greenY, DataType: "float64"},
		{ID: "PNG:BlueX", Name: "BlueX", Value: blueX, DataType: "float64"},
		{ID: "PNG:BlueY", Name: "BlueY", Value: blueY, DataType: "float64"},
	}
}

// parsegAMAChunk parses a gAMA chunk (gamma)
func (p *Parser) parsegAMAChunk(r io.ReaderAt, chunk *Chunk) *parser.Tag {
	if chunk.Length != gamaChunkSize {
		return nil
	}

	data := make([]byte, gamaChunkSize)
	_, err := r.ReadAt(data, chunk.DataOffset)
	if err != nil {
		return nil
	}

	// Gamma is stored as integer / 100000
	gamma := float64(binary.BigEndian.Uint32(data)) / gamaScale

	return &parser.Tag{
		ID:       "PNG:Gamma",
		Name:     "Gamma",
		Value:    gamma,
		DataType: "float64",
	}
}

// parsepHYsChunk parses a pHYs chunk (physical dimensions)
func (p *Parser) parsepHYsChunk(r io.ReaderAt, chunk *Chunk) []parser.Tag {
	if chunk.Length != physChunkSize {
		return nil
	}

	data := make([]byte, physChunkSize)
	_, err := r.ReadAt(data, chunk.DataOffset)
	if err != nil {
		return nil
	}

	pixelsPerUnitX := binary.BigEndian.Uint32(data[physPixelsXOffset : physPixelsXOffset+4])
	pixelsPerUnitY := binary.BigEndian.Uint32(data[physPixelsYOffset : physPixelsYOffset+4])
	unit := data[physUnitOffset]

	unitStr := "Unknown"
	if unit == physUnitUnknown {
		unitStr = "Unspecified"
	} else if unit == physUnitMeter {
		unitStr = "Meters"
	}

	return []parser.Tag{
		{ID: "PNG:PixelsPerUnitX", Name: "PixelsPerUnitX", Value: pixelsPerUnitX, DataType: "uint32"},
		{ID: "PNG:PixelsPerUnitY", Name: "PixelsPerUnitY", Value: pixelsPerUnitY, DataType: "uint32"},
		{ID: "PNG:PixelUnits", Name: "PixelUnits", Value: unitStr, DataType: "string"},
	}
}

// parsetIMEChunk parses a tIME chunk (modification time)
func (p *Parser) parsetIMEChunk(r io.ReaderAt, chunk *Chunk) *parser.Tag {
	if chunk.Length != timeChunkSize {
		return nil
	}

	data := make([]byte, timeChunkSize)
	_, err := r.ReadAt(data, chunk.DataOffset)
	if err != nil {
		return nil
	}

	year := binary.BigEndian.Uint16(data[timeYearOffset : timeYearOffset+2])
	month := data[timeMonthOffset]
	day := data[timeDayOffset]
	hour := data[timeHourOffset]
	minute := data[timeMinOffset]
	second := data[timeSecOffset]

	timeStr := fmt.Sprintf("%04d:%02d:%02d %02d:%02d:%02d", year, month, day, hour, minute, second)

	return &parser.Tag{
		ID:       "PNG:ModifyDate",
		Name:     "ModifyDate",
		Value:    timeStr,
		DataType: "string",
	}
}

// parsebKGDChunk parses a bKGD chunk (background color)
func (p *Parser) parsebKGDChunk(r io.ReaderAt, chunk *Chunk) *parser.Tag {
	if chunk.Length == 0 {
		return nil
	}

	data := make([]byte, chunk.Length)
	_, err := r.ReadAt(data, chunk.DataOffset)
	if err != nil {
		return nil
	}

	// Background color format depends on color type (stored in IHDR)
	// For simplicity, we'll store the raw values
	var value string
	if chunk.Length == bkgdGrayscaleSize {
		// Grayscale or palette index
		value = fmt.Sprintf("%d", data[0])
	} else if chunk.Length == bkgdGrayscale16Size {
		// Grayscale (16-bit)
		value = fmt.Sprintf("%d", binary.BigEndian.Uint16(data))
	} else if chunk.Length == bkgdRGBSize {
		// RGB (16-bit per channel)
		r := binary.BigEndian.Uint16(data[0:2])
		g := binary.BigEndian.Uint16(data[2:4])
		b := binary.BigEndian.Uint16(data[4:6])
		value = fmt.Sprintf("%d %d %d", r, g, b)
	}

	return &parser.Tag{
		ID:       "PNG:BackgroundColor",
		Name:     "BackgroundColor",
		Value:    value,
		DataType: "string",
	}
}
