package twig

import (
	"testing"
)

// TestEdgeCases tests various edge cases in the template engine
func TestEdgeCases(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		// Empty values
		{
			name:     "Empty string value",
			source:   "Value: '{{ value }}'",
			context:  map[string]interface{}{"value": ""},
			expected: "Value: ''",
		},
		{
			name:     "Nil value",
			source:   "Value: '{{ value }}'",
			context:  map[string]interface{}{"value": nil},
			expected: "Value: ''",
		},
		{
			name:     "Empty array",
			source:   "Count: {{ items|length }}",
			context:  map[string]interface{}{"items": []interface{}{}},
			expected: "Count: 0",
		},
		{
			name:     "For loop with empty array",
			source:   "Items: {% for item in items %}{{ item }}{% else %}none{% endfor %}",
			context:  map[string]interface{}{"items": []interface{}{}},
			expected: "Items: none",
		},

		// Type conversion edge cases
		{
			name:     "Boolean to string in concatenation",
			source:   "Result: {{ 'Value is ' ~ value }}",
			context:  map[string]interface{}{"value": true},
			expected: "Result: Value is true",
		},
		{
			name:     "Number to string in concatenation",
			source:   "Result: {{ 'Number: ' ~ number }}",
			context:  map[string]interface{}{"number": 42.5},
			expected: "Result: Number: 42.5",
		},
		{
			name:     "String as number in addition",
			source:   "Result: {{ '5' + 3 }}",
			context:  nil,
			expected: "Result: 8",
		},

		// Undefined variables
		{
			name:     "Undefined variable",
			source:   "Value: '{{ undefined_var }}'",
			context:  map[string]interface{}{},
			expected: "Value: ''",
		},
		{
			name:     "Undefined variable in if condition",
			source:   "{% if undefined_var %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{},
			expected: "false",
		},
		{
			name:     "Default filter with undefined variable",
			source:   "{{ undefined_var|default('default value') }}",
			context:  map[string]interface{}{},
			expected: "default value",
		},
		{
			name:     "Defined test with undefined variable",
			source:   "{% if undefined_var is defined %}defined{% else %}undefined{% endif %}",
			context:  map[string]interface{}{},
			expected: "undefined",
		},

		// Special characters
		{
			name:     "Template with special characters",
			source:   "{{ '\\\"special\\\" & <chars>' }}",
			context:  nil,
			expected: "\"special\" & <chars>",
		},
		{
			name:     "UTF-8 characters",
			source:   "{{ 'UTF-8: いろはにほへと' }}",
			context:  nil,
			expected: "UTF-8: いろはにほへと",
		},

		// Nested attribute access
		{
			name:   "Deeply nested attribute access",
			source: "{{ user.profile.contact.email }}",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"contact": map[string]interface{}{
							"email": "test@example.com",
						},
					},
				},
			},
			expected: "test@example.com",
		},
		{
			name:   "Attribute access with nil intermediate",
			source: "{{ user.profile.contact.email|default('no email') }}",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": nil,
				},
			},
			expected: "no email",
		},

		// Whitespace control
		// TODO: Add proper whitespace trimming support
		// {
		// 	name:     "Whitespace control with dash",
		// 	source:   "{%- if true -%}content{%- endif -%}",
		// 	context:  nil,
		// 	expected: "content",
		// },
		// {
		// 	name:     "Whitespace control in blocks",
		// 	source:   "{% set value %}   spaced   content   {% endset %}|{{ value|trim }}|",
		// 	context:  nil,
		// 	expected: "|   spaced   content   |",
		// },

		// Escaping syntax
		// TODO: Add proper escape sequence support for template tags
		// {
		// 	name:     "Escaped delimiters with backslash",
		// 	source:   "\\{{ Not a variable \\}}",
		// 	context:  nil,
		// 	expected: "{{ Not a variable }}",
		// },
		{
			name:     "Literal braces in strings",
			source:   "{{ '{not a variable}' }}",
			context:  nil,
			expected: "{not a variable}",
		},

		// Complex expressions
		{
			name:     "Complex nested expressions",
			source:   "{{ ((1 + 2) * 3) / (4 - 2) }}",
			context:  nil,
			expected: "4.5",
		},
		{
			name:     "Operator precedence",
			source:   "{{ 1 + 2 * 3 }}",
			context:  nil,
			expected: "7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.RegisterString(tt.name, tt.source)
			if err != nil {
				t.Fatalf("Error registering template: %v", err)
			}

			result, err := engine.Render(tt.name, tt.context)
			if err != nil {
				t.Fatalf("Error rendering template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestErrorConditions tests that appropriate errors are generated
func TestErrorConditions(t *testing.T) {
	engine := New()

	tests := []struct {
		name        string
		source      string
		context     map[string]interface{}
		shouldError bool
	}{
		// Syntax errors
		{
			name:        "Unclosed tag",
			source:      "{% if true %}content",
			context:     nil,
			shouldError: true,
		},
		{
			name:        "Unclosed variable",
			source:      "{{ variable",
			context:     nil,
			shouldError: true,
		},
		{
			name:        "Invalid operator",
			source:      "{{ 1 ++ 2 }}",
			context:     nil,
			shouldError: true,
		},
		{
			name:        "Unbalanced parentheses",
			source:      "{{ (1 + 2 }}",
			context:     nil,
			shouldError: true,
		},

		// Runtime errors
		{
			name:        "Division by zero",
			source:      "{{ 1 / 0 }}",
			context:     nil,
			shouldError: true,
		},
		{
			name:        "Invalid filter arguments",
			source:      "{{ 'hello'|slice('invalid', 5) }}",
			context:     nil,
			shouldError: true,
		},
		{
			name:        "Invalid attribute type",
			source:      "{{ 42.field }}",
			context:     nil,
			shouldError: true,
		},
		{
			name:        "Array out of bounds",
			source:      "{{ [1, 2, 3][10] }}",
			context:     nil,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.RegisterString(tt.name, tt.source)
			if err != nil && !tt.shouldError {
				t.Fatalf("Unexpected error registering template: %v", err)
			}

			if err == nil {
				// Template registered successfully, try rendering
				_, err = engine.Render(tt.name, tt.context)
			}

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
