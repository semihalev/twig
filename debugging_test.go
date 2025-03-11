package twig

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// Debugging tests
// Consolidated from: debug_test.go, debug_conditional_test.go, debug_equals_test.go,
// debug_print_test.go, elseif_debug_test.go, etc.

// TestOrganizedDebugLevels tests the different debug logging levels
func TestOrganizedDebugLevels(t *testing.T) {
	// Save and restore original debugger state
	origLevel := debugger.level
	origWriter := debugger.writer
	defer func() {
		debugger.level = origLevel
		debugger.writer = origWriter
	}()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	SetDebugWriter(&buf)

	// Test all debug levels
	tests := []struct {
		level       DebugLevel
		shouldLog   bool
		description string
	}{
		{DebugOff, false, "DebugOff should not log"},
		{DebugError, true, "DebugError should log errors"},
		{DebugWarning, true, "DebugWarning should log warnings"},
		{DebugInfo, true, "DebugInfo should log info"},
		{DebugVerbose, true, "DebugVerbose should log verbose"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			buf.Reset()
			SetDebugLevel(test.level)

			testErr := errors.New("test error")
			LogError(testErr, "test context")

			hasOutput := buf.Len() > 0
			if hasOutput != test.shouldLog {
				t.Errorf("Expected logging to be %v, but got %v", test.shouldLog, hasOutput)
			}
		})
	}
}

// TestDebugMode tests debug mode functionality
func TestDebugMode(t *testing.T) {
	// Create a test template
	engine := New()
	engine.SetDebug(true)

	// Save and restore original debugger state
	origLevel := debugger.level
	origWriter := debugger.writer
	defer func() {
		debugger.level = origLevel
		debugger.writer = origWriter
	}()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	SetDebugWriter(&buf)
	SetDebugLevel(DebugVerbose)

	// Create a simple template
	source := "{{ 'Hello' }}"
	engine.RegisterString("debug_test", source)

	// Render the template
	_, err := engine.Render("debug_test", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	// Verify debug output was logged
	if buf.Len() == 0 {
		t.Error("Expected debug output, but got none")
	}
}

// TestDebugConditionals tests debugging of conditionals
func TestDebugConditionals(t *testing.T) {
	// Create a test template
	engine := New()
	engine.SetDebug(true)

	// Save and restore original debugger state
	origLevel := debugger.level
	origWriter := debugger.writer
	defer func() {
		debugger.level = origLevel
		debugger.writer = origWriter
	}()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	SetDebugWriter(&buf)
	SetDebugLevel(DebugVerbose)

	// Create a conditional template
	source := "{% if test %}true{% else %}false{% endif %}"
	engine.RegisterString("debug_conditional", source)

	// Render the template
	_, err := engine.Render("debug_conditional", map[string]interface{}{"test": true})
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	// Verify debug output was logged
	if buf.Len() == 0 {
		t.Error("Expected debug output, but got none")
	}

	// Verify debug output contains conditional evaluation info
	output := buf.String()
	
	// Check for specific debug messages we expect to see
	expectedMessages := []string{
		"Evaluating 'if' condition",
		"Condition result:",
		"Entering 'if' block",
	}
	
	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected debug output to contain '%s', but it was not found", msg)
		}
	}
}

// TestDebugErrorReporting tests error reporting during template execution
func TestDebugErrorReporting(t *testing.T) {
	// Create a test template
	engine := New()
	engine.SetDebug(true)
	// Also enable strict vars for error reporting of undefined variables
	engine.SetStrictVars(true)

	// Save and restore original debugger state
	origLevel := debugger.level
	origWriter := debugger.writer
	defer func() {
		debugger.level = origLevel
		debugger.writer = origWriter
	}()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	SetDebugWriter(&buf)
	SetDebugLevel(DebugError)

	// Create a template with a syntax error rather than an undefined variable
	// Since undefined variables don't cause errors by default in twig
	source := "{{ 1 / 0 }}"  // Division by zero will cause an error
	engine.RegisterString("debug_error", source)

	// Render the template - expect an error
	_, err := engine.Render("debug_error", nil)
	if err == nil {
		t.Fatal("Expected error but got none")
	}

	// Verify the error type and message
	errMsg := err.Error()
	if !strings.Contains(errMsg, "division by zero") && 
	   !strings.Contains(errMsg, "divide by zero") {
		t.Errorf("Expected error message to contain division error, got: %s", errMsg)
	}

	// The error might be directly returned rather than logged
	// Check both the log output and the error message
	output := buf.String()
	if len(output) > 0 && !strings.Contains(output, "ERROR") {
		t.Error("Expected debug output to contain error information")
	}
}
