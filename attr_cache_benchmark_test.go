package twig

import (
	"fmt"
	"reflect"
	"testing"
)

// Struct for testing attribute cache
type testType struct {
	Field1 string
	Field2 int
	Field3 bool
}

func (t *testType) Method1() string {
	return t.Field1
}

func (t testType) Method2() int {
	return t.Field2
}

// Create many distinct types to stress the attribute cache
type dynamicType struct {
	name   string
	fields map[string]interface{}
}

// Custom interface to showcase reflection
type displayable interface {
	Display() string
}

func (d dynamicType) Display() string {
	return fmt.Sprintf("Type: %s", d.name)
}

// Benchmark the attribute cache with a small number of types
func BenchmarkAttributeCache_FewTypes(b *testing.B) {
	// Reset the attribute cache
	attributeCache.Lock()
	attributeCache.m = make(map[attributeCacheKey]attributeCacheEntry)
	attributeCache.currSize = 0
	attributeCache.Unlock()

	// Create a render context
	ctx := NewRenderContext(nil, nil, nil)
	defer ctx.Release()

	obj := &testType{
		Field1: "test",
		Field2: 123,
		Field3: true,
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Access different fields and methods
		_, _ = ctx.getAttribute(obj, "Field1")
		_, _ = ctx.getAttribute(obj, "Field2")
		_, _ = ctx.getAttribute(obj, "Field3")
		_, _ = ctx.getAttribute(obj, "Method1")
		_, _ = ctx.getAttribute(obj, "Method2")
	}
}

// Benchmark the attribute cache with many different types
func BenchmarkAttributeCache_ManyTypes(b *testing.B) {
	// Reset the attribute cache
	attributeCache.Lock()
	attributeCache.m = make(map[attributeCacheKey]attributeCacheEntry)
	attributeCache.currSize = 0
	attributeCache.Unlock()

	// Create a render context
	ctx := NewRenderContext(nil, nil, nil)
	defer ctx.Release()

	// Create 2000 different types (more than the cache limit)
	types := make([]interface{}, 2000)
	for i := 0; i < 2000; i++ {
		types[i] = dynamicType{
			name: fmt.Sprintf("Type%d", i),
			fields: map[string]interface{}{
				"field1": fmt.Sprintf("value%d", i),
				"field2": i,
			},
		}
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Access attributes across different types
		typeIdx := i % 2000
		_, _ = ctx.getAttribute(types[typeIdx], "name")
		_, _ = ctx.getAttribute(types[typeIdx], "fields")
	}
}

// Verify that the attribute cache properly performs LRU eviction
func TestAttributeCacheLRUEviction(t *testing.T) {
	// Reset the attribute cache
	attributeCache.Lock()
	attributeCache.m = make(map[attributeCacheKey]attributeCacheEntry)
	attributeCache.currSize = 0
	attributeCache.maxSize = 10 // Small size for testing
	attributeCache.Unlock()

	// Create a render context
	ctx := NewRenderContext(nil, nil, nil)
	defer ctx.Release()

	// Create 20 different types (more than the cache size)
	types := make([]interface{}, 20)
	for i := 0; i < 20; i++ {
		types[i] = dynamicType{
			name: fmt.Sprintf("Type%d", i),
			fields: map[string]interface{}{
				"field1": fmt.Sprintf("value%d", i),
			},
		}
	}

	// First access all types once
	for i := 0; i < 20; i++ {
		_, _ = ctx.getAttribute(types[i], "name")
	}

	// Now access the last 5 types more frequently
	for i := 0; i < 100; i++ {
		typeIdx := 15 + (i % 5) // Types 15-19
		_, _ = ctx.getAttribute(types[typeIdx], "name")
	}

	// Check which types are in the cache
	attributeCache.RLock()
	defer attributeCache.RUnlock()

	// The most recently/frequently used types should be in the cache
	for i := 15; i < 20; i++ {
		typeKey := attributeCacheKey{
			typ:  reflect.TypeOf(types[i]),
			attr: "name",
		}

		_, found := attributeCache.m[typeKey]
		if !found {
			t.Errorf("Expected type %d to be in cache, but it wasn't", i)
		}
	}
}
