package twig

import (
	"fmt"
	"strings"
)

func (p *Parser) parseInclude(parser *Parser) (Node, error) {
	// Get the line number of the include token
	includeLine := parser.tokens[parser.tokenIndex-2].Line

	// Get the template expression
	templateExpr, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// Check for optional parameters
	var variables map[string]Node
	var ignoreMissing bool
	var onlyContext bool

	// Look for 'with', 'ignore missing', or 'only'
	for parser.tokenIndex < len(parser.tokens) &&
		parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {

		keyword := parser.tokens[parser.tokenIndex].Value
		parser.tokenIndex++

		switch keyword {
		case "with":
			// Parse variables as a hash
			if variables == nil {
				variables = make(map[string]Node)
			}

			// Check for opening brace
			if parser.tokenIndex < len(parser.tokens) &&
				parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION &&
				parser.tokens[parser.tokenIndex].Value == "{" {
				parser.tokenIndex++ // Skip opening brace

				// Parse key-value pairs
				for {
					// If we see a closing brace, we're done
					if parser.tokenIndex < len(parser.tokens) &&
						parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION &&
						parser.tokens[parser.tokenIndex].Value == "}" {
						parser.tokenIndex++ // Skip closing brace
						break
					}

					// Get the variable name - can be string literal or name token
					var varName string
					if parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_STRING {
						// It's a quoted string key
						varName = parser.tokens[parser.tokenIndex].Value
						parser.tokenIndex++
					} else if parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {
						// It's an unquoted key
						varName = parser.tokens[parser.tokenIndex].Value
						parser.tokenIndex++
					} else {
						return nil, fmt.Errorf("expected variable name or string at line %d", includeLine)
					}

					// Expect colon or equals
					if parser.tokenIndex >= len(parser.tokens) ||
						((parser.tokens[parser.tokenIndex].Type != TOKEN_PUNCTUATION &&
							parser.tokens[parser.tokenIndex].Value != ":") &&
							(parser.tokens[parser.tokenIndex].Type != TOKEN_OPERATOR &&
								parser.tokens[parser.tokenIndex].Value != "=")) {
						return nil, fmt.Errorf("expected ':' or '=' after variable name at line %d", includeLine)
					}
					parser.tokenIndex++ // Skip : or =

					// Parse the value expression
					varExpr, err := parser.parseExpression()
					if err != nil {
						return nil, err
					}

					// Add to variables map
					variables[varName] = varExpr

					// If there's a comma, skip it
					if parser.tokenIndex < len(parser.tokens) &&
						parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION &&
						parser.tokens[parser.tokenIndex].Value == "," {
						parser.tokenIndex++
					}

					// If we see whitespace, skip it
					for parser.tokenIndex < len(parser.tokens) &&
						parser.tokens[parser.tokenIndex].Type == TOKEN_TEXT &&
						strings.TrimSpace(parser.tokens[parser.tokenIndex].Value) == "" {
						parser.tokenIndex++
					}
				}
			} else {
				// If there's no opening brace, expect name-value pairs in the old format
				for parser.tokenIndex < len(parser.tokens) &&
					parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {

					// Get the variable name
					varName := parser.tokens[parser.tokenIndex].Value
					parser.tokenIndex++

					// Expect '='
					if parser.tokenIndex >= len(parser.tokens) ||
						parser.tokens[parser.tokenIndex].Type != TOKEN_OPERATOR ||
						parser.tokens[parser.tokenIndex].Value != "=" {
						return nil, fmt.Errorf("expected '=' after variable name at line %d", includeLine)
					}
					parser.tokenIndex++

					// Parse the value expression
					varExpr, err := parser.parseExpression()
					if err != nil {
						return nil, err
					}

					// Add to variables map
					variables[varName] = varExpr

					// If there's a comma, skip it
					if parser.tokenIndex < len(parser.tokens) &&
						parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION &&
						parser.tokens[parser.tokenIndex].Value == "," {
						parser.tokenIndex++
					} else {
						break
					}
				}
			}

		case "ignore":
			// Check for 'missing' keyword
			if parser.tokenIndex >= len(parser.tokens) ||
				parser.tokens[parser.tokenIndex].Type != TOKEN_NAME ||
				parser.tokens[parser.tokenIndex].Value != "missing" {
				return nil, fmt.Errorf("expected 'missing' after 'ignore' at line %d", includeLine)
			}
			parser.tokenIndex++

			ignoreMissing = true

		case "only":
			onlyContext = true

		default:
			return nil, fmt.Errorf("unexpected keyword '%s' in include at line %d", keyword, includeLine)
		}
	}

	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
		return nil, fmt.Errorf("expected block end token after include at line %d, found token type %d with value '%s'",
			includeLine,
			parser.tokens[parser.tokenIndex].Type,
			parser.tokens[parser.tokenIndex].Value)
	}
	parser.tokenIndex++

	// Create the include node
	includeNode := &IncludeNode{
		template:      templateExpr,
		variables:     variables,
		ignoreMissing: ignoreMissing,
		only:          onlyContext,
		line:          includeLine,
	}

	return includeNode, nil
}
