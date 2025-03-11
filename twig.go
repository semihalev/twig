package twig

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Engine represents the Twig template engine
type Engine struct {
	templates       map[string]*Template
	mu              sync.RWMutex
	autoReload      bool
	strictVars      bool
	loaders         []Loader
	environment     *Environment
	debug           bool
	currentTemplate string // Tracks the name of the template currently being rendered

	// Test helper - override Parse function
	Parse func(source string) (*Template, error)
}

// Template represents a parsed and compiled Twig template
type Template struct {
	name         string
	source       string
	nodes        Node
	env          *Environment
	engine       *Engine // Reference back to the engine for loading parent templates
	loader       Loader  // The loader that loaded this template
	lastModified int64   // Last modified timestamp for this template
}

// Environment holds configuration and context for template rendering
type Environment struct {
	globals    map[string]interface{}
	filters    map[string]FilterFunc
	functions  map[string]FunctionFunc
	tests      map[string]TestFunc
	operators  map[string]OperatorFunc
	extensions []Extension
	cache      bool
	autoescape bool
	debug      bool
	sandbox    bool
}

// New creates a new Twig engine instance
func New() *Engine {
	env := &Environment{
		globals:    make(map[string]interface{}),
		filters:    make(map[string]FilterFunc),
		functions:  make(map[string]FunctionFunc),
		tests:      make(map[string]TestFunc),
		operators:  make(map[string]OperatorFunc),
		autoescape: true,
		cache:      true,  // Enable caching by default
		debug:      false, // Disable debug mode by default
	}

	engine := &Engine{
		templates:   make(map[string]*Template),
		environment: env,
		autoReload:  false, // Disable auto-reload by default
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

// SetDebug enables or disables debug mode
func (e *Engine) SetDebug(enabled bool) {
	e.environment.debug = enabled
	e.debug = enabled

	// Set the debug level based on the debug flag
	if enabled {
		SetDebugLevel(DebugInfo)
	} else {
		SetDebugLevel(DebugOff)
	}
}

// SetCache enables or disables template caching
func (e *Engine) SetCache(enabled bool) {
	e.environment.cache = enabled
}

// SetDevelopmentMode enables settings appropriate for development
// This sets debug mode on, enables auto-reload, and disables caching
func (e *Engine) SetDevelopmentMode(enabled bool) {
	e.environment.debug = enabled
	e.debug = enabled
	e.autoReload = enabled
	e.environment.cache = !enabled
}

// Render renders a template with the given context
func (e *Engine) Render(name string, context map[string]interface{}) (string, error) {
	LogInfo("Rendering template: %s", name)

	// Store current template name and previous template name
	prevTemplate := e.currentTemplate
	e.currentTemplate = name
	defer func() {
		// Restore previous template name when we're done
		e.currentTemplate = prevTemplate
	}()

	template, err := e.Load(name)
	if err != nil {
		LogError(err, fmt.Sprintf("Failed to load template: %s", name))
		return "", err
	}

	// If debug is enabled, use more detailed error reporting
	if e.environment.debug {
		var buf StringBuffer
		ctx := NewRenderContext(e.environment, context, e)
		defer ctx.Release()

		// Use debug rendering with enhanced error reporting
		err = DebugRender(&buf, template, ctx)
		if err != nil {
			LogError(err, fmt.Sprintf("Error rendering template: %s", name))
			// Enhance error with template information
			if enhancedErr, ok := err.(*EnhancedError); !ok {
				err = NewError(err, name, 0, 0, template.source)
			} else {
				// Already enhanced, just ensure template name is set
				if enhancedErr.Template == "" {
					enhancedErr.Template = name
				}
			}
			return "", err
		}

		return buf.String(), nil
	}

	// Normal rendering path without debug overhead
	return template.Render(context)
}

// RenderTo renders a template to a writer
func (e *Engine) RenderTo(w io.Writer, name string, context map[string]interface{}) error {
	LogInfo("Rendering template to writer: %s", name)

	// Store current template name and previous template name
	prevTemplate := e.currentTemplate
	e.currentTemplate = name
	defer func() {
		// Restore previous template name when we're done
		e.currentTemplate = prevTemplate
	}()

	template, err := e.Load(name)
	if err != nil {
		LogError(err, fmt.Sprintf("Failed to load template: %s", name))
		return err
	}

	// If debug is enabled, use more detailed error reporting
	if e.environment.debug {
		ctx := NewRenderContext(e.environment, context, e)
		defer ctx.Release()

		// Use debug rendering with enhanced error reporting
		err = DebugRender(w, template, ctx)
		if err != nil {
			LogError(err, fmt.Sprintf("Error rendering template: %s", name))
			// Enhance error with template information
			if enhancedErr, ok := err.(*EnhancedError); !ok {
				err = NewError(err, name, 0, 0, template.source)
			} else {
				// Already enhanced, just ensure template name is set
				if enhancedErr.Template == "" {
					enhancedErr.Template = name
				}
			}
			return err
		}

		return nil
	}

	// Normal rendering path without debug overhead
	return template.RenderTo(w, context)
}

// Load loads a template by name
func (e *Engine) Load(name string) (*Template, error) {
	// Only check the cache if caching is enabled
	if e.environment.cache {
		// Use a quick check under read lock first to avoid contention
		e.mu.RLock()
		tmpl, ok := e.templates[name]
		e.mu.RUnlock()

		// If template exists in cache
		if ok {
			// If auto-reload is disabled, return the cached template immediately
			if !e.autoReload {
				return tmpl, nil
			}

			// If auto-reload is enabled, check if the template has been modified
			needsReload := false

			if tmpl.loader != nil {
				// Check if the loader supports timestamp checking
				if tsLoader, ok := tmpl.loader.(TimestampAwareLoader); ok {
					// Get the current modification time
					currentModTime, err := tsLoader.GetModifiedTime(name)
					if err != nil || currentModTime > tmpl.lastModified {
						needsReload = true
					}
				}
			}

			// If no reload needed, return cached template
			if !needsReload {
				return tmpl, nil
			}
		}
	}

	// Template not in cache or cache disabled or needs reloading
	// Try to load the template
	var lastModified int64
	var sourceLoader Loader
	var loaderErrors []error
	var template *Template

	for _, loader := range e.loaders {
		source, err := loader.Load(name)
		if err != nil {
			// Collect loader errors for better diagnostics
			loaderErrors = append(loaderErrors, fmt.Errorf("loader %T: %w", loader, err))
			continue
		}

		// If this loader supports modification times, get the time
		if tsLoader, ok := loader.(TimestampAwareLoader); ok {
			lastModified, _ = tsLoader.GetModifiedTime(name)
		}

		sourceLoader = loader
		LogInfo("Template '%s' loaded from %T", name, loader)

		parser := &Parser{}
		nodes, err := parser.Parse(source)
		if err != nil {
			// Include more context in parsing errors
			return nil, NewError(err, name, 0, 0, source)
		}

		template = &Template{
			name:         name,
			source:       source,
			nodes:        nodes,
			env:          e.environment,
			engine:       e, // Add reference to the engine
			loader:       sourceLoader,
			lastModified: lastModified,
		}

		// Successfully loaded template
		break
	}

	// If we failed to load the template from any loader
	if template == nil {
		// If we have collected errors from loaders, include them in the error message
		if len(loaderErrors) > 0 {
			errorDetails := strings.Builder{}
			errorDetails.WriteString(fmt.Sprintf("Template '%s' not found. Tried %d loaders:\n", name, len(loaderErrors)))

			for i, err := range loaderErrors {
				errorDetails.WriteString(fmt.Sprintf("  %d) %s\n", i+1, err.Error()))
			}

			LogError(ErrTemplateNotFound, errorDetails.String())
			return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, errorDetails.String())
		}

		LogError(ErrTemplateNotFound, fmt.Sprintf("Template '%s' not found. No loaders configured.", name))
		return nil, fmt.Errorf("%w: '%s'", ErrTemplateNotFound, name)
	}

	// Only cache if caching is enabled
	if e.environment.cache {
		e.mu.Lock()
		e.templates[name] = template
		e.mu.Unlock()
	}

	return template, nil
}

// RegisterString registers a template from a string source
func (e *Engine) RegisterString(name string, source string) error {
	parser := &Parser{}
	nodes, err := parser.Parse(source)
	if err != nil {
		return err
	}

	// String templates always use the current time as modification time
	// and null loader since they're not loaded from a file
	now := time.Now().Unix()

	template := &Template{
		name:         name,
		source:       source,
		nodes:        nodes,
		env:          e.environment,
		engine:       e,
		lastModified: now,
		loader:       nil, // String templates don't have a loader
	}

	// Only cache if caching is enabled
	if e.environment.cache {
		e.mu.Lock()
		e.templates[name] = template
		e.mu.Unlock()
	}

	return nil
}

// GetEnvironment returns the engine's environment
func (e *Engine) GetEnvironment() *Environment {
	return e.environment
}

// IsDebugEnabled returns true if debug mode is enabled
func (e *Engine) IsDebugEnabled() bool {
	return e.environment.debug
}

// IsCacheEnabled returns true if caching is enabled
func (e *Engine) IsCacheEnabled() bool {
	return e.environment.cache
}

// IsAutoReloadEnabled returns true if auto-reload is enabled
func (e *Engine) IsAutoReloadEnabled() bool {
	return e.autoReload
}

// GetCachedTemplateCount returns the number of templates in the cache
func (e *Engine) GetCachedTemplateCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.templates)
}

// GetCachedTemplateNames returns a list of template names in the cache
func (e *Engine) GetCachedTemplateNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		names = append(names, name)
	}
	return names
}

// RegisterTemplate directly registers a pre-built template
func (e *Engine) RegisterTemplate(name string, template *Template) {
	// Set the lastModified timestamp if it's not already set
	if template.lastModified == 0 {
		template.lastModified = time.Now().Unix()
	}

	// Only cache if caching is enabled
	if e.environment.cache {
		e.mu.Lock()
		e.templates[name] = template
		e.mu.Unlock()
	}
}

// CompileTemplate compiles a template for faster rendering
func (e *Engine) CompileTemplate(name string) (*CompiledTemplate, error) {
	template, err := e.Load(name)
	if err != nil {
		return nil, err
	}

	// Compile the template
	return CompileTemplate(template)
}

// RegisterCompiledTemplate registers a compiled template with the engine
func (e *Engine) RegisterCompiledTemplate(compiled *CompiledTemplate) error {
	if compiled == nil {
		return errors.New("cannot register nil compiled template")
	}

	// Load the template from the compiled representation
	template, err := LoadFromCompiled(compiled, e.environment, e)
	if err != nil {
		return err
	}

	// Register the template
	e.RegisterTemplate(compiled.Name, template)

	return nil
}

// LoadFromCompiledData loads a template from serialized compiled data
func (e *Engine) LoadFromCompiledData(data []byte) error {
	// Deserialize the compiled template
	compiled, err := DeserializeCompiledTemplate(data)
	if err != nil {
		return err
	}

	// Register the compiled template
	return e.RegisterCompiledTemplate(compiled)
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
		name:         name,
		source:       source,
		nodes:        nodes,
		env:          e.environment,
		engine:       e,
		lastModified: time.Now().Unix(),
		loader:       nil,
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
		source:       source,
		nodes:        nodes,
		env:          e.environment,
		engine:       e,
		lastModified: time.Now().Unix(),
		loader:       nil,
	}

	return template, nil
}

// Render renders a template with the given context
func (t *Template) Render(context map[string]interface{}) (string, error) {
	// Get a string buffer from the pool
	buf := NewStringBuffer()
	defer buf.Release()

	err := t.RenderTo(buf, context)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderTo renders a template to a writer
func (t *Template) RenderTo(w io.Writer, context map[string]interface{}) error {
	// Get a render context from the pool
	ctx := NewRenderContext(t.env, context, t.engine)

	// Debug logging only when enabled
	LogDebug("Rendering template '%s'", t.name)

	// Ensure the context is returned to the pool
	defer ctx.Release()

	return t.nodes.Render(w, ctx)
}

// Compile compiles the template to a CompiledTemplate
func (t *Template) Compile() (*CompiledTemplate, error) {
	return CompileTemplate(t)
}

// SaveCompiled serializes the compiled template to a byte array
func (t *Template) SaveCompiled() ([]byte, error) {
	// Compile the template
	compiled, err := t.Compile()
	if err != nil {
		return nil, err
	}

	// Serialize the compiled template
	return SerializeCompiledTemplate(compiled)
}

// StringBuffer is a simple buffer for string building
type StringBuffer struct {
	buf bytes.Buffer
}

// stringBufferPool is a sync.Pool for StringBuffer objects
var stringBufferPool = sync.Pool{
	New: func() interface{} {
		return &StringBuffer{}
	},
}

// NewStringBuffer gets a StringBuffer from the pool
func NewStringBuffer() *StringBuffer {
	buffer := stringBufferPool.Get().(*StringBuffer)
	buffer.buf.Reset() // Clear any previous content
	return buffer
}

// Release returns the StringBuffer to the pool
func (b *StringBuffer) Release() {
	stringBufferPool.Put(b)
}

// Write implements io.Writer
func (b *StringBuffer) Write(p []byte) (n int, err error) {
	return b.buf.Write(p)
}

// String returns the buffer's contents as a string
func (b *StringBuffer) String() string {
	return b.buf.String()
}
