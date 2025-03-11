package twig

import (
	"strings"
	"testing"
)

func TestSpacelessTag(t *testing.T) {
	engine := New()

	// Test simple case with spaceless tag
	template1 := `{% spaceless %}
<div>
  <strong>foo</strong>
</div>
{% endspaceless %}`

	result, err := engine.ParseTemplate(template1)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	output, err := result.Render(nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "<div><strong>foo</strong></div>"
	if normalizeSpacelessOutput(output) != normalizeSpacelessOutput(expected) {
		t.Errorf("Test 1 failed. Expected '%s', got '%s'", expected, output)
	}

	// Test with nested tags
	template2 := `{% spaceless %}
<div>
  <p>
    <span>Hello</span>
    <span>World</span>
  </p>
</div>
{% endspaceless %}`

	result, err = engine.ParseTemplate(template2)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	output, err = result.Render(nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected = "<div><p><span>Hello</span><span>World</span></p></div>"
	if normalizeSpacelessOutput(output) != normalizeSpacelessOutput(expected) {
		t.Errorf("Test 2 failed. Expected '%s', got '%s'", expected, output)
	}

	// Test with variable content
	template3 := `{% spaceless %}
<div>
  {{ content }}
</div>
{% endspaceless %}`

	result, err = engine.ParseTemplate(template3)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	context := map[string]interface{}{
		"content": "<strong>Important</strong>",
	}

	output, err = result.Render(context)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected = "<div><strong>Important</strong></div>"
	if normalizeSpacelessOutput(output) != normalizeSpacelessOutput(expected) {
		t.Errorf("Test 3 failed. Expected '%s', got '%s'", expected, output)
	}
}

func normalizeSpacelessOutput(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
