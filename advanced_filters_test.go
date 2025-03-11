package twig

import (
	"testing"
)

// TestAdvancedFilters tests advanced filter functionality
// Note: Some of these tests are aspirational and test features that may not be fully
// implemented yet. They serve as a roadmap for future development.
func TestAdvancedFilters(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		// String manipulation filters
		{
			name:     "Slice filter with positive indexes",
			source:   "{{ 'hello world'|slice(0, 5) }}",
			context:  nil,
			expected: "hello",
		},
		{
			name:     "Slice filter with negative start index",
			source:   "{{ 'hello world'|slice(-5, 5) }}",
			context:  nil,
			expected: "world",
		},
		{
			name:     "Slice filter with negative length",
			source:   "{{ 'hello world'|slice(0, -6) }}",
			context:  nil,
			expected: "hello",
		},
		{
			name:     "Title filter",
			source:   "{{ 'hello WORLD'|title }}",
			context:  nil,
			expected: "Hello World",
		},
		{
			name:     "Split filter",
			source:   "{% set parts = 'one,two,three'|split(',') %}{{ parts[1] }}",
			context:  nil,
			expected: "two",
		},

		// HTML filters
		{
			name:     "Escape filter",
			source:   "{{ '<strong>bold</strong>'|escape }}",
			context:  nil,
			expected: "&lt;strong&gt;bold&lt;/strong&gt;",
		},
		{
			name:     "Striptags filter",
			source:   "{{ '<p>paragraph <b>with bold</b> text</p>'|striptags }}",
			context:  nil,
			expected: "paragraph with bold text",
		},

		// Format filters
		{
			name:     "Format filter with numbered placeholders",
			source:   "{{ 'Hello %s, you have %d new messages'|format('John', 5) }}",
			context:  nil,
			expected: "Hello John, you have 5 new messages",
		},
		{
			name:     "Raw filter to prevent escaping",
			source:   "{{ '<div>content</div>'|raw }}",
			context:  nil,
			expected: "<div>content</div>",
		},

		// Array filters
		{
			name:     "Join filter",
			source:   "{{ ['apple', 'banana', 'cherry']|join(', ') }}",
			context:  nil,
			expected: "apple, banana, cherry",
		},
		{
			name:     "Join filter with variable",
			source:   "{{ items|join(', ') }}",
			context:  map[string]interface{}{"items": []string{"one", "two", "three"}},
			expected: "one, two, three",
		},
		{
			name:     "First filter",
			source:   "{{ [10, 20, 30]|first }}",
			context:  nil,
			expected: "10",
		},
		{
			name:     "Last filter",
			source:   "{{ [10, 20, 30]|last }}",
			context:  nil,
			expected: "30",
		},
		{
			name:     "Reverse filter",
			source:   "{{ ['a', 'b', 'c']|reverse|join(',') }}",
			context:  nil,
			expected: "c,b,a",
		},

		// Number filters
		{
			name:     "Abs filter",
			source:   "{{ (-50)|abs }}",
			context:  nil,
			expected: "50",
		},
		{
			name:     "Round filter (default)",
			source:   "{{ 3.7|round }}",
			context:  nil,
			expected: "4",
		},
		{
			name:     "Round filter (floor)",
			source:   "{{ 3.7|round(0, 'floor') }}",
			context:  nil,
			expected: "3",
		},
		{
			name:     "Round filter (ceil)",
			source:   "{{ 3.2|round(0, 'ceil') }}",
			context:  nil,
			expected: "4",
		},
		{
			name:     "Round filter with precision",
			source:   "{{ 3.1415926|round(4) }}",
			context:  nil,
			expected: "3.1416",
		},

		// Chained filters
		{
			name:     "Multiple chained filters",
			source:   "{{ 'HELLO WORLD'|lower|capitalize|replace('World', 'everyone') }}",
			context:  nil,
			expected: "Hello everyone",
		},
		{
			name:     "Complex filter chain with arrays",
			source:   "{{ ['B', 'A', 'C']|sort|reverse|join('-') }}",
			context:  nil,
			expected: "C-B-A",
		},
		{
			name:     "Filter with dynamic argument",
			source:   "{{ name|slice(0, length) }}",
			context:  map[string]interface{}{"name": "Elizabeth", "length": 4},
			expected: "Eliz",
		},
		// Note: Arrow function syntax requires parser changes and is not supported yet
		/*{
			name:     "Map filter with arrow function syntax",
			source:   "{{ items|map(item => item * 2)|join(', ') }}",
			context:  map[string]interface{}{"items": []int{1, 2, 3}},
			expected: "2, 4, 6",
		},
		{
			name:   "Filter with nested expression",
			source: "{{ items|sort((a, b) => a.age <=> b.age)|map(item => item.name)|join(', ') }}",
			context: map[string]interface{}{
				"items": []map[string]interface{}{
					{"name": "John", "age": 30},
					{"name": "Alice", "age": 25},
					{"name": "Bob", "age": 35},
				},
			},
			expected: "Alice, John, Bob",
		},*/
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

// TestFilterCombinations tests various combinations of filters with other template features
// Note: Some of these tests are aspirational and test features that may not be fully
// implemented yet. They serve as a roadmap for future development.
func TestFilterCombinations(t *testing.T) {
	// Enable debug logging
	SetDebugLevel(DebugVerbose)
	defer SetDebugLevel(DebugOff)

	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		// Filters in conditionals
		{
			name:     "Filter in if condition",
			source:   "{% if name|length > 5 %}long{% else %}short{% endif %}",
			context:  map[string]interface{}{"name": "Elizabeth"},
			expected: "long",
		},
		{
			name:     "Multiple filters in if condition",
			source:   "{% if name|lower|replace('e', '')|length <= 3 %}yes{% else %}no{% endif %}",
			context:  map[string]interface{}{"name": "Jane"},
			expected: "yes",
		},

		// Filters in loop expressions
		{
			name:     "Filter in for loop",
			source:   "{% for item in items|sort %}{{ item }}{% endfor %}",
			context:  map[string]interface{}{"items": []string{"b", "c", "a"}},
			expected: "abc",
		},
		// Additional test variations to debug the sort filter in for loops
		{
			name:     "Direct sort with simplified output",
			source:   "{% for item in ['b', 'c', 'a']|sort %}{{ item }}{% endfor %}",
			context:  nil,
			expected: "abc",
		},
		{
			name:     "Sort filter with explicit join",
			source:   "{% set sorted = items|sort %}{% for item in sorted %}{{ item }}{% endfor %}",
			context:  map[string]interface{}{"items": []string{"b", "c", "a"}},
			expected: "abc",
		},
		{
			name:     "Filter with loop variables",
			source:   "{% for i in range(1, 3) %}{{ i|abs }}{% endfor %}",
			context:  nil,
			expected: "123",
		},

		// Filters with attribute access
		{
			name:     "Filter with attribute access",
			source:   "{{ user.name|upper }}",
			context:  map[string]interface{}{"user": map[string]interface{}{"name": "john"}},
			expected: "JOHN",
		},
		{
			name:     "Filter result with attribute access",
			source:   "{{ ('name:' ~ user.name)|split(':')|last }}",
			context:  map[string]interface{}{"user": map[string]interface{}{"name": "john"}},
			expected: "john",
		},

		// Filters in ternary expressions
		{
			name:     "Filter in ternary condition",
			source:   "{{ name|length > 5 ? 'long' : 'short' }}",
			context:  map[string]interface{}{"name": "Elizabeth"},
			expected: "long",
		},
		{
			name:     "Filters in both branches of ternary",
			source:   "{{ condition ? name|upper : name|lower }}",
			context:  map[string]interface{}{"condition": true, "name": "John"},
			expected: "JOHN",
		},

		// Nested filter applications
		{
			name:     "Nested filter application",
			source:   "{{ text|replace('world', name|capitalize) }}",
			context:  map[string]interface{}{"text": "hello world", "name": "john"},
			expected: "hello John",
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
