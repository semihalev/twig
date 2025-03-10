// Code generated by lexgen; DO NOT EDIT.
package twig

import (
	"fmt"
)

// Lexer tokenizes a template string
type Lexer struct {
	source string
	tokens []Token
	line   int
	col    int
	pos    int
}

// NewLexer creates a new lexer
func NewLexer(source string) *Lexer {
	return &Lexer{
		source: source,
		tokens: []Token{},
		line:   1,
		col:    1,
		pos:    0,
	}
}

// Tokenize tokenizes the source and returns tokens
func (l *Lexer) Tokenize() ([]Token, error) {
	// Skip if already tokenized
	if len(l.tokens) > 0 {
		return l.tokens, nil
	}

	for l.pos < len(l.source) {
		switch {
		case l.source[l.pos] == '\n':
			// Newline
			l.line++
			l.col = 1
			l.pos++
			
		case l.isWhitespace(l.source[l.pos]):
			// Skip whitespace
			l.col++
			l.pos++
			
		case l.match("{{"):
			// Variable start
			l.addToken(T_VAR_START, "{{")
			l.advance(2)

			// After variable start, scan for identifiers, operators, etc.
			l.scanExpressionTokens()
			
		case l.match("}}"):
			// Variable end
			l.addToken(T_VAR_END, "}}")
			l.advance(2)
			
		case l.match("{%"):
			// Block start
			l.addToken(T_BLOCK_START, "{%")
			l.advance(2)

			// After block start, scan for identifiers, operators, etc.
			l.scanExpressionTokens()
			
		case l.match("%}"):
			// Block end
			l.addToken(T_BLOCK_END, "%}")
			l.advance(2)
			
		default:
			// Text content (anything that's not a special delimiter)
			start := l.pos
			for l.pos < len(l.source) && 
				!l.match("{{") && !l.match("}}") && 
				!l.match("{%") && !l.match("%}") {
				if l.source[l.pos] == '\n' {
					l.line++
					l.col = 1
				} else {
					l.col++
				}
				l.pos++
			}
			
			if start != l.pos {
				l.addToken(T_TEXT, l.source[start:l.pos])
			} else {
				// If we get here, we have an unrecognized character
				return nil, fmt.Errorf("unexpected character %q at line %d, column %d",
					l.source[l.pos], l.line, l.col)
			}
		}
	}

	// Add EOF token
	l.addToken(T_EOF, "")
	
	return l.tokens, nil
}

// scanExpressionTokens scans for tokens inside expressions (between {{ }} or {% %})
func (l *Lexer) scanExpressionTokens() {
	// Skip leading whitespace
	for l.pos < len(l.source) && l.isWhitespace(l.source[l.pos]) {
		l.advance(1)
	}

	// Continue scanning until we reach the end tag
	for l.pos < len(l.source) && 
		!l.match("}}") && !l.match("%}") {
		
		// Skip whitespace
		if l.isWhitespace(l.source[l.pos]) {
			l.advance(1)
			continue
		}

		// Scan for different token types
		switch {
		case l.isAlpha(l.source[l.pos]):
			// Identifier or keyword
			l.scanIdentifierOrKeyword()

		case l.isDigit(l.source[l.pos]):
			// Number
			l.scanNumber()

		case l.source[l.pos] == '"' || l.source[l.pos] == '\'':
			// String literal
			l.scanString()

		case l.isPunctuation(l.source[l.pos]):
			// Punctuation
			l.addToken(T_PUNCTUATION, string(l.source[l.pos]))
			l.advance(1)

		case l.isOperator(l.source[l.pos]):
			// Operator
			l.scanOperator()

		default:
			// Skip any other characters
			l.advance(1)
		}
	}
}

// scanIdentifierOrKeyword scans an identifier or keyword
func (l *Lexer) scanIdentifierOrKeyword() {
	start := l.pos
	
	// First character is already checked to be alpha
	l.advance(1)
	
	// Keep scanning alphanumeric and underscore characters
	for l.pos < len(l.source) && (l.isAlphaNumeric(l.source[l.pos]) || l.source[l.pos] == '_') {
		l.advance(1)
	}
	
	// Extract the identifier
	text := l.source[start:l.pos]
	
	// Check if it's a keyword
	switch text {
	case "macro":
		l.addToken(T_MACRO, text)
	case "endmacro":
		l.addToken(T_ENDMACRO, text)
	case "import":
		l.addToken(T_IMPORT, text)
	case "from":
		l.addToken(T_FROM, text)
	case "as":
		l.addToken(T_AS, text)
	case "with":
		l.addToken(T_WITH, text)
	case "only":
		l.addToken(T_ONLY, text)
	case "ignore":
		l.addToken(T_IGNORE, text)
	case "missing":
		l.addToken(T_MISSING, text)
	case "in":
		l.addToken(T_IN, text)
	default:
		l.addToken(T_IDENT, text)
	}
}

// scanNumber scans a number (integer or float)
func (l *Lexer) scanNumber() {
	start := l.pos
	
	// Keep scanning digits
	for l.pos < len(l.source) && l.isDigit(l.source[l.pos]) {
		l.advance(1)
	}
	
	// Look for fractional part
	if l.pos < len(l.source) && l.source[l.pos] == '.' && 
		l.pos+1 < len(l.source) && l.isDigit(l.source[l.pos+1]) {
		// Consume the dot
		l.advance(1)
		
		// Consume digits after the dot
		for l.pos < len(l.source) && l.isDigit(l.source[l.pos]) {
			l.advance(1)
		}
	}
	
	l.addToken(T_NUMBER, l.source[start:l.pos])
}

// scanString scans a string literal
func (l *Lexer) scanString() {
	start := l.pos
	quote := l.source[l.pos] // ' or "
	
	// Consume the opening quote
	l.advance(1)
	
	// Keep scanning until closing quote or end of file
	for l.pos < len(l.source) && l.source[l.pos] != quote {
		// Handle escape sequence
		if l.source[l.pos] == '\\' && l.pos+1 < len(l.source) {
			l.advance(2) // Skip the escape sequence
		} else {
			l.advance(1)
		}
	}
	
	// Consume the closing quote
	if l.pos < len(l.source) {
		l.advance(1)
	}
	
	l.addToken(T_STRING, l.source[start:l.pos])
}

// scanOperator scans an operator
func (l *Lexer) scanOperator() {
	// Check for multi-character operators first
	if l.pos+1 < len(l.source) {
		twoChars := l.source[l.pos:l.pos+2]
		
		switch twoChars {
		case "==", "!=", ">=", "<=", "&&", "||", "+=", "-=", "*=", "/=", "%=", "~=":
			l.addToken(T_OPERATOR, twoChars)
			l.advance(2)
			return
		}
	}
	
	// Single character operators
	l.addToken(T_OPERATOR, string(l.source[l.pos]))
	l.advance(1)
}

// Helper methods for character classification

func (l *Lexer) isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func (l *Lexer) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (l *Lexer) isAlphaNumeric(c byte) bool {
	return l.isAlpha(c) || l.isDigit(c)
}

func (l *Lexer) isPunctuation(c byte) bool {
	return c == '(' || c == ')' || c == '[' || c == ']' || c == '{' || c == '}' || 
		   c == ',' || c == '.' || c == ':' || c == ';'
}

func (l *Lexer) isOperator(c byte) bool {
	return c == '+' || c == '-' || c == '*' || c == '/' || c == '%' || c == '=' || 
		   c == '<' || c == '>' || c == '!' || c == '&' || c == '|' || c == '~'
}

// Helper methods

func (l *Lexer) addToken(typ TokenType, value string) {
	l.tokens = append(l.tokens, Token{
		Type:  typ,
		Value: value,
		Line:  l.line,
		Col:   l.col - len(value),
	})
}

func (l *Lexer) advance(n int) {
	l.pos += n
	l.col += n
}

func (l *Lexer) match(s string) bool {
	if l.pos+len(s) > len(l.source) {
		return false
	}
	return l.source[l.pos:l.pos+len(s)] == s
}

func (l *Lexer) isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r'
}