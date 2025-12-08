package xmp

import (
	"testing"
)

func TestHandlerRegistry(t *testing.T) {
	t.Run("NewHandlerRegistry creates registry with all handlers", func(t *testing.T) {
		registry := NewHandlerRegistry()
		if registry == nil {
			t.Fatal("NewHandlerRegistry returned nil")
		}

		// Verify all context types have handlers
		contexts := []ContextType{
			CTX_ROOT,
			CTX_RDF,
			CTX_DESCRIPTION,
			CTX_PROPERTY,
			CTX_ARRAY,
			CTX_LI,
			CTX_STRUCT_FIELD,
		}

		for _, ctx := range contexts {
			handler := registry.Get(ctx)
			if handler == nil {
				t.Errorf("No handler registered for context %s", ctx)
			}
		}
	})

	t.Run("Get returns fallback for unknown context", func(t *testing.T) {
		registry := NewHandlerRegistry()

		// Request handler for invalid context type
		unknownCtx := ContextType(999)
		handler := registry.Get(unknownCtx)

		if handler == nil {
			t.Error("Get should return fallback handler for unknown context, not nil")
		}

		// The fallback should be the ROOT handler
		rootHandler := registry.Get(CTX_ROOT)
		if handler != rootHandler {
			t.Error("Fallback handler should be ROOT handler")
		}
	})
}
