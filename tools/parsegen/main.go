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
	parserContent, err := os.ReadFile("tools/parsegen/parser.template.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading parser template: %v\n", err)
		os.Exit(1)
	}

	nodesContent, err := os.ReadFile("tools/parsegen/nodes.template.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading nodes template: %v\n", err)
		os.Exit(1)
	}

	// Generate parser file
	parserFilePath := filepath.Join(*outputDir, "parser.gen.go")
	if err := os.WriteFile(parserFilePath, parserContent, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating parser file: %v\n", err)
		os.Exit(1)
	}

	// Generate nodes file
	nodesFilePath := filepath.Join(*outputDir, "nodes.gen.go")
	if err := os.WriteFile(nodesFilePath, nodesContent, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating nodes file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated files:\n- %s\n- %s\n", parserFilePath, nodesFilePath)
}
