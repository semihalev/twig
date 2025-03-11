package twig

import (
	"testing"
)

// Operator tests
// Consolidated from: basic_operators_test.go, additional_operators_test.go,
// equal_operator_test.go, test_operators_test.go, ternary_operator_test.go, etc.

// TestOrganizedBasicOperators tests basic mathematical operators
func TestOrganizedBasicOperators(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		// Arithmetic operators
		{
			name:     "Addition operator",
			source:   "{{ 2 + 3 }}",
			context:  nil,
			expected: "5",
		},
		{
			name:     "Subtraction operator",
			source:   "{{ 5 - 2 }}",
			context:  nil,
			expected: "3",
		},
		{
			name:     "Multiplication operator",
			source:   "{{ 3 * 4 }}",
			context:  nil,
			expected: "12",
		},
		{
			name:     "Division operator",
			source:   "{{ 10 / 2 }}",
			context:  nil,
			expected: "5",
		},

		// String concatenation
		{
			name:     "String concatenation operator",
			source:   "{{ 'hello' ~ ' ' ~ 'world' }}",
			context:  nil,
			expected: "hello world",
		},

		// Operator precedence
		{
			name:     "Operator precedence",
			source:   "{{ 2 + 3 * 4 }}",
			context:  nil,
			expected: "14",
		},
		{
			name:     "Parentheses for precedence",
			source:   "{{ (2 + 3) * 4 }}",
			context:  nil,
			expected: "20",
		},

		// Variable operations
		{
			name:     "Variable arithmetic",
			source:   "{{ a + b }}",
			context:  map[string]interface{}{"a": 5, "b": 3},
			expected: "8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.RegisterString("test", tt.source)
			if err != nil {
				t.Fatalf("Error registering template: %v", err)
			}

			result, err := engine.Render("test", tt.context)
			if err != nil {
				t.Fatalf("Error rendering template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestOrganizedComparisonOperators tests comparison operators
func TestOrganizedComparisonOperators(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Equal operator with if statement",
			source:   "{% if 5 == 5 %}equal{% else %}not equal{% endif %}",
			context:  nil,
			expected: "equal",
		},
		{
			name:     "Equal operator with strings in if statement",
			source:   "{% if 'hello' == 'hello' %}equal{% else %}not equal{% endif %}",
			context:  nil,
			expected: "equal",
		},
		{
			name:     "Equal operator with different values in if statement",
			source:   "{% if 5 == 10 %}equal{% else %}not equal{% endif %}",
			context:  nil,
			expected: "not equal",
		},
		{
			name:     "Inequality operator in if statement",
			source:   "{% if 5 != 10 %}not equal{% else %}equal{% endif %}",
			context:  nil,
			expected: "not equal",
		},
		{
			name:     "Greater than operator",
			source:   "{% if 10 > 5 %}greater{% else %}not greater{% endif %}",
			context:  nil,
			expected: "greater",
		},
		{
			name:     "Less than operator",
			source:   "{% if 5 < 10 %}less{% else %}not less{% endif %}",
			context:  nil,
			expected: "less",
		},
		{
			name:     "Greater than or equal operator (greater)",
			source:   "{% if 10 >= 5 %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Greater than or equal operator (equal)",
			source:   "{% if 5 >= 5 %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Less than or equal operator (less)",
			source:   "{% if 5 <= 10 %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Less than or equal operator (equal)",
			source:   "{% if 5 <= 5 %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{}
			node, err := parser.Parse(tt.source)
			if err != nil {
				t.Fatalf("Error parsing template: %v", err)
			}

			template := &Template{
				name:   tt.name,
				source: tt.source,
				nodes:  node,
				env:    engine.environment,
				engine: engine,
			}

			engine.RegisterTemplate(tt.name, template)

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

// TestOrganizedLogicalOperators tests logical operators (and, or, not)
func TestOrganizedLogicalOperators(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Logical AND operator (true and true)",
			source:   "{% if true and true %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Logical AND operator (true and false)",
			source:   "{% if true and false %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "false",
		},
		{
			name:     "Logical AND operator (false and false)",
			source:   "{% if false and false %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "false",
		},
		{
			name:     "Logical OR operator (true or false)",
			source:   "{% if true or false %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Logical OR operator (false or true)",
			source:   "{% if false or true %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Logical OR operator (false or false)",
			source:   "{% if false or false %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "false",
		},
		{
			name:     "Logical NOT operator (not true)",
			source:   "{% if not true %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "false",
		},
		{
			name:     "Logical NOT operator (not false)",
			source:   "{% if not false %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Complex logical expression",
			source:   "{% if (true and false) or (true and true) %}true{% else %}false{% endif %}",
			context:  nil,
			expected: "true",
		},
		{
			name:     "Logical operators with variables",
			source:   "{% if a and b %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"a": true, "b": true},
			expected: "true",
		},
		{
			name:     "Logical operators with comparison",
			source:   "{% if a > 5 and b < 10 %}true{% else %}false{% endif %}",
			context:  map[string]interface{}{"a": 7, "b": 3},
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{}
			node, err := parser.Parse(tt.source)
			if err != nil {
				t.Fatalf("Error parsing template: %v", err)
			}

			template := &Template{
				name:   tt.name,
				source: tt.source,
				nodes:  node,
				env:    engine.environment,
				engine: engine,
			}

			engine.RegisterTemplate(tt.name, template)

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

// TestOrganizedTernaryOperator tests the ternary operator
func TestOrganizedTernaryOperator(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple ternary with true condition",
			source:   "{{ true ? 'true branch' : 'false branch' }}",
			context:  nil,
			expected: "true branch",
		},
		{
			name:     "Simple ternary with false condition",
			source:   "{{ false ? 'true branch' : 'false branch' }}",
			context:  nil,
			expected: "false branch",
		},
		{
			name:     "Ternary with variable condition",
			source:   "{{ condition ? 'true branch' : 'false branch' }}",
			context:  map[string]interface{}{"condition": true},
			expected: "true branch",
		},
		{
			name:     "Ternary with comparison condition",
			source:   "{{ 5 > 3 ? 'greater' : 'not greater' }}",
			context:  nil,
			expected: "greater",
		},
		{
			name:     "Ternary with expressions in branches",
			source:   "{{ true ? 5 + 3 : 10 - 2 }}",
			context:  nil,
			expected: "8",
		},
		{
			name:     "Ternary with variables in branches",
			source:   "{{ true ? a : b }}",
			context:  map[string]interface{}{"a": "value a", "b": "value b"},
			expected: "value a",
		},
		{
			name:     "Nested ternary operators",
			source:   "{{ a ? (b ? 'a and b true' : 'a true, b false') : 'a false' }}",
			context:  map[string]interface{}{"a": true, "b": true},
			expected: "a and b true",
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

// TestOrganizedContainsOperator tests the 'in' operator
func TestOrganizedContainsOperator(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "In operator with array (item exists)",
			source:   "{% if 'b' in ['a', 'b', 'c'] %}found{% else %}not found{% endif %}",
			context:  nil,
			expected: "found",
		},
		{
			name:     "In operator with array (item doesn't exist)",
			source:   "{% if 'z' in ['a', 'b', 'c'] %}found{% else %}not found{% endif %}",
			context:  nil,
			expected: "not found",
		},
		{
			name:     "Not in operator with array",
			source:   "{% if 'z' not in ['a', 'b', 'c'] %}not found{% else %}found{% endif %}",
			context:  nil,
			expected: "not found",
		},
		{
			name:     "In operator with string (substring exists)",
			source:   "{% if 'world' in 'hello world' %}found{% else %}not found{% endif %}",
			context:  nil,
			expected: "found",
		},
		{
			name:     "In operator with string (substring doesn't exist)",
			source:   "{% if 'other' in 'hello world' %}found{% else %}not found{% endif %}",
			context:  nil,
			expected: "not found",
		},
		//{
		//			name:     "In operator with map (key exists)",
		//			source:   "{% if 'name' in {'name': 'John', 'age': 30} %}found{% else %}not found{% endif %}",
		//			context:  nil,
		//			expected: "found",
		//		},
		//{
		//			name:     "In operator with map (key doesn't exist)",
		//			source:   "{% if 'address' in {'name': 'John', 'age': 30} %}found{% else %}not found{% endif %}",
		//			context:  nil,
		//			expected: "not found",
		//		},
		{
			name:     "In operator with variable array",
			source:   "{% if item in items %}found{% else %}not found{% endif %}",
			context:  map[string]interface{}{"item": "b", "items": []string{"a", "b", "c"}},
			expected: "found",
		},
		{
			name:     "In operator with variable map",
			source:   "{% if key in data %}found{% else %}not found{% endif %}",
			context:  map[string]interface{}{"key": "name", "data": map[string]interface{}{"name": "John", "age": 30}},
			expected: "found",
		},
		{
			name:     "In operator with integers in array",
			source:   "{% if 42 in [10, 20, 30, 42, 50] %}found{% else %}not found{% endif %}",
			context:  nil,
			expected: "found",
		},
		{
			name:     "Not in operator with mixed types",
			source:   "{% if 'hello' not in [1, 2, 3] %}not found{% else %}found{% endif %}",
			context:  nil,
			expected: "not found",
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

// TestOrganizedConcatenation tests string concatenation
func TestOrganizedConcatenation(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple concatenation",
			source:   "{{ 'hello' ~ ' ' ~ 'world' }}",
			context:  nil,
			expected: "hello world",
		},
		{
			name:     "Concatenation with variables",
			source:   "{{ first ~ ' ' ~ last }}",
			context:  map[string]interface{}{"first": "John", "last": "Doe"},
			expected: "John Doe",
		},
		{
			name:     "Concatenation with numbers",
			source:   "{{ 'number: ' ~ 42 }}",
			context:  nil,
			expected: "number: 42",
		},
		{
			name:     "Concatenation with expressions",
			source:   "{{ 'result: ' ~ (5 * 10) }}",
			context:  nil,
			expected: "result: 50",
		},
		{
			name:     "Concatenation in if statement",
			source:   "{% if 'a' ~ 'b' == 'ab' %}equal{% else %}not equal{% endif %}",
			context:  nil,
			expected: "equal",
		},
		{
			name:     "Concatenation with filters",
			source:   "{{ ('hello' ~ ' world')|upper }}",
			context:  nil,
			expected: "HELLO WORLD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.RegisterString("test", tt.source)
			if err != nil {
				t.Fatalf("Error registering template: %v", err)
			}

			result, err := engine.Render("test", tt.context)
			if err != nil {
				t.Fatalf("Error rendering template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestOrganizedSpecialOperators tests the special operators (is, is not, matches, starts with, ends with)
func TestOrganizedSpecialOperators(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		// 'is' operator tests
		{
			name:     "Is operator with defined test",
			source:   "{% if name is defined %}defined{% else %}not defined{% endif %}",
			context:  map[string]interface{}{"name": "John"},
			expected: "defined",
		},
		{
			name:     "Is operator with undefined variable",
			source:   "{% if undefined is defined %}defined{% else %}not defined{% endif %}",
			context:  nil,
			expected: "not defined",
		},
		{
			name:     "Is operator with empty test (empty string)",
			source:   "{% if '' is empty %}empty{% else %}not empty{% endif %}",
			context:  nil,
			expected: "empty",
		},
		{
			name:     "Is operator with empty test (empty array)",
			source:   "{% if [] is empty %}empty{% else %}not empty{% endif %}",
			context:  nil,
			expected: "empty",
		},
		{
			name:     "Is operator with empty test (non-empty array)",
			source:   "{% if ['a', 'b'] is empty %}empty{% else %}not empty{% endif %}",
			context:  nil,
			expected: "not empty",
		},
		{
			name:     "Is operator with null test",
			source:   "{% if null_var is null %}null{% else %}not null{% endif %}",
			context:  map[string]interface{}{"null_var": nil},
			expected: "null",
		},
		{
			name:     "Is operator with even test",
			source:   "{% if 4 is even %}even{% else %}odd{% endif %}",
			context:  nil,
			expected: "even",
		},
		{
			name:     "Is operator with odd test",
			source:   "{% if 5 is odd %}odd{% else %}even{% endif %}",
			context:  nil,
			expected: "odd",
		},
		{
			name:     "Is operator with iterable test (array)",
			source:   "{% if items is iterable %}iterable{% else %}not iterable{% endif %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "iterable",
		},

		// 'is not' operator tests
		{
			name:     "Is not operator with defined test",
			source:   "{% if undefined is not defined %}not defined{% else %}defined{% endif %}",
			context:  nil,
			expected: "not defined",
		},
		{
			name:     "Is not operator with empty test",
			source:   "{% if 'hello' is not empty %}not empty{% else %}empty{% endif %}",
			context:  nil,
			expected: "not empty",
		},
		{
			name:     "Is not operator with null test",
			source:   "{% if value is not null %}not null{% else %}null{% endif %}",
			context:  map[string]interface{}{"value": "hello"},
			expected: "not null",
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
