# Twig

Twig is a fast, memory-efficient Twig template engine implementation for Go. It aims to provide full support for the Twig template language in a Go-native way.

## Features

- Zero-allocation rendering where possible
- Full Twig syntax support
- Template inheritance
- Extensible with filters, functions, tests, and operators
- Multiple loader types (filesystem, in-memory)
- Compatible with Go's standard library interfaces

## Installation

```bash
go get github.com/semihalev/twig
```

## Basic Usage

```go
package main

import (
    "fmt"
    "github.com/semihalev/twig"
    "os"
)

func main() {
    // Create a new Twig engine
    engine := twig.New()
    
    // Add a template loader
    loader := twig.NewFileSystemLoader([]string{"./templates"})
    engine.RegisterLoader(loader)
    
    // Render a template
    context := map[string]interface{}{
        "name": "World",
        "items": []string{"apple", "banana", "orange"},
    }
    
    // Render to a string
    result, err := engine.Render("index.twig", context)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    fmt.Println(result)
    
    // Or render directly to a writer
    err = engine.RenderTo(os.Stdout, "index.twig", context)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
}
```

## Supported Twig Syntax

- Variable printing: `{{ variable }}`
- Control structures: `{% if %}`, `{% for %}`, etc.
- Filters: `{{ variable|filter }}`
- Functions: `{{ function(args) }}`
- Template inheritance: `{% extends %}`, `{% block %}`
- Includes: `{% include %}`
- Comments: `{# comment #}`
- Array literals: `[1, 2, 3]`
- Conditional expressions: `condition ? true_expr : false_expr`
- And more...

## Filter Support

Twig filters allow you to modify variables and expressions. Filters are applied using the pipe (`|`) character:

```twig
{{ 'hello'|upper }}
```

### Supported Filters

This implementation supports many standard Twig filters:

- `upper`: Converts a string to uppercase
- `lower`: Converts a string to lowercase
- `capitalize`: Capitalizes a string
- `trim`: Removes whitespace from both sides of a string
- `slice`: Extracts a slice of a string or array
- `default`: Returns a default value if the variable is empty or undefined
- `join`: Joins array elements with a delimiter
- `split`: Splits a string by a delimiter
- `length` / `count`: Returns the length of a string, array, or collection
- `replace`: Replaces occurrences of a substring
- `escape` / `e`: HTML-escapes a string
- `raw`: Marks the value as safe (no escaping)
- `first`: Returns the first element of an array or first character of a string
- `last`: Returns the last element of an array or last character of a string
- `reverse`: Reverses a string or array
- `sort`: Sorts an array
- `keys`: Returns the keys of an array or map
- `merge`: Merges arrays or maps
- `date`: Formats a date
- `number_format`: Formats a number
- `abs`: Returns the absolute value of a number
- `round`: Rounds a number
- `striptags`: Strips HTML tags from a string
- `nl2br`: Replaces newlines with HTML line breaks

### Filter Usage Examples

**Basic filters:**
```twig
{{ 'hello'|upper }}                          {# Output: HELLO #}
{{ name|capitalize }}                        {# Output: Name #}
{{ 'hello world'|split(' ')|first }}         {# Output: hello #}
```

**Filters with arguments:**
```twig
{{ 'hello world'|slice(0, 5) }}              {# Output: hello #}
{{ [1, 2, 3]|join('-') }}                    {# Output: 1-2-3 #}
{{ 'hello'|replace('e', 'a') }}              {# Output: hallo #}
```

**Chained filters:**
```twig
{{ 'hello'|upper|trim }}                     {# Output: HELLO #}
{{ ['a', 'b', 'c']|join(', ')|upper }}       {# Output: A, B, C #}
```

**Filters in expressions:**
```twig
{{ (name|capitalize) ~ ' ' ~ (greeting|upper) }}
{% if name|length > 3 %}long{% else %}short{% endif %}
```

## Custom Filter and Function Registration

Twig allows you to register custom filters and functions to extend its functionality.

### Adding Custom Filters

```go
// Create a new Twig engine
engine := twig.New()

// Add a simple filter that reverses words in a string
engine.AddFilter("reverse_words", func(value interface{}, args ...interface{}) (interface{}, error) {
    s := toString(value)
    words := strings.Fields(s)
    
    // Reverse the order of words
    for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
        words[i], words[j] = words[j], words[i]
    }
    
    return strings.Join(words, " "), nil
})

// Use it in a template
template, _ := engine.ParseTemplate("{{ 'hello world'|reverse_words }}")
result, _ := template.Render(nil)
// Result: "world hello"
```

### Adding Custom Functions

```go
// Add a custom function that repeats a string n times
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

// Use it in a template
template, _ := engine.ParseTemplate("{{ repeat('abc', 3) }}")
result, _ := template.Render(nil)
// Result: "abcabcabc"
```

### Creating a Custom Extension

You can also create a custom extension with multiple filters and functions:

```go
// Create and register a custom extension
engine.RegisterExtension("my_extension", func(ext *twig.CustomExtension) {
    // Add a filter
    ext.Filters["shuffle"] = func(value interface{}, args ...interface{}) (interface{}, error) {
        s := toString(value)
        runes := []rune(s)
        // Simple shuffle algorithm
        rand.Shuffle(len(runes), func(i, j int) {
            runes[i], runes[j] = runes[j], runes[i]
        })
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
```

## Performance

The library is designed with performance in mind:
- Minimal memory allocations
- Efficient parsing and rendering
- Template caching

## License

This project is licensed under the MIT License - see the LICENSE file for details.