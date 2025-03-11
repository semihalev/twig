package twig

import (
	"bytes"
	"io"
	"sync"
)

// byteBufferPool is used to reuse byte buffers during node rendering
var byteBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// GetByteBuffer gets a bytes.Buffer from the pool
func GetByteBuffer() *bytes.Buffer {
	buf := byteBufferPool.Get().(*bytes.Buffer)
	buf.Reset() // Clear any previous content
	return buf
}

// PutByteBuffer returns a bytes.Buffer to the pool
func PutByteBuffer(buf *bytes.Buffer) {
	byteBufferPool.Put(buf)
}

// WriteString optimally writes a string to a writer
// This avoids allocating a new byte slice for each string written
func WriteString(w io.Writer, s string) (int, error) {
	// Fast path for strings.Builder, bytes.Buffer and similar structs that have WriteString
	if sw, ok := w.(io.StringWriter); ok {
		return sw.WriteString(s)
	}
	
	// Fallback path - reuse buffer from pool to avoid allocation
	buf := GetByteBuffer()
	buf.WriteString(s)
	n, err := w.Write(buf.Bytes())
	PutByteBuffer(buf)
	return n, err
}