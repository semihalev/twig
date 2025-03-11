package twig

import (
	"strings"
	"testing"
)

// TestSimpleControlStructures tests simple control structures (if, for)
func TestSimpleControlStructures(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		// If statements
		{
			name:     "Simple if (true)",
			source:   "{% if true %}yes{% endif %}",
			context:  nil,
			expected: "yes",
		},
		{
			name:     "Simple if (false)",
			source:   "{% if false %}yes{% endif %}",
			context:  nil,
			expected: "",
		},
		{
			name:     "If-else (true)",
			source:   "{% if true %}yes{% else %}no{% endif %}",
			context:  nil,
			expected: "yes",
		},
		{
			name:     "If-else (false)",
			source:   "{% if false %}yes{% else %}no{% endif %}",
			context:  nil,
			expected: "no",
		},
		{
			name:     "If-elseif-else (first true)",
			source:   "{% if true %}1{% elseif true %}2{% else %}3{% endif %}",
			context:  nil,
			expected: "1",
		},
		{
			name:     "If-elseif-else (second true)",
			source:   "{% if false %}1{% elseif true %}2{% else %}3{% endif %}",
			context:  nil,
			expected: "2",
		},
		{
			name:     "If-elseif-else (all false)",
			source:   "{% if false %}1{% elseif false %}2{% else %}3{% endif %}",
			context:  nil,
			expected: "3",
		},
		{
			name:     "Nested if statements",
			source:   "{% if true %}outer{% if false %}inner-if{% else %}inner-else{% endif %}{% endif %}",
			context:  nil,
			expected: "outerinner-else",
		},
		{
			name:     "If with variable condition",
			source:   "{% if value %}yes{% else %}no{% endif %}",
			context:  map[string]interface{}{"value": true},
			expected: "yes",
		},
		{
			name:     "If with complex condition (and)",
			source:   "{% if value and otherValue %}yes{% else %}no{% endif %}",
			context:  map[string]interface{}{"value": true, "otherValue": true},
			expected: "yes",
		},
		{
			name:     "If with complex condition (or)",
			source:   "{% if value or otherValue %}yes{% else %}no{% endif %}",
			context:  map[string]interface{}{"value": false, "otherValue": true},
			expected: "yes",
		},
		{
			name:     "If with is test",
			source:   "{% if value is defined %}yes{% else %}no{% endif %}",
			context:  map[string]interface{}{"value": "something"},
			expected: "yes",
		},
		{
			name:     "If with complex condition (parentheses)",
			source:   "{% if (a or b) and c %}yes{% else %}no{% endif %}",
			context:  map[string]interface{}{"a": true, "b": false, "c": true},
			expected: "yes",
		},

		// For loops
		{
			name:     "For loop with array",
			source:   "{% for item in items %}{{ item }}{% endfor %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "abc",
		},
		{
			name:     "For loop with array literal",
			source:   "{% for item in ['a', 'b', 'c'] %}{{ item }}{% endfor %}",
			context:  nil,
			expected: "abc",
		},
		{
			name:     "For loop with map",
			source:   "{% for key, value in data %}{{ key }}:{{ value }};{% endfor %}",
			context:  map[string]interface{}{"data": map[string]interface{}{"a": 1, "b": 2, "c": 3}},
			expected: "a:1;b:2;c:3;",  // This is just a reference; actual map iteration order is checked separately
		},
		{
			name:     "For loop with loop variable",
			source:   "{% for item in items %}{{ loop.index }}:{{ item }};{% endfor %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "1:a;2:b;3:c;",
		},
		{
			name:     "Nested for loops",
			source:   "{% for i in items %}{% for j in items %}({{ i }},{{ j }}){% endfor %}{% endfor %}",
			context:  map[string]interface{}{"items": []string{"a", "b"}},
			expected: "(a,a)(a,b)(b,a)(b,b)",
		},
		{
			name:     "For loop with if condition",
			source:   "{% for item in items %}{% if loop.first %}first:{% endif %}{{ item }}{% endfor %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "first:abc",
		},
		{
			name:     "For loop with else (non-empty)",
			source:   "{% for item in items %}{{ item }}{% else %}empty{% endfor %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "abc",
		},
		{
			name:     "For loop with else (empty)",
			source:   "{% for item in items %}{{ item }}{% else %}empty{% endfor %}",
			context:  map[string]interface{}{"items": []string{}},
			expected: "empty",
		},
		{
			name:     "For loop with array access",
			source:   "{% for i in range(0, 2) %}{{ items[i] }}{% endfor %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "abc",
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

			// Special handling for map iteration which is non-deterministic in Go
			if tt.name == "For loop with map" {
				// Check that the result contains all the expected key-value pairs
				expectedPairs := []string{"a:1;", "b:2;", "c:3;"}
				for _, pair := range expectedPairs {
					if !strings.Contains(result, pair) {
						t.Errorf("Expected result to contain %q, but got: %q", pair, result)
					}
				}
			} else if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestRangeFunctionInForLoop tests the range function directly in a for loop
func TestRangeFunctionInForLoop(t *testing.T) {
	engine := New()

	// Test the actual rendering
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "Simple range function in for loop",
			source:   "{% for i in range(1, 3) %}{{ i }}{% endfor %}",
			expected: "123",
		},
		{
			name:     "Range function with step in for loop",
			source:   "{% for i in range(1, 9, 2) %}{{ i }}{% endfor %}",
			expected: "13579",
		},
		{
			name:     "Range function with loop variable",
			source:   "{% for i in range(1, 3) %}{{ loop.index }}:{{ i }};{% endfor %}",
			expected: "1:1;2:2;3:3;",
		},
		{
			name:     "Range function with direct negative start",
			source:   "{% for i in range(-3, 0) %}{{ i }}{% endfor %}",
			expected: "-3-2-10",
		},
		{
			name:     "Range function with parenthesized negative start",
			source:   "{% for i in range((-5), 0) %}{{ i }}{% endfor %}",
			expected: "-5-4-3-2-10",
		},
		{
			name:     "Range with negative end value",
			source:   "{% for i in range(0, -3, -1) %}{{ i }}{% endfor %}",
			expected: "0-1-2-3",
		},
		{
			name:     "Range with complex negative literals",
			source:   "{% for i in range((-10), (-5)) %}{{ i }}{% endfor %}",
			expected: "-10-9-8-7-6-5",
		},
		{
			name:     "Range with arithmetic expressions for bounds",
			source:   "{% for i in range(0-5, 0) %}{{ i }}{% endfor %}",
			expected: "-5-4-3-2-10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.RegisterString("test", tt.source)
			if err != nil {
				t.Fatalf("Error registering template: %v", err)
			}

			result, err := engine.Render("test", nil)
			if err != nil {
				t.Fatalf("Error rendering template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestConditionalExpressions tests conditional expressions (ternary)
func TestConditionalExpressions(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple ternary (true)",
			source:   "{{ true ? 'yes' : 'no' }}",
			context:  nil,
			expected: "yes",
		},
		{
			name:     "Simple ternary (false)",
			source:   "{{ false ? 'yes' : 'no' }}",
			context:  nil,
			expected: "no",
		},
		{
			name:     "Ternary with variable",
			source:   "{{ value ? 'yes' : 'no' }}",
			context:  map[string]interface{}{"value": true},
			expected: "yes",
		},
		{
			name:     "Ternary with complex condition",
			source:   "{{ (a or b) and c ? 'yes' : 'no' }}",
			context:  map[string]interface{}{"a": true, "b": false, "c": true},
			expected: "yes",
		},
		{
			name:     "Nested ternary (outer true)",
			source:   "{{ true ? (true ? '1' : '2') : '3' }}",
			context:  nil,
			expected: "1",
		},
		{
			name:     "Nested ternary (outer false)",
			source:   "{{ false ? '1' : (true ? '2' : '3') }}",
			context:  nil,
			expected: "2",
		},
		{
			name:     "Ternary with arithmetic",
			source:   "{{ 5 > 2 ? 5 + 3 : 2 - 1 }}",
			context:  nil,
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