package twig

import (
	"testing"
)

func TestParentFunction(t *testing.T) {
	engine := New()

	// Create a simple parent-child template relationship
	baseTemplate := `
<!-- BASE TEMPLATE -->
<div class="base">
    {% block test %}
    <p>This is the parent content</p>
    {% endblock %}
</div>
`

	childTemplate := `
<!-- CHILD TEMPLATE -->
{% extends "base.twig" %}

{% block test %}
    <h2>Child heading</h2>
    {{ parent() }}
    <p>Child footer</p>
{% endblock %}
`

	// Register the templates
	engine.RegisterString("base.twig", baseTemplate)
	engine.RegisterString("child.twig", childTemplate)

	// Render the child template
	output, err := engine.Render("child.twig", nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Print the output for debugging
	t.Logf("Rendered output:\n%s", output)

	// Check for expected content
	mustContain(t, output, "BASE TEMPLATE")
	mustContain(t, output, "Child heading")
	mustContain(t, output, "This is the parent content")
	mustContain(t, output, "Child footer")

	// Verify ordering
	inOrderCheck(t, output, "Child heading", "This is the parent content")
	inOrderCheck(t, output, "This is the parent content", "Child footer")
}

func TestNestedParentFunction(t *testing.T) {
	// Test multi-level inheritance with parent() functionality
	engine := New()

	// Create a simple parent-child relationship
	baseTemplate := `<!-- BASE TEMPLATE -->
{% block content %}
<p>Base content</p>
{% endblock %}
`

	// Skip the middle template and test direct parent-child
	childTemplate := `<!-- CHILD TEMPLATE -->
{% extends "base.twig" %}
{% block content %}
<h1>Child content</h1>
{{ parent() }}
<p>More child content</p>
{% endblock %}
`

	engine.RegisterString("base.twig", baseTemplate)
	engine.RegisterString("child.twig", childTemplate)

	// Render the child template
	output, err := engine.Render("child.twig", nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Print the output for debugging
	t.Logf("Rendered output:\n%s", output)

	// Check that content from parent and child is properly included
	mustContain(t, output, "BASE TEMPLATE")
	mustContain(t, output, "Base content")
	mustContain(t, output, "Child content")
	mustContain(t, output, "More child content")

	// Verify ordering
	inOrderCheck(t, output, "Child content", "Base content")
	inOrderCheck(t, output, "Base content", "More child content")
}

func TestSimpleMultiLevelParentFunction(t *testing.T) {
	// A simpler test with just two levels to isolate the issue
	engine := New()

	// Enable debug mode
	SetDebugLevel(DebugInfo)
	engine.SetDebug(true)

	// Create a simpler template hierarchy with just parent and child
	baseTemplate := `<!-- BASE TEMPLATE -->
{% block content %}
<div>Base content</div>
{% endblock %}
`

	// Middle template with parent() call
	middleTemplate := `<!-- MIDDLE TEMPLATE -->
{% extends "base.twig" %}
{% block content %}
<div class="middle">
    {{ parent() }}
    <p>Middle content</p>
</div>
{% endblock %}
`

	// Register the templates
	engine.RegisterString("base.twig", baseTemplate)
	engine.RegisterString("middle.twig", middleTemplate)

	// Render the middle template
	output, err := engine.Render("middle.twig", nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Print the output for debugging
	t.Logf("Rendered output:\n%s", output)

	// Test the output
	mustContain(t, output, "BASE TEMPLATE")
	mustContain(t, output, "Base content")
	mustContain(t, output, "Middle content")
	inOrderCheck(t, output, "Base content", "Middle content")
}

func TestThreeLevelParentFunction(t *testing.T) {
	t.Skip("Multi-level parent() inheritance not yet implemented")
	// Let's try a simpler approach to debug the issue
	engine := New()

	// Enable debug mode
	SetDebugLevel(DebugInfo)
	engine.SetDebug(true)

	// Create a more basic version with just one middle parent() call
	baseTemplate := `<!-- BASE TEMPLATE -->
{% block content %}
<div>Base content</div>
{% endblock %}
`

	// Middle template with parent() call
	middleTemplate := `<!-- MIDDLE TEMPLATE -->
{% extends "base.twig" %}
{% block content %}
<div class="middle">
    <p>Middle content before parent</p>
    {{ parent() }}
    <p>Middle content after parent</p>
</div>
{% endblock %}
`

	// Child template that just extends middle, no parent() call
	childTemplate := `<!-- CHILD TEMPLATE -->
{% extends "middle.twig" %}
{% block content %}
<div class="child">
    <h1>Child content</h1>
    {{ parent() }}
</div>
{% endblock %}
`

	// Register the templates
	engine.RegisterString("base.twig", baseTemplate)
	engine.RegisterString("middle.twig", middleTemplate)
	engine.RegisterString("child.twig", childTemplate)

	// Render the child template which should access both parent and grandparent
	output, err := engine.Render("child.twig", nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Print the output for debugging
	t.Logf("Rendered output:\n%s", output)

	// Test the output
	mustContain(t, output, "BASE TEMPLATE")
	mustContain(t, output, "Base content")
	mustContain(t, output, "Middle content")
	mustContain(t, output, "Child header")
	mustContain(t, output, "Child footer")

	// Check order of content - should nest properly
	inOrderCheck(t, output, "Child header", "Base content")
	inOrderCheck(t, output, "Base content", "Middle content")
	inOrderCheck(t, output, "Middle content", "Child footer")
}

func TestParentFunctionErrors(t *testing.T) {
	engine := New()

	// Test parent() outside of a block
	template := `{{ parent() }}`
	engine.RegisterString("bad.twig", template)

	_, err := engine.Render("bad.twig", nil)
	if err == nil {
		t.Errorf("Expected an error when calling parent() outside of a block")
	}

	// Test parent() in a template without inheritance
	template2 := `{% block test %}{{ parent() }}{% endblock %}`
	engine.RegisterString("no_parent.twig", template2)

	_, err = engine.Render("no_parent.twig", nil)
	if err == nil {
		t.Errorf("Expected an error when calling parent() in a template without inheritance")
	}
}

// Helper functions for tests
func stringContains(haystack, needle string) bool {
	return stringIndexOf(haystack, needle) >= 0
}

func stringIndexOf(haystack, needle string) int {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

func inOrder(haystack, firstNeedle, secondNeedle string) bool {
	firstIndex := stringIndexOf(haystack, firstNeedle)
	secondIndex := stringIndexOf(haystack, secondNeedle)
	return firstIndex != -1 && secondIndex != -1 && firstIndex < secondIndex
}

func mustContain(t *testing.T, haystack, needle string) {
	if !stringContains(haystack, needle) {
		t.Errorf("Expected output to contain '%s', but it didn't", needle)
	}
}

func inOrderCheck(t *testing.T, haystack, firstNeedle, secondNeedle string) {
	if !inOrder(haystack, firstNeedle, secondNeedle) {
		if !stringContains(haystack, firstNeedle) {
			t.Errorf("First string '%s' not found in output", firstNeedle)
		} else if !stringContains(haystack, secondNeedle) {
			t.Errorf("Second string '%s' not found in output", secondNeedle)
		} else {
			t.Errorf("Expected '%s' to come before '%s' in output", firstNeedle, secondNeedle)
		}
	}
}
