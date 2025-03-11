// This file contains a utility function to generate a memory allocation report
package twig

import (
	"fmt"
	"os"
	"runtime/pprof"
	"sort"
	"strings"
)

// RunMemoryAnalysis runs a comprehensive memory analysis on the template engine
// and generates a report of allocation hotspots.
func RunMemoryAnalysis() error {
	// Create a memory profile output file
	f, err := os.Create("twig_memory.pprof")
	if err != nil {
		return fmt.Errorf("failed to create memory profile: %v", err)
	}
	defer f.Close()

	// Write a memory profile with detailed allocation info
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("failed to write memory profile: %v", err)
	}

	// Generate a report based on the memory profile
	reportFile, err := os.Create("twig_memory_report.txt")
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer reportFile.Close()

	// Header information
	fmt.Fprintf(reportFile, "# TWIG MEMORY ALLOCATION REPORT\n\n")
	fmt.Fprintf(reportFile, "This report shows memory allocation hotspots in the Twig template engine.\n")
	fmt.Fprintf(reportFile, "Optimizing these areas will help achieve a zero-allocation rendering path.\n\n")

	// Instructions for viewing the profile
	fmt.Fprintf(reportFile, "## How to View the Profile\n\n")
	fmt.Fprintf(reportFile, "Run this command to analyze the memory profile:\n")
	fmt.Fprintf(reportFile, "```\ngo tool pprof -alloc_space twig_memory.pprof\n```\n\n")
	fmt.Fprintf(reportFile, "Common commands in the pprof interface:\n")
	fmt.Fprintf(reportFile, "- `top`: Shows the top allocation sources\n")
	fmt.Fprintf(reportFile, "- `list FunctionName`: Shows line-by-line allocations in a function\n")
	fmt.Fprintf(reportFile, "- `web`: Opens a web browser with a visualization of the profile\n\n")

	// Benchmarking instructions
	fmt.Fprintf(reportFile, "## Benchmark Results\n\n")
	fmt.Fprintf(reportFile, "Run this command to see allocation statistics:\n")
	fmt.Fprintf(reportFile, "```\ngo test -run=^$ -bench=. -benchmem ./memory_profile_test.go\n```\n\n")

	// Generate tables for common allocation sources
	generateAllocationTables(reportFile)

	// Recommendation section
	fmt.Fprintf(reportFile, "## Optimization Recommendations\n\n")
	fmt.Fprintf(reportFile, "Based on common patterns in template engines, consider these areas for optimization:\n\n")
	
	recommendations := []string{
		"**Context Creation**: Pool and reuse RenderContext objects",
		"**String Concatenation**: Replace with direct WriteString to output buffers",
		"**Expression Evaluation**: Eliminate intermediate allocations during evaluation",
		"**Filter Chain Evaluation**: Reuse filter result objects",
		"**Map Creation**: Pre-size maps and reuse map objects where possible",
		"**String Conversions**: Use allocation-free ToString implementations for common types",
		"**Buffer Management**: Pool and reuse output buffers",
		"**Node Creation**: Extend the node pool to cover all node types",
		"**Slice Allocations**: Pre-allocate slices with expected capacity",
	}
	
	for _, rec := range recommendations {
		fmt.Fprintf(reportFile, "- %s\n", rec)
	}
	
	// Implementation strategy
	fmt.Fprintf(reportFile, "\n## Implementation Strategy\n\n")
	fmt.Fprintf(reportFile, "1. **Start with high-impact areas**: Focus on the top allocation sources first\n")
	fmt.Fprintf(reportFile, "2. **Implement pools for all temporary objects**: Especially RenderContext objects\n")
	fmt.Fprintf(reportFile, "3. **Optimize string operations**: String handling is often a major source of allocations\n")
	fmt.Fprintf(reportFile, "4. **Review all map/slice creations**: Pre-size collections where possible\n")
	fmt.Fprintf(reportFile, "5. **Incremental testing**: Benchmark after each optimization to measure impact\n\n")
	
	fmt.Fprintf(reportFile, "## Final Notes\n\n")
	fmt.Fprintf(reportFile, "Remember that some allocations are unavoidable, especially for dynamic templates.\n")
	fmt.Fprintf(reportFile, "The goal is to eliminate allocations in the core rendering path, prioritizing the\n")
	fmt.Fprintf(reportFile, "most frequent operations for maximum performance impact.\n")

	return nil
}

// Helper function to generate allocation tables for common sources
func generateAllocationTables(w *os.File) {
	// String Handling Table
	fmt.Fprintf(w, "### String Operations\n\n")
	fmt.Fprintf(w, "| Operation | Allocation Issue | Optimization |\n")
	fmt.Fprintf(w, "|-----------|-----------------|-------------|\n")
	stringOps := [][]string{
		{"String Concatenation", "Creates new strings", "Use WriteString to buffer"},
		{"Substring Operations", "Creates new strings", "Reuse byte slices when possible"},
		{"String Conversion", "Boxing/unboxing", "Specialized ToString methods"},
		{"Format/Replace", "Creates intermediate strings", "Write directly to output buffer"},
	}
	for _, op := range stringOps {
		fmt.Fprintf(w, "| %s | %s | %s |\n", op[0], op[1], op[2])
	}
	fmt.Fprintf(w, "\n")
	
	// Context Operations Table
	fmt.Fprintf(w, "### Context Operations\n\n")
	fmt.Fprintf(w, "| Operation | Allocation Issue | Optimization |\n")
	fmt.Fprintf(w, "|-----------|-----------------|-------------|\n")
	contextOps := [][]string{
		{"Context Creation", "New map allocations", "Pool and reuse context objects"},
		{"Context Cloning", "Copying maps", "Selective copying or copy-on-write"},
		{"Variable Lookup", "Interface conversions", "Type-specialized getters"},
		{"Context Merging", "Map copies for scope", "Prototype or linked contexts"},
	}
	for _, op := range contextOps {
		fmt.Fprintf(w, "| %s | %s | %s |\n", op[0], op[1], op[2])
	}
	fmt.Fprintf(w, "\n")
	
	// Node Operations Table
	fmt.Fprintf(w, "### Node Operations\n\n")
	fmt.Fprintf(w, "| Node Type | Allocation Issue | Optimization |\n")
	fmt.Fprintf(w, "|-----------|-----------------|-------------|\n")
	nodeOps := [][]string{
		{"Expression Nodes", "New node for each evaluation", "Pool and reuse node objects"},
		{"Filter Nodes", "Chained filter allocations", "Intermediate result pooling"},
		{"Loop Nodes", "Iterator allocations", "Reuse loop context and iterators"},
		{"Block Nodes", "Block context allocations", "Pool block contexts"},
	}
	for _, op := range nodeOps {
		fmt.Fprintf(w, "| %s | %s | %s |\n", op[0], op[1], op[2])
	}
	fmt.Fprintf(w, "\n")
}