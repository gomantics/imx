package png

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// Image chunk tests

func TestParse_IHDRChunk_RGB(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(1920, 1080, 8, 2) // RGB
	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := dirs[0]

	// Check width
	widthTag := findTag(pngDir.Tags, "ImageWidth")
	if widthTag == nil {
		t.Fatal("ImageWidth tag not found")
	}
	if widthTag.Value != uint32(1920) {
		t.Errorf("ImageWidth = %v, want 1920", widthTag.Value)
	}

	// Check height
	heightTag := findTag(pngDir.Tags, "ImageHeight")
	if heightTag == nil {
		t.Fatal("ImageHeight tag not found")
	}
	if heightTag.Value != uint32(1080) {
		t.Errorf("ImageHeight = %v, want 1080", heightTag.Value)
	}

	// Check color type
	colorTag := findTag(pngDir.Tags, "ColorType")
	if colorTag == nil {
		t.Fatal("ColorType tag not found")
	}
	if colorTag.Value != "RGB" {
		t.Errorf("ColorType = %v, want RGB", colorTag.Value)
	}
}

func TestParse_IHDRChunk_ColorTypes(t *testing.T) {
	tests := []struct {
		colorType byte
		expected  string
	}{
		{0, "Grayscale"},
		{2, "RGB"},
		{3, "Palette"},
		{4, "Grayscale with Alpha"},
		{6, "RGB with Alpha"},
		{99, "Unknown (99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			var buf bytes.Buffer
			buf.Write(pngSignature)

			ihdr := createIHDR(10, 10, 8, tt.colorType)
			writeChunk(&buf, "IHDR", ihdr)
			writeChunk(&buf, "IEND", nil)

			r := bytes.NewReader(buf.Bytes())
			p := New()
			dirs, _ := p.Parse(r)

			colorTag := findTag(dirs[0].Tags, "ColorType")
			if colorTag.Value != tt.expected {
				t.Errorf("ColorType = %v, want %v", colorTag.Value, tt.expected)
			}
		})
	}
}

func TestParse_IHDRChunk_UnknownColorType(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	// Create IHDR with unknown color type (99)
	ihdr := createIHDR(1920, 1080, 8, 99) // Unknown color type
	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "ColorType")
	if tag == nil {
		t.Fatal("ColorType tag not found")
	}

	expected := "Unknown (99)"
	if tag.Value != expected {
		t.Errorf("ColorType = %q, want %q", tag.Value, expected)
	}
}

func TestParse_IHDRChunk_UnknownCompression(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	// Create IHDR with unknown compression method
	ihdr := make([]byte, ihdrChunkSize)
	binary.BigEndian.PutUint32(ihdr[ihdrWidthOffset:ihdrWidthOffset+4], 100)
	binary.BigEndian.PutUint32(ihdr[ihdrHeightOffset:ihdrHeightOffset+4], 100)
	ihdr[ihdrBitDepthOffset] = 8
	ihdr[ihdrColorTypeOffset] = colorTypeRGB
	ihdr[ihdrCompressionOffset] = 99 // Unknown compression
	ihdr[ihdrFilterOffset] = filterAdaptive
	ihdr[ihdrInterlaceOffset] = interlaceNone

	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "Compression")
	if tag == nil {
		t.Fatal("Compression tag not found")
	}

	expected := "Unknown (99)"
	if tag.Value != expected {
		t.Errorf("Compression = %q, want %q", tag.Value, expected)
	}
}

func TestParse_IHDRChunk_UnknownInterlace(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	// Create IHDR with unknown interlace method
	ihdr := make([]byte, ihdrChunkSize)
	binary.BigEndian.PutUint32(ihdr[ihdrWidthOffset:ihdrWidthOffset+4], 100)
	binary.BigEndian.PutUint32(ihdr[ihdrHeightOffset:ihdrHeightOffset+4], 100)
	ihdr[ihdrBitDepthOffset] = 8
	ihdr[ihdrColorTypeOffset] = colorTypeRGB
	ihdr[ihdrCompressionOffset] = compressionDeflate
	ihdr[ihdrFilterOffset] = filterAdaptive
	ihdr[ihdrInterlaceOffset] = 99 // Unknown interlace

	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "Interlace")
	if tag == nil {
		t.Fatal("Interlace tag not found")
	}

	expected := "Unknown (99)"
	if tag.Value != expected {
		t.Errorf("Interlace = %q, want %q", tag.Value, expected)
	}
}

func TestParse_IHDRChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading IHDR chunk data
	ihdrChunkOffset := int64(len(pngSignature) + 8)
	r := &errorReaderAt{data: data, errorAtOffset: ihdrChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should have empty or no PNG directory
	if len(dirs) > 0 {
		pngDir := findDir(dirs, "PNG")
		if pngDir != nil && len(pngDir.Tags) > 0 {
			t.Error("Expected no IHDR tags when read fails")
		}
	}
}

func TestParse_IHDRChunk_InvalidLength(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	// Create IHDR chunk with invalid length (12 bytes instead of 13)
	ihdr := make([]byte, 12)
	binary.BigEndian.PutUint32(ihdr[0:4], 100)
	binary.BigEndian.PutUint32(ihdr[4:8], 100)
	ihdr[8] = 8
	ihdr[9] = 2  // RGB
	ihdr[10] = 0 // Compression
	ihdr[11] = 0 // Filter

	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, _ := p.Parse(r)

	// Should have empty or no PNG directory due to invalid IHDR length
	if len(dirs) > 0 {
		pngDir := findDir(dirs, "PNG")
		if pngDir != nil && len(pngDir.Tags) > 0 {
			t.Error("Expected no IHDR tags when length is invalid")
		}
	}
}

func TestParse_IHDRChunk_UnknownFilter(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	// Create IHDR with unknown filter method
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], 100)
	binary.BigEndian.PutUint32(ihdr[4:8], 100)
	ihdr[8] = 8
	ihdr[9] = 2
	ihdr[10] = 0
	ihdr[11] = 99 // Unknown filter
	ihdr[12] = 0

	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "Filter")
	if tag == nil {
		t.Fatal("Filter tag not found")
	}

	expected := "Unknown (99)"
	if tag.Value != expected {
		t.Errorf("Filter = %q, want %q", tag.Value, expected)
	}
}

func TestParse_IHDRChunk_Adam7Interlace(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	// Create IHDR with Adam7 interlace
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], 100)
	binary.BigEndian.PutUint32(ihdr[4:8], 100)
	ihdr[8] = 8
	ihdr[9] = 2
	ihdr[10] = 0
	ihdr[11] = 0
	ihdr[12] = 1 // Adam7 interlace

	writeChunk(&buf, "IHDR", ihdr)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "Interlace")
	if tag == nil {
		t.Fatal("Interlace tag not found")
	}

	expected := "Adam7 Interlace"
	if tag.Value != expected {
		t.Errorf("Interlace = %q, want %q", tag.Value, expected)
	}
}

func TestParse_cHRMChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// cHRM chunk: 8 x uint32 values / 100000
	// White point, Red, Green, Blue (x,y for each)
	chromData := make([]byte, 32)
	binary.BigEndian.PutUint32(chromData[0:4], 31270)   // White X
	binary.BigEndian.PutUint32(chromData[4:8], 32900)   // White Y
	binary.BigEndian.PutUint32(chromData[8:12], 64000)  // Red X
	binary.BigEndian.PutUint32(chromData[12:16], 33000) // Red Y
	binary.BigEndian.PutUint32(chromData[16:20], 30000) // Green X
	binary.BigEndian.PutUint32(chromData[20:24], 60000) // Green Y
	binary.BigEndian.PutUint32(chromData[24:28], 15000) // Blue X
	binary.BigEndian.PutUint32(chromData[28:32], 6000)  // Blue Y

	writeChunk(&buf, "cHRM", chromData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	whiteXTag := findTag(pngDir.Tags, "WhitePointX")
	if whiteXTag == nil {
		t.Fatal("WhitePointX tag not found")
	}

	whiteX, ok := whiteXTag.Value.(float64)
	if !ok {
		t.Fatalf("WhitePointX type = %T, want float64", whiteXTag.Value)
	}

	expected := 31270.0 / 100000.0
	if whiteX < expected-0.001 || whiteX > expected+0.001 {
		t.Errorf("WhitePointX = %v, want ~%v", whiteX, expected)
	}
}

func TestParse_cHRMChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	chromData := make([]byte, 32)
	binary.BigEndian.PutUint32(chromData[0:4], 31270)
	binary.BigEndian.PutUint32(chromData[4:8], 32900)
	binary.BigEndian.PutUint32(chromData[8:12], 64000)
	binary.BigEndian.PutUint32(chromData[12:16], 33000)
	binary.BigEndian.PutUint32(chromData[16:20], 30000)
	binary.BigEndian.PutUint32(chromData[20:24], 60000)
	binary.BigEndian.PutUint32(chromData[24:28], 15000)
	binary.BigEndian.PutUint32(chromData[28:32], 6000)

	writeChunk(&buf, "cHRM", chromData)
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading cHRM chunk data
	chrmChunkOffset := int64(len(pngSignature) + 8 + 13 + 4 + 8)
	r := &errorReaderAt{data: data, errorAtOffset: chrmChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should still parse IHDR but skip cHRM
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Expected PNG directory from IHDR")
	}
	if findTag(pngDir.Tags, "WhitePointX") != nil {
		t.Error("Expected no cHRM tags when read fails")
	}
}

func TestParse_cHRMChunk_InvalidLength(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// cHRM chunk with wrong length (30 bytes instead of 32)
	chromData := make([]byte, 30)
	writeChunk(&buf, "cHRM", chromData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	// Should not have cHRM tags due to invalid length
	if findTag(pngDir.Tags, "WhitePointX") != nil {
		t.Error("WhitePointX tag found when length is invalid")
	}
}

func TestParse_gAMAChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// gAMA chunk: gamma as uint32 / 100000
	// gamma 2.2 = 220000
	gammaData := make([]byte, 4)
	binary.BigEndian.PutUint32(gammaData, 220000)
	writeChunk(&buf, "gAMA", gammaData)

	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "Gamma")
	if tag == nil {
		t.Fatal("Gamma tag not found")
	}

	gamma, ok := tag.Value.(float64)
	if !ok {
		t.Fatalf("Gamma value type = %T, want float64", tag.Value)
	}

	if gamma < 2.19 || gamma > 2.21 {
		t.Errorf("Gamma = %v, want ~2.2", gamma)
	}
}

func TestParse_gAMAChunk_InvalidLength(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// gAMA chunk with wrong length (3 bytes instead of 4)
	gammaData := []byte{0x01, 0x02, 0x03}
	writeChunk(&buf, "gAMA", gammaData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	// Should not have Gamma tag due to invalid length
	tag := findTag(pngDir.Tags, "Gamma")
	if tag != nil {
		t.Error("Gamma tag found when length is invalid")
	}
}

func TestParse_gAMAChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	gammaData := make([]byte, 4)
	binary.BigEndian.PutUint32(gammaData, 220000)
	writeChunk(&buf, "gAMA", gammaData)
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading gAMA chunk data
	gamaChunkOffset := int64(len(pngSignature) + 8 + 13 + 4 + 8)
	r := &errorReaderAt{data: data, errorAtOffset: gamaChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should still parse IHDR but skip gAMA
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Expected PNG directory from IHDR")
	}
	if findTag(pngDir.Tags, "Gamma") != nil {
		t.Error("Expected no Gamma tag when read fails")
	}
}

func TestParse_pHYsChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// pHYs chunk: pixels per unit X, Y (uint32 each), unit specifier (byte)
	physData := make([]byte, 9)
	binary.BigEndian.PutUint32(physData[0:4], 2835) // ~72 DPI
	binary.BigEndian.PutUint32(physData[4:8], 2835)
	physData[8] = 1 // Meters

	writeChunk(&buf, "pHYs", physData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	xTag := findTag(pngDir.Tags, "PixelsPerUnitX")
	if xTag == nil {
		t.Fatal("PixelsPerUnitX tag not found")
	}

	if xTag.Value != uint32(2835) {
		t.Errorf("PixelsPerUnitX = %v, want 2835", xTag.Value)
	}

	unitTag := findTag(pngDir.Tags, "PixelUnits")
	if unitTag == nil {
		t.Fatal("PixelUnits tag not found")
	}

	if unitTag.Value != "Meters" {
		t.Errorf("PixelUnits = %v, want Meters", unitTag.Value)
	}
}

func TestParse_pHYsChunk_UnspecifiedUnit(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// pHYs chunk with unit = 0 (unspecified)
	physData := make([]byte, 9)
	binary.BigEndian.PutUint32(physData[0:4], 1000)
	binary.BigEndian.PutUint32(physData[4:8], 1000)
	physData[8] = 0 // Unspecified unit

	writeChunk(&buf, "pHYs", physData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	unitTag := findTag(pngDir.Tags, "PixelUnits")
	if unitTag == nil {
		t.Fatal("PixelUnits tag not found")
	}

	if unitTag.Value != "Unspecified" {
		t.Errorf("PixelUnits = %v, want Unspecified", unitTag.Value)
	}
}

func TestParse_pHYsChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	physData := make([]byte, 9)
	binary.BigEndian.PutUint32(physData[0:4], 2835)
	binary.BigEndian.PutUint32(physData[4:8], 2835)
	physData[8] = 1

	writeChunk(&buf, "pHYs", physData)
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading pHYs chunk data
	physChunkOffset := int64(len(pngSignature) + 8 + 13 + 4 + 8)
	r := &errorReaderAt{data: data, errorAtOffset: physChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should still parse IHDR but skip pHYs
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Expected PNG directory from IHDR")
	}
	if findTag(pngDir.Tags, "PixelsPerUnitX") != nil {
		t.Error("Expected no pHYs tags when read fails")
	}
}

func TestParse_pHYsChunk_InvalidLength(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// pHYs chunk with wrong length (8 bytes instead of 9)
	physData := make([]byte, 8)
	writeChunk(&buf, "pHYs", physData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	// Should not have pHYs tags due to invalid length
	if findTag(pngDir.Tags, "PixelsPerUnitX") != nil {
		t.Error("PixelsPerUnitX tag found when length is invalid")
	}
}

func TestParse_pHYsChunk_UnknownUnit(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// pHYs chunk with unknown unit (not 0 or 1)
	physData := make([]byte, 9)
	binary.BigEndian.PutUint32(physData[0:4], 1000)
	binary.BigEndian.PutUint32(physData[4:8], 1000)
	physData[8] = 99 // Unknown unit

	writeChunk(&buf, "pHYs", physData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	unitTag := findTag(pngDir.Tags, "PixelUnits")
	if unitTag == nil {
		t.Fatal("PixelUnits tag not found")
	}

	if unitTag.Value != "Unknown" {
		t.Errorf("PixelUnits = %v, want Unknown", unitTag.Value)
	}
}

func TestParse_tIMEChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// tIME chunk: year (uint16), month, day, hour, minute, second (bytes)
	timeData := make([]byte, 7)
	binary.BigEndian.PutUint16(timeData[0:2], 2024)
	timeData[2] = 12 // Month
	timeData[3] = 25 // Day
	timeData[4] = 14 // Hour
	timeData[5] = 30 // Minute
	timeData[6] = 45 // Second

	writeChunk(&buf, "tIME", timeData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "ModifyDate")
	if tag == nil {
		t.Fatal("ModifyDate tag not found")
	}

	expected := "2024:12:25 14:30:45"
	if tag.Value != expected {
		t.Errorf("ModifyDate = %q, want %q", tag.Value, expected)
	}
}

func TestParse_tIMEChunk_InvalidLength(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// tIME chunk with wrong length (6 bytes instead of 7)
	timeData := make([]byte, 6)
	binary.BigEndian.PutUint16(timeData[0:2], 2024)
	timeData[2] = 12
	timeData[3] = 25
	timeData[4] = 14
	timeData[5] = 30

	writeChunk(&buf, "tIME", timeData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	// Should not have ModifyDate tag due to invalid length
	tag := findTag(pngDir.Tags, "ModifyDate")
	if tag != nil {
		t.Error("ModifyDate tag found when length is invalid")
	}
}

func TestParse_tIMEChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	timeData := make([]byte, 7)
	binary.BigEndian.PutUint16(timeData[0:2], 2024)
	timeData[2] = 12
	timeData[3] = 25
	timeData[4] = 14
	timeData[5] = 30
	timeData[6] = 45

	writeChunk(&buf, "tIME", timeData)
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading tIME chunk data
	timeChunkOffset := int64(len(pngSignature) + 8 + 13 + 4 + 8)
	r := &errorReaderAt{data: data, errorAtOffset: timeChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should still parse IHDR but skip tIME
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Expected PNG directory from IHDR")
	}
	if findTag(pngDir.Tags, "ModifyDate") != nil {
		t.Error("Expected no ModifyDate tag when read fails")
	}
}

func TestParse_bKGDChunk_Grayscale(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 0) // Grayscale
	writeChunk(&buf, "IHDR", ihdr)

	// bKGD chunk: 2 bytes for grayscale (16-bit)
	bgData := make([]byte, 2)
	binary.BigEndian.PutUint16(bgData, 32768) // Mid-gray

	writeChunk(&buf, "bKGD", bgData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "BackgroundColor")
	if tag == nil {
		t.Fatal("BackgroundColor tag not found")
	}

	if tag.Value != "32768" {
		t.Errorf("BackgroundColor = %v, want 32768", tag.Value)
	}
}

func TestParse_bKGDChunk_RGB(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2) // RGB
	writeChunk(&buf, "IHDR", ihdr)

	// bKGD chunk: 6 bytes for RGB (16-bit per channel)
	bgData := make([]byte, 6)
	binary.BigEndian.PutUint16(bgData[0:2], 65535) // R
	binary.BigEndian.PutUint16(bgData[2:4], 32768) // G
	binary.BigEndian.PutUint16(bgData[4:6], 0)     // B

	writeChunk(&buf, "bKGD", bgData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "BackgroundColor")
	if tag == nil {
		t.Fatal("BackgroundColor tag not found")
	}

	if tag.Value != "65535 32768 0" {
		t.Errorf("BackgroundColor = %v, want '65535 32768 0'", tag.Value)
	}
}

func TestParse_bKGDChunk_Palette(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 3) // Palette
	writeChunk(&buf, "IHDR", ihdr)

	// bKGD chunk: 1 byte for palette index
	bgData := []byte{5} // Palette index 5

	writeChunk(&buf, "bKGD", bgData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	tag := findTag(pngDir.Tags, "BackgroundColor")
	if tag == nil {
		t.Fatal("BackgroundColor tag not found")
	}

	if tag.Value != "5" {
		t.Errorf("BackgroundColor = %v, want '5'", tag.Value)
	}
}

func TestParse_bKGDChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	bgData := make([]byte, 6)
	binary.BigEndian.PutUint16(bgData[0:2], 65535)
	binary.BigEndian.PutUint16(bgData[2:4], 32768)
	binary.BigEndian.PutUint16(bgData[4:6], 0)

	writeChunk(&buf, "bKGD", bgData)
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading bKGD chunk data
	bkgdChunkOffset := int64(len(pngSignature) + 8 + 13 + 4 + 8)
	r := &errorReaderAt{data: data, errorAtOffset: bkgdChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should still parse IHDR but skip bKGD
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Expected PNG directory from IHDR")
	}
	if findTag(pngDir.Tags, "BackgroundColor") != nil {
		t.Error("Expected no BackgroundColor tag when read fails")
	}
}

func TestParse_bKGDChunk_InvalidLength(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// bKGD chunk with invalid length (4 bytes, not 1, 2, or 6)
	bgData := make([]byte, 4)
	writeChunk(&buf, "bKGD", bgData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	// With invalid length, tag may exist but with empty value
	tag := findTag(pngDir.Tags, "BackgroundColor")
	if tag != nil && tag.Value != "" {
		t.Errorf("BackgroundColor tag has value %q when length is invalid, want empty", tag.Value)
	}
}

func TestParse_bKGDChunk_EmptyChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// Empty bKGD chunk
	writeChunk(&buf, "bKGD", nil)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}

	// Should not have BackgroundColor tag due to empty chunk
	if findTag(pngDir.Tags, "BackgroundColor") != nil {
		t.Error("BackgroundColor tag found when chunk is empty")
	}
}
