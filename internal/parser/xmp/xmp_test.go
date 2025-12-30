package xmp

import (
	"bytes"
	"io"
	"testing"

	"github.com/gomantics/imx/internal/parser"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.handlers == nil {
		t.Error("New() created parser with nil handlers")
	}
	// Verify all handlers are registered
	expectedHandlers := []ContextType{
		CTX_ROOT, CTX_RDF, CTX_DESCRIPTION,
		CTX_PROPERTY, CTX_ARRAY, CTX_LI, CTX_STRUCT_FIELD,
	}
	for _, ctx := range expectedHandlers {
		if _, ok := p.handlers[ctx]; !ok {
			t.Errorf("Handler for %v not registered", ctx)
		}
	}
}

func TestParser_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "XMP" {
		t.Errorf("Name() = %q, want %q", got, "XMP")
	}
}

func TestParser_Detect(t *testing.T) {
	// Note: Detect() reads first 100 bytes, so test data must be at least 100 bytes
	// to avoid read errors, or we test the error path
	makeData := func(prefix string) []byte {
		data := make([]byte, 100)
		copy(data, prefix)
		return data
	}

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid xpacket",
			data: makeData(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>`),
			want: true,
		},
		{
			name: "valid x:xmpmeta",
			data: makeData(`<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF></rdf:RDF>`),
			want: true,
		},
		{
			name: "xpacket in middle of buffer",
			data: makeData(`<!-- comment --><?xpacket begin=""`),
			want: true,
		},
		{
			name: "invalid - no XMP markers",
			data: makeData(`<html><body>Not XMP</body></html>`),
			want: false,
		},
		{
			name: "invalid - random bytes",
			data: makeData(string([]byte{0x00, 0x01, 0x02, 0x03, 0x04})),
			want: false,
		},
		{
			name: "too short - triggers read error",
			data: []byte("abc"),
			want: false,
		},
		{
			name: "empty - triggers read error",
			data: []byte{},
			want: false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			if got := p.Detect(r); got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantDirs bool
		wantErr  bool
	}{
		{
			name: "simple XMP with property",
			data: []byte(`<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/" dc:title="Test"/>
</rdf:RDF>
</x:xmpmeta>`),
			wantDirs: true,
			wantErr:  false,
		},
		{
			name: "XMP with nested property",
			data: []byte(`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:creator>Author Name</dc:creator>
</rdf:Description>
</rdf:RDF>`),
			wantDirs: true,
			wantErr:  false,
		},
		{
			name: "XMP with array",
			data: []byte(`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:subject>
<rdf:Bag>
<rdf:li>keyword1</rdf:li>
<rdf:li>keyword2</rdf:li>
</rdf:Bag>
</dc:subject>
</rdf:Description>
</rdf:RDF>`),
			wantDirs: true,
			wantErr:  false,
		},
		{
			name: "XMP with Seq array",
			data: []byte(`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:creator>
<rdf:Seq>
<rdf:li>Author 1</rdf:li>
<rdf:li>Author 2</rdf:li>
</rdf:Seq>
</dc:creator>
</rdf:Description>
</rdf:RDF>`),
			wantDirs: true,
			wantErr:  false,
		},
		{
			name: "XMP with Alt array",
			data: []byte(`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:title>
<rdf:Alt>
<rdf:li xml:lang="en">English Title</rdf:li>
</rdf:Alt>
</dc:title>
</rdf:Description>
</rdf:RDF>`),
			wantDirs: true,
			wantErr:  false,
		},
		{
			name: "XMP with struct",
			data: []byte(`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/">
<xmpMM:History>
<rdf:Seq>
<rdf:li rdf:parseType="Resource">
<stEvt:action xmlns:stEvt="http://ns.adobe.com/xap/1.0/sType/ResourceEvent#">created</stEvt:action>
</rdf:li>
</rdf:Seq>
</xmpMM:History>
</rdf:Description>
</rdf:RDF>`),
			wantDirs: true,
			wantErr:  false,
		},
		{
			name:     "empty XML",
			data:     []byte(`<?xml version="1.0"?>`),
			wantDirs: false,
			wantErr:  false,
		},
		{
			name:     "invalid XML",
			data:     []byte(`<unclosed`),
			wantDirs: false,
			wantErr:  true,
		},
		{
			name:     "empty data",
			data:     []byte{},
			wantDirs: false,
			wantErr:  false,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			dirs, parseErr := p.Parse(r)

			hasErr := parseErr != nil && parseErr.Error() != ""
			if hasErr != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", parseErr, tt.wantErr)
			}
			if (len(dirs) > 0) != tt.wantDirs {
				t.Errorf("Parse() dirs = %d, wantDirs %v", len(dirs), tt.wantDirs)
			}
		})
	}
}

func TestParser_parsePacket_Errors(t *testing.T) {
	tests := []struct {
		name       string
		nodeMap    NodeMap
		namespaces map[string]string
		wantErr    bool
	}{
		{
			name:       "nil nodeMap",
			nodeMap:    nil,
			namespaces: make(map[string]string),
			wantErr:    true,
		},
		{
			name:       "nil namespaces",
			nodeMap:    make(NodeMap),
			namespaces: nil,
			wantErr:    true,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(`<root/>`))
			err := p.parsePacket(r, tt.nodeMap, tt.namespaces)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePacket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReaderAtWrapper(t *testing.T) {
	data := []byte("Hello, World!")
	r := bytes.NewReader(data)
	wrapper := &readerAtWrapper{r: r, offset: 0}

	// Read first chunk
	buf := make([]byte, 5)
	n, err := wrapper.Read(buf)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
	if n != 5 {
		t.Errorf("Read() n = %d, want 5", n)
	}
	if string(buf) != "Hello" {
		t.Errorf("Read() got %q, want %q", buf, "Hello")
	}

	// Read second chunk
	buf = make([]byte, 8)
	n, err = wrapper.Read(buf)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
	if n != 8 {
		t.Errorf("Read() n = %d, want 8", n)
	}
	if string(buf) != ", World!" {
		t.Errorf("Read() got %q, want %q", buf, ", World!")
	}

	// Read past end
	buf = make([]byte, 10)
	n, err = wrapper.Read(buf)
	if err != io.EOF {
		t.Errorf("Read() error = %v, want io.EOF", err)
	}
}

func TestParser_ImplementsInterface(t *testing.T) {
	var _ parser.Parser = (*Parser)(nil)
}

func TestParser_Parse_ComplexXMP(t *testing.T) {
	// Test with more complex XMP structure
	xmp := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description rdf:about=""
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:xmp="http://ns.adobe.com/xap/1.0/"
    xmlns:photoshop="http://ns.adobe.com/photoshop/1.0/"
    dc:format="image/jpeg"
    xmp:Rating="5"
    photoshop:ColorMode="3">
<dc:title>
<rdf:Alt>
<rdf:li xml:lang="x-default">My Photo</rdf:li>
</rdf:Alt>
</dc:title>
<dc:subject>
<rdf:Bag>
<rdf:li>nature</rdf:li>
<rdf:li>landscape</rdf:li>
<rdf:li>sunset</rdf:li>
</rdf:Bag>
</dc:subject>
<dc:creator>
<rdf:Seq>
<rdf:li>John Doe</rdf:li>
</rdf:Seq>
</dc:creator>
</rdf:Description>
</rdf:RDF>
</x:xmpmeta>`

	p := New()
	r := bytes.NewReader([]byte(xmp))
	dirs, parseErr := p.Parse(r)

	if parseErr != nil && parseErr.Error() != "" {
		t.Errorf("Parse() error = %v", parseErr)
	}
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories")
	}
}

func TestParser_Parse_NestedStructs(t *testing.T) {
	xmp := `<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:mwg-rs="http://www.metadataworkinggroup.com/schemas/regions/">
<mwg-rs:Regions rdf:parseType="Resource">
<mwg-rs:AppliedToDimensions rdf:parseType="Resource">
<stDim:w xmlns:stDim="http://ns.adobe.com/xap/1.0/sType/Dimensions#">1920</stDim:w>
<stDim:h xmlns:stDim="http://ns.adobe.com/xap/1.0/sType/Dimensions#">1080</stDim:h>
</mwg-rs:AppliedToDimensions>
</mwg-rs:Regions>
</rdf:Description>
</rdf:RDF>`

	p := New()
	r := bytes.NewReader([]byte(xmp))
	dirs, parseErr := p.Parse(r)

	if parseErr != nil && parseErr.Error() != "" {
		t.Errorf("Parse() error = %v", parseErr)
	}
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories for nested structs")
	}
}

func TestParser_Parse_BoolAndNumericValues(t *testing.T) {
	xmp := `<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:xmp="http://ns.adobe.com/xap/1.0/">
<xmp:Rating>5</xmp:Rating>
<xmp:Flag>true</xmp:Flag>
<xmp:FlagFalse>false</xmp:FlagFalse>
<xmp:Float>3.14</xmp:Float>
</rdf:Description>
</rdf:RDF>`

	p := New()
	r := bytes.NewReader([]byte(xmp))
	dirs, parseErr := p.Parse(r)

	if parseErr != nil && parseErr.Error() != "" {
		t.Errorf("Parse() error = %v", parseErr)
	}
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories")
	}
}

func TestParser_Parse_MultipleDescriptions(t *testing.T) {
	xmp := `<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/" dc:title="Title1"/>
<rdf:Description xmlns:xmp="http://ns.adobe.com/xap/1.0/" xmp:Rating="3"/>
</rdf:RDF>`

	p := New()
	r := bytes.NewReader([]byte(xmp))
	dirs, parseErr := p.Parse(r)

	if parseErr != nil && parseErr.Error() != "" {
		t.Errorf("Parse() error = %v", parseErr)
	}
	if len(dirs) == 0 {
		t.Error("Parse() returned no directories")
	}
}

func TestParser_ConcurrentParse(t *testing.T) {
	// Create minimal valid XMP data
	xmp := `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/" dc:title="Test"/>
</rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`

	p := New()
	r := bytes.NewReader([]byte(xmp))

	const goroutines = 10
	done := make(chan bool, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			p.Parse(r)
			done <- true
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}
