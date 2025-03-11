package twig

import (
	"strings"
	"testing"
)

// TestFromTagBasic tests the most basic from tag use case
func TestFromTagBasic(t *testing.T) {
	engine := New()

	// Macro library template with a simple macro
	macroLib := `{% macro hello(name) %}Hello, {{ name }}!{% endmacro %}`

	// Simple template using the from tag
	mainTemplate := `{% from "macros.twig" import hello %}
{{ hello('World') }}`

	// Register templates
	err := engine.RegisterString("macros.twig", macroLib)
	if err != nil {
		t.Fatalf("Error registering macros.twig: %v", err)
	}

	err = engine.RegisterString("main.twig", mainTemplate)
	if err != nil {
		t.Fatalf("Error registering main.twig: %v", err)
	}

	// Render the template
	result, err := engine.Render("main.twig", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	// Check the output
	expected := "Hello, World!"
	if !strings.Contains(result, expected) {
		t.Errorf("Expected %q in result, but got: %s", expected, result)
	}
}

// TestFromTagWithAlias tests the from tag with an alias
func TestFromTagWithAlias(t *testing.T) {
	engine := New()

	// Macro library template
	macroLib := `{% macro greet(name) %}Hello, {{ name }}!{% endmacro %}
{% macro farewell(name) %}Goodbye, {{ name }}!{% endmacro %}`

	// Template using from import with aliases
	template := `{% from "macros.twig" import greet as hello, farewell as bye %}
{{ hello('John') }}
{{ bye('Jane') }}`

	// Register templates
	err := engine.RegisterString("macros.twig", macroLib)
	if err != nil {
		t.Fatalf("Error registering macros.twig: %v", err)
	}

	err = engine.RegisterString("template.twig", template)
	if err != nil {
		t.Fatalf("Error registering template.twig: %v", err)
	}

	// Render the template
	result, err := engine.Render("template.twig", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	// Check the output
	expectedHello := "Hello, John!"
	expectedBye := "Goodbye, Jane!"

	if !strings.Contains(result, expectedHello) {
		t.Errorf("Expected %q in result, but got: %s", expectedHello, result)
	}

	if !strings.Contains(result, expectedBye) {
		t.Errorf("Expected %q in result, but got: %s", expectedBye, result)
	}
}

// TestFromTagMultipleImports tests importing multiple macros from a template
func TestFromTagMultipleImports(t *testing.T) {
	engine := New()

	// Macro library template with multiple macros
	macroLib := `{% macro input(name, value) %}
<input name="{{ name }}" value="{{ value }}">
{% endmacro %}

{% macro label(text) %}
<label>{{ text }}</label>
{% endmacro %}

{% macro button(text) %}
<button>{{ text }}</button>
{% endmacro %}`

	// Template importing multiple macros
	template := `{% from "form_macros.twig" import input, label, button %}
<form>
  {{ label('Username') }}
  {{ input('username', 'john') }}
  {{ button('Submit') }}
</form>`

	// Register templates
	err := engine.RegisterString("form_macros.twig", macroLib)
	if err != nil {
		t.Fatalf("Error registering form_macros.twig: %v", err)
	}

	err = engine.RegisterString("form.twig", template)
	if err != nil {
		t.Fatalf("Error registering form.twig: %v", err)
	}

	// Render the template
	result, err := engine.Render("form.twig", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	// Check the output
	expectedElements := []string{
		`<label>Username</label>`,
		`<input name="username" value="john">`,
		`<button>Submit</button>`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected %q in result, but got: %s", expected, result)
		}
	}
}

// TestFromTagMixedAliases tests importing some macros with aliases and some without
func TestFromTagMixedAliases(t *testing.T) {
	engine := New()

	// Macro library template
	macroLib := `{% macro header(text) %}
<h1>{{ text }}</h1>
{% endmacro %}

{% macro paragraph(text) %}
<p>{{ text }}</p>
{% endmacro %}

{% macro link(href, text) %}
<a href="{{ href }}">{{ text }}</a>
{% endmacro %}`

	// Template with mixed alias usage
	template := `{% from "content_macros.twig" import header, paragraph as p, link as a %}
{{ header('Title') }}
{{ p('This is a paragraph.') }}
{{ a('#', 'Click here') }}`

	// Register templates
	err := engine.RegisterString("content_macros.twig", macroLib)
	if err != nil {
		t.Fatalf("Error registering content_macros.twig: %v", err)
	}

	err = engine.RegisterString("content.twig", template)
	if err != nil {
		t.Fatalf("Error registering content.twig: %v", err)
	}

	// Render the template
	result, err := engine.Render("content.twig", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	// Check the output
	expectedElements := []string{
		`<h1>Title</h1>`,
		`<p>This is a paragraph.</p>`,
		`<a href="#">Click here</a>`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected %q in result, but got: %s", expected, result)
		}
	}
}
