package twig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRelativePathsWithIncludes(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "twig-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory
	layoutDir := filepath.Join(tempDir, "layout")
	err = os.Mkdir(layoutDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create layout dir: %v", err)
	}

	// Create a simple test to isolate the issue
	baseTemplate := `Base template: {% include "./include.twig" %}`
	includeTemplate := `Included content`

	// Write templates to files
	err = os.WriteFile(filepath.Join(layoutDir, "base.twig"), []byte(baseTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write base template: %v", err)
	}

	err = os.WriteFile(filepath.Join(layoutDir, "include.twig"), []byte(includeTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write include template: %v", err)
	}

	// Create a new Twig engine
	engine := New()

	// Register the template directory
	loader := NewFileSystemLoader([]string{tempDir})
	engine.RegisterLoader(loader)

	// Render the base template directly
	output, err := engine.Render("layout/base.twig", nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Verify the output
	expectedOutput := `Base template: Included content`

	// Normalize whitespace for comparison
	expectedOutput = normalizeWhitespace(expectedOutput)
	output = normalizeWhitespace(output)

	if output != expectedOutput {
		t.Errorf("Template rendering did not match expected output.\nGot:\n%s\n\nExpected:\n%s", output, expectedOutput)
	}
}

// Test with macros in subdirectories
func TestRelativePathsWithFromImport(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "twig-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory
	macrosDir := filepath.Join(tempDir, "macros")
	err = os.Mkdir(macrosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create macros dir: %v", err)
	}

	// Create simple macro templates
	macrosTemplate := `{% macro simple() %}Macro output{% endmacro %}`
	useTemplate := `{% from "./simple.twig" import simple %}{{ simple() }}`

	// Write templates to files
	err = os.WriteFile(filepath.Join(macrosDir, "simple.twig"), []byte(macrosTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write macro template: %v", err)
	}

	err = os.WriteFile(filepath.Join(macrosDir, "use.twig"), []byte(useTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write use template: %v", err)
	}

	// Create a new Twig engine
	engine := New()

	// Register the template directory
	loader := NewFileSystemLoader([]string{tempDir})
	engine.RegisterLoader(loader)

	// Render the use template
	output, err := engine.Render("macros/use.twig", nil)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Verify the output
	expectedOutput := `Macro output`

	// Normalize whitespace for comparison
	expectedOutput = normalizeWhitespace(expectedOutput)
	output = normalizeWhitespace(output)

	if output != expectedOutput {
		t.Errorf("Template rendering did not match expected output.\nGot:\n%s\n\nExpected:\n%s", output, expectedOutput)
	}
}

func normalizeWhitespace(s string) string {
	// Replace multiple whitespace characters with a single space
	result := strings.Join(strings.Fields(s), " ")
	return result
}
