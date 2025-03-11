package twig

import (
	"fmt"
)

// parseIf parses if/elseif/else/endif block structure
// Examples:
// {% if condition %}...{% endif %}
// {% if condition %}...{% else %}...{% endif %}
// {% if condition %}...{% elseif condition2 %}...{% else %}...{% endif %}
func (p *Parser) parseIf(parser *Parser) (Node, error) {
	// Get the line number of the if token
	ifLine := parser.tokens[parser.tokenIndex-2].Line

	// Parse the condition expression
	condition, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect the block end token (either regular or whitespace-trimming variant)
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
		return nil, fmt.Errorf("expected block end after if condition at line %d", ifLine)
	}
	parser.tokenIndex++

	// Parse the if body (statements between if and endif/else/elseif)
	ifBody, err := parser.parseOuterTemplate()
	if err != nil {
		return nil, err
	}

	// Initialize conditions and bodies arrays with the initial if condition and body
	conditions := []Node{condition}
	bodies := [][]Node{ifBody}
	var elseBody []Node

	// Keep track of whether we've seen an else block
	var hasElseBlock bool

	// Process subsequent tags (elseif, else, endif)
	for {
		// We expect a block start token for elseif, else, or endif
		if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START {
			return nil, fmt.Errorf("unexpected end of template, expected endif at line %d", ifLine)
		}
		parser.tokenIndex++

		// We expect a name token (elseif, else, or endif)
		if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
			return nil, fmt.Errorf("expected block name at line %d", parser.tokens[parser.tokenIndex-1].Line)
		}

		// Get the tag name
		blockName := parser.tokens[parser.tokenIndex].Value
		blockLine := parser.tokens[parser.tokenIndex].Line
		parser.tokenIndex++

		// Process based on the tag type
		if blockName == "elseif" {
			// Check if we've already seen an else block - elseif can't come after else
			if hasElseBlock {
				return nil, fmt.Errorf("unexpected elseif after else at line %d", blockLine)
			}

			// Handle elseif condition
			elseifCondition, err := parser.parseExpression()
			if err != nil {
				return nil, err
			}

			// Expect block end token
			if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
				return nil, fmt.Errorf("expected block end after elseif condition at line %d", blockLine)
			}
			parser.tokenIndex++

			// Parse the elseif body
			elseifBody, err := parser.parseOuterTemplate()
			if err != nil {
				return nil, err
			}

			// Add condition and body to our arrays
			conditions = append(conditions, elseifCondition)
			bodies = append(bodies, elseifBody)

			// Continue checking for more elseif/else/endif tags
		} else if blockName == "else" {
			// Check if we've already seen an else block - can't have multiple else blocks
			if hasElseBlock {
				return nil, fmt.Errorf("multiple else blocks found at line %d", blockLine)
			}

			// Mark that we've seen an else block
			hasElseBlock = true

			// Expect block end token
			if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
				return nil, fmt.Errorf("expected block end after else tag at line %d", blockLine)
			}
			parser.tokenIndex++

			// Parse the else body
			elseBody, err = parser.parseOuterTemplate()
			if err != nil {
				return nil, err
			}

			// After else, we need to find endif next (handled in next iteration)
		} else if blockName == "endif" {
			// Expect block end token
			if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
				return nil, fmt.Errorf("expected block end after endif at line %d", blockLine)
			}
			parser.tokenIndex++

			// We found the endif, we're done
			break
		} else {
			return nil, fmt.Errorf("expected elseif, else, or endif, got %s at line %d", blockName, blockLine)
		}
	}

	// Create and return the if node
	ifNode := &IfNode{
		conditions: conditions,
		bodies:     bodies,
		elseBranch: elseBody,
		line:       ifLine,
	}

	return ifNode, nil
}

// Helper function to check if a token type is a block end token
func isBlockEndToken(tokenType int) bool {
	return tokenType == TOKEN_BLOCK_END || tokenType == TOKEN_BLOCK_END_TRIM
}

// Helper function to check if a token is any kind of variable end token (regular or trim variant)
func isVarEndToken(tokenType int) bool {
	return tokenType == TOKEN_VAR_END || tokenType == TOKEN_VAR_END_TRIM
}
