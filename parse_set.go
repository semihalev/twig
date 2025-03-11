package twig

import "fmt"

func (p *Parser) parseSet(parser *Parser) (Node, error) {
	// Get the line number of the set token
	setLine := parser.tokens[parser.tokenIndex-2].Line

	// Get the variable name
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
		return nil, fmt.Errorf("expected variable name after set at line %d", setLine)
	}

	varName := parser.tokens[parser.tokenIndex].Value
	parser.tokenIndex++

	// Expect '='
	if parser.tokenIndex >= len(parser.tokens) ||
		parser.tokens[parser.tokenIndex].Type != TOKEN_OPERATOR ||
		parser.tokens[parser.tokenIndex].Value != "=" {
		return nil, fmt.Errorf("expected '=' after variable name at line %d", setLine)
	}
	parser.tokenIndex++

	// Parse the value expression
	valueExpr, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// For expressions like 5 + 10, we need to parse both sides and make a binary node
	// Check if there's an operator after the first token
	if parser.tokenIndex < len(parser.tokens) &&
		parser.tokens[parser.tokenIndex].Type == TOKEN_OPERATOR &&
		parser.tokens[parser.tokenIndex].Value != "=" {

		// Get the operator
		operator := parser.tokens[parser.tokenIndex].Value
		parser.tokenIndex++

		// Parse the right side
		rightExpr, err := parser.parseExpression()
		if err != nil {
			return nil, err
		}

		// Create a binary node
		valueExpr = NewBinaryNode(operator, valueExpr, rightExpr, setLine)
	}

	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after set expression at line %d", setLine)
	}
	parser.tokenIndex++

	// Create the set node
	setNode := NewSetNode(varName, valueExpr, setLine)

	return setNode, nil
}
