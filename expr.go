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
	name string
	args []Node
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
	
	_, err := w.Write([]byte(str))
	return err
}

// NewLiteralNode creates a new literal node
func NewLiteralNode(value interface{}, line int) *LiteralNode {
	return &LiteralNode{
		ExpressionNode: ExpressionNode{
			exprType: ExprLiteral,
			line:     line,
		},
		value: value,
	}
}

// NewVariableNode creates a new variable node
func NewVariableNode(name string, line int) *VariableNode {
	return &VariableNode{
		ExpressionNode: ExpressionNode{
			exprType: ExprVariable,
			line:     line,
		},
		name: name,
	}
}

// NewBinaryNode creates a new binary operation node
func NewBinaryNode(operator string, left, right Node, line int) *BinaryNode {
	return &BinaryNode{
		ExpressionNode: ExpressionNode{
			exprType: ExprBinary,
			line:     line,
		},
		operator: operator,
		left:     left,
		right:    right,
	}
}

// NewGetAttrNode creates a new attribute access node
func NewGetAttrNode(node, attribute Node, line int) *GetAttrNode {
	return &GetAttrNode{
		ExpressionNode: ExpressionNode{
			exprType: ExprGetAttr,
			line:     line,
		},
		node:      node,
		attribute: attribute,
	}
}

// Render implementation for VariableNode
func (n *VariableNode) Render(w io.Writer, ctx *RenderContext) error {
	value, err := ctx.GetVariable(n.name)
	if err != nil {
		return err
	}
	
	str := ctx.ToString(value)
	_, err = w.Write([]byte(str))
	return err
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
	_, err = w.Write([]byte(str))
	return err
}

// Render implementation for BinaryNode
func (n *BinaryNode) Render(w io.Writer, ctx *RenderContext) error {
	result, err := ctx.EvaluateExpression(n)
	if err != nil {
		return err
	}
	
	str := ctx.ToString(result)
	_, err = w.Write([]byte(str))
	return err
}