package twig

import (
	"fmt"
	"io"
	"reflect"
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

// SetNode represents a variable assignment
type SetNode struct {
	name       string
	value      Node
	line       int
}

// CommentNode represents a {# comment #}
type CommentNode struct {
	content string
	line    int
}

// Implement Node interface for RootNode
func (n *RootNode) Render(w io.Writer, ctx *RenderContext) error {
	// Check if this is an extending template
	var extendsNode *ExtendsNode
	for _, child := range n.children {
		if node, ok := child.(*ExtendsNode); ok {
			extendsNode = node
			break
		}
	}
	
	// If this template extends another, don't render text nodes and only collect blocks
	if extendsNode != nil {
		// Set the extending flag
		ctx.extending = true
		
		// First, collect all blocks
		for _, child := range n.children {
			if _, ok := child.(*BlockNode); ok {
				// Render block nodes to register them in the context
				if err := child.Render(w, ctx); err != nil {
					return err
				}
			}
		}
		
		// Turn off the extending flag for the parent template
		ctx.extending = false
		
		// Then render the extends node
		if err := extendsNode.Render(w, ctx); err != nil {
			return err
		}
		
		return nil
	}
	
	// Regular template (not extending)
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

// Implement Node interface for IfNode
func (n *IfNode) Render(w io.Writer, ctx *RenderContext) error {
	// Evaluate the conditions one by one
	for i, condition := range n.conditions {
		// Evaluate the condition
		result, err := ctx.EvaluateExpression(condition)
		if err != nil {
			return err
		}
		
		// Check if the condition is true
		if ctx.toBool(result) {
			// Render the corresponding body
			for _, node := range n.bodies[i] {
				if err := node.Render(w, ctx); err != nil {
					return err
				}
			}
			return nil
		}
	}
	
	// If no condition matched, render the else branch if it exists
	if len(n.elseBranch) > 0 {
		for _, node := range n.elseBranch {
			if err := node.Render(w, ctx); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (n *IfNode) Type() NodeType {
	return NodeIf
}

func (n *IfNode) Line() int {
	return n.line
}

// Implement Node interface for ForNode
func (n *ForNode) Render(w io.Writer, ctx *RenderContext) error {
	// Evaluate the sequence expression
	sequence, err := ctx.EvaluateExpression(n.sequence)
	if err != nil {
		return err
	}
	
	// Check if we have anything to iterate over
	hasItems := false
	
	// Create local context for the loop variables
	// Use reflect to handle different types of sequences
	v := reflect.ValueOf(sequence)
	
	// For nil values or empty collections, skip to else branch
	if sequence == nil || 
	  (v.Kind() == reflect.Slice && v.Len() == 0) || 
	  (v.Kind() == reflect.Map && v.Len() == 0) || 
	  (v.Kind() == reflect.String && v.Len() == 0) {
		if len(n.elseBranch) > 0 {
			for _, node := range n.elseBranch {
				if err := node.Render(w, ctx); err != nil {
					return err
				}
			}
		}
		return nil
	}
	
	// Create a child context
	loopContext := &RenderContext{
		env:     ctx.env,
		context: make(map[string]interface{}),
		blocks:  ctx.blocks,
		macros:  ctx.macros,
		parent:  ctx,
	}
	
	// Copy all variables from the parent context
	for k, v := range ctx.context {
		loopContext.context[k] = v
	}
	
	// Handle different types of sequences
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		hasItems = v.Len() > 0
		// Iterate over the slice/array
		for i := 0; i < v.Len(); i++ {
			// Set loop variables
			if n.keyVar != "" {
				loopContext.context[n.keyVar] = i
			}
			loopContext.context[n.valueVar] = v.Index(i).Interface()
			
			// Add loop metadata
			loopContext.context["loop"] = map[string]interface{}{
				"index":     i + 1,
				"index0":    i,
				"revindex":  v.Len() - i,
				"revindex0": v.Len() - i - 1,
				"first":     i == 0,
				"last":      i == v.Len()-1,
				"length":    v.Len(),
			}
			
			// Render the loop body
			for _, node := range n.body {
				if err := node.Render(w, loopContext); err != nil {
					return err
				}
			}
		}
		
	case reflect.Map:
		hasItems = v.Len() > 0
		// Get the map keys
		keys := v.MapKeys()
		
		// Iterate over the map
		for i, key := range keys {
			// Set loop variables
			if n.keyVar != "" {
				loopContext.context[n.keyVar] = key.Interface()
			}
			loopContext.context[n.valueVar] = v.MapIndex(key).Interface()
			
			// Add loop metadata
			loopContext.context["loop"] = map[string]interface{}{
				"index":     i + 1,
				"index0":    i,
				"revindex":  len(keys) - i,
				"revindex0": len(keys) - i - 1,
				"first":     i == 0,
				"last":      i == len(keys)-1,
				"length":    len(keys),
			}
			
			// Render the loop body
			for _, node := range n.body {
				if err := node.Render(w, loopContext); err != nil {
					return err
				}
			}
		}
		
	case reflect.String:
		str := v.String()
		hasItems = len(str) > 0
		
		// Iterate over string characters
		for i, char := range str {
			// Set loop variables
			if n.keyVar != "" {
				loopContext.context[n.keyVar] = i
			}
			loopContext.context[n.valueVar] = string(char)
			
			// Add loop metadata
			loopContext.context["loop"] = map[string]interface{}{
				"index":     i + 1,
				"index0":    i,
				"revindex":  len(str) - i,
				"revindex0": len(str) - i - 1,
				"first":     i == 0,
				"last":      i == len(str)-1,
				"length":    len(str),
			}
			
			// Render the loop body
			for _, node := range n.body {
				if err := node.Render(w, loopContext); err != nil {
					return err
				}
			}
		}
		
	default:
		// For non-iterable types, just use the value as is
		// This might not be ideal, but it's more forgiving
		loopContext.context[n.valueVar] = sequence
		
		// Add minimal loop metadata
		loopContext.context["loop"] = map[string]interface{}{
			"index":     1,
			"index0":    0,
			"revindex":  1,
			"revindex0": 0,
			"first":     true,
			"last":      true,
			"length":    1,
		}
		
		// Render the loop body
		for _, node := range n.body {
			if err := node.Render(w, loopContext); err != nil {
				return err
			}
		}
		
		hasItems = true
	}
	
	// If no items and we have an else branch, render it
	if !hasItems && len(n.elseBranch) > 0 {
		for _, node := range n.elseBranch {
			if err := node.Render(w, ctx); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (n *ForNode) Type() NodeType {
	return NodeFor
}

func (n *ForNode) Line() int {
	return n.line
}

// Implement Node interface for BlockNode
func (n *BlockNode) Render(w io.Writer, ctx *RenderContext) error {
	// During extension, just register the block
	if ctx.extending {
		if ctx.blocks == nil {
			ctx.blocks = make(map[string][]Node)
		}
		ctx.blocks[n.name] = []Node{n} // Replace any parent blocks - child blocks take precedence
		return nil
	}
	
	// For regular rendering, find the block to use
	var blockToRender *BlockNode
	
	// Use the most recent block definition
	if blocks, ok := ctx.blocks[n.name]; ok && len(blocks) > 0 {
		// Use the block definition from the child template
		blockToRender = blocks[len(blocks)-1].(*BlockNode)
	} else {
		// If no overrides found, use this one
		blockToRender = n
	}
	
	// Set current block for parent() function
	oldBlock := ctx.currentBlock
	ctx.currentBlock = blockToRender
	
	// Render the block contents
	for _, node := range blockToRender.body {
		if err := node.Render(w, ctx); err != nil {
			return err
		}
	}
	
	// Restore previous block
	ctx.currentBlock = oldBlock
	
	return nil
}

func (n *BlockNode) Type() NodeType {
	return NodeBlock
}

func (n *BlockNode) Line() int {
	return n.line
}

// Implement Node interface for ExtendsNode
func (n *ExtendsNode) Render(w io.Writer, ctx *RenderContext) error {
	// Get the parent template name
	parentName, err := ctx.EvaluateExpression(n.parent)
	if err != nil {
		return err
	}
	
	// Convert to string if needed
	parentNameStr := ctx.ToString(parentName)
	
	// Get the parent template from the engine
	engine := ctx.engine
	if engine == nil {
		return fmt.Errorf("no engine available in context")
	}
	
	parent, err := engine.Load(parentNameStr)
	if err != nil {
		return err
	}
	
	// Setup a new context for the parent template, using the blocks that have already been collected
	parentCtx := &RenderContext{
		env:          ctx.env,
		context:      ctx.context,
		blocks:       ctx.blocks,  // Share blocks with the child template
		macros:       ctx.macros,
		parent:       ctx.parent,
		engine:       ctx.engine,
		extending:    false,       // The parent will be the final rendering
		currentBlock: nil,
	}
	
	// Now render the parent template which will use our blocks
	if err := parent.nodes.Render(w, parentCtx); err != nil {
		return err
	}
	
	return nil
}

func (n *ExtendsNode) Type() NodeType {
	return NodeExtends
}

func (n *ExtendsNode) Line() int {
	return n.line
}

// Implement Node interface for IncludeNode
func (n *IncludeNode) Render(w io.Writer, ctx *RenderContext) error {
	// Get the template name
	templateName, err := ctx.EvaluateExpression(n.template)
	if err != nil {
		return err
	}
	
	// Convert to string if needed
	templateNameStr := ctx.ToString(templateName)
	
	// Get the template from the engine
	engine := ctx.engine
	if engine == nil {
		return fmt.Errorf("no engine available in context")
	}
	
	// Load the template
	template, err := engine.Load(templateNameStr)
	if err != nil {
		if n.ignoreMissing {
			// Skip if we're ignoring missing templates
			return nil
		}
		return err
	}
	
	// Create a context for the included template
	includeCtx := &RenderContext{
		env:          ctx.env,
		context:      make(map[string]interface{}),
		blocks:       ctx.blocks,  // Share blocks with the parent template
		macros:       ctx.macros,
		parent:       nil,         // Will set this based on the 'only' flag below
		engine:       ctx.engine,
		extending:    false,
		currentBlock: nil,
	}
	
	// If not "only", copy all variables from parent context
	if !n.only {
		for k, v := range ctx.context {
			includeCtx.context[k] = v
		}
		
		// Also allow access to parent for variable lookup
		includeCtx.parent = ctx
	} else {
		// When using 'only', don't allow access to parent context
		includeCtx.parent = nil
	}
	
	// Evaluate and add the with variables
	if n.variables != nil {
		for name, node := range n.variables {
			value, err := ctx.EvaluateExpression(node)
			if err != nil {
				return err
			}
			includeCtx.context[name] = value
		}
	}
	
	// Render the included template
	return template.nodes.Render(w, includeCtx)
}

func (n *IncludeNode) Type() NodeType {
	return NodeInclude
}

func (n *IncludeNode) Line() int {
	return n.line
}

// Implement Node interface for SetNode
func (n *SetNode) Render(w io.Writer, ctx *RenderContext) error {
	// Evaluate the value expression
	value, err := ctx.EvaluateExpression(n.value)
	if err != nil {
		return err
	}
	
	// Set the variable in the context
	ctx.SetVariable(n.name, value)
	
	// The set tag doesn't output anything
	return nil
}

func (n *SetNode) Type() NodeType {
	return NodeSet
}

func (n *SetNode) Line() int {
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

// NewForNode creates a new for loop node
func NewForNode(keyVar, valueVar string, sequence Node, body, elseBranch []Node, line int) *ForNode {
	return &ForNode{
		keyVar:     keyVar,
		valueVar:   valueVar,
		sequence:   sequence,
		body:       body,
		elseBranch: elseBranch,
		line:       line,
	}
}

// NewIfNode creates a new if node
func NewIfNode(condition Node, body, elseBranch []Node, line int) *IfNode {
	return &IfNode{
		conditions: []Node{condition},
		bodies:     [][]Node{body},
		elseBranch: elseBranch,
		line:       line,
	}
}

// NewBlockNode creates a new block node
func NewBlockNode(name string, body []Node, line int) *BlockNode {
	return &BlockNode{
		name: name,
		body: body,
		line: line,
	}
}

// NewExtendsNode creates a new extends node
func NewExtendsNode(parent Node, line int) *ExtendsNode {
	return &ExtendsNode{
		parent: parent,
		line:   line,
	}
}

// NewIncludeNode creates a new include node
func NewIncludeNode(template Node, variables map[string]Node, ignoreMissing, only bool, line int) *IncludeNode {
	return &IncludeNode{
		template:     template,
		variables:    variables,
		ignoreMissing: ignoreMissing,
		only:         only,
		line:         line,
	}
}

// NewSetNode creates a new set node
func NewSetNode(name string, value Node, line int) *SetNode {
	return &SetNode{
		name:  name,
		value: value,
		line:  line,
	}
}