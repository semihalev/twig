package twig

import (
	"fmt"
	"strings"
)

// parseFrom handles the from tag which imports macros from another template
// Example: {% from "macros.twig" import input, button %}
// Example: {% from "macros.twig" import input as field, button as btn %}
func (p *Parser) parseFrom(parser *Parser) (Node, error) {
	// Get the line number of the from token
	fromLine := parser.tokens[parser.tokenIndex-1].Line
	
	// Debugging: Print out tokens for debugging purposes
	if IsDebugEnabled() {
		LogDebug("Parsing from tag. Next tokens (up to 10):")
		for i := 0; i < 10 && i+parser.tokenIndex < len(parser.tokens); i++ {
			if parser.tokenIndex+i < len(parser.tokens) {
				token := parser.tokens[parser.tokenIndex+i]
				LogDebug("  Token %d: Type=%d, Value=%q", i, token.Type, token.Value)
			}
		}
	}
	
	// First try to parse in token-by-token format (from zero-allocation tokenizer)
	// Check if we have a proper sequence: [STRING/NAME, NAME="import", NAME="macroname", ...]
	if parser.tokenIndex+1 < len(parser.tokens) {
		// Check for (template path) followed by "import" keyword
		firstToken := parser.tokens[parser.tokenIndex]
		secondToken := parser.tokens[parser.tokenIndex+1]
		
		isTemplatePath := firstToken.Type == TOKEN_STRING || firstToken.Type == TOKEN_NAME
		isImportKeyword := secondToken.Type == TOKEN_NAME && secondToken.Value == "import"
		
		if isTemplatePath && isImportKeyword {
			LogDebug("Found tokenized from...import pattern")
			
			// Get template path
			templatePath := firstToken.Value
			if firstToken.Type == TOKEN_NAME {
				// For paths like ./file.twig, just use the value
				templatePath = strings.Trim(templatePath, "\"'")
			}
			
			// Create template expression
			templateExpr := &LiteralNode{
				ExpressionNode: ExpressionNode{
					exprType: ExprLiteral,
					line:     fromLine,
				},
				value: templatePath,
			}
			
			// Skip past the template path and import keyword
			parser.tokenIndex += 2
			
			// Parse macros and aliases
			macros := []string{}
			aliases := map[string]string{}
			
			// Process tokens until end of block
			for parser.tokenIndex < len(parser.tokens) {
				token := parser.tokens[parser.tokenIndex]
				
				// Stop at block end
				if token.Type == TOKEN_BLOCK_END || token.Type == TOKEN_BLOCK_END_TRIM {
					parser.tokenIndex++
					break
				}
				
				// Skip punctuation (commas)
				if token.Type == TOKEN_PUNCTUATION {
					parser.tokenIndex++
					continue
				}
				
				// Handle macro name
				if token.Type == TOKEN_NAME {
					macroName := token.Value
					
					// Add to macros list
					macros = append(macros, macroName)
					
					// Check for alias
					parser.tokenIndex++
					if parser.tokenIndex < len(parser.tokens) && 
					   parser.tokens[parser.tokenIndex].Type == TOKEN_NAME &&
					   parser.tokens[parser.tokenIndex].Value == "as" {
						
						// Skip 'as' keyword
						parser.tokenIndex++
						
						// Get alias
						if parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {
							aliases[macroName] = parser.tokens[parser.tokenIndex].Value
							parser.tokenIndex++
						}
					}
				} else {
					// Skip any other token
					parser.tokenIndex++
				}
			}
			
			// If we found macros, return a FromImportNode
			if len(macros) > 0 {
				return NewFromImportNode(templateExpr, macros, aliases, fromLine), nil
			}
		}
	}
	
	// Fall back to the original approach (for backward compatibility)
	// We need to extract the template path, import keyword, and macro(s) from
	// the current token. The tokenizer may be combining them.
	if parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {
		// Extract parts from the combined token value
		tokenValue := parser.tokens[parser.tokenIndex].Value

		// Try to extract template path and remaining parts
		matches := strings.Split(tokenValue, " import ")
		if len(matches) == 2 {
			// We found the import keyword in the token value
			templatePath := strings.TrimSpace(matches[0])
			// Remove quotes if present
			templatePath = strings.Trim(templatePath, "\"'")
			macrosList := strings.TrimSpace(matches[1])

			// Create template expression
			templateExpr := &LiteralNode{
				ExpressionNode: ExpressionNode{
					exprType: ExprLiteral,
					line:     fromLine,
				},
				value: templatePath,
			}

			// Parse macros list
			macros := []string{}
			aliases := map[string]string{}

			// Split macros by comma if multiple
			macroItems := strings.Split(macrosList, ",")
			for _, item := range macroItems {
				item = strings.TrimSpace(item)

				// Check for "as" alias
				asParts := strings.Split(item, " as ")
				if len(asParts) == 2 {
					// We have an alias
					macroName := strings.TrimSpace(asParts[0])
					aliasName := strings.TrimSpace(asParts[1])
					aliases[macroName] = aliasName
					// Still add the macro name to macros list, even with alias
					macros = append(macros, macroName)
				} else {
					// No alias
					macros = append(macros, item)
				}
			}

			// Skip the current token
			parser.tokenIndex++

			// Skip to the block end token
			for parser.tokenIndex < len(parser.tokens) {
				if parser.tokens[parser.tokenIndex].Type == TOKEN_BLOCK_END ||
					parser.tokens[parser.tokenIndex].Type == TOKEN_BLOCK_END_TRIM {
					parser.tokenIndex++
					break
				}
				parser.tokenIndex++
			}

			// Create and return the FromImportNode
			return NewFromImportNode(templateExpr, macros, aliases, fromLine), nil
		}
	}

	// If we're here, the standard parsing approach failed, so return an error
	return nil, fmt.Errorf("expected 'import' after template path at line %d", fromLine)
}
