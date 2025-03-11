# TWIG GO PROJECT GUIDELINES

## Project Overview
This is a Go implementation of the Twig template engine, designed to be lightweight, efficient, and closely follow the original Twig syntax and behavior while maintaining Go idioms and best practices.

## Development Rules

### Code Style and Philosophy
- Follow Go idioms and best practices
- No external dependencies beyond the Go standard library
- Clear and consistent naming conventions
- Comprehensive documentation for public APIs
- Expressive, readable code over clever tricks

### Core Engine Rules
- NEVER inject code directly into the core engine to work around failing tests
- NEVER modify test expectations to match implementation - fix the implementation instead
- Maintain backward compatibility whenever possible
- Keep the API surface clean and consistent
- Follow the original Twig semantics where appropriate

### Testing Philosophy
- All changes must pass existing tests
- Add new tests for new features or bug fixes
- Tests should be descriptive and illustrate usage
- If a test is failing, analyze and fix the core issue, not the test

### HTML Handling
- Don't modify whitespace in HTML content
- Only manage template code blocks, not the HTML itself
- Preserve content structure and formatting

### Build and Test Commands
```bash
# Run all tests
go test ./...

# Build the project
go build

# Lint the code
go vet
```

### Project Structure
- `parser.go`: Template parsing logic
- `tokenizer.go`: Template tokenization
- `expr.go`: Expression node definitions
- `render.go`: Template rendering
- `extension.go`: Extensions and filters
- `twig.go`: Main engine API

### Common Issues
- Template parsing issues often relate to operator precedence or tokenization
- Map key order is not guaranteed in Go - don't hardcode expected order in tests
- For backwards compatibility, ensure all filters return the same types as the PHP Twig implementation