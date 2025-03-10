//go:generate go run tools/lexgen/main.go -output gen
//go:generate go run tools/parsegen/main.go -output gen
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
	engine *Engine       // Reference back to the engine for loading parent templates
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

	engine := &Engine{
		templates:   make(map[string]*Template),
		environment: env,
	}
	
	// Register the core extension by default
	engine.AddExtension(&CoreExtension{})
	
	return engine
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
			engine: e,  // Add reference to the engine
		}

		e.mu.Lock()
		e.templates[name] = template
		e.mu.Unlock()

		return template, nil
	}

	return nil, ErrTemplateNotFound
}

// RegisterString registers a template from a string source
func (e *Engine) RegisterString(name string, source string) error {
	parser := &Parser{}
	nodes, err := parser.Parse(source)
	if err != nil {
		return err
	}

	template := &Template{
		name:   name,
		source: source,
		nodes:  nodes,
		env:    e.environment,
		engine: e,
	}

	e.mu.Lock()
	e.templates[name] = template
	e.mu.Unlock()

	return nil
}

// GetEnvironment returns the engine's environment
func (e *Engine) GetEnvironment() *Environment {
	return e.environment
}

// RegisterTemplate directly registers a pre-built template
func (e *Engine) RegisterTemplate(name string, template *Template) {
	e.mu.Lock()
	e.templates[name] = template
	e.mu.Unlock()
}

// AddFilter registers a custom filter function
func (e *Engine) AddFilter(name string, filter FilterFunc) {
	e.environment.filters[name] = filter
}

// AddFunction registers a custom function
func (e *Engine) AddFunction(name string, function FunctionFunc) {
	e.environment.functions[name] = function
}

// AddTest registers a custom test function
func (e *Engine) AddTest(name string, test TestFunc) {
	e.environment.tests[name] = test
}

// AddGlobal adds a global variable to the template environment
func (e *Engine) AddGlobal(name string, value interface{}) {
	e.environment.globals[name] = value
}

// AddExtension registers a Twig extension
func (e *Engine) AddExtension(extension Extension) {
	e.environment.extensions = append(e.environment.extensions, extension)
	
	// Register all filters from the extension
	for name, filter := range extension.GetFilters() {
		e.environment.filters[name] = filter
	}
	
	// Register all functions from the extension
	for name, function := range extension.GetFunctions() {
		e.environment.functions[name] = function
	}
	
	// Register all tests from the extension
	for name, test := range extension.GetTests() {
		e.environment.tests[name] = test
	}
	
	// Register all operators from the extension
	for name, operator := range extension.GetOperators() {
		e.environment.operators[name] = operator
	}
	
	// Initialize the extension
	extension.Initialize(e)
}

// CreateExtension creates a new custom extension with the given name
func (e *Engine) CreateExtension(name string) *CustomExtension {
	extension := &CustomExtension{
		Name:      name,
		Filters:   make(map[string]FilterFunc),
		Functions: make(map[string]FunctionFunc),
		Tests:     make(map[string]TestFunc),
		Operators: make(map[string]OperatorFunc),
	}
	
	return extension
}

// AddFilterToExtension adds a filter to a custom extension
func (e *Engine) AddFilterToExtension(extension *CustomExtension, name string, filter FilterFunc) {
	if extension.Filters == nil {
		extension.Filters = make(map[string]FilterFunc)
	}
	extension.Filters[name] = filter
}

// AddFunctionToExtension adds a function to a custom extension
func (e *Engine) AddFunctionToExtension(extension *CustomExtension, name string, function FunctionFunc) {
	if extension.Functions == nil {
		extension.Functions = make(map[string]FunctionFunc)
	}
	extension.Functions[name] = function
}

// AddTestToExtension adds a test to a custom extension
func (e *Engine) AddTestToExtension(extension *CustomExtension, name string, test TestFunc) {
	if extension.Tests == nil {
		extension.Tests = make(map[string]TestFunc)
	}
	extension.Tests[name] = test
}

// RegisterExtension creates, configures and registers a custom extension
func (e *Engine) RegisterExtension(name string, config func(*CustomExtension)) {
	extension := e.CreateExtension(name)
	if config != nil {
		config(extension)
	}
	e.AddExtension(extension)
}

// NewTemplate creates a new template with the given parameters
func (e *Engine) NewTemplate(name string, source string, nodes Node) *Template {
	return &Template{
		name:   name,
		source: source,
		nodes:  nodes,
		env:    e.environment,
		engine: e,
	}
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
		engine: e,
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
	
	// Create a render context with access to the engine
	ctx := &RenderContext{
		env:          t.env,
		context:      context,
		blocks:       make(map[string][]Node),
		macros:       make(map[string]Node),
		engine:       t.engine,
		extending:    false,
		currentBlock: nil,
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