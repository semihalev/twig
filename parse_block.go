package twig

import "fmt"

func (p *Parser) parseBlock(parser *Parser) (Node, error) {
	// Get the line number of the block token
	blockLine := parser.tokens[parser.tokenIndex-2].Line

	// Get the block name
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
		return nil, fmt.Errorf("expected block name at line %d", blockLine)
	}

	blockName := parser.tokens[parser.tokenIndex].Value
	parser.tokenIndex++

	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after block name at line %d", blockLine)
	}
	parser.tokenIndex++

	// Parse the block body
	blockBody, err := parser.parseOuterTemplate()
	if err != nil {
		return nil, err
	}

	// Expect endblock tag
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START {
		return nil, fmt.Errorf("expected endblock tag at line %d", blockLine)
	}
	parser.tokenIndex++

	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME ||
		parser.tokens[parser.tokenIndex].Value != "endblock" {
		return nil, fmt.Errorf("expected endblock at line %d", parser.tokens[parser.tokenIndex-1].Line)
	}
	parser.tokenIndex++

	// Check for optional block name in endblock
	if parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {
		endBlockName := parser.tokens[parser.tokenIndex].Value
		if endBlockName != blockName {
			return nil, fmt.Errorf("mismatched block name, expected %s but got %s at line %d",
				blockName, endBlockName, parser.tokens[parser.tokenIndex].Line)
		}
		parser.tokenIndex++
	}

	// Expect the final block end token
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after endblock at line %d", parser.tokens[parser.tokenIndex-1].Line)
	}
	parser.tokenIndex++

	// Create the block node
	blockNode := &BlockNode{
		name: blockName,
		body: blockBody,
		line: blockLine,
	}

	return blockNode, nil
}
