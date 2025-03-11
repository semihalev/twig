package twig

import (
	"testing"
)

// TestForLoopEdgeCases tests for loop functionality with edge cases
func TestForLoopEdgeCases(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "For loop with nested negative range",
			source:   "{% for i in range(-2, 2) %}{% for j in range(0, (-3), -1) %}({{ i }},{{ j }}){% endfor %}{% endfor %}",
			context:  nil,
			expected: "(-2,0)(-2,-1)(-2,-2)(-2,-3)(-1,0)(-1,-1)(-1,-2)(-1,-3)(0,0)(0,-1)(0,-2)(0,-3)(1,0)(1,-1)(1,-2)(1,-3)(2,0)(2,-1)(2,-2)(2,-3)",
		},
		{
			name:     "For loop with arithmetic in range bounds",
			source:   "{% for i in range(-(2+3), 0) %}{{ i }}{% endfor %}",
			context:  nil,
			expected: "-5-4-3-2-10",
		},
		{
			name:     "For loop with complex negative step calculation",
			source:   "{% for i in range(10, 0, -1 * 2) %}{{ i }}{% endfor %}",
			context:  nil,
			expected: "1086420",
		},
		{
			name:     "For loop with negative start and positive step",
			source:   "{% set negative = -5 %}{% for i in range(negative, 0) %}{{ i }}{% endfor %}",
			context:  nil,
			expected: "-5-4-3-2-10",
		},
		{
			name:     "For loop with negative range in conditional",
			source:   "{% if true %}{% for i in range(-3, 0) %}{{ i }}{% endfor %}{% endif %}",
			context:  nil,
			expected: "-3-2-10",
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