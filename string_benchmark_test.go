package twig

import (
	"io/ioutil"
	"testing"
)

func BenchmarkWriteStringDirect(b *testing.B) {
	buf := NewStringBuffer()
	defer buf.Release()
	longStr := "This is a test string for benchmarking the write performance of direct byte slice conversion"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.buf.Reset()
		buf.buf.Write([]byte(longStr))
	}
}

func BenchmarkWriteStringOptimized(b *testing.B) {
	buf := NewStringBuffer()
	defer buf.Release()
	longStr := "This is a test string for benchmarking the write performance of optimized string writing"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.buf.Reset()
		WriteString(&buf.buf, longStr)
	}
}

func BenchmarkWriteStringDirect_Discard(b *testing.B) {
	longStr := "This is a test string for benchmarking the write performance of direct byte slice conversion"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ioutil.Discard.Write([]byte(longStr))
	}
}

func BenchmarkWriteStringOptimized_Discard(b *testing.B) {
	longStr := "This is a test string for benchmarking the write performance of optimized string writing"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WriteString(ioutil.Discard, longStr)
	}
}
