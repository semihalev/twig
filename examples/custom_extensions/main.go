package main

import (
	"fmt"
	"github.com/semihalev/twig"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Create a new Twig engine
	engine := twig.New()

	// Add custom filter - reverses words in a string
	engine.AddFilter("reverse_words", func(value interface{}, args ...interface{}) (interface{}, error) {
		s := toString(value)
		words := strings.Fields(s)
		
		// Reverse the order of words
		for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
			words[i], words[j] = words[j], words[i]
		}
		
		return strings.Join(words, " "), nil
	})

	// Add custom function - repeats a string n times
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

	// Register a custom extension with multiple filters and functions
	engine.RegisterExtension("demo_extension", func(ext *twig.CustomExtension) {
		// Initialize random seed
		rand.Seed(time.Now().UnixNano())
		
		// Add a filter that shuffles characters in a string
		ext.Filters["shuffle"] = func(value interface{}, args ...interface{}) (interface{}, error) {
			s := toString(value)
			runes := []rune(s)
			// Shuffle algorithm
			rand.Shuffle(len(runes), func(i, j int) {
				runes[i], runes[j] = runes[j], runes[i]
			})
			return string(runes), nil
		}
		
		// Add a filter that formats a number with a prefix/suffix
		ext.Filters["format_number"] = func(value interface{}, args ...interface{}) (interface{}, error) {
			num, err := toFloat64(value)
			if err != nil {
				return value, nil
			}
			
			// Default format
			format := "%.2f"
			if len(args) > 0 {
				if fmt, ok := args[0].(string); ok {
					format = fmt
				}
			}
			
			// Default prefix
			prefix := ""
			if len(args) > 1 {
				if pre, ok := args[1].(string); ok {
					prefix = pre
				}
			}
			
			// Default suffix
			suffix := ""
			if len(args) > 2 {
				if suf, ok := args[2].(string); ok {
					suffix = suf
				}
			}
			
			return prefix + fmt.Sprintf(format, num) + suffix, nil
		}
		
		// Add a function that generates a random number between min and max
		ext.Functions["random_between"] = func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return rand.Intn(100), nil
			}
			
			min, err := toInt(args[0])
			if err != nil {
				return nil, err
			}
			
			max, err := toInt(args[1])
			if err != nil {
				return nil, err
			}
			
			if max <= min {
				return nil, fmt.Errorf("max must be greater than min")
			}
			
			return min + rand.Intn(max-min+1), nil
		}
	})

	// Create a template that uses all the custom filters and functions
	templateContent := `
Custom Filter and Function Demo:
-------------------------------

1. reverse_words filter:
   Original: "Hello world from Twig"
   Reversed: "{{ 'Hello world from Twig'|reverse_words }}"

2. repeat function:
   {{ repeat('=-', 20) }}

3. shuffle filter (from extension):
   Original: "abcdefghijklmnopqrstuvwxyz"
   Shuffled: "{{ 'abcdefghijklmnopqrstuvwxyz'|shuffle }}"

4. format_number filter (from extension):
   Original: {{ price }}
   Formatted: {{ price|format_number('%.2f', '$', ' USD') }}

5. random_between function (from extension):
   Random number between 1 and 100: {{ random_between(1, 100) }}

6. Combining filters:
   {{ 'hello world'|reverse_words|shuffle }}

7. Calculation with formatted output:
   {{ (price * 0.2)|format_number('%.2f', '$', ' (20% of original price)') }}
`

	// Parse the template
	template, err := engine.ParseTemplate(templateContent)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		os.Exit(1)
	}

	// Render the template with context
	context := map[string]interface{}{
		"price": 99.99,
	}
	
	result, err := template.Render(context)
	if err != nil {
		fmt.Println("Error rendering template:", err)
		os.Exit(1)
	}
	
	fmt.Println(result)
}

// Helper functions

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toInt(v interface{}) (int, error) {
	if v == nil {
		return 0, fmt.Errorf("cannot convert nil to int")
	}
	
	switch val := v.(type) {
	case int:
		return val, nil
	case float64:
		return int(val), nil
	case string:
		i, err := strconv.Atoi(val)
		if err != nil {
			return 0, err
		}
		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

func toFloat64(v interface{}) (float64, error) {
	if v == nil {
		return 0, fmt.Errorf("cannot convert nil to float64")
	}
	
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}