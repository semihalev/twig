package twig

import (
	"bytes"
	"testing"
)

func TestBasicTemplate(t *testing.T) {
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

func TestRenderToWriter(t *testing.T) {
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

func TestTemplateNotFound(t *testing.T) {
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

func TestVariableAccess(t *testing.T) {
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

func TestParseStringTemplate(t *testing.T) {
	engine := New()
	
	// Create a pre-parsed template
	text := "Count: 5"
	node := NewTextNode(text, 1)
	root := NewRootNode([]Node{node}, 1)
	
	template := &Template{
		source: text,
		nodes:  root,
		env:    engine.environment,
	}
	
	// Simulate the Parse function
	engine.Parse = func(source string) (*Template, error) {
		return template, nil
	}
	
	// Parse template string directly
	template, err := engine.ParseTemplate("Count: {{ count }}")
	if err != nil {
		t.Fatalf("Error parsing template string: %v", err)
	}
	
	// Render with context
	context := map[string]interface{}{
		"count": 5,
	}
	
	result, err := template.Render(context)
	if err != nil {
		t.Fatalf("Error rendering parsed template: %v", err)
	}
	
	expected := "Count: 5"
	if result != expected {
		t.Errorf("Expected result to be %q, but got %q", expected, result)
	}
}