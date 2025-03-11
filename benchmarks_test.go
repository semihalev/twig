package twig

import (
	"testing"
)

// Benchmark tests
// Consolidated from: twig_bench_test.go

// BenchmarkOrganizedTemplateCaching benchmarks template caching
func BenchmarkOrganizedTemplateCaching(b *testing.B) {
	engine := New()
	engine.RegisterString("template", "Hello {{ name }}!")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, _ := engine.Render("template", map[string]interface{}{
			"name": "World",
		})
		_ = result
	}
}

// BenchmarkOrganizedTemplateWithAttributes benchmarks accessing struct attributes
func BenchmarkOrganizedTemplateWithAttributes(b *testing.B) {
	engine := New()
	engine.RegisterString("template", "Hello {{ user.name }}! Age: {{ user.age }}")

	type User struct {
		Name string
		Age  int
	}

	user := User{
		Name: "John",
		Age:  30,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, _ := engine.Render("template", map[string]interface{}{
			"user": user,
		})
		_ = result
	}
}

// BenchmarkOrganizedTemplateWithAttributesRepeated benchmarks accessing struct attributes repeatedly
func BenchmarkOrganizedTemplateWithAttributesRepeated(b *testing.B) {
	engine := New()
	engine.RegisterString("template", "Hello {{ user.name }}! Age: {{ user.age }} Name again: {{ user.name }}")

	type User struct {
		Name string
		Age  int
	}

	user := User{
		Name: "John",
		Age:  30,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, _ := engine.Render("template", map[string]interface{}{
			"user": user,
		})
		_ = result
	}
}

// BenchmarkSimpleTemplate benchmarks rendering of simple templates
func BenchmarkSimpleTemplate(b *testing.B) {
	engine := New()
	engine.RegisterString("simple", "Hello, World!")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, _ := engine.Render("simple", nil)
		_ = result
	}
}

// BenchmarkComplexTemplate benchmarks rendering of complex templates
func BenchmarkComplexTemplate(b *testing.B) {
	engine := New()
	source := `
	{% for i in range(1, 10) %}
		{% if i % 2 == 0 %}
			{{ i }} is even
		{% else %}
			{{ i }} is odd
		{% endif %}
	{% endfor %}
	`
	engine.RegisterString("complex", source)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, _ := engine.Render("complex", nil)
		_ = result
	}
}

// BenchmarkFilters benchmarks filter execution
func BenchmarkFilters(b *testing.B) {
	engine := New()
	source := "{{ 'hello world'|upper|trim|slice(0, 5)|replace('H', 'J') }}"
	engine.RegisterString("filters", source)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, _ := engine.Render("filters", nil)
		_ = result
	}
}

// BenchmarkFunctions benchmarks function execution
func BenchmarkFunctions(b *testing.B) {
	engine := New()
	source := "{{ range(1, 10)|join(', ') }}"
	engine.RegisterString("functions", source)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, _ := engine.Render("functions", nil)
		_ = result
	}
}

// BenchmarkLoops benchmarks loop execution
func BenchmarkLoops(b *testing.B) {
	engine := New()
	source := `
	{% for i in range(1, 10) %}
		Item {{ i }}
	{% endfor %}
	`
	engine.RegisterString("loops", source)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, _ := engine.Render("loops", nil)
		_ = result
	}
}
