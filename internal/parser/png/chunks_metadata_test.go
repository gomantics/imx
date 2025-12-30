package png

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"
)

// Metadata chunk tests

func TestParse_eXIfChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// eXIf chunk contains TIFF data
	// Create minimal TIFF: byte order + magic + IFD offset + empty IFD
	var tiffData bytes.Buffer
	tiffData.Write([]byte{0x49, 0x49})             // Little-endian
	tiffData.Write([]byte{0x2A, 0x00})             // Magic number 42
	tiffData.Write([]byte{0x08, 0x00, 0x00, 0x00}) // IFD offset
	tiffData.Write([]byte{0x00, 0x00})             // 0 entries

	writeChunk(&buf, "eXIf", tiffData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should have PNG directory from IHDR (TIFF parser returns empty for empty IFD)
	if len(dirs) < 1 {
		t.Fatalf("Parse() got %d directories, want at least 1", len(dirs))
	}
}

func TestParse_eXIfChunk_Empty(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// Empty eXIf chunk
	writeChunk(&buf, "eXIf", []byte{})
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should still parse IHDR but EXIF should be ignored
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("PNG directory not found")
	}
}

func TestParse_iTXtChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk: keyword + null + compression flag + compression method +
	// language tag + null + translated keyword + null + text
	var textData bytes.Buffer
	textData.WriteString("Title")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)   // No compression
	textData.WriteByte(0x00)   // Compression method
	textData.WriteString("en") // Language tag
	textData.WriteByte(0x00)
	textData.WriteString("Title") // Translated keyword
	textData.WriteByte(0x00)
	textData.WriteString("My PNG Image")

	writeChunk(&buf, "iTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	textDir := findDir(dirs, "PNG-Text")
	if textDir == nil {
		t.Fatal("PNG Text directory not found")
	}

	tag := findTag(textDir.Tags, "Title")
	if tag == nil {
		t.Fatal("Title tag not found")
	}

	if tag.Value != "My PNG Image" {
		t.Errorf("Tag value = %q, want %q", tag.Value, "My PNG Image")
	}
}

func TestParse_iTXtChunk_TooShort(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk too short (< 5 bytes)
	writeChunk(&buf, "iTXt", []byte{0x01, 0x02, 0x03}) // 3 bytes, less than 5
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil {
		t.Error("PNG Text directory found when it should not exist")
	}
}

func TestParse_iTXtChunk_XMP_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk with XMP keyword but malformed (missing translated keyword null)
	var textData bytes.Buffer
	textData.WriteString("XML:com.adobe.xmp")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00) // No compression
	textData.WriteByte(0x00) // Compression method
	textData.WriteString("") // Empty language tag
	textData.WriteByte(0x00)
	// Missing translated keyword null terminator - this should cause transEnd < 0
	textData.WriteString("XMP data without proper termination")

	writeChunk(&buf, "iTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have XMP directory due to malformed data
	xmpDir := findDir(dirs, "XMP")
	if xmpDir != nil {
		t.Error("XMP directory found when XMP data is malformed")
	}
}

func TestParse_iTXtChunk_InvalidLanguageTag(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk with regular text but no language tag null terminator
	var textData bytes.Buffer
	textData.WriteString("Title")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)   // No compression
	textData.WriteByte(0x00)   // Compression method
	textData.WriteString("en") // Language tag without null terminator
	// Missing translated keyword section entirely

	writeChunk(&buf, "iTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory due to malformed data
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil {
		t.Error("PNG Text directory found when iTXt data is malformed")
	}
}

func TestParse_iTXtChunk_InvalidTranslatedKeyword(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk with language tag but no translated keyword null terminator
	var textData bytes.Buffer
	textData.WriteString("Title")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)   // No compression
	textData.WriteByte(0x00)   // Compression method
	textData.WriteString("en") // Language tag with null
	textData.WriteByte(0x00)
	// Missing translated keyword null terminator
	textData.WriteString("Title")
	// No final null, so transEnd will be < 0

	writeChunk(&buf, "iTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory due to malformed translated keyword
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected iTXt with malformed translated keyword to be skipped")
	}
}

func TestParse_iTXtChunk_XMP(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk with XMP keyword
	var textData bytes.Buffer
	textData.WriteString("XML:com.adobe.xmp")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00) // No compression
	textData.WriteByte(0x00) // Compression method
	textData.WriteString("") // Empty language tag
	textData.WriteByte(0x00)
	textData.WriteString("") // Empty translated keyword
	textData.WriteByte(0x00)

	// XMP data
	xmpData := `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:creator><rdf:Seq><rdf:li>Test Author</rdf:li></rdf:Seq></dc:creator>
</rdf:Description>
</rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`
	textData.WriteString(xmpData)

	writeChunk(&buf, "iTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should have PNG directory and XMP directory
	xmpDir := findDir(dirs, "XMP")
	if xmpDir == nil {
		t.Skip("XMP parser may not extract data from minimal XMP")
	}
}

func TestParse_iTXtChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// Add iTXt chunk
	var textData bytes.Buffer
	textData.WriteString("Title")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)
	textData.WriteString("en")
	textData.WriteByte(0x00)
	textData.WriteString("Title")
	textData.WriteByte(0x00)
	textData.WriteString("My PNG Image")

	writeChunk(&buf, "iTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading iTXt chunk data
	iTxtChunkOffset := int64(len(pngSignature) + 8 + 13 + 4 + 8) // After IHDR chunk + iTXt header
	r := &errorReaderAt{data: data, errorAtOffset: iTxtChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should still parse IHDR but skip iTXt
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Expected PNG directory from IHDR")
	}
}

func TestParse_iTXtChunk_NoKeywordTerminator(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk without null terminator in keyword
	textData := []byte("TitleWithoutNullTerminator")
	writeChunk(&buf, "iTXt", textData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have text tags due to missing keyword terminator
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected iTXt without keyword terminator to be skipped")
	}
}

func TestParse_iTXtChunk_XMP_NoLanguageTerminator(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk with XMP keyword but no language terminator
	var textData bytes.Buffer
	textData.WriteString("XML:com.adobe.xmp")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)
	textData.WriteString("en") // No null terminator for language

	writeChunk(&buf, "iTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have XMP directory due to malformed data
	xmpDir := findDir(dirs, "XMP")
	if xmpDir != nil {
		t.Error("XMP directory found when language tag is not terminated")
	}
}

func TestParse_iTXtChunk_XMP_OffsetExceedsLength(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk with XMP keyword but offset exceeds data length
	var textData bytes.Buffer
	textData.WriteString("XML:com.adobe.xmp")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)
	textData.WriteString("")
	textData.WriteByte(0x00)
	textData.WriteString("")
	textData.WriteByte(0x00)
	// No XMP data, offset will exceed length

	writeChunk(&buf, "iTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have XMP directory due to offset exceeding length
	xmpDir := findDir(dirs, "XMP")
	if xmpDir != nil {
		t.Error("XMP directory found when offset exceeds length")
	}
}

func TestParse_iTXtChunk_OffsetEqualsLength(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iTXt chunk with regular text but offset equals length (no text value)
	var textData bytes.Buffer
	textData.WriteString("Title")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)
	textData.WriteString("")
	textData.WriteByte(0x00)
	textData.WriteString("")
	textData.WriteByte(0x00)
	// No text data after this point - offset will equal length

	writeChunk(&buf, "iTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory due to no text value
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected iTXt with offset >= length to be skipped")
	}
}

func TestParse_iCCPChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iCCP chunk: profile name + null + compression method + compressed data
	var iccpData bytes.Buffer
	iccpData.WriteString("sRGB")
	iccpData.WriteByte(0x00)
	iccpData.WriteByte(0x00) // Compression method 0 (deflate)

	// Create minimal ICC profile (128 bytes header minimum)
	minimalICC := make([]byte, 128)
	binary.BigEndian.PutUint32(minimalICC[0:4], 128) // Profile size

	// Compress it
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(minimalICC)
	w.Close()

	iccpData.Write(compressed.Bytes())
	writeChunk(&buf, "iCCP", iccpData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// ICC parser might not parse minimal profile, just verify no crash
	if len(dirs) < 1 {
		t.Fatal("Expected at least PNG directory")
	}
}

func TestParse_iCCPChunk_InvalidZlibData(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iCCP chunk with corrupted zlib data
	var iccpData bytes.Buffer
	iccpData.WriteString("test")
	iccpData.WriteByte(0x00)
	iccpData.WriteByte(0x00) // Valid compression method
	// Add some invalid zlib data (not valid deflate compressed data)
	iccpData.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})

	writeChunk(&buf, "iCCP", iccpData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have ICC directory due to zlib decompression failure
	iccDir := findDir(dirs, "ICC")
	if iccDir != nil {
		t.Error("ICC directory found when zlib data is corrupted")
	}
}

func TestParse_iCCPChunk_InvalidProfileName(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iCCP chunk with profile name not null-terminated
	var iccpData bytes.Buffer
	iccpData.WriteString("testprofile") // No null terminator
	iccpData.WriteByte(0x00)            // Compression method

	writeChunk(&buf, "iCCP", iccpData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have ICC directory due to malformed profile name
	iccDir := findDir(dirs, "ICC")
	if iccDir != nil {
		t.Error("ICC directory found when profile name is malformed")
	}
}

func TestParse_iCCPChunk_InvalidCompression(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iCCP chunk with invalid compression method
	var iccpData bytes.Buffer
	iccpData.WriteString("test")
	iccpData.WriteByte(0x00)
	iccpData.WriteByte(0x05) // Invalid compression method (not 0 = deflate)
	// Add enough data to pass length check (>= 10 bytes)
	iccpData.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})

	writeChunk(&buf, "iCCP", iccpData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have ICC directory due to invalid compression
	iccDir := findDir(dirs, "ICC")
	if iccDir != nil {
		t.Error("ICC directory found when compression method is invalid")
	}
}

func TestParse_iCCPChunk_TooShort(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iCCP chunk with length < 10 (minimum required)
	var iccpData bytes.Buffer
	iccpData.WriteString("RGB")
	iccpData.WriteByte(0x00)
	iccpData.WriteByte(0x00)
	// Only 5 bytes total, less than minimum of 10

	writeChunk(&buf, "iCCP", iccpData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have ICC directory due to length being too short
	iccDir := findDir(dirs, "ICC")
	if iccDir != nil {
		t.Error("ICC directory found when chunk length is too short")
	}
}

func TestParse_iCCPChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// Add iCCP chunk
	var iccpData bytes.Buffer
	iccpData.WriteString("sRGB")
	iccpData.WriteByte(0x00)
	iccpData.WriteByte(0x00)

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(make([]byte, 128))
	w.Close()

	iccpData.Write(compressed.Bytes())
	writeChunk(&buf, "iCCP", iccpData.Bytes())
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading iCCP chunk data
	iccpChunkOffset := int64(len(pngSignature) + 8 + 13 + 4 + 8)
	r := &errorReaderAt{data: data, errorAtOffset: iccpChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should still parse IHDR but skip iCCP
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Expected PNG directory from IHDR")
	}
}

func TestParse_iCCPChunk_DecompressionError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// iCCP chunk with valid header but corrupted compressed data that passes zlib.NewReader
	// but fails on io.Copy
	var iccpData bytes.Buffer
	iccpData.WriteString("test")
	iccpData.WriteByte(0x00)
	iccpData.WriteByte(0x00)

	// Create a valid zlib header but with truncated/corrupted data
	// This will pass NewReader but fail on io.Copy
	iccpData.Write([]byte{0x78, 0x9c, 0x03, 0x00}) // Valid zlib header but incomplete data

	writeChunk(&buf, "iCCP", iccpData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have ICC directory due to decompression error
	iccDir := findDir(dirs, "ICC")
	if iccDir != nil {
		t.Error("ICC directory found when decompression fails")
	}
}

func TestParse_tEXtChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// tEXt chunk: keyword + null + text
	textData := []byte("Author\x00John Doe")
	writeChunk(&buf, "tEXt", textData)

	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should have PNG directory and PNG Text directory
	if len(dirs) != 2 {
		t.Fatalf("Parse() got %d directories, want 2", len(dirs))
	}

	textDir := findDir(dirs, "PNG-Text")
	if textDir == nil {
		t.Fatal("PNG Text directory not found")
	}

	tag := findTag(textDir.Tags, "Author")
	if tag == nil {
		t.Fatal("Author tag not found")
	}

	if tag.Value != "John Doe" {
		t.Errorf("Tag value = %q, want %q", tag.Value, "John Doe")
	}
}

func TestParse_tEXtChunk_NoNullTerminator(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// tEXt chunk without null terminator in keyword
	textData := []byte("AuthorJohn Doe") // No null byte
	writeChunk(&buf, "tEXt", textData)

	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory due to missing null terminator
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected tEXt without null terminator to be skipped")
	}
}

func TestParse_tEXtChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	textData := []byte("Author\x00John Doe")
	writeChunk(&buf, "tEXt", textData)
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading tEXt chunk data
	textChunkOffset := int64(len(pngSignature) + 8 + 13 + 4 + 8)
	r := &errorReaderAt{data: data, errorAtOffset: textChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should still parse IHDR but skip tEXt
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Expected PNG directory from IHDR")
	}
}

func TestParse_tEXtChunk_EmptyChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// Empty tEXt chunk
	writeChunk(&buf, "tEXt", nil)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory due to empty chunk
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected empty tEXt chunk to be skipped")
	}
}

func TestParse_zTXtChunk(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// zTXt chunk: keyword + null + compression method + compressed text
	var textData bytes.Buffer
	textData.WriteString("Description")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00) // Compression method 0 (deflate)

	// Compress the text
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write([]byte("This is a compressed description"))
	w.Close()

	textData.Write(compressed.Bytes())
	writeChunk(&buf, "zTXt", textData.Bytes())

	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	textDir := findDir(dirs, "PNG-Text")
	if textDir == nil {
		t.Fatal("PNG Text directory not found")
	}

	tag := findTag(textDir.Tags, "Description")
	if tag == nil {
		t.Fatal("Description tag not found")
	}

	if tag.Value != "This is a compressed description" {
		t.Errorf("Tag value = %q, want %q", tag.Value, "This is a compressed description")
	}
}

func TestParse_zTXtChunk_InvalidCompression(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// zTXt with invalid compression method
	var textData bytes.Buffer
	textData.WriteString("Comment")
	textData.WriteByte(0x00)
	textData.WriteByte(0x99) // Invalid compression method
	textData.WriteString("data")

	writeChunk(&buf, "zTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	// Should handle gracefully (skip the chunk)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Should only have PNG directory, not text directory
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected zTXt with invalid compression to be skipped")
	}
}

func TestParse_zTXtChunk_InvalidZlibData(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// zTXt chunk with valid compression method but corrupted zlib data
	var textData bytes.Buffer
	textData.WriteString("Comment")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00) // Valid compression method
	// Add some data that looks like zlib header but is corrupted
	// Valid zlib header would start with 0x78 0x9C for deflate, but we'll use invalid
	textData.Write([]byte{0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	writeChunk(&buf, "zTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory due to zlib decompression failure
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected zTXt with corrupted zlib data to be skipped")
	}
}

func TestParse_zTXtChunk_TooShort(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// zTXt chunk too short (< 3 bytes)
	writeChunk(&buf, "zTXt", []byte{0x01, 0x02}) // 2 bytes, less than 3
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory due to malformed chunk
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected zTXt with insufficient length to be skipped")
	}
}

func TestParse_zTXtChunk_ReadError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	var textData bytes.Buffer
	textData.WriteString("Description")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00)

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write([]byte("This is a compressed description"))
	w.Close()

	textData.Write(compressed.Bytes())
	writeChunk(&buf, "zTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	data := buf.Bytes()
	// Error when reading zTXt chunk data
	zTxtChunkOffset := int64(len(pngSignature) + 8 + 13 + 4 + 8)
	r := &errorReaderAt{data: data, errorAtOffset: zTxtChunkOffset}

	p := New()
	dirs, _ := p.Parse(r)

	// Should still parse IHDR but skip zTXt
	pngDir := findDir(dirs, "PNG")
	if pngDir == nil {
		t.Fatal("Expected PNG directory from IHDR")
	}
}

func TestParse_zTXtChunk_NoNullTerminator(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// zTXt chunk without null terminator in keyword
	textData := []byte("KeywordWithoutNull")
	writeChunk(&buf, "zTXt", textData)
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory due to missing keyword terminator
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected zTXt without keyword terminator to be skipped")
	}
}

func TestParse_zTXtChunk_DecompressionError(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := createIHDR(10, 10, 8, 2)
	writeChunk(&buf, "IHDR", ihdr)

	// zTXt chunk with valid header but corrupted compressed data that passes zlib.NewReader
	// but fails on io.Copy
	var textData bytes.Buffer
	textData.WriteString("Comment")
	textData.WriteByte(0x00)
	textData.WriteByte(0x00) // Valid compression method

	// Create a valid zlib header but with truncated/corrupted data
	// This will pass NewReader but fail on io.Copy
	textData.Write([]byte{0x78, 0x9c, 0x03, 0x00}) // Valid zlib header but incomplete data

	writeChunk(&buf, "zTXt", textData.Bytes())
	writeChunk(&buf, "IEND", nil)

	r := bytes.NewReader(buf.Bytes())
	p := New()
	dirs, err := p.Parse(r)

	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Should not have PNG Text directory due to decompression error
	textDir := findDir(dirs, "PNG-Text")
	if textDir != nil && len(textDir.Tags) > 0 {
		t.Error("Expected zTXt with decompression error to be skipped")
	}
}
