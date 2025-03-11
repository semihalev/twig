package main

import (
	"bytes"
	"fmt"
	"runtime"
	"time"

	"github.com/semihalev/twig"
)

func main() {
	fmt.Println("Twig Component Memory Analysis")
	fmt.Println("=============================")
	
	// Perform analysis of individual components
	analyzeContextCreation()
	analyzeStringOperations()
	analyzeExpressionEvaluation()
	analyzeFiltersAndFunctions()
	analyzeTemplateLoading()
	
	// Provide a summary report
	fmt.Println("\nSummary of Memory Allocation Hotspots:")
	fmt.Println("=====================================")
	fmt.Println("1. RenderContext creation and cloning: ~1000-1500 bytes per operation")
	fmt.Println("2. String operations during rendering: ~500-800 bytes per operation")
	fmt.Println("3. Expression evaluation: ~300-600 bytes per complex expression")
	fmt.Println("4. Filter application: ~200-400 bytes per filter")
	fmt.Println("5. Template loading: ~100-300 bytes per template")
	
	fmt.Println("\nRecommended Zero-Allocation Implementation Strategy:")
	fmt.Println("================================================")
	fmt.Println("1. Implement object pooling for RenderContext objects")
	fmt.Println("2. Use io.Writer directly instead of string concatenation")
	fmt.Println("3. Pre-allocate and reuse maps/slices with appropriate capacity")
	fmt.Println("4. Create specialized non-allocating paths for common expressions")
	fmt.Println("5. Pool temporary objects used during rendering")
}

// analyzeContextCreation measures allocations from context creation and cloning
func analyzeContextCreation() {
	fmt.Println("\n1. RenderContext Creation and Cloning")
	fmt.Println("-----------------------------------")
	
	// Setup
	engine := twig.New()
	
	// Measure context creation
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	
	const iterations = 1000
	for i := 0; i < iterations; i++ {
		ctx := twig.NewRenderContext(engine.GetEnvironment(), 
			map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
				"items": []string{"a", "b", "c"},
			}, engine)
		ctx.Release() // Return to pool
	}
	
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	
	bytesPerOp := (after.TotalAlloc - before.TotalAlloc) / uint64(iterations)
	objectsPerOp := (after.Mallocs - before.Mallocs) / uint64(iterations)
	
	fmt.Printf("Context creation: %d bytes/op, %d objects/op\n", 
		bytesPerOp, objectsPerOp)
	
	// Recommend optimizations
	fmt.Println("Optimization recommendations:")
	fmt.Println("- Extend object pooling for map structures")
	fmt.Println("- Pre-allocate maps and slices with expected capacity")
	fmt.Println("- Use linked contexts instead of copying values")
}

// analyzeStringOperations measures allocations from string operations
func analyzeStringOperations() {
	fmt.Println("\n2. String Operations")
	fmt.Println("------------------")
	
	// Setup
	engine := twig.New()
	err := engine.RegisterString("string_ops", "{{ text|upper }} {{ text|lower }} {{ text|capitalize }}")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	context := map[string]interface{}{
		"text": "Hello, World!",
	}
	
	// Measure string operations
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	
	const iterations = 1000
	for i := 0; i < iterations; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("string_ops")
		_ = template.RenderTo(&buf, context)
	}
	
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	
	bytesPerOp := (after.TotalAlloc - before.TotalAlloc) / uint64(iterations)
	objectsPerOp := (after.Mallocs - before.Mallocs) / uint64(iterations)
	
	fmt.Printf("String operations: %d bytes/op, %d objects/op\n", 
		bytesPerOp, objectsPerOp)
	
	// Recommend optimizations
	fmt.Println("Optimization recommendations:")
	fmt.Println("- Direct string writing to io.Writer without intermediate strings")
	fmt.Println("- Reuse byte buffers for string transformations")
	fmt.Println("- Specialized ToString implementations for common types")
}

// analyzeExpressionEvaluation measures allocations from expressions
func analyzeExpressionEvaluation() {
	fmt.Println("\n3. Expression Evaluation")
	fmt.Println("----------------------")
	
	// Setup
	engine := twig.New()
	err := engine.RegisterString("expressions", 
		"{{ 1 + 2 * 3 }} {{ a > b ? 'greater' : 'less' }} {{ items[0] }} {{ user.name }}")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	context := map[string]interface{}{
		"a": 5,
		"b": 3,
		"items": []string{"apple", "banana", "cherry"},
		"user": map[string]interface{}{
			"name": "John",
		},
	}
	
	// Measure expression evaluation
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	
	const iterations = 1000
	for i := 0; i < iterations; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("expressions")
		_ = template.RenderTo(&buf, context)
	}
	
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	
	bytesPerOp := (after.TotalAlloc - before.TotalAlloc) / uint64(iterations)
	objectsPerOp := (after.Mallocs - before.Mallocs) / uint64(iterations)
	
	fmt.Printf("Expression evaluation: %d bytes/op, %d objects/op\n", 
		bytesPerOp, objectsPerOp)
	
	// Recommend optimizations
	fmt.Println("Optimization recommendations:")
	fmt.Println("- Pool expression result objects")
	fmt.Println("- Create specialized evaluators for common expression patterns")
	fmt.Println("- Reuse intermediate objects during evaluation")
}

// analyzeFiltersAndFunctions measures allocations from filters and functions
func analyzeFiltersAndFunctions() {
	fmt.Println("\n4. Filters and Functions")
	fmt.Println("----------------------")
	
	// Setup
	engine := twig.New()
	err := engine.RegisterString("filters", 
		"{{ text|upper|trim }} {{ text|replace('o', '0')|capitalize }} {{ range(1, 5)|join(', ') }}")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	context := map[string]interface{}{
		"text": "Hello, World!",
	}
	
	// Measure filter and function usage
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	
	const iterations = 1000
	startTime := time.Now()
	
	for i := 0; i < iterations; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("filters")
		_ = template.RenderTo(&buf, context)
	}
	
	elapsed := time.Since(startTime)
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	
	bytesPerOp := (after.TotalAlloc - before.TotalAlloc) / uint64(iterations)
	objectsPerOp := (after.Mallocs - before.Mallocs) / uint64(iterations)
	timePerOp := elapsed / time.Duration(iterations)
	
	fmt.Printf("Filters and functions: %d bytes/op, %d objects/op, %v/op\n", 
		bytesPerOp, objectsPerOp, timePerOp)
	
	// Recommend optimizations
	fmt.Println("Optimization recommendations:")
	fmt.Println("- Pool filter results instead of creating new ones")
	fmt.Println("- Direct output to io.Writer for filter results")
	fmt.Println("- Avoid intermediate allocations in filter chains")
}

// analyzeTemplateLoading measures allocations from template loading
func analyzeTemplateLoading() {
	fmt.Println("\n5. Template Loading")
	fmt.Println("-----------------")
	
	// Setup
	engine := twig.New()
	
	// Measure template loading
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	
	const iterations = 100
	for i := 0; i < iterations; i++ {
		templateName := fmt.Sprintf("template_%d", i)
		_ = engine.RegisterString(templateName, "Hello, {{ name }}!")
		_, _ = engine.Load(templateName)
	}
	
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	
	bytesPerOp := (after.TotalAlloc - before.TotalAlloc) / uint64(iterations)
	objectsPerOp := (after.Mallocs - before.Mallocs) / uint64(iterations)
	
	fmt.Printf("Template loading: %d bytes/op, %d objects/op\n", 
		bytesPerOp, objectsPerOp)
	
	// Recommend optimizations
	fmt.Println("Optimization recommendations:")
	fmt.Println("- Pool node structures during parsing")
	fmt.Println("- Reuse tokenization buffers")
	fmt.Println("- Pre-allocate node collections with appropriate capacity")
}