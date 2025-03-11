package twig

import (
	"bytes"
	"testing"
)

func TestVerbatimTag(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Basic verbatim tag",
			template: "{% verbatim %}Hello {{ name }}{% endverbatim %}",
			context:  map[string]interface{}{"name": "World"},
			expected: "Hello {{name}}",
		},
		{
			name:     "Verbatim with multiple variables",
			template: "{% verbatim %}{{ foo }} and {{ bar }}{% endverbatim %}",
			context:  map[string]interface{}{"foo": "value1", "bar": "value2"},
			expected: "{{foo}} and {{bar}}",
		},
		{
			name:     "Verbatim with block tags",
			template: "{% verbatim %}{% if condition %}True{% else %}False{% endif %}{% endverbatim %}",
			context:  map[string]interface{}{"condition": true},
			expected: "{%if condition%}True{%else %}False{%endif %}",
		},
		{
			name:     "Verbatim with mixed content",
			template: "Before {% verbatim %}{{ var }} and {% if x %}{% endif %}{% endverbatim %} After",
			context:  map[string]interface{}{"var": "Value", "x": true},
			expected: "Before {{var}} and {%if x%}{%endif %} After",
		},
		{
			name: "Multiple verbatim blocks",
			template: `First: {% verbatim %}{{ foo }}{% endverbatim %}
                       Second: {% verbatim %}{% if bar %}{% endif %}{% endverbatim %}`,
			context: map[string]interface{}{"foo": "value1", "bar": true},
			expected: `First: {{foo}}
                       Second: {%if bar%}{%endif %}`,
		},
		{
			name:     "Verbatim with comments",
			template: "{% verbatim %}{# This is a comment #}{% endverbatim %}",
			context:  nil,
			expected: "{# This is a comment #}",
		},
		{
			name:     "Nested twig syntax in verbatim",
			template: "{% verbatim %}{% for item in items %}{{ item }}{% endfor %}{% endverbatim %}",
			context:  map[string]interface{}{"items": []string{"a", "b", "c"}},
			expected: "{%for iteminitems%}{{item}}{%endfor %}",
		},
		{
			name:     "HTML content in verbatim",
			template: "{% verbatim %}<div class=\"{{ class }}\">content</div>{% endverbatim %}",
			context:  map[string]interface{}{"class": "highlight"},
			expected: "<div class=\"{{class}}\">content</div>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := New()
			tmpl, err := engine.ParseTemplate(test.template)
			if err != nil {
				t.Fatalf("Template parsing error: %s", err)
			}

			result, err := tmpl.Render(test.context)
			if err != nil {
				t.Fatalf("Template rendering error: %s", err)
			}
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestVerbatimTagErrors(t *testing.T) {
	tests := []struct {
		name     string
		template string
		errorMsg string
	}{
		{
			name:     "Unclosed verbatim tag",
			template: "{% verbatim %}Hello {{ name }}",
			errorMsg: "unclosed verbatim tag",
		},
		{
			name:     "Missing endverbatim closing tag",
			template: "{% verbatim %}Content{% endblock %}",
			errorMsg: "unclosed verbatim tag",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := New()
			_, err := engine.ParseTemplate(test.template)
			if err == nil {
				t.Fatalf("Expected an error but got none")
			}

			if err != nil && !containsString(err.Error(), test.errorMsg) {
				t.Errorf("Expected error containing %q, got: %q", test.errorMsg, err.Error())
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && bytes.Contains([]byte(s), []byte(substr))
}
