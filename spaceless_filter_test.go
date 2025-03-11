package twig

import "testing"

func TestSpacelessFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple text with spaceless filter",
			template: `{{ "<div>   <strong>foo</strong>   </div>" | spaceless }}`,
			context:  nil,
			expected: "<div><strong>foo</strong></div>",
		},
		{
			name:     "HTML with newlines and spaces",
			template: `{{ "<div>\n  <p>  Hello  </p>\n  <p>  World  </p>\n</div>" | spaceless }}`,
			context:  nil,
			expected: "<div><p>  Hello  </p><p>  World  </p></div>",
		},
		{
			name:     "Variable with spaceless filter",
			template: `{{ html | spaceless }}`,
			context: map[string]interface{}{
				"html": "<div>\n  <strong>foo</strong>\n</div>",
			},
			expected: "<div><strong>foo</strong></div>",
		},
		{
			name:     "Chain filters ending with spaceless",
			template: `{{ "<div>\n  <p>hello</p>\n</div>" | upper | spaceless }}`,
			context:  nil,
			expected: "<DIV><P>HELLO</P></DIV>",
		},
	}

	engine := New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := engine.ParseTemplate(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			output, err := result.Render(test.context)
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if output != test.expected {
				t.Errorf("Template rendered incorrectly. Expected '%s', got '%s'", test.expected, output)
			}
		})
	}
}
