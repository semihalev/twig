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

	// We need to manually extract the template path, import keyword, and macro(s) from
	// the current token. The tokenizer seems to be combining them.
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
