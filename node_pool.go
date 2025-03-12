package twig

import (
	"sync"
)

// NodePool provides object pooling for all node types
// This significantly reduces GC pressure by reusing node objects

// TextNodePool provides a pool for TextNode objects
var TextNodePool = sync.Pool{
	New: func() interface{} {
		return &TextNode{}
	},
}

// GetTextNode gets a TextNode from the pool and initializes it
func GetTextNode(content string, line int) *TextNode {
	node := TextNodePool.Get().(*TextNode)
	node.content = content
	node.line = line
	return node
}

// ReleaseTextNode returns a TextNode to the pool
func ReleaseTextNode(node *TextNode) {
	if node == nil {
		return
	}
	node.content = ""
	TextNodePool.Put(node)
}

// PrintNodePool provides a pool for PrintNode objects
var PrintNodePool = sync.Pool{
	New: func() interface{} {
		return &PrintNode{}
	},
}

// GetPrintNode gets a PrintNode from the pool and initializes it
func GetPrintNode(expression Node, line int) *PrintNode {
	node := PrintNodePool.Get().(*PrintNode)
	node.expression = expression
	node.line = line
	return node
}

// ReleasePrintNode returns a PrintNode to the pool
func ReleasePrintNode(node *PrintNode) {
	if node == nil {
		return
	}
	node.expression = nil
	PrintNodePool.Put(node)
}

// RootNodePool provides a pool for RootNode objects
var RootNodePool = sync.Pool{
	New: func() interface{} {
		return &RootNode{}
	},
}

// GetRootNode gets a RootNode from the pool and initializes it
func GetRootNode(children []Node, line int) *RootNode {
	node := RootNodePool.Get().(*RootNode)
	node.children = children
	node.line = line
	return node
}

// ReleaseRootNode returns a RootNode to the pool
func ReleaseRootNode(node *RootNode) {
	if node == nil {
		return
	}
	node.children = nil
	RootNodePool.Put(node)
}

// Note: LiteralNodePool, GetLiteralNode, ReleaseLiteralNode moved to expr_pool.go
// Note: VariableNodePool, GetVariableNode, ReleaseVariableNode moved to expr_pool.go

// TokenPool provides a pool for Token objects
var TokenPool = sync.Pool{
	New: func() interface{} {
		return &Token{}
	},
}

// GetToken gets a Token from the pool and initializes it
func GetToken(tokenType int, value string, line int) *Token {
	token := TokenPool.Get().(*Token)
	token.Type = tokenType
	token.Value = value
	token.Line = line
	return token
}

// ReleaseToken returns a Token to the pool
func ReleaseToken(token *Token) {
	if token == nil {
		return
	}
	token.Value = ""
	TokenPool.Put(token)
}

// SlicePool provides pools for commonly used slice types
// to reduce allocations when working with collections

// IfNodePool provides a pool for IfNode objects
var IfNodePool = sync.Pool{
	New: func() interface{} {
		return &IfNode{}
	},
}

// GetIfNode gets an IfNode from the pool and initializes it
func GetIfNode(conditions []Node, bodies [][]Node, elseBranch []Node, line int) *IfNode {
	node := IfNodePool.Get().(*IfNode)
	node.conditions = conditions
	node.bodies = bodies
	node.elseBranch = elseBranch
	node.line = line
	return node
}

// ReleaseIfNode returns an IfNode to the pool
func ReleaseIfNode(node *IfNode) {
	if node == nil {
		return
	}
	node.conditions = nil
	node.bodies = nil
	node.elseBranch = nil
	IfNodePool.Put(node)
}

// ForNodePool provides a pool for ForNode objects
var ForNodePool = sync.Pool{
	New: func() interface{} {
		return &ForNode{}
	},
}

// GetForNode gets a ForNode from the pool and initializes it
func GetForNode(keyVar, valueVar string, sequence Node, body, elseBranch []Node, line int) *ForNode {
	node := ForNodePool.Get().(*ForNode)
	node.keyVar = keyVar
	node.valueVar = valueVar
	node.sequence = sequence
	node.body = body
	node.elseBranch = elseBranch
	node.line = line
	return node
}

// ReleaseForNode returns a ForNode to the pool
func ReleaseForNode(node *ForNode) {
	if node == nil {
		return
	}
	node.keyVar = ""
	node.valueVar = ""
	node.sequence = nil
	node.body = nil
	node.elseBranch = nil
	ForNodePool.Put(node)
}

// NodeSlicePool provides a pool for []Node slices
var NodeSlicePool = sync.Pool{
	New: func() interface{} {
		slice := make([]Node, 0, 8) // Default capacity of 8 for common cases
		return &slice
	},
}

// GetNodeSlice gets a slice of Node from the pool
func GetNodeSlice() *[]Node {
	slice := NodeSlicePool.Get().(*[]Node)
	*slice = (*slice)[:0] // Clear slice but keep capacity
	return slice
}

// ReleaseNodeSlice returns a slice of Node to the pool
func ReleaseNodeSlice(slice *[]Node) {
	if slice == nil {
		return
	}
	// Clear references to help GC
	for i := range *slice {
		(*slice)[i] = nil
	}
	*slice = (*slice)[:0]
	NodeSlicePool.Put(slice)
}

// TokenSlicePool provides a pool for []Token slices
var TokenSlicePool = sync.Pool{
	New: func() interface{} {
		slice := make([]Token, 0, 32) // Higher default capacity for tokens
		return &slice
	},
}

// GetTokenSlice gets a slice of Token from the pool with optional capacity hint
func GetTokenSlice(capacityHint int) []Token {
	if capacityHint <= 0 {
		// Use default capacity
		slice := TokenSlicePool.Get().(*[]Token)
		return (*slice)[:0]
	}

	// For large token slices, allocate directly
	if capacityHint > 1000 {
		return make([]Token, 0, capacityHint)
	}

	// Get from pool and ensure it has enough capacity
	slice := TokenSlicePool.Get().(*[]Token)
	if cap(*slice) < capacityHint {
		// Current slice is too small, allocate a new one
		*slice = make([]Token, 0, capacityHint)
	} else {
		*slice = (*slice)[:0] // Clear but keep capacity
	}
	return *slice
}

// ReleaseTokenSlice returns a slice of Token to the pool
func ReleaseTokenSlice(slice []Token) {
	// Only return reasonable sized slices to the pool
	if cap(slice) > 1000 || cap(slice) < 32 {
		return // Don't pool very large or very small slices
	}

	// Clear the slice to help GC
	slicePtr := &slice
	*slicePtr = (*slicePtr)[:0]
	TokenSlicePool.Put(slicePtr)
}
