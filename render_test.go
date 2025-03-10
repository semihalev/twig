package twig

import (
	"testing"
)

func BenchmarkGetAttribute(b *testing.B) {
	type testStruct struct {
		Name string
		Age  int
	}

	obj := testStruct{
		Name: "John",
		Age:  30,
	}

	ctx := NewRenderContext(nil, nil, nil)
	defer ctx.Release()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.getAttribute(obj, "Name")
		_, _ = ctx.getAttribute(obj, "Age")
	}
}

// Run with multiple iterations to show the caching effect
func BenchmarkGetAttributeRepeated(b *testing.B) {
	type ComplexStruct struct {
		Name   string
		Values []int
		Nested struct {
			Data string
		}
	}

	obj := ComplexStruct{
		Name:   "Complex",
		Values: []int{1, 2, 3},
		Nested: struct {
			Data string
		}{
			Data: "NestedData",
		},
	}

	ctx := NewRenderContext(nil, nil, nil)
	defer ctx.Release()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// First access causes cache miss, subsequent ones use cache
		_, _ = ctx.getAttribute(obj, "Name")
		_, _ = ctx.getAttribute(obj, "Name") // This should be faster - cached
		_, _ = ctx.getAttribute(obj, "Values")
		_, _ = ctx.getAttribute(obj, "Values") // This should be faster - cached
	}
}

func BenchmarkRenderContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx := NewRenderContext(nil, nil, nil)
		ctx.Release()
	}
}

func BenchmarkStringBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := NewStringBuffer()
		buf.Write([]byte("Hello World"))
		_ = buf.String()
		buf.Release()
	}
}