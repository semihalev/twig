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

// Placeholder methods for block handlers - to be implemented
func (p *Parser) parseIf(parser *Parser) (Node, error) {
	// Placeholder for if block parsing
	return nil, fmt.Errorf("if blocks not implemented yet")
}

func (p *Parser) parseFor(parser *Parser) (Node, error) {
	// Placeholder for for loop parsing
	return nil, fmt.Errorf("for loops not implemented yet")
}

func (p *Parser) parseBlock(parser *Parser) (Node, error) {
	// Placeholder for block definition parsing
	return nil, fmt.Errorf("blocks not implemented yet")
}

func (p *Parser) parseExtends(parser *Parser) (Node, error) {
	// Placeholder for extends parsing
	return nil, fmt.Errorf("extends not implemented yet")
}

func (p *Parser) parseInclude(parser *Parser) (Node, error) {
	// Placeholder for include parsing
	return nil, fmt.Errorf("include not implemented yet")
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