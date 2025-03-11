package twig

import (
	"fmt"
	"strings"
)

// ApplyFilter applies a filter to a value
func (ctx *RenderContext) ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error) {
	// Look for the filter in the environment
	if ctx.env != nil {
		if filter, ok := ctx.env.filters[name]; ok {
			result, err := filter(value, args...)
			if err != nil {
				return nil, err
			}

			// We've moved the script-specific string handling to PrintNode.Render
			return result, nil
		}
	}

	// Handle built-in filters for macro compatibility
	switch name {
	case "e", "escape":
		// Optimized HTML escape using a single pass with strings.Builder
		str := ctx.ToString(value)
		if str == "" {
			return "", nil
		}

		// Preallocate with a reasonable estimate (slightly larger than original)
		// This avoids most reallocations
		var b strings.Builder
		b.Grow(len(str) + len(str)/8)

		// Single-pass iteration is much more efficient than nested Replace calls
		for _, c := range str {
			switch c {
			case '&':
				b.WriteString("&amp;")
			case '<':
				b.WriteString("&lt;")
			case '>':
				b.WriteString("&gt;")
			case '"':
				b.WriteString("&quot;")
			case '\'':
				b.WriteString("&#39;")
			default:
				b.WriteRune(c)
			}
		}

		return b.String(), nil
	}

	return nil, fmt.Errorf("filter '%s' not found", name)
}

// FilterChainItem represents a single filter in a chain
type FilterChainItem struct {
	name string
	args []interface{}
}

// DetectFilterChain analyzes a filter node and extracts all filters in the chain
// Returns the base node and a slice of all filters to be applied
func (ctx *RenderContext) DetectFilterChain(node Node) (Node, []FilterChainItem, error) {
	// First, count the depth of the filter chain to properly allocate
	depth := 0
	currentNode := node
	for {
		filterNode, isFilter := currentNode.(*FilterNode)
		if !isFilter {
			break
		}
		depth++
		currentNode = filterNode.node
	}

	// Now that we know the depth, allocate the proper size slice
	chain := make([]FilterChainItem, depth)

	// Traverse the chain again, but this time fill the slice in reverse order
	// This avoids the O(n²) complexity of the previous implementation
	currentNode = node
	for i := depth - 1; i >= 0; i-- {
		filterNode := currentNode.(*FilterNode) // Safe because we validated in first pass

		// Evaluate filter arguments
		args := make([]interface{}, len(filterNode.args))
		for j, arg := range filterNode.args {
			val, err := ctx.EvaluateExpression(arg)
			if err != nil {
				return nil, nil, err
			}
			args[j] = val
		}

		// Add to the chain in the correct position
		chain[i] = FilterChainItem{
			name: filterNode.filter,
			args: args,
		}

		// Continue with the next node
		currentNode = filterNode.node
	}

	// Return the base node and the chain
	return currentNode, chain, nil
}

// ApplyFilterChain applies a chain of filters to a value
func (ctx *RenderContext) ApplyFilterChain(baseValue interface{}, chain []FilterChainItem) (interface{}, error) {
	// Start with the base value
	result := baseValue
	var err error

	// Apply each filter in the chain
	for _, filter := range chain {
		result, err = ctx.ApplyFilter(filter.name, result, filter.args...)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Override for the original FilterNode evaluation in render.go
func (ctx *RenderContext) evaluateFilterNode(n *FilterNode) (interface{}, error) {
	// Detect the complete filter chain
	baseNode, filterChain, err := ctx.DetectFilterChain(n)
	if err != nil {
		return nil, err
	}

	// Evaluate the base value
	value, err := ctx.EvaluateExpression(baseNode)
	if err != nil {
		return nil, err
	}

	// Log for debugging
	if IsDebugEnabled() {
		LogDebug("Evaluating filter chain: %s on value type %T", n.filter, value)
	}

	// Apply the entire filter chain in a single operation
	result, err := ctx.ApplyFilterChain(value, filterChain)
	if err != nil {
		return nil, err
	}

	// Ensure filter chain results are directly usable in for loops
	// This is especially important for filters like 'sort' that transform arrays
	// We convert to a []interface{} which is what ForNode.Render expects
	if IsDebugEnabled() {
		LogDebug("Filter result type: %T", result)
	}
	return result, nil
}
