package twig

import (
	"testing"
)

// TestDirectNegativeSteps tests the range function with direct negative literals
// This should now work with our parser improvements
func TestDirectNegativeSteps(t *testing.T) {
	engine := New()

	// Test with direct negative step
	err := engine.RegisterString("range_direct_neg", `{% for i in range(5, 1, -1) %}{{ i }}{% endfor %}`)
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	result, err := engine.Render("range_direct_neg", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	expected := "54321"
	if result != expected {
		t.Fatalf("Expected: %q, Got: %q", expected, result)
	}

	// Test with direct negative step and more complex expressions
	err = engine.RegisterString("range_complex_neg", `{% for i in range(10, 0, (-2)) %}{{ i }}{% endfor %}`)
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	result, err = engine.Render("range_complex_neg", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	expected = "1086420"
	if result != expected {
		t.Fatalf("Expected: %q, Got: %q", expected, result)
	}
}

// TestNegativeNumbersInExpressions tests more complex expressions with negative numbers
func TestNegativeNumbersInExpressions(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Negative number in arithmetic",
			source:   "{{ 10 + (-5) }}",
			context:  nil,
			expected: "5",
		},
		{
			name:     "Multiple negative numbers",
			source:   "{{ (-2) * (-3) }}",
			context:  nil,
			expected: "6",
		},
		{
			name:     "Nested expressions with negatives",
			source:   "{{ (-(5 + 3)) + 10 }}",
			context:  nil,
			expected: "2",
		},
		{
			name:     "Negative with filter",
			source:   "{{ (-123.45)|number_format(1, '.', ',') }}",
			context:  nil,
			expected: "-123.5",
		},
		{
			name:     "Complex filter args with negatives",
			source:   "{{ 1234.5678|round((-1) * 2) }}",
			context:  nil,
			expected: "1200",
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