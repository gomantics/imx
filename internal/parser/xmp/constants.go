package xmp

const (
	nsRDF = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	nsXML = "http://www.w3.org/XML/1998/namespace"
)

const (
	defaultPrefix   = "ns"
	unknownDataType = "unknown"
)

// Safety limits to prevent maliciously deep or large XMP packets
// are defined in internal/parser/limits.
