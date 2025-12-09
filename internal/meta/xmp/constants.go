package xmp

// XML namespace constants used in XMP parsing
const (
	// nsRDF is the RDF syntax namespace
	nsRDF = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"

	// nsXML is the XML namespace
	nsXML = "http://www.w3.org/XML/1998/namespace"
)

// Default values and fallback strings
const (
	// defaultPrefix is the fallback prefix for unknown namespaces
	defaultPrefix = "ns"

	// directoryName is the name used for XMP directories
	directoryName = "XMP"

	// unknownDataType is returned when property kind is not recognized
	unknownDataType = "unknown"
)
