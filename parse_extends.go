package twig

import "fmt"

func (p *Parser) parseExtends(parser *Parser) (Node, error) {
	// Get the line number of the extends token
	extendsLine := parser.tokens[parser.tokenIndex-2].Line

	// Get the parent template expression
	parentExpr, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after extends at line %d", extendsLine)
	}
	parser.tokenIndex++

	// Create the extends node
	extendsNode := &ExtendsNode{
		parent: parentExpr,
		line:   extendsLine,
	}

	return extendsNode, nil
}
