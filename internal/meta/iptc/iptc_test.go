package iptc

import (
	"testing"

	"github.com/gomantics/imx/internal/format"
	"github.com/gomantics/imx/internal/meta"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
}

func TestParser_Spec(t *testing.T) {
	p := New()
	if p.Spec() != meta.SpecIPTC {
		t.Errorf("Spec() = %v, want %v", p.Spec(), meta.SpecIPTC)
	}
}

func TestParser_Parse_EmptyBlocks(t *testing.T) {
	p := New()
	dirs, err := p.Parse(nil)
	if err != nil {
		t.Errorf("Parse(nil) error = %v", err)
	}
	if dirs != nil {
		t.Errorf("Parse(nil) = %v, want nil", dirs)
	}
}

func TestParser_Parse_NonIPTCBlocks(t *testing.T) {
	p := New()
	blocks := []format.RawBlock{
		{
			Spec:    int(meta.SpecEXIF),
			Payload: []byte{1, 2, 3},
		},
	}
	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}
	if dirs != nil {
		t.Error("Parse() should return nil for non-IPTC blocks")
	}
}

func TestParser_Parse_ValidIPTC(t *testing.T) {
	p := New()

	// Build IPTC data
	iptcData := buildIPTCDataset(RecordApplication, 5, []byte("Test Title"))
	iptcData = append(iptcData, buildIPTCDataset(RecordApplication, 80, []byte("John Doe"))...)
	iptcData = append(iptcData, buildIPTCDataset(RecordApplication, 25, []byte("keyword1"))...)
	iptcData = append(iptcData, buildIPTCDataset(RecordApplication, 25, []byte("keyword2"))...)

	// Wrap in Photoshop IRB
	irbData := buildPhotoshopIRB(ResourceIPTC, iptcData)

	blocks := []format.RawBlock{
		{
			Spec:    int(meta.SpecIPTC),
			Payload: irbData,
			Origin:  "APP13 IPTC",
			Format:  format.FormatJPEG,
		},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() returned %d directories, want 1", len(dirs))
	}

	dir := dirs[0]
	if dir.Spec != meta.SpecIPTC {
		t.Errorf("dir.Spec = %v, want %v", dir.Spec, meta.SpecIPTC)
	}

	// Check for expected tags
	if _, ok := dir.Tags["IPTC:ObjectName"]; !ok {
		t.Error("Missing IPTC:ObjectName tag")
	}
	if _, ok := dir.Tags["IPTC:Byline"]; !ok {
		t.Error("Missing IPTC:Byline tag")
	}
	// Keywords should be indexed
	if _, ok := dir.Tags["IPTC:Keywords"]; !ok {
		t.Error("Missing IPTC:Keywords tag")
	}
	if _, ok := dir.Tags["IPTC:Keywords[1]"]; !ok {
		t.Error("Missing IPTC:Keywords[1] tag")
	}
}

func TestParser_Parse_RawIPTC(t *testing.T) {
	p := New()

	// Raw IPTC data without Photoshop wrapper
	iptcData := buildIPTCDataset(RecordApplication, 5, []byte("Direct Title"))

	blocks := []format.RawBlock{
		{
			Spec:    int(meta.SpecIPTC),
			Payload: iptcData,
		},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() returned %d directories, want 1", len(dirs))
	}

	if _, ok := dirs[0].Tags["IPTC:ObjectName"]; !ok {
		t.Error("Missing IPTC:ObjectName tag")
	}
}

func TestParser_Parse_MalformedIRB(t *testing.T) {
	p := New()

	blocks := []format.RawBlock{
		{
			Spec:    int(meta.SpecIPTC),
			Payload: []byte("invalid data"),
		},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}
	// Should return nil or empty for malformed data
	if len(dirs) != 0 {
		t.Errorf("Parse() should return empty for malformed data, got %d dirs", len(dirs))
	}
}

func TestParser_Parse_EnvelopeRecord(t *testing.T) {
	p := New()

	// Build envelope record data
	iptcData := buildIPTCDataset(RecordEnvelope, 70, []byte("20231215"))
	irbData := buildPhotoshopIRB(ResourceIPTC, iptcData)

	blocks := []format.RawBlock{
		{
			Spec:    int(meta.SpecIPTC),
			Payload: irbData,
		},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() returned %d directories, want 1", len(dirs))
	}

	if dirs[0].Name != "IPTC-Envelope" {
		t.Errorf("dir.Name = %q, want %q", dirs[0].Name, "IPTC-Envelope")
	}
}

func TestParser_Parse_MultipleBlocks(t *testing.T) {
	p := New()

	// First block
	iptc1 := buildIPTCDataset(RecordApplication, 5, []byte("Title 1"))
	irb1 := buildPhotoshopIRB(ResourceIPTC, iptc1)

	// Second block
	iptc2 := buildIPTCDataset(RecordApplication, 80, []byte("Author"))
	irb2 := buildPhotoshopIRB(ResourceIPTC, iptc2)

	blocks := []format.RawBlock{
		{Spec: int(meta.SpecIPTC), Payload: irb1},
		{Spec: int(meta.SpecIPTC), Payload: irb2},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// All datasets from same record should be in one directory
	if len(dirs) != 1 {
		t.Fatalf("Parse() returned %d directories, want 1", len(dirs))
	}

	if len(dirs[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(dirs[0].Tags))
	}
}

func TestParser_Parse_IntegerValue(t *testing.T) {
	p := New()

	// RecordVersion returns int
	iptcData := buildIPTCDataset(RecordApplication, 0, []byte{0x00, 0x04})
	irbData := buildPhotoshopIRB(ResourceIPTC, iptcData)

	blocks := []format.RawBlock{
		{Spec: int(meta.SpecIPTC), Payload: irbData},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(dirs) != 1 {
		t.Fatalf("Parse() returned %d directories, want 1", len(dirs))
	}

	tag, ok := dirs[0].Tags["IPTC:RecordVersion"]
	if !ok {
		t.Fatal("Missing IPTC:RecordVersion tag")
	}

	if tag.DataType != "int" {
		t.Errorf("tag.DataType = %q, want %q", tag.DataType, "int")
	}
	if tag.Value != 4 {
		t.Errorf("tag.Value = %v, want 4", tag.Value)
	}
}

func TestParser_BuildDirectories_Empty(t *testing.T) {
	p := New()
	dirs := p.buildDirectories(nil)
	if len(dirs) != 0 {
		t.Errorf("buildDirectories(nil) returned %d directories, want 0", len(dirs))
	}
}

func TestParser_BuildDirectories_MultipleRecords(t *testing.T) {
	p := New()
	datasets := []Dataset{
		{Record: RecordEnvelope, DatasetID: 70, Name: "DateSent", Value: "2023-12-15"},
		{Record: RecordApplication, DatasetID: 5, Name: "ObjectName", Value: "Title"},
	}

	dirs := p.buildDirectories(datasets)
	if len(dirs) != 2 {
		t.Errorf("buildDirectories() returned %d directories, want 2", len(dirs))
	}
}

func TestParser_Parse_IRBError(t *testing.T) {
	p := New()

	// Block with too short data to trigger IRB error, followed by valid block
	iptcData := buildIPTCDataset(RecordApplication, 5, []byte("Title"))
	irbData := buildPhotoshopIRB(ResourceIPTC, iptcData)

	blocks := []format.RawBlock{
		{
			Spec:    int(meta.SpecIPTC),
			Payload: []byte{1, 2, 3}, // Too short, triggers IRB error
		},
		{
			Spec:    int(meta.SpecIPTC),
			Payload: irbData, // Valid
		},
	}

	dirs, err := p.Parse(blocks)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should still get the valid block
	if len(dirs) != 1 {
		t.Errorf("Parse() returned %d directories, want 1", len(dirs))
	}
}
