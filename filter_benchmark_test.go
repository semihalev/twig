package twig

import (
	"fmt"
	"html"
	"strings"
	"testing"
)

// Benchmark HTML escape filter
func BenchmarkHtmlEscapeFilter(b *testing.B) {
	ctx := NewRenderContext(nil, nil, nil)
	defer ctx.Release()

	// Create a string with all the special characters
	testString := `This is a "test" with <tags> & special 'characters' that need escaping`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, _ := ctx.ApplyFilter("escape", testString)
		_ = result
	}
}

// Benchmark the original nested strings.Replace approach (for comparison)
func BenchmarkHtmlEscapeOriginal(b *testing.B) {
	testString := `This is a "test" with <tags> & special 'characters' that need escaping`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := strings.Replace(
			strings.Replace(
				strings.Replace(
					strings.Replace(
						strings.Replace(
							testString,
							"&", "&amp;", -1),
						"<", "&lt;", -1),
					">", "&gt;", -1),
				"\"", "&quot;", -1),
			"'", "&#39;", -1)
		_ = result
	}
}

// Benchmark the standard library's html.EscapeString (for comparison)
func BenchmarkHtmlEscapeStdLib(b *testing.B) {
	testString := `This is a "test" with <tags> & special 'characters' that need escaping`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := html.EscapeString(testString)
		_ = result
	}
}

// Helper for filter chain benchmarks - builds a deep filter chain
func createFilterChain(depth int) *FilterNode {
	var node Node

	// Create a literal base node
	node = NewLiteralNode("test", 1)

	// Add filters in sequence
	for i := 0; i < depth; i++ {
		// Use common filters
		var filterName string
		switch i % 4 {
		case 0:
			filterName = "upper"
		case 1:
			filterName = "lower"
		case 2:
			filterName = "capitalize"
		case 3:
			filterName = "escape"
		}

		node = NewFilterNode(node, filterName, nil, 1)
	}

	return node.(*FilterNode)
}

// Benchmark filter chain with different depths
func BenchmarkFilterChainDepth1(b *testing.B) {
	benchmarkFilterChain(b, 1)
}

func BenchmarkFilterChainDepth5(b *testing.B) {
	benchmarkFilterChain(b, 5)
}

func BenchmarkFilterChainDepth10(b *testing.B) {
	benchmarkFilterChain(b, 10)
}

func benchmarkFilterChain(b *testing.B, depth int) {
	// Create environment with test filters
	env := &Environment{
		filters: make(map[string]FilterFunc),
	}

	// Add filters directly to the map since Environment doesn't expose AddFilter
	env.filters["upper"] = func(value interface{}, args ...interface{}) (interface{}, error) {
		return strings.ToUpper(value.(string)), nil
	}
	env.filters["lower"] = func(value interface{}, args ...interface{}) (interface{}, error) {
		return strings.ToLower(value.(string)), nil
	}
	env.filters["capitalize"] = func(value interface{}, args ...interface{}) (interface{}, error) {
		s := value.(string)
		if s == "" {
			return s, nil
		}
		return strings.ToUpper(s[:1]) + s[1:], nil
	}
	env.filters["escape"] = func(value interface{}, args ...interface{}) (interface{}, error) {
		return html.EscapeString(value.(string)), nil
	}

	ctx := NewRenderContext(env, nil, nil)
	defer ctx.Release()

	// Create a filter chain of the specified depth
	filterNode := createFilterChain(depth)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.evaluateFilterNode(filterNode)
	}
}

// Helper function specific to this benchmark
func benchmarkToString(value interface{}) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", value)
}
