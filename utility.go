package twig

import (
	"io"
	"strconv"
)

// countNewlines counts newlines in a string without allocations.
// This is a zero-allocation replacement for strings.Count(s, "\n")
func countNewlines(s string) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			count++
		}
	}
	return count
}

// WriteString optimally writes a string to a writer
// This avoids allocating a new byte slice for each string written
// Uses our optimized Buffer pool for better performance
func WriteString(w io.Writer, s string) (int, error) {
	// Fast path for strings.Builder, bytes.Buffer and similar structs that have WriteString
	if sw, ok := w.(io.StringWriter); ok {
		return sw.WriteString(s)
	}

	// Fast path for our own Buffer type
	if buf, ok := w.(*Buffer); ok {
		return buf.WriteString(s)
	}

	// Fallback path - reuse buffer from pool to avoid allocation
	buf := GetBuffer()
	buf.WriteString(s)
	n, err := w.Write(buf.Bytes())
	buf.Release()
	return n, err
}

// WriteFormat writes a formatted string to a writer with minimal allocations
// Similar to fmt.Fprintf but uses our optimized Buffer for better performance
func WriteFormat(w io.Writer, format string, args ...interface{}) (int, error) {
	// Fast path for our own Buffer type
	if buf, ok := w.(*Buffer); ok {
		return buf.WriteFormat(format, args...)
	}

	// Use a pooled buffer for other writer types
	buf := GetBuffer()
	defer buf.Release()

	// Write the formatted string to the buffer
	buf.WriteFormat(format, args...)

	// Write the buffer to the writer
	return w.Write(buf.Bytes())
}

// FormatInt formats an integer without allocations
// Returns a string representation using cached small integers
func FormatInt(i int) string {
	// Use pre-computed strings for small integers
	if i >= 0 && i < 100 {
		return smallIntStrings[i]
	} else if i > -100 && i < 0 {
		return smallNegIntStrings[-i]
	}

	// Fall back to standard formatting
	return strconv.Itoa(i)
}
