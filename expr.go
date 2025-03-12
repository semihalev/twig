package twig

import (
	"fmt"
	"io"
	"strconv"
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
