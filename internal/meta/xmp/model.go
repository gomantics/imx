package xmp

import (
	"strings"
)

// ContextType represents the current state in the XMP state machine.

type ContextType int

const (
	CTX_ROOT ContextType = iota
	CTX_RDF
	CTX_DESCRIPTION
	CTX_PROPERTY
	CTX_ARRAY
	CTX_LI
	CTX_STRUCT_FIELD // For when a property is treated as a struct field (nested)
)

// NSFrame tracks namespaces at a specific depth
type NSFrame struct {
	prefixToURI map[string]string
	uriToPrefix map[string]string
}

// ContextFrame tracks parsing state
type ContextFrame struct {
	Type ContextType

	// For CTX_PROPERTY / CTX_STRUCT_FIELD
	propURI    string
	propLocal  string
	propPrefix string
	propKind   PropKind // inferred kind

	// Buffers
	text   strings.Builder // for Simple values
	items  []PropertyValue // for Array items
	fields []StructField   // for Struct fields
}

type PropKind int

const (
	KindUnknown PropKind = iota
	KindSimple
	KindArray
	KindStruct
)

type PropertyValue struct {
	Kind   PropKind
	Scalar string          // for Simple
	Items  []PropertyValue // for Array
	Fields []StructField   // for Struct
}

type StructField struct {
	Prefix string
	URI    string
	Name   string
	Value  PropertyValue
}

type PropertyKey struct {
	URI   string
	Local string
}

type NodeMap map[PropertyKey][]PropertyValue
