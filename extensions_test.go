package twig

import (
	"strings"
	"testing"
)

// Extension tests
// Consolidated from: extension_test.go, macro_test.go, etc.

// TestExtensionSystem tests the extension system
func TestExtensionSystem(t *testing.T) {
	engine := New()

	// Add a custom extension
	ext := &CustomExtension{
		Filters:   make(map[string]FilterFunc),
		Functions: make(map[string]FunctionFunc),
	}

	// Add filters and functions to the extension
	ext.Filters["custom_filter"] = func(value interface{}, args ...interface{}) (interface{}, error) {
		return "filtered", nil
	}

	ext.Functions["custom_function"] = func(args ...interface{}) (interface{}, error) {
		return "function result", nil
	}

	// Register the extension
	engine.AddExtension(ext)

	// Test filter
	source := "{{ 'test'|custom_filter }}"
	engine.RegisterString("test_filter", source)
	result, err := engine.Render("test_filter", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	if result != "filtered" {
		t.Errorf("Expected 'filtered', got %q", result)
	}

	// Test function
	source = "{{ custom_function() }}"
	engine.RegisterString("test_function", source)
	result, err = engine.Render("test_function", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	if result != "function result" {
		t.Errorf("Expected 'function result', got %q", result)
	}
}

// TestOrganizedCustomFilters tests custom filter registration
func TestOrganizedCustomFilters(t *testing.T) {
	engine := New()

	// Add a custom filter
	engine.AddFilter("reverse_words", func(value interface{}, args ...interface{}) (interface{}, error) {
		s := toString(value)
		words := strings.Fields(s)

		// Reverse the order of words
		for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
			words[i], words[j] = words[j], words[i]
		}

		return strings.Join(words, " "), nil
	})

	// Create a test template using the custom filter
	source := "{{ 'hello world'|reverse_words }}"
	template, err := engine.ParseTemplate(source)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}

	// Render the template
	result, err := template.Render(nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	expected := "world hello"
	if result != expected {
		t.Errorf("Expected result to be %q, but got %q", expected, result)
	}

	// Test a filter with arguments
	engine.AddFilter("take", func(value interface{}, args ...interface{}) (interface{}, error) {
		s := toString(value)
		count := 3 // default
		if len(args) > 0 {
			if n, ok := args[0].(int); ok {
				count = n
			}
		}

		if len(s) <= count {
			return s, nil
		}
		return s[:count], nil
	})

	// Test with arguments
	source = "{{ 'hello'|take(2) }}"
	engine.RegisterString("test_take", source)
	result, err = engine.Render("test_take", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	if result != "he" {
		t.Errorf("Expected 'he', got %q", result)
	}
}

// TestOrganizedCustomFunctions tests custom function registration
func TestOrganizedCustomFunctions(t *testing.T) {
	engine := New()

	// Add a custom function
	engine.AddFunction("repeat", func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return "", nil
		}

		text := toString(args[0])
		count, err := toInt(args[1])
		if err != nil {
			return "", err
		}

		return strings.Repeat(text, count), nil
	})

	// Create a test template using the custom function
	source := "{{ repeat('abc', 3) }}"
	template, err := engine.ParseTemplate(source)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}

	// Render the template
	result, err := template.Render(nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	expected := "abcabcabc"
	if result != expected {
		t.Errorf("Expected result to be %q, but got %q", expected, result)
	}

	// Test function with no arguments
	engine.AddFunction("get_greeting", func(args ...interface{}) (interface{}, error) {
		return "Hello!", nil
	})

	source = "{{ get_greeting() }}"
	engine.RegisterString("test_greeting", source)
	result, err = engine.Render("test_greeting", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	if result != "Hello!" {
		t.Errorf("Expected 'Hello!', got %q", result)
	}
}

// TestMacros tests macro functionality
func TestMacros(t *testing.T) {
	engine := New()

	// Create a template with a macro
	source := `
	{% macro input(name, value, type = "text") %}
	<input type="{{ type }}" name="{{ name }}" value="{{ value }}">
	{% endmacro %}
	
	{{ input('username', 'john') }}
	{{ input('password', '****', 'password') }}
	`

	engine.RegisterString("test_macro", source)
	result, err := engine.Render("test_macro", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	// Check the output contains the expected HTML
	if !strings.Contains(result, `<input type="text" name="username" value="john">`) {
		t.Errorf("Expected text input in result, but got: %s", result)
	}

	if !strings.Contains(result, `<input type="password" name="password" value="****">`) {
		t.Errorf("Expected password input in result, but got: %s", result)
	}
}

// TestExtensionIntegration tests integration with the extension system
func TestExtensionIntegration(t *testing.T) {
	engine := New()

	// Register a custom extension
	engine.RegisterExtension("test_extension", func(ext *CustomExtension) {
		// Add a filter
		ext.Filters["capitalize_words"] = func(value interface{}, args ...interface{}) (interface{}, error) {
			s := toString(value)
			words := strings.Fields(s)

			for i, word := range words {
				if len(word) > 0 {
					words[i] = strings.ToUpper(word[:1]) + word[1:]
				}
			}

			return strings.Join(words, " "), nil
		}

		// Add a function
		ext.Functions["join_words"] = func(args ...interface{}) (interface{}, error) {
			var words []string
			for _, arg := range args {
				words = append(words, toString(arg))
			}
			return strings.Join(words, " "), nil
		}
	})

	// Test filter
	source := "{{ 'hello world'|capitalize_words }}"
	engine.RegisterString("test_capitalize", source)
	result, err := engine.Render("test_capitalize", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got %q", result)
	}

	// Test function
	source = "{{ join_words('hello', 'beautiful', 'world') }}"
	engine.RegisterString("test_join", source)
	result, err = engine.Render("test_join", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	if result != "hello beautiful world" {
		t.Errorf("Expected 'hello beautiful world', got %q", result)
	}
}
