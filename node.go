package twig

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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

// NewIfNode creates a new if node
func NewIfNode(conditions []Node, bodies [][]Node, elseBranch []Node, line int) *IfNode {
	return &IfNode{
		conditions: conditions,
		bodies:     bodies,
		elseBranch: elseBranch,
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
		template:      template,
		variables:     variables,
		ignoreMissing: ignoreMissing,
		only:          only,
		line:          line,
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

// NewCommentNode creates a new comment node
func NewCommentNode(content string, line int) *CommentNode {
	return &CommentNode{
		content: content,
		line:    line,
	}
}

// NewMacroNode creates a new macro node
func NewMacroNode(name string, params []string, defaults map[string]Node, body []Node, line int) *MacroNode {
	return &MacroNode{
		name:     name,
		params:   params,
		defaults: defaults,
		body:     body,
		line:     line,
	}
}

// NewImportNode creates a new import node
func NewImportNode(template Node, module string, line int) *ImportNode {
	return &ImportNode{
		template: template,
		module:   module,
		line:     line,
	}
}

// NewFromImportNode creates a new from import node
func NewFromImportNode(template Node, macros []string, aliases map[string]string, line int) *FromImportNode {
	return &FromImportNode{
		template: template,
		macros:   macros,
		aliases:  aliases,
		line:     line,
	}
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
	NodeFunction
	NodeSpaceless
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

// String implementation for debugging
func (n *TextNode) String() string {
	// Display spaces as visible characters for easier debugging
	spacesVisual := strings.ReplaceAll(n.content, " ", "·")
	return fmt.Sprintf("TextNode(%q [%s], line: %d)", n.content, spacesVisual, n.line)
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
	template      Node
	variables     map[string]Node
	ignoreMissing bool
	only          bool
	line          int
}

// SetNode represents a variable assignment
type SetNode struct {
	name  string
	value Node
	line  int
}

// CommentNode represents a {# comment #}
type CommentNode struct {
	content string
	line    int
}

// We use the FunctionNode from expr.go

// MacroNode represents a macro definition
type MacroNode struct {
	name     string
	params   []string
	defaults map[string]Node
	body     []Node
	line     int
}

// ImportNode represents an import statement
type ImportNode struct {
	template Node
	module   string
	line     int
}

// FromImportNode represents a from import statement
type FromImportNode struct {
	template Node
	macros   []string
	aliases  map[string]string
	line     int
}

// NullWriter is a writer that discards all data
type NullWriter struct{}

// Write implements io.Writer for NullWriter
func (w *NullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
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
	content := n.content
	
	// Track entry into script and style tags
	// This affects how we handle whitespace and quoting of values
	lowerContent := strings.ToLower(content)
	
	// Check for <script> tags (case-insensitive)
	if strings.Contains(lowerContent, "<script") && !strings.Contains(lowerContent, "</script") {
		ctx.inScriptTag = true
	} else if strings.Contains(lowerContent, "</script>") {
		// Set flag to false AFTER rendering this node
		defer func() { ctx.inScriptTag = false }()
	}
	
	// Check for <style> tags (case-insensitive)
	if strings.Contains(lowerContent, "<style") && !strings.Contains(lowerContent, "</style") {
		ctx.inStyleTag = true
	} else if strings.Contains(lowerContent, "</style>") {
		// Set flag to false AFTER rendering this node
		defer func() { ctx.inStyleTag = false }()
	}
	
	// Apply whitespace processing based on Environment settings
	if ctx.env != nil {
		// Preserve whitespace feature is enabled by default
		if ctx.env.preserveWhitespace {
			_, err := w.Write([]byte(content))
			return err
		}
		
		// Process internal Script/Style tag contents specially or entering nodes
		if ctx.inScriptTag || ctx.inStyleTag || 
		   strings.Contains(lowerContent, "<script") || strings.Contains(lowerContent, "<style") {
			_, err := w.Write([]byte(content))
			return err
		}
		
		// Process HTML content
		containsHTML := strings.Contains(content, "<") && strings.Contains(content, ">")
		
		if containsHTML && ctx.env.prettyOutputHTML {
			// Add spaces after colons in text content to improve readability
			colonFixRegex := regexp.MustCompile(`:\s*(\S)`)
			content = colonFixRegex.ReplaceAllString(content, ": $1")
			
			// Process HTML attributes if enabled
			if ctx.env.preserveAttributes {
				// Ensure quotes around attribute values
				attrQuoteRegex := regexp.MustCompile(`=([a-zA-Z0-9_.-]+)(\s|>)`)
				content = attrQuoteRegex.ReplaceAllString(content, `="$1"$2`)
				
				// Fix self-closing tags to have proper spacing
				selfCloseRegex := regexp.MustCompile(`/\s*>`)
				content = selfCloseRegex.ReplaceAllString(content, " />")
			}
			
			// Process HTML tags for better spacing if pretty output is enabled
			// Add space after > and before < to separate tags from content
			tagTextRegex := regexp.MustCompile(`(>)(\S)`)
			content = tagTextRegex.ReplaceAllString(content, "$1 $2")
			
			// Add space before < if preceded by non-whitespace
			textTagRegex := regexp.MustCompile(`(\S)(<)`)
			content = textTagRegex.ReplaceAllString(content, "$1 $2")
		} else if !containsHTML && ctx.env.prettyOutputHTML {
			// For plain text content, normalize multiple spaces to single space
			wsRegex := regexp.MustCompile(`[ \t]+`)
			content = wsRegex.ReplaceAllString(content, " ")
		}
	}
	
	_, err := w.Write([]byte(content))
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

	// Check if result is a callable (for macros)
	if callable, ok := result.(func(io.Writer) error); ok {
		// Execute the callable directly
		return callable(w)
	}

	// Convert result to string
	str := ctx.ToString(result)

	// Special handling for default filter values to ensure proper string formatting
	// in both JavaScript and CSS contexts
	filter, ok := n.expression.(*FilterNode)
	if ok && filter.filter == "default" {
		// Check if this looks like a CSS value that shouldn't be quoted
		isCSSValue := false

		// Common CSS units that shouldn't be quoted
		if strings.HasPrefix(str, "#") || // Colors
			strings.HasSuffix(str, "px") || // Pixel values
			strings.HasSuffix(str, "em") || // Em values
			strings.HasSuffix(str, "rem") || // Root em values
			strings.HasSuffix(str, "vh") || // Viewport height
			strings.HasSuffix(str, "vw") || // Viewport width
			strings.HasSuffix(str, "%") { // Percentage
			isCSSValue = true
		}

		// Font-family values should be treated as already correctly formatted
		// in the template where individual font names are quoted as needed
		if !isCSSValue && strings.Contains(str, ",") &&
			(strings.Contains(strings.ToLower(str), "serif") ||
				strings.Contains(strings.ToLower(str), "sans") ||
				strings.Contains(strings.ToLower(str), "monospace") ||
				strings.Contains(strings.ToLower(str), "arial") ||
				strings.Contains(strings.ToLower(str), "helvetica") ||
				strings.Contains(strings.ToLower(str), "roboto") ||
				strings.Contains(strings.ToLower(str), "font")) {
			// Font family specifications should be treated as CSS values
			isCSSValue = true
		}

		// Check for numeric values
		isNumber := false
		_, err1 := strconv.ParseInt(str, 10, 64)
		if err1 == nil {
			isNumber = true
		} else {
			_, err2 := strconv.ParseFloat(str, 64)
			if err2 == nil {
				isNumber = true
			}
		}

		// If not a CSS value that should remain unquoted, apply JavaScript string quoting
		if !isCSSValue {
			// Skip quoting for null, undefined, numeric values, and boolean literals
			if str != "null" && str != "undefined" && !isNumber &&
				str != "true" && str != "false" {
				// Check if it's already quoted
				if !(len(str) >= 2 &&
					((str[0] == '\'' && str[len(str)-1] == '\'') ||
						(str[0] == '"' && str[len(str)-1] == '"'))) {
					// Add quotes
					str = "'" + str + "'"
				}
			}
		}
	}

	// Write the result
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

	// Iterate over the sequence based on its type
	switch v.Kind() {
	case reflect.Map:
		// Initialize loop.index and loop.index0
		loopContext.context["loop"] = map[string]interface{}{
			"index":  1,
			"index0": 0,
			"first":  true,
			"last":   false,
			"length": v.Len(),
		}

		keys := v.MapKeys()
		for i, key := range keys {
			// Update loop variables
			loop := loopContext.context["loop"].(map[string]interface{})
			loop["index"] = i + 1
			loop["index0"] = i
			loop["first"] = i == 0
			loop["last"] = i == len(keys)-1

			// Set the key and value variables
			if n.keyVar != "" {
				loopContext.context[n.keyVar] = key.Interface()
			}
			loopContext.context[n.valueVar] = v.MapIndex(key).Interface()

			// Render the loop body
			for _, node := range n.body {
				if err := node.Render(w, loopContext); err != nil {
					return err
				}
			}

			hasItems = true
		}

	case reflect.Slice, reflect.Array:
		// Initialize loop.index and loop.index0
		loopContext.context["loop"] = map[string]interface{}{
			"index":  1,
			"index0": 0,
			"first":  true,
			"last":   false,
			"length": v.Len(),
		}

		for i := 0; i < v.Len(); i++ {
			// Update loop variables
			loop := loopContext.context["loop"].(map[string]interface{})
			loop["index"] = i + 1
			loop["index0"] = i
			loop["first"] = i == 0
			loop["last"] = i == v.Len()-1

			// Set the key and value variables
			if n.keyVar != "" {
				loopContext.context[n.keyVar] = i
			}
			loopContext.context[n.valueVar] = v.Index(i).Interface()

			// Render the loop body
			for _, node := range n.body {
				if err := node.Render(w, loopContext); err != nil {
					return err
				}
			}

			hasItems = true
		}

	case reflect.String:
		// String iteration iterates over runes (unicode code points)
		str := v.String()
		runes := []rune(str)

		// Initialize loop.index and loop.index0
		loopContext.context["loop"] = map[string]interface{}{
			"index":  1,
			"index0": 0,
			"first":  true,
			"last":   false,
			"length": len(runes),
		}

		for i, r := range runes {
			// Update loop variables
			loop := loopContext.context["loop"].(map[string]interface{})
			loop["index"] = i + 1
			loop["index0"] = i
			loop["first"] = i == 0
			loop["last"] = i == len(runes)-1

			// Set the key and value variables
			if n.keyVar != "" {
				loopContext.context[n.keyVar] = i
			}
			loopContext.context[n.valueVar] = string(r)

			// Render the loop body
			for _, node := range n.body {
				if err := node.Render(w, loopContext); err != nil {
					return err
				}
			}

			hasItems = true
		}

	default:
		// For other types, we can't iterate
		// Render the else branch if present
		if len(n.elseBranch) > 0 {
			for _, node := range n.elseBranch {
				if err := node.Render(w, ctx); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// If we didn't iterate over anything and there's an else branch, render it
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
	// Register this block in the context
	if ctx.blocks == nil {
		ctx.blocks = make(map[string][]Node)
	}
	ctx.blocks[n.name] = n.body

	// If this is an extending template, don't render the block yet
	if ctx.extending {
		return nil
	}

	// Save the current block for the parent() function
	oldBlock := ctx.currentBlock
	ctx.currentBlock = n

	// Render the block content
	for _, node := range n.body {
		if err := node.Render(w, ctx); err != nil {
			return err
		}
	}

	// Restore the previous block
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
	// Evaluate the parent template name
	parentName, err := ctx.EvaluateExpression(n.parent)
	if err != nil {
		return err
	}

	// Get the parent template name as a string
	parentNameStr := ctx.ToString(parentName)

	// Check if we have an engine to load the parent template
	if ctx.engine == nil {
		return fmt.Errorf("cannot extend without an engine")
	}

	// Load the parent template
	parentTemplate, err := ctx.engine.Load(parentNameStr)
	if err != nil {
		return err
	}

	// Render the parent template
	return parentTemplate.nodes.Render(w, ctx)
}

func (n *ExtendsNode) Type() NodeType {
	return NodeExtends
}

func (n *ExtendsNode) Line() int {
	return n.line
}

// Implement Node interface for IncludeNode
func (n *IncludeNode) Render(w io.Writer, ctx *RenderContext) error {
	// Evaluate the template name
	templateName, err := ctx.EvaluateExpression(n.template)
	if err != nil {
		return err
	}

	// Get the template name as a string
	templateNameStr := ctx.ToString(templateName)

	// Check if we have an engine to load the included template
	if ctx.engine == nil {
		return fmt.Errorf("cannot include without an engine")
	}

	// Load the included template
	includedTemplate, err := ctx.engine.Load(templateNameStr)
	if err != nil {
		if n.ignoreMissing {
			// If ignore_missing is true, silently skip the inclusion
			return nil
		}
		return err
	}

	// Create a child context for the included template
	childCtx := &RenderContext{
		env:     ctx.env,
		context: make(map[string]interface{}),
		blocks:  ctx.blocks,
		macros:  ctx.macros,
		engine:  ctx.engine,
		parent:  ctx,
	}

	// If only=true, don't inherit the parent context
	if !n.only {
		// Copy all variables from the parent context
		for k, v := range ctx.context {
			childCtx.context[k] = v
		}
	}

	// Add the variables defined in the include tag
	if n.variables != nil {
		for name, valueNode := range n.variables {
			value, err := ctx.EvaluateExpression(valueNode)
			if err != nil {
				return err
			}
			childCtx.context[name] = value
		}
	}

	// Render the included template
	return includedTemplate.nodes.Render(w, childCtx)
}

func (n *IncludeNode) Type() NodeType {
	return NodeInclude
}

func (n *IncludeNode) Line() int {
	return n.line
}

// Implement Node interface for SetNode
func (n *SetNode) Render(w io.Writer, ctx *RenderContext) error {
	// Evaluate the expression
	value, err := ctx.EvaluateExpression(n.value)
	if err != nil {
		return err
	}

	// Set the variable in the context
	ctx.context[n.name] = value

	return nil
}

func (n *SetNode) Type() NodeType {
	return NodeSet
}

func (n *SetNode) Line() int {
	return n.line
}

// Implement Node interface for CommentNode
func (n *CommentNode) Render(w io.Writer, ctx *RenderContext) error {
	// Comments don't render anything
	return nil
}

func (n *CommentNode) Type() NodeType {
	return NodeComment
}

func (n *CommentNode) Line() int {
	return n.line
}

// Implement Node interface for MacroNode
func (n *MacroNode) Render(w io.Writer, ctx *RenderContext) error {
	// Create a macro function that captures the current context
	macro := func(w io.Writer, args ...interface{}) error {
		// Create a child context for macro execution
		macroCtx := &RenderContext{
			env:     ctx.env,
			context: make(map[string]interface{}),
			blocks:  ctx.blocks,
			macros:  ctx.macros,
			engine:  ctx.engine,
			parent:  ctx,
		}

		// Map positional arguments to parameter names
		for i, param := range n.params {
			if i < len(args) {
				macroCtx.context[param] = args[i]
			} else if defaultExpr, hasDefault := n.defaults[param]; hasDefault {
				// Use default value if provided
				defaultVal, err := ctx.EvaluateExpression(defaultExpr)
				if err != nil {
					return err
				}
				macroCtx.context[param] = defaultVal
			}
		}

		// Render the macro body
		for _, node := range n.body {
			if err := node.Render(w, macroCtx); err != nil {
				return err
			}
		}

		return nil
	}

	// Register the macro in the context
	if ctx.macros == nil {
		ctx.macros = make(map[string]Node)
	}
	// Store the macro node directly
	ctx.macros[n.name] = n

	// Also store a function that can be called
	ctx.context[n.name] = func(args ...interface{}) func(io.Writer) error {
		return func(w io.Writer) error {
			return macro(w, args...)
		}
	}

	return nil
}

func (n *MacroNode) Type() NodeType {
	return NodeMacro
}

func (n *MacroNode) Line() int {
	return n.line
}

// CallMacro calls a macro with the given arguments
func (n *MacroNode) CallMacro(w io.Writer, ctx *RenderContext, args ...interface{}) error {
	// Create a child context for macro execution
	macroCtx := &RenderContext{
		env:     ctx.env,
		context: make(map[string]interface{}),
		blocks:  ctx.blocks,
		macros:  ctx.macros,
		engine:  ctx.engine,
		parent:  ctx,
	}

	// Map positional arguments to parameter names
	for i, param := range n.params {
		if i < len(args) {
			macroCtx.context[param] = args[i]
		} else if defaultExpr, hasDefault := n.defaults[param]; hasDefault {
			// Use default value if provided
			defaultVal, err := ctx.EvaluateExpression(defaultExpr)
			if err != nil {
				return err
			}
			macroCtx.context[param] = defaultVal
		}
	}

	// Render the macro body
	for _, node := range n.body {
		if err := node.Render(w, macroCtx); err != nil {
			return err
		}
	}

	return nil
}

// Implement Node interface for ImportNode
func (n *ImportNode) Render(w io.Writer, ctx *RenderContext) error {
	// Evaluate the template name
	templateName, err := ctx.EvaluateExpression(n.template)
	if err != nil {
		return err
	}

	// Get the template name as a string
	templateNameStr := ctx.ToString(templateName)

	// Check if we have an engine to load the imported template
	if ctx.engine == nil {
		return fmt.Errorf("cannot import without an engine")
	}

	// Load the imported template
	importedTemplate, err := ctx.engine.Load(templateNameStr)
	if err != nil {
		return err
	}

	// Create a context for the imported template
	importCtx := &RenderContext{
		env:     ctx.env,
		context: make(map[string]interface{}),
		blocks:  make(map[string][]Node),
		macros:  make(map[string]Node),
		engine:  ctx.engine,
	}

	// Render the imported template to a null writer to collect macros
	err = importedTemplate.nodes.Render(&NullWriter{}, importCtx)
	if err != nil {
		return err
	}

	// Create a module object with all the macros
	moduleObj := make(map[string]interface{})
	for name, macro := range importCtx.macros {
		if macroNode, ok := macro.(*MacroNode); ok {
			// Create a wrapper function that captures the macro node
			moduleObj[name] = func(args ...interface{}) func(io.Writer) error {
				return func(w io.Writer) error {
					return macroNode.CallMacro(w, ctx, args...)
				}
			}
		}
	}

	// Register the module in the context
	ctx.context[n.module] = moduleObj

	return nil
}

func (n *ImportNode) Type() NodeType {
	return NodeImport
}

func (n *ImportNode) Line() int {
	return n.line
}

// Implement Node interface for FromImportNode
func (n *FromImportNode) Render(w io.Writer, ctx *RenderContext) error {
	// Evaluate the template name
	templateName, err := ctx.EvaluateExpression(n.template)
	if err != nil {
		return err
	}

	// Get the template name as a string
	templateNameStr := ctx.ToString(templateName)

	// Check if we have an engine to load the imported template
	if ctx.engine == nil {
		return fmt.Errorf("cannot import without an engine")
	}

	// Load the imported template
	importedTemplate, err := ctx.engine.Load(templateNameStr)
	if err != nil {
		return err
	}

	// Create a context for the imported template
	importCtx := &RenderContext{
		env:     ctx.env,
		context: make(map[string]interface{}),
		blocks:  make(map[string][]Node),
		macros:  make(map[string]Node),
		engine:  ctx.engine,
	}

	// Render the imported template to a null writer to collect macros
	err = importedTemplate.nodes.Render(&NullWriter{}, importCtx)
	if err != nil {
		return err
	}

	// Import the requested macros
	for _, name := range n.macros {
		// Check if we need to use an alias
		alias := name
		if n.aliases != nil {
			if a, ok := n.aliases[name]; ok {
				alias = a
			}
		}

		// Get the macro from the imported template
		if macro, ok := importCtx.macros[name]; ok {
			if macroNode, ok := macro.(*MacroNode); ok {
				// Create a wrapper function that captures the macro node
				ctx.context[alias] = func(args ...interface{}) func(io.Writer) error {
					return func(w io.Writer) error {
						return macroNode.CallMacro(w, ctx, args...)
					}
				}
			}
		}
	}

	return nil
}

func (n *FromImportNode) Type() NodeType {
	return NodeImport
}

func (n *FromImportNode) Line() int {
	return n.line
}