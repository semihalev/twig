package twig

import (
	"bytes"
	"testing"
)

// BenchmarkRenderSimpleTemplate benchmarks rendering a simple template with just variables
func BenchmarkRenderSimpleTemplate(b *testing.B) {
	engine := New()
	err := engine.RegisterString("simple", "Hello, {{ name }}!")
	if err != nil {
		b.Fatalf("Error registering template: %v", err)
	}

	context := map[string]interface{}{
		"name": "World",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("simple")
		_ = template.RenderTo(&buf, context)
	}
}

// BenchmarkRenderComplexTemplate benchmarks rendering a template with conditionals, loops, and filters
func BenchmarkRenderComplexTemplate(b *testing.B) {
	engine := New()
	templateContent := `
<div class="container">
  <h1>{{ title|upper }}</h1>
  {% if showHeader %}
    <div class="header">Welcome, {{ user.name }}!</div>
  {% endif %}
  <ul class="items">
    {% for item in items %}
      <li>{{ item.name }} - {{ item.price|format("$%.2f") }}</li>
    {% endfor %}
  </ul>
  {% set total = 0 %}
  {% for item in items %}
    {% set total = total + item.price %}
  {% endfor %}
  <div class="total">Total: {{ total|format("$%.2f") }}</div>
</div>
`
	err := engine.RegisterString("complex", templateContent)
	if err != nil {
		b.Fatalf("Error registering template: %v", err)
	}

	context := map[string]interface{}{
		"title":      "Product List",
		"showHeader": true,
		"user": map[string]interface{}{
			"name": "John",
		},
		"items": []map[string]interface{}{
			{"name": "Item 1", "price": 10.5},
			{"name": "Item 2", "price": 15.0},
			{"name": "Item 3", "price": 8.75},
			{"name": "Item 4", "price": 12.25},
			{"name": "Item 5", "price": 9.99},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("complex")
		_ = template.RenderTo(&buf, context)
	}
}

// BenchmarkRenderMacros benchmarks rendering a template with macro definitions and calls
func BenchmarkRenderMacros(b *testing.B) {
	engine := New()
	templateContent := `
{% macro input(name, value='', type='text') %}
  <input type="{{ type }}" name="{{ name }}" value="{{ value|e }}">
{% endmacro %}

{% macro form(action, method='post') %}
  <form action="{{ action }}" method="{{ method }}">
    {{ _self.input('username', user.username) }}
    {{ _self.input('password', '', 'password') }}
    {{ _self.input('submit', 'Login', 'submit') }}
  </form>
{% endmacro %}

{{ _self.form('/login') }}
`
	err := engine.RegisterString("macros", templateContent)
	if err != nil {
		b.Fatalf("Error registering template: %v", err)
	}

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"username": "johndoe",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("macros")
		_ = template.RenderTo(&buf, context)
	}
}

// BenchmarkRenderInheritance benchmarks rendering a template with inheritance
func BenchmarkRenderInheritance(b *testing.B) {
	engine := New()
	
	// Base template
	baseTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
</head>
<body>
    <header>{% block header %}Default Header{% endblock %}</header>
    <main>{% block content %}Default Content{% endblock %}</main>
    <footer>{% block footer %}Default Footer{% endblock %}</footer>
</body>
</html>
`
	err := engine.RegisterString("base", baseTemplate)
	if err != nil {
		b.Fatalf("Error registering template: %v", err)
	}
	
	// Child template
	childTemplate := `
{% extends "base" %}

{% block title %}{{ pageTitle }} - {{ parent() }}{% endblock %}

{% block header %}
    <h1>{{ pageTitle }}</h1>
    {{ parent() }}
{% endblock %}

{% block content %}
    {% for item in items %}
        <div class="item">{{ item }}</div>
    {% endfor %}
{% endblock %}
`
	err = engine.RegisterString("child", childTemplate)
	if err != nil {
		b.Fatalf("Error registering template: %v", err)
	}

	context := map[string]interface{}{
		"pageTitle": "Products Page",
		"items": []string{
			"Product 1",
			"Product 2",
			"Product 3",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("child")
		_ = template.RenderTo(&buf, context)
	}
}

// BenchmarkRenderFilters benchmarks heavy use of filter chains
func BenchmarkRenderFilters(b *testing.B) {
	engine := New()
	templateContent := `
{% set text = "Hello, this is some example text for filtering!" %}

<p>{{ text|upper|trim }}</p>
<p>{{ text|lower|replace("example", "sample")|capitalize }}</p>
<p>{{ text|split(" ")|join("-")|upper }}</p>
<p>{{ text|length }}</p>
<p>{{ '2023-05-15'|date("Y-m-d") }}</p>
<p>{{ 123.456|number_format(2, ".", ",") }}</p>
<p>{{ ['a', 'b', 'c']|join(", ")|upper }}</p>
<p>{{ text|slice(7, 10)|capitalize }}</p>
<p>{{ text|replace({"example": "great", "text": "content"}) }}</p>
<p>{{ text|default("No text provided")|upper }}</p>
`
	err := engine.RegisterString("filters", templateContent)
	if err != nil {
		b.Fatalf("Error registering template: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("filters")
		_ = template.RenderTo(&buf, nil)
	}
}

// BenchmarkRenderWithLargeContext benchmarks rendering with a large context
func BenchmarkRenderWithLargeContext(b *testing.B) {
	engine := New()
	templateContent := `
<ul>
{% for user in users %}
  <li>{{ user.id }}: {{ user.name }} ({{ user.email }})
    <ul>
      {% for role in user.roles %}
        <li>{{ role }}</li>
      {% endfor %}
    </ul>
  </li>
{% endfor %}
</ul>
`
	err := engine.RegisterString("large_context", templateContent)
	if err != nil {
		b.Fatalf("Error registering template: %v", err)
	}

	// Create a large context with 100 users
	users := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		users[i] = map[string]interface{}{
			"id":    i + 1,
			"name":  "User " + string(rune(65+i%26)),
			"email": "user" + string(rune(65+i%26)) + "@example.com",
			"roles": []string{"User", "Editor", "Admin", "Viewer"}[0:1+(i%4)],
		}
	}
	
	context := map[string]interface{}{
		"users": users,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("large_context")
		_ = template.RenderTo(&buf, context)
	}
}

// BenchmarkContextCloning benchmarks the RenderContext cloning operation
func BenchmarkContextCloning(b *testing.B) {
	engine := New()
	
	// Create a base context with some data
	baseContext := NewRenderContext(engine.environment, map[string]interface{}{
		"user": map[string]interface{}{
			"id":   123,
			"name": "John Doe",
			"roles": []string{
				"admin", "editor", "user",
			},
		},
		"settings": map[string]interface{}{
			"theme":        "dark",
			"notifications": true,
			"language":     "en",
		},
		"items": []map[string]interface{}{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
			{"id": 3, "name": "Item 3"},
		},
	}, engine)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a clone of the context
		clonedCtx := baseContext.Clone()
		clonedCtx.Release() // Return to pool after use
	}
	
	baseContext.Release() // Clean up
}

// BenchmarkExpressionEvaluation benchmarks various expression evaluations
func BenchmarkExpressionEvaluation(b *testing.B) {
	engine := New()
	
	// Register a simple template with different expression types
	templateContent := `
{{ 1 + 2 * 3 }}
{{ "Hello " ~ name ~ "!" }}
{{ items[0] }}
{{ user.name }}
{{ items|length > 3 ? "Many items" : "Few items" }}
{{ range(1, 10)|join(", ") }}
`
	err := engine.RegisterString("expressions", templateContent)
	if err != nil {
		b.Fatalf("Error registering template: %v", err)
	}

	context := map[string]interface{}{
		"name": "World",
		"user": map[string]interface{}{
			"name": "John",
			"age":  30,
		},
		"items": []string{"a", "b", "c", "d", "e"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("expressions")
		_ = template.RenderTo(&buf, context)
	}
}

// BenchmarkStringOperations benchmarks string manipulation operations
func BenchmarkStringOperations(b *testing.B) {
	engine := New()
	
	// Register a template with various string operations
	templateContent := `
{{ "  Hello, World!  "|trim }}
{{ text|replace("o", "0") }}
{{ text|upper }}
{{ text|lower }}
{{ text|capitalize }}
{{ text|slice(7, 5) }}
{{ text|split(", ")|join("-") }}
{{ "%s, %s!"|format("Hello", "World") }}
`
	err := engine.RegisterString("string_ops", templateContent)
	if err != nil {
		b.Fatalf("Error registering template: %v", err)
	}

	context := map[string]interface{}{
		"text": "Hello, World!",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("string_ops")
		_ = template.RenderTo(&buf, context)
	}
}