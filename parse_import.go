package twig

import (
	"fmt"
	"strings"
)

func (p *Parser) parseImport(parser *Parser) (Node, error) {
	// Use debug logging if enabled
	if IsDebugEnabled() && debugger.level >= DebugVerbose {
		tokenIndex := parser.tokenIndex - 2
		LogVerbose("Parsing import, tokens available:")
		for i := 0; i < 10 && tokenIndex+i < len(parser.tokens); i++ {
			token := parser.tokens[tokenIndex+i]
			LogVerbose("  Token %d: Type=%d, Value=%q, Line=%d", i, token.Type, token.Value, token.Line)
		}
	}

	// Get the line number of the import token
	importLine := parser.tokens[parser.tokenIndex-2].Line

	// Check for incorrectly tokenized import syntax
	if parser.tokenIndex < len(parser.tokens) &&
		parser.tokens[parser.tokenIndex].Type == TOKEN_NAME &&
		strings.Contains(parser.tokens[parser.tokenIndex].Value, " as ") {

		// Special handling for combined syntax like "path.twig as alias"
		parts := strings.SplitN(parser.tokens[parser.tokenIndex].Value, " as ", 2)
		if len(parts) == 2 {
			templatePath := strings.TrimSpace(parts[0])
			alias := strings.TrimSpace(parts[1])

			if IsDebugEnabled() && debugger.level >= DebugVerbose {
				LogVerbose("Found combined import syntax: template=%q, alias=%q", templatePath, alias)
			}

			// Create an expression node for the template path
			var templateExpr Node
			if strings.HasPrefix(templatePath, "\"") && strings.HasSuffix(templatePath, "\"") {
				// It's already a quoted string
				templateExpr = NewLiteralNode(templatePath[1:len(templatePath)-1], importLine)
			} else {
				// Create a string literal node
				templateExpr = NewLiteralNode(templatePath, importLine)
			}

			// Skip to end of token
			parser.tokenIndex++

			// Expect block end
			if parser.tokenIndex >= len(parser.tokens) ||
				(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
					parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
				return nil, fmt.Errorf("expected block end token after import statement at line %d", importLine)
			}
			parser.tokenIndex++

			// Create import node
			return NewImportNode(templateExpr, alias, importLine), nil
		}
	}

	// Standard parsing path
	// Get the template to import
	templateExpr, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect 'as' keyword
	if parser.tokenIndex >= len(parser.tokens) ||
		parser.tokens[parser.tokenIndex].Type != TOKEN_NAME ||
		parser.tokens[parser.tokenIndex].Value != "as" {
		return nil, fmt.Errorf("expected 'as' after template path at line %d", importLine)
	}
	parser.tokenIndex++

	// Get the alias name
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
		return nil, fmt.Errorf("expected identifier after 'as' at line %d", importLine)
	}

	alias := parser.tokens[parser.tokenIndex].Value
	parser.tokenIndex++

	// Expect block end
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
		return nil, fmt.Errorf("expected block end token after import statement at line %d", importLine)
	}
	parser.tokenIndex++

	// Create import node
	return NewImportNode(templateExpr, alias, importLine), nil
}
