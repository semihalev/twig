package twig

import (
	"strings"
	"testing"
)

func TestBasicOperators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple addition",
			source:   "{{ 1 + 2 }}",
			expected: "3",
		},
		{
			name:     "Simple subtraction",
			source:   "{{ 5 - 2 }}",
			expected: "3",
		},
		{
			name:     "Simple multiplication",
			source:   "{{ 2 * 3 }}",
			expected: "6",
		},
		{
			name:     "Simple division",
			source:   "{{ 6 / 2 }}",
			expected: "3",
		},
		{
			name:     "Simple modulo",
			source:   "{{ 7 % 3 }}",
			expected: "1",
		},
		{
			name:     "String concatenation",
			source:   "{{ 'hello' ~ ' ' ~ 'world' }}",
			expected: "hello world",
		},
		{
			name:     "Complex expression",
			source:   "{{ 1 + 2 * 3 }}",
			expected: "7",
		},
		{
			name:     "Parenthesized expression",
			source:   "{{ (1 + 2) * 3 }}",
			expected: "9",
		},
		{
			name:     "Variable addition",
			source:   "{{ a + b }}",
			context:  map[string]interface{}{"a": 5, "b": 3},
			expected: "8",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := New()
			template, err := engine.ParseTemplate(test.source)
			if err != nil {
				t.Fatalf("Error parsing template: %s", err)
			}

			output, err := template.Render(test.context)
			if err != nil {
				t.Fatalf("Error rendering template: %s", err)
			}

			if strings.TrimSpace(output) != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, strings.TrimSpace(output))
			}
		})
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Equal",
			source:   "{% if 1 == 1 %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "Not equal",
			source:   "{% if 1 != 2 %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "Less than",
			source:   "{% if 1 < 2 %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "Greater than",
			source:   "{% if 2 > 1 %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "Less than or equal",
			source:   "{% if 1 <= 1 %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "Greater than or equal",
			source:   "{% if 1 >= 1 %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "Contains",
			source:   "{% if 'hello' in 'hello world' %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "Not contains",
			source:   "{% if 'xyz' not in 'hello world' %}true{% else %}false{% endif %}",
			expected: "true",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := New()
			template, err := engine.ParseTemplate(test.source)
			if err != nil {
				t.Fatalf("Error parsing template: %s", err)
			}

			output, err := template.Render(test.context)
			if err != nil {
				t.Fatalf("Error rendering template: %s", err)
			}

			if strings.TrimSpace(output) != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, strings.TrimSpace(output))
			}
		})
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple AND",
			source:   "{% if true and true %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "Simple OR",
			source:   "{% if true or false %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "Simple NOT",
			source:   "{% if not false %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "AND with first operand false (short-circuit)",
			source:   "{% if false and nonexistentvar %}true{% else %}false{% endif %}",
			expected: "false",
		},
		{
			name:     "OR with first operand true (short-circuit)",
			source:   "{% if true or nonexistentvar %}true{% else %}false{% endif %}",
			expected: "true",
		},
		{
			name:     "AND with variable existence check (short-circuit)",
			source:   "{% if foo is defined and foo > 5 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{},
			expected: "false",
		},
		{
			name:     "AND with variable existence check (positive case)",
			source:   "{% if foo is defined and foo > 5 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"foo": 10},
			expected: "true",
		},
		{
			name:     "AND with variable existence check (negative case)",
			source:   "{% if foo is defined and foo > 5 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"foo": 3},
			expected: "false",
		},
		{
			name:     "OR with variable existence check",
			source:   "{% if foo is not defined or foo > 5 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{},
			expected: "true",
		},
		{
			name:     "Multiple conditions with AND",
			source:   "{% if foo is defined and bar is defined and foo > 5 and bar < 10 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"foo": 7, "bar": 3},
			expected: "true",
		},
		{
			name:     "Multiple conditions with AND (short-circuit)",
			source:   "{% if foo is defined and bar is defined and foo > 5 and bar < 10 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{},
			expected: "false",
		},
		{
			name:     "Complex mixed conditions",
			source:   "{% if (foo is defined and foo > 5) or (bar is defined and bar < 10) %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"bar": 5},
			expected: "true",
		},
		{
			name:     "Nested conditions with undefined variables",
			source:   "{% if foo is defined and (bar is defined and bar > foo) %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"foo": 3},
			expected: "false",
		},
		{
			name:     "Nested conditions with all variables",
			source:   "{% if foo is defined and (bar is defined and bar > foo) %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"foo": 3, "bar": 5},
			expected: "true",
		},
		{
			name:     "Multiple operations with precedence",
			source:   "{% if 1 < 2 and 3 > 2 or 5 < 4 %}true{% else %}false{% endif %}",
			expected: "true",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := New()
			template, err := engine.ParseTemplate(test.source)
			if err != nil {
				t.Fatalf("Error parsing template: %s", err)
			}

			output, err := template.Render(test.context)
			if err != nil {
				t.Fatalf("Error rendering template: %s", err)
			}

			if strings.TrimSpace(output) != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, strings.TrimSpace(output))
			}
		})
	}
}
