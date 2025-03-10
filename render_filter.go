package twig

import "fmt"

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

// Override for the original FilterNode evaluation in render.go
func (ctx *RenderContext) evaluateFilterNode(n *FilterNode) (interface{}, error) {
	// Evaluate the base value
	value, err := ctx.EvaluateExpression(n.node)
	if err != nil {
		return nil, err
	}

	// Evaluate filter arguments
	args := make([]interface{}, len(n.args))
	for i, arg := range n.args {
		val, err := ctx.EvaluateExpression(arg)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	return ctx.ApplyFilter(n.filter, value, args...)
}