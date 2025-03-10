package twig

import (
	"testing"
)

func TestWhitespaceControl(t *testing.T) {
	testCases := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Simple whitespace control - opening tag",
			template: "Hello   {{- name }}",
			expected: "HelloWorld",
		},
		{
			name:     "Simple whitespace control - closing tag",
			template: "{{ name -}}   World",
			expected: "WorldWorld",
		},
		{
			name:     "Simple whitespace control - both sides",
			template: "Hello   {{- name -}}   World",
			expected: "HelloWorldWorld",
		},
		{
			name:     "Block tag whitespace control",
			template: "Hello   {%- if true %}Yes{% endif -%}   World", // Original test case
			expected: "HelloYesWorld",
		},
		{
			name:     "Newlines trimmed",
			template: "Hello\n{{- name }}\n",
			expected: "HelloWorld",
		},
		{
			name:     "Complex mixed examples",
			template: "Hello\n  {{- name -}}  \nWorld",
			expected: "HelloWorldWorld",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := New()

			// Register the template
			err := engine.RegisterString("test.twig", tc.template)
			if err != nil {
				t.Fatalf("Failed to register template: %v", err)
			}

			// Render with context
			result, err := engine.Render("test.twig", map[string]interface{}{
				"name": "World",
			})

			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected: %q, got: %q for template: %q", tc.expected, result, tc.template)
			}
		})
	}
}

func TestSpacelessTag(t *testing.T) {
	testCases := []struct {
		name     string
		template string
		expected string
	}{
		{
			name: "Basic HTML compression",
			template: `{% spaceless %}
				<div>
					<p>Hello</p>
					<p>World</p>
				</div>
			{% endspaceless %}`,
			expected: "<div><p>Hello</p><p>World</p></div>",
		},
		{
			name: "With variables",
			template: `{% spaceless %}
				<div>
					<p>{{ greeting }}</p>
					<p>{{ name }}</p>
				</div>
			{% endspaceless %}`,
			expected: "<div><p>Hello</p><p>World</p></div>",
		},
		{
			name: "Mixed with other tags",
			template: `{% spaceless %}
				<div>
					{% if true %}
						<p>Condition is true</p>
					{% endif %}
				</div>
			{% endspaceless %}`,
			expected: "<div><p>Condition is true</p></div>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := New()

			// Register the template
			err := engine.RegisterString("test.twig", tc.template)
			if err != nil {
				t.Fatalf("Failed to register template: %v", err)
			}

			// Render with context
			result, err := engine.Render("test.twig", map[string]interface{}{
				"greeting": "Hello",
				"name":     "World",
			})

			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected: %q, got: %q for template: %q", tc.expected, result, tc.template)
			}
		})
	}
}
