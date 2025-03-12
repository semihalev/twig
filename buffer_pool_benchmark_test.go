package twig

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"testing"
)

func BenchmarkBufferWrite(b *testing.B) {
	buf := GetBuffer()
	defer buf.Release()
	longStr := "This is a test string for benchmarking the write performance of the new buffer pool"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.WriteString(longStr)
	}
}

func BenchmarkStandardBufferWrite(b *testing.B) {
	buf := &bytes.Buffer{}
	longStr := "This is a test string for benchmarking the write performance of standard byte buffer"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.WriteString(longStr)
	}
}

func BenchmarkBufferIntegerFormatting(b *testing.B) {
	buf := GetBuffer()
	defer buf.Release()

	vals := []int{0, 5, -5, 123, -123, 9999, -9999, 123456789, -123456789}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		for _, v := range vals {
			buf.WriteInt(v)
		}
	}
}

func BenchmarkStandardIntegerFormatting(b *testing.B) {
	buf := &bytes.Buffer{}

	vals := []int{0, 5, -5, 123, -123, 9999, -9999, 123456789, -123456789}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		for _, v := range vals {
			buf.WriteString(strconv.Itoa(v))
		}
	}
}

func BenchmarkSmallIntegerFormatting(b *testing.B) {
	// Test specifically for small integers which should use the
	// pre-computed string table
	buf := GetBuffer()
	defer buf.Release()

	b.Run("Optimized_Small_Ints", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			for j := 0; j < 100; j++ {
				buf.WriteInt(j)
			}
		}
	})

	b.Run("Standard_Small_Ints", func(b *testing.B) {
		sbuf := &bytes.Buffer{}
		for i := 0; i < b.N; i++ {
			sbuf.Reset()
			for j := 0; j < 100; j++ {
				sbuf.WriteString(strconv.Itoa(j))
			}
		}
	})
}

func BenchmarkFloatFormatting(b *testing.B) {
	buf := GetBuffer()
	defer buf.Release()

	vals := []float64{
		0.0, 1.0, -1.0, // Whole numbers
		3.14, -2.718, // Common constants
		123.456, -789.012, // Medium floats
		0.123, 0.001, 9.999, // Small decimals
		1234567.89, -9876543.21, // Large numbers
	}

	b.Run("OptimizedFloat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			for _, v := range vals {
				buf.WriteFloat(v, 'f', -1)
			}
		}
	})

	b.Run("StandardFloat", func(b *testing.B) {
		sbuf := &bytes.Buffer{}
		for i := 0; i < b.N; i++ {
			sbuf.Reset()
			for _, v := range vals {
				sbuf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
			}
		}
	})
}

func BenchmarkFormatString(b *testing.B) {
	buf := GetBuffer()
	defer buf.Release()

	format := "Hello, %s! Count: %d, Value: %v"
	name := "World"
	count := 42
	value := true

	b.Run("BufferFormat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			buf.WriteFormat(format, name, count, value)
		}
	})

	b.Run("FmtSprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Each fmt.Sprintf creates a new string
			_ = fmt.Sprintf(format, name, count, value)
		}
	})
}

func BenchmarkFormatInt(b *testing.B) {
	b.Run("SmallInt_Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FormatInt(42)
		}
	})

	b.Run("SmallInt_Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = strconv.Itoa(42)
		}
	})

	b.Run("LargeInt_Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FormatInt(12345678)
		}
	})

	b.Run("LargeInt_Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = strconv.Itoa(12345678)
		}
	})
}

func BenchmarkWriteValue(b *testing.B) {
	buf := GetBuffer()
	defer buf.Release()

	values := []interface{}{
		"string value",
		123,
		-456,
		3.14159,
		true,
		[]byte("byte slice"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		for _, v := range values {
			writeValueToBuffer(buf, v)
		}
	}
}

func BenchmarkStringifyValues(b *testing.B) {
	values := []interface{}{
		"string value",
		123,
		-456,
		3.14159,
		true,
		[]byte("byte slice"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range values {
			_ = stringify(v)
		}
	}
}

func BenchmarkBufferGrowth(b *testing.B) {
	// Test how the buffer handles growing for larger strings
	b.Run("Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := GetBuffer()
			buf.WriteString("small")
			buf.Release()
		}
	})

	b.Run("Medium", func(b *testing.B) {
		mediumStr := "medium string that is longer than the small one but still reasonable"
		for i := 0; i < b.N; i++ {
			buf := GetBuffer()
			buf.WriteString(mediumStr)
			buf.Release()
		}
	})

	b.Run("Large", func(b *testing.B) {
		largeStr := string(make([]byte, 2048)) // 2KB string
		for i := 0; i < b.N; i++ {
			buf := GetBuffer()
			buf.WriteString(largeStr)
			buf.Release()
		}
	})
}

func BenchmarkBufferToWriter(b *testing.B) {
	buf := GetBuffer()
	defer buf.Release()

	str := "This is a test string that will be written to a discard writer"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.WriteString(str)
		buf.WriteTo(io.Discard)
	}
}
