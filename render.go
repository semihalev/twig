// Revert to backup file and use the content with our changes
package twig

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RenderContext holds the state during template rendering
type RenderContext struct {
	env          *Environment
	context      map[string]interface{}
	blocks       map[string][]Node
	parentBlocks map[string][]Node // Original block content from parent templates
	macros       map[string]Node
	parent       *RenderContext
	engine       *Engine    // Reference to engine for loading templates
	extending    bool       // Whether this template extends another
	currentBlock *BlockNode // Current block being rendered (for parent() function)
	inParentCall bool       // Flag to indicate if we're currently rendering a parent() call
	sandboxed    bool       // Flag indicating if this context is sandboxed
}

// contextMapPool is a pool for the maps used in RenderContext
var contextMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 16) // Pre-allocate with reasonable size
	},
}

// blocksMapPool is a pool for the blocks map used in RenderContext
var blocksMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string][]Node, 8) // Pre-allocate with reasonable size
	},
}

// macrosMapPool is a pool for the macros map used in RenderContext
var macrosMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]Node, 8) // Pre-allocate with reasonable size
	},
}

// renderContextPool is a sync.Pool for RenderContext objects
var renderContextPool = sync.Pool{
	New: func() interface{} {
		return &RenderContext{
			context:      contextMapPool.Get().(map[string]interface{}),
			blocks:       blocksMapPool.Get().(map[string][]Node),
			parentBlocks: blocksMapPool.Get().(map[string][]Node),
			macros:       macrosMapPool.Get().(map[string]Node),
		}
	},
}

// NewRenderContext gets a RenderContext from the pool and initializes it
func NewRenderContext(env *Environment, context map[string]interface{}, engine *Engine) *RenderContext {
	ctx := renderContextPool.Get().(*RenderContext)

	// Ensure all maps are initialized (should be from the pool)
	if ctx.context == nil {
		ctx.context = contextMapPool.Get().(map[string]interface{})
	} else {
		// Clear any existing data
		for k := range ctx.context {
			delete(ctx.context, k)
		}
	}

	if ctx.blocks == nil {
		ctx.blocks = blocksMapPool.Get().(map[string][]Node)
	} else {
		// Clear any existing data
		for k := range ctx.blocks {
			delete(ctx.blocks, k)
		}
	}

	if ctx.parentBlocks == nil {
		ctx.parentBlocks = blocksMapPool.Get().(map[string][]Node)
	} else {
		// Clear any existing data
		for k := range ctx.parentBlocks {
			delete(ctx.parentBlocks, k)
		}
	}

	if ctx.macros == nil {
		ctx.macros = macrosMapPool.Get().(map[string]Node)
	} else {
		// Clear any existing data
		for k := range ctx.macros {
			delete(ctx.macros, k)
		}
	}

	// Set basic properties
	ctx.env = env
	ctx.engine = engine
	ctx.extending = false
	ctx.currentBlock = nil
	ctx.parent = nil
	ctx.inParentCall = false
	ctx.sandboxed = false

	// Copy the context values directly
	if context != nil {
		for k, v := range context {
			ctx.context[k] = v
		}
	}

	return ctx
}

// Release returns the RenderContext to the pool with proper cleanup
func (ctx *RenderContext) Release() {
	// Clear references to large objects to prevent memory leaks
	ctx.env = nil
	ctx.engine = nil
	ctx.currentBlock = nil

	// Save the maps so we can return them to their respective pools
	contextMap := ctx.context
	blocksMap := ctx.blocks
	parentBlocksMap := ctx.parentBlocks
	macrosMap := ctx.macros

	// Clear the maps from the context
	ctx.context = nil
	ctx.blocks = nil
	ctx.parentBlocks = nil
	ctx.macros = nil

	// Don't release parent contexts - they'll be released separately
	ctx.parent = nil

	// Return to pool
	renderContextPool.Put(ctx)

	// Clear map contents and return them to their pools
	if contextMap != nil {
		for k := range contextMap {
			delete(contextMap, k)
		}
		contextMapPool.Put(contextMap)
	}

	if blocksMap != nil {
		for k := range blocksMap {
			delete(blocksMap, k)
		}
		blocksMapPool.Put(blocksMap)
	}

	if parentBlocksMap != nil {
		for k := range parentBlocksMap {
			delete(parentBlocksMap, k)
		}
		blocksMapPool.Put(parentBlocksMap)
	}

	if macrosMap != nil {
		for k := range macrosMap {
			delete(macrosMap, k)
		}
		macrosMapPool.Put(macrosMap)
	}
}

// Error types
var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrUndefinedVar     = errors.New("undefined variable")
	ErrInvalidAttribute = errors.New("invalid attribute access")
	ErrCompilation      = errors.New("compilation error")
	ErrRender           = errors.New("render error")
)

// GetVariable gets a variable from the context
func (ctx *RenderContext) GetVariable(name string) (interface{}, error) {
	// Check if this looks like an array literal - this is a hack to handle
	// the case where the array literal is parsed as a variable name
	if len(name) >= 2 && name[0] == '[' && strings.Contains(name, "]") {
		// This looks like an array literal that was parsed as a variable name
		// We'll parse it manually here

		// Extract the content between [ and ]
		content := name[1:strings.LastIndex(name, "]")]

		// Split by commas
		parts := strings.Split(content, ",")

		// Create a result array
		result := make([]interface{}, 0, len(parts))

		// Process each element
		for _, part := range parts {
			// Trim whitespace and quotes
			element := strings.TrimSpace(part)

			// If it's a quoted string, remove the quotes
			if len(element) >= 2 && (element[0] == '"' || element[0] == '\'') && element[0] == element[len(element)-1] {
				element = element[1 : len(element)-1]
			}

			result = append(result, element)
		}

		return result, nil
	}

	// Fallback ternary expression parser for backward compatibility
	// This handles cases where the parser didn't correctly handle ternary expressions
	if strings.Contains(name, "?") && strings.Contains(name, ":") {
		LogDebug("Parsing inline ternary expression: %s", name)

		// Simple ternary expression handler
		parts := strings.SplitN(name, "?", 2)
		condition := strings.TrimSpace(parts[0])
		branches := strings.SplitN(parts[1], ":", 2)

		if len(branches) != 2 {
			return nil, fmt.Errorf("malformed ternary expression: %s", name)
		}

		trueExpr := strings.TrimSpace(branches[0])
		falseExpr := strings.TrimSpace(branches[1])

		// Evaluate condition
		var condValue bool
		if condition == "true" {
			condValue = true
		} else if condition == "false" {
			condValue = false
		} else {
			// Try to get variable value
			condVar, _ := ctx.GetVariable(condition)
			condValue = ctx.toBool(condVar)
		}

		// Evaluate the appropriate branch
		if condValue {
			return ctx.GetVariable(trueExpr)
		} else {
			return ctx.GetVariable(falseExpr)
		}
	}

	// Check local context first
	if value, ok := ctx.context[name]; ok {
		return value, nil
	}

	// Check globals
	if ctx.env != nil {
		if value, ok := ctx.env.globals[name]; ok {
			return value, nil
		}
	}

	// Check parent context
	if ctx.parent != nil {
		return ctx.parent.GetVariable(name)
	}

	// Return nil with no error for undefined variables
	// Twig treats undefined variables as empty strings during rendering
	return nil, nil
}

// GetVariableOrNil gets a variable from the context, returning nil silently if not found
func (ctx *RenderContext) GetVariableOrNil(name string) interface{} {
	value, _ := ctx.GetVariable(name)
	return value
}

// SetVariable sets a variable in the context
func (ctx *RenderContext) SetVariable(name string, value interface{}) {
	ctx.context[name] = value
}

// GetEnvironment returns the environment
func (ctx *RenderContext) GetEnvironment() *Environment {
	return ctx.env
}

// GetEngine returns the engine
func (ctx *RenderContext) GetEngine() *Engine {
	return ctx.engine
}

// SetParent sets the parent context
func (ctx *RenderContext) SetParent(parent *RenderContext) {
	ctx.parent = parent
}

// EnableSandbox enables sandbox mode on this context
func (ctx *RenderContext) EnableSandbox() {
	ctx.sandboxed = true
}

// IsSandboxed returns whether this context is sandboxed
func (ctx *RenderContext) IsSandboxed() bool {
	return ctx.sandboxed
}

// Clone creates a new context as a child of the current context
func (ctx *RenderContext) Clone() *RenderContext {
	// Get a new context from the pool with empty maps
	newCtx := renderContextPool.Get().(*RenderContext)

	// Initialize the context
	newCtx.env = ctx.env
	newCtx.engine = ctx.engine
	newCtx.extending = false
	newCtx.currentBlock = nil
	newCtx.parent = ctx
	newCtx.inParentCall = false

	// Inherit sandbox state
	newCtx.sandboxed = ctx.sandboxed

	// Ensure maps are initialized (they should be from the pool already)
	if newCtx.context == nil {
		newCtx.context = contextMapPool.Get().(map[string]interface{})
	} else {
		// Clear any existing data
		for k := range newCtx.context {
			delete(newCtx.context, k)
		}
	}

	if newCtx.blocks == nil {
		newCtx.blocks = blocksMapPool.Get().(map[string][]Node)
	} else {
		// Clear any existing data
		for k := range newCtx.blocks {
			delete(newCtx.blocks, k)
		}
	}

	if newCtx.macros == nil {
		newCtx.macros = macrosMapPool.Get().(map[string]Node)
	} else {
		// Clear any existing data
		for k := range newCtx.macros {
			delete(newCtx.macros, k)
		}
	}

	if newCtx.parentBlocks == nil {
		newCtx.parentBlocks = blocksMapPool.Get().(map[string][]Node)
	} else {
		// Clear any existing data
		for k := range newCtx.parentBlocks {
			delete(newCtx.parentBlocks, k)
		}
	}

	// Copy blocks by reference (no need to deep copy)
	for name, nodes := range ctx.blocks {
		newCtx.blocks[name] = nodes
	}

	// Copy macros by reference (no need to deep copy)
	for name, macro := range ctx.macros {
		newCtx.macros[name] = macro
	}

	return newCtx
}

// GetMacro gets a macro from the context
func (ctx *RenderContext) GetMacro(name string) (interface{}, bool) {
	// Check local macros first
	if macro, ok := ctx.macros[name]; ok {
		return macro, true
	}

	// Check parent context
	if ctx.parent != nil {
		return ctx.parent.GetMacro(name)
	}

	return nil, false
}

// GetMacros returns the macros map
func (ctx *RenderContext) GetMacros() map[string]Node {
	return ctx.macros
}

// InitMacros initializes the macros map if it's nil
func (ctx *RenderContext) InitMacros() {
	if ctx.macros == nil {
		ctx.macros = macrosMapPool.Get().(map[string]Node)
	}
}

// SetMacro sets a macro in the context
func (ctx *RenderContext) SetMacro(name string, macro Node) {
	if ctx.macros == nil {
		ctx.macros = macrosMapPool.Get().(map[string]Node)
	}
	ctx.macros[name] = macro
}

// CallMacro calls a macro with the given arguments
func (ctx *RenderContext) CallMacro(w io.Writer, name string, args []interface{}) error {
	// Find the macro
	macro, ok := ctx.GetMacro(name)
	if !ok {
		return fmt.Errorf("macro '%s' not found", name)
	}

	// Check if it's a MacroNode
	macroNode, ok := macro.(*MacroNode)
	if !ok {
		return fmt.Errorf("'%s' is not a macro", name)
	}

	// Call the macro
	return macroNode.CallMacro(w, ctx, args...)
}

// CallFunction calls a function with the given arguments
func (ctx *RenderContext) CallFunction(name string, args []interface{}) (interface{}, error) {
	// Check if it's a function in the environment
	if ctx.env != nil {
		if fn, ok := ctx.env.functions[name]; ok {
			// Special case for parent() function which needs access to the RenderContext
			if name == "parent" {
				return fn(args...)
			}

			// Regular function call
			return fn(args...)
		}
	}

	// Check if it's a built-in function
	switch name {
	case "range":
		return ctx.callRangeFunction(args)
	case "length", "count":
		return ctx.callLengthFunction(args)
	case "max":
		return ctx.callMaxFunction(args)
	case "min":
		return ctx.callMinFunction(args)
	}

	// Check if it's a macro
	if macro, ok := ctx.GetMacro(name); ok {
		// Return a callable function
		return func(w io.Writer) error {
			macroNode, ok := macro.(*MacroNode)
			if !ok {
				return fmt.Errorf("'%s' is not a macro", name)
			}
			return macroNode.CallMacro(w, ctx, args...)
		}, nil
	}

	return nil, fmt.Errorf("function '%s' not found", name)
}

// callRangeFunction implements the range function
func (ctx *RenderContext) callRangeFunction(args []interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("range function requires at least 2 arguments")
	}

	// Get the start and end values
	start, ok1 := ctx.toNumber(args[0])
	end, ok2 := ctx.toNumber(args[1])

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("range arguments must be numbers")
	}

	// Get the step value (default is 1)
	step := 1.0
	if len(args) > 2 {
		if s, ok := ctx.toNumber(args[2]); ok {
			step = s
		}
	}

	// Create the range
	result := make([]interface{}, 0)

	if step > 0 {
		for i := start; i <= end; i += step {
			result = append(result, int(i))
		}
	} else {
		for i := start; i >= end; i += step {
			result = append(result, int(i))
		}
	}

	// Always return a non-nil slice for the for loop
	if len(result) == 0 {
		return []interface{}{}, nil
	}

	return result, nil
}

// callLengthFunction implements the length/count function
func (ctx *RenderContext) callLengthFunction(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("length/count function requires exactly 1 argument")
	}

	val := args[0]
	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.String:
		return len(v.String()), nil
	case reflect.Slice, reflect.Array:
		return v.Len(), nil
	case reflect.Map:
		return v.Len(), nil
	default:
		return 0, nil
	}
}

// callMaxFunction implements the max function
func (ctx *RenderContext) callMaxFunction(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("max function requires at least 1 argument")
	}

	// If the argument is a slice or array, find the max value in it
	if len(args) == 1 {
		v := reflect.ValueOf(args[0])
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			if v.Len() == 0 {
				return nil, nil
			}

			max := v.Index(0).Interface()
			maxNum, ok := ctx.toNumber(max)
			if !ok {
				return max, nil
			}

			for i := 1; i < v.Len(); i++ {
				val := v.Index(i).Interface()
				if valNum, ok := ctx.toNumber(val); ok {
					if valNum > maxNum {
						max = val
						maxNum = valNum
					}
				}
			}

			return max, nil
		}
	}

	// Find the max value in the arguments
	max := args[0]
	maxNum, ok := ctx.toNumber(max)
	if !ok {
		return max, nil
	}

	for i := 1; i < len(args); i++ {
		val := args[i]
		if valNum, ok := ctx.toNumber(val); ok {
			if valNum > maxNum {
				max = val
				maxNum = valNum
			}
		}
	}

	return max, nil
}

// callMinFunction implements the min function
func (ctx *RenderContext) callMinFunction(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("min function requires at least 1 argument")
	}

	// If the argument is a slice or array, find the min value in it
	if len(args) == 1 {
		v := reflect.ValueOf(args[0])
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			if v.Len() == 0 {
				return nil, nil
			}

			min := v.Index(0).Interface()
			minNum, ok := ctx.toNumber(min)
			if !ok {
				return min, nil
			}

			for i := 1; i < v.Len(); i++ {
				val := v.Index(i).Interface()
				if valNum, ok := ctx.toNumber(val); ok {
					if valNum < minNum {
						min = val
						minNum = valNum
					}
				}
			}

			return min, nil
		}
	}

	// Find the min value in the arguments
	min := args[0]
	minNum, ok := ctx.toNumber(min)
	if !ok {
		return min, nil
	}

	for i := 1; i < len(args); i++ {
		val := args[i]
		if valNum, ok := ctx.toNumber(val); ok {
			if valNum < minNum {
				min = val
				minNum = valNum
			}
		}
	}

	return min, nil
}

// EvaluateExpression evaluates an expression node
func (ctx *RenderContext) EvaluateExpression(node Node) (interface{}, error) {
	if node == nil {
		return nil, nil
	}

	// Check sandbox security if enabled
	if ctx.sandboxed && ctx.env.securityPolicy != nil {
		switch n := node.(type) {
		case *FunctionNode:
			if !ctx.env.securityPolicy.IsFunctionAllowed(n.name) {
				return nil, NewFunctionViolation(n.name)
			}
		case *FilterNode:
			if !ctx.env.securityPolicy.IsFilterAllowed(n.filter) {
				return nil, NewFilterViolation(n.filter)
			}
		}
	}

	switch n := node.(type) {
	case *LiteralNode:
		return n.value, nil

	case *VariableNode:
		// Check if it's a macro first
		if macro, ok := ctx.GetMacro(n.name); ok {
			return macro, nil
		}

		// Otherwise, look up variable
		return ctx.GetVariable(n.name)

	case *GetAttrNode:
		obj, err := ctx.EvaluateExpression(n.node)
		if err != nil {
			return nil, err
		}

		attrName, err := ctx.EvaluateExpression(n.attribute)
		if err != nil {
			return nil, err
		}

		attrStr, ok := attrName.(string)
		if !ok {
			return nil, fmt.Errorf("attribute name must be a string")
		}

		// Check if obj is a map containing macros (from import)
		if moduleMap, ok := obj.(map[string]interface{}); ok {
			if macro, ok := moduleMap[attrStr]; ok {
				return macro, nil
			}
		}

		return ctx.getAttribute(obj, attrStr)

	case *GetItemNode:
		// Evaluate the container (array, slice, map)
		container, err := ctx.EvaluateExpression(n.node)
		if err != nil {
			return nil, err
		}

		// Evaluate the item index/key
		index, err := ctx.EvaluateExpression(n.item)
		if err != nil {
			return nil, err
		}

		return ctx.getItem(container, index)

	case *BinaryNode:
		// First, evaluate the left side of the expression
		left, err := ctx.EvaluateExpression(n.left)
		if err != nil {
			return nil, err
		}

		// Implement short-circuit evaluation for logical operators
		if n.operator == "and" || n.operator == "&&" {
			// For "and" operator, if left side is false, return false without evaluating right side
			if !ctx.toBool(left) {
				return false, nil
			}
		} else if n.operator == "or" || n.operator == "||" {
			// For "or" operator, if left side is true, return true without evaluating right side
			if ctx.toBool(left) {
				return true, nil
			}
		}

		// For other operators or if short-circuit condition not met, evaluate right side
		right, err := ctx.EvaluateExpression(n.right)
		if err != nil {
			return nil, err
		}

		return ctx.evaluateBinaryOp(n.operator, left, right)

	case *ConditionalNode:
		// Evaluate the condition
		condResult, err := ctx.EvaluateExpression(n.condition)
		if err != nil {
			// Log error if debug is enabled
			if IsDebugEnabled() {
				LogError(err, "Error evaluating 'if' condition")
			}
			return nil, err
		}

		// Log result if debug is enabled
		conditionResult := ctx.toBool(condResult)
		if IsDebugEnabled() {
			LogDebug("Ternary condition result: %v (type: %T, raw value: %v)", conditionResult, condResult, condResult)
			LogDebug("Branches: true=%T, false=%T", n.trueExpr, n.falseExpr)
		}

		// If condition is true, evaluate the true expression, otherwise evaluate the false expression
		if ctx.toBool(condResult) {
			return ctx.EvaluateExpression(n.trueExpr)
		} else {
			return ctx.EvaluateExpression(n.falseExpr)
		}

	case *ArrayNode:
		// We need to avoid pooling for arrays that might be directly used by filters like merge
		// as those filters return the slice directly to the user
		items := make([]interface{}, 0, len(n.items))
		
		for i := 0; i < len(n.items); i++ {
			val, err := ctx.EvaluateExpression(n.items[i])
			if err != nil {
				return nil, err
			}
			items = append(items, val)
		}

		// Always return a non-nil slice, even if empty
		if len(items) == 0 {
			return []interface{}{}, nil
		}

		return items, nil

	case *HashNode:
		// Evaluate each key-value pair in the hash using a new map
		// We can't use pooling with defer here because the map is returned directly
		result := make(map[string]interface{}, len(n.items))
		
		for k, v := range n.items {
			// Evaluate the key
			keyVal, err := ctx.EvaluateExpression(k)
			if err != nil {
				return nil, err
			}

			// Convert key to string
			key := ctx.ToString(keyVal)

			// Evaluate the value
			val, err := ctx.EvaluateExpression(v)
			if err != nil {
				return nil, err
			}

			// Store in the map
			result[key] = val
		}
		return result, nil

	case *FunctionNode:
		// Check if this is a module.function() call (moduleExpr will be non-nil)
		if n.moduleExpr != nil {
			if IsDebugEnabled() && debugger.level >= DebugVerbose {
				LogVerbose("Handling module.function() call with module expression")
			}

			// Evaluate the module expression first
			moduleObj, err := ctx.EvaluateExpression(n.moduleExpr)
			if err != nil {
				return nil, err
			}

			// Evaluate all arguments - need direct allocation
			args := make([]interface{}, len(n.args))
			
			for i := 0; i < len(n.args); i++ {
				val, err := ctx.EvaluateExpression(n.args[i])
				if err != nil {
					return nil, err
				}
				args[i] = val
			}

			// Check if moduleObj is a map that contains macros
			if moduleMap, ok := moduleObj.(map[string]interface{}); ok {
				if macroObj, ok := moduleMap[n.name]; ok {
					if IsDebugEnabled() && debugger.level >= DebugVerbose {
						LogVerbose("Found macro '%s' in module map", n.name)
					}

					// If the macro is a MacroNode, return a callable to render it
					if macroNode, ok := macroObj.(*MacroNode); ok {
						// Return a callable that can be rendered later
						return func(w io.Writer) error {
							return macroNode.CallMacro(w, ctx, args...)
						}, nil
					}
				}
			}

			// Fallback - try calling it like a regular function
			if IsDebugEnabled() && debugger.level >= DebugVerbose {
				LogVerbose("Fallback - calling '%s' as a regular function", n.name)
			}
			result, err := ctx.CallFunction(n.name, args)
			if err != nil {
				return nil, err
			}

			return result, nil
		}

		// Check if it's a macro call
		if macro, ok := ctx.GetMacro(n.name); ok {
			// Evaluate arguments - need direct allocation for macro calls
			args := make([]interface{}, len(n.args))
			
			// Evaluate arguments
			for i := 0; i < len(n.args); i++ {
				val, err := ctx.EvaluateExpression(n.args[i])
				if err != nil {
					return nil, err
				}
				args[i] = val
			}

			// Return a callable that can be rendered later
			return func(w io.Writer) error {
				macroNode, ok := macro.(*MacroNode)
				if !ok {
					return fmt.Errorf("'%s' is not a macro", n.name)
				}
				return macroNode.CallMacro(w, ctx, args...)
			}, nil
		}

		// Otherwise, it's a regular function call
		// Evaluate arguments - need direct allocation for function calls
		args := make([]interface{}, len(n.args))
		
		// Evaluate arguments
		for i := 0; i < len(n.args); i++ {
			val, err := ctx.EvaluateExpression(n.args[i])
			if err != nil {
				return nil, err
			}
			args[i] = val
		}

		result, err := ctx.CallFunction(n.name, args)
		if err != nil {
			return nil, err
		}

		// Make sure function results that should be iterable actually are
		if result == nil && (n.name == "range" || n.name == "length") {
			return []interface{}{}, nil
		}

		return result, nil

	case *FilterNode:
		// Use the optimized filter chain implementation from render_filter.go
		result, err := ctx.evaluateFilterNode(n)
		if err != nil {
			return nil, err
		}

		// Ensure filter results are never nil if they're expected to be iterable
		if result == nil {
			return "", nil
		}

		return result, nil

	case *TestNode:
		// Handle special "not defined" test (from parseBinaryExpression)
		if n.test == "not defined" {
			// Check if it's a variable reference
			if varNode, ok := n.node.(*VariableNode); ok {
				// Check directly in context
				if ctx.context != nil {
					_, exists := ctx.context[varNode.name]
					if exists {
						// If it exists, "not defined" is false
						return false, nil
					}
				}

				// Try full variable lookup
				val, err := ctx.GetVariable(varNode.name)
				// Return true if not defined (err != nil or val is nil)
				return err != nil || val == nil, nil
			}

			// For non-variable nodes, assume defined
			return false, nil
		}

		// Special handling for "is defined" test with attribute access
		if n.test == "defined" {
			// Check if this is a GetAttrNode
			if getAttrNode, ok := n.node.(*GetAttrNode); ok {
				// Evaluate the object
				obj, err := ctx.EvaluateExpression(getAttrNode.node)
				if err != nil {
					return false, nil // If can't evaluate the object, it's not defined
				}

				// If obj is nil, attribute not defined
				if obj == nil {
					return false, nil
				}

				// Evaluate the attribute name
				attrNameNode, err := ctx.EvaluateExpression(getAttrNode.attribute)
				if err != nil {
					return false, nil
				}

				attrName, ok := attrNameNode.(string)
				if !ok {
					return false, nil
				}

				// For maps, directly check if the key exists
				if objMap, ok := obj.(map[string]interface{}); ok {
					_, exists := objMap[attrName]
					return exists, nil
				}

				// For other types, try to get the attribute but catch the error
				_, err = ctx.getAttribute(obj, attrName)
				return err == nil, nil
			}

			// Check for simple variable references
			if varNode, ok := n.node.(*VariableNode); ok {
				// Check directly in context
				if ctx.context != nil {
					_, exists := ctx.context[varNode.name]
					if exists {
						return true, nil
					}
				}

				// Try full variable lookup
				val, err := ctx.GetVariable(varNode.name)
				return err == nil && val != nil, nil
			}
		}

		// Standard test evaluation for all other cases
		// Evaluate the tested value
		value, err := ctx.EvaluateExpression(n.node)
		if err != nil {
			return nil, err
		}

		// Evaluate test arguments - need direct allocation
		args := make([]interface{}, len(n.args))
		
		// Evaluate arguments
		for i := 0; i < len(n.args); i++ {
			val, err := ctx.EvaluateExpression(n.args[i])
			if err != nil {
				return nil, err
			}
			args[i] = val
		}

		// Look for the test in the environment
		if ctx.env != nil {
			if test, ok := ctx.env.tests[n.test]; ok {
				// Call the test with the value and any arguments
				return test(value, args...)
			}
		}

		return false, fmt.Errorf("test '%s' not found", n.test)

	case *UnaryNode:
		// Evaluate the operand
		operand, err := ctx.EvaluateExpression(n.node)
		if err != nil {
			return nil, err
		}

		// Apply the operator
		switch n.operator {
		case "not", "!":
			// Ensuring that the boolean conversion is correct before negation
			result := ctx.toBool(operand)
			return !result, nil
		case "+":
			if num, ok := ctx.toNumber(operand); ok {
				return num, nil
			}
			return 0, nil
		case "-":
			if num, ok := ctx.toNumber(operand); ok {
				return -num, nil
			}
			return 0, nil
		default:
			return nil, fmt.Errorf("unsupported unary operator: %s", n.operator)
		}

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", node)
	}
}

// attributeCacheKey is used as a key for the attribute cache
type attributeCacheKey struct {
	typ  reflect.Type
	attr string
}

// attributeCacheEntry represents a cached attribute lookup result
type attributeCacheEntry struct {
	fieldIndex  int       // Index of the field (-1 if not a field)
	isMethod    bool      // Whether this is a method
	methodIndex int       // Index of the method (-1 if not a method)
	ptrMethod   bool      // Whether the method is on the pointer type
	lastAccess  time.Time // When this entry was last accessed
	accessCount int       // How many times this entry has been accessed
}

// attributeCache caches attribute lookups by type and attribute name
// Uses a simplified LRU strategy for eviction - when cache fills up,
// we remove 10% of the least recently used entries to make room
var attributeCache = struct {
	sync.RWMutex
	m           map[attributeCacheKey]attributeCacheEntry
	maxSize     int     // Maximum number of entries to cache
	currSize    int     // Current number of entries
	evictionPct float64 // Percentage of cache to evict when full (0.0-1.0)
}{
	m:           make(map[attributeCacheKey]attributeCacheEntry),
	maxSize:     1000, // Limit cache to 1000 entries to prevent unbounded growth
	evictionPct: 0.1,  // Evict 10% of entries when cache is full
}

// evictLRUEntries removes the least recently used entries from the cache
// This function assumes that the caller holds the attributeCache lock
func evictLRUEntries() {
	// Calculate how many entries to evict
	numToEvict := int(float64(attributeCache.maxSize) * attributeCache.evictionPct)
	if numToEvict < 1 {
		numToEvict = 1 // Always evict at least one entry
	}

	// Create a slice of entries to sort by last access time
	type cacheItem struct {
		key   attributeCacheKey
		entry attributeCacheEntry
	}

	entries := make([]cacheItem, 0, attributeCache.currSize)
	for k, v := range attributeCache.m {
		entries = append(entries, cacheItem{k, v})
	}

	// Sort entries by last access time (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		// If access counts differ by a significant amount, prefer keeping frequently accessed items
		if entries[i].entry.accessCount < entries[j].entry.accessCount/2 {
			return true
		}
		// Otherwise, use recency as the deciding factor
		return entries[i].entry.lastAccess.Before(entries[j].entry.lastAccess)
	})

	// Remove the oldest entries
	for i := 0; i < numToEvict && i < len(entries); i++ {
		delete(attributeCache.m, entries[i].key)
		attributeCache.currSize--
	}
}

// getItem gets an item from a container (array, slice, map) by index or key
func (ctx *RenderContext) getItem(container, index interface{}) (interface{}, error) {
	if container == nil {
		return nil, nil
	}

	// Convert numeric indices to int for consistent handling
	idx, _ := ctx.toNumber(index)
	intIndex := int(idx)

	// Handle different container types
	switch c := container.(type) {
	case []interface{}:
		// Check bounds
		if intIndex < 0 || intIndex >= len(c) {
			return nil, fmt.Errorf("array index out of bounds: %d", intIndex)
		}
		return c[intIndex], nil

	case map[string]interface{}:
		// Try string key
		if strKey, ok := index.(string); ok {
			if value, exists := c[strKey]; exists {
				return value, nil
			}
		}

		// Try numeric key as string
		strKey := ctx.ToString(index)
		if value, exists := c[strKey]; exists {
			return value, nil
		}

		return nil, nil // Nil for missing keys

	default:
		// Use reflection for other types
		v := reflect.ValueOf(container)

		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			// Check bounds
			if intIndex < 0 || intIndex >= v.Len() {
				return nil, fmt.Errorf("array index out of bounds: %d", intIndex)
			}
			return v.Index(intIndex).Interface(), nil

		case reflect.Map:
			// Try to find the key
			var mapKey reflect.Value

			// Convert the index to the map's key type if possible
			keyType := v.Type().Key()
			indexValue := reflect.ValueOf(index)

			if indexValue.Type().ConvertibleTo(keyType) {
				mapKey = indexValue.Convert(keyType)
			} else {
				// Try string conversion for the key
				strKey := ctx.ToString(index)
				if reflect.TypeOf(strKey).ConvertibleTo(keyType) {
					mapKey = reflect.ValueOf(strKey).Convert(keyType)
				} else {
					return nil, nil // Key type mismatch
				}
			}

			mapValue := v.MapIndex(mapKey)
			if mapValue.IsValid() {
				return mapValue.Interface(), nil
			}
		}
	}

	return nil, nil // Default nil for non-indexable types
}

// getAttribute gets an attribute from an object
func (ctx *RenderContext) getAttribute(obj interface{}, attr string) (interface{}, error) {
	if obj == nil {
		// Instead of returning an error for nil objects, return nil value
		return nil, nil
	}

	// Fast path for maps
	if objMap, ok := obj.(map[string]interface{}); ok {
		if value, exists := objMap[attr]; exists {
			return value, nil
		}

		// For non-existent keys in maps, return nil instead of an error
		return nil, nil
	}

	// Get the reflect.Value and type for the object
	objValue := reflect.ValueOf(obj)
	origType := objValue.Type()

	// Handle pointer indirection
	isPtr := origType.Kind() == reflect.Ptr
	if isPtr {
		objValue = objValue.Elem()
	}

	// Only use caching for struct types
	if objValue.Kind() != reflect.Struct {
		// Instead of returning an error for non-struct types, return nil
		return nil, nil
	}

	objType := objValue.Type()

	// Create a cache key
	key := attributeCacheKey{
		typ:  objType,
		attr: attr,
	}

	// Get a read lock to check the cache first
	attributeCache.RLock()
	entry, found := attributeCache.m[key]
	if found {
		// Found in cache, update access stats later with a write lock
		attributeCache.RUnlock()

		// Update the entry's access statistics with a write lock
		attributeCache.Lock()
		// Need to check again after acquiring write lock
		if cachedEntry, stillExists := attributeCache.m[key]; stillExists {
			// Update access time and count
			cachedEntry.lastAccess = time.Now()
			cachedEntry.accessCount++
			attributeCache.m[key] = cachedEntry
			entry = cachedEntry
		}
		attributeCache.Unlock()
	} else {
		// Not found in cache - release read lock and get write lock for update
		attributeCache.RUnlock()
		attributeCache.Lock()

		// Double-check if another goroutine added it while we were waiting
		entry, found = attributeCache.m[key]
		if !found {
			// Still not found, need to populate the cache

			// Check if cache has reached maximum size
			if attributeCache.currSize >= attributeCache.maxSize {
				// Cache is full, use our LRU eviction strategy
				evictLRUEntries()
			}

			// Create a new entry with current timestamp
			entry = attributeCacheEntry{
				fieldIndex:  -1,
				methodIndex: -1,
				lastAccess:  time.Now(),
				accessCount: 1,
			}

			// Look for a field
			field, found := objType.FieldByName(attr)
			if found {
				entry.fieldIndex = field.Index[0] // Assuming single-level field access
			}

			// Look for a method on the value
			method, found := objType.MethodByName(attr)
			if found && method.Type.NumIn() == 1 { // The receiver is the first argument
				entry.isMethod = true
				entry.methodIndex = method.Index
			} else {
				// Look for a method on the pointer to the value
				ptrType := reflect.PtrTo(objType)
				method, found := ptrType.MethodByName(attr)
				if found && method.Type.NumIn() == 1 {
					entry.isMethod = true
					entry.ptrMethod = true
					entry.methodIndex = method.Index
				}
			}

			// Store in cache
			attributeCache.m[key] = entry
			attributeCache.currSize++
		}
		attributeCache.Unlock()
	}

	// Use the cached lookup information to get the attribute

	// Try field access first
	if entry.fieldIndex >= 0 {
		field := objValue.Field(entry.fieldIndex)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), nil
		}
	}

	// Try method access
	if entry.isMethod && entry.methodIndex >= 0 {
		var method reflect.Value

		if entry.ptrMethod {
			// Need a pointer to the struct
			if isPtr {
				// Object is already a pointer, use the original value
				method = reflect.ValueOf(obj).Method(entry.methodIndex)
			} else {
				// Create a new pointer to the struct
				ptrValue := reflect.New(objType)
				ptrValue.Elem().Set(objValue)
				method = ptrValue.Method(entry.methodIndex)
			}
		} else {
			// Method is directly on the struct type
			method = objValue.Method(entry.methodIndex)
		}

		if method.IsValid() {
			results := method.Call(nil)
			if len(results) > 0 {
				return results[0].Interface(), nil
			}
			return nil, nil
		}
	}

	// Instead of returning an error for attributes not found, just return nil
	return nil, nil
}

// evaluateBinaryOp evaluates a binary operation
func (ctx *RenderContext) evaluateBinaryOp(operator string, left, right interface{}) (interface{}, error) {
	// Check for the special case: 'not defined' test
	if operator == "not" && (right == "defined" || right == string("defined")) {
		// Special case for variable node check
		if varNode, ok := left.(*VariableNode); ok {
			// Get the variable name from the node
			varName := varNode.name
			// Check if the variable exists in the context
			_, err := ctx.GetVariable(varName)
			// Variable not defined = return true, otherwise false
			return err != nil, nil
		}
		// Default to false for other cases
		return false, nil
	}

	// For standard 'not' operator
	if operator == "not" {
		// Handle regular boolean negation
		return !ctx.toBool(right), nil
	}

	switch operator {
	case "+":
		// Check if both values can be interpreted as numbers first for proper type handling
		lNum, lok := ctx.toNumber(left)
		rNum, rok := ctx.toNumber(right)

		if lok && rok {
			// If both can be numbers, perform numeric addition
			return lNum + rNum, nil
		}

		// Otherwise, handle string concatenation
		if lStr, lok := left.(string); lok {
			if rStr, rok := right.(string); rok {
				return lStr + rStr, nil
			}
			return lStr + ctx.ToString(right), nil
		}

	case "-":
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				return lNum - rNum, nil
			}
		}

	case "*":
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				return lNum * rNum, nil
			}
		}

	case "/":
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				if rNum == 0 {
					return nil, errors.New("division by zero")
				}
				return lNum / rNum, nil
			}
		}

	case "%":
		// Modulo operator
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				if rNum == 0 {
					return nil, errors.New("modulo by zero")
				}
				return math.Mod(lNum, rNum), nil
			}
		}

	case "^":
		// Exponentiation operator
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				return math.Pow(lNum, rNum), nil
			}
		}

	case "==":
		return ctx.equals(left, right), nil

	case "!=":
		return !ctx.equals(left, right), nil

	case "<":
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				return lNum < rNum, nil
			}
		}

	case ">":
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				return lNum > rNum, nil
			}
		}

	case "<=":
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				return lNum <= rNum, nil
			}
		}

	case ">=":
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				return lNum >= rNum, nil
			}
		}

	case "and", "&&":
		// Note: Short-circuit evaluation is already handled in EvaluateExpression
		// This is just the final boolean combination
		return ctx.toBool(left) && ctx.toBool(right), nil

	case "or", "||":
		// Note: Short-circuit evaluation is already handled in EvaluateExpression
		// This is just the final boolean combination
		return ctx.toBool(left) || ctx.toBool(right), nil

	case "~":
		// String concatenation
		return ctx.ToString(left) + ctx.ToString(right), nil

	case "in":
		// Check if left is in right (for arrays, slices, maps, strings)
		return ctx.contains(right, left)

	case "not in":
		// Check if left is not in right
		contains, err := ctx.contains(right, left)
		if err != nil {
			return false, err
		}
		return !contains, nil

	case "matches":
		// Regular expression match
		pattern := ctx.ToString(right)
		str := ctx.ToString(left)

		// Check for flags in the pattern
		caseInsensitive := false
		if len(pattern) >= 3 && pattern[0] == '/' {
			// Check for /pattern/i format (i after the slash)
			if len(pattern) >= 4 && pattern[len(pattern)-1] == 'i' && pattern[len(pattern)-2] == '/' {
				caseInsensitive = true
				pattern = pattern[1 : len(pattern)-2]
			} else if pattern[len(pattern)-1] == '/' {
				// Check for /pattern/i format (i before the slash)
				if len(pattern) >= 4 && pattern[len(pattern)-2] == 'i' {
					caseInsensitive = true
					pattern = pattern[1 : len(pattern)-2]
				} else {
					// Regular /pattern/ without flags
					pattern = pattern[1 : len(pattern)-1]
				}
			}
		}

		// Handle escaped character sequences
		pattern = strings.ReplaceAll(pattern, "\\\\", "\\")

		// Special handling for regex character classes
		// When working with backslashes in strings, we need 2 levels of escaping
		// 1. In Go source, \d is written as \\d
		// 2. After string processing, we need to handle it specially
		pattern = strings.ReplaceAll(pattern, "\\d", "[0-9]")        // digits
		pattern = strings.ReplaceAll(pattern, "\\w", "[a-zA-Z0-9_]") // word chars
		pattern = strings.ReplaceAll(pattern, "\\s", "[ \\t\\n\\r]") // whitespace

		// Compile the regex with appropriate flags
		var regex *regexp.Regexp
		var err error
		if caseInsensitive {
			regex, err = regexp.Compile("(?i)" + pattern)
		} else {
			regex, err = regexp.Compile(pattern)
		}

		if err != nil {
			return false, fmt.Errorf("invalid regular expression: %s", err)
		}

		result := regex.MatchString(str)
		// Add debug logging for the regex matches
		if IsDebugEnabled() {
			LogDebug("Regex match: pattern=%q, text=%q, result=%v", pattern, str, result)
		}
		return result, nil

	case "starts with":
		// String prefix check
		str := ctx.ToString(left)
		prefix := ctx.ToString(right)
		return strings.HasPrefix(str, prefix), nil

	case "ends with":
		// String suffix check
		str := ctx.ToString(left)
		suffix := ctx.ToString(right)
		return strings.HasSuffix(str, suffix), nil
	}

	return nil, fmt.Errorf("unsupported binary operator: %s", operator)
}

// contains checks if a value is contained in a container (string, slice, array, map)
func (ctx *RenderContext) contains(container, item interface{}) (bool, error) {
	if container == nil {
		return false, nil
	}

	// Fast path for common types
	switch c := container.(type) {
	case string:
		// Use string conversion only once
		return strings.Contains(c, ctx.ToString(item)), nil
	case []interface{}:
		// For small slices, linear search is fine
		// For larger slices (>50 items), consider a map-based approach
		if len(c) > 50 {
			// Create a temporary map for O(1) lookups
			// Only worth doing for sufficiently large slices
			tempMap := make(map[interface{}]struct{}, len(c))
			for _, v := range c {
				tempMap[v] = struct{}{}
			}

			// For numeric items, try direct lookup first
			if _, ok := tempMap[item]; ok {
				return true, nil
			}

			// For string-comparable items, try string version
			if _, ok := tempMap[ctx.ToString(item)]; ok {
				return true, nil
			}

			// Fall back to deep equality comparison
			for k := range tempMap {
				if ctx.equals(k, item) {
					return true, nil
				}
			}
			return false, nil
		}

		// For small slices, linear search
		for _, v := range c {
			if ctx.equals(v, item) {
				return true, nil
			}
		}
		return false, nil
	case map[string]interface{}:
		// Convert only once for map key lookup
		itemStr := ctx.ToString(item)
		_, exists := c[itemStr]
		return exists, nil
	}

	// Handle other types via reflection
	rv := reflect.ValueOf(container)
	switch rv.Kind() {
	case reflect.String:
		return strings.Contains(rv.String(), ctx.ToString(item)), nil
	case reflect.Array, reflect.Slice:
		// Optimize for large slices/arrays
		if rv.Len() > 50 {
			// Same map-based optimization as above
			tempMap := make(map[interface{}]struct{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				tempMap[rv.Index(i).Interface()] = struct{}{}
			}

			// Try direct lookup
			if _, ok := tempMap[item]; ok {
				return true, nil
			}

			// Try string-based lookup
			if _, ok := tempMap[ctx.ToString(item)]; ok {
				return true, nil
			}

			// Fall back to equality comparison
			for k := range tempMap {
				if ctx.equals(k, item) {
					return true, nil
				}
			}
			return false, nil
		}

		// For small collections, linear search
		for i := 0; i < rv.Len(); i++ {
			if ctx.equals(rv.Index(i).Interface(), item) {
				return true, nil
			}
		}
	case reflect.Map:
		// For maps, we're already doing key lookup which is O(1)
		for _, key := range rv.MapKeys() {
			if ctx.equals(key.Interface(), item) {
				return true, nil
			}
		}
	}

	return false, nil
}

// equals checks if two values are equal
func (ctx *RenderContext) equals(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Try numeric comparison
	if aNum, aok := ctx.toNumber(a); aok {
		if bNum, bok := ctx.toNumber(b); bok {
			return aNum == bNum
		}
	}

	// Try string comparison
	return ctx.ToString(a) == ctx.ToString(b)
}

// toNumber converts a value to a float64, returning ok=false if not possible
func (ctx *RenderContext) toNumber(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		// Try to parse as float64
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
		return 0, false
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	}

	// Try reflection for custom types
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	}

	return 0, false
}

// toBool converts a value to a boolean
func (ctx *RenderContext) toBool(val interface{}) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case int, int8, int16, int32, int64:
		return v != 0
	case uint, uint8, uint16, uint32, uint64:
		return v != 0
	case float32, float64:
		return v != 0
	case string:
		return v != ""
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	}

	// Try reflection for other types
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	case reflect.String:
		return rv.String() != ""
	case reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() > 0
	}

	// Default to true for other non-nil values
	return true
}

// ToString converts a value to a string
func (ctx *RenderContext) ToString(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	}

	return fmt.Sprintf("%v", val)
}
