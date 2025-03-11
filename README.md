# Twig

<p align="center">
  <img src="twig.svg" alt="Twig Logo" width="150" height="180">
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/semihalev/twig"><img src="https://goreportcard.com/badge/github.com/semihalev/twig?style=flat-square" alt="Go Report Card"></a>
  <a href="https://pkg.go.dev/github.com/semihalev/twig"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?style=flat-square" alt="go.dev reference"></a>
  <a href="https://github.com/semihalev/twig/releases"><img src="https://img.shields.io/github/v/release/semihalev/twig?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/semihalev/twig/blob/master/LICENSE"><img src="https://img.shields.io/github/license/semihalev/twig?style=flat-square" alt="License"></a>
</p>

<p align="center">
  A fast, memory-efficient Twig template engine implementation for Go.<br>
  Provides full support for the Twig template language in a Go-native way.
</p>

Twig for Go is a comprehensive template engine that implements the Twig syntax popularized in PHP, but designed specifically for Go applications. It offers exceptional performance with minimal memory usage while providing a rich feature set including template inheritance, macros, filters, and more. Whether you're building a website, API responses, or any text generation system, Twig makes template management elegant and maintainable.

## Why Choose Twig?

- **Superior Performance**: Benchmarks show Twig is up to 57× faster than Go's standard template engine for complex templates
- **Memory Efficiency**: Uses up to 90% less memory than standard Go templates while being significantly faster
- **Powerful Features**: Template inheritance, macros, filters, and imports create a robust ecosystem for template reuse
- **Developer Friendly**: Clean, readable syntax with clear error messages that help debug template issues
- **Zero Dependencies**: No external Go dependencies means easy integration in any project
- **Production Ready**: Template caching, compilation, and development mode provide flexibility for all environments

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Supported Twig Syntax](#supported-twig-syntax)
- [Filter Support](#filter-support)
- [Custom Filter and Function Registration](#custom-filter-and-function-registration)
- [Macros and Reusability](#macros-and-reusability)
- [Development Mode and Caching](#development-mode-and-caching)
- [Debugging and Error Handling](#debugging-and-error-handling)
- [String Escape Sequences](#string-escape-sequences)
- [Whitespace Handling](#whitespace-handling)
- [Performance](#performance)
- [Examples](#examples)
- [Template Compilation](#template-compilation)
- [Installation Requirements](#installation-requirements)
- [Running Tests](#running-tests)
- [Compatibility](#compatibility)
- [Versioning Policy](#versioning-policy)
- [Security Considerations](#security-considerations)
- [Contributing](#contributing)
- [Roadmap](#roadmap)
- [Community & Support](#community--support)
- [License](#license)

## Features

- Zero-allocation rendering where possible
- Full Twig syntax support including macros, imports, and includes
- Template inheritance
- Extensible with filters, functions, tests, and operators
- Multiple loader types (filesystem, in-memory, compiled)
- Template compilation for maximum performance
- Whitespace control features (trim modifiers)
- Compatible with Go's standard library interfaces
- Memory pooling for improved performance
- Attribute caching to reduce reflection overhead
- Detailed error reporting and debugging tools
- Thread-safe and concurrency optimized
- Robust escape sequence handling in string literals

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
- Macros: `{% macro name(args) %}...{% endmacro %}`
- Imports: `{% import "template.twig" as alias %}`
- Selective imports: `{% from "template.twig" import macro1, macro2 as alias %}`
- Apply filters to blocks: `{% apply filter %}...{% endapply %}`
- Verbatim content: `{% verbatim %}...{% endverbatim %}`
- Comments: `{# comment #}`
- Array literals: `[1, 2, 3]`
- Conditional expressions: `condition ? true_expr : false_expr`
- String escape sequences: `\n`, `\"`, `\\`, `\{`, etc.
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
- `spaceless`: Removes whitespace between HTML tags

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

## Macros and Reusability

Twig macros are a powerful way to create reusable templates and components. They work like functions that can output template content.

### Defining Macros

Macros are defined using the `{% macro %}` tag:

```twig
{% macro input(name, value = '', type = 'text') %}
    <input type="{{ type }}" name="{{ name }}" value="{{ value|e }}" />
{% endmacro %}
```

### Using Macros

Once defined, macros can be used directly in the same template:

```twig
{{ input('username', 'johndoe') }}
{{ input('password', '', 'password') }}
{{ input('submit', 'Login', 'submit') }}
```

### Macros with Default Parameters

Macros can have default parameter values:

```twig
{% macro button(text, type = 'button', class = 'btn') %}
    <button type="{{ type }}" class="{{ class }}">{{ text }}</button>
{% endmacro %}

{# Using with defaults #}
{{ button('Click Me') }}

{# Overriding defaults #}
{{ button('Submit', 'submit', 'btn btn-primary') }}
```

### Importing Macros from Other Templates

Macros can be defined in one template and imported into another:

```twig
{# In forms.twig #}
{% macro input(name, value = '', type = 'text') %}
    <input type="{{ type }}" name="{{ name }}" value="{{ value|e }}" />
{% endmacro %}

{# In template.twig #}
{% import "forms.twig" as forms %}

{{ forms.input('username') }}
```

### Nested Macros

Macros can call other macros:

```twig
{% macro field(name, value = '', type = 'text', label = '') %}
    <div class="field">
        {% if label %}
            <label for="{{ name }}">{{ label }}</label>
        {% endif %}
        {{ input(name, value, type) }}
    </div>
{% endmacro %}

{% macro input(name, value = '', type = 'text') %}
    <input type="{{ type }}" name="{{ name }}" value="{{ value|e }}" />
{% endmacro %}

{{ field('username', user.username, 'text', 'Username') }}
```

### Macro Variable Scope

Macros have their own variable scope, separate from the template that calls them:

```twig
{% set name = 'Global' %}

{% macro greet(name = 'Default') %}
    {# This only sees the 'name' parameter, not the global 'name' #}
    Hello, {{ name }}!
{% endmacro %}

{{ greet() }}                 {# Output: Hello, Default! #}
{{ greet('Local') }}          {# Output: Hello, Local! #}
{{ name }}                    {# Output: Global #}
```

To access the template's context from within a macro, pass the special `_context` variable:

```twig
{% macro listItems(items, _context) %}
    {% for item in items %}
        <li>{{ item }} (from {{ _context.templateName }})</li>
    {% endfor %}
{% endmacro %}

{% set templateName = 'product-list' %}
{{ listItems(products, _context) }}
```

### Self-Referencing with _self

You can reference macros from the same template using the `_self` variable:

```twig
{% macro input(name, value) %}<input name="{{ name }}" value="{{ value }}">{% endmacro %}
{% macro form(action) %}
    <form action="{{ action }}">
        {{ _self.input('username', '') }}
        <button type="submit">Submit</button>
    </form>
{% endmacro %}

{{ _self.form('/submit') }}
```

### Organizing Macro Libraries

For larger applications, organize macros into component libraries:

```twig
{# components/forms.twig #}
{% macro input(name, value = '') %}...{% endmacro %}
{% macro textarea(name, value = '') %}...{% endmacro %}

{# components/layout.twig #}
{% macro card(title, content) %}...{% endmacro %}
{% macro panel(title, content) %}...{% endmacro %}

{# template.twig #}
{% import "components/forms.twig" as forms %}
{% import "components/layout.twig" as layout %}

{{ layout.card('Login', forms.input('username')) }}
```

### Error Handling in Macros

Common errors to watch for:

1. **Undefined macros**: Ensure macros are defined before calling them
2. **Parameter mismatches**: Check parameter names and types
3. **Scope issues**: Remember macros can't access parent scope without explicitly passing context

When debugging, enable debug mode to get detailed error messages:

```go
engine.SetDebug(true)
```

### Performance Considerations

According to our benchmarks, imported macros perform 27% better than direct macro usage due to optimizations in the caching system:

| Macro Usage Type | Time (µs/op) | Relative Performance |
|------------------|--------------|----------------------|
| Direct           | 3.16         | 1.00x                |
| Imported         | 2.30         | 0.73x (27% faster)   |
| Nested           | 2.98         | 0.94x (6% faster)    |

For high-performance applications:
- Prefer imported macros over direct macro usage
- Group related macros in dedicated template files
- Use the import cache efficiently by importing each macro file once

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

## Debugging and Error Handling

Twig provides enhanced error reporting and debugging tools to help during development:

```go
// Enable debug mode
engine.SetDebug(true)

// Set custom debug level for more detailed logging
twig.SetDebugLevel(twig.DebugVerbose) // Options: DebugOff, DebugError, DebugWarning, DebugInfo, DebugVerbose

// Redirect debug output to a file
logFile, _ := os.Create("twig_debug.log")
twig.SetDebugWriter(logFile)
```

### Debug Features

When debug is enabled, you get:
- **Enhanced Error Messages**: Includes template name, line number, and source context
- **Performance Tracing**: Log rendering times for templates and template sections
- **Variable Inspection**: Log template variable values and types
- **Hierarchical Error Reporting**: Proper error propagation through template inheritance

Example error output:
```
Error in template 'user_profile.twig' at line 45, column 12: 
undefined variable "user"
Line 45: <h1>Welcome, {{ user.name }}!</h1>
                        ^
```

### Error Handling Best Practices

```go
// Render with proper error handling
result, err := engine.Render("template.twig", context)
if err != nil {
    // Enhanced errors with full context
    fmt.Printf("Rendering failed: %v\n", err)
    return
}
```

## String Escape Sequences

Twig supports standard string escape sequences to include special characters in string literals:

```twig
{{ "Line with \n a newline character" }}
{{ "Quotes need escaping: \"quoted text\"" }}
{{ "Use \\ for a literal backslash" }}
{{ "Escape Twig syntax: \{\{ this is not a variable \}\}" }}
```

The following escape sequences are supported:
- `\n`: Newline
- `\r`: Carriage return
- `\t`: Tab
- `\"`: Double quote
- `\'`: Single quote
- `\\`: Backslash
- `\{`: Left curly brace (to avoid being interpreted as Twig syntax)
- `\}`: Right curly brace (to avoid being interpreted as Twig syntax)

This is particularly useful in JavaScript blocks or when you need to include literal braces in your output.

## Whitespace Handling

Twig templates can have significant whitespace that affects the rendered output. This implementation supports the following mechanism for controlling whitespace:

### Apply Tag

The `apply` tag allows you to apply a filter to an entire block of template content:

```twig
{% apply upper %}
    This text will be converted to uppercase.
    {{ variable }} will also be uppercase.
{% endapply %}
```

This is particularly useful for:
- Applying transformations to both static and dynamic content
- Handling complex HTML with the spaceless filter
- Creating consistent text formatting across multiple lines
- Processing content with custom filters

Example usages:

```twig
{# Remove whitespace between HTML tags #}
{% apply spaceless %}
    <div>
        <strong>Hello</strong>
    </div>
{% endapply %}

{# Result: <div><strong>Hello</strong></div> #}

{# Convert content to uppercase #}
{% apply upper %}
    Welcome, {{ user.name }}!
{% endapply %}

{# Replace content #}
{% apply replace('e', 'a') %}
    Hello there
{% endapply %}
{# Result: Hallo thara #}
```

### Verbatim Tag

The `verbatim` tag allows you to output Twig syntax without it being processed:

```twig
{% verbatim %}
    This {{ will not be processed }} as a variable.
    {% if statements won't be executed %}
{% endverbatim %}
```

This is useful when:
- Displaying Twig syntax as part of documentation
- Including template examples in your output
- Working with JavaScript frameworks that use similar syntax (Vue.js, Angular, etc.)
- Creating code examples that include Twig syntax

### Whitespace Control

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

This feature helps you create cleaner output, especially when generating HTML with proper indentation in templates but needing compact output for production.

## Performance

The library is designed with performance in mind:
- Minimal memory allocations
- Efficient parsing and rendering
- Memory pooling for frequently allocated objects
- Attribute caching to reduce reflection overhead
- Template caching
- Production/development mode toggle
- Optimized filter chain processing
- Thread-safe concurrent rendering

### Benchmark Results

Twig consistently outperforms other Go template engines, especially for complex templates:

| Engine      | Simple (µs/op) | Medium (µs/op) | Complex (µs/op) |
|-------------|----------------|----------------|-----------------|
| Twig        | 0.28           | 0.14           | 0.14            |
| Go Template | 0.90           | 0.94           | 7.98            |
| Pongo2      | 0.86           | 0.91           | 4.57            |
| Stick       | 4.00           | 15.85          | 54.56           |

For complex templates, Twig is:
- **57x faster** than Go's standard library
- **33x faster** than Pongo2
- **390x faster** than Stick

Twig also uses approximately **90% less memory** than Go's standard library while being **57x faster**.

### Macro Performance

| Macro Usage Type | Time (µs/op) | Relative Performance |
|------------------|--------------|----------------------|
| Direct           | 3.16         | 1.00x                |
| Imported         | 2.30         | 0.73x (27% faster)   |
| Nested           | 2.98         | 0.94x (6% faster)    |

See [full benchmark results](benchmark/BENCHMARK_RESULTS.md) for detailed comparison.

## Examples

The repository includes several example applications demonstrating various features of Twig:

### Simple Example

A basic example showing how to use Twig templates:

```go
// From examples/simple/main.go
package main

import (
    "fmt"
    "github.com/semihalev/twig"
    "os"
)

func main() {
    // Create a Twig engine
    engine := twig.New()
    
    // Load templates from memory
    template := "Hello, {{ name }}!"
    engine.AddTemplateString("greeting", template)
    
    // Render the template
    context := map[string]interface{}{
        "name": "World",
    }
    
    result, err := engine.Render("greeting", context)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    fmt.Println(result) // Output: Hello, World!
}
```

### Development Mode Example

```go
// From examples/development_mode/main.go
package main

import (
    "fmt"
    "github.com/semihalev/twig"
    "os"
)

func main() {
    // Create a Twig engine with development mode enabled
    engine := twig.New()
    engine.SetDevelopmentMode(true)
    
    // Add a template loader
    loader := twig.NewFileSystemLoader([]string{"./templates"})
    engine.RegisterLoader(loader)
    
    // Render a template
    context := map[string]interface{}{
        "name": "Developer",
    }
    
    // Templates will auto-reload if changed
    result, err := engine.Render("hello.twig", context)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    fmt.Println(result)
}
```

### Custom Extensions Example

Example showing how to create custom Twig extensions:

```go
// From examples/custom_extensions/main.go
package main

import (
    "fmt"
    "github.com/semihalev/twig"
    "strings"
)

func main() {
    // Create a Twig engine
    engine := twig.New()
    
    // Register a custom extension
    engine.RegisterExtension("text_tools", func(ext *twig.CustomExtension) {
        // Add a filter to count words
        ext.Filters["word_count"] = func(value interface{}, args ...interface{}) (interface{}, error) {
            str, ok := value.(string)
            if !ok {
                return 0, nil
            }
            return len(strings.Fields(str)), nil
        }
        
        // Add a function to generate Lorem Ipsum text
        ext.Functions["lorem"] = func(args ...interface{}) (interface{}, error) {
            count := 5
            if len(args) > 0 {
                if c, ok := args[0].(int); ok {
                    count = c
                }
            }
            return strings.Repeat("Lorem ipsum dolor sit amet. ", count), nil
        }
    })
    
    // Use the custom extensions in a template
    template := `
    The following text has {{ text|word_count }} words:
    
    {{ text }}
    
    Generated text:
    {{ lorem(3) }}
    `
    
    engine.AddTemplateString("example", template)
    
    // Render the template
    context := map[string]interface{}{
        "text": "This is an example of a custom filter in action.",
    }
    
    result, err := engine.Render("example", context)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    fmt.Println(result)
}
```

### Macros Example

Example showing how to use macros for reusable UI components:

```go
// From examples/macros/main.go
package main

import (
    "fmt"
    "github.com/semihalev/twig"
    "os"
)

func main() {
    // Create a new Twig engine
    engine := twig.New()

    // Create template with macros
    macrosTemplate := `
    {# Define macros in a separate template #}
    {% macro input(name, value = '', type = 'text', size = 20) %}
      <input type="{{ type }}" name="{{ name }}" value="{{ value|e }}" size="{{ size }}">
    {% endmacro %}

    {% macro label(text, for = '') %}
      <label{% if for %} for="{{ for }}"{% endif %}>{{ text }}</label>
    {% endmacro %}
    `

    // Create a template that imports and uses macros
    mainTemplate := `
    {% import "macros.twig" as forms %}

    <form>
      <div class="form-row">
        {{ forms.label('Username', 'username') }}
        {{ forms.input('username', user.username) }}
      </div>
      <div class="form-row">
        {{ forms.input('submit', 'Submit', 'submit') }}
      </div>
    </form>
    `

    // Register templates
    engine.RegisterString("macros.twig", macrosTemplate)
    engine.RegisterString("main.twig", mainTemplate)

    // Create context with user data
    context := map[string]interface{}{
        "user": map[string]interface{}{
            "username": "johndoe",
        },
    }

    // Render the template
    err := engine.RenderTo(os.Stdout, "main.twig", context)
    if err != nil {
        fmt.Printf("Error rendering template: %v\n", err)
    }
}
```

More examples can be found in the `examples/` directory:
- `examples/compiled_templates/` - Shows how to compile and use compiled templates
- `examples/macros/` - Demonstrates the use of macros in templates with nested and imported examples
- `examples/development_mode/` - Shows how to use the development mode for template auto-reloading
- `examples/simple/` - Basic usage examples for quick reference

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

## Installation Requirements

- Go 1.18 or higher
- No external dependencies required (all dependencies are included in Go's standard library)

## Running Tests

To run the test suite:

```bash
go test ./...
```

For tests with coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Compatibility

This implementation aims to be compatible with Twig PHP version 3.x syntax and features. While we strive for full compatibility, there may be some minor differences due to the nature of the Go language compared to PHP.

## Versioning Policy

This project follows Semantic Versioning:
- MAJOR version for incompatible API changes
- MINOR version for backwards-compatible functionality additions
- PATCH version for backwards-compatible bug fixes

## Security Considerations

When using Twig or any template engine:

- Never allow untrusted users to modify or create templates directly
- Be cautious with user-provided variables in templates
- Consider using the HTML escaping filters (`escape` or `e`) for user-provided content
- In sandbox mode (if implementing custom functions/filters), carefully validate inputs

## Contributing

Contributions are welcome! Here's how you can contribute:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure your code passes all tests and follows the existing code style.

## Roadmap

✅ Features already implemented:
- Full macro functionality with imports
- Comprehensive benchmarking including memory usage
- Template inheritance and includes
- Filters and functions
- HTML escaping and safety

Future development plans include:

- Expanded sandbox mode for enhanced security
- Additional optimization techniques for macro evaluation
- Template profiling tools for performance analysis
- Additional loader types
- Enhanced from ... import syntax for selective macro imports
- Support for lazy-loading templates
- Test coverage improvements
- Plugin system for third-party extensions

## Community & Support

- Submit bug reports and feature requests through GitHub Issues
- Ask questions using GitHub Discussions
- Contribute to the project by submitting Pull Requests

## License

This project is licensed under the MIT License - see the LICENSE file for details.