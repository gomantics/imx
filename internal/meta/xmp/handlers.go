package xmp

import (
	"encoding/xml"
)

// StateHandler handles state transitions for a specific context type.
// Each handler is responsible for processing XML elements when the parser
// is in a particular state (e.g., inside an RDF element, property element, etc.).
type StateHandler interface {
	// HandleStart processes a start element and returns a new context.
	// It is called when the parser encounters an opening XML tag.
	//
	// Parameters:
	//   - elem: The XML start element being processed
	//   - parent: The parent context frame
	//   - ns: The current namespace frame
	//   - namespaces: Global namespace mapping (URI -> Prefix)
	//   - nodeMap: Global node map to store properties (needed for Description attrs)
	//
	// Returns:
	//   - New context frame for the element
	//   - Error if the element cannot be processed
	HandleStart(elem xml.StartElement, parent *ContextFrame, ns *NSFrame, namespaces map[string]string, nodeMap NodeMap) (*ContextFrame, error)

	// HandleEnd processes an end element and finalizes the context.
	// It is called when the parser encounters a closing XML tag.
	//
	// Parameters:
	//   - curr: The current context frame being closed
	//   - parent: The parent context frame
	//   - nodeMap: Global node map to store finalized properties
	//
	// Returns:
	//   - Error if the element cannot be finalized
	HandleEnd(curr *ContextFrame, parent *ContextFrame, nodeMap NodeMap) error
}

// HandlerRegistry manages the mapping from ContextType to StateHandler.
type HandlerRegistry struct {
	handlers map[ContextType]StateHandler
}

// NewHandlerRegistry creates and initializes a new handler registry
// with all state handlers registered.
func NewHandlerRegistry() *HandlerRegistry {
	r := &HandlerRegistry{
		handlers: make(map[ContextType]StateHandler),
	}

	// Register all handlers
	r.handlers[CTX_ROOT] = &RootStateHandler{}
	r.handlers[CTX_RDF] = &RDFStateHandler{}
	r.handlers[CTX_DESCRIPTION] = &DescriptionStateHandler{}
	r.handlers[CTX_PROPERTY] = &PropertyStateHandler{}
	r.handlers[CTX_ARRAY] = &ArrayStateHandler{}
	r.handlers[CTX_LI] = &LiStateHandler{}
	r.handlers[CTX_STRUCT_FIELD] = &StructFieldStateHandler{}

	return r
}

// Get returns the handler for the given context type.
// Returns the root handler as fallback if the context type is not registered.
// Panics if the registry is corrupted (should never happen in normal operation).
func (r *HandlerRegistry) Get(ctx ContextType) StateHandler {
	if r == nil || r.handlers == nil {
		panic("handler registry is nil or uninitialized")
	}

	if handler, ok := r.handlers[ctx]; ok && handler != nil {
		return handler
	}

	// Fallback to root handler for unknown contexts
	if root := r.handlers[CTX_ROOT]; root != nil {
		return root
	}

	panic("handler registry corrupted: missing root handler")
}
