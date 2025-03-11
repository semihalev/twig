# Zero-Allocation Rendering Path Implementation Plan

This document outlines the strategy for implementing a zero-allocation rendering path in the Twig template engine based on detailed memory profiling.

## Memory Allocation Analysis Results

Our detailed profiling shows the following memory allocation hotspots, listed in order of priority:

1. **Template Loading**: 3,303 bytes/op, 40 objects/op
   - The highest per-operation allocation occurs during template parsing and node creation

2. **RenderContext Creation/Cloning**: 409 bytes/op, 4 objects/op
   - Context objects allocate maps and slices for variable storage

3. **String Operations**: 85 bytes/op, 2 objects/op
   - String manipulations during rendering create new allocations

4. **Expression Evaluation**: 86 bytes/op, 2 objects/op
   - Evaluating expressions creates temporary objects

5. **Filters and Functions**: 86 bytes/op, 2 objects/op
   - Filter chains allocate intermediate results

## Implementation Strategy

Based on the profiling results, we should implement optimizations in this order:

### Phase 1: Template Node Pooling

1. **Extend Node Pool System**
   - Expand the existing pool to cover all node types
   - Implement a node recycling mechanism for the parser
   - Pre-allocate node slices with appropriate capacity

2. **Tokenizer Optimization**
   - Reuse token buffers
   - Implement zero-allocation token scanning
   - Pool token objects

### Phase 2: RenderContext Optimization

1. **Context Object Pooling**
   - Extend the current RenderContext pool
   - Ensure proper cleanup on Release()
   - Pre-size context maps based on template requirements

2. **Linked Context Implementation**
   - Replace map copying with linked context references
   - Implement copy-on-write for variable updates
   - Pool nested context objects

### Phase 3: String Handling Improvements

1. **Direct Writer Output**
   - Replace string concatenation with direct io.Writer calls
   - Implement specialized WriteString methods
   - Use buffer pooling for temporary string operations

2. **Type-Specialized ToString Methods**
   - Create non-allocating ToString implementations for common types
   - Implement direct writing of primitive values
   - Pool string builders for complex conversions

### Phase 4: Expression Evaluation Optimization

1. **Expression Result Pooling**
   - Pool temporary objects used during expression evaluation
   - Reuse expression result containers
   - Create specialized evaluators for common expressions

2. **Reduce Interface Conversions**
   - Minimize boxing/unboxing of primitive types
   - Use type assertions instead of reflective operations where possible
   - Implement specialized paths for known types

### Phase 5: Filter and Function Optimization

1. **Filter Chain Optimization**
   - Pool filter result objects
   - Implement direct-to-writer filters
   - Eliminate intermediate allocations in filter chains

2. **Function Argument Handling**
   - Pre-allocate function argument slices
   - Reuse argument containers
   - Implement zero-allocation parameter passing

## Implementation Details

### Enhanced Object Pooling

Extend the object pool system to more types:

```go
// NodePool for each node type
var (
    variableNodePool = sync.Pool{
        New: func() interface{} { return &VariableNode{} },
    }
    filterNodePool = sync.Pool{
        New: func() interface{} { return &FilterNode{} },
    }
    // ... more pools for other node types
)

// Get/Release functions for each type
func GetVariableNode() *VariableNode {
    node := variableNodePool.Get().(*VariableNode)
    // Reset state
    node.name = ""
    return node
}

func ReleaseVariableNode(node *VariableNode) {
    // Clear references to help GC
    variableNodePool.Put(node)
}
```

### Zero-Allocation String Operations

Replace string concatenation with direct writing:

```go
// Instead of:
func (n *TextNode) Render(w io.Writer, ctx *RenderContext) error {
    result := n.content
    _, err := w.Write([]byte(result))
    return err
}

// Use:
func (n *TextNode) Render(w io.Writer, ctx *RenderContext) error {
    _, err := WriteString(w, n.content)
    return err
}
```

### Pre-sized Collections

Pre-allocate maps and slices with appropriate capacity:

```go
// Instead of:
children := make([]Node, 0)
for _, token := range tokens {
    // Process and append nodes
}

// Use:
children := make([]Node, 0, len(tokens)) // Pre-allocate with expected capacity
for _, token := range tokens {
    // Process and append nodes
}
```

### Linked Context Implementation

Replace copying with linking:

```go
// LinkedRenderContext holds a reference to a parent context
type LinkedRenderContext struct {
    parent *RenderContext
    local  map[string]interface{} // Only stores overrides
}

// GetVariable tries local context first, then parent
func (c *LinkedRenderContext) GetVariable(name string) (interface{}, error) {
    if val, ok := c.local[name]; ok {
        return val, nil
    }
    if c.parent != nil {
        return c.parent.GetVariable(name)
    }
    return nil, ErrUndefinedVar
}
```

## Benchmarking

Implement benchmarks for each optimization phase:

```go
func BenchmarkRenderZeroAlloc(b *testing.B) {
    engine := New()
    err := engine.RegisterString("template", "Hello, {{ name }}!")
    if err != nil {
        b.Fatalf("Error registering template: %v", err)
    }

    context := map[string]interface{}{
        "name": "World",
    }

    b.ReportAllocs() // Should report 0 allocs once optimized
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        var buf bytes.Buffer
        template, _ := engine.Load("template")
        template.RenderTo(&buf, context)
    }
}
```

## Phased Implementation Approach

1. **Implement one component at a time**, starting with the highest allocation sources
2. **Add benchmarks for each component** to validate zero allocations
3. **Maintain backward compatibility** with existing API
4. **Integrate incrementally** to avoid breaking changes
5. **Document the zero-allocation API** for users

## Expected Outcome

After implementing all optimizations, we expect:
- **Zero allocations** for most common template rendering operations
- **Significantly reduced allocations** for complex templates
- **Better performance under high load** due to reduced GC pressure
- **More predictable performance** with fewer GC pauses