package twig

import (
	"fmt"
)

// parseSandbox parses a sandbox tag
// {% sandbox %} ... {% endsandbox %}
func (p *Parser) parseSandbox() (Node, error) {
	// Get the line number of this tag for error reporting
	line := p.getCurrentToken().Line

	// Consume the sandbox token
	p.nextToken()

	// Consume the block end token
	if err := p.expectTokenType(TOKEN_BLOCK_END, TOKEN_BLOCK_END_TRIM); err != nil {
		return nil, err
	}

	// Parse the body of the sandbox
	body, err := p.parseUntilTag("endsandbox")
	if err != nil {
		return nil, err
	}

	// Consume the end sandbox token
	if err := p.expectTokenType(TOKEN_BLOCK_END, TOKEN_BLOCK_END_TRIM); err != nil {
		return nil, err
	}

	// Create and return the sandbox node
	return NewSandboxNode(body, line), nil
}

// SandboxNode represents a {% sandbox %} ... {% endsandbox %} block
type SandboxNode struct {
	body []Node
	line int
}

// NewSandboxNode creates a new sandbox node
func NewSandboxNode(body []Node, line int) *SandboxNode {
	return &SandboxNode{
		body: body,
		line: line,
	}
}

// Render renders the node to a writer
func (n *SandboxNode) Render(w io.Writer, ctx *RenderContext) error {
	// Create a sandboxed rendering context
	sandboxCtx := ctx.Clone()
	sandboxCtx.EnableSandbox()

	if sandboxCtx.environment.securityPolicy == nil {
		return fmt.Errorf("sandbox error: no security policy defined")
	}

	// Render all body nodes within the sandboxed context
	for _, node := range n.body {
		err := node.Render(w, sandboxCtx)
		if err != nil {
			return fmt.Errorf("sandbox error: %w", err)
		}
	}

	return nil
}

// Line returns the line number of the node
func (n *SandboxNode) Line() int {
	return n.line
}

// Type returns the node type
func (n *SandboxNode) Type() NodeType {
	return NodeSandbox
}