package imx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/gomantics/imx/internal/parser"
)

// Re-export parser types as the public API types
type (
	TagID     = parser.TagID
	Tag       = parser.Tag
	Directory = parser.Directory
)

// Metadata is the top-level container for all parsed metadata.
// Fields are unexported to prevent external mutation; use accessor methods instead.
type Metadata struct {
	directories []Directory // All parsed directories
	errors      []error     // All errors encountered during parsing

	index map[TagID]*Tag // Lazy-built index for O(1) tag lookup
	mu    sync.RWMutex   // Protects index during lazy initialization
}

// Directories returns a slice of all parsed metadata directories.
// The returned slice is a copy to prevent external modification.
func (m *Metadata) Directories() []Directory {
	if m == nil {
		return nil
	}
	dirs := make([]Directory, len(m.directories))
	copy(dirs, m.directories)
	return dirs
}

// Errors returns a slice of all errors encountered during parsing.
// The returned slice is a copy to prevent external modification.
func (m *Metadata) Errors() []error {
	if m == nil {
		return nil
	}
	errs := make([]error, len(m.errors))
	copy(errs, m.errors)
	return errs
}

// Directory returns the directory with the given name
func (m *Metadata) Directory(name string) (Directory, bool) {
	for _, dir := range m.directories {
		if dir.Name == name {
			return dir, true
		}
	}
	return Directory{}, false
}

// Tag returns the tag with the given ID using an efficient index.
// The index is built lazily on first call and cached for subsequent calls.
func (m *Metadata) Tag(id TagID) (Tag, bool) {
	// Fast path: check if index exists (read lock)
	m.mu.RLock()
	if m.index != nil {
		tag, ok := m.index[id]
		m.mu.RUnlock()
		if ok {
			return *tag, true
		}
		return Tag{}, false
	}
	m.mu.RUnlock()

	// Slow path: build index (write lock)
	m.mu.Lock()
	// Double-check in case another goroutine built it
	if m.index == nil {
		m.buildIndex()
	}
	tag, ok := m.index[id]
	m.mu.Unlock()

	if ok {
		return *tag, true
	}
	return Tag{}, false
}

// buildIndex builds the internal index for O(1) tag lookup.
// Caller must hold m.mu.
func (m *Metadata) buildIndex() {
	m.index = make(map[TagID]*Tag)
	for i := range m.directories {
		dir := &m.directories[i]
		for j := range dir.Tags {
			tag := &dir.Tags[j]
			m.index[tag.ID] = tag
		}
	}
}

// GetAll returns a map of values for the given tag IDs
func (m *Metadata) GetAll(ids ...TagID) map[TagID]any {
	result := make(map[TagID]any, len(ids))
	for _, id := range ids {
		if tag, ok := m.Tag(id); ok {
			result[id] = tag.Value
		}
	}
	return result
}

// Each iterates over all tags, calling fn for each tag.
// If fn returns false, iteration stops.
func (m *Metadata) Each(fn func(Directory, Tag) bool) {
	for _, dir := range m.directories {
		for _, tag := range dir.Tags {
			if !fn(dir, tag) {
				return
			}
		}
	}
}

// EachTag iterates over all tags across all directories.
// If fn returns false, iteration stops.
func (m *Metadata) EachTag(fn func(Tag) bool) {
	for _, dir := range m.directories {
		for _, tag := range dir.Tags {
			if !fn(tag) {
				return
			}
		}
	}
}

// EachInDirectory iterates over tags in the given directory.
// If fn returns false, iteration stops.
func (m *Metadata) EachInDirectory(name string, fn func(Tag) bool) {
	for _, dir := range m.directories {
		if dir.Name == name {
			for _, tag := range dir.Tags {
				if !fn(tag) {
					return
				}
			}
			return
		}
	}
}

// AllTags returns a flat slice of all tags across all directories.
// The order matches the iteration order (directory order, then tag order within each directory).
func (m *Metadata) AllTags() []Tag {
	var tags []Tag
	for _, dir := range m.directories {
		tags = append(tags, dir.Tags...)
	}
	return tags
}

// DirectoryNames returns a list of all directory names present in the metadata.
func (m *Metadata) DirectoryNames() []string {
	names := make([]string, 0, len(m.directories))
	for _, dir := range m.directories {
		names = append(names, dir.Name)
	}
	return names
}

// TagCount returns the total number of tags across all directories.
func (m *Metadata) TagCount() int {
	count := 0
	for _, dir := range m.directories {
		count += len(dir.Tags)
	}
	return count
}

// GetString returns the tag value as a string.
//
// Conversion rules:
//   - string: returned as-is
//   - []byte: converted to string
//   - fmt.Stringer: calls String() method
//   - all other types: converted using fmt.Sprintf("%v", value)
//
// The fallback conversion allows numeric types (int, float, etc.) commonly found
// in metadata to be displayed as strings. For type-safe numeric conversions,
// use GetInt or GetFloat instead.
//
// Returns an error only if the tag doesn't exist.
func (m *Metadata) GetString(id TagID) (string, error) {
	tag, ok := m.Tag(id)
	if !ok {
		return "", fmt.Errorf("tag %q not found", id)
	}

	switch v := tag.Value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		// Fallback for numeric and other types commonly found in metadata
		return fmt.Sprintf("%v", v), nil
	}
}

// GetInt returns the tag value as an int64.
// Returns an error if the tag doesn't exist or cannot be converted to int64.
func (m *Metadata) GetInt(id TagID) (int64, error) {
	tag, ok := m.Tag(id)
	if !ok {
		return 0, fmt.Errorf("tag %q not found", id)
	}

	switch v := tag.Value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if v > 1<<63-1 {
			return 0, fmt.Errorf("value %d overflows int64", v)
		}
		return int64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// GetFloat returns the tag value as a float64.
// Returns an error if the tag doesn't exist or cannot be converted to float64.
func (m *Metadata) GetFloat(id TagID) (float64, error) {
	tag, ok := m.Tag(id)
	if !ok {
		return 0, fmt.Errorf("tag %q not found", id)
	}

	switch v := tag.Value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// GetBytes returns the tag value as a byte slice.
// Returns an error if the tag doesn't exist or is not a byte slice or string.
func (m *Metadata) GetBytes(id TagID) ([]byte, error) {
	tag, ok := m.Tag(id)
	if !ok {
		return nil, fmt.Errorf("tag %q not found", id)
	}

	switch v := tag.Value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []byte", v)
	}
}

// MarshalJSON implements json.Marshaler for Metadata.
// The JSON structure is:
//
//	{
//	  "directories": [...],
//	  "errors": [...]
//	}
func (m *Metadata) MarshalJSON() ([]byte, error) {
	type Alias Metadata

	// Convert errors to strings for JSON serialization
	var errorStrings []string
	if len(m.errors) > 0 {
		errorStrings = make([]string, len(m.errors))
		for i, err := range m.errors {
			errorStrings[i] = err.Error()
		}
	}

	return json.Marshal(&struct {
		Directories []Directory `json:"directories"`
		Errors      []string    `json:"errors,omitempty"`
	}{
		Directories: m.directories,
		Errors:      errorStrings,
	})
}

// readerAdapter implements io.ReaderAt by buffering data from an io.Reader.
//
// This adapter enables parsers that require random access (io.ReaderAt) to work
// with streaming sources (io.Reader) like HTTP responses or pipes. It achieves
// this by buffering data on-demand as the parser requests it.
//
// Buffering strategy:
//   - Data is read from the underlying io.Reader only when needed
//   - All read data is cached in an internal buffer
//   - Subsequent reads from already-buffered regions are served from cache
//   - Memory usage grows only as needed by the parser
//
// Performance characteristics:
//   - First read at offset N: O(N) - must buffer all data up to N
//   - Subsequent reads: O(1) - served directly from buffer
//   - Memory: O(max offset accessed)
//   - Best for: Sequential or forward-seeking access patterns
//   - Worst for: Random backward seeks (entire stream must be buffered)
//
// This design is optimized for image metadata parsers, which typically:
//   - Read headers sequentially from the beginning
//   - Occasionally seek to known offsets for specific data blocks
//   - Rarely seek backward to earlier positions
type readerAdapter struct {
	r       io.Reader     // Underlying streaming source
	buffer  *bytes.Buffer // Accumulated data buffer
	eof     bool          // Whether we've reached EOF on the source
	limit   int64         // Maximum bytes to buffer (0 = unlimited)
	bufSize int           // Read chunk size
	lastErr error         // Sticky error (e.g., max bytes exceeded)
}

// boundedReaderAt wraps an io.ReaderAt and enforces a byte limit.
type boundedReaderAt struct {
	r       io.ReaderAt
	limit   int64 // 0 = unlimited
	lastErr error
}

// newReaderAdapter creates a new adapter that wraps an io.Reader.
// The adapter starts with an empty buffer and reads data on-demand.
func newReaderAdapter(r io.Reader, maxBytes int64, bufferSize int) *readerAdapter {
	if bufferSize <= 0 {
		bufferSize = 64 << 10 // default 64KB
	}
	return &readerAdapter{
		r:       r,
		buffer:  &bytes.Buffer{},
		eof:     false,
		limit:   maxBytes,
		bufSize: bufferSize,
	}
}

// ReadAt enforces the configured byte limit before delegating.
func (b *boundedReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if b.limit > 0 && off+int64(len(p)) > b.limit {
		b.lastErr = ErrMaxBytesExceeded
		return 0, ErrMaxBytesExceeded
	}
	return b.r.ReadAt(p, off)
}

// LastError returns the most recent error encountered by the bounded reader.
func (b *boundedReaderAt) LastError() error {
	return b.lastErr
}

// ReadAt reads len(p) bytes into p starting at offset off.
// It implements the io.ReaderAt interface by buffering data from the underlying reader.
// Returns io.ErrUnexpectedEOF if we hit EOF before reading all requested bytes.
func (ra *readerAdapter) ReadAt(p []byte, off int64) (n int, err error) {
	// Enforce max bytes limit
	if ra.limit > 0 && off+int64(len(p)) > ra.limit {
		ra.lastErr = ErrMaxBytesExceeded
		return 0, ErrMaxBytesExceeded
	}

	// Ensure we have enough data buffered
	currentSize := int64(ra.buffer.Len())
	needed := off + int64(len(p))

	if needed > currentSize && !ra.eof {
		// Need to read more data from the source
		toRead := needed - currentSize
		chunkSize := int64(ra.bufSize)
		if chunkSize <= 0 {
			chunkSize = toRead
		}

		for toRead > 0 {
			readLen := chunkSize
			if toRead < readLen {
				readLen = toRead
			}
			chunk := make([]byte, readLen)
			nr, readErr := io.ReadFull(ra.r, chunk)
			if nr > 0 {
				ra.buffer.Write(chunk[:nr])
				toRead -= int64(nr)
			}

			if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
				ra.eof = true
				break
			} else if readErr != nil && readErr != io.EOF {
				ra.lastErr = readErr
				return 0, readErr
			}

			// Stop early if we've met the required buffer size
			if toRead <= 0 {
				break
			}

			// Respect limit
			if ra.limit > 0 && int64(ra.buffer.Len()) >= ra.limit {
				ra.lastErr = ErrMaxBytesExceeded
				return 0, ErrMaxBytesExceeded
			}
		}
	}

	// Read from buffer
	bufData := ra.buffer.Bytes()
	if off >= int64(len(bufData)) {
		return 0, io.EOF
	}

	n = copy(p, bufData[off:])
	if n < len(p) {
		// Couldn't read all requested bytes - return UnexpectedEOF
		return n, io.ErrUnexpectedEOF
	}

	return n, nil
}

// LastError returns the last sticky error encountered by the adapter.
func (ra *readerAdapter) LastError() error {
	return ra.lastErr
}
