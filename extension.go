package twig

import (
	"errors"
	"fmt"
	"html"
	"math/rand"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FilterFunc is a function that can be used as a filter
type FilterFunc func(value interface{}, args ...interface{}) (interface{}, error)

// FunctionFunc is a function that can be used in templates
type FunctionFunc func(args ...interface{}) (interface{}, error)

// TestFunc is a function that can be used for testing conditions
type TestFunc func(value interface{}, args ...interface{}) (bool, error)

// OperatorFunc is a function that implements a custom operator
type OperatorFunc func(left, right interface{}) (interface{}, error)

// Extension represents a Twig extension
type Extension interface {
	// GetName returns the name of the extension
	GetName() string
	
	// GetFilters returns the filters defined by this extension
	GetFilters() map[string]FilterFunc
	
	// GetFunctions returns the functions defined by this extension
	GetFunctions() map[string]FunctionFunc
	
	// GetTests returns the tests defined by this extension
	GetTests() map[string]TestFunc
	
	// GetOperators returns the operators defined by this extension
	GetOperators() map[string]OperatorFunc
	
	// GetTokenParsers returns any custom token parsers
	GetTokenParsers() []TokenParser
	
	// Initialize initializes the extension
	Initialize(*Engine)
}

// TokenParser provides a way to parse custom tags
type TokenParser interface {
	// GetTag returns the tag this parser handles
	GetTag() string
	
	// Parse parses the tag and returns a node
	Parse(*Parser, *Token) (Node, error)
}

// CoreExtension provides the core Twig functionality
type CoreExtension struct{}

// GetName returns the name of the core extension
func (e *CoreExtension) GetName() string {
	return "core"
}

// GetFilters returns the core filters
func (e *CoreExtension) GetFilters() map[string]FilterFunc {
	return map[string]FilterFunc{
		"default":   e.filterDefault,
		"escape":    e.filterEscape,
		"upper":     e.filterUpper,
		"lower":     e.filterLower,
		"trim":      e.filterTrim,
		"raw":       e.filterRaw,
		"length":    e.filterLength,
		"join":      e.filterJoin,
		"split":     e.filterSplit,
		"date":      e.filterDate,
		"url_encode": e.filterUrlEncode,
	}
}

// GetFunctions returns the core functions
func (e *CoreExtension) GetFunctions() map[string]FunctionFunc {
	return map[string]FunctionFunc{
		"range":    e.functionRange,
		"date":     e.functionDate,
		"random":   e.functionRandom,
		"max":      e.functionMax,
		"min":      e.functionMin,
		"dump":     e.functionDump,
		"constant": e.functionConstant,
	}
}

// GetTests returns the core tests
func (e *CoreExtension) GetTests() map[string]TestFunc {
	return map[string]TestFunc{
		"defined":   e.testDefined,
		"empty":     e.testEmpty,
		"null":      e.testNull,
		"even":      e.testEven,
		"odd":       e.testOdd,
		"iterable":  e.testIterable,
		"same_as":   e.testSameAs,
		"divisible_by": e.testDivisibleBy,
	}
}

// GetOperators returns the core operators
func (e *CoreExtension) GetOperators() map[string]OperatorFunc {
	return map[string]OperatorFunc{
		"in": e.operatorIn,
	}
}

// GetTokenParsers returns the core token parsers
func (e *CoreExtension) GetTokenParsers() []TokenParser {
	return nil
}

// Initialize initializes the core extension
func (e *CoreExtension) Initialize(engine *Engine) {
	// Nothing to initialize for core extension
}

// Filter implementations

func (e *CoreExtension) filterDefault(value interface{}, args ...interface{}) (interface{}, error) {
	if isEmptyValue(value) && len(args) > 0 {
		return args[0], nil
	}
	return value, nil
}

func (e *CoreExtension) filterEscape(value interface{}, args ...interface{}) (interface{}, error) {
	s := toString(value)
	return escapeHTML(s), nil
}

func (e *CoreExtension) filterUpper(value interface{}, args ...interface{}) (interface{}, error) {
	s := toString(value)
	return strings.ToUpper(s), nil
}

func (e *CoreExtension) filterLower(value interface{}, args ...interface{}) (interface{}, error) {
	s := toString(value)
	return strings.ToLower(s), nil
}

func (e *CoreExtension) filterTrim(value interface{}, args ...interface{}) (interface{}, error) {
	s := toString(value)
	return strings.TrimSpace(s), nil
}

func (e *CoreExtension) filterRaw(value interface{}, args ...interface{}) (interface{}, error) {
	// Raw just returns the value without any processing
	return value, nil
}

func (e *CoreExtension) filterLength(value interface{}, args ...interface{}) (interface{}, error) {
	return length(value)
}

func (e *CoreExtension) filterJoin(value interface{}, args ...interface{}) (interface{}, error) {
	delimiter := " "
	if len(args) > 0 {
		if d, ok := args[0].(string); ok {
			delimiter = d
		}
	}
	
	return join(value, delimiter)
}

func (e *CoreExtension) filterSplit(value interface{}, args ...interface{}) (interface{}, error) {
	delimiter := " "
	if len(args) > 0 {
		if d, ok := args[0].(string); ok {
			delimiter = d
		}
	}
	
	s := toString(value)
	return strings.Split(s, delimiter), nil
}

func (e *CoreExtension) filterDate(value interface{}, args ...interface{}) (interface{}, error) {
	// Implement date formatting
	return value, nil
}

func (e *CoreExtension) filterUrlEncode(value interface{}, args ...interface{}) (interface{}, error) {
	s := toString(value)
	return url.QueryEscape(s), nil
}

// Function implementations

func (e *CoreExtension) functionRange(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, errors.New("range function requires at least 2 arguments")
	}
	
	start, err := toInt(args[0])
	if err != nil {
		return nil, err
	}
	
	end, err := toInt(args[1])
	if err != nil {
		return nil, err
	}
	
	step := 1
	if len(args) > 2 {
		s, err := toInt(args[2])
		if err != nil {
			return nil, err
		}
		step = s
	}
	
	if step == 0 {
		return nil, errors.New("step cannot be zero")
	}
	
	var result []int
	if step > 0 {
		for i := start; i <= end; i += step {
			result = append(result, i)
		}
	} else {
		for i := start; i >= end; i += step {
			result = append(result, i)
		}
	}
	
	return result, nil
}

func (e *CoreExtension) functionDate(args ...interface{}) (interface{}, error) {
	// Implement date function
	return time.Now(), nil
}

func (e *CoreExtension) functionRandom(args ...interface{}) (interface{}, error) {
	// Implement random function
	return rand.Intn(100), nil
}

func (e *CoreExtension) functionMax(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("max function requires at least one argument")
	}
	
	var max float64
	var initialized bool
	
	for i, arg := range args {
		num, err := toFloat64(arg)
		if err != nil {
			return nil, fmt.Errorf("argument %d is not a number", i)
		}
		
		if !initialized || num > max {
			max = num
			initialized = true
		}
	}
	
	return max, nil
}

func (e *CoreExtension) functionMin(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("min function requires at least one argument")
	}
	
	var min float64
	var initialized bool
	
	for i, arg := range args {
		num, err := toFloat64(arg)
		if err != nil {
			return nil, fmt.Errorf("argument %d is not a number", i)
		}
		
		if !initialized || num < min {
			min = num
			initialized = true
		}
	}
	
	return min, nil
}

func (e *CoreExtension) functionDump(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	
	var result strings.Builder
	for i, arg := range args {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("%#v", arg))
	}
	
	return result.String(), nil
}

func (e *CoreExtension) functionConstant(args ...interface{}) (interface{}, error) {
	// Not applicable in Go, but included for compatibility
	return nil, errors.New("constant function not supported in Go")
}

// Test implementations

func (e *CoreExtension) testDefined(value interface{}, args ...interface{}) (bool, error) {
	return value != nil, nil
}

func (e *CoreExtension) testEmpty(value interface{}, args ...interface{}) (bool, error) {
	return isEmptyValue(value), nil
}

func (e *CoreExtension) testNull(value interface{}, args ...interface{}) (bool, error) {
	return value == nil, nil
}

func (e *CoreExtension) testEven(value interface{}, args ...interface{}) (bool, error) {
	i, err := toInt(value)
	if err != nil {
		return false, err
	}
	return i%2 == 0, nil
}

func (e *CoreExtension) testOdd(value interface{}, args ...interface{}) (bool, error) {
	i, err := toInt(value)
	if err != nil {
		return false, err
	}
	return i%2 != 0, nil
}

func (e *CoreExtension) testIterable(value interface{}, args ...interface{}) (bool, error) {
	return isIterable(value), nil
}

func (e *CoreExtension) testSameAs(value interface{}, args ...interface{}) (bool, error) {
	if len(args) == 0 {
		return false, errors.New("same_as test requires an argument")
	}
	return value == args[0], nil
}

func (e *CoreExtension) testDivisibleBy(value interface{}, args ...interface{}) (bool, error) {
	if len(args) == 0 {
		return false, errors.New("divisible_by test requires a divisor argument")
	}
	
	dividend, err := toInt(value)
	if err != nil {
		return false, err
	}
	
	divisor, err := toInt(args[0])
	if err != nil {
		return false, err
	}
	
	if divisor == 0 {
		return false, errors.New("division by zero")
	}
	
	return dividend%divisor == 0, nil
}

// Operator implementations

func (e *CoreExtension) operatorIn(left, right interface{}) (interface{}, error) {
	if !isIterable(right) {
		return false, errors.New("right operand must be iterable")
	}
	
	return contains(right, left)
}

// Helper functions

func isEmptyValue(v interface{}) bool {
	if v == nil {
		return true
	}
	
	switch value := v.(type) {
	case string:
		return value == ""
	case bool:
		return !value
	case int, int8, int16, int32, int64:
		return value == 0
	case uint, uint8, uint16, uint32, uint64:
		return value == 0
	case float32, float64:
		return value == 0
	case []interface{}:
		return len(value) == 0
	case map[string]interface{}:
		return len(value) == 0
	}
	
	// Use reflection for other types
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.String:
		return rv.String() == ""
	}
	
	// Default behavior for other types
	return false
}

func isIterable(v interface{}) bool {
	if v == nil {
		return false
	}
	
	switch v.(type) {
	case string, []interface{}, map[string]interface{}:
		return true
	}
	
	// Use reflection for other types
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return true
	}
	
	return false
}

func length(v interface{}) (int, error) {
	if v == nil {
		return 0, nil
	}
	
	switch value := v.(type) {
	case string:
		return len(value), nil
	case []interface{}:
		return len(value), nil
	case map[string]interface{}:
		return len(value), nil
	}
	
	// Use reflection for other types
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return rv.Len(), nil
	}
	
	return 0, fmt.Errorf("cannot get length of %T", v)
}

func join(v interface{}, delimiter string) (string, error) {
	var items []string
	
	if v == nil {
		return "", nil
	}
	
	// Handle different types
	switch value := v.(type) {
	case []string:
		return strings.Join(value, delimiter), nil
	case []interface{}:
		for _, item := range value {
			items = append(items, toString(item))
		}
	default:
		// Try reflection for other types
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < rv.Len(); i++ {
				items = append(items, toString(rv.Index(i).Interface()))
			}
		default:
			return toString(v), nil
		}
	}
	
	return strings.Join(items, delimiter), nil
}

func contains(container, item interface{}) (bool, error) {
	if container == nil {
		return false, nil
	}
	
	itemStr := toString(item)
	
	// Handle different container types
	switch c := container.(type) {
	case string:
		return strings.Contains(c, itemStr), nil
	case []interface{}:
		for _, v := range c {
			if toString(v) == itemStr {
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
				if toString(rv.Index(i).Interface()) == itemStr {
					return true, nil
				}
			}
		case reflect.Map:
			for _, key := range rv.MapKeys() {
				if toString(key.Interface()) == itemStr {
					return true, nil
				}
			}
		}
	}
	
	return false, nil
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case []byte:
		return string(val)
	case fmt.Stringer:
		return val.String()
	}
	
	return fmt.Sprintf("%v", v)
}

func toInt(v interface{}) (int, error) {
	if v == nil {
		return 0, errors.New("cannot convert nil to int")
	}
	
	switch val := v.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		i, err := strconv.Atoi(val)
		if err != nil {
			return 0, err
		}
		return i, nil
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	}
	
	return 0, fmt.Errorf("cannot convert %T to int", v)
}

func toFloat64(v interface{}) (float64, error) {
	if v == nil {
		return 0, errors.New("cannot convert nil to float64")
	}
	
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	}
	
	return 0, fmt.Errorf("cannot convert %T to float64", v)
}

func escapeHTML(s string) string {
	return html.EscapeString(s)
}