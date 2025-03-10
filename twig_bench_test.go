package twig

import (
	"testing"
)

func BenchmarkTemplateCaching(b *testing.B) {
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

func BenchmarkTemplateWithAttributes(b *testing.B) {
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

func BenchmarkTemplateWithAttributesRepeated(b *testing.B) {
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
