package twig

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// StringLoader is a simple template loader that delegates to engine's registered templates
type StringLoader struct {
	templates map[string]string
}

// Load implements the Loader interface and returns a template by name
func (l *StringLoader) Load(name string) (string, error) {
	// The engine already has the templates registered via RegisterString
	// This is just a dummy implementation to satisfy the interface
	// The actual template loading is handled by the engine's internal cache
	return "", fmt.Errorf("template not found: '%s'", name)
}

// Exists implements the Loader interface
func (l *StringLoader) Exists(name string) bool {
	// Always return false to let the engine load from its internal cache
	return false
}

// TestExtension is a simple extension for testing
type TestExtension struct {
	functions map[string]FunctionFunc
	filters   map[string]FilterFunc
}

func (e *TestExtension) GetName() string {
	return "test_extension"
}

func (e *TestExtension) GetFilters() map[string]FilterFunc {
	return e.filters
}

func (e *TestExtension) GetFunctions() map[string]FunctionFunc {
	return e.functions
}

func (e *TestExtension) GetTests() map[string]TestFunc {
	return nil
}

func (e *TestExtension) GetOperators() map[string]OperatorFunc {
	return nil
}

func (e *TestExtension) GetTokenParsers() []TokenParser {
	return nil
}

func (e *TestExtension) Initialize(engine *Engine) {
	// Nothing to initialize
}

// TestSandboxFunctions tests if the sandbox can restrict function access
func TestSandboxFunctions(t *testing.T) {
	// Create a fresh engine
	engine := New()

	// Create a default security policy that doesn't allow any functions
	policy := NewDefaultSecurityPolicy()
	policy.AllowedFunctions = map[string]bool{} // Start with no allowed functions

	// Register a test function through a custom extension
	engine.AddExtension(&TestExtension{
		functions: map[string]FunctionFunc{
			"test_func": func(args ...interface{}) (interface{}, error) {
				return "test function called", nil
			},
		},
	})

	// Enable sandbox mode with the restrictive policy
	engine.EnableSandbox(policy)

	// Register a template that uses the function
	err := engine.RegisterString("sandbox_test", "{{ test_func() }}")
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	// Render in sandbox mode (should fail)
	ctx := NewRenderContext(engine.environment, nil, engine)
	ctx.EnableSandbox() // Enable sandbox mode explicitly in context

	// Try to render
	var buf bytes.Buffer
	template, err := engine.Load("sandbox_test")
	if err != nil {
		t.Fatalf("Error loading template: %v", err)
	}

	// Rendering should fail because the function is not allowed
	err = template.nodes.Render(&buf, ctx)
	if err == nil {
		t.Errorf("Expected sandbox to block unauthorized function, but it didn't")
	} else {
		t.Logf("Correctly got error: %v", err)
	}

	// Now allow the function and try again
	policy.AllowedFunctions["test_func"] = true

	// Create a new context (with sandbox enabled)
	ctx = NewRenderContext(engine.environment, nil, engine)
	ctx.EnableSandbox()

	// Reset buffer
	buf.Reset()

	// Rendering should succeed now
	err = template.nodes.Render(&buf, ctx)
	if err != nil {
		t.Errorf("Rendering failed after allowing function: %v", err)
	}

	expected := "test function called"
	if buf.String() != expected {
		t.Errorf("Expected rendered output '%s', got '%s'", expected, buf.String())
	}
}

// TestSandboxFilters tests if the sandbox can restrict filter access
func TestSandboxFilters(t *testing.T) {
	// Create a fresh engine
	engine := New()

	// Create a default security policy that doesn't allow any filters
	policy := NewDefaultSecurityPolicy()
	policy.AllowedFilters = map[string]bool{} // Start with no allowed filters

	// Register a test filter through a custom extension
	engine.AddExtension(&TestExtension{
		filters: map[string]FilterFunc{
			"test_filter": func(value interface{}, args ...interface{}) (interface{}, error) {
				return "filtered content", nil
			},
		},
	})

	// Enable sandbox mode with the restrictive policy
	engine.EnableSandbox(policy)

	// Register a template that uses the filter
	err := engine.RegisterString("sandbox_filter_test", "{{ 'anything'|test_filter }}")
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	// Render in sandbox mode (should fail)
	ctx := NewRenderContext(engine.environment, nil, engine)
	ctx.EnableSandbox() // Enable sandbox mode explicitly in context

	// Try to render
	var buf bytes.Buffer
	template, err := engine.Load("sandbox_filter_test")
	if err != nil {
		t.Fatalf("Error loading template: %v", err)
	}

	// Rendering should fail because the filter is not allowed
	err = template.nodes.Render(&buf, ctx)
	if err == nil {
		t.Errorf("Expected sandbox to block unauthorized filter, but it didn't")
	} else {
		t.Logf("Correctly got error: %v", err)
	}

	// Now allow the filter and try again
	policy.AllowedFilters["test_filter"] = true

	// Create a new context (with sandbox enabled)
	ctx = NewRenderContext(engine.environment, nil, engine)
	ctx.EnableSandbox()

	// Reset buffer
	buf.Reset()

	// Rendering should succeed now
	err = template.nodes.Render(&buf, ctx)
	if err != nil {
		t.Errorf("Rendering failed after allowing filter: %v", err)
	}

	expected := "filtered content"
	if buf.String() != expected {
		t.Errorf("Expected rendered output '%s', got '%s'", expected, buf.String())
	}
}

// TestSandboxOption tests the sandbox flag on render context
func TestSandboxOption(t *testing.T) {
	// Create a fresh engine
	engine := New()

	// Create a security policy that allows specific functions
	policy := NewDefaultSecurityPolicy()
	policy.AllowedFunctions = map[string]bool{
		"safe_func": true, // This function is allowed in sandboxed includes
	}

	// Register both safe and dangerous functions
	engine.AddExtension(&TestExtension{
		functions: map[string]FunctionFunc{
			"safe_func": func(args ...interface{}) (interface{}, error) {
				return "safe function called", nil
			},
			"dangerous_func": func(args ...interface{}) (interface{}, error) {
				return "dangerous function called", nil
			},
		},
	})

	// Enable sandbox mode with the policy
	engine.EnableSandbox(policy)

	// Create a standard (non-sandboxed) context
	ctx := NewRenderContext(engine.environment, nil, engine)

	// Verify the context is not sandboxed initially
	if ctx.IsSandboxed() {
		t.Errorf("Context should not be sandboxed initially")
	}

	// Create a child context for an include with sandbox option
	// This simulates what happens in IncludeNode.Render
	includeCtx := NewRenderContext(ctx.env, make(map[string]interface{}), ctx.engine)

	// Explicitly enable sandbox
	includeCtx.EnableSandbox()

	// Verify the child context is now sandboxed
	if !includeCtx.IsSandboxed() {
		t.Errorf("Child context should be sandboxed after EnableSandbox()")
	}

	// Verify safe function works in sandbox mode
	evalNode := &FunctionNode{
		name: "safe_func",
		args: []Node{},
	}

	result, err := includeCtx.EvaluateExpression(evalNode)
	if err != nil {
		t.Errorf("Safe function should work in sandbox mode: %v", err)
	} else {
		if result != "safe function called" {
			t.Errorf("Unexpected result from safe function: got %v, expected 'safe function called'", result)
		}
	}

	// Verify dangerous function fails in sandbox mode
	evalDangerousNode := &FunctionNode{
		name: "dangerous_func",
		args: []Node{},
	}

	_, err = includeCtx.EvaluateExpression(evalDangerousNode)
	if err == nil {
		t.Errorf("Dangerous function should be blocked in sandbox mode")
	} else {
		// Verify the error message mentions the dangerous function
		if msg := err.Error(); !strings.Contains(msg, "dangerous_func") {
			t.Errorf("Expected error to mention 'dangerous_func', but got: %v", err)
		} else {
			t.Logf("Correctly got error: %v", err)
		}
	}
}
