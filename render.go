package twig

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// RenderContext holds the state during template rendering
type RenderContext struct {
	env     *Environment
	context map[string]interface{}
	blocks  map[string][]Node
	macros  map[string]Node
	parent  *RenderContext
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
	
	return nil, fmt.Errorf("%w: %s", ErrUndefinedVar, name)
}

// SetVariable sets a variable in the context
func (ctx *RenderContext) SetVariable(name string, value interface{}) {
	ctx.context[name] = value
}

// EvaluateExpression evaluates an expression node
func (ctx *RenderContext) EvaluateExpression(node Node) (interface{}, error) {
	switch n := node.(type) {
	case *LiteralNode:
		return n.value, nil
		
	case *VariableNode:
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
	}
	
	return nil, fmt.Errorf("unsupported binary operator: %s", operator)
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