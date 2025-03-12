package twig

import (
	"sync"
)

// This file implements object pooling for expression node types
// to reduce memory allocations during evaluation.

// BinaryNodePool provides a pool for BinaryNode objects
var BinaryNodePool = sync.Pool{
	New: func() interface{} {
		return &BinaryNode{}
	},
}

// GetBinaryNode gets a BinaryNode from the pool and initializes it
func GetBinaryNode(operator string, left, right Node, line int) *BinaryNode {
	node := BinaryNodePool.Get().(*BinaryNode)
	node.ExpressionNode.exprType = ExprBinary
	node.ExpressionNode.line = line
	node.operator = operator
	node.left = left
	node.right = right
	return node
}

// ReleaseBinaryNode returns a BinaryNode to the pool
func ReleaseBinaryNode(node *BinaryNode) {
	if node == nil {
		return
	}
	node.operator = ""
	node.left = nil
	node.right = nil
	BinaryNodePool.Put(node)
}

// GetAttrNodePool provides a pool for GetAttrNode objects
var GetAttrNodePool = sync.Pool{
	New: func() interface{} {
		return &GetAttrNode{}
	},
}

// GetGetAttrNode gets a GetAttrNode from the pool and initializes it
func GetGetAttrNode(node, attribute Node, line int) *GetAttrNode {
	n := GetAttrNodePool.Get().(*GetAttrNode)
	n.ExpressionNode.exprType = ExprGetAttr
	n.ExpressionNode.line = line
	n.node = node
	n.attribute = attribute
	return n
}

// ReleaseGetAttrNode returns a GetAttrNode to the pool
func ReleaseGetAttrNode(node *GetAttrNode) {
	if node == nil {
		return
	}
	node.node = nil
	node.attribute = nil
	GetAttrNodePool.Put(node)
}

// GetItemNodePool provides a pool for GetItemNode objects
var GetItemNodePool = sync.Pool{
	New: func() interface{} {
		return &GetItemNode{}
	},
}

// GetGetItemNode gets a GetItemNode from the pool and initializes it
func GetGetItemNode(node, item Node, line int) *GetItemNode {
	n := GetItemNodePool.Get().(*GetItemNode)
	n.ExpressionNode.exprType = ExprGetItem
	n.ExpressionNode.line = line
	n.node = node
	n.item = item
	return n
}

// ReleaseGetItemNode returns a GetItemNode to the pool
func ReleaseGetItemNode(node *GetItemNode) {
	if node == nil {
		return
	}
	node.node = nil
	node.item = nil
	GetItemNodePool.Put(node)
}

// FilterNodePool provides a pool for FilterNode objects
var FilterNodePool = sync.Pool{
	New: func() interface{} {
		return &FilterNode{}
	},
}

// GetFilterNode gets a FilterNode from the pool and initializes it
func GetFilterNode(node Node, filter string, args []Node, line int) *FilterNode {
	n := FilterNodePool.Get().(*FilterNode)
	n.ExpressionNode.exprType = ExprFilter
	n.ExpressionNode.line = line
	n.node = node
	n.filter = filter
	n.args = args
	return n
}

// ReleaseFilterNode returns a FilterNode to the pool
func ReleaseFilterNode(node *FilterNode) {
	if node == nil {
		return
	}
	node.node = nil
	node.filter = ""
	node.args = nil
	FilterNodePool.Put(node)
}

// TestNodePool provides a pool for TestNode objects
var TestNodePool = sync.Pool{
	New: func() interface{} {
		return &TestNode{}
	},
}

// GetTestNode gets a TestNode from the pool and initializes it
func GetTestNode(node Node, test string, args []Node, line int) *TestNode {
	n := TestNodePool.Get().(*TestNode)
	n.ExpressionNode.exprType = ExprTest
	n.ExpressionNode.line = line
	n.node = node
	n.test = test
	n.args = args
	return n
}

// ReleaseTestNode returns a TestNode to the pool
func ReleaseTestNode(node *TestNode) {
	if node == nil {
		return
	}
	node.node = nil
	node.test = ""
	node.args = nil
	TestNodePool.Put(node)
}

// UnaryNodePool provides a pool for UnaryNode objects
var UnaryNodePool = sync.Pool{
	New: func() interface{} {
		return &UnaryNode{}
	},
}

// GetUnaryNode gets a UnaryNode from the pool and initializes it
func GetUnaryNode(operator string, node Node, line int) *UnaryNode {
	n := UnaryNodePool.Get().(*UnaryNode)
	n.ExpressionNode.exprType = ExprUnary
	n.ExpressionNode.line = line
	n.operator = operator
	n.node = node
	return n
}

// ReleaseUnaryNode returns a UnaryNode to the pool
func ReleaseUnaryNode(node *UnaryNode) {
	if node == nil {
		return
	}
	node.operator = ""
	node.node = nil
	UnaryNodePool.Put(node)
}

// ConditionalNodePool provides a pool for ConditionalNode objects
var ConditionalNodePool = sync.Pool{
	New: func() interface{} {
		return &ConditionalNode{}
	},
}

// GetConditionalNode gets a ConditionalNode from the pool and initializes it
func GetConditionalNode(condition, trueExpr, falseExpr Node, line int) *ConditionalNode {
	node := ConditionalNodePool.Get().(*ConditionalNode)
	node.ExpressionNode.exprType = ExprConditional
	node.ExpressionNode.line = line
	node.condition = condition
	node.trueExpr = trueExpr
	node.falseExpr = falseExpr
	return node
}

// ReleaseConditionalNode returns a ConditionalNode to the pool
func ReleaseConditionalNode(node *ConditionalNode) {
	if node == nil {
		return
	}
	node.condition = nil
	node.trueExpr = nil
	node.falseExpr = nil
	ConditionalNodePool.Put(node)
}

// ArrayNodePool provides a pool for ArrayNode objects
var ArrayNodePool = sync.Pool{
	New: func() interface{} {
		return &ArrayNode{}
	},
}

// GetArrayNode gets an ArrayNode from the pool and initializes it
func GetArrayNode(items []Node, line int) *ArrayNode {
	node := ArrayNodePool.Get().(*ArrayNode)
	node.ExpressionNode.exprType = ExprArray
	node.ExpressionNode.line = line
	node.items = items
	return node
}

// ReleaseArrayNode returns an ArrayNode to the pool
func ReleaseArrayNode(node *ArrayNode) {
	if node == nil {
		return
	}
	node.items = nil
	ArrayNodePool.Put(node)
}

// HashNodePool provides a pool for HashNode objects
var HashNodePool = sync.Pool{
	New: func() interface{} {
		return &HashNode{}
	},
}

// GetHashNode gets a HashNode from the pool and initializes it
func GetHashNode(items map[Node]Node, line int) *HashNode {
	node := HashNodePool.Get().(*HashNode)
	node.ExpressionNode.exprType = ExprHash
	node.ExpressionNode.line = line
	node.items = items
	return node
}

// ReleaseHashNode returns a HashNode to the pool
func ReleaseHashNode(node *HashNode) {
	if node == nil {
		return
	}
	node.items = nil
	HashNodePool.Put(node)
}

// FunctionNodePool provides a pool for FunctionNode objects
var FunctionNodePool = sync.Pool{
	New: func() interface{} {
		return &FunctionNode{}
	},
}

// GetFunctionNode gets a FunctionNode from the pool and initializes it
func GetFunctionNode(name string, args []Node, line int) *FunctionNode {
	node := FunctionNodePool.Get().(*FunctionNode)
	node.ExpressionNode.exprType = ExprFunction
	node.ExpressionNode.line = line
	node.name = name
	node.args = args
	return node
}

// ReleaseFunctionNode returns a FunctionNode to the pool
func ReleaseFunctionNode(node *FunctionNode) {
	if node == nil {
		return
	}
	node.name = ""
	node.args = nil
	node.moduleExpr = nil
	FunctionNodePool.Put(node)
}

// VariableNodePool provides a pool for VariableNode objects
var VariableNodePool = sync.Pool{
	New: func() interface{} {
		return &VariableNode{}
	},
}

// GetVariableNode gets a VariableNode from the pool and initializes it
func GetVariableNode(name string, line int) *VariableNode {
	node := VariableNodePool.Get().(*VariableNode)
	node.ExpressionNode.exprType = ExprVariable
	node.ExpressionNode.line = line
	node.name = name
	return node
}

// ReleaseVariableNode returns a VariableNode to the pool
func ReleaseVariableNode(node *VariableNode) {
	if node == nil {
		return
	}
	node.name = ""
	VariableNodePool.Put(node)
}

// LiteralNodePool provides a pool for LiteralNode objects
var LiteralNodePool = sync.Pool{
	New: func() interface{} {
		return &LiteralNode{}
	},
}

// GetLiteralNode gets a LiteralNode from the pool and initializes it
func GetLiteralNode(value interface{}, line int) *LiteralNode {
	node := LiteralNodePool.Get().(*LiteralNode)
	node.ExpressionNode.exprType = ExprLiteral
	node.ExpressionNode.line = line
	node.value = value
	return node
}

// ReleaseLiteralNode returns a LiteralNode to the pool
func ReleaseLiteralNode(node *LiteralNode) {
	if node == nil {
		return
	}
	node.value = nil
	LiteralNodePool.Put(node)
}

// --- Slice Pools for Expression Evaluation ---

// smallArgSlicePool provides a pool for small argument slices (0-2 items)
var smallArgSlicePool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate a slice with capacity 2
		return make([]interface{}, 0, 2)
	},
}

// mediumArgSlicePool provides a pool for medium argument slices (3-5 items)
var mediumArgSlicePool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate a slice with capacity 5
		return make([]interface{}, 0, 5)
	},
}

// largeArgSlicePool provides a pool for large argument slices (6-10 items)
var largeArgSlicePool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate a slice with capacity 10
		return make([]interface{}, 0, 10)
	},
}

// GetArgSlice gets an appropriately sized slice for the given number of arguments
func GetArgSlice(size int) []interface{} {
	if size <= 0 {
		return nil
	}

	var slice []interface{}

	switch {
	case size <= 2:
		slice = smallArgSlicePool.Get().([]interface{})
	case size <= 5:
		slice = mediumArgSlicePool.Get().([]interface{})
	case size <= 10:
		slice = largeArgSlicePool.Get().([]interface{})
	default:
		// For very large slices, just allocate directly
		return make([]interface{}, 0, size)
	}

	// Clear the slice but maintain capacity
	return slice[:0]
}

// ReleaseArgSlice returns an argument slice to the appropriate pool
func ReleaseArgSlice(slice []interface{}) {
	if slice == nil {
		return
	}

	// Clear all references
	for i := range slice {
		slice[i] = nil
	}

	// Reset length to 0
	slice = slice[:0]

	// Return to appropriate pool based on capacity
	switch cap(slice) {
	case 2:
		smallArgSlicePool.Put(slice)
	case 5:
		mediumArgSlicePool.Put(slice)
	case 10:
		largeArgSlicePool.Put(slice)
	}
}

// --- Map Pools for HashNode Evaluation ---

// smallHashMapPool provides a pool for small hash maps (1-5 items)
var smallHashMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 5)
	},
}

// mediumHashMapPool provides a pool for medium hash maps (6-15 items)
var mediumHashMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 15)
	},
}

// GetHashMap gets an appropriately sized map for hash operations
func GetHashMap(size int) map[string]interface{} {
	if size <= 5 {
		hashMap := smallHashMapPool.Get().(map[string]interface{})
		// Clear any existing entries
		for k := range hashMap {
			delete(hashMap, k)
		}
		return hashMap
	} else if size <= 15 {
		hashMap := mediumHashMapPool.Get().(map[string]interface{})
		// Clear any existing entries
		for k := range hashMap {
			delete(hashMap, k)
		}
		return hashMap
	}

	// For larger maps, just allocate directly
	return make(map[string]interface{}, size)
}

// ReleaseHashMap returns a hash map to the appropriate pool
func ReleaseHashMap(hashMap map[string]interface{}) {
	if hashMap == nil {
		return
	}

	// Clear all entries (not used directly in our defer block)
	// for k := range hashMap {
	//     delete(hashMap, k)
	// }

	// Return to appropriate pool based on capacity
	// We don't actually clear the map when releasing through the defer,
	// because we return the map as the result and deleting entries would
	// clear the returned result

	// Map doesn't have a built-in cap function
	// Not using pool return for maps directly returned as results

	/*
		switch {
		case len(hashMap) <= 5:
			smallHashMapPool.Put(hashMap)
		case len(hashMap) <= 15:
			mediumHashMapPool.Put(hashMap)
		}
	*/
}
