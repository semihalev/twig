package twig

import (
	"testing"
)

func TestDoTag(t *testing.T) {
	engine := New()
	
	// Add a custom function for testing
	engine.AddFunction("double", func(args ...interface{}) (interface{}, error) {
		if len(args) == 0 {
			return 0, nil
		}
		
		// Type conversion
		var num float64
		switch v := args[0].(type) {
		case int:
			num = float64(v)
		case float64:
			num = v
		case string:
			// Try to convert, or just return 0
			return 0, nil
		default:
			// Return 0 for unsupported types
			return 0, nil
		}
		
		return num * 2, nil
	})
	
	// Disable debug output
	SetDebugLevel(DebugOff)

	tests := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple do tag",
			template: `{% do 5 + 3 %}Output`,
			context:  nil,
			expected: "Output",
		},
		{
			name:     "Set tag",
			template: `{% set count = 1 %}{{ count }}`,
			context:  nil,
			expected: "1",
		},
		{
			name:     "Do tag with assignment",
			template: `{% set count = 0 %}{% do count = 1 %}{{ count }}`,
			context:  nil,
			expected: "1",
		},
		{
			name:     "Do tag with variable and expression",
			template: `{% set x = 5 %}{% do x = x + 10 %}{{ x }}`,
			context:  nil,
			expected: "15",
		},
		{
			name:     "Do tag with function call",
			template: `{% set x = 1 %}{% do x = x * 2 + 3 %}{{ x }}`,
			context:  nil,
			expected: "5",
		},
		{
			name:     "Multiple do tags",
			template: `{% set count = 0 %}{% do count = count + 1 %}{% do count = count + 1 %}{{ count }}`,
			context:  nil,
			expected: "2",
		},
		{
			name:     "Do tag with custom function",
			template: `{% set val = 10 %}{% do val = double(val) %}{{ val }}`,
			context:  nil, // Function is registered with the engine
			expected: "20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := engine.ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(tt.context)
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDoTagErrors(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "Empty do tag",
			template: `{% do %}`,
		},
		{
			name:     "Unclosed do tag",
			template: `{% do 5 + 3`,
		},
		{
			name:     "Invalid expression in do tag",
			template: `{% do 5 + %}`,
		},
		{
			name:     "Invalid variable in assignment",
			template: `{% do 123 = 456 %}`,
		},
		{
			name:     "Missing value in assignment",
			template: `{% do x = %}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.ParseTemplate(tt.template)
			if err == nil {
				t.Errorf("Expected error for invalid template, but got none")
			}
		})
	}
}