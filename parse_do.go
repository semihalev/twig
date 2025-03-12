package twig

import (
	"fmt"
	"strconv"
)

func (p *Parser) parseDo(parser *Parser) (Node, error) {
	// Get the line number for error reporting
	doLine := parser.tokens[parser.tokenIndex-2].Line

	// Check if we have an empty do tag ({% do %})
	if parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_BLOCK_END {
		// Empty do tag is not valid
		return nil, fmt.Errorf("do tag cannot be empty at line %d", doLine)
	}

	// Check for special case: assignment expressions
	// These need to be handled specially since they're not normal expressions
	if parser.tokenIndex < len(parser.tokens) {
		// Look ahead to find possible assignment patterns
		// We need to check for NUMBER = EXPR which is invalid
		// as well as NAME = EXPR which is valid

		// Check if we have an equals sign in the next few tokens
		hasAssignment := false
		equalsPosition := -1

		// Scan ahead a bit to find possible equals sign
		for i := 0; i < 3 && parser.tokenIndex+i < len(parser.tokens); i++ {
			token := parser.tokens[parser.tokenIndex+i]
			if token.Type == TOKEN_OPERATOR && token.Value == "=" {
				hasAssignment = true
				equalsPosition = i
				break
			}

			// Stop scanning if we hit the end of the block
			if token.Type == TOKEN_BLOCK_END {
				break
			}
		}

		// If we found an equals sign, analyze the left-hand side
		if hasAssignment && equalsPosition > 0 {
			firstToken := parser.tokens[parser.tokenIndex]

			// Check if the left-hand side is a valid variable name
			isValidVariableName := firstToken.Type == TOKEN_NAME

			// If the left-hand side is a number or literal, that's an error
			if !isValidVariableName {
				return nil, fmt.Errorf("invalid variable name %q in do tag assignment at line %d", firstToken.Value, doLine)
			}

			// Handle assignment case
			if isValidVariableName && hasAssignment {
				varName := parser.tokens[parser.tokenIndex].Value

				// Skip tokens up to and including the equals sign
				parser.tokenIndex += equalsPosition + 1

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

				// Additional validation for variable name
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
		}
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
