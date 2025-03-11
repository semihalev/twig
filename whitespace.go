package twig

import (
	"io"
	"strings"
)

// trimLeadingWhitespace removes leading whitespace from a string
// This is only used for whitespace control in templates ({{- and -}})
func trimLeadingWhitespace(s string) string {
	return strings.TrimLeft(s, " \t\n\r")
}

// trimTrailingWhitespace removes trailing whitespace from a string
// This is only used for whitespace control in templates ({{- and -}})
func trimTrailingWhitespace(s string) string {
	return strings.TrimRight(s, " \t\n\r")
}

// SpacelessNode represents a {% spaceless %} ... {% endspaceless %} block
type SpacelessNode struct {
	body []Node
	line int
}

// NewSpacelessNode creates a new spaceless node
func NewSpacelessNode(body []Node, line int) *SpacelessNode {
	return &SpacelessNode{
		body: body,
		line: line,
	}
}

// Render renders the node to a writer
func (n *SpacelessNode) Render(w io.Writer, ctx *RenderContext) error {
	// Just render the content directly - we don't manipulate HTML
	for _, node := range n.body {
		if err := node.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

// Line returns the line number of the node
func (n *SpacelessNode) Line() int {
	return n.line
}

// Type returns the node type
func (n *SpacelessNode) Type() NodeType {
	return NodeSpaceless
}
