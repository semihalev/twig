# Twig

Twig is a fast, memory-efficient Twig template engine implementation for Go. It aims to provide full support for the Twig template language in a Go-native way.

## Features

- Zero-allocation rendering where possible
- Full Twig syntax support
- Template inheritance
- Extensible with filters, functions, tests, and operators
- Multiple loader types (filesystem, in-memory, compiled)
- Template compilation for maximum performance
- Whitespace control features (trim modifiers, spaceless tag)
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

## Development Mode and Caching

Twig provides several options to control template caching and debug behavior:

```go
// Create a new Twig engine
engine := twig.New()

// Enable development mode (enables debug, enables auto-reload, disables caching)
engine.SetDevelopmentMode(true)

// Or control individual settings
engine.SetDebug(true)        // Enable debug mode
engine.SetCache(false)       // Disable template caching
engine.SetAutoReload(true)   // Enable template auto-reloading
```

### Development Mode

When development mode is enabled:
- Template caching is disabled, ensuring you always see the latest changes
- Auto-reload is enabled, which will check for template modifications
- Debug mode is enabled for more detailed error messages

This is ideal during development to avoid having to restart your application when templates change.

### Auto-Reload & Template Modification Checking

The engine can automatically detect when template files change on disk and reload them:

```go
// Enable auto-reload to detect template changes
engine.SetAutoReload(true)
```

When auto-reload is enabled:
1. The engine tracks the last modification time of each template
2. When a template is requested, it checks if the file has been modified
3. If the file has changed, it automatically reloads the template
4. If the file hasn't changed, it uses the cached version (if caching is enabled)

This provides the best of both worlds:
- Fast performance (no unnecessary file system access for unchanged templates)
- Always up-to-date content (automatic reload when templates change)

### Production Mode

By default, Twig runs in production mode:
- Template caching is enabled for maximum performance
- Auto-reload is disabled to avoid unnecessary file system checks
- Debug mode is disabled to reduce overhead

## Whitespace Handling

Twig templates can have significant whitespace that affects the rendered output. This implementation supports several mechanisms for controlling whitespace:

### Whitespace Control Features

1. **Whitespace Control Modifiers**
   
   The whitespace control modifiers (`-` character) allow you to trim whitespace around tags:
   
   ```twig
   <div>
       {{- greeting -}}     {# Removes whitespace before and after #}
   </div>
   ```
   
   Using these modifiers:
   - `{{- ... }}`: Removes whitespace before the variable output
   - `{{ ... -}}`: Removes whitespace after the variable output
   - `{{- ... -}}`: Removes whitespace both before and after
   - `{%- ... %}`: Removes whitespace before the block tag
   - `{% ... -%}`: Removes whitespace after the block tag
   - `{%- ... -%}`: Removes whitespace both before and after

2. **Spaceless Tag**
   
   The `spaceless` tag removes whitespace between HTML tags (but preserves whitespace within text content):
   
   ```twig
   {% spaceless %}
       <div>
           <strong>Whitespace is removed between HTML tags</strong>
       </div>
   {% endspaceless %}
   ```
   
   This produces: `<div><strong>Whitespace is removed between HTML tags</strong></div>`

These features help you create cleaner output, especially when generating HTML with proper indentation in templates but needing compact output for production.

## Performance

The library is designed with performance in mind:
- Minimal memory allocations
- Efficient parsing and rendering
- Template caching
- Production/development mode toggle
- Optimized filter chain processing

## Template Compilation

For maximum performance in production environments, Twig supports compiling templates to a binary format:

### Benefits of Template Compilation

1. **Faster Rendering**: Pre-compiled templates skip the parsing step, leading to faster rendering
2. **Reduced Memory Usage**: Compiled templates can be more memory-efficient
3. **Better Deployment Options**: Compile during build and distribute only compiled templates
4. **No Source Required**: Run without needing access to the original template files

### Using Compiled Templates

```go
// Create a new engine
engine := twig.New()

// Compile a template
template, _ := engine.Load("template_name")
compiled, _ := template.Compile()

// Serialize to binary data
data, _ := twig.SerializeCompiledTemplate(compiled)

// Save to disk or transmit elsewhere...
ioutil.WriteFile("template.compiled", data, 0644)

// In production, load the compiled template
compiledData, _ := ioutil.ReadFile("template.compiled")
engine.LoadFromCompiledData(compiledData)
```

### Compiled Template Loader

A dedicated `CompiledLoader` provides easy handling of compiled templates:

```go
// Create a loader for compiled templates
loader := twig.NewCompiledLoader("./compiled_templates")

// Compile all templates in the engine
loader.CompileAll(engine)

// In production
loader.LoadAll(engine)
```

See the `examples/compiled_templates` directory for a complete example.

## License

This project is licensed under the MIT License - see the LICENSE file for details.