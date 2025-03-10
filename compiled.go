package twig

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

func init() {
	// Register all node types for serialization
	gob.Register(&RootNode{})
	gob.Register(&TextNode{})
	gob.Register(&PrintNode{})
	gob.Register(&BlockNode{})
	gob.Register(&ForNode{})
	gob.Register(&IfNode{})
	gob.Register(&SetNode{})
	gob.Register(&IncludeNode{})
	gob.Register(&ExtendsNode{})
	gob.Register(&MacroNode{})
	gob.Register(&ImportNode{})
	gob.Register(&FromImportNode{})
	gob.Register(&FunctionNode{})
	gob.Register(&CommentNode{})
	
	// Expression nodes
	gob.Register(&ExpressionNode{})
	gob.Register(&VariableNode{})
	gob.Register(&LiteralNode{})
	gob.Register(&UnaryNode{})
	gob.Register(&BinaryNode{})
}

// CompiledTemplate represents a compiled Twig template
type CompiledTemplate struct {
	Name         string // Template name
	Source       string // Original template source
	LastModified int64  // Last modification timestamp
	CompileTime  int64  // Time when compilation occurred
	AST          []byte // Serialized AST data
}

// CompileTemplate compiles a parsed template into a compiled format
func CompileTemplate(tmpl *Template) (*CompiledTemplate, error) {
	if tmpl == nil {
		return nil, fmt.Errorf("cannot compile nil template")
	}

	// Serialize the AST (Node tree)
	var astBuf bytes.Buffer
	enc := gob.NewEncoder(&astBuf)
	
	// Encode the AST
	if err := enc.Encode(tmpl.nodes); err != nil {
		// If serialization fails, we'll still create the template but without AST
		// and log the error
		fmt.Printf("Warning: Failed to serialize AST: %v\n", err)
	}

	// Store the template source, metadata, and AST
	compiled := &CompiledTemplate{
		Name:         tmpl.name,
		Source:       tmpl.source,
		LastModified: tmpl.lastModified,
		CompileTime:  time.Now().Unix(),
		AST:          astBuf.Bytes(),
	}

	return compiled, nil
}

// LoadFromCompiled loads a template from its compiled representation
func LoadFromCompiled(compiled *CompiledTemplate, env *Environment, engine *Engine) (*Template, error) {
	if compiled == nil {
		return nil, fmt.Errorf("cannot load from nil compiled template")
	}

	var nodes Node
	
	// Try to use the cached AST if available
	if len(compiled.AST) > 0 {
		// Attempt to deserialize the AST
		dec := gob.NewDecoder(bytes.NewReader(compiled.AST))
		
		// Try to decode the AST
		err := dec.Decode(&nodes)
		if err != nil {
			// Fall back to parsing if AST deserialization fails
			fmt.Printf("Warning: Failed to deserialize AST, falling back to parsing: %v\n", err)
			nodes = nil
		}
	}
	
	// If AST deserialization failed or AST is not available, parse the source
	if nodes == nil {
		parser := &Parser{}
		var err error
		nodes, err = parser.Parse(compiled.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse compiled template: %w", err)
		}
	}

	// Create the template from the nodes (either from AST or freshly parsed)
	tmpl := &Template{
		name:         compiled.Name,
		source:       compiled.Source,
		nodes:        nodes,
		env:          env,
		engine:       engine,
		lastModified: compiled.LastModified,
	}

	return tmpl, nil
}

// SerializeCompiledTemplate serializes a compiled template to a byte array
func SerializeCompiledTemplate(compiled *CompiledTemplate) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(compiled); err != nil {
		return nil, fmt.Errorf("failed to serialize compiled template: %w", err)
	}

	return buf.Bytes(), nil
}

// DeserializeCompiledTemplate deserializes a compiled template from a byte array
func DeserializeCompiledTemplate(data []byte) (*CompiledTemplate, error) {
	dec := gob.NewDecoder(bytes.NewReader(data))

	var compiled CompiledTemplate
	if err := dec.Decode(&compiled); err != nil {
		return nil, fmt.Errorf("failed to deserialize compiled template: %w", err)
	}

	return &compiled, nil
}
