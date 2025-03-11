package twig

import (
	"strings"
	"testing"
)

func TestApplyTag(t *testing.T) {
	engine := New()

	// Test simple spaceless filter with apply tag
	template1 := `{% apply spaceless %}
<div>
  <strong>foo</strong>
</div>
{% endapply %}`

	result, err := engine.ParseTemplate(template1)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	output, err := result.Render(nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "<div><strong>foo</strong></div>"
	if normalizeOutput(output) != normalizeOutput(expected) {
		t.Errorf("Test 1 failed. Expected '%s', got '%s'", expected, output)
	}

	// Test upper filter with apply tag
	template2 := `{% apply upper %}hello world{% endapply %}`

	result, err = engine.ParseTemplate(template2)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	output, err = result.Render(nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected = "HELLO WORLD"
	if normalizeOutput(output) != normalizeOutput(expected) {
		t.Errorf("Test 2 failed. Expected '%s', got '%s'", expected, output)
	}

	// Test with context variable
	template3 := `{% apply upper %}{{ name }}{% endapply %}`

	result, err = engine.ParseTemplate(template3)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	context := map[string]interface{}{
		"name": "john",
	}

	output, err = result.Render(context)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected = "JOHN"
	if normalizeOutput(output) != normalizeOutput(expected) {
		t.Errorf("Test 3 failed. Expected '%s', got '%s'", expected, output)
	}
}

func normalizeOutput(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
