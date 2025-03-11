package twig

import (
	"testing"
)

// TestAdvancedComparisonOperators tests advanced comparison operators
func TestAdvancedComparisonOperators(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		// Starts with operator
		{
			name:     "Starts with operator - true case",
			source:   "{% if 'Hello world' starts with 'Hello' %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Starts with operator - false case",
			source:   "{% if 'Hello world' starts with 'world' %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "false",
		},
		{
			name:     "Starts with operator with variables",
			source:   "{% if text starts with prefix %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"text": "Hello world", "prefix": "Hello"},
			expected: "true",
		},

		// Ends with operator
		{
			name:     "Ends with operator - true case",
			source:   "{% if 'Hello world' ends with 'world' %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Ends with operator - false case",
			source:   "{% if 'Hello world' ends with 'Hello' %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "false",
		},
		{
			name:     "Ends with operator with variables",
			source:   "{% if text ends with suffix %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"text": "Hello world", "suffix": "world"},
			expected: "true",
		},

		// Matches operator (regex)
		{
			name:     "Matches operator - true case",
			source:   "{% if 'abc123' matches '/^[a-z]+\\d+$/' %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Matches operator - false case",
			source:   "{% if '123abc' matches '/^[a-z]+\\d+$/' %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "false",
		},
		{
			name:     "Matches operator with case insensitive flag",
			source:   "{% if 'HELLO' matches '/hello/i' %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Matches operator with variables",
			source:   "{% if text matches pattern %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"text": "abc123", "pattern": "/^[a-z]+\\d+$/"},
			expected: "true",
		},

		// Combined operators
		{
			name:     "Combined comparison operators (and)",
			source:   "{% if num > 5 and num < 15 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"num": 10},
			expected: "true",
		},
		{
			name:     "Combined comparison operators (or)",
			source:   "{% if num < 5 or num > 15 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"num": 20},
			expected: "true",
		},
		{
			name:     "Mixed string and comparison operators",
			source:   "{% if text starts with 'H' and num > 5 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"text": "Hello", "num": 10},
			expected: "true",
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

// TestComplexConditionalExpressions tests more complex conditional expressions
func TestComplexConditionalExpressions(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		// Nested ternary operators
		{
			name:     "Deeply nested ternary operators",
			source:   "{{ a ? (b ? 'a and b' : 'a not b') : (c ? 'not a but c' : 'neither a nor c') }}",
			context:  map[string]interface{}{"a": true, "b": false, "c": true},
			expected: "a not b",
		},

		// Complex test expressions
		{
			name:     "Complex test expressions with 'is' operator",
			source:   "{% if value is defined and value is not empty and value > 10 %}valid{% else %}invalid{% endif %}",
			context:  map[string]interface{}{"value": 15},
			expected: "valid",
		},

		// Complex expressions with grouping
		{
			name:     "Complex expressions with parentheses for grouping",
			source:   "{% if (a or b) and (c or d) %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"a": false, "b": true, "c": true, "d": false},
			expected: "true",
		},

		// Complex negation
		{
			name:     "Complex negation with 'not' operator",
			source:   "{% if not (a or b) %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"a": false, "b": false},
			expected: "true",
		},

		// Attribute access in conditions
		{
			name:     "Attribute access in conditions",
			source:   "{% if user.active and user.age >= 18 %}allowed{% else %}denied{% endif %}",
			context:  map[string]interface{}{"user": map[string]interface{}{"active": true, "age": 25}},
			expected: "allowed",
		},

		// Nested attribute access in ternary
		{
			name:     "Nested attribute access in ternary",
			source:   "{{ user.active ? user.name : 'Guest' }}",
			context:  map[string]interface{}{"user": map[string]interface{}{"active": true, "name": "John"}},
			expected: "John",
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
