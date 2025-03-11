package twig

import (
	"fmt"
)

func (p *Parser) parseApply(parser *Parser) (Node, error) {
	// Get the line number of the apply token
	applyLine := parser.tokens[parser.tokenIndex-2].Line

	// Parse the filter name
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
		return nil, fmt.Errorf("expected filter name after apply tag at line %d", applyLine)
	}

	filterName := parser.tokens[parser.tokenIndex].Value
	parser.tokenIndex++

	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
		return nil, fmt.Errorf("expected block end token after apply filter at line %d", applyLine)
	}
	parser.tokenIndex++

	// Parse the apply body
	applyBody, err := parser.parseOuterTemplate()
	if err != nil {
		return nil, err
	}

	// Expect endapply tag
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START {
		return nil, fmt.Errorf("expected endapply tag at line %d", applyLine)
	}
	parser.tokenIndex++

	// Expect 'endapply' token
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME || parser.tokens[parser.tokenIndex].Value != "endapply" {
		return nil, fmt.Errorf("expected 'endapply' at line %d", applyLine)
	}
	parser.tokenIndex++

	// Expect block end token
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
		return nil, fmt.Errorf("expected block end token after endapply at line %d", applyLine)
	}
	parser.tokenIndex++

	// Create apply node (no arguments for now)
	return NewApplyNode(applyBody, filterName, nil, applyLine), nil
}
