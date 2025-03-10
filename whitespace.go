package twig

import (
	"io"
	"regexp"
	"strings"
)

// trimLeadingWhitespace removes leading whitespace from a string
func trimLeadingWhitespace(s string) string {
	return strings.TrimLeft(s, " \t\n\r")
}

// trimTrailingWhitespace removes trailing whitespace from a string
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
	// First, render the content to a string using a buffer
	var buf StringBuffer

	for _, node := range n.body {
		if err := node.Render(&buf, ctx); err != nil {
			return err
		}
	}

	// Get the rendered content as a string
	content := buf.String()

	// Apply spaceless processing (remove whitespace between HTML tags)
	result := removeWhitespaceBetweenTags(content)

	// Write the processed result
	_, err := w.Write([]byte(result))
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

// removeWhitespaceBetweenTags removes whitespace between HTML tags
// but preserves whitespace between words
func removeWhitespaceBetweenTags(content string) string {
	// This regex matches whitespace between HTML tags only
	re := regexp.MustCompile(">\\s+<")
	return re.ReplaceAllString(content, "><")
}
