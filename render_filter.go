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
		// Simple HTML escape
		return strings.Replace(
			strings.Replace(
				strings.Replace(
					strings.Replace(
						strings.Replace(
							ctx.ToString(value),
							"&", "&amp;", -1),
						"<", "&lt;", -1),
					">", "&gt;", -1),
				"\"", "&quot;", -1),
			"'", "&#39;", -1), nil
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
	// Preallocate with a reasonable capacity for typical filter chains
	chain := make([]FilterChainItem, 0, 4)
	currentNode := node

	// Traverse down the filter chain, collecting filters as we go
	for {
		// Check if the current node is a filter node
		filterNode, isFilter := currentNode.(*FilterNode)
		if !isFilter {
			// We've reached the base value node
			break
		}

		// Evaluate filter arguments
		args := make([]interface{}, len(filterNode.args))
		for i, arg := range filterNode.args {
			val, err := ctx.EvaluateExpression(arg)
			if err != nil {
				return nil, nil, err
			}
			args[i] = val
		}

		// Insert this filter at the beginning of the chain
		// (prepend to maintain order)
		chain = append([]FilterChainItem{{
			name: filterNode.filter,
			args: args,
		}}, chain...)

		// Continue with the next node in the chain
		currentNode = filterNode.node
	}

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
