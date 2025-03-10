package twig

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// RenderContext holds the state during template rendering
type RenderContext struct {
	env          *Environment
	context      map[string]interface{}
	blocks       map[string][]Node
	macros       map[string]Node
	parent       *RenderContext
	engine       *Engine         // Reference to engine for loading templates
	extending    bool            // Whether this template extends another
	currentBlock *BlockNode      // Current block being rendered (for parent() function)
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

// SetVariable sets a variable in the context
func (ctx *RenderContext) SetVariable(name string, value interface{}) {
	ctx.context[name] = value
}

// GetMacro gets a macro from the context
func (ctx *RenderContext) GetMacro(name string) (Node, bool) {
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
	return macroNode.Call(w, ctx, args)
}

// CallFunction calls a function with the given arguments
func (ctx *RenderContext) CallFunction(name string, args []interface{}) (interface{}, error) {
	// Check if it's a function in the environment
	if ctx.env != nil {
		if fn, ok := ctx.env.functions[name]; ok {
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
			return macroNode.Call(w, ctx, args)
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
	var result []int
	for i := start; i <= end; i += step {
		result = append(result, int(i))
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
		
	case *BinaryNode:
		left, err := ctx.EvaluateExpression(n.left)
		if err != nil {
			return nil, err
		}
		
		right, err := ctx.EvaluateExpression(n.right)
		if err != nil {
			return nil, err
		}
		
		return ctx.evaluateBinaryOp(n.operator, left, right)
		
	case *ConditionalNode:
		// Evaluate the condition
		condResult, err := ctx.EvaluateExpression(n.condition)
		if err != nil {
			return nil, err
		}
		
		// If condition is true, evaluate the true expression, otherwise evaluate the false expression
		if ctx.toBool(condResult) {
			return ctx.EvaluateExpression(n.trueExpr)
		} else {
			return ctx.EvaluateExpression(n.falseExpr)
		}
		
	case *ArrayNode:
		// Evaluate each item in the array
		items := make([]interface{}, len(n.items))
		for i, item := range n.items {
			val, err := ctx.EvaluateExpression(item)
			if err != nil {
				return nil, err
			}
			items[i] = val
		}
		return items, nil
		
	case *FunctionNode:
		// Check if it's a macro call
		if macro, ok := ctx.GetMacro(n.name); ok {
			// Evaluate arguments
			args := make([]interface{}, len(n.args))
			for i, arg := range n.args {
				val, err := ctx.EvaluateExpression(arg)
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
				return macroNode.Call(w, ctx, args)
			}, nil
		}
		
		// Otherwise, it's a regular function call
		// Evaluate arguments
		args := make([]interface{}, len(n.args))
		for i, arg := range n.args {
			val, err := ctx.EvaluateExpression(arg)
			if err != nil {
				return nil, err
			}
			args[i] = val
		}
		
		return ctx.CallFunction(n.name, args)
		
	case *FilterNode:
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
		
		// Look for the filter in the environment
		if ctx.env != nil {
			if filter, ok := ctx.env.filters[n.filter]; ok {
				// Call the filter with the value and any arguments
				return filter(value, args...)
			}
		}
		
		return nil, fmt.Errorf("filter '%s' not found", n.filter)
		
	case *TestNode:
		// Evaluate the tested value
		value, err := ctx.EvaluateExpression(n.node)
		if err != nil {
			return nil, err
		}
		
		// Evaluate test arguments
		args := make([]interface{}, len(n.args))
		for i, arg := range n.args {
			val, err := ctx.EvaluateExpression(arg)
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
			return !ctx.toBool(operand), nil
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

// getAttribute gets an attribute from an object
func (ctx *RenderContext) getAttribute(obj interface{}, attr string) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("%w: cannot get attribute %s of nil", ErrInvalidAttribute, attr)
	}
	
	// Handle maps
	if objMap, ok := obj.(map[string]interface{}); ok {
		if value, exists := objMap[attr]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("%w: map has no key %s", ErrInvalidAttribute, attr)
	}
	
	// Use reflection for structs
	objValue := reflect.ValueOf(obj)
	
	// Handle pointer indirection
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}
	
	// Handle structs
	if objValue.Kind() == reflect.Struct {
		// Try field access first
		field := objValue.FieldByName(attr)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), nil
		}
		
		// Try method access (both with and without parameters)
		method := objValue.MethodByName(attr)
		if method.IsValid() {
			if method.Type().NumIn() == 0 {
				results := method.Call(nil)
				if len(results) > 0 {
					return results[0].Interface(), nil
				}
				return nil, nil
			}
		}
		
		// Try method on pointer to struct
		ptrValue := reflect.New(objValue.Type())
		ptrValue.Elem().Set(objValue)
		method = ptrValue.MethodByName(attr)
		if method.IsValid() {
			if method.Type().NumIn() == 0 {
				results := method.Call(nil)
				if len(results) > 0 {
					return results[0].Interface(), nil
				}
				return nil, nil
			}
		}
	}
	
	return nil, fmt.Errorf("%w: %s", ErrInvalidAttribute, attr)
}

// evaluateBinaryOp evaluates a binary operation
func (ctx *RenderContext) evaluateBinaryOp(operator string, left, right interface{}) (interface{}, error) {
	switch operator {
	case "+":
		// Handle string concatenation
		if lStr, lok := left.(string); lok {
			if rStr, rok := right.(string); rok {
				return lStr + rStr, nil
			}
			return lStr + ctx.ToString(right), nil
		}
		
		// Handle numeric addition
		if lNum, lok := ctx.toNumber(left); lok {
			if rNum, rok := ctx.toNumber(right); rok {
				return lNum + rNum, nil
			}
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
		return ctx.toBool(left) && ctx.toBool(right), nil
		
	case "or", "||":
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
		
		// Compile the regex
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return false, fmt.Errorf("invalid regular expression: %s", err)
		}
		
		return regex.MatchString(str), nil
		
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
	
	itemStr := ctx.ToString(item)
	
	// Handle different container types
	switch c := container.(type) {
	case string:
		return strings.Contains(c, itemStr), nil
	case []interface{}:
		for _, v := range c {
			if ctx.equals(v, item) {
				return true, nil
			}
		}
	case map[string]interface{}:
		for k := range c {
			if k == itemStr {
				return true, nil
			}
		}
	default:
		// Try reflection for other types
		rv := reflect.ValueOf(container)
		switch rv.Kind() {
		case reflect.String:
			return strings.Contains(rv.String(), itemStr), nil
		case reflect.Array, reflect.Slice:
			for i := 0; i < rv.Len(); i++ {
				if ctx.equals(rv.Index(i).Interface(), item) {
					return true, nil
				}
			}
		case reflect.Map:
			for _, key := range rv.MapKeys() {
				if ctx.equals(key.Interface(), item) {
					return true, nil
				}
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