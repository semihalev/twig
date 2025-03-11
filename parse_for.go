package twig

import (
	"fmt"
)

// parseFor parses a for loop construct in Twig templates
// Examples:
// {% for item in items %}...{% endfor %}
// {% for key, value in items %}...{% endfor %}
// {% for item in items %}...{% else %}...{% endfor %}
func (p *Parser) parseFor(parser *Parser) (Node, error) {
	// Get the line number of the for token
	forLine := parser.tokens[parser.tokenIndex-2].Line

	// Parse the loop variable name(s)
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
		return nil, fmt.Errorf("expected variable name after for at line %d", forLine)
	}

	// Get value variable name
	valueVar := parser.tokens[parser.tokenIndex].Value
	parser.tokenIndex++

	var keyVar string

	// Check for key, value syntax
	if parser.tokenIndex < len(parser.tokens) &&
		parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION &&
		parser.tokens[parser.tokenIndex].Value == "," {

		// Move past the comma
		parser.tokenIndex++

		// Now valueVar is actually the key, and we need to get the value
		keyVar = valueVar

		if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
			return nil, fmt.Errorf("expected value variable name after comma at line %d", forLine)
		}

		valueVar = parser.tokens[parser.tokenIndex].Value
		parser.tokenIndex++
	}

	// Expect 'in' keyword
	if parser.tokenIndex >= len(parser.tokens) ||
		parser.tokens[parser.tokenIndex].Type != TOKEN_NAME ||
		parser.tokens[parser.tokenIndex].Value != "in" {
		return nil, fmt.Errorf("expected 'in' keyword after variable name at line %d", forLine)
	}
	parser.tokenIndex++

	// Parse the sequence expression
	sequence, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// Check for filter operator (|) - needed for cases where filter detection might be missed
	if IsDebugEnabled() {
		LogDebug("For loop sequence expression type: %T", sequence)
	}

	// Expect the block end token (either regular or trim variant)
	if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
		return nil, fmt.Errorf("expected block end after for statement at line %d", forLine)
	}
	parser.tokenIndex++

	// Parse the for loop body
	loopBody, err := parser.parseOuterTemplate()
	if err != nil {
		return nil, err
	}

	var elseBody []Node

	// Check for else or endfor
	if parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_BLOCK_START {
		parser.tokenIndex++

		if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
			return nil, fmt.Errorf("expected block name at line %d", parser.tokens[parser.tokenIndex-1].Line)
		}

		// Check if this is an else block
		if parser.tokens[parser.tokenIndex].Value == "else" {
			parser.tokenIndex++

			// Expect the block end token
			if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
				return nil, fmt.Errorf("expected block end after else at line %d", parser.tokens[parser.tokenIndex-1].Line)
			}
			parser.tokenIndex++

			// Parse the else body
			elseBody, err = parser.parseOuterTemplate()
			if err != nil {
				return nil, err
			}

			// Now expect the endfor
			if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START {
				return nil, fmt.Errorf("expected endfor block at line %d", parser.tokens[parser.tokenIndex-1].Line)
			}
			parser.tokenIndex++

			if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected endfor at line %d", parser.tokens[parser.tokenIndex-1].Line)
			}

			if parser.tokens[parser.tokenIndex].Value != "endfor" {
				return nil, fmt.Errorf("expected endfor, got %s at line %d", parser.tokens[parser.tokenIndex].Value, parser.tokens[parser.tokenIndex].Line)
			}
			parser.tokenIndex++
		} else if parser.tokens[parser.tokenIndex].Value == "endfor" {
			parser.tokenIndex++
		} else {
			return nil, fmt.Errorf("expected else or endfor, got %s at line %d", parser.tokens[parser.tokenIndex].Value, parser.tokens[parser.tokenIndex].Line)
		}

		// Expect the final block end token
		if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
			return nil, fmt.Errorf("expected block end after endfor at line %d", parser.tokens[parser.tokenIndex-1].Line)
		}
		parser.tokenIndex++
	} else {
		return nil, fmt.Errorf("unexpected end of template, expected endfor at line %d", forLine)
	}

	// Create the for node
	forNode := &ForNode{
		keyVar:     keyVar,
		valueVar:   valueVar,
		sequence:   sequence,
		body:       loopBody,
		elseBranch: elseBody,
		line:       forLine,
	}

	return forNode, nil
}
