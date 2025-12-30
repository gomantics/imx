package testing

// DirectoryExpectation defines expected directory with all its tags.
type DirectoryExpectation struct {
	Name          string           // Directory name (e.g., "IFD0", "ExifIFD")
	ExactTagCount int              // MUST have exactly this many tags (catches missing/extra tags)
	Tags          []TagExpectation // ALL tags in this directory
}

// TagExpectation defines tag validation requirements.
// Use ONE of: Value (exact match), Type (type check), Pattern (regex), or just name (presence).
type TagExpectation struct {
	Name string // Tag name (required)

	// Validation level (pick one):
	Value   interface{} // Exact value match (for critical tags like "Make", "Model")
	Type    string      // Just check type: "string", "uint16", "uint32", "[]uint16", etc.
	Pattern string      // Regex pattern for string values (e.g., for dates, versions)
}
