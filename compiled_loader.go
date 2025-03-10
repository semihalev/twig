package twig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CompiledLoader loads templates from compiled files
type CompiledLoader struct {
	directory     string
	fileExtension string
}

// NewCompiledLoader creates a new compiled loader
func NewCompiledLoader(directory string) *CompiledLoader {
	return &CompiledLoader{
		directory:     directory,
		fileExtension: ".twig.compiled",
	}
}

// Load implements the Loader interface
func (l *CompiledLoader) Load(name string) (string, error) {
	filePath := filepath.Join(l.directory, name+l.fileExtension)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("compiled template file not found: %s", filePath)
	}

	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read compiled template file: %w", err)
	}

	// Deserialize the compiled template
	compiled, err := DeserializeCompiledTemplate(data)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize compiled template: %w", err)
	}

	// Return the source from the compiled template
	return compiled.Source, nil
}

// Exists checks if a compiled template exists
func (l *CompiledLoader) Exists(name string) bool {
	filePath := filepath.Join(l.directory, name+l.fileExtension)
	_, err := os.Stat(filePath)
	return err == nil
}

// LoadCompiled loads a compiled template from file and registers it with the engine
func (l *CompiledLoader) LoadCompiled(engine *Engine, name string) error {
	// The Load method of this loader already handles reading the compiled template
	// Just force a load of the template by the engine
	_, err := engine.Load(name)
	return err
}

// SaveCompiled saves a compiled template to file
func (l *CompiledLoader) SaveCompiled(engine *Engine, name string) error {
	// Get the template
	template, err := engine.Load(name)
	if err != nil {
		return err
	}

	// Compile the template
	compiled, err := template.Compile()
	if err != nil {
		return err
	}

	// Serialize the compiled template
	data, err := SerializeCompiledTemplate(compiled)
	if err != nil {
		return err
	}

	// Ensure the directory exists
	if err := os.MkdirAll(l.directory, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Save the compiled template
	filePath := filepath.Join(l.directory, name+l.fileExtension)
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write compiled template file: %w", err)
	}

	return nil
}

// CompileAll compiles all templates loaded by the engine and saves them
func (l *CompiledLoader) CompileAll(engine *Engine) error {
	// Get all cached template names
	templateNames := engine.GetCachedTemplateNames()

	// Compile and save each template
	for _, name := range templateNames {
		if err := l.SaveCompiled(engine, name); err != nil {
			return fmt.Errorf("failed to compile template %s: %w", name, err)
		}
	}

	return nil
}

// LoadAll loads all compiled templates from the directory
func (l *CompiledLoader) LoadAll(engine *Engine) error {
	// Register loader with the engine
	engine.RegisterLoader(l)

	// List all files in the directory
	files, err := ioutil.ReadDir(l.directory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Collect errors during loading
	var loadErrors []error
	
	// Load each compiled template
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Check if it's a compiled template file
		ext := filepath.Ext(file.Name())
		if ext == l.fileExtension {
			// Get the template name (filename without extension)
			name := file.Name()[:len(file.Name())-len(ext)]

			// Load the template
			if err := l.LoadCompiled(engine, name); err != nil {
				// Collect the error but continue loading other templates
				loadErrors = append(loadErrors, fmt.Errorf("failed to load compiled template %s: %w", name, err))
			}
		}
	}

	// If there were any errors, return a combined error
	if len(loadErrors) > 0 {
		var errMsg string
		for i, err := range loadErrors {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += err.Error()
		}
		return fmt.Errorf("errors loading compiled templates: %s", errMsg)
	}

	return nil
}

// Implement TimestampAwareLoader interface
func (l *CompiledLoader) GetModifiedTime(name string) (int64, error) {
	filePath := filepath.Join(l.directory, name+l.fileExtension)

	// Check if the file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	// Return the file modification time
	return info.ModTime().Unix(), nil
}
