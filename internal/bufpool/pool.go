// Package bufpool provides a buffer pool for reducing GC pressure from repeated small buffer allocations.
//
// Usage:
//
//	buf := bufpool.Get(4)
//	defer bufpool.Put(buf)
//	n, err := r.ReadAt(buf, offset)
//	// use buf...
//
// The pool manages buffers of standard sizes (2, 4, 8, 16, 256, 4096 bytes).
// Requesting a non-standard size returns the next larger pool size.
// Buffers larger than 4096 bytes are not pooled and are allocated directly.
package bufpool

import "sync"

// Standard buffer sizes managed by the pool
const (
	Size2    = 2
	Size4    = 4
	Size8    = 8
	Size16   = 16
	Size256  = 256
	Size4096 = 4096
)

var (
	pool2    = &sync.Pool{New: func() interface{} { return make([]byte, Size2) }}
	pool4    = &sync.Pool{New: func() interface{} { return make([]byte, Size4) }}
	pool8    = &sync.Pool{New: func() interface{} { return make([]byte, Size8) }}
	pool16   = &sync.Pool{New: func() interface{} { return make([]byte, Size16) }}
	pool256  = &sync.Pool{New: func() interface{} { return make([]byte, Size256) }}
	pool4096 = &sync.Pool{New: func() interface{} { return make([]byte, Size4096) }}
)

// Get returns a buffer of at least the requested size from the pool.
func Get(size int) []byte {
	switch {
	case size <= Size2:
		return pool2.Get().([]byte)
	case size <= Size4:
		return pool4.Get().([]byte)
	case size <= Size8:
		return pool8.Get().([]byte)
	case size <= Size16:
		return pool16.Get().([]byte)
	case size <= Size256:
		return pool256.Get().([]byte)
	case size <= Size4096:
		return pool4096.Get().([]byte)
	default:
		// Don't pool very large buffers
		return make([]byte, size)
	}
}

// Put returns a buffer to the pool for reuse.
func Put(buf []byte) {
	if buf == nil {
		return
	}

	// Only return standard-sized buffers to the pool
	switch cap(buf) {
	case Size2:
		pool2.Put(buf[:Size2])
	case Size4:
		pool4.Put(buf[:Size4])
	case Size8:
		pool8.Put(buf[:Size8])
	case Size16:
		pool16.Put(buf[:Size16])
	case Size256:
		pool256.Put(buf[:Size256])
	case Size4096:
		pool4096.Put(buf[:Size4096])
	}
}
