package twig

import (
	"bytes"
	"testing"
)

// Benchmark for expression evaluation
func BenchmarkExpressionEvaluation(b *testing.B) {
	engine := New()
	ctx := NewRenderContext(engine.environment, map[string]interface{}{
		"a": 10,
		"b": 20,
		"c": "hello",
		"d": []interface{}{1, 2, 3, 4, 5},
		"e": map[string]interface{}{
			"f": "world",
			"g": 42,
		},
	}, engine)
	defer ctx.Release()

	// Setup expression nodes for testing
	tests := []struct {
		name string
		node Node
	}{
		{
			name: "LiteralNode",
			node: NewLiteralNode("test string", 1),
		},
		{
			name: "VariableNode",
			node: NewVariableNode("a", 1),
		},
		{
			name: "BinaryNode-Simple",
			node: NewBinaryNode("+", NewVariableNode("a", 1), NewVariableNode("b", 1), 1),
		},
		{
			name: "BinaryNode-Complex",
			node: NewBinaryNode(
				"+",
				NewBinaryNode("*", NewVariableNode("a", 1), NewLiteralNode(2, 1), 1),
				NewBinaryNode("/", NewVariableNode("b", 1), NewLiteralNode(4, 1), 1),
				1,
			),
		},
		{
			name: "GetAttrNode",
			node: NewGetAttrNode(NewVariableNode("e", 1), NewLiteralNode("f", 1), 1),
		},
		{
			name: "GetItemNode",
			node: NewGetItemNode(NewVariableNode("d", 1), NewLiteralNode(2, 1), 1),
		},
		{
			name: "ArrayNode",
			node: NewArrayNode([]Node{
				NewVariableNode("a", 1),
				NewVariableNode("b", 1),
				NewLiteralNode("test", 1),
			}, 1),
		},
		{
			name: "HashNode",
			node: func() *HashNode {
				items := make(map[Node]Node)
				items[NewLiteralNode("key1", 1)] = NewVariableNode("a", 1)
				items[NewLiteralNode("key2", 1)] = NewVariableNode("b", 1)
				items[NewLiteralNode("key3", 1)] = NewLiteralNode("value", 1)
				return NewHashNode(items, 1)
			}(),
		},
		{
			name: "ConditionalNode",
			node: NewConditionalNode(
				NewBinaryNode(">", NewVariableNode("a", 1), NewLiteralNode(5, 1), 1),
				NewVariableNode("b", 1),
				NewLiteralNode(0, 1),
				1,
			),
		},
	}

	// Run each benchmark
	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = ctx.EvaluateExpression(tc.node)
			}
		})
	}
}

// BenchmarkExpressionRender benchmarks the rendering of expressions
func BenchmarkExpressionRender(b *testing.B) {
	engine := New()
	ctx := NewRenderContext(engine.environment, map[string]interface{}{
		"a": 10,
		"b": 20,
		"c": "hello",
		"d": []interface{}{1, 2, 3, 4, 5},
		"e": map[string]interface{}{
			"f": "world",
			"g": 42,
		},
	}, engine)
	defer ctx.Release()

	// Setup expression nodes for testing
	tests := []struct {
		name string
		node Node
	}{
		{
			name: "LiteralNode",
			node: NewLiteralNode("test string", 1),
		},
		{
			name: "VariableNode",
			node: NewVariableNode("a", 1),
		},
		{
			name: "BinaryNode",
			node: NewBinaryNode("+", NewVariableNode("a", 1), NewVariableNode("b", 1), 1),
		},
		{
			name: "GetAttrNode",
			node: NewGetAttrNode(NewVariableNode("e", 1), NewLiteralNode("f", 1), 1),
		},
		{
			name: "GetItemNode",
			node: NewGetItemNode(NewVariableNode("d", 1), NewLiteralNode(2, 1), 1),
		},
	}

	// Create a buffer for testing
	buf := bytes.NewBuffer(nil)

	// Run each benchmark
	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				_ = tc.node.Render(buf, ctx)
			}
		})
	}
}

// BenchmarkFilterChain benchmarks filter chain evaluation
func BenchmarkFilterChain(b *testing.B) {
	engine := New()
	ctx := NewRenderContext(engine.environment, map[string]interface{}{
		"a":    10,
		"text": "Hello, World!",
		"html": "<p>This is a paragraph</p>",
	}, engine)
	defer ctx.Release()

	// Setup filter nodes for testing
	tests := []struct {
		name string
		node Node
	}{
		{
			name: "SingleFilter",
			node: NewFilterNode(NewVariableNode("text", 1), "upper", nil, 1),
		},
		{
			name: "FilterChain",
			node: NewFilterNode(
				NewFilterNode(
					NewVariableNode("text", 1),
					"upper",
					nil,
					1,
				),
				"trim",
				nil,
				1,
			),
		},
		{
			name: "FilterWithArgs",
			node: NewFilterNode(
				NewVariableNode("text", 1),
				"replace",
				[]Node{
					NewLiteralNode("World", 1),
					NewLiteralNode("Universe", 1),
				},
				1,
			),
		},
	}

	// Run each benchmark
	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = ctx.EvaluateExpression(tc.node)
			}
		})
	}
}

// BenchmarkFunctionCall benchmarks function calls
func BenchmarkFunctionCall(b *testing.B) {
	engine := New()
	ctx := NewRenderContext(engine.environment, map[string]interface{}{
		"numbers": []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		"start":   1,
		"end":     10,
	}, engine)
	defer ctx.Release()

	// Setup function nodes for testing
	tests := []struct {
		name string
		node Node
	}{
		{
			name: "RangeFunction",
			node: NewFunctionNode("range", []Node{
				NewVariableNode("start", 1),
				NewVariableNode("end", 1),
			}, 1),
		},
		{
			name: "LengthFunction",
			node: NewFunctionNode("length", []Node{
				NewVariableNode("numbers", 1),
			}, 1),
		},
		{
			name: "MaxFunction",
			node: NewFunctionNode("max", []Node{
				NewVariableNode("numbers", 1),
			}, 1),
		},
	}

	// Run each benchmark
	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = ctx.EvaluateExpression(tc.node)
			}
		})
	}
}

// BenchmarkArgSlicePooling specifically tests array arguments allocation
func BenchmarkArgSlicePooling(b *testing.B) {
	engine := New()
	ctx := NewRenderContext(engine.environment, nil, engine)
	defer ctx.Release()

	smallArgs := []Node{
		NewLiteralNode(1, 1),
		NewLiteralNode(2, 1),
	}

	mediumArgs := []Node{
		NewLiteralNode(1, 1),
		NewLiteralNode(2, 1),
		NewLiteralNode(3, 1),
		NewLiteralNode(4, 1),
		NewLiteralNode(5, 1),
	}

	largeArgs := make([]Node, 10)
	for i := 0; i < 10; i++ {
		largeArgs[i] = NewLiteralNode(i, 1)
	}

	tests := []struct {
		name string
		node Node
	}{
		{
			name: "NoArgs",
			node: NewFunctionNode("range", nil, 1),
		},
		{
			name: "SmallArgs",
			node: NewFunctionNode("range", smallArgs, 1),
		},
		{
			name: "MediumArgs",
			node: NewFunctionNode("range", mediumArgs, 1),
		},
		{
			name: "LargeArgs",
			node: NewFunctionNode("range", largeArgs, 1),
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = ctx.EvaluateExpression(tc.node)
			}
		})
	}
}
