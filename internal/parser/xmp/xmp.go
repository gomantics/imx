package xmp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/gomantics/imx/internal/parser"
	"github.com/gomantics/imx/internal/parser/limits"
)

type Parser struct {
	handlers map[ContextType]StateHandler
}

func New() *Parser {
	handlers := make(map[ContextType]StateHandler, 7)
	handlers[CTX_ROOT] = &RootStateHandler{}
	handlers[CTX_RDF] = &RDFStateHandler{}
	handlers[CTX_DESCRIPTION] = &DescriptionStateHandler{}
	handlers[CTX_PROPERTY] = &PropertyStateHandler{}
	handlers[CTX_ARRAY] = &ArrayStateHandler{}
	handlers[CTX_LI] = &LiStateHandler{}
	handlers[CTX_STRUCT_FIELD] = &StructFieldStateHandler{}

	return &Parser{
		handlers: handlers,
	}
}

func (p *Parser) Name() string {
	return "XMP"
}

func (p *Parser) Detect(r io.ReaderAt) bool {
	buf := make([]byte, 100)
	_, err := r.ReadAt(buf, 0)
	if err != nil {
		return false
	}
	return bytes.Contains(buf, []byte("<?xpacket")) || bytes.Contains(buf, []byte("<x:xmpmeta"))
}

func (p *Parser) Parse(r io.ReaderAt) ([]parser.Directory, *parser.ParseError) {
	parseErr := parser.NewParseError()

	nodeMap := make(NodeMap)
	namespaces := make(map[string]string)

	reader := &readerAtWrapper{r: r, offset: 0}

	if err := p.parsePacket(reader, nodeMap, namespaces); err != nil {
		parseErr.Add(fmt.Errorf("parse XMP packet: %w", err))
		return nil, parseErr
	}

	if len(nodeMap) == 0 {
		return nil, nil
	}

	dirs := flattenNodeMap(nodeMap, namespaces)
	return dirs, parseErr.OrNil()
}

type readerAtWrapper struct {
	r      io.ReaderAt
	offset int64
}

func (w *readerAtWrapper) Read(p []byte) (int, error) {
	n, err := w.r.ReadAt(p, w.offset)
	w.offset += int64(n)
	return n, err
}

func (p *Parser) parsePacket(r io.Reader, nodeMap NodeMap, namespaces map[string]string) error {
	if nodeMap == nil {
		return fmt.Errorf("nodeMap cannot be nil")
	}
	if namespaces == nil {
		return fmt.Errorf("namespaces map cannot be nil")
	}

	decoder := xml.NewDecoder(r)

	nsStack := []*NSFrame{replaceNSFrame(nil, nil)}
	ctxStack := []*ContextFrame{{Type: CTX_ROOT}}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decode XML token: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if len(ctxStack) >= limits.MaxXMPDepth {
				return fmt.Errorf("XML depth exceeds limit (%d)", limits.MaxXMPDepth)
			}
			parentNS := nsStack[len(nsStack)-1]
			currNS := replaceNSFrame(parentNS, t.Attr)
			nsStack = append(nsStack, currNS)

			parent := ctxStack[len(ctxStack)-1]
			handler := p.handlers[parent.Type]
			newCtx := handler.HandleStart(t, parent, currNS, namespaces, nodeMap)
			ctxStack = append(ctxStack, newCtx)

		case xml.EndElement:
			curr := ctxStack[len(ctxStack)-1]
			parent := ctxStack[len(ctxStack)-2]
			handler := p.handlers[curr.Type]
			handler.HandleEnd(curr, parent, nodeMap)

			ctxStack = ctxStack[:len(ctxStack)-1]
			nsStack = nsStack[:len(nsStack)-1]

		case xml.CharData:
			top := ctxStack[len(ctxStack)-1]
			if top.text.Len()+len(t) > limits.MaxXMPTextBytes {
				return fmt.Errorf("XMP text exceeds limit (%d bytes)", limits.MaxXMPTextBytes)
			}
			top.text.Write(t)
		}
	}

	return nil
}
