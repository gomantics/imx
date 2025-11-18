package formats

// Result captures format-specific metadata returned by parsers.
type Result struct {
	Width         int
	Height        int
	ColorDepth    int
	ColorSpace    string
	HasICCProfile bool
	EXIF          map[string]interface{}
	Additional    map[string]interface{}
}

// newResult allocates a result with initialized maps.
func newResult() *Result {
	return &Result{
		EXIF:       make(map[string]interface{}),
		Additional: make(map[string]interface{}),
	}
}
