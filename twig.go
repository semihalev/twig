package twig

import (
	"bytes"
	"io"
	"sync"
)

// Engine represents the Twig template engine
type Engine struct {
	templates   map[string]*Template
	mu          sync.RWMutex
	autoReload  bool
	strictVars  bool
	loaders     []Loader
	environment *Environment
	
	// Test helper - override Parse function
	Parse func(source string) (*Template, error)
}

// Template represents a parsed and compiled Twig template
type Template struct {
	name   string
	source string
	nodes  Node
	env    *Environment
}

// Environment holds configuration and context for template rendering
type Environment struct {
	globals     map[string]interface{}
	filters     map[string]FilterFunc
	functions   map[string]FunctionFunc
	tests       map[string]TestFunc
	operators   map[string]OperatorFunc
	extensions  []Extension
	cache       bool
	autoescape bool
	debug      bool
	sandbox    bool
}

// New creates a new Twig engine instance
func New() *Engine {
	env := &Environment{
		globals:     make(map[string]interface{}),
		filters:     make(map[string]FilterFunc),
		functions:   make(map[string]FunctionFunc),
		tests:       make(map[string]TestFunc),
		operators:   make(map[string]OperatorFunc),
		autoescape: true,
	}

	return &Engine{
		templates:   make(map[string]*Template),
		environment: env,
	}
}

// RegisterLoader adds a template loader to the engine
func (e *Engine) RegisterLoader(loader Loader) {
	e.loaders = append(e.loaders, loader)
}

// SetAutoReload sets whether templates should be reloaded on change
func (e *Engine) SetAutoReload(autoReload bool) {
	e.autoReload = autoReload
}

// SetStrictVars sets whether strict variable access is enabled
func (e *Engine) SetStrictVars(strictVars bool) {
	e.strictVars = strictVars
}

// Render renders a template with the given context
func (e *Engine) Render(name string, context map[string]interface{}) (string, error) {
	template, err := e.Load(name)
	if err != nil {
		return "", err
	}

	return template.Render(context)
}

// RenderTo renders a template to a writer
func (e *Engine) RenderTo(w io.Writer, name string, context map[string]interface{}) error {
	template, err := e.Load(name)
	if err != nil {
		return err
	}

	return template.RenderTo(w, context)
}

// Load loads a template by name
func (e *Engine) Load(name string) (*Template, error) {
	e.mu.RLock()
	if tmpl, ok := e.templates[name]; ok {
		e.mu.RUnlock()
		return tmpl, nil
	}
	e.mu.RUnlock()

	for _, loader := range e.loaders {
		source, err := loader.Load(name)
		if err != nil {
			continue
		}

		parser := &Parser{}
		nodes, err := parser.Parse(source)
		if err != nil {
			return nil, err
		}

		template := &Template{
			name:   name,
			source: source,
			nodes:  nodes,
			env:    e.environment,
		}

		e.mu.Lock()
		e.templates[name] = template
		e.mu.Unlock()

		return template, nil
	}

	return nil, ErrTemplateNotFound
}

// ParseTemplate parses a template string
func (e *Engine) ParseTemplate(source string) (*Template, error) {
	// Use the override Parse function if it's set (for testing)
	if e.Parse != nil {
		return e.Parse(source)
	}
	
	parser := &Parser{}
	nodes, err := parser.Parse(source)
	if err != nil {
		return nil, err
	}

	template := &Template{
		source: source,
		nodes:  nodes,
		env:    e.environment,
	}

	return template, nil
}

// Render renders a template with the given context
func (t *Template) Render(context map[string]interface{}) (string, error) {
	var buf StringBuffer
	err := t.RenderTo(&buf, context)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderTo renders a template to a writer
func (t *Template) RenderTo(w io.Writer, context map[string]interface{}) error {
	if context == nil {
		context = make(map[string]interface{})
	}
	
	// Create a simple render context 
	ctx := &RenderContext{
		env:     t.env,
		context: context,
		blocks:  make(map[string][]Node),
		macros:  make(map[string]Node),
	}
	
	return t.nodes.Render(w, ctx)
}

// StringBuffer is a simple buffer for string building
type StringBuffer struct {
	buf bytes.Buffer
}

// Write implements io.Writer
func (b *StringBuffer) Write(p []byte) (n int, err error) {
	return b.buf.Write(p)
}

// String returns the buffer's contents as a string
func (b *StringBuffer) String() string {
	return b.buf.String()
}