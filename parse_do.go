package twig

import (
	"fmt"
	"strconv"
)

func (p *Parser) parseDo(parser *Parser) (Node, error) {
	// Get the line number for error reporting
	doLine := parser.tokens[parser.tokenIndex-2].Line

	// Check for special case: assignment expressions
	// These need to be handled specially since they're not normal expressions
	if parser.tokenIndex < len(parser.tokens) &&
		parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {

		varName := parser.tokens[parser.tokenIndex].Value
		parser.tokenIndex++

		if parser.tokenIndex < len(parser.tokens) &&
			parser.tokens[parser.tokenIndex].Type == TOKEN_OPERATOR &&
			parser.tokens[parser.tokenIndex].Value == "=" {

			// Skip the equals sign
			parser.tokenIndex++

			// Parse the right side expression
			expr, err := parser.parseExpression()
			if err != nil {
				return nil, fmt.Errorf("error parsing expression in do assignment at line %d: %w", doLine, err)
			}

			// Make sure we have the closing tag
			if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
				return nil, fmt.Errorf("expecting end of do tag at line %d", doLine)
			}
			parser.tokenIndex++

			// Validate the variable name - it should not be a numeric literal
			if _, err := strconv.Atoi(varName); err == nil {
				return nil, fmt.Errorf("invalid variable name %q in do tag assignment at line %d", varName, doLine)
			}

			// Create a SetNode instead of DoNode for assignments
			return &SetNode{
				name:  varName,
				value: expr,
				line:  doLine,
			}, nil
		}

		// If it wasn't an assignment, backtrack to parse it as a normal expression
		parser.tokenIndex -= 1
	}

	// Parse the expression to be executed
	expr, err := parser.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("error parsing expression in do tag at line %d: %w", doLine, err)
	}

	// Make sure we have the closing tag
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expecting end of do tag at line %d", doLine)
	}
	parser.tokenIndex++

	// Create and return the DoNode
	return NewDoNode(expr, doLine), nil
}
