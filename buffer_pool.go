package twig

import (
	"fmt"
	"io"
	"strconv"
	"sync"
)

// BufferPool is a specialized pool for string building operations
// Designed for zero allocation rendering of templates
type BufferPool struct {
	pool sync.Pool
}

// Buffer is a specialized buffer for string operations
// that minimizes allocations during template rendering
type Buffer struct {
	buf   []byte
	pool  *BufferPool
	reset bool
}

// Global buffer pool instance
var globalBufferPool = NewBufferPool()

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				// Start with a reasonable capacity
				return &Buffer{
					buf:   make([]byte, 0, 1024),
					reset: true,
				}
			},
		},
	}
}

// Get retrieves a buffer from the pool
func (p *BufferPool) Get() *Buffer {
	buffer := p.pool.Get().(*Buffer)
	if buffer.reset {
		buffer.buf = buffer.buf[:0] // Reset length but keep capacity
	} else {
		buffer.buf = buffer.buf[:0] // Ensure buffer is empty
		buffer.reset = true
	}
	buffer.pool = p
	return buffer
}

// GetBuffer retrieves a buffer from the global pool
func GetBuffer() *Buffer {
	return globalBufferPool.Get()
}

// Release returns the buffer to its pool
func (b *Buffer) Release() {
	if b.pool != nil {
		b.pool.pool.Put(b)
	}
}

// Write implements io.Writer
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteString writes a string to the buffer with zero allocation
func (b *Buffer) WriteString(s string) (n int, err error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// WriteByte writes a single byte to the buffer
func (b *Buffer) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// WriteSpecialized functions for common types to avoid string conversions

// WriteInt writes an integer to the buffer without allocations
func (b *Buffer) WriteInt(i int) (n int, err error) {
	// For small integers, use a table-based approach
	if i >= 0 && i < 10 {
		err = b.WriteByte('0' + byte(i))
		if err == nil {
			n = 1
		}
		return
	} else if i < 0 && i > -10 {
		err = b.WriteByte('-')
		if err != nil {
			return 0, err
		}
		err = b.WriteByte('0' + byte(-i))
		if err == nil {
			n = 2
		}
		return
	}
	
	// Convert to string, this will allocate but is handled later
	s := strconv.Itoa(i)
	return b.WriteString(s)
}

// WriteFloat writes a float to the buffer
func (b *Buffer) WriteFloat(f float64, fmt byte, prec int) (n int, err error) {
	// Use strconv for now - future optimization could implement
	// this without allocation for common cases
	s := strconv.FormatFloat(f, fmt, prec, 64)
	return b.WriteString(s)
}

// WriteBool writes a boolean value to the buffer
func (b *Buffer) WriteBool(v bool) (n int, err error) {
	if v {
		return b.WriteString("true")
	}
	return b.WriteString("false")
}

// Len returns the current length of the buffer
func (b *Buffer) Len() int {
	return len(b.buf)
}

// String returns the contents as a string
func (b *Buffer) String() string {
	return string(b.buf)
}

// Bytes returns the contents as a byte slice
func (b *Buffer) Bytes() []byte {
	return b.buf
}

// Reset empties the buffer
func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
}

// WriteTo writes the buffer to an io.Writer
func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b.buf)
	return int64(n), err
}

// Global-level utility functions for writing values with minimal allocations

// WriteValue writes any value to a writer in the most efficient way possible
func WriteValue(w io.Writer, val interface{}) (n int, err error) {
	// First check if we can use optimized path for known writer types
	if bw, ok := w.(*Buffer); ok {
		return writeValueToBuffer(bw, val)
	}
	
	// If writer is a StringWriter, we can optimize some cases
	if sw, ok := w.(io.StringWriter); ok {
		return writeValueToStringWriter(sw, val)
	}
	
	// Fallback path - use temp buffer for conversion to avoid allocating strings
	buf := GetBuffer()
	defer buf.Release()
	
	_, _ = writeValueToBuffer(buf, val)
	return w.Write(buf.Bytes())
}

// writeValueToBuffer writes a value to a Buffer using type-specific optimizations
func writeValueToBuffer(b *Buffer, val interface{}) (n int, err error) {
	if val == nil {
		return 0, nil
	}
	
	switch v := val.(type) {
	case string:
		return b.WriteString(v)
	case int:
		return b.WriteInt(v)
	case int64:
		return b.WriteString(strconv.FormatInt(v, 10))
	case float64:
		return b.WriteFloat(v, 'f', -1)
	case bool:
		return b.WriteBool(v)
	case []byte:
		return b.Write(v)
	default:
		// Fall back to string conversion
		return b.WriteString(defaultToString(val))
	}
}

// writeValueToStringWriter writes a value to an io.StringWriter
func writeValueToStringWriter(w io.StringWriter, val interface{}) (n int, err error) {
	if val == nil {
		return 0, nil
	}
	
	switch v := val.(type) {
	case string:
		return w.WriteString(v)
	case int:
		if v >= 0 && v < 10 {
			// Single digit optimization
			return w.WriteString(string([]byte{'0' + byte(v)}))
		}
		return w.WriteString(strconv.Itoa(v))
	case int64:
		return w.WriteString(strconv.FormatInt(v, 10))
	case float64:
		return w.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		if v {
			return w.WriteString("true")
		}
		return w.WriteString("false")
	case []byte:
		return w.WriteString(string(v))
	default:
		// Fall back to string conversion
		return w.WriteString(defaultToString(val))
	}
}

// defaultToString converts a value to a string using the default method
func defaultToString(val interface{}) string {
	return stringify(val)
}

// stringify is a helper to convert any value to string
func stringify(val interface{}) string {
	if val == nil {
		return ""
	}
	
	// Use type switch for efficient handling of common types
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	}
	
	// Fall back to fmt.Sprintf for complex types
	return fmt.Sprintf("%v", val)
}