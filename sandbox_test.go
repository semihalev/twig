package twig

import (
	"testing"
)

// TestSandboxIncludes tests the sandboxed option for include tags
func TestSandboxIncludes(t *testing.T) {
	engine := New()

	// Create a security policy
	policy := NewDefaultSecurityPolicy()

	// Enable sandbox mode with the policy
	engine.EnableSandbox(policy)

	// Register templates for testing
	err := engine.RegisterString("sandbox_parent", "Parent: {% include 'sandbox_child' sandboxed %}")
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	err = engine.RegisterString("sandbox_child", "{{ harmful_function() }}")
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	// Define a harmful function in the context
	context := map[string]interface{}{
		"harmful_function": func() string {
			return "This should be blocked in sandbox mode"
		},
	}

	// Register the function
	engine.RegisterFunction("harmful_function", func() string {
		return "This should be blocked in sandbox mode"
	})

	// The harmful function is not in the allowed list, so it should fail
	_, err = engine.Render("sandbox_parent", context)
	if err == nil {
		t.Errorf("Expected sandbox violation error but got none")
	}

	// Now allow the function in the security policy
	policy.AllowedFunctions["harmful_function"] = true

	// Now it should work
	result, err := engine.Render("sandbox_parent", context)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Parent: This should be blocked in sandbox mode"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test non-sandboxed include
	err = engine.RegisterString("non_sandbox_parent", "Parent: {% include 'sandbox_child' %}")
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	// Non-sandboxed includes should always work regardless of security policy
	result, err = engine.Render("non_sandbox_parent", context)
	if err != nil {
		t.Errorf("Unexpected error in non-sandboxed include: %v", err)
	}

	// The result should be the same
	if result != expected {
		t.Errorf("Expected '%s' for non-sandboxed include, got '%s'", expected, result)
	}
}