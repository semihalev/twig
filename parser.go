package twig

import (
	"fmt"
	"strconv"
	"strings"
)

// Token types
const (
	TOKEN_TEXT = iota
	TOKEN_VAR_START
	TOKEN_VAR_END
	TOKEN_BLOCK_START
	TOKEN_BLOCK_END
	TOKEN_COMMENT_START
	TOKEN_COMMENT_END
	TOKEN_NAME
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_OPERATOR
	TOKEN_PUNCTUATION
	TOKEN_EOF
)

// Parser handles parsing Twig templates into node trees
type Parser struct {
	source        string
	tokens        []Token
	tokenIndex    int
	filename      string
	cursor        int
	line          int
	blockHandlers map[string]blockHandlerFunc
}

type blockHandlerFunc func(*Parser) (Node, error)

// Token represents a lexical token
type Token struct {
	Type  int
	Value string
	Line  int
}

// Parse parses a template source into a node tree
func (p *Parser) Parse(source string) (Node, error) {
	p.source = source
	p.cursor = 0
	p.line = 1
	p.tokenIndex = 0
	
	// Initialize default block handlers
	p.initBlockHandlers()
	
	// Tokenize source
	var err error
	p.tokens, err = p.tokenize()
	if err != nil {
		return nil, err
	}
	
	// Parse tokens into nodes
	nodes, err := p.parseOuterTemplate()
	if err != nil {
		return nil, err
	}
	
	return NewRootNode(nodes, 1), nil
}

// Initialize block handlers for different tag types
func (p *Parser) initBlockHandlers() {
	p.blockHandlers = map[string]blockHandlerFunc{
		"if":      p.parseIf,
		"for":     p.parseFor,
		"block":   p.parseBlock,
		"extends": p.parseExtends,
		"include": p.parseInclude,
		"set":     p.parseSet,
		"do":      p.parseDo,
		"macro":   p.parseMacro,
		"import":  p.parseImport,
		"from":    p.parseFrom,
	}
}

// Tokenize the source into a list of tokens
func (p *Parser) tokenize() ([]Token, error) {
	var tokens []Token
	
	for p.cursor < len(p.source) {
		// Check for variable syntax {{ }}
		if p.matchString("{{") {
			tokens = append(tokens, Token{Type: TOKEN_VAR_START, Line: p.line})
			p.cursor += 2
			// Skip whitespace after opening braces
			for p.cursor < len(p.source) && isWhitespace(p.current()) {
				if p.current() == '\n' {
					p.line++
				}
				p.cursor++
			}
			continue
		}
		
		if p.matchString("}}") {
			tokens = append(tokens, Token{Type: TOKEN_VAR_END, Line: p.line})
			p.cursor += 2
			continue
		}
		
		// Check for block syntax {% %}
		if p.matchString("{%") {
			tokens = append(tokens, Token{Type: TOKEN_BLOCK_START, Line: p.line})
			p.cursor += 2
			// Skip whitespace after opening braces
			for p.cursor < len(p.source) && isWhitespace(p.current()) {
				if p.current() == '\n' {
					p.line++
				}
				p.cursor++
			}
			continue
		}
		
		if p.matchString("%}") {
			tokens = append(tokens, Token{Type: TOKEN_BLOCK_END, Line: p.line})
			p.cursor += 2
			continue
		}
		
		// Check for comment syntax {# #}
		if p.matchString("{#") {
			tokens = append(tokens, Token{Type: TOKEN_COMMENT_START, Line: p.line})
			p.cursor += 2
			// Skip whitespace after opening braces
			for p.cursor < len(p.source) && isWhitespace(p.current()) {
				if p.current() == '\n' {
					p.line++
				}
				p.cursor++
			}
			continue
		}
		
		if p.matchString("#}") {
			tokens = append(tokens, Token{Type: TOKEN_COMMENT_END, Line: p.line})
			p.cursor += 2
			continue
		}
		
		// Check for string literals
		if p.current() == '"' || p.current() == '\'' {
			quote := p.current()
			p.cursor++
			
			start := p.cursor
			for p.cursor < len(p.source) && p.current() != quote {
				if p.current() == '\\' {
					p.cursor++
				}
				if p.current() == '\n' {
					p.line++
				}
				p.cursor++
			}
			
			if p.cursor >= len(p.source) {
				return nil, fmt.Errorf("unterminated string at line %d", p.line)
			}
			
			value := p.source[start:p.cursor]
			tokens = append(tokens, Token{Type: TOKEN_STRING, Value: value, Line: p.line})
			p.cursor++
			continue
		}
		
		// Check for numbers
		if isDigit(p.current()) {
			start := p.cursor
			for p.cursor < len(p.source) && (isDigit(p.current()) || p.current() == '.') {
				p.cursor++
			}
			
			value := p.source[start:p.cursor]
			tokens = append(tokens, Token{Type: TOKEN_NUMBER, Value: value, Line: p.line})
			continue
		}
		
		// Check for identifiers/names
		if isAlpha(p.current()) {
			start := p.cursor
			for p.cursor < len(p.source) && isAlphaNumeric(p.current()) {
				p.cursor++
			}
			
			value := p.source[start:p.cursor]
			tokens = append(tokens, Token{Type: TOKEN_NAME, Value: value, Line: p.line})
			continue
		}
		
		// Check for operators
		if isOperator(p.current()) {
			start := p.cursor
			for p.cursor < len(p.source) && isOperator(p.current()) {
				p.cursor++
			}
			
			value := p.source[start:p.cursor]
			tokens = append(tokens, Token{Type: TOKEN_OPERATOR, Value: value, Line: p.line})
			continue
		}
		
		// Check for punctuation
		if isPunctuation(p.current()) {
			tokens = append(tokens, Token{
				Type:  TOKEN_PUNCTUATION,
				Value: string(p.current()),
				Line:  p.line,
			})
			p.cursor++
			continue
		}
		
		// Check for whitespace and newlines
		if isWhitespace(p.current()) {
			if p.current() == '\n' {
				p.line++
			}
			p.cursor++
			continue
		}
		
		// Handle plain text
		start := p.cursor
		for p.cursor < len(p.source) &&
			!p.matchString("{{") && !p.matchString("}}") &&
			!p.matchString("{%") && !p.matchString("%}") &&
			!p.matchString("{#") && !p.matchString("#}") {
			if p.current() == '\n' {
				p.line++
			}
			p.cursor++
		}
		
		if start != p.cursor {
			value := p.source[start:p.cursor]
			tokens = append(tokens, Token{Type: TOKEN_TEXT, Value: value, Line: p.line})
		}
	}
	
	tokens = append(tokens, Token{Type: TOKEN_EOF, Line: p.line})
	return tokens, nil
}

// Helper methods for tokenization
func (p *Parser) current() byte {
	if p.cursor >= len(p.source) {
		return 0
	}
	return p.source[p.cursor]
}

func (p *Parser) matchString(s string) bool {
	if p.cursor+len(s) > len(p.source) {
		return false
	}
	return p.source[p.cursor:p.cursor+len(s)] == s
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

func isOperator(c byte) bool {
	return strings.ContainsRune("+-*/=<>!&|~^%", rune(c))
}

func isPunctuation(c byte) bool {
	return strings.ContainsRune("()[]{},.:", rune(c))
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// Parse the outer level of a template (text, print tags, blocks)
func (p *Parser) parseOuterTemplate() ([]Node, error) {
	var nodes []Node
	
	for p.tokenIndex < len(p.tokens) && p.tokens[p.tokenIndex].Type != TOKEN_EOF {
		token := p.tokens[p.tokenIndex]
		
		switch token.Type {
		case TOKEN_TEXT:
			nodes = append(nodes, NewTextNode(token.Value, token.Line))
			p.tokenIndex++
			
		case TOKEN_VAR_START:
			p.tokenIndex++
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			
			nodes = append(nodes, NewPrintNode(expr, token.Line))
			
			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != TOKEN_VAR_END {
				return nil, fmt.Errorf("expected }} at line %d", token.Line)
			}
			p.tokenIndex++
			
		case TOKEN_BLOCK_START:
			p.tokenIndex++
			
			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected block name at line %d", token.Line)
			}
			
			blockName := p.tokens[p.tokenIndex].Value
			p.tokenIndex++
			
			// Check if we have a handler for this block type
			handler, ok := p.blockHandlers[blockName]
			if !ok {
				return nil, fmt.Errorf("unknown block type '%s' at line %d", blockName, token.Line)
			}
			
			node, err := handler(p)
			if err != nil {
				return nil, err
			}
			
			nodes = append(nodes, node)
			
		case TOKEN_COMMENT_START:
			// Skip comments
			p.tokenIndex++
			startLine := token.Line
			
			// Find the end of the comment
			for p.tokenIndex < len(p.tokens) && p.tokens[p.tokenIndex].Type != TOKEN_COMMENT_END {
				p.tokenIndex++
			}
			
			if p.tokenIndex >= len(p.tokens) {
				return nil, fmt.Errorf("unclosed comment starting at line %d", startLine)
			}
			
			p.tokenIndex++
			
		default:
			return nil, fmt.Errorf("unexpected token %v at line %d", token.Type, token.Line)
		}
	}
	
	return nodes, nil
}

// Parse an expression
func (p *Parser) parseExpression() (Node, error) {
	// Placeholder - implement actual expression parsing
	// This is a simplified version that just handles literals and variables
	
	if p.tokenIndex >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of template")
	}
	
	token := p.tokens[p.tokenIndex]
	
	switch token.Type {
	case TOKEN_STRING:
		p.tokenIndex++
		return NewLiteralNode(token.Value, token.Line), nil
		
	case TOKEN_NUMBER:
		p.tokenIndex++
		// Attempt to convert to int or float
		if strings.Contains(token.Value, ".") {
			// It's a float
			// Note: error handling omitted for brevity
			val, _ := strconv.ParseFloat(token.Value, 64)
			return NewLiteralNode(val, token.Line), nil
		} else {
			// It's an int
			val, _ := strconv.Atoi(token.Value)
			return NewLiteralNode(val, token.Line), nil
		}
		
	case TOKEN_NAME:
		p.tokenIndex++
		
		// First create the variable node
		var result Node = NewVariableNode(token.Value, token.Line)
		
		// Check for attribute access (obj.attr)
		for p.tokenIndex < len(p.tokens) && 
			p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION && 
			p.tokens[p.tokenIndex].Value == "." {
			
			p.tokenIndex++
			
			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected attribute name at line %d", token.Line)
			}
			
			attrName := p.tokens[p.tokenIndex].Value
			attrNode := NewLiteralNode(attrName, p.tokens[p.tokenIndex].Line)
			result = NewGetAttrNode(result, attrNode, token.Line)
			p.tokenIndex++
		}
		
		return result, nil
		
	default:
		return nil, fmt.Errorf("unexpected token in expression at line %d", token.Line)
	}
}

// Parse if statement
func (p *Parser) parseIf(parser *Parser) (Node, error) {
	// Get the line number of the if token
	ifLine := parser.tokens[parser.tokenIndex-2].Line
	
	// Parse the condition expression
	condition, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}
	
	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end after if condition at line %d", ifLine)
	}
	parser.tokenIndex++
	
	// Parse the if body (statements between if and endif/else)
	ifBody, err := parser.parseOuterTemplate()
	if err != nil {
		return nil, err
	}
	
	var elseBody []Node
	
	// Check for else or endif
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
			
			// Now expect the endif
			if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START {
				return nil, fmt.Errorf("expected endif block at line %d", parser.tokens[parser.tokenIndex-1].Line)
			}
			parser.tokenIndex++
			
			if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected endif at line %d", parser.tokens[parser.tokenIndex-1].Line)
			}
			
			if parser.tokens[parser.tokenIndex].Value != "endif" {
				return nil, fmt.Errorf("expected endif, got %s at line %d", parser.tokens[parser.tokenIndex].Value, parser.tokens[parser.tokenIndex].Line)
			}
			parser.tokenIndex++
		} else if parser.tokens[parser.tokenIndex].Value == "endif" {
			parser.tokenIndex++
		} else {
			return nil, fmt.Errorf("expected else or endif, got %s at line %d", parser.tokens[parser.tokenIndex].Value, parser.tokens[parser.tokenIndex].Line)
		}
		
		// Expect the final block end token
		if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
			return nil, fmt.Errorf("expected block end after endif at line %d", parser.tokens[parser.tokenIndex-1].Line)
		}
		parser.tokenIndex++
	} else {
		return nil, fmt.Errorf("unexpected end of template, expected endif at line %d", ifLine)
	}
	
	// Create the if node
	ifNode := &IfNode{
		conditions: []Node{condition},
		bodies:     [][]Node{ifBody},
		elseBranch: elseBody,
		line:       ifLine,
	}
	
	return ifNode, nil
}

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
	
	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
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

func (p *Parser) parseInclude(parser *Parser) (Node, error) {
	// Get the line number of the include token
	includeLine := parser.tokens[parser.tokenIndex-2].Line
	
	// Get the template expression
	templateExpr, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}
	
	// Check for optional parameters
	var variables map[string]Node
	var ignoreMissing bool
	var onlyContext bool
	
	// Look for 'with', 'ignore missing', or 'only'
	for parser.tokenIndex < len(parser.tokens) && 
		parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {
		
		keyword := parser.tokens[parser.tokenIndex].Value
		parser.tokenIndex++
		
		switch keyword {
		case "with":
			// Parse variables as a hash
			if variables == nil {
				variables = make(map[string]Node)
			}
			
			// Parse the variable assignments
			for parser.tokenIndex < len(parser.tokens) && 
				parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {
				
				// Get the variable name
				varName := parser.tokens[parser.tokenIndex].Value
				parser.tokenIndex++
				
				// Expect '='
				if parser.tokenIndex >= len(parser.tokens) || 
					parser.tokens[parser.tokenIndex].Type != TOKEN_OPERATOR || 
					parser.tokens[parser.tokenIndex].Value != "=" {
					return nil, fmt.Errorf("expected '=' after variable name at line %d", includeLine)
				}
				parser.tokenIndex++
				
				// Parse the value expression
				varExpr, err := parser.parseExpression()
				if err != nil {
					return nil, err
				}
				
				// Add to variables map
				variables[varName] = varExpr
				
				// If there's a comma, skip it
				if parser.tokenIndex < len(parser.tokens) && 
					parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION && 
					parser.tokens[parser.tokenIndex].Value == "," {
					parser.tokenIndex++
				} else {
					break
				}
			}
			
		case "ignore":
			// Check for 'missing' keyword
			if parser.tokenIndex >= len(parser.tokens) || 
				parser.tokens[parser.tokenIndex].Type != TOKEN_NAME || 
				parser.tokens[parser.tokenIndex].Value != "missing" {
				return nil, fmt.Errorf("expected 'missing' after 'ignore' at line %d", includeLine)
			}
			parser.tokenIndex++
			
			ignoreMissing = true
			
		case "only":
			onlyContext = true
			
		default:
			return nil, fmt.Errorf("unexpected keyword '%s' in include at line %d", keyword, includeLine)
		}
	}
	
	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after include at line %d", includeLine)
	}
	parser.tokenIndex++
	
	// Create the include node
	includeNode := &IncludeNode{
		template:     templateExpr,
		variables:    variables,
		ignoreMissing: ignoreMissing,
		only:         onlyContext,
		line:         includeLine,
	}
	
	return includeNode, nil
}

func (p *Parser) parseSet(parser *Parser) (Node, error) {
	// Placeholder for set parsing
	return nil, fmt.Errorf("set not implemented yet")
}

func (p *Parser) parseDo(parser *Parser) (Node, error) {
	// Placeholder for do parsing
	return nil, fmt.Errorf("do not implemented yet")
}

func (p *Parser) parseMacro(parser *Parser) (Node, error) {
	// Placeholder for macro parsing
	return nil, fmt.Errorf("macros not implemented yet")
}

func (p *Parser) parseImport(parser *Parser) (Node, error) {
	// Placeholder for import parsing
	return nil, fmt.Errorf("import not implemented yet")
}

func (p *Parser) parseFrom(parser *Parser) (Node, error) {
	// Placeholder for from parsing
	return nil, fmt.Errorf("from not implemented yet")
}