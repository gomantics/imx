package testing

// ValuesEqual compares two values for equality, handling type conversions.
// This is needed because Go is strict about types, but we want to allow
// comparing uint16(8688) == uint32(8688) conceptually.
func ValuesEqual(got, want interface{}) bool {
	// String comparison
	gotStr, gotIsStr := got.(string)
	wantStr, wantIsStr := want.(string)
	if gotIsStr && wantIsStr {
		return gotStr == wantStr
	}

	// Slice comparisons
	switch w := want.(type) {
	case []uint16:
		if g, ok := got.([]uint16); ok {
			if len(g) != len(w) {
				return false
			}
			for i := range g {
				if g[i] != w[i] {
					return false
				}
			}
			return true
		}
	case []byte: // []byte is the same as []uint8 in Go
		if g, ok := got.([]byte); ok {
			if len(g) != len(w) {
				return false
			}
			for i := range g {
				if g[i] != w[i] {
					return false
				}
			}
			return true
		}
	case []string:
		if g, ok := got.([]string); ok {
			if len(g) != len(w) {
				return false
			}
			for i := range g {
				if g[i] != w[i] {
					return false
				}
			}
			return true
		}
	}

	// Boolean comparison
	gotBool, gotIsBool := got.(bool)
	wantBool, wantIsBool := want.(bool)
	if gotIsBool && wantIsBool {
		return gotBool == wantBool
	}

	// Numeric comparisons with type conversions
	switch w := want.(type) {
	case uint8:
		if g, ok := got.(uint8); ok {
			return g == w
		}
	case uint16:
		if g, ok := got.(uint16); ok {
			return g == w
		}
	case uint32:
		if g, ok := got.(uint32); ok {
			return g == w
		}
	case uint64:
		if g, ok := got.(uint64); ok {
			return g == w
		}
	case int:
		if g, ok := got.(int); ok {
			return g == w
		}
		// Handle int stored as uint32
		if g, ok := got.(uint32); ok {
			return int64(g) == int64(w)
		}
	case int64:
		if g, ok := got.(int64); ok {
			return g == w
		}
	case float64:
		if g, ok := got.(float64); ok {
			return g == w
		}
	}

	return false
}

// TypeMatches checks if value matches expected type string.
// Supports common types used in metadata parsing.
func TypeMatches(value interface{}, typeName string) bool {
	switch typeName {
	case "string":
		_, ok := value.(string)
		return ok
	case "uint8":
		_, ok := value.(uint8)
		return ok
	case "uint16":
		_, ok := value.(uint16)
		return ok
	case "uint32":
		_, ok := value.(uint32)
		return ok
	case "uint64":
		_, ok := value.(uint64)
		return ok
	case "int":
		_, ok := value.(int)
		return ok
	case "int64":
		_, ok := value.(int64)
		return ok
	case "float64":
		_, ok := value.(float64)
		return ok
	case "[]uint16":
		_, ok := value.([]uint16)
		return ok
	case "[]uint8":
		_, ok := value.([]uint8)
		return ok
	case "[]byte":
		_, ok := value.([]byte)
		return ok
	case "[]string":
		_, ok := value.([]string)
		return ok
	case "bool":
		_, ok := value.(bool)
		return ok
	default:
		// Unknown type - fail the check
		return false
	}
}
