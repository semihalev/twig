package twig

import (
	"io"
)

// Node represents a node in the template parse tree
type Node interface {
	// Render renders the node to the output
	Render(w io.Writer, ctx *RenderContext) error
	
	// Type returns the node type
	Type() NodeType
	
	// Line returns the source line number
	Line() int
}

// NodeType represents the type of a node
type NodeType int

// Node types
const (
	NodeRoot NodeType = iota
	NodeText
	NodePrint
	NodeIf
	NodeFor
	NodeBlock
	NodeExtends
	NodeInclude
	NodeImport
	NodeMacro
	NodeSet
	NodeExpression
	NodeComment
	NodeVerbatim
	NodeElement
)

// RootNode represents the root of a template
type RootNode struct {
	children []Node
	line     int
}

// TextNode represents a raw text node
type TextNode struct {
	content string
	line    int
}

// PrintNode represents a {{ expression }} node
type PrintNode struct {
	expression Node
	line       int
}

// IfNode represents an if block
type IfNode struct {
	conditions []Node
	bodies     [][]Node
	elseBranch []Node
	line       int
}

// ForNode represents a for loop
type ForNode struct {
	keyVar     string
	valueVar   string
	sequence   Node
	body       []Node
	elseBranch []Node
	line       int
}

// BlockNode represents a block definition
type BlockNode struct {
	name string
	body []Node
	line int
}

// ExtendsNode represents template inheritance
type ExtendsNode struct {
	parent Node
	line   int
}

// IncludeNode represents template inclusion
type IncludeNode struct {
	template   Node
	variables  map[string]Node
	ignoreMissing bool
	only       bool
	line       int
}

// CommentNode represents a {# comment #}
type CommentNode struct {
	content string
	line    int
}

// Implement Node interface for RootNode
func (n *RootNode) Render(w io.Writer, ctx *RenderContext) error {
	for _, child := range n.children {
		if err := child.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (n *RootNode) Type() NodeType {
	return NodeRoot
}

func (n *RootNode) Line() int {
	return n.line
}

// Implement Node interface for TextNode
func (n *TextNode) Render(w io.Writer, ctx *RenderContext) error {
	_, err := w.Write([]byte(n.content))
	return err
}

func (n *TextNode) Type() NodeType {
	return NodeText
}

func (n *TextNode) Line() int {
	return n.line
}

// Implement Node interface for PrintNode
func (n *PrintNode) Render(w io.Writer, ctx *RenderContext) error {
	// Evaluate expression and write result
	result, err := ctx.EvaluateExpression(n.expression)
	if err != nil {
		return err
	}
	
	// Convert result to string and write
	str := ctx.ToString(result)
	_, err = w.Write([]byte(str))
	return err
}

func (n *PrintNode) Type() NodeType {
	return NodePrint
}

func (n *PrintNode) Line() int {
	return n.line
}

// NewRootNode creates a new root node
func NewRootNode(children []Node, line int) *RootNode {
	return &RootNode{
		children: children,
		line:     line,
	}
}

// NewTextNode creates a new text node
func NewTextNode(content string, line int) *TextNode {
	return &TextNode{
		content: content,
		line:    line,
	}
}

// NewPrintNode creates a new print node
func NewPrintNode(expression Node, line int) *PrintNode {
	return &PrintNode{
		expression: expression,
		line:       line,
	}
}