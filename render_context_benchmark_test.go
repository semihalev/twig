package twig

import (
	"testing"
)

func BenchmarkRenderContextCreation(b *testing.B) {
	engine := New()

	// Create a simple context with a few variables
	contextVars := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		},
		"items": []string{"item1", "item2", "item3"},
		"count": 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := NewRenderContext(engine.environment, contextVars, engine)
		ctx.Release() // Return to pool after use
	}
}

func BenchmarkRenderContextCloning(b *testing.B) {
	engine := New()

	// Create a parent context with some variables, blocks, and macros
	parentContext := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
		"items": []string{"item1", "item2", "item3"},
	}

	// Create the parent context
	parent := NewRenderContext(engine.environment, parentContext, engine)

	// Add some blocks
	header := &TextNode{content: "Header Content", line: 1}
	footer := &TextNode{content: "Footer Content", line: 2}
	parent.blocks["header"] = []Node{header}
	parent.blocks["footer"] = []Node{footer}

	// Add a simple macro
	macroNode := &MacroNode{
		name: "format",
		line: 3,
	}
	parent.macros["format"] = macroNode

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clone the context - this should reuse memory from pools
		child := parent.Clone()
		
		// Do some operations on the child context
		child.SetVariable("newVar", "test value")
		
		// Release the child context
		child.Release()
	}

	// Clean up parent context
	parent.Release()
}

func BenchmarkNestedContextCreation(b *testing.B) {
	engine := New()
	baseContext := map[string]interface{}{"base": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a chain of 5 nested contexts
		ctx1 := NewRenderContext(engine.environment, baseContext, engine)
		ctx2 := ctx1.Clone()
		ctx3 := ctx2.Clone()
		ctx4 := ctx3.Clone()
		ctx5 := ctx4.Clone()

		// Make some changes to test variable lookup
		ctx5.SetVariable("level5", "value5")
		ctx3.SetVariable("level3", "value3")

		// Release in reverse order
		ctx5.Release()
		ctx4.Release()
		ctx3.Release()
		ctx2.Release()
		ctx1.Release()
	}
}

func BenchmarkContextVariableLookup(b *testing.B) {
	engine := New()

	// Create a chain of contexts with variables at different levels
	rootCtx := NewRenderContext(engine.environment, map[string]interface{}{
		"rootVar": "root value",
		"shared":  "root version",
	}, engine)

	level1 := rootCtx.Clone()
	level1.SetVariable("level1Var", "level1 value")
	level1.SetVariable("shared", "level1 version")

	level2 := level1.Clone()
	level2.SetVariable("level2Var", "level2 value")

	// Setup complete, start benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test variable lookup at different levels
		level2.GetVariable("level2Var") // Local var
		level2.GetVariable("level1Var") // Parent var
		level2.GetVariable("rootVar")   // Root var
		level2.GetVariable("shared")    // Shadowed var
		level2.GetVariable("nonExistentVar") // Missing var
	}

	// Clean up
	level2.Release()
	level1.Release()
	rootCtx.Release()
}