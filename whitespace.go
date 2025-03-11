package twig

import (
	"bytes"
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
	// First render body content to a buffer
	var buf bytes.Buffer

	// Render all body nodes
	for _, node := range n.body {
		err := node.Render(&buf, ctx)
		if err != nil {
			return err
		}
	}

	// Apply spaceless filter to the rendered content
	result, err := ctx.ApplyFilter("spaceless", buf.String())
	if err != nil {
		// Fall back to original content on filter error
		_, err = w.Write(buf.Bytes())
		return err
	}

	// Write the processed result
	_, err = WriteString(w, ctx.ToString(result))
	return err
}

// Line returns the line number of the node
func (n *SpacelessNode) Line() int {
	return n.line
}

// Type returns the node type
func (n *SpacelessNode) Type() NodeType {
	return NodeSpaceless
}
