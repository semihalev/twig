package twig

import (
	"fmt"
	"os"
	"path/filepath"
)

// Loader defines the interface for template loading
type Loader interface {
	// Load loads a template by name, returning its source code
	Load(name string) (string, error)
	
	// Exists checks if a template exists
	Exists(name string) bool
}

// TimestampAwareLoader is an interface for loaders that can check modification times
type TimestampAwareLoader interface {
	Loader
	
	// GetModifiedTime returns the last modification time of a template
	GetModifiedTime(name string) (int64, error)
}

// FileSystemLoader loads templates from the file system
type FileSystemLoader struct {
	paths        []string
	suffix       string
	defaultPaths []string
	// Stores paths for each loaded template to avoid repeatedly searching for the file
	templatePaths map[string]string
}

// ArrayLoader loads templates from an in-memory array
type ArrayLoader struct {
	templates map[string]string
}

// ChainLoader chains multiple loaders together
type ChainLoader struct {
	loaders []Loader
}

// NewFileSystemLoader creates a new file system loader
func NewFileSystemLoader(paths []string) *FileSystemLoader {
	// Add default path
	defaultPaths := []string{"."}
	
	// If no paths provided, use default
	if len(paths) == 0 {
		paths = defaultPaths
	}
	
	// Normalize paths
	normalizedPaths := make([]string, len(paths))
	for i, path := range paths {
		normalizedPaths[i] = filepath.Clean(path)
	}
	
	return &FileSystemLoader{
		paths:        normalizedPaths,
		suffix:       ".twig",
		defaultPaths: defaultPaths,
		templatePaths: make(map[string]string),
	}
}

// Load loads a template from the file system
func (l *FileSystemLoader) Load(name string) (string, error) {
	// Check if we already know the location of this template
	if filePath, ok := l.templatePaths[name]; ok {
		// Check if file still exists at this path
		if _, err := os.Stat(filePath); err == nil {
			// Read file content
			content, err := os.ReadFile(filePath)
			if err != nil {
				return "", fmt.Errorf("error reading template %s: %w", name, err)
			}
			
			return string(content), nil
		}
		// If file doesn't exist anymore, remove from cache and search again
		delete(l.templatePaths, name)
	}

	// Check each path for the template
	for _, path := range l.paths {
		filePath := filepath.Join(path, name)
		
		// Add suffix if not already present
		if !hasSuffix(filePath, l.suffix) {
			filePath = filePath + l.suffix
		}
		
		// Check if file exists
		if _, err := os.Stat(filePath); err == nil {
			// Save the path for future lookups
			l.templatePaths[name] = filePath
			
			// Read file content
			content, err := os.ReadFile(filePath)
			if err != nil {
				return "", fmt.Errorf("error reading template %s: %w", name, err)
			}
			
			return string(content), nil
		}
	}
	
	return "", fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
}

// Exists checks if a template exists in the file system
func (l *FileSystemLoader) Exists(name string) bool {
	// Check each path for the template
	for _, path := range l.paths {
		filePath := filepath.Join(path, name)
		
		// Add suffix if not already present
		if !hasSuffix(filePath, l.suffix) {
			filePath = filePath + l.suffix
		}
		
		// Check if file exists
		if _, err := os.Stat(filePath); err == nil {
			return true
		}
	}
	
	return false
}

// SetSuffix sets the file suffix for templates
func (l *FileSystemLoader) SetSuffix(suffix string) {
	l.suffix = suffix
}

// GetModifiedTime returns the last modification time of a template file
func (l *FileSystemLoader) GetModifiedTime(name string) (int64, error) {
	// If we already know where this template is, check that path directly
	if filePath, ok := l.templatePaths[name]; ok {
		info, err := os.Stat(filePath)
		if err != nil {
			// If file doesn't exist anymore, remove from cache
			if os.IsNotExist(err) {
				delete(l.templatePaths, name)
			}
			return 0, err
		}
		
		return info.ModTime().Unix(), nil
	}
	
	// Otherwise search for the template
	for _, path := range l.paths {
		filePath := filepath.Join(path, name)
		
		// Add suffix if not already present
		if !hasSuffix(filePath, l.suffix) {
			filePath = filePath + l.suffix
		}
		
		// Check if file exists
		info, err := os.Stat(filePath)
		if err == nil {
			// Save the path for future lookups
			l.templatePaths[name] = filePath
			
			return info.ModTime().Unix(), nil
		}
	}
	
	return 0, fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
}

// NewArrayLoader creates a new array loader
func NewArrayLoader(templates map[string]string) *ArrayLoader {
	return &ArrayLoader{
		templates: templates,
	}
}

// Load loads a template from the array
func (l *ArrayLoader) Load(name string) (string, error) {
	if template, ok := l.templates[name]; ok {
		return template, nil
	}
	
	return "", fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
}

// Exists checks if a template exists in the array
func (l *ArrayLoader) Exists(name string) bool {
	_, ok := l.templates[name]
	return ok
}

// SetTemplate adds or updates a template in the array
func (l *ArrayLoader) SetTemplate(name, template string) {
	l.templates[name] = template
}

// NewChainLoader creates a new chain loader
func NewChainLoader(loaders []Loader) *ChainLoader {
	return &ChainLoader{
		loaders: loaders,
	}
}

// Load loads a template from the first loader that has it
func (l *ChainLoader) Load(name string) (string, error) {
	for _, loader := range l.loaders {
		if loader.Exists(name) {
			return loader.Load(name)
		}
	}
	
	return "", fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
}

// Exists checks if a template exists in any of the loaders
func (l *ChainLoader) Exists(name string) bool {
	for _, loader := range l.loaders {
		if loader.Exists(name) {
			return true
		}
	}
	
	return false
}

// AddLoader adds a loader to the chain
func (l *ChainLoader) AddLoader(loader Loader) {
	l.loaders = append(l.loaders, loader)
}

// Helper function to check if a string has a suffix
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}