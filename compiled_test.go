package twig

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTemplateCompilation(t *testing.T) {
	// Create a new engine
	engine := New()

	// Register a template string
	source := "Hello, {{ name }}!"
	name := "test"

	err := engine.RegisterString(name, source)
	if err != nil {
		t.Fatalf("Failed to register template: %v", err)
	}

	// Compile the template
	compiled, err := engine.CompileTemplate(name)
	if err != nil {
		t.Fatalf("Failed to compile template: %v", err)
	}

	// Check the compiled template
	if compiled.Name != name {
		t.Errorf("Expected compiled name to be %s, got %s", name, compiled.Name)
	}

	// Serialize the compiled template
	data, err := SerializeCompiledTemplate(compiled)
	if err != nil {
		t.Fatalf("Failed to serialize compiled template: %v", err)
	}

	// Create a new engine
	newEngine := New()

	// Load the compiled template
	err = newEngine.LoadFromCompiledData(data)
	if err != nil {
		t.Fatalf("Failed to load compiled template: %v", err)
	}

	// Render the template
	context := map[string]interface{}{
		"name": "World",
	}

	result, err := newEngine.Render(name, context)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Check the result
	expected := "Hello,World!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCompiledLoader(t *testing.T) {
	// Create a temporary directory for compiled templates
	tempDir, err := ioutil.TempDir("", "twig-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new engine
	engine := New()

	// Register some templates
	templates := map[string]string{
		"test1": "Hello, {{ name }}!",
		"test2": "{% if show %}Shown{% else %}Hidden{% endif %}",
		"test3": "{% for item in items %}{{ item }}{% endfor %}",
	}

	for name, source := range templates {
		err := engine.RegisterString(name, source)
		if err != nil {
			t.Fatalf("Failed to register template %s: %v", name, err)
		}
	}

	// Create a compiled loader
	loader := NewCompiledLoader(tempDir)

	// Compile all templates
	err = loader.CompileAll(engine)
	if err != nil {
		t.Fatalf("Failed to compile templates: %v", err)
	}

	// Check that the compiled files exist
	for name := range templates {
		path := filepath.Join(tempDir, name+".twig.compiled")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Compiled file %s does not exist", path)
		}
	}

	// Create a new engine
	newEngine := New()

	// Load all compiled templates
	err = loader.LoadAll(newEngine)
	if err != nil {
		t.Fatalf("Failed to load compiled templates: %v", err)
	}

	// Test rendering
	testCases := []struct {
		name     string
		context  map[string]interface{}
		expected string
	}{
		{
			name: "test1",
			context: map[string]interface{}{
				"name": "World",
			},
			expected: "Hello,World!",
		},
		{
			name: "test2",
			context: map[string]interface{}{
				"show": true,
			},
			expected: "Shown",
		},
		{
			name: "test3",
			context: map[string]interface{}{
				"items": []string{"a", "b", "c"},
			},
			expected: "abc",
		},
	}

	for _, tc := range testCases {
		result, err := newEngine.Render(tc.name, tc.context)
		if err != nil {
			t.Errorf("Failed to render template %s: %v", tc.name, err)
			continue
		}

		if result != tc.expected {
			t.Errorf("Template %s: expected %q, got %q", tc.name, tc.expected, result)
		}
	}
}

func TestCompiledLoaderReload(t *testing.T) {
	// Create a temporary directory for compiled templates
	tempDir, err := ioutil.TempDir("", "twig-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new engine with the compiled loader
	engine := New()
	loader := NewCompiledLoader(tempDir)

	// Register a template
	name := "test"
	source := "Version 1"

	err = engine.RegisterString(name, source)
	if err != nil {
		t.Fatalf("Failed to register template: %v", err)
	}

	// Compile the template
	err = loader.SaveCompiled(engine, name)
	if err != nil {
		t.Fatalf("Failed to compile template: %v", err)
	}

	// Create a new engine and load the compiled template
	newEngine := New()
	newEngine.SetAutoReload(true)
	newEngine.RegisterLoader(loader)

	err = loader.LoadCompiled(newEngine, name)
	if err != nil {
		t.Fatalf("Failed to load compiled template: %v", err)
	}

	// Render the template
	result, err := newEngine.Render(name, nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Check the result
	expected := "Version1"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Wait a bit to ensure file system time difference
	time.Sleep(100 * time.Millisecond)

	// Update the template
	source = "Version 2"
	err = engine.RegisterString(name, source)
	if err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	// Compile the updated template
	err = loader.SaveCompiled(engine, name)
	if err != nil {
		t.Fatalf("Failed to compile updated template: %v", err)
	}

	// Ensure the file system timestamp is updated
	filePath := filepath.Join(tempDir, name+".twig.compiled")
	current := time.Now().Local()
	err = os.Chtimes(filePath, current, current)
	if err != nil {
		t.Fatalf("Failed to update file time: %v", err)
	}

	// Force cache to be cleared by directly unregistering the template
	newEngine.mu.Lock()
	delete(newEngine.templates, name)
	newEngine.mu.Unlock()

	// Render the template again (should load the new version)
	result, err = newEngine.Render(name, nil)
	if err != nil {
		t.Fatalf("Failed to render updated template: %v", err)
	}

	// Check the result
	expected = "Version2"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func BenchmarkCompiledTemplateRendering(b *testing.B) {
	// Create a direct template
	directEngine := New()
	source := "Hello, {{ name }}! {% if show %}Shown{% else %}Hidden{% endif %} {% for item in items %}{{ item }}{% endfor %}"
	name := "bench"

	err := directEngine.RegisterString(name, source)
	if err != nil {
		b.Fatalf("Failed to register template: %v", err)
	}

	// Create a compiled template
	compiledEngine := New()
	compiled, err := directEngine.CompileTemplate(name)
	if err != nil {
		b.Fatalf("Failed to compile template: %v", err)
	}

	data, err := SerializeCompiledTemplate(compiled)
	if err != nil {
		b.Fatalf("Failed to serialize template: %v", err)
	}

	err = compiledEngine.LoadFromCompiledData(data)
	if err != nil {
		b.Fatalf("Failed to load compiled template: %v", err)
	}

	// Context for rendering
	context := map[string]interface{}{
		"name":  "World",
		"show":  true,
		"items": []string{"a", "b", "c"},
	}

	// Benchmark rendering the direct template
	b.Run("Direct", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			buf.Reset()
			directEngine.RenderTo(&buf, name, context)
		}
	})

	// Benchmark rendering the compiled template
	b.Run("Compiled", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			buf.Reset()
			compiledEngine.RenderTo(&buf, name, context)
		}
	})
}
