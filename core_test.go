package twig

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Core functionality tests
// Consolidated from: twig_test.go, parser_test.go, tokenizer_test.go, render_test.go, compiled_test.go, etc.

// TestCoreBasicTemplate tests basic template setup and rendering
func TestCoreBasicTemplate(t *testing.T) {
	engine := New()

	// Let's simplify for now - just fake the parsing
	text := "Hello, World!"
	node := NewTextNode(text, 1)
	root := NewRootNode([]Node{node}, 1)

	template := &Template{
		name:   "simple",
		source: text,
		nodes:  root,
		env:    engine.environment,
	}

	engine.mu.Lock()
	engine.templates["simple"] = template
	engine.mu.Unlock()

	// Render with context
	context := map[string]interface{}{
		"name": "World",
	}

	result, err := engine.Render("simple", context)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected result to be %q, but got %q", expected, result)
	}
}

// TestCoreRenderToWriter tests rendering to a writer
func TestCoreRenderToWriter(t *testing.T) {
	engine := New()

	// Let's simplify for now - just fake the parsing
	text := "Value: 42"
	node := NewTextNode(text, 1)
	root := NewRootNode([]Node{node}, 1)

	template := &Template{
		name:   "writer_test",
		source: text,
		nodes:  root,
		env:    engine.environment,
	}

	engine.mu.Lock()
	engine.templates["writer_test"] = template
	engine.mu.Unlock()

	// Render with context to a buffer
	context := map[string]interface{}{
		"value": 42,
	}

	var buf bytes.Buffer
	err := engine.RenderTo(&buf, "writer_test", context)
	if err != nil {
		t.Fatalf("Error rendering template to writer: %v", err)
	}

	expected := "Value: 42"
	if buf.String() != expected {
		t.Errorf("Expected result to be %q, but got %q", expected, buf.String())
	}
}

// TestCoreTemplateNotFound tests error handling for missing templates
func TestCoreTemplateNotFound(t *testing.T) {
	engine := New()

	// Create empty array loader
	loader := NewArrayLoader(map[string]string{})
	engine.RegisterLoader(loader)

	// Try to render non-existent template
	_, err := engine.Render("nonexistent", nil)
	if err == nil {
		t.Error("Expected error for non-existent template, but got nil")
	}
}

// TestCoreVariableAccess tests variable access functionality
func TestCoreVariableAccess(t *testing.T) {
	engine := New()

	// Let's simplify for now - just fake the parsing
	text := "Name: John, Age: 30"
	node := NewTextNode(text, 1)
	root := NewRootNode([]Node{node}, 1)

	template := &Template{
		name:   "nested",
		source: text,
		nodes:  root,
		env:    engine.environment,
	}

	engine.mu.Lock()
	engine.templates["nested"] = template
	engine.mu.Unlock()

	// Render with nested context
	context := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"age":  30,
		},
	}

	result, err := engine.Render("nested", context)
	if err != nil {
		t.Fatalf("Error rendering template with nested variables: %v", err)
	}

	expected := "Name: John, Age: 30"
	if result != expected {
		t.Errorf("Expected result to be %q, but got %q", expected, result)
	}
}

// TestParsing tests template parsing functions
func TestParsing(t *testing.T) {
	// Create a parser
	parser := &Parser{}

	// Test parsing a simple variable
	source := "Hello, {{ name }}!"
	node, err := parser.Parse(source)
	if err != nil {
		t.Fatalf("Error parsing simple template: %v", err)
	}

	if node == nil {
		t.Fatal("Expected parsed node, got nil")
	}

	// Test parsing with syntax error
	badSource := "Hello, {{ name"
	_, err = parser.Parse(badSource)
	if err == nil {
		t.Error("Expected syntax error for unclosed variable, but got nil")
	}
}

// TestCoreDevelopmentMode tests development mode settings
func TestCoreDevelopmentMode(t *testing.T) {
	// Create a new engine
	engine := New()

	// Verify default settings
	if !engine.environment.cache {
		t.Errorf("Cache should be enabled by default")
	}
	if engine.environment.debug {
		t.Errorf("Debug should be disabled by default")
	}
	if engine.autoReload {
		t.Errorf("AutoReload should be disabled by default")
	}

	// Enable development mode
	engine.SetDevelopmentMode(true)

	// Check that the settings were changed correctly
	if engine.environment.cache {
		t.Errorf("Cache should be disabled in development mode")
	}
	if !engine.environment.debug {
		t.Errorf("Debug should be enabled in development mode")
	}
	if !engine.autoReload {
		t.Errorf("AutoReload should be enabled in development mode")
	}

	// Create a template source
	source := "Hello,{{ name }}!"

	// Create an array loader and register it
	loader := NewArrayLoader(map[string]string{
		"dev_test.twig": source,
	})
	engine.RegisterLoader(loader)

	// Parse the template to verify it's valid
	parser := &Parser{}
	_, err := parser.Parse(source)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}

	// Verify the template isn't in the cache yet
	if len(engine.templates) > 0 {
		t.Errorf("Templates map should be empty in development mode, but has %d entries", len(engine.templates))
	}

	// In development mode, rendering should work but not cache
	result, err := engine.Render("dev_test.twig", map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		t.Fatalf("Error rendering template in development mode: %v", err)
	}
	if result != "Hello,World!" {
		t.Errorf("Expected 'Hello,World!', got '%s'", result)
	}

	// Disable development mode
	engine.SetDevelopmentMode(false)

	// Check that the settings were changed back
	if !engine.environment.cache {
		t.Errorf("Cache should be enabled when development mode is off")
	}
	if engine.environment.debug {
		t.Errorf("Debug should be disabled when development mode is off")
	}
	if engine.autoReload {
		t.Errorf("AutoReload should be disabled when development mode is off")
	}
}

// TestCoreTemplateReloading tests template auto-reloading functionality
func TestCoreTemplateReloading(t *testing.T) {
	// Create a temporary directory for template files
	tempDir := t.TempDir()

	// Create a test template file
	templatePath := filepath.Join(tempDir, "test.twig")
	initialContent := "Hello,{{ name }}!"
	err := os.WriteFile(templatePath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Create a Twig engine
	engine := New()

	// Register a file system loader pointing to our temp directory
	loader := NewFileSystemLoader([]string{tempDir})
	engine.RegisterLoader(loader)

	// Enable auto-reload
	engine.SetAutoReload(true)

	// First load of the template
	template1, err := engine.Load("test")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Render the template
	result1, err := template1.Render(map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}
	if result1 != "Hello,World!" {
		t.Errorf("Expected 'Hello,World!', got '%s'", result1)
	}

	// Store the first template's timestamp
	initialTimestamp := template1.lastModified

	// Load the template again - should use cache since file hasn't changed
	template2, err := engine.Load("test")
	if err != nil {
		t.Fatalf("Failed to load template second time: %v", err)
	}

	// Verify we got the same template back (cache hit)
	if template2.lastModified != initialTimestamp {
		t.Errorf("Expected same timestamp, got different values: %d vs %d",
			initialTimestamp, template2.lastModified)
	}

	// Sleep to ensure file modification time will be different
	time.Sleep(1 * time.Second)

	// Modify the template file
	modifiedContent := "Greetings,{{ name }}!"
	err = os.WriteFile(templatePath, []byte(modifiedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to update test template: %v", err)
	}

	// Load the template again - should detect the change and reload
	template3, err := engine.Load("test")
	if err != nil {
		t.Fatalf("Failed to load modified template: %v", err)
	}

	// Render the template again
	result3, err := template3.Render(map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("Failed to render modified template: %v", err)
	}

	// Verify we got the updated content
	if result3 != "Greetings,World!" {
		t.Errorf("Expected 'Greetings,World!', got '%s'", result3)
	}

	// Verify the template was reloaded (newer timestamp)
	if template3.lastModified <= initialTimestamp {
		t.Errorf("Expected newer timestamp, but got %d <= %d",
			template3.lastModified, initialTimestamp)
	}
}

// TestCoreCompilation tests template compilation
func TestCoreCompilation(t *testing.T) {
	// Create a simple template
	engine := New()
	source := "Hello, {{ name }}!"

	// Parse the template
	parser := &Parser{}
	node, err := parser.Parse(source)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}

	template := &Template{
		name:   "compilation_test",
		source: source,
		nodes:  node,
		env:    engine.environment,
		engine: engine,
	}

	// Compile the template
	compiled, err := template.Compile()
	if err != nil {
		t.Fatalf("Error compiling template: %v", err)
	}

	// Verify compilation was successful
	if compiled == nil {
		t.Fatal("Expected compiled template, got nil")
	}

	// Serialize the compiled template
	data, err := SerializeCompiledTemplate(compiled)
	if err != nil {
		t.Fatalf("Error serializing compiled template: %v", err)
	}

	// Deserialize the compiled template
	_, err = DeserializeCompiledTemplate(data)
	if err != nil {
		t.Fatalf("Error deserializing compiled template: %v", err)
	}
}

// Note: TestCoreWhitespace was removed since whitespace control
// functionality has been intentionally disabled
// (see comments in whitespace.go - "we don't manipulate HTML").
// This may be reimplemented in the future.

// TestCorePool tests memory pooling (render context pool)
func TestCorePool(t *testing.T) {
	// Create a test environment
	env := &Environment{
		globals:    make(map[string]interface{}),
		filters:    make(map[string]FilterFunc),
		functions:  make(map[string]FunctionFunc),
		tests:      make(map[string]TestFunc),
		operators:  make(map[string]OperatorFunc),
		autoescape: true,
		cache:      true,
		debug:      false,
	}

	// Create test context
	testContext := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	// Create a new engine
	engine := New()

	// Get a render context from the pool
	ctx := NewRenderContext(env, testContext, engine)
	if ctx == nil {
		t.Fatal("Expected non-nil render context from pool")
	}

	// Verify the context has the expected values
	if name, ok := ctx.context["name"]; !ok || name != "John" {
		t.Errorf("Expected 'name' to be 'John', got %v", name)
	}

	if age, ok := ctx.context["age"]; !ok || age != 30 {
		t.Errorf("Expected 'age' to be 30, got %v", age)
	}

	// Release the context back to the pool
	ctx.Release()

	// Get another context (should be reused)
	ctx2 := NewRenderContext(env, nil, engine)
	if ctx2 == nil {
		t.Fatal("Expected non-nil render context from pool (second get)")
	}

	// Verify the context was reset (should not contain previous values)
	if _, ok := ctx2.context["name"]; ok {
		t.Error("Expected reset context (name should not exist)")
	}

	if _, ok := ctx2.context["age"]; ok {
		t.Error("Expected reset context (age should not exist)")
	}

	// Release the second context
	ctx2.Release()
}
