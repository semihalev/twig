package twig

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ExpressionType represents the type of an expression
type ExpressionType int

// Expression types
const (
	ExprLiteral ExpressionType = iota
	ExprVariable
	ExprUnary
	ExprBinary
	ExprFunction
	ExprFilter
	ExprTest
	ExprGetAttr
	ExprGetItem
	ExprMethodCall
	ExprArray
	ExprHash
	ExprConditional
	ExprModuleMethod
)

// ExpressionNode represents a Twig expression
type ExpressionNode struct {
	exprType ExpressionType
	line     int
}

// LiteralNode represents a literal value (string, number, boolean, null)
type LiteralNode struct {
	ExpressionNode
	value interface{}
}

// VariableNode represents a variable reference
type VariableNode struct {
	ExpressionNode
	name string
}

// UnaryNode represents a unary operation (not, -, +)
type UnaryNode struct {
	ExpressionNode
	operator string
	node     Node
}

// BinaryNode represents a binary operation (+, -, *, /, etc)
type BinaryNode struct {
	ExpressionNode
	operator string
	left     Node
	right    Node
}

// FunctionNode represents a function call
type FunctionNode struct {
	ExpressionNode
	name       string
	args       []Node
	moduleExpr Node // Optional module for module.function() calls
}

// FilterNode represents a filter application
type FilterNode struct {
	ExpressionNode
	node   Node
	filter string
	args   []Node
}

// TestNode represents a test (is defined, is null, etc)
type TestNode struct {
	ExpressionNode
	node Node
	test string
	args []Node
}

// GetAttrNode represents attribute access (obj.attr)
type GetAttrNode struct {
	ExpressionNode
	node      Node
	attribute Node
}

// GetItemNode represents item access (array[key])
type GetItemNode struct {
	ExpressionNode
	node Node
	item Node
}

// MethodCallNode represents method call (obj.method())
type MethodCallNode struct {
	ExpressionNode
	node   Node
	method string
	args   []Node
}

// ArrayNode represents an array literal
type ArrayNode struct {
	ExpressionNode
	items []Node
}

// HashNode represents a hash/map literal
type HashNode struct {
	ExpressionNode
	items map[Node]Node
}

// ConditionalNode represents ternary operator (condition ? true : false)
type ConditionalNode struct {
	ExpressionNode
	condition Node
	trueExpr  Node
	falseExpr Node
}

// Type implementation for ExpressionNode
func (n *ExpressionNode) Type() NodeType {
	return NodeExpression
}

func (n *ExpressionNode) Line() int {
	return n.line
}

// Render implementation for LiteralNode
func (n *LiteralNode) Render(w io.Writer, ctx *RenderContext) error {
	var str string

	switch v := n.value.(type) {
	case string:
		str = v
	case int:
		str = strconv.Itoa(v)
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		str = strconv.FormatBool(v)
	case nil:
		str = ""
	default:
		str = ctx.ToString(v)
	}

	_, err := WriteString(w, str)
	return err
}

// Release returns a LiteralNode to the pool
func (n *LiteralNode) Release() {
	ReleaseLiteralNode(n)
}

// NewLiteralNode creates a new literal node
func NewLiteralNode(value interface{}, line int) *LiteralNode {
	node := GetLiteralNode(value, line)
	node.ExpressionNode.exprType = ExprLiteral
	return node
}

// NewVariableNode creates a new variable node
func NewVariableNode(name string, line int) *VariableNode {
	node := GetVariableNode(name, line)
	node.ExpressionNode.exprType = ExprVariable
	return node
}

// NewBinaryNode creates a new binary operation node
func NewBinaryNode(operator string, left, right Node, line int) *BinaryNode {
	return GetBinaryNode(operator, left, right, line)
}

// NewGetAttrNode creates a new attribute access node
func NewGetAttrNode(node, attribute Node, line int) *GetAttrNode {
	return GetGetAttrNode(node, attribute, line)
}

// NewGetItemNode creates a new item access node
func NewGetItemNode(node, item Node, line int) *GetItemNode {
	return GetGetItemNode(node, item, line)
}

// Render implementation for VariableNode
func (n *VariableNode) Render(w io.Writer, ctx *RenderContext) error {
	value, err := ctx.GetVariable(n.name)
	if err != nil {
		return err
	}

	// If debug is enabled, log variable access and value
	if IsDebugEnabled() {
		if value == nil {
			// Log undefined variable at error level if debug is enabled
			message := fmt.Sprintf("Variable lookup at line %d", n.line)
			LogError(fmt.Errorf("%w: %s", ErrUndefinedVar, n.name), message)

			// If in strict debug mode with error level, return an error for undefined variables
			if debugger.level >= DebugError && ctx.engine != nil && ctx.engine.debug {
				templateName := "unknown"
				if ctx.engine.currentTemplate != "" {
					templateName = ctx.engine.currentTemplate
				}
				return NewError(fmt.Errorf("%w: %s", ErrUndefinedVar, n.name), templateName, n.line, 0, "")
			}
		} else if debugger.level >= DebugVerbose {
			// Log defined variables at verbose level
			LogVerbose("Variable access at line %d: %s = %v (type: %T)", n.line, n.name, value, value)
		}
	}

	str := ctx.ToString(value)
	_, err = WriteString(w, str)
	return err
}

// Release returns a VariableNode to the pool
func (n *VariableNode) Release() {
	ReleaseVariableNode(n)
}

// Render implementation for GetAttrNode
func (n *GetAttrNode) Render(w io.Writer, ctx *RenderContext) error {
	obj, err := ctx.EvaluateExpression(n.node)
	if err != nil {
		return err
	}

	attrName, err := ctx.EvaluateExpression(n.attribute)
	if err != nil {
		return err
	}

	attrStr, ok := attrName.(string)
	if !ok {
		return fmt.Errorf("attribute name must be a string")
	}

	value, err := ctx.getAttribute(obj, attrStr)
	if err != nil {
		return err
	}

	str := ctx.ToString(value)
	_, err = WriteString(w, str)
	return err
}

// Release returns a GetAttrNode to the pool
func (n *GetAttrNode) Release() {
	ReleaseGetAttrNode(n)
}

// Render implementation for GetItemNode
func (n *GetItemNode) Render(w io.Writer, ctx *RenderContext) error {
	container, err := ctx.EvaluateExpression(n.node)
	if err != nil {
		return err
	}

	index, err := ctx.EvaluateExpression(n.item)
	if err != nil {
		return err
	}

	value, err := ctx.getItem(container, index)
	if err != nil {
		return err
	}

	str := ctx.ToString(value)
	_, err = WriteString(w, str)
	return err
}

// Release returns a GetItemNode to the pool
func (n *GetItemNode) Release() {
	ReleaseGetItemNode(n)
}

// Render implementation for BinaryNode
func (n *BinaryNode) Render(w io.Writer, ctx *RenderContext) error {
	result, err := ctx.EvaluateExpression(n)
	if err != nil {
		return err
	}

	str := ctx.ToString(result)
	_, err = WriteString(w, str)
	return err
}

// Release returns a BinaryNode to the pool
func (n *BinaryNode) Release() {
	ReleaseBinaryNode(n)
}

// Render implementation for FilterNode
func (n *FilterNode) Render(w io.Writer, ctx *RenderContext) error {
	result, err := ctx.EvaluateExpression(n)
	if err != nil {
		return err
	}

	str := ctx.ToString(result)
	_, err = WriteString(w, str)
	return err
}

// Release returns a FilterNode to the pool
func (n *FilterNode) Release() {
	ReleaseFilterNode(n)
}

// Render implementation for TestNode
func (n *TestNode) Render(w io.Writer, ctx *RenderContext) error {
	result, err := ctx.EvaluateExpression(n)
	if err != nil {
		return err
	}

	str := ctx.ToString(result)
	_, err = WriteString(w, str)
	return err
}

// Release returns a TestNode to the pool
func (n *TestNode) Release() {
	ReleaseTestNode(n)
}

// Render implementation for UnaryNode
func (n *UnaryNode) Render(w io.Writer, ctx *RenderContext) error {
	result, err := ctx.EvaluateExpression(n)
	if err != nil {
		return err
	}

	str := ctx.ToString(result)
	_, err = WriteString(w, str)
	return err
}

// Release returns a UnaryNode to the pool
func (n *UnaryNode) Release() {
	ReleaseUnaryNode(n)
}

// Render implementation for ConditionalNode
func (n *ConditionalNode) Render(w io.Writer, ctx *RenderContext) error {
	result, err := ctx.EvaluateExpression(n)
	if err != nil {
		return err
	}

	str := ctx.ToString(result)
	_, err = WriteString(w, str)
	return err
}

// Release returns a ConditionalNode to the pool
func (n *ConditionalNode) Release() {
	ReleaseConditionalNode(n)
}

// Render implementation for ArrayNode
func (n *ArrayNode) Render(w io.Writer, ctx *RenderContext) error {
	result, err := ctx.EvaluateExpression(n)
	if err != nil {
		return err
	}

	str := ctx.ToString(result)
	_, err = WriteString(w, str)
	return err
}

// Release returns an ArrayNode to the pool
func (n *ArrayNode) Release() {
	ReleaseArrayNode(n)
}

// Render implementation for HashNode
func (n *HashNode) Render(w io.Writer, ctx *RenderContext) error {
	result, err := ctx.EvaluateExpression(n)
	if err != nil {
		return err
	}

	str := ctx.ToString(result)
	_, err = WriteString(w, str)
	return err
}

// Release returns a HashNode to the pool
func (n *HashNode) Release() {
	ReleaseHashNode(n)
}

// Render implementation for FunctionNode
func (n *FunctionNode) Render(w io.Writer, ctx *RenderContext) error {
	result, err := ctx.EvaluateExpression(n)
	if err != nil {
		return err
	}

	str := ctx.ToString(result)
	_, err = WriteString(w, str)
	return err
}

// Release returns a FunctionNode to the pool
func (n *FunctionNode) Release() {
	ReleaseFunctionNode(n)
}

// NewFilterNode creates a new filter node
func NewFilterNode(node Node, filter string, args []Node, line int) *FilterNode {
	return GetFilterNode(node, filter, args, line)
}

// NewTestNode creates a new test node
func NewTestNode(node Node, test string, args []Node, line int) *TestNode {
	return GetTestNode(node, test, args, line)
}

// NewUnaryNode creates a new unary operation node
func NewUnaryNode(operator string, node Node, line int) *UnaryNode {
	return GetUnaryNode(operator, node, line)
}

// NewConditionalNode creates a new conditional (ternary) node
func NewConditionalNode(condition, trueExpr, falseExpr Node, line int) *ConditionalNode {
	return GetConditionalNode(condition, trueExpr, falseExpr, line)
}

// NewArrayNode creates a new array node
func NewArrayNode(items []Node, line int) *ArrayNode {
	return GetArrayNode(items, line)
}

// NewHashNode creates a new hash node
func NewHashNode(items map[Node]Node, line int) *HashNode {
	return GetHashNode(items, line)
}

// NewFunctionNode creates a new function call node
func NewFunctionNode(name string, args []Node, line int) *FunctionNode {
	return GetFunctionNode(name, args, line)
}

// ParseExpressionOptimized parses simple expressions with minimal allocations
// Returns the parsed expression node and a boolean indicating success
func ParseExpressionOptimized(expr string) (Node, bool) {
	// Trim whitespace
	expr = strings.TrimSpace(expr)
	
	// Quick empty check
	if len(expr) == 0 {
		return nil, false
	}
	
	// Try parsing as literal
	if val, ok := ParseLiteralOptimized(expr); ok {
		return NewLiteralNode(val, 0), true
	}
	
	// Check for simple variable references
	if IsValidVariableName(expr) {
		return NewVariableNode(expr, 0), true
	}
	
	// More complex expression - will need the full parser
	return nil, false
}

// ParseLiteralOptimized parses literals (strings, numbers, booleans) with minimal allocations
func ParseLiteralOptimized(expr string) (interface{}, bool) {
	// Quick check for common literal types
	if len(expr) == 0 {
		return nil, false
	}
	
	// Check for string literals
	if len(expr) >= 2 && ((expr[0] == '"' && expr[len(expr)-1] == '"') || 
	                      (expr[0] == '\'' && expr[len(expr)-1] == '\'')) {
		// String literal
		return expr[1 : len(expr)-1], true
	}
	
	// Check for number literals
	if isDigit(expr[0]) || (expr[0] == '-' && len(expr) > 1 && isDigit(expr[1])) {
		// Try parsing as integer first (most common case)
		if i, err := strconv.Atoi(expr); err == nil {
			return i, true
		}
		
		// Try parsing as float
		if f, err := strconv.ParseFloat(expr, 64); err == nil {
			return f, true
		}
	}
	
	// Check for boolean literals
	if expr == "true" {
		return true, true
	}
	if expr == "false" {
		return false, true
	}
	if expr == "null" || expr == "nil" {
		return nil, true
	}
	
	// Not a simple literal
	return nil, false
}

// IsValidVariableName checks if a string is a valid variable name
func IsValidVariableName(name string) bool {
	if len(name) == 0 {
		return false
	}
	
	// First character must be a letter or underscore
	if !isAlpha(name[0]) {
		return false
	}
	
	// Rest can be letters, digits, or underscores
	for i := 1; i < len(name); i++ {
		if !isNameChar(name[i]) {
			return false
		}
	}
	
	// Check for reserved keywords
	switch name {
	case "true", "false", "null", "nil", "not", "and", "or", "in", "is":
		return false
	}
	
	return true
}

// ProcessStringEscapes efficiently processes string escapes with minimal allocations
// Returns the processed string and a flag indicating if any processing was done
func ProcessStringEscapes(text string) (string, bool) {
	// Check if there are any escapes to process
	hasEscape := false
	for i := 0; i < len(text); i++ {
		if text[i] == '\\' && i+1 < len(text) {
			hasEscape = true
			break
		}
	}
	
	// If no escapes, return the original string
	if !hasEscape {
		return text, false
	}
	
	// Process escapes
	var sb strings.Builder
	sb.Grow(len(text))
	
	i := 0
	for i < len(text) {
		if text[i] == '\\' && i+1 < len(text) {
			// Escape sequence found
			i++
			
			// Process the escape sequence
			switch text[i] {
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case 'b':
				sb.WriteByte('\b')
			case 'f':
				sb.WriteByte('\f')
			case '\'':
				sb.WriteByte('\'')
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			default:
				// Unknown escape, just keep it as-is
				sb.WriteByte('\\')
				sb.WriteByte(text[i])
			}
		} else {
			// Regular character
			sb.WriteByte(text[i])
		}
		i++
	}
	
	// Return the processed string
	return sb.String(), true
}

// ParseNumberOptimized parses a number without using strconv for standard cases
// This provides better performance for typical integers and simple floats
func ParseNumberOptimized(text string) (interface{}, bool) {
	// Empty string is not a number
	if len(text) == 0 {
		return nil, false
	}
	
	// Check for negative sign
	isNegative := false
	pos := 0
	
	if text[pos] == '-' {
		isNegative = true
		pos++
		
		// Just a minus sign is not a number
		if pos >= len(text) {
			return nil, false
		}
	}
	
	// Parse the integer part
	var intValue int64
	hasDigits := false
	
	for pos < len(text) && isDigit(text[pos]) {
		digit := int64(text[pos] - '0')
		intValue = intValue*10 + digit
		hasDigits = true
		pos++
	}
	
	if !hasDigits {
		return nil, false
	}
	
	// Check for decimal point
	hasDecimal := false
	var floatValue float64
	
	if pos < len(text) && text[pos] == '.' {
		hasDecimal = true
		pos++
		
		// Need at least one digit after decimal point
		if pos >= len(text) || !isDigit(text[pos]) {
			return nil, false
		}
		
		// Parse the fractional part
		floatValue = float64(intValue)
		fraction := 0.0
		multiplier := 0.1
		
		for pos < len(text) && isDigit(text[pos]) {
			digit := float64(text[pos] - '0')
			fraction += digit * multiplier
			multiplier *= 0.1
			pos++
		}
		
		floatValue += fraction
	}
	
	// Check for exponent
	if pos < len(text) && (text[pos] == 'e' || text[pos] == 'E') {
		hasDecimal = true
		pos++
		
		// Parse exponent sign
		expNegative := false
		if pos < len(text) && (text[pos] == '+' || text[pos] == '-') {
			expNegative = text[pos] == '-'
			pos++
		}
		
		// Need at least one digit in exponent
		if pos >= len(text) || !isDigit(text[pos]) {
			return nil, false
		}
		
		// Parse exponent value
		exp := 0
		for pos < len(text) && isDigit(text[pos]) {
			digit := int(text[pos] - '0')
			exp = exp*10 + digit
			pos++
		}
		
		// Apply exponent
		// Convert to float regardless of hasDecimal
		hasDecimal = true
		if floatValue == 0 { // Not set yet
			floatValue = float64(intValue)
		}
		
		// Apply exponent using a more efficient approach
		if expNegative {
			// For negative exponents, divide by 10^exp
			multiplier := 1.0
			for i := 0; i < exp; i++ {
				multiplier *= 0.1
			}
			floatValue *= multiplier
		} else {
			// For positive exponents, multiply by 10^exp
			multiplier := 1.0
			for i := 0; i < exp; i++ {
				multiplier *= 10
			}
			floatValue *= multiplier
		}
	}
	
	// Check that we consumed the whole string
	if pos < len(text) {
		return nil, false
	}
	
	// Return the appropriate value
	if hasDecimal {
		if isNegative {
			return -floatValue, true
		}
		return floatValue, true
	} else {
		if isNegative {
			return -intValue, true
		}
		return intValue, true
	}
}

// Helper functions for character classification

// isAlpha checks if a character is a letter or underscore
func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// isNameChar checks if a character is valid in a name
func isNameChar(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

// isDigit is defined elsewhere

// Helper for optimized hash calculation
// This is useful for consistent hashing of variable names or strings
func calcStringHash(s string) uint32 {
	var h uint32
	
	for i := 0; i < len(s); i++ {
		h = 31*h + uint32(s[i])
	}
	
	return h
}
