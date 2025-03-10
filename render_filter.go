package twig

import (
	"fmt"
)

// ApplyFilter applies a filter to a value
func (ctx *RenderContext) ApplyFilter(name string, value interface{}, args ...interface{}) (interface{}, error) {
	// Look for the filter in the environment
	if ctx.env != nil {
		if filter, ok := ctx.env.filters[name]; ok {
			return filter(value, args...)
		}
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
	var chain []FilterChainItem
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

		// Add this filter to the chain (prepend to maintain order)
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

	// Apply the entire filter chain in a single operation
	return ctx.ApplyFilterChain(value, filterChain)
}
