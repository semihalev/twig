package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Parse command line arguments
	outputDir := flag.String("output", ".", "Output directory for generated files")
	flag.Parse()

	// Create output directory if it doesn't exist
	err := os.MkdirAll(*outputDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Read template files
	tokenContent, err := os.ReadFile("tools/lexgen/token.template.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading token template: %v\n", err)
		os.Exit(1)
	}

	lexerContent, err := os.ReadFile("tools/lexgen/lexer.template.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading lexer template: %v\n", err)
		os.Exit(1)
	}

	// Generate token file
	tokenFilePath := filepath.Join(*outputDir, "tokens.gen.go")
	if err := os.WriteFile(tokenFilePath, tokenContent, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating token file: %v\n", err)
		os.Exit(1)
	}

	// Generate lexer file
	lexerFilePath := filepath.Join(*outputDir, "lexer.gen.go")
	if err := os.WriteFile(lexerFilePath, lexerContent, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating lexer file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated files:\n- %s\n- %s\n", tokenFilePath, lexerFilePath)
}
