package twig

import (
	"sync"
)

// This file extends the node pool system to cover all node types,
// following the implementation strategy in the zero-allocation plan.

// BlockNodePool provides a pool for BlockNode objects
var BlockNodePool = sync.Pool{
	New: func() interface{} {
		return &BlockNode{}
	},
}

// GetBlockNode gets a BlockNode from the pool and initializes it
func GetBlockNode(name string, body []Node, line int) *BlockNode {
	node := BlockNodePool.Get().(*BlockNode)
	node.name = name
	node.body = body
	node.line = line
	return node
}

// ReleaseBlockNode returns a BlockNode to the pool
func ReleaseBlockNode(node *BlockNode) {
	if node == nil {
		return
	}
	node.name = ""
	node.body = nil
	BlockNodePool.Put(node)
}

// ExtendsNodePool provides a pool for ExtendsNode objects
var ExtendsNodePool = sync.Pool{
	New: func() interface{} {
		return &ExtendsNode{}
	},
}

// GetExtendsNode gets an ExtendsNode from the pool and initializes it
func GetExtendsNode(parent Node, line int) *ExtendsNode {
	node := ExtendsNodePool.Get().(*ExtendsNode)
	node.parent = parent
	node.line = line
	return node
}

// ReleaseExtendsNode returns an ExtendsNode to the pool
func ReleaseExtendsNode(node *ExtendsNode) {
	if node == nil {
		return
	}
	node.parent = nil
	ExtendsNodePool.Put(node)
}

// IncludeNodePool provides a pool for IncludeNode objects
var IncludeNodePool = sync.Pool{
	New: func() interface{} {
		return &IncludeNode{}
	},
}

// GetIncludeNode gets an IncludeNode from the pool and initializes it
func GetIncludeNode(template Node, variables map[string]Node, ignoreMissing, only, sandboxed bool, line int) *IncludeNode {
	node := IncludeNodePool.Get().(*IncludeNode)
	node.template = template
	node.variables = variables
	node.ignoreMissing = ignoreMissing
	node.only = only
	node.sandboxed = sandboxed
	node.line = line
	return node
}

// ReleaseIncludeNode returns an IncludeNode to the pool
func ReleaseIncludeNode(node *IncludeNode) {
	if node == nil {
		return
	}
	node.template = nil
	node.variables = nil
	node.ignoreMissing = false
	node.only = false
	node.sandboxed = false
	IncludeNodePool.Put(node)
}

// SetNodePool provides a pool for SetNode objects
var SetNodePool = sync.Pool{
	New: func() interface{} {
		return &SetNode{}
	},
}

// GetSetNode gets a SetNode from the pool and initializes it
func GetSetNode(name string, value Node, line int) *SetNode {
	node := SetNodePool.Get().(*SetNode)
	node.name = name
	node.value = value
	node.line = line
	return node
}

// ReleaseSetNode returns a SetNode to the pool
func ReleaseSetNode(node *SetNode) {
	if node == nil {
		return
	}
	node.name = ""
	node.value = nil
	SetNodePool.Put(node)
}

// CommentNodePool provides a pool for CommentNode objects
var CommentNodePool = sync.Pool{
	New: func() interface{} {
		return &CommentNode{}
	},
}

// GetCommentNode gets a CommentNode from the pool and initializes it
func GetCommentNode(content string, line int) *CommentNode {
	node := CommentNodePool.Get().(*CommentNode)
	node.content = content
	node.line = line
	return node
}

// ReleaseCommentNode returns a CommentNode to the pool
func ReleaseCommentNode(node *CommentNode) {
	if node == nil {
		return
	}
	node.content = ""
	CommentNodePool.Put(node)
}

// MacroNodePool provides a pool for MacroNode objects
var MacroNodePool = sync.Pool{
	New: func() interface{} {
		return &MacroNode{}
	},
}

// GetMacroNode gets a MacroNode from the pool and initializes it
func GetMacroNode(name string, params []string, defaults map[string]Node, body []Node, line int) *MacroNode {
	node := MacroNodePool.Get().(*MacroNode)
	node.name = name
	node.params = params
	node.defaults = defaults
	node.body = body
	node.line = line
	return node
}

// ReleaseMacroNode returns a MacroNode to the pool
func ReleaseMacroNode(node *MacroNode) {
	if node == nil {
		return
	}
	node.name = ""
	node.params = nil
	node.defaults = nil
	node.body = nil
	MacroNodePool.Put(node)
}

// ImportNodePool provides a pool for ImportNode objects
var ImportNodePool = sync.Pool{
	New: func() interface{} {
		return &ImportNode{}
	},
}

// GetImportNode gets an ImportNode from the pool and initializes it
func GetImportNode(template Node, module string, line int) *ImportNode {
	node := ImportNodePool.Get().(*ImportNode)
	node.template = template
	node.module = module
	node.line = line
	return node
}

// ReleaseImportNode returns an ImportNode to the pool
func ReleaseImportNode(node *ImportNode) {
	if node == nil {
		return
	}
	node.template = nil
	node.module = ""
	ImportNodePool.Put(node)
}

// FromImportNodePool provides a pool for FromImportNode objects
var FromImportNodePool = sync.Pool{
	New: func() interface{} {
		return &FromImportNode{}
	},
}

// GetFromImportNode gets a FromImportNode from the pool and initializes it
func GetFromImportNode(template Node, macros []string, aliases map[string]string, line int) *FromImportNode {
	node := FromImportNodePool.Get().(*FromImportNode)
	node.template = template
	node.macros = macros
	node.aliases = aliases
	node.line = line
	return node
}

// ReleaseFromImportNode returns a FromImportNode to the pool
func ReleaseFromImportNode(node *FromImportNode) {
	if node == nil {
		return
	}
	node.template = nil
	node.macros = nil
	node.aliases = nil
	FromImportNodePool.Put(node)
}

// VerbatimNodePool provides a pool for VerbatimNode objects
var VerbatimNodePool = sync.Pool{
	New: func() interface{} {
		return &VerbatimNode{}
	},
}

// GetVerbatimNode gets a VerbatimNode from the pool and initializes it
func GetVerbatimNode(content string, line int) *VerbatimNode {
	node := VerbatimNodePool.Get().(*VerbatimNode)
	node.content = content
	node.line = line
	return node
}

// ReleaseVerbatimNode returns a VerbatimNode to the pool
func ReleaseVerbatimNode(node *VerbatimNode) {
	if node == nil {
		return
	}
	node.content = ""
	VerbatimNodePool.Put(node)
}

// DoNodePool provides a pool for DoNode objects
var DoNodePool = sync.Pool{
	New: func() interface{} {
		return &DoNode{}
	},
}

// GetDoNode gets a DoNode from the pool and initializes it
func GetDoNode(expression Node, line int) *DoNode {
	node := DoNodePool.Get().(*DoNode)
	node.expression = expression
	node.line = line
	return node
}

// ReleaseDoNode returns a DoNode to the pool
func ReleaseDoNode(node *DoNode) {
	if node == nil {
		return
	}
	node.expression = nil
	DoNodePool.Put(node)
}

// ApplyNodePool provides a pool for ApplyNode objects
var ApplyNodePool = sync.Pool{
	New: func() interface{} {
		return &ApplyNode{}
	},
}

// GetApplyNode gets an ApplyNode from the pool and initializes it
func GetApplyNode(body []Node, filter string, args []Node, line int) *ApplyNode {
	node := ApplyNodePool.Get().(*ApplyNode)
	node.body = body
	node.filter = filter
	node.args = args
	node.line = line
	return node
}

// ReleaseApplyNode returns an ApplyNode to the pool
func ReleaseApplyNode(node *ApplyNode) {
	if node == nil {
		return
	}
	node.body = nil
	node.filter = ""
	node.args = nil
	ApplyNodePool.Put(node)
}