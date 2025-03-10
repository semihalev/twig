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
- And more...

## Performance

The library is designed with performance in mind:
- Minimal memory allocations
- Efficient parsing and rendering
- Template caching

## License

This project is licensed under the MIT License - see the LICENSE file for details.