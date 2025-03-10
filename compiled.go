package twig

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

// CompiledTemplate represents a compiled Twig template
type CompiledTemplate struct {
	Name         string // Template name
	Source       string // Original template source
	LastModified int64  // Last modification timestamp
	CompileTime  int64  // Time when compilation occurred
}

// CompileTemplate compiles a parsed template into a compiled format
func CompileTemplate(tmpl *Template) (*CompiledTemplate, error) {
	if tmpl == nil {
		return nil, fmt.Errorf("cannot compile nil template")
	}

	// Store the template source and metadata
	compiled := &CompiledTemplate{
		Name:         tmpl.name,
		Source:       tmpl.source,
		LastModified: tmpl.lastModified,
		CompileTime:  time.Now().Unix(),
	}

	return compiled, nil
}

// LoadFromCompiled loads a template from its compiled representation
func LoadFromCompiled(compiled *CompiledTemplate, env *Environment, engine *Engine) (*Template, error) {
	if compiled == nil {
		return nil, fmt.Errorf("cannot load from nil compiled template")
	}

	// Parse the template source
	parser := &Parser{}
	nodes, err := parser.Parse(compiled.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse compiled template: %w", err)
	}

	// Create the template from the parsed nodes
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
