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

// Pre-computed small integer strings to avoid allocations
var smallIntStrings = [...]string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
	"20", "21", "22", "23", "24", "25", "26", "27", "28", "29",
	"30", "31", "32", "33", "34", "35", "36", "37", "38", "39",
	"40", "41", "42", "43", "44", "45", "46", "47", "48", "49",
	"50", "51", "52", "53", "54", "55", "56", "57", "58", "59",
	"60", "61", "62", "63", "64", "65", "66", "67", "68", "69",
	"70", "71", "72", "73", "74", "75", "76", "77", "78", "79",
	"80", "81", "82", "83", "84", "85", "86", "87", "88", "89",
	"90", "91", "92", "93", "94", "95", "96", "97", "98", "99",
}

// Pre-computed small negative integer strings
var smallNegIntStrings = [...]string{
	"0", "-1", "-2", "-3", "-4", "-5", "-6", "-7", "-8", "-9",
	"-10", "-11", "-12", "-13", "-14", "-15", "-16", "-17", "-18", "-19",
	"-20", "-21", "-22", "-23", "-24", "-25", "-26", "-27", "-28", "-29",
	"-30", "-31", "-32", "-33", "-34", "-35", "-36", "-37", "-38", "-39",
	"-40", "-41", "-42", "-43", "-44", "-45", "-46", "-47", "-48", "-49",
	"-50", "-51", "-52", "-53", "-54", "-55", "-56", "-57", "-58", "-59",
	"-60", "-61", "-62", "-63", "-64", "-65", "-66", "-67", "-68", "-69",
	"-70", "-71", "-72", "-73", "-74", "-75", "-76", "-77", "-78", "-79",
	"-80", "-81", "-82", "-83", "-84", "-85", "-86", "-87", "-88", "-89",
	"-90", "-91", "-92", "-93", "-94", "-95", "-96", "-97", "-98", "-99",
}

// WriteInt writes an integer to the buffer with minimal allocations
// Uses a fast path for common integer values
func (b *Buffer) WriteInt(i int) (n int, err error) {
	// Fast path for small integers using pre-computed strings
	if i >= 0 && i < 100 {
		return b.WriteString(smallIntStrings[i])
	} else if i > -100 && i < 0 {
		return b.WriteString(smallNegIntStrings[-i])
	}
	
	// Optimization: manual integer formatting for common sizes
	// Avoid the allocations in strconv.Itoa for numbers we can handle directly
	if i >= -999999 && i <= 999999 {
		return b.formatInt(int64(i))
	}
	
	// For larger integers, fallback to standard formatting
	// This still allocates, but is rare enough to be acceptable
	s := strconv.FormatInt(int64(i), 10)
	return b.WriteString(s)
}

// formatInt does manual string formatting for integers without allocation
// This is a specialized version that handles integers up to 6 digits
func (b *Buffer) formatInt(i int64) (int, error) {
	// Handle negative numbers
	if i < 0 {
		b.WriteByte('-')
		i = -i
	}
	
	// Count digits to determine buffer size
	var digits int
	if i < 10 {
		digits = 1
	} else if i < 100 {
		digits = 2
	} else if i < 1000 {
		digits = 3
	} else if i < 10000 {
		digits = 4
	} else if i < 100000 {
		digits = 5
	} else {
		digits = 6
	}
	
	// Reserve space for the digits
	// Compute in reverse order, then reverse the result
	start := len(b.buf)
	for j := 0; j < digits; j++ {
		digit := byte('0' + i%10)
		b.buf = append(b.buf, digit)
		i /= 10
	}
	
	// Reverse the digits
	end := len(b.buf) - 1
	for j := 0; j < digits/2; j++ {
		b.buf[start+j], b.buf[end-j] = b.buf[end-j], b.buf[start+j]
	}
	
	return digits, nil
}

// WriteFloat writes a float to the buffer with optimizations for common cases
func (b *Buffer) WriteFloat(f float64, fmt byte, prec int) (n int, err error) {
	// Special case for integers or near-integers with default precision
	if prec == -1 && fmt == 'f' {
		// If it's a whole number within integer range, use integer formatting
		if f == float64(int64(f)) && f <= 9007199254740991 && f >= -9007199254740991 {
			// It's a whole number that can be represented exactly as an int64
			return b.formatInt(int64(f))
		}
	}
	
	// Special case for small, common floating-point values with 1-2 decimal places
	if fmt == 'f' && f >= 0 && f < 1000 && (prec == 1 || prec == 2 || prec == -1) {
		// Try to format common floats manually without allocation
		intPart := int64(f)
		
		// Get the fractional part based on precision
		var fracFactor int64
		var fracPrec int
		if prec == -1 {
			// Default precision, up to 6 decimal places
			// Check if we can represent this exactly with fewer digits
			fracPart := f - float64(intPart)
			if fracPart == 0 {
				// It's a whole number
				return b.formatInt(intPart)
			}
			
			// Test if 1-2 decimal places is enough
			if fracPart*100 == float64(int64(fracPart*100)) {
				// Two decimal places is sufficient
				fracFactor = 100
				fracPrec = 2
			} else if fracPart*10 == float64(int64(fracPart*10)) {
				// One decimal place is sufficient
				fracFactor = 10
				fracPrec = 1
			} else {
				// Needs more precision, use strconv
				goto useStrconv
			}
		} else if prec == 1 {
			fracFactor = 10
			fracPrec = 1
		} else {
			fracFactor = 100
			fracPrec = 2
		}
		
		// Format integer part first
		intLen, err := b.formatInt(intPart)
		if err != nil {
			return intLen, err
		}
		
		// Add decimal point
		if err := b.WriteByte('.'); err != nil {
			return intLen, err
		}
		
		// Format fractional part, ensuring proper padding with zeros
		fracPart := int64((f - float64(intPart)) * float64(fracFactor) + 0.5) // Round
		if fracPart >= fracFactor {
			// Rounding caused carry
			fracPart = 0
			// Adjust integer part
			b.Reset()
			intLen, err = b.formatInt(intPart + 1)
			if err != nil {
				return intLen, err
			}
			if err := b.WriteByte('.'); err != nil {
				return intLen, err
			}
		}
		
		// Write fractional part with leading zeros if needed
		if fracPrec == 2 && fracPart < 10 {
			if err := b.WriteByte('0'); err != nil {
				return intLen + 1, err
			}
		}
		
		fracLen, err := b.formatInt(fracPart)
		if err != nil {
			return intLen + 1, err
		}
		
		return intLen + 1 + fracLen, nil
	}
	
useStrconv:
	// Fallback to standard formatting for complex or unusual cases
	s := strconv.FormatFloat(f, fmt, prec, 64)
	return b.WriteString(s)
}

// WriteFormat appends a formatted string to the buffer with minimal allocations
// Similar to fmt.Sprintf but reuses the buffer and avoids allocations
// Only handles a limited set of format specifiers: %s, %d, %v
func (b *Buffer) WriteFormat(format string, args ...interface{}) (n int, err error) {
	// Fast path for simple string with no format specifiers
	if len(args) == 0 {
		return b.WriteString(format)
	}
	
	startIdx := 0
	argIdx := 0
	totalWritten := 0
	
	// Scan the format string for format specifiers
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		
		// Found a potential format specifier
		if i+1 >= len(format) {
			// % at the end of the string is invalid
			break
		}
		
		// Check next character
		next := format[i+1]
		if next == '%' {
			// It's an escaped %
			// Write everything up to and including the first %
			if i > startIdx {
				written, err := b.WriteString(format[startIdx:i+1])
				totalWritten += written
				if err != nil {
					return totalWritten, err
				}
			}
			// Skip the second %
			i++
			startIdx = i+1
			continue
		}
		
		// Write the part before the format specifier
		if i > startIdx {
			written, err := b.WriteString(format[startIdx:i])
			totalWritten += written
			if err != nil {
				return totalWritten, err
			}
		}
		
		// Make sure we have an argument for this specifier
		if argIdx >= len(args) {
			// More specifiers than arguments, skip
			startIdx = i
			continue
		}
		
		arg := args[argIdx]
		argIdx++
		
		// Process the format specifier
		switch next {
		case 's':
			// String format
			if str, ok := arg.(string); ok {
				written, err := b.WriteString(str)
				totalWritten += written
				if err != nil {
					return totalWritten, err
				}
			} else {
				// Convert to string
				written, err := writeValueToBuffer(b, arg)
				totalWritten += written
				if err != nil {
					return totalWritten, err
				}
			}
		case 'd', 'v':
			// Integer or default format
			if i, ok := arg.(int); ok {
				written, err := b.WriteInt(i)
				totalWritten += written
				if err != nil {
					return totalWritten, err
				}
			} else {
				// Use general value formatting
				written, err := writeValueToBuffer(b, arg)
				totalWritten += written
				if err != nil {
					return totalWritten, err
				}
			}
		default:
			// Unsupported format specifier, just output it as-is
			if err := b.WriteByte('%'); err != nil {
				return totalWritten, err
			}
			totalWritten++
			if err := b.WriteByte(next); err != nil {
				return totalWritten, err
			}
			totalWritten++
		}
		
		// Move past the format specifier
		i++
		startIdx = i+1
	}
	
	// Write any remaining part of the format string
	if startIdx < len(format) {
		written, err := b.WriteString(format[startIdx:])
		totalWritten += written
		if err != nil {
			return totalWritten, err
		}
	}
	
	return totalWritten, nil
}

// Grow ensures the buffer has enough capacity for n more bytes
// This helps avoid multiple small allocations during growth
func (b *Buffer) Grow(n int) {
	// Calculate new capacity needed
	needed := len(b.buf) + n
	if cap(b.buf) >= needed {
		return // Already have enough capacity
	}
	
	// Grow capacity with a smart algorithm that avoids frequent resizing
	// Double the capacity until we have enough, but with some optimizations:
	// - For small buffers (<1KB), grow more aggressively (2x)
	// - For medium buffers (1KB-64KB), grow at 1.5x
	// - For large buffers (>64KB), grow at 1.25x to avoid excessive memory usage
	
	newCap := cap(b.buf)
	const (
		smallBuffer  = 1024      // 1KB
		mediumBuffer = 64 * 1024 // 64KB
	)
	
	for newCap < needed {
		if newCap < smallBuffer {
			newCap *= 2 // Double small buffers
		} else if newCap < mediumBuffer {
			newCap = newCap + newCap/2 // Grow medium buffers by 1.5x
		} else {
			newCap = newCap + newCap/4 // Grow large buffers by 1.25x
		}
	}
	
	// Create new buffer with the calculated capacity
	newBuf := make([]byte, len(b.buf), newCap)
	copy(newBuf, b.buf)
	b.buf = newBuf
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
		if v >= 0 && v < 100 {
			// Use cached small integers for fast path
			return w.WriteString(smallIntStrings[v])
		} else if v > -100 && v < 0 {
			return w.WriteString(smallNegIntStrings[-v])
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
		// Use cached small integers for fast path
		if v >= 0 && v < 100 {
			return smallIntStrings[v]
		} else if v > -100 && v < 0 {
			return smallNegIntStrings[-v]
		}
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