package twig

import (
	"bytes"
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

func BenchmarkByteBufferWrite(b *testing.B) {
	buf := GetByteBuffer()
	defer PutByteBuffer(buf)
	longStr := "This is a test string for benchmarking the write performance of the byte buffer pool"

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