package twig

import (
	"testing"
)

// Control structures tests
// Consolidated from: if_elseif_test.go, elseif_test.go, fixed_elseif_test.go,
// is_defined_test.go, test_direct_test.go, etc.

// TestOrganizedIfStatement tests basic if statement functionality
func TestOrganizedIfStatement(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple if (true condition)",
			source:   "{% if true %}yes{% endif %}",
			context:  nil,
			expected: "yes",
		},
		{
			name:     "Simple if (false condition)",
			source:   "{% if false %}yes{% endif %}",
			context:  nil,
			expected: "",
		},
		{
			name:     "If with variable (true)",
			source:   "{% if flag %}Enabled{% endif %}",
			context:  map[string]interface{}{"flag": true},
			expected: "Enabled",
		},
		{
			name:     "If with variable (false)",
			source:   "{% if flag %}Enabled{% endif %}",
			context:  map[string]interface{}{"flag": false},
			expected: "",
		},
		{
			name:     "If with comparison",
			source:   "{% if value > 5 %}Greater{% endif %}",
			context:  map[string]interface{}{"value": 10},
			expected: "Greater",
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

// TestIfElseStatement tests if-else functionality
func TestIfElseStatement(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "If-else (true condition)",
			source:   "{% if true %}yes{% else %}no{% endif %}",
			context:  nil,
			expected: "yes",
		},
		{
			name:     "If-else (false condition)",
			source:   "{% if false %}yes{% else %}no{% endif %}",
			context:  nil,
			expected: "no",
		},
		{
			name:     "If-else with variable",
			source:   "{% if flag %}Enabled{% else %}Disabled{% endif %}",
			context:  map[string]interface{}{"flag": false},
			expected: "Disabled",
		},
		{
			name:     "If-else with comparison",
			source:   "{% if value > 5 %}Greater{% else %}Less or equal{% endif %}",
			context:  map[string]interface{}{"value": 3},
			expected: "Less or equal",
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

// TestIfElseIfStatement tests if-elseif-else functionality
func TestIfElseIfStatement(t *testing.T) {
	// Create a new engine
	engine := New()

	// Create a parser to parse a template
	parser := &Parser{}
	source := "{% if score > 90 %}A{% elseif score > 80 %}B{% elseif score > 70 %}C{% else %}F{% endif %}"

	// Parse the template
	node, err := parser.Parse(source)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create a template
	template := &Template{
		name:   "grade",
		source: source,
		nodes:  node,
		env:    engine.environment,
		engine: engine,
	}

	// Register the template
	engine.RegisterTemplate("grade", template)

	// Test cases with different scores
	tests := []struct {
		name     string
		score    int
		expected string
	}{
		{"A grade", 95, "A"},
		{"B grade", 85, "B"},
		{"C grade", 75, "C"},
		{"F grade", 65, "F"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create context with score
			context := map[string]interface{}{"score": test.score}

			// Render the template
			result, err := engine.Render("grade", context)
			if err != nil {
				t.Fatalf("Error rendering template: %v", err)
			}

			if result != test.expected {
				t.Errorf("Expected result %q, got %q", test.expected, result)
			}
		})
	}
}

// TestOrganizedForLoop tests for loop functionality
func TestOrganizedForLoop(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple for loop over array",
			source:   "{% for item in items %}-{{ item }}-{% endfor %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "-a--b--c-",
		},
		{
			name:     "For loop with else (non-empty array)",
			source:   "{% for item in items %}-{{ item }}-{% else %}No items{% endfor %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "-a--b--c-",
		},
		{
			name:     "For loop with else (empty array)",
			source:   "{% for item in items %}-{{ item }}-{% else %}No items{% endfor %}",
			context:  map[string]interface{}{"items": []string{}},
			expected: "No items",
		},
		{
			name:     "For loop over map",
			source:   "{% for key, value in map %}{{ key }}:{{ value }};{% endfor %}",
			context:  map[string]interface{}{"map": map[string]string{"a": "1", "b": "2"}},
			expected: "a:1;b:2;",
		},
		{
			name:     "For loop with loop variable",
			source:   "{% for item in items %}{{ loop.index }}:{{ item }};{% endfor %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "1:a;2:b;3:c;",
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

			// For maps, the order is not guaranteed, so we can't do an exact match for map cases
			if tt.name != "For loop over map" && result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestNestedLoops tests nested loop functionality
func TestNestedLoops(t *testing.T) {
	engine := New()

	// A simpler test for nested loops
	source := `{% for i in [1, 2, 3] %}{% for j in [1, 2, 3] %}({{ i }}:{{ j }}){% endfor %}{% endfor %}`

	err := engine.RegisterString("nested_loops", source)
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	result, err := engine.Render("nested_loops", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	// Check the result contains each combination
	expected := "(1:1)(1:2)(1:3)(2:1)(2:2)(2:3)(3:1)(3:2)(3:3)"
	if result != expected {
		t.Errorf("Expected result to be %q, got %q", expected, result)
	}
}

// TestNegativeRangeStep tests the range function with negative step values
// This test is replaced by TestRangeNegativeStepWorkaround which directly tests the function
func TestNegativeRangeStep(t *testing.T) {
	t.Skip("Skipping: This test is replaced by TestRangeNegativeStepWorkaround")
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
			source:   "{% for i in range(1, 10, 2) %}{{ i }}{% endfor %}",
			expected: "13579",
		},
		{
			name:     "Range function with loop variable",
			source:   "{% for i in range(1, 3) %}{{ loop.index }}:{{ i }};{% endfor %}",
			expected: "1:1;2:2;3:3;",
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

// Helper function to check for substring
func containsSubstring(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
			source:   "{{ flag ? 'Enabled' : 'Disabled' }}",
			context:  map[string]interface{}{"flag": true},
			expected: "Enabled",
		},
		{
			name:     "Ternary with comparison",
			source:   "{{ value > 5 ? 'Greater' : 'Less or equal' }}",
			context:  map[string]interface{}{"value": 3},
			expected: "Less or equal",
		},
		{
			name:     "Ternary with expressions",
			source:   "{{ flag ? a + b : a - b }}",
			context:  map[string]interface{}{"flag": true, "a": 5, "b": 3},
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

// TestIsDefined tests the 'defined' test function
func TestIsDefined(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Defined test (variable exists)",
			source:   "{% if variable is defined %}defined{% else %}not defined{% endif %}",
			context:  map[string]interface{}{"variable": "value"},
			expected: "defined",
		},
		{
			name:     "Defined test (variable does not exist)",
			source:   "{% if variable is defined %}defined{% else %}not defined{% endif %}",
			context:  nil,
			expected: "not defined",
		},
		{
			name:     "Not defined test (variable exists)",
			source:   "{% if variable is not defined %}not defined{% else %}defined{% endif %}",
			context:  map[string]interface{}{"variable": "value"},
			expected: "defined",
		},
		{
			name:     "Not defined test (variable does not exist)",
			source:   "{% if variable is not defined %}not defined{% else %}defined{% endif %}",
			context:  nil,
			expected: "not defined",
		},
		{
			name:     "Not defined test alternative syntax (variable exists)",
			source:   "{% if variable not defined %}not defined{% else %}defined{% endif %}",
			context:  map[string]interface{}{"variable": "value"},
			expected: "defined",
		},
		{
			name:     "Not defined test alternative syntax (variable does not exist)",
			source:   "{% if variable not defined %}not defined{% else %}defined{% endif %}",
			context:  nil,
			expected: "not defined",
		},
		{
			name:     "Strategy name not defined test",
			source:   "{% if strategy_name not defined %}undefined{% else %}defined{% endif %}",
			context:  map[string]interface{}{},
			expected: "undefined",
		},
		{
			name:     "Strategy name not defined test with defined var",
			source:   "{% if strategy_name not defined %}undefined{% else %}defined{% endif %}",
			context:  map[string]interface{}{"strategy_name": "test"},
			expected: "defined",
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
