package main

import (
	"fmt"
	"os"

	"github.com/semihalev/twig"
)

func main() {
	// Create a new Twig engine
	engine := twig.New()

	// Create an in-memory template loader
	templates := map[string]string{
		"hello": "Hello, {{ name }}!",
		"page": `<!DOCTYPE html>
<html>
<head>
    <title>{{ title }}</title>
</head>
<body>
    <h1>{{ title }}</h1>
    <ul>
    {% for item in items %}
        <li>{{ item }}</li>
    {% endfor %}
    </ul>
</body>
</html>`,
	}
	loader := twig.NewArrayLoader(templates)
	engine.RegisterLoader(loader)

	// Render a simple template
	context := map[string]interface{}{
		"name": "World",
	}

	result, err := engine.Render("hello", context)
	if err != nil {
		fmt.Printf("Error rendering hello template: %v\n", err)
		return
	}

	fmt.Println("Result of 'hello' template:")
	fmt.Println(result)
	fmt.Println()

	// Render a more complex template
	pageContext := map[string]interface{}{
		"title": "My Page",
		"items": []string{"Item 1", "Item 2", "Item 3"},
	}

	fmt.Println("Result of 'page' template:")
	err = engine.RenderTo(os.Stdout, "page", pageContext)
	if err != nil {
		fmt.Printf("Error rendering page template: %v\n", err)
		return
	}
}