package twig

import (
	"strings"
	"testing"
)

func TestBlocksInConditionals(t *testing.T) {
	tests := []struct {
		name      string
		base      string
		extend    string
		expected  string
		shouldErr bool
	}{
		{
			name: "block inside if condition",
			base: `{% if true %}
		{% block my_block %}foo{% endblock %}
	{% endif %}`,
			extend: `{% extends "base" %}

	{% block my_block %}
		<p>{{ parent() }}</p>
	{% endblock %}`,
			expected: `<p>foo</p>`,
		},
		{
			name: "block inside else branch",
			base: `{% if false %}
		nothing
	{% else %}
		{% block my_block %}bar{% endblock %}
	{% endif %}`,
			extend: `{% extends "base" %}

	{% block my_block %}
		<div>{{ parent() }}</div>
	{% endblock %}`,
			expected: `<div>bar</div>`,
		},
		{
			name: "block inside nested if",
			base: `{% if true %}
		{% if true %}
			{% block my_block %}nested{% endblock %}
		{% endif %}
	{% endif %}`,
			extend: `{% extends "base" %}

	{% block my_block %}
		{{ parent() }}-extended
	{% endblock %}`,
			expected: `nested-extended`,
		},
		{
			name: "block inside for loop",
			base: `{% for i in [1] %}
		{% block my_block %}loop{{ i }}{% endblock %}
	{% endfor %}`,
			extend: `{% extends "base" %}

	{% block my_block %}
		prefix-{{ parent() }}
	{% endblock %}`,
			expected: `prefix-loop1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()
			loader := NewArrayLoader(map[string]string{
				"base":   tt.base,
				"extend": tt.extend,
			})
			engine.RegisterLoader(loader)

			result, err := engine.Render("extend", nil)
			if err != nil {
				if !tt.shouldErr {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if tt.shouldErr {
				t.Fatal("expected error but got none")
			}

			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tt.expected)

			// Remove extra whitespace
			result = strings.Join(strings.Fields(result), " ")
			expected = strings.Join(strings.Fields(expected), " ")

			if result != expected {
				t.Errorf("expected %q, got %q", expected, result)
			}
		})
	}
}

func TestBlocksInApplyFilter(t *testing.T) {
	tests := []struct {
		name      string
		base      string
		extend    string
		expected  string
		shouldErr bool
	}{
		{
			name: "block inside apply spaceless",
			base: `{% apply spaceless %}
		{% block my_block %}foo{% endblock %}
	{% endapply %}`,
			extend: `{% extends "base" %}

	{% block my_block %}
		<p>{{ parent() }}</p>
	{% endblock %}`,
			expected: `<p>foo</p>`,
		},
		{
			name: "block inside apply upper",
			base: `{% apply upper %}
		{% block my_block %}hello{% endblock %}
	{% endapply %}`,
			extend: `{% extends "base" %}

	{% block my_block %}
		{{ parent() }} world
	{% endblock %}`,
			expected: `HELLO WORLD`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()
			loader := NewArrayLoader(map[string]string{
				"base":   tt.base,
				"extend": tt.extend,
			})
			engine.RegisterLoader(loader)

			result, err := engine.Render("extend", nil)
			if err != nil {
				if !tt.shouldErr {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if tt.shouldErr {
				t.Fatal("expected error but got none")
			}

			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tt.expected)

			if result != expected {
				t.Errorf("expected %q, got %q", expected, result)
			}
		})
	}
}

func TestComplexNestedBlocks(t *testing.T) {
	// Test the exact example from the issue
	engine := New()
	loader := NewArrayLoader(map[string]string{
		"base": `
		{% apply spaceless %}
			{% block my_block %}foo{% endblock %}
		{% endapply %}
`,
		"extend": `
		{% extends "base" %}

		{% block my_block %}
			<p>{{ parent() }}</p>
		{% endblock %}
`,
	})

	engine.RegisterLoader(loader)
	result, err := engine.Render("extend", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The expected result with spaceless applied
	expected := "<p>foo</p>"
	result = strings.TrimSpace(result)

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
