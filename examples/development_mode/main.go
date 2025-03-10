package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/semihalev/twig"
)

func main() {
	// Create template directories
	templatesDir := "./templates"
	os.MkdirAll(templatesDir, 0755)

	// Create a sample template file
	templatePath := filepath.Join(templatesDir, "hello.twig")
	templateContent := `
<!DOCTYPE html>
<html>
<head>
    <title>{{ title }}</title>
</head>
<body>
    <h1>Hello, {{ name }}!</h1>
    <p>Welcome to Twig in {{ mode }} mode.</p>
    
    {% if items %}
    <ul>
        {% for item in items %}
        <li>{{ item }}</li>
        {% endfor %}
    </ul>
    {% endif %}
</body>
</html>
`
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		fmt.Printf("Error creating template: %v\n", err)
		return
	}
	fmt.Println("Created template:", templatePath)

	// Create a Twig engine in development mode
	engine := twig.New()

	// Enable development mode
	engine.SetDevelopmentMode(true)
	fmt.Println("Development mode enabled:")
	fmt.Printf("  - Cache enabled:    %v\n", engine.IsCacheEnabled())
	fmt.Printf("  - Debug enabled:    %v\n", engine.IsDebugEnabled())
	fmt.Printf("  - AutoReload:       %v\n", engine.IsAutoReloadEnabled())

	// Register a template loader
	loader := twig.NewFileSystemLoader([]string{templatesDir})
	engine.RegisterLoader(loader)

	// Context for rendering
	context := map[string]interface{}{
		"title": "Twig Development Mode Example",
		"name":  "Developer",
		"mode":  "development",
		"items": []string{"Easy to use", "No template caching", "Auto-reloading enabled"},
	}

	// Render the template to stdout
	fmt.Println("\n--- Rendering in Development Mode ---")
	err = engine.RenderTo(os.Stdout, "hello.twig", context)
	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return
	}

	// Cached templates should be empty in development mode
	fmt.Printf("\n\nNumber of cached templates: %d\n", engine.GetCachedTemplateCount())

	// Now switch to production mode
	engine.SetDevelopmentMode(false)
	fmt.Println("\nProduction mode enabled:")
	fmt.Printf("  - Cache enabled:    %v\n", engine.IsCacheEnabled())
	fmt.Printf("  - Debug enabled:    %v\n", engine.IsDebugEnabled())
	fmt.Printf("  - AutoReload:       %v\n", engine.IsAutoReloadEnabled())

	// Change the context
	context["mode"] = "production"
	context["items"] = []string{"Maximum performance", "Template caching", "Optimized for speed"}

	// Render again in production mode
	fmt.Println("\n--- Rendering in Production Mode ---")
	err = engine.RenderTo(os.Stdout, "hello.twig", context)
	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return
	}

	// Now we should have templates in the cache
	fmt.Printf("\n\nNumber of cached templates: %d\n", engine.GetCachedTemplateCount())
	fmt.Println("Template names in cache:")
	for _, name := range engine.GetCachedTemplateNames() {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println("\nDemonstrating auto-reload with file modification:")

	// Let's modify the template file
	fmt.Println("Modifying template file...")
	modifiedContent := `
<!DOCTYPE html>
<html>
<head>
    <title>{{ title }} - UPDATED</title>
</head>
<body>
    <h1>Hello, {{ name }}!</h1>
    <p>This template was modified while the application is running.</p>
    <p>Welcome to Twig in {{ mode }} mode.</p>
    
    {% if items %}
    <ul>
        {% for item in items %}
        <li>{{ item }}</li>
        {% endfor %}
    </ul>
    {% endif %}
    
    <footer>Auto-reload detected this change automatically!</footer>
</body>
</html>
`
	err = os.WriteFile(templatePath, []byte(modifiedContent), 0644)
	if err != nil {
		fmt.Printf("Error modifying template: %v\n", err)
		return
	}

	// Enable development mode
	engine.SetDevelopmentMode(true)
	context["title"] = "Twig Auto-Reload Example"

	// Pause to make sure the file system registers the change
	fmt.Println("Waiting for file system to register change...")
	time.Sleep(1 * time.Second)

	// Render the template again - it should load the modified version
	fmt.Println("\n--- Rendering with Auto-Reload ---")
	err = engine.RenderTo(os.Stdout, "hello.twig", context)
	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return
	}

	fmt.Println("\n\nSuccess! The template was automatically reloaded.")
}
