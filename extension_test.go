package twig

import (
	"strings"
	"testing"
)

// TestAddFilter tests the AddFilter function
func TestAddFilter(t *testing.T) {
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
}

// TestAddFunction tests the AddFunction function
func TestAddFunction(t *testing.T) {
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
}

// TestCustomExtension tests the custom extension functionality
func TestCustomExtension(t *testing.T) {
	engine := New()
	
	// Create and register a custom extension
	engine.RegisterExtension("test_extension", func(ext *CustomExtension) {
		// Add a filter
		ext.Filters["shuffle"] = func(value interface{}, args ...interface{}) (interface{}, error) {
			s := toString(value)
			runes := []rune(s)
			// Simple shuffle algorithm (not random for reproducible tests)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return string(runes), nil
		}
		
		// Add a function
		ext.Functions["add"] = func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return 0, nil
			}
			
			a, errA := toFloat64(args[0])
			b, errB := toFloat64(args[1])
			
			if errA != nil || errB != nil {
				return 0, nil
			}
			
			return a + b, nil
		}
	})
	
	// Test the filter
	filterSource := "{{ 'hello'|shuffle }}"
	filterTemplate, err := engine.ParseTemplate(filterSource)
	if err != nil {
		t.Fatalf("Error parsing filter template: %v", err)
	}
	
	filterResult, err := filterTemplate.Render(nil)
	if err != nil {
		t.Fatalf("Error rendering filter template: %v", err)
	}
	
	// Due to how our shuffle works, we know the exact result
	filterExpected := "olleh"
	if filterResult != filterExpected {
		t.Errorf("Expected filter result to be %q, but got %q", filterExpected, filterResult)
	}
	
	// Test the function
	funcSource := "{{ add(2, 3) }}"
	funcTemplate, err := engine.ParseTemplate(funcSource)
	if err != nil {
		t.Fatalf("Error parsing function template: %v", err)
	}
	
	funcResult, err := funcTemplate.Render(nil)
	if err != nil {
		t.Fatalf("Error rendering function template: %v", err)
	}
	
	funcExpected := "5"
	if funcResult != funcExpected {
		t.Errorf("Expected function result to be %q, but got %q", funcExpected, funcResult)
	}
}

// TestMultipleExtensions tests registering multiple extensions
func TestMultipleExtensions(t *testing.T) {
	engine := New()
	
	// Register first extension
	engine.RegisterExtension("first_extension", func(ext *CustomExtension) {
		ext.Filters["double"] = func(value interface{}, args ...interface{}) (interface{}, error) {
			num, err := toFloat64(value)
			if err != nil {
				return value, nil
			}
			return num * 2, nil
		}
	})
	
	// Register second extension
	engine.RegisterExtension("second_extension", func(ext *CustomExtension) {
		ext.Filters["triple"] = func(value interface{}, args ...interface{}) (interface{}, error) {
			num, err := toFloat64(value)
			if err != nil {
				return value, nil
			}
			return num * 3, nil
		}
	})
	
	// Test using both extensions in a single template
	source := "{{ 5|double|triple }}"
	template, err := engine.ParseTemplate(source)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}
	
	result, err := template.Render(nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}
	
	// 5 doubled is 10, then tripled is 30
	expected := "30"
	if result != expected {
		t.Errorf("Expected result to be %q, but got %q", expected, result)
	}
}