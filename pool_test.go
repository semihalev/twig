package twig

import (
	"sync"
	"testing"
)

func TestRenderContextPool(t *testing.T) {
	// Create a new context from the pool
	ctx := NewRenderContext(nil, nil, nil)

	// Test that it's initialized properly
	if ctx.context == nil {
		t.Error("Context map should be initialized")
	}
	if ctx.blocks == nil {
		t.Error("Blocks map should be initialized")
	}
	if ctx.macros == nil {
		t.Error("Macros map should be initialized")
	}

	// Add some data to the context
	ctx.context["test"] = "value"
	ctx.blocks["block"] = []Node{&TextNode{content: "test", line: 1}}
	ctx.macros["macro"] = &MacroNode{name: "test", line: 1}

	// Release it back to the pool
	ctx.Release()

	// Get another context from the pool (should be the same one)
	ctx2 := NewRenderContext(nil, nil, nil)

	// Data should be cleared
	if len(ctx2.context) != 0 {
		t.Errorf("Context should be empty, but has %d items", len(ctx2.context))
	}
	if len(ctx2.blocks) != 0 {
		t.Errorf("Blocks should be empty, but has %d items", len(ctx2.blocks))
	}
	if len(ctx2.macros) != 0 {
		t.Errorf("Macros should be empty, but has %d items", len(ctx2.macros))
	}

	// Clean up
	ctx2.Release()
}

func TestStringBufferPool(t *testing.T) {
	// Get a buffer from the pool
	buf := NewStringBuffer()

	// Write to it
	testData := "Hello, World!"
	buf.Write([]byte(testData))

	// Check the content
	if buf.String() != testData {
		t.Errorf("Expected '%s', got '%s'", testData, buf.String())
	}

	// Release it back to the pool
	buf.Release()

	// Get another buffer (should be the same one)
	buf2 := NewStringBuffer()

	// Should be empty
	if buf2.String() != "" {
		t.Errorf("Buffer should be empty, got '%s'", buf2.String())
	}

	// Clean up
	buf2.Release()
}

func TestRenderContextPoolConcurrency(t *testing.T) {
	// Test concurrent access to the render context pool
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Create channels to coordinate the test
	startCh := make(chan struct{})

	// Launch goroutines that will get and release contexts
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Wait for the start signal
			<-startCh

			// Get a context from the pool
			ctx := NewRenderContext(nil, map[string]interface{}{"id": id}, nil)

			// Verify it's properly initialized
			if val, exists := ctx.context["id"]; !exists || val != id {
				t.Errorf("Context data not properly initialized: expected id=%d, got %v", id, ctx.context["id"])
			}

			// Release it
			ctx.Release()
		}(i)
	}

	// Start all goroutines simultaneously
	close(startCh)

	// Wait for all goroutines to complete
	wg.Wait()
}

func TestStringBufferPoolConcurrency(t *testing.T) {
	// Test concurrent access to the string buffer pool
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Create channels to coordinate the test
	startCh := make(chan struct{})

	// Launch goroutines that will get and release buffers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Wait for the start signal
			<-startCh

			// Get a buffer from the pool
			buf := NewStringBuffer()

			// Write to it
			data := []byte("Test from goroutine ")
			buf.Write(data)

			// Release it
			buf.Release()
		}(i)
	}

	// Start all goroutines simultaneously
	close(startCh)

	// Wait for all goroutines to complete
	wg.Wait()
}

func TestNestedRenderContexts(t *testing.T) {
	// Create a parent context
	parent := NewRenderContext(nil, map[string]interface{}{"parent": "value"}, nil)

	// Create a child context with the parent
	child := NewRenderContext(nil, map[string]interface{}{"child": "value"}, nil)
	child.parent = parent

	// Test variable inheritance
	val, err := child.GetVariable("parent")
	if err != nil {
		t.Errorf("Failed to get parent variable: %v", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got '%v'", val)
	}

	// Release the child - should not release the parent
	child.Release()

	// Parent should still have its data
	if val, exists := parent.context["parent"]; !exists || val != "value" {
		t.Errorf("Parent context corrupted after child release")
	}

	// Release the parent
	parent.Release()
}

func TestAttributeCache(t *testing.T) {
	// Create a test struct
	type TestStruct struct {
		Name string
		Age  int
	}

	// Create a render context
	ctx := NewRenderContext(nil, nil, nil)
	defer ctx.Release()

	obj := TestStruct{
		Name: "John",
		Age:  30,
	}

	// First access - should cache
	name, err := ctx.getAttribute(obj, "Name")
	if err != nil {
		t.Fatalf("Failed to get attribute: %v", err)
	}
	if name != "John" {
		t.Errorf("Expected 'John', got '%v'", name)
	}

	age, err := ctx.getAttribute(obj, "Age")
	if err != nil {
		t.Fatalf("Failed to get attribute: %v", err)
	}
	if age != 30 {
		t.Errorf("Expected 30, got '%v'", age)
	}

	// Second access - should use cache
	name2, err := ctx.getAttribute(obj, "Name")
	if err != nil {
		t.Fatalf("Failed to get attribute from cache: %v", err)
	}
	if name2 != "John" {
		t.Errorf("Expected 'John' from cache, got '%v'", name2)
	}

	// Different object with the same type - should use same cache entry
	obj2 := TestStruct{
		Name: "Jane",
		Age:  25,
	}

	name3, err := ctx.getAttribute(obj2, "Name")
	if err != nil {
		t.Fatalf("Failed to get attribute from second object: %v", err)
	}
	if name3 != "Jane" {
		t.Errorf("Expected 'Jane', got '%v'", name3)
	}
}

func TestAttributeCacheConcurrency(t *testing.T) {
	// Test concurrent access to the attribute cache
	type TestStruct struct {
		Name string
		Age  int
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Create channels to coordinate the test
	startCh := make(chan struct{})

	// Launch goroutines that will access attributes concurrently
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a local context
			ctx := NewRenderContext(nil, nil, nil)
			defer ctx.Release()

			// Create a test object
			obj := TestStruct{
				Name: "Test",
				Age:  id,
			}

			// Wait for the start signal
			<-startCh

			// Access attributes - should use and update cache
			for j := 0; j < 10; j++ {
				_, err := ctx.getAttribute(obj, "Name")
				if err != nil {
					t.Errorf("Failed to get Name attribute: %v", err)
				}

				_, err = ctx.getAttribute(obj, "Age")
				if err != nil {
					t.Errorf("Failed to get Age attribute: %v", err)
				}
			}
		}(i)
	}

	// Start all goroutines simultaneously
	close(startCh)

	// Wait for all goroutines to complete
	wg.Wait()
}
