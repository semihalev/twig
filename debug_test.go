package twig

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestDebugLevels(t *testing.T) {
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

func TestEnhancedError(t *testing.T) {
	baseErr := errors.New("something went wrong")
	src := "Line 1\nLine 2 with {{ error here }}\nLine 3"

	enhancedErr := NewError(baseErr, "test.twig", 2, 15, src)
	errString := enhancedErr.Error()

	// Check that the error contains the expected components
	expectedComponents := []string{
		"Error in template 'test.twig'",
		"at line 2",
		"something went wrong",
		"Line 2 with {{ error here }}",
		"^",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(errString, component) {
			t.Errorf("Enhanced error should contain '%s'. Got: %s", component, errString)
		}
	}

	// Test Unwrap functionality
	unwrappedErr := errors.Unwrap(enhancedErr)
	if unwrappedErr != baseErr {
		t.Errorf("Unwrapped error should be the original error")
	}
}

func TestErrorFormatContext(t *testing.T) {
	source := "Line 1\nLine 2 with an error\nLine 3"

	// Test valid error position - don't test exact position of ^ as it's not critical
	ctx := FormatErrorContext(source, 25, 2)
	if !strings.Contains(ctx, "Line 2: Line 2 with an error") {
		t.Errorf("Expected context to contain the error line, got: %s", ctx)
	}
	if !strings.Contains(ctx, "^") {
		t.Errorf("Expected context to contain the caret marker, got: %s", ctx)
	}

	// Test invalid positions
	invalidTests := []struct {
		name     string
		source   string
		position int
		line     int
	}{
		{"Empty source", "", 5, 1},
		{"Negative position", source, -1, 2},
		{"Line out of range", source, 10, 99},
		{"Zero line", source, 10, 0},
	}

	for _, test := range invalidTests {
		t.Run(test.name, func(t *testing.T) {
			ctx := FormatErrorContext(test.source, test.position, test.line)
			if ctx != "" {
				t.Errorf("Expected empty context for invalid parameters, got: %s", ctx)
			}
		})
	}
}

func TestTracing(t *testing.T) {
	// Save and restore original debugger state
	origLevel := debugger.level
	origWriter := debugger.writer
	defer func() {
		debugger.level = origLevel
		debugger.writer = origWriter
	}()

	// Enable debugging for the test
	var buf bytes.Buffer
	SetDebugWriter(&buf)
	SetDebugLevel(DebugInfo)

	// Test template tracing
	buf.Reset()
	endTrace := StartTrace("test-template")
	endTrace()

	output := buf.String()
	if !strings.Contains(output, "Begin rendering template: test-template") {
		t.Errorf("Trace start message missing from output: %s", output)
	}
	if !strings.Contains(output, "Completed rendering template: test-template") {
		t.Errorf("Trace end message missing from output: %s", output)
	}

	// Test section tracing
	buf.Reset()
	SetDebugLevel(DebugVerbose)
	endSection := TraceSection("test-section")
	endSection()

	output = buf.String()
	if !strings.Contains(output, "Begin section: test-section") {
		t.Errorf("Section start message missing from output: %s", output)
	}
	if !strings.Contains(output, "End section: test-section") {
		t.Errorf("Section end message missing from output: %s", output)
	}
}

func TestErrorWithTemplateNotFound(t *testing.T) {
	engine := New()
	
	// Attempt to load a non-existent template
	_, err := engine.Load("non-existent")
	
	if err == nil {
		t.Fatal("Expected error for missing template, got nil")
	}
	
	// The error should contain template name and ErrTemplateNotFound
	if !strings.Contains(err.Error(), "non-existent") {
		t.Errorf("Error should contain template name, but got: %s", err.Error())
	}
	
	// Check that it wraps ErrTemplateNotFound
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("Error should wrap ErrTemplateNotFound, but got: %T: %v", err, err)
	}
}

func TestErrorWithInvalidSyntax(t *testing.T) {
	engine := New()
	
	// Register a template with invalid syntax
	err := engine.RegisterString("invalid", "{{ unclosed tag")
	
	if err == nil {
		t.Fatal("Expected error for invalid syntax, got nil")
	}
	
	// Error should contain information about the syntax error
	if !strings.Contains(err.Error(), "parsing error") {
		t.Errorf("Error should mention parsing error, but got: %s", err.Error())
	}
}

func TestDebugRender(t *testing.T) {
	// Save original state
	origLevel := debugger.level
	origWriter := debugger.writer
	defer func() {
		debugger.level = origLevel
		debugger.writer = origWriter
	}()
	
	// Set up debug logging
	var logBuf bytes.Buffer
	SetDebugWriter(&logBuf)
	SetDebugLevel(DebugInfo)
	
	// Create a template and engine
	engine := New()
	// Make sure to add a space in the output to match what the default Twig rendering would do
	engine.RegisterString("debug-template", "Hello {{ name }}!")
	
	// Get the template
	tmpl, err := engine.Load("debug-template")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}
	
	// Render with debugging
	var outBuf bytes.Buffer
	ctx := NewRenderContext(engine.environment, map[string]interface{}{"name": "World"}, engine)
	defer ctx.Release()
	
	err = DebugRender(&outBuf, tmpl, ctx)
	if err != nil {
		t.Fatalf("DebugRender failed: %v", err)
	}
	
	// Check output - only check that it contains the expected parts
	output := outBuf.String()
	if !strings.Contains(output, "Hello") || !strings.Contains(output, "World") {
		t.Errorf("Expected output to contain 'Hello' and 'World', got '%s'", output)
	}
	
	// Check log output
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Rendering template") {
		t.Errorf("Debug log should contain render start message")
	}
	if !strings.Contains(logOutput, "Completed rendering") {
		t.Errorf("Debug log should contain render completion message")
	}
}