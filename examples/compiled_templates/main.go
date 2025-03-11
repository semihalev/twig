package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/semihalev/twig"
)

func main() {
	// Create a new engine
	engine := twig.New()

	// Directory for compiled templates
	compiledDir := "./compiled_templates"

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(compiledDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Create a loader for source templates
	sourceLoader := twig.NewFileSystemLoader([]string{"./templates"})
	engine.RegisterLoader(sourceLoader)

	// Create a loader for compiled templates
	compiledLoader := twig.NewCompiledLoader(compiledDir)

	// Two operation modes:
	// 1. Compile mode: Compile templates and save them for later use
	// 2. Production mode: Load pre-compiled templates

	// Check if we're in compile mode
	if len(os.Args) > 1 && os.Args[1] == "compile" {
		fmt.Println("Compiling templates...")

		// List all template files
		templateFiles, err := os.ReadDir("./templates")
		if err != nil {
			fmt.Printf("Error reading templates directory: %v\n", err)
			return
		}

		// Load and compile each template
		for _, file := range templateFiles {
			if file.IsDir() {
				continue
			}

			// Get the template name (filename without extension)
			name := filepath.Base(file.Name())
			if ext := filepath.Ext(name); ext != "" {
				name = name[:len(name)-len(ext)]
			}

			fmt.Printf("Compiling template: %s\n", name)

			// Load the template
			_, err := engine.Load(name)
			if err != nil {
				fmt.Printf("Error loading template %s: %v\n", name, err)
				continue
			}

			// Save the compiled template
			if err := compiledLoader.SaveCompiled(engine, name); err != nil {
				fmt.Printf("Error compiling template %s: %v\n", name, err)
				continue
			}
		}

		fmt.Println("Templates compiled successfully!")
		return
	}

	// Production mode: Load compiled templates
	fmt.Println("Loading compiled templates...")

	// Try to load all compiled templates
	err := compiledLoader.LoadAll(engine)
	if err != nil {
		fmt.Printf("Error loading compiled templates: %v\n", err)
		fmt.Println("Falling back to source templates...")
	} else {
		fmt.Printf("Loaded %d compiled templates\n", engine.GetCachedTemplateCount())
	}

	// Test rendering a template
	context := map[string]interface{}{
		"name":  "World",
		"items": []string{"apple", "banana", "cherry"},
		"user": map[string]interface{}{
			"username": "testuser",
			"email":    "test@example.com",
		},
	}

	// Try to render each template
	templates := []string{"welcome", "products", "user_profile"}
	for _, name := range templates {
		result, err := engine.Render(name, context)
		if err != nil {
			fmt.Printf("Error rendering template %s: %v\n", name, err)
			continue
		}

		fmt.Printf("\n--- Rendered template: %s ---\n", name)
		fmt.Println(result)
	}
}
