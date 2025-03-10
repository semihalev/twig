package twig

import (
	"bytes"
	"testing"
)

func TestBasicMacroString(t *testing.T) {
	engine := New()

	source := `
{% macro test(a = "def") %}
  {{ a }}
{% endmacro %}

{{ test() }}
`

	// Register the template
	err := engine.RegisterString("test_template", source)
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	// Render the template
	var buf bytes.Buffer
	err = engine.RenderTo(&buf, "test_template", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	// Output should contain the default value
	t.Logf("Output: %q", buf.String())
}
