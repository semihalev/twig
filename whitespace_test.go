package twig

import (
	"testing"
	"strings"
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

func TestHTMLWhitespacePreservation(t *testing.T) {
	testCases := []struct {
		name               string
		template           string
		preserveWhitespace bool
		prettyOutputHTML   bool
		preserveAttributes bool
		expected           string
	}{
		{
			name: "Default settings - HTML spacing and attribute formatting",
			template: `<div><p>Hello</p><p>World</p></div>
<div class=test id=example>This is a test</div>
<input type=text value=test/>`,
			preserveWhitespace: false,
			prettyOutputHTML:   true,
			preserveAttributes: true,
			expected: `<div> <p>Hello</p> <p>World</p> </div><div class=test id=example>This is a test</div><input type=text value=test/>`,
		},
		{
			name: "Preserve whitespace setting - no formatting",
			template: `<div><p>Hello</p><p>World</p></div>
<div class=test id=example>This is a test</div>
<input type=text value=test/>`,
			preserveWhitespace: true,
			prettyOutputHTML:   true,
			preserveAttributes: true,
			expected: `<div><p>Hello</p><p>World</p></div><div class=test id=example>This is a test</div><input type=text value=test/>`,
		},
		{
			name: "Pretty HTML off - default",
			template: `<div><p>Hello</p><p>World</p></div>
<div class=test id=example>This is a test</div>`,
			preserveWhitespace: false,
			prettyOutputHTML:   false,
			preserveAttributes: true,
			expected: `<div><p>Hello</p><p>World</p></div><div class=test id=example>This is a test</div>`,
		},
		{
			name: "Script tags special handling",
			template: `<script>
  var name = "John";
  var age = 30;
  if (age > 18) {
    console.log("Adult");
  }
</script>`,
			preserveWhitespace: false,
			prettyOutputHTML:   true,
			preserveAttributes: true,
			expected: `<script>var name=John;
  var age = 30;
  if (age > 18) {
    console.log("Adult");
  }
</script>`,
		},
		{
			name: "Style tags special handling",
			template: `<style>
  body { 
    font-family: Arial, sans-serif;
    color: #333;
  }
  .container {
    margin: 0 auto;
  }
</style>`,
			preserveWhitespace: false,
			prettyOutputHTML:   true,
			preserveAttributes: true,
			expected: `<style>body{font-family:Arial,sans-serif;
    color: #333;
  }
  .container {
    margin: 0 auto;
  }
</style>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := New()
			
			// Configure whitespace settings
			engine.SetPreserveWhitespace(tc.preserveWhitespace)
			engine.SetPrettyOutputHTML(tc.prettyOutputHTML)
			engine.SetPreserveAttributes(tc.preserveAttributes)

			// Register the template
			err := engine.RegisterString("test.twig", tc.template)
			if err != nil {
				t.Fatalf("Failed to register template: %v", err)
			}

			// Render with context
			result, err := engine.Render("test.twig", nil)
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			// Normalize line endings for comparison
			result = strings.ReplaceAll(result, "\r\n", "\n")
			expected := strings.ReplaceAll(tc.expected, "\r\n", "\n")

			if result != expected {
				t.Errorf("\nExpected: %q\nGot:      %q\nFor template: %q", expected, result, tc.template)
				
				// Show the differences more clearly by comparing character by character
				t.Log("Difference visualization:")
				minLen := len(result)
				if len(expected) < minLen {
					minLen = len(expected)
				}
				
				// Show first differing character
				for i := 0; i < minLen; i++ {
					if result[i] != expected[i] {
						t.Logf("First difference at position %d: expected '%c' (ASCII %d), got '%c' (ASCII %d)", 
							i, expected[i], expected[i], result[i], result[i])
						break
					}
				}
				
				// Show length difference if applicable
				if len(result) != len(expected) {
					t.Logf("Length difference: expected %d, got %d", len(expected), len(result))
				}
			}
		})
	}
}