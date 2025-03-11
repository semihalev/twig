package twig

import (
	"testing"
)

// TestWhitespaceControl tests the whitespace control with dash modifiers
func TestWhitespaceControl(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		source   string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Whitespace control with dash",
			source:   "Before{%- if true -%}content{% endif -%}After",
			context:  nil,
			expected: "BeforecontentAfter",
		},
		{
			name:     "Whitespace control with left dash only",
			source:   "Before {%- if true %}content{% endif %} After",
			context:  nil,
			expected: "Beforecontent After",
		},
		{
			name:     "Whitespace control with right dash only",
			source:   "Before {% if true -%}content{% endif -%} After",
			context:  nil,
			expected: "Before contentAfter",
		},
		{
			name:     "Whitespace control with variable tags",
			source:   "Before {{- 'content' -}} After",
			context:  nil,
			expected: "BeforecontentAfter",
		},
		{
			name:     "Mixed whitespace control",
			source:   "Line1\n   {{- 'content' }}   \n   {% if true -%}   more   {% endif -%}\nLine2",
			context:  nil,
			expected: "Line1content   \n   more   Line2",
		},
		{
			name:     "Newline control with dash",
			source:   "Line1\n{{- 'content' -}}\nLine2",
			context:  nil,
			expected: "Line1contentLine2",
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