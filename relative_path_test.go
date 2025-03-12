package twig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRelativePathsWithIncludes(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "twig-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory
	layoutDir := filepath.Join(tempDir, "layout")
	err = os.Mkdir(layoutDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create layout dir: %v", err)
	}

	// Create a simple test to isolate the issue
	baseTemplate := `Base template: {% include "./include.twig" %}`
	includeTemplate := `Included content`

	// Write templates to files
	err = os.WriteFile(filepath.Join(layoutDir, "base.twig"), []byte(baseTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write base template: %v", err)
	}

	err = os.WriteFile(filepath.Join(layoutDir, "include.twig"), []byte(includeTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write include template: %v", err)
	}

	// Create a new Twig engine
	engine := New()

	// Register the template directory
	loader := NewFileSystemLoader([]string{tempDir})
	engine.RegisterLoader(loader)

	// Render the base template directly
	output, err := engine.Render("layout/base.twig", nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Verify the output
	expectedOutput := `Base template: Included content`

	// Normalize whitespace for comparison
	expectedOutput = normalizeWhitespace(expectedOutput)
	output = normalizeWhitespace(output)

	if output != expectedOutput {
		t.Errorf("Template rendering did not match expected output.\nGot:\n%s\n\nExpected:\n%s", output, expectedOutput)
	}
}

// Test with macros in subdirectories
func TestRelativePathsWithFromImport(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "twig-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory
	macrosDir := filepath.Join(tempDir, "macros")
	err = os.Mkdir(macrosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create macros dir: %v", err)
	}

	// Create simple macro templates
	macrosTemplate := `{% macro simple() %}Macro output{% endmacro %}`

	// Print the template content for debugging
	t.Logf("Simple template content: %s", macrosTemplate)

	// Note: The template needs to be in the format: {% from "template" import macro %}
	useTemplate := `{% from "./simple.twig" import simple %}{{ simple() }}`
	t.Logf("Use template content: %s", useTemplate)

	// Write templates to files
	err = os.WriteFile(filepath.Join(macrosDir, "simple.twig"), []byte(macrosTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write macro template: %v", err)
	}

	err = os.WriteFile(filepath.Join(macrosDir, "use.twig"), []byte(useTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write use template: %v", err)
	}

	// Create a new Twig engine
	engine := New()

	// Register the template directory
	loader := NewFileSystemLoader([]string{tempDir})
	engine.RegisterLoader(loader)

	// Render the use template
	output, err := engine.Render("macros/use.twig", nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Verify the output
	expectedOutput := `Macro output`

	// Normalize whitespace for comparison
	expectedOutput = normalizeWhitespace(expectedOutput)
	output = normalizeWhitespace(output)

	if output != expectedOutput {
		t.Errorf("Template rendering did not match expected output.\nGot:\n%s\n\nExpected:\n%s", output, expectedOutput)
	}
}

func normalizeWhitespace(s string) string {
	// Replace multiple whitespace characters with a single space
	result := strings.Join(strings.Fields(s), " ")
	return result
}

// TestRelativePathsWithExtendsInSubfolder tests a specific scenario where a template
// in a child folder extends another template in the same folder using a relative path
func TestRelativePathsWithExtendsInSubfolder(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "twig-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a main templates directory and a child subdirectory
	templatesDir := filepath.Join(tempDir, "templates")
	err = os.Mkdir(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	t.Logf("Template dir: %s", templatesDir)

	childDir := filepath.Join(templatesDir, "child")
	err = os.Mkdir(childDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create child dir: %v", err)
	}

	// Create the layout template in the templates directory
	layoutTemplate := `<!DOCTYPE html>
<html>
<head>
	<title>{% block title %}Main Layout Title{% endblock %}</title>
	{% block stylesheet %}{% endblock stylesheet %}	
</head>
<body>
	{% block content %}
	{% endblock content %}
	{% block javascript %}{% endblock javascript %}
</body>
</html>`

	// Create the base template in the child directory
	baseTemplate := `{% extends '../layout.html.twig' %}

{% block title %}{{ title }}{% if pair is defined %} | {{ pair | split('.', 2) | first | capitalize }} | {{ pair | split('.', 2) | last | replace('something', '') }}{% endif %} | Text{% endblock %}

{% block stylesheet %}
	<style>
	.brand-image {
		margin-top: -.5rem;
		margin-right: .2rem;
		height: 33px;
	}
	</style>
{% endblock %}
		{% block content %}
		<div class="container">
			<h1>{{ title }}</h1>
			<p>{{ content }}</p>
			<ul>
			{% for item in items %}
				<li>{{ item }}</li>
			{% endfor %}
			</ul>
			<a href="{{ '/' }}">main</a>
		</div>
        {% endblock %}
		{% block javascript %}
		<script>
			console.log('{{ content | split(' ', 2) | first }}');
        {% set break = false %}
        {% for item in items %}
			{% if not break %}
            {% set list = item %}
            window.top.location = '/{{ list }}';
            {% set break = true %}
			{% endif %}
        {% endfor %}
		</script>
		{% endblock javascript %}
		`

	// Create the child template that extends the base using a relative path
	childTemplate := `{% extends './layout.html.twig' %}`

	// Write templates to files
	err = os.WriteFile(filepath.Join(templatesDir, "layout.html.twig"), []byte(layoutTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write base template: %v", err)
	}

	err = os.WriteFile(filepath.Join(childDir, "layout.html.twig"), []byte(baseTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write base template: %v", err)
	}

	err = os.WriteFile(filepath.Join(childDir, "child.html.twig"), []byte(childTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write child template: %v", err)
	}

	// Create a new Twig engine with debug enabled
	engine := New()
	engine.SetDevelopmentMode(true)
	engine.SetDebug(true)
	engine.SetAutoReload(true)

	// Register the template directory
	loader := NewFileSystemLoader([]string{templatesDir})
	engine.RegisterLoader(loader)

	// Render the child template from the child directory
	output, err := engine.Render("child/child.html.twig", map[string]interface{}{
		"title":   "Base Title",
		"content": "This is the base content.",
		"items":   []string{"Item 1", "Item 2", "Item 3"},
	})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	//t.Logf("Output: %s", output)

	// Verify the output contains elements from both templates
	if !strings.Contains(output, "Base Title") {
		t.Errorf("Output should contain 'Base Title' from base template title block")
	}

	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Errorf("Output should contain DOCTYPE declaration from main layout template")
	}

	if !strings.Contains(output, "<div class=\"container\">") {
		t.Errorf("Output should contain container div from base template template")
	}

	// The child content should contain the base content
	if !strings.Contains(output, "This is the base content") {
		t.Errorf("Output should contain 'This is the base content' from base template")
	}

	t.Logf("Successfully rendered template with relative extends path in subfolder")
}
