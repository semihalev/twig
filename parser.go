package twig

import (
	"fmt"
	"strconv"
	"strings"
)

// Token types
const (
	TOKEN_TEXT          = iota
	TOKEN_VAR_START     // {{
	TOKEN_VAR_END       // }}
	TOKEN_BLOCK_START   // {%
	TOKEN_BLOCK_END     // %}
	TOKEN_COMMENT_START // {#
	TOKEN_COMMENT_END   // #}
	TOKEN_NAME
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_OPERATOR
	TOKEN_PUNCTUATION
	TOKEN_EOF

	// Whitespace control token types
	TOKEN_VAR_START_TRIM   // {{-
	TOKEN_VAR_END_TRIM     // -}}
	TOKEN_BLOCK_START_TRIM // {%-
	TOKEN_BLOCK_END_TRIM   // -%}
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
		return nil, fmt.Errorf("tokenization error: %w", err)
	}

	// Debug tokenization output
	/*
		fmt.Println("Tokenized template:")
		for i, t := range p.tokens {
			fmt.Printf("Token %d: Type=%d, Value=%q, Line=%d\n", i, t.Type, t.Value, t.Line)
		}
	*/

	// Apply whitespace control processing to the tokens to handle
	// the whitespace trimming between template elements
	p.tokens = processWhitespaceControl(p.tokens)

	// Parse tokens into nodes
	nodes, err := p.parseOuterTemplate()
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	return NewRootNode(nodes, 1), nil
}

// Initialize block handlers for different tag types
func (p *Parser) initBlockHandlers() {
	p.blockHandlers = map[string]blockHandlerFunc{
		"if":        p.parseIf,
		"for":       p.parseFor,
		"block":     p.parseBlock,
		"extends":   p.parseExtends,
		"include":   p.parseInclude,
		"set":       p.parseSet,
		"do":        p.parseDo,
		"macro":     p.parseMacro,
		"import":    p.parseImport,
		"from":      p.parseFrom,
		"spaceless": p.parseSpaceless,

		// Special closing tags - they will be handled in their corresponding open tag parsers
		"endif":        p.parseEndTag,
		"endfor":       p.parseEndTag,
		"endmacro":     p.parseEndTag,
		"endblock":     p.parseEndTag,
		"endspaceless": p.parseEndTag,
		"else":         p.parseEndTag,
		"elseif":       p.parseEndTag,
	}
}

// Tokenize the source into a list of tokens
func (p *Parser) tokenize() ([]Token, error) {
	var tokens []Token

	for p.cursor < len(p.source) {
		// Check for variable syntax with whitespace control {{ }} or {{- -}}
		if p.matchString("{{-") {
			tokens = append(tokens, Token{Type: TOKEN_VAR_START_TRIM, Line: p.line})
			p.cursor += 3
			// Skip whitespace after opening braces
			for p.cursor < len(p.source) && isWhitespace(p.current()) {
				if p.current() == '\n' {
					p.line++
				}
				p.cursor++
			}
			continue
		} else if p.matchString("{{") {
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

		if p.matchString("-}}") {
			tokens = append(tokens, Token{Type: TOKEN_VAR_END_TRIM, Line: p.line})
			p.cursor += 3
			continue
		} else if p.matchString("}}") {
			tokens = append(tokens, Token{Type: TOKEN_VAR_END, Line: p.line})
			p.cursor += 2
			continue
		}

		// Check for block syntax with whitespace control {% %} or {%- -%}
		if p.matchString("{%-") {
			tokens = append(tokens, Token{Type: TOKEN_BLOCK_START_TRIM, Line: p.line})
			p.cursor += 3
			// Skip whitespace after opening braces
			for p.cursor < len(p.source) && isWhitespace(p.current()) {
				if p.current() == '\n' {
					p.line++
				}
				p.cursor++
			}
			continue
		} else if p.matchString("{%") {
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

		if p.matchString("-%}") {
			tokens = append(tokens, Token{Type: TOKEN_BLOCK_END_TRIM, Line: p.line})
			p.cursor += 3
			continue
		} else if p.matchString("%}") {
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
			var inEmbeddedVar bool = false
			for p.cursor < len(p.source) && (p.current() != quote || inEmbeddedVar) {
				// Handle embedded Twig syntax like {{ }}
				if p.cursor+1 < len(p.source) && p.current() == '{' && (p.source[p.cursor+1] == '{' || p.source[p.cursor+1] == '%') {
					inEmbeddedVar = true
				}

				// Check for end of embedded variable
				if inEmbeddedVar && p.cursor+1 < len(p.source) && p.current() == '}' && (p.source[p.cursor+1] == '}' || p.source[p.cursor+1] == '%') {
					p.cursor += 2 // Skip the closing brackets
					inEmbeddedVar = false
					continue
				}

				// Skip escaped quote characters
				if p.current() == '\\' && p.cursor+1 < len(p.source) {
					// Skip the backslash and the next character (which might be a quote)
					p.cursor += 2
					continue
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
			!p.matchString("{{-") && !p.matchString("{{") &&
			!p.matchString("-}}") && !p.matchString("}}") &&
			!p.matchString("{%-") && !p.matchString("{%") &&
			!p.matchString("-%}") && !p.matchString("%}") &&
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

// Helper function to check if a token is any kind of block end token (regular or trim variant)
func isBlockEndToken(tokenType int) bool {
	return tokenType == TOKEN_BLOCK_END || tokenType == TOKEN_BLOCK_END_TRIM
}

// Helper function to check if a token is any kind of variable end token (regular or trim variant)
func isVarEndToken(tokenType int) bool {
	return tokenType == TOKEN_VAR_END || tokenType == TOKEN_VAR_END_TRIM
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
	return strings.ContainsRune("+-*/=<>!&~^%?:", rune(c))
}

func isPunctuation(c byte) bool {
	return strings.ContainsRune("()[]{},.:|", rune(c))
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// processEscapeSequences handles escape sequences in string literals
func processEscapeSequences(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			case '\\':
				result.WriteByte('\\')
			case '"':
				result.WriteByte('"')
			case '\'':
				result.WriteByte('\'')
			default:
				result.WriteByte(s[i])
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}

// Replace HTML attributes like type="{{ type }}" with actual Twig variables in HTML
func fixHTMLAttributes(input string) string {
	// Search for patterns like: type="{{ type }}" or name="{{ name }}"
	for i := 0; i < len(input); i++ {
		// Find potential attribute patterns
		attrStart := strings.Index(input[i:], "=\"{{")
		if attrStart == -1 {
			break // No more attributes with embedded variables
		}

		attrStart += i // Adjust to full string position

		// Find the end of the attribute value
		attrEnd := strings.Index(input[attrStart+3:], "}}\"")
		if attrEnd == -1 {
			break // No closing variable
		}

		attrEnd += attrStart + 3 // Adjust to full string position

		// Extract the variable name (between {{ and }})
		varName := strings.TrimSpace(input[attrStart+3 : attrEnd])

		// Replace the attribute string with an empty string for now
		// We'll need to handle this specially in the parsing logic
		input = input[:attrStart] + "=" + varName + input[attrEnd+2:]
	}

	return input
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

		case TOKEN_VAR_START, TOKEN_VAR_START_TRIM:
			// Handle both normal and whitespace trimming var start tokens
			p.tokenIndex++
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			nodes = append(nodes, NewPrintNode(expr, token.Line))

			// Check for either normal or whitespace trimming var end tokens
			if p.tokenIndex >= len(p.tokens) || !isVarEndToken(p.tokens[p.tokenIndex].Type) {
				return nil, fmt.Errorf("expected }} or -}} at line %d", token.Line)
			}
			p.tokenIndex++

		case TOKEN_BLOCK_START, TOKEN_BLOCK_START_TRIM:
			// Handle both normal and whitespace trimming block start tokens
			p.tokenIndex++

			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected block name at line %d", token.Line)
			}

			blockName := p.tokens[p.tokenIndex].Value
			p.tokenIndex++

			// Check if this is a control ending tag (endif, endfor, endblock, etc.)
			if blockName == "endif" || blockName == "endfor" || blockName == "endblock" ||
				blockName == "endmacro" || blockName == "else" || blockName == "elseif" ||
				blockName == "endspaceless" {
				// We should return to the parent parser that's handling the parent block
				// First move back two steps to the start of the block tag
				p.tokenIndex -= 2
				return nodes, nil
			}

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

		// Add special handling for trim token types
		case TOKEN_VAR_END_TRIM, TOKEN_BLOCK_END_TRIM:
			// These should have been handled with their corresponding start tokens
			return nil, fmt.Errorf("unexpected token %v at line %d", token.Type, token.Line)

		// Add special handling for TOKEN_NAME outside of a tag
		case TOKEN_NAME, TOKEN_PUNCTUATION, TOKEN_OPERATOR, TOKEN_STRING, TOKEN_NUMBER:
			// For raw names, punctuation, operators, and literals not inside tags, convert to text
			// In many languages, the text "true" is a literal boolean, but in our parser it's just a name token
			// outside of an expression context

			// Special handling for text content words - add spaces between consecutive text tokens
			// This fixes issues with the spaceless tag's handling of text content
			if token.Type == TOKEN_NAME && p.tokenIndex+1 < len(p.tokens) &&
				p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
				p.tokens[p.tokenIndex+1].Line == token.Line {
				// Look ahead for consecutive name tokens and join them with spaces
				var textContent strings.Builder
				textContent.WriteString(token.Value)

				currentLine := token.Line
				p.tokenIndex++ // Skip the first token as we've already added it

				// Collect consecutive name tokens on the same line
				for p.tokenIndex < len(p.tokens) &&
					p.tokens[p.tokenIndex].Type == TOKEN_NAME &&
					p.tokens[p.tokenIndex].Line == currentLine {
					textContent.WriteString(" ") // Add space between words
					textContent.WriteString(p.tokens[p.tokenIndex].Value)
					p.tokenIndex++
				}

				nodes = append(nodes, NewTextNode(textContent.String(), token.Line))
			} else {
				// Regular handling for single text tokens
				nodes = append(nodes, NewTextNode(token.Value, token.Line))
				p.tokenIndex++
			}

		default:
			return nil, fmt.Errorf("unexpected token %v at line %d", token.Type, token.Line)
		}
	}

	return nodes, nil
}

// Parse an expression
func (p *Parser) parseExpression() (Node, error) {
	// Parse the primary expression first
	expr, err := p.parseSimpleExpression()
	if err != nil {
		return nil, err
	}

	// Now check for filter operator (|)
	if p.tokenIndex < len(p.tokens) &&
		p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
		p.tokens[p.tokenIndex].Value == "|" {

		expr, err = p.parseFilters(expr)
		if err != nil {
			return nil, err
		}
	}

	// Check for binary operators (and, or, ==, !=, <, >, etc.)
	if p.tokenIndex < len(p.tokens) &&
		(p.tokens[p.tokenIndex].Type == TOKEN_OPERATOR ||
			(p.tokens[p.tokenIndex].Type == TOKEN_NAME &&
				(p.tokens[p.tokenIndex].Value == "and" ||
					p.tokens[p.tokenIndex].Value == "or" ||
					p.tokens[p.tokenIndex].Value == "in" ||
					p.tokens[p.tokenIndex].Value == "not" ||
					p.tokens[p.tokenIndex].Value == "is" ||
					p.tokens[p.tokenIndex].Value == "matches" ||
					p.tokens[p.tokenIndex].Value == "starts" ||
					p.tokens[p.tokenIndex].Value == "ends"))) {

		expr, err = p.parseBinaryExpression(expr)
		if err != nil {
			return nil, err
		}
	}

	// Check for ternary operator (? :)
	if p.tokenIndex < len(p.tokens) &&
		p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
		p.tokens[p.tokenIndex].Value == "?" {

		return p.parseConditionalExpression(expr)
	}

	return expr, nil
}

// Parse ternary conditional expression (condition ? true_expr : false_expr)
func (p *Parser) parseConditionalExpression(condition Node) (Node, error) {
	line := p.tokens[p.tokenIndex].Line

	// Skip the "?" token
	p.tokenIndex++

	// Parse the "true" expression
	trueExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect ":" token
	if p.tokenIndex >= len(p.tokens) ||
		p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
		p.tokens[p.tokenIndex].Value != ":" {
		return nil, fmt.Errorf("expected ':' after true expression in conditional at line %d", line)
	}
	p.tokenIndex++ // Skip ":"

	// Parse the "false" expression
	falseExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Create a conditional node
	return &ConditionalNode{
		ExpressionNode: ExpressionNode{
			exprType: ExprConditional,
			line:     line,
		},
		condition: condition,
		trueExpr:  trueExpr,
		falseExpr: falseExpr,
	}, nil
}

// Parse a simple expression (literal, variable, function call, array)
func (p *Parser) parseSimpleExpression() (Node, error) {
	if p.tokenIndex >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of template")
	}

	token := p.tokens[p.tokenIndex]

	switch token.Type {
	case TOKEN_STRING:
		p.tokenIndex++
		// For string literals, process escape sequences
		processedValue := processEscapeSequences(token.Value)
		return NewLiteralNode(processedValue, token.Line), nil

	case TOKEN_NUMBER:
		p.tokenIndex++
		// Attempt to convert to int or float
		if strings.Contains(token.Value, ".") {
			// It's a float
			val, _ := strconv.ParseFloat(token.Value, 64)
			return NewLiteralNode(val, token.Line), nil
		} else {
			// It's an int
			val, _ := strconv.Atoi(token.Value)
			return NewLiteralNode(val, token.Line), nil
		}

	case TOKEN_NAME:
		p.tokenIndex++

		// Store the variable name for function calls
		varName := token.Value
		varLine := token.Line

		// Special handling for boolean literals and null
		if varName == "true" {
			return NewLiteralNode(true, varLine), nil
		} else if varName == "false" {
			return NewLiteralNode(false, varLine), nil
		} else if varName == "null" || varName == "nil" {
			return NewLiteralNode(nil, varLine), nil
		}

		// Check if this is a function call (name followed by opening parenthesis)
		if p.tokenIndex < len(p.tokens) &&
			p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
			p.tokens[p.tokenIndex].Value == "(" {

			// This is a function call
			p.tokenIndex++ // Skip the opening parenthesis

			// Parse arguments list
			var args []Node

			// If there are arguments (not empty parentheses)
			if p.tokenIndex < len(p.tokens) &&
				!(p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
					p.tokens[p.tokenIndex].Value == ")") {

				for {
					// Parse each argument expression
					argExpr, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					args = append(args, argExpr)

					// Check for comma separator between arguments
					if p.tokenIndex < len(p.tokens) &&
						p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
						p.tokens[p.tokenIndex].Value == "," {
						p.tokenIndex++ // Skip comma
						continue
					}

					// No comma, so must be end of argument list
					break
				}
			}

			// Expect closing parenthesis
			if p.tokenIndex >= len(p.tokens) ||
				p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
				p.tokens[p.tokenIndex].Value != ")" {
				return nil, fmt.Errorf("expected closing parenthesis after function arguments at line %d", varLine)
			}
			p.tokenIndex++ // Skip closing parenthesis

			// Create and return function node
			return NewFunctionNode(varName, args, varLine), nil
		}

		// If not a function call, it's a regular variable
		var result Node = NewVariableNode(varName, varLine)

		// Check for attribute access (obj.attr)
		for p.tokenIndex < len(p.tokens) &&
			p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
			p.tokens[p.tokenIndex].Value == "." {

			p.tokenIndex++

			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected attribute name at line %d", varLine)
			}

			attrName := p.tokens[p.tokenIndex].Value
			attrNode := NewLiteralNode(attrName, p.tokens[p.tokenIndex].Line)
			result = NewGetAttrNode(result, attrNode, varLine)
			p.tokenIndex++
		}

		return result, nil

	case TOKEN_PUNCTUATION:
		// Handle array literals [1, 2, 3]
		if token.Value == "[" {
			return p.parseArrayExpression()
		}

		// Handle parenthesized expressions
		if token.Value == "(" {
			p.tokenIndex++ // Skip "("
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			// Expect closing parenthesis
			if p.tokenIndex >= len(p.tokens) ||
				p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
				p.tokens[p.tokenIndex].Value != ")" {
				return nil, fmt.Errorf("expected closing parenthesis at line %d", token.Line)
			}
			p.tokenIndex++ // Skip ")"

			return expr, nil
		}

	default:
		return nil, fmt.Errorf("unexpected token in expression at line %d", token.Line)
	}

	return nil, fmt.Errorf("unexpected token in expression at line %d", token.Line)
}

// Parse array expression [item1, item2, ...]
func (p *Parser) parseArrayExpression() (Node, error) {
	// Save the line number for error reporting
	line := p.tokens[p.tokenIndex].Line

	// Skip the opening bracket
	p.tokenIndex++

	// Parse the array items
	var items []Node

	// Check if there are any items
	if p.tokenIndex < len(p.tokens) &&
		!(p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
			p.tokens[p.tokenIndex].Value == "]") {

		for {
			// Parse each item expression
			itemExpr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			items = append(items, itemExpr)

			// Check for comma separator between items
			if p.tokenIndex < len(p.tokens) &&
				p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
				p.tokens[p.tokenIndex].Value == "," {
				p.tokenIndex++ // Skip comma
				continue
			}

			// No comma, so must be end of array
			break
		}
	}

	// Expect closing bracket
	if p.tokenIndex >= len(p.tokens) ||
		p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
		p.tokens[p.tokenIndex].Value != "]" {
		return nil, fmt.Errorf("expected closing bracket after array items at line %d", line)
	}
	p.tokenIndex++ // Skip closing bracket

	// Create array node
	return &ArrayNode{
		ExpressionNode: ExpressionNode{
			exprType: ExprArray,
			line:     line,
		},
		items: items,
	}, nil
}

// Parse filter expressions: variable|filter(args)
func (p *Parser) parseFilters(node Node) (Node, error) {
	line := p.tokens[p.tokenIndex].Line

	// Loop to handle multiple filters (e.g. var|filter1|filter2)
	for p.tokenIndex < len(p.tokens) &&
		p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
		p.tokens[p.tokenIndex].Value == "|" {

		p.tokenIndex++ // Skip the | token

		// Expect filter name
		if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != TOKEN_NAME {
			return nil, fmt.Errorf("expected filter name at line %d", line)
		}

		filterName := p.tokens[p.tokenIndex].Value
		p.tokenIndex++

		// Check for filter arguments
		var args []Node

		// If there are arguments in parentheses
		if p.tokenIndex < len(p.tokens) &&
			p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
			p.tokens[p.tokenIndex].Value == "(" {

			p.tokenIndex++ // Skip opening parenthesis

			// Parse arguments
			if p.tokenIndex < len(p.tokens) &&
				!(p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
					p.tokens[p.tokenIndex].Value == ")") {

				for {
					// Parse each argument expression
					argExpr, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					args = append(args, argExpr)

					// Check for comma separator
					if p.tokenIndex < len(p.tokens) &&
						p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
						p.tokens[p.tokenIndex].Value == "," {
						p.tokenIndex++ // Skip comma
						continue
					}

					// No comma, so end of argument list
					break
				}
			}

			// Expect closing parenthesis
			if p.tokenIndex >= len(p.tokens) ||
				p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
				p.tokens[p.tokenIndex].Value != ")" {
				return nil, fmt.Errorf("expected closing parenthesis after filter arguments at line %d", line)
			}
			p.tokenIndex++ // Skip closing parenthesis
		}

		// Create a new FilterNode
		node = &FilterNode{
			ExpressionNode: ExpressionNode{
				exprType: ExprFilter,
				line:     line,
			},
			node:   node,
			filter: filterName,
			args:   args,
		}
	}

	return node, nil
}

// Parse binary expressions (a + b, a and b, a in b, etc.)
func (p *Parser) parseBinaryExpression(left Node) (Node, error) {
	token := p.tokens[p.tokenIndex]
	operator := token.Value
	line := token.Line

	// Process multi-word operators
	if token.Type == TOKEN_NAME {
		// Handle 'not in' operator
		if token.Value == "not" && p.tokenIndex+1 < len(p.tokens) &&
			p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
			p.tokens[p.tokenIndex+1].Value == "in" {
			operator = "not in"
			p.tokenIndex += 2 // Skip both 'not' and 'in'
		} else if token.Value == "is" && p.tokenIndex+1 < len(p.tokens) &&
			p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
			p.tokens[p.tokenIndex+1].Value == "not" {
			// Handle 'is not' operator
			operator = "is not"
			p.tokenIndex += 2 // Skip both 'is' and 'not'
		} else if token.Value == "starts" && p.tokenIndex+1 < len(p.tokens) &&
			p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
			p.tokens[p.tokenIndex+1].Value == "with" {
			// Handle 'starts with' operator
			operator = "starts with"
			p.tokenIndex += 2 // Skip both 'starts' and 'with'
		} else if token.Value == "ends" && p.tokenIndex+1 < len(p.tokens) &&
			p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
			p.tokens[p.tokenIndex+1].Value == "with" {
			// Handle 'ends with' operator
			operator = "ends with"
			p.tokenIndex += 2 // Skip both 'ends' and 'with'
		} else {
			// Single word operators like 'is', 'and', 'or', 'in', 'matches'
			p.tokenIndex++ // Skip the operator token
		}
	} else {
		// Regular operators like +, -, *, /, etc.
		p.tokenIndex++ // Skip the operator token
	}

	// Handle 'is' followed by a test
	if operator == "is" || operator == "is not" {
		// Check if this is a test
		if p.tokenIndex < len(p.tokens) && p.tokens[p.tokenIndex].Type == TOKEN_NAME {
			testName := p.tokens[p.tokenIndex].Value
			p.tokenIndex++ // Skip the test name

			// Parse test arguments if any
			var args []Node

			// If there's an opening parenthesis, parse arguments
			if p.tokenIndex < len(p.tokens) &&
				p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
				p.tokens[p.tokenIndex].Value == "(" {

				p.tokenIndex++ // Skip opening parenthesis

				// Parse arguments
				if p.tokenIndex < len(p.tokens) &&
					!(p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
						p.tokens[p.tokenIndex].Value == ")") {

					for {
						// Parse each argument expression
						argExpr, err := p.parseExpression()
						if err != nil {
							return nil, err
						}
						args = append(args, argExpr)

						// Check for comma separator
						if p.tokenIndex < len(p.tokens) &&
							p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
							p.tokens[p.tokenIndex].Value == "," {
							p.tokenIndex++ // Skip comma
							continue
						}

						// No comma, so end of argument list
						break
					}
				}

				// Expect closing parenthesis
				if p.tokenIndex >= len(p.tokens) ||
					p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
					p.tokens[p.tokenIndex].Value != ")" {
					return nil, fmt.Errorf("expected closing parenthesis after test arguments at line %d", line)
				}
				p.tokenIndex++ // Skip closing parenthesis
			}

			// Create the test node
			test := &TestNode{
				ExpressionNode: ExpressionNode{
					exprType: ExprTest,
					line:     line,
				},
				node: left,
				test: testName,
				args: args,
			}

			// If it's a negated test (is not), create a unary 'not' node
			if operator == "is not" {
				return &UnaryNode{
					ExpressionNode: ExpressionNode{
						exprType: ExprUnary,
						line:     line,
					},
					operator: "not",
					node:     test,
				}, nil
			}

			return test, nil
		}
	}

	// If we get here, we have a regular binary operator

	// For regular binary operators, parse the right operand
	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return NewBinaryNode(operator, left, right, line), nil
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

	// Expect the block end token (either regular or whitespace-trimming variant)
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
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

		// Expect the final block end token (either regular or trim variant)
		if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
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
		template:      templateExpr,
		variables:     variables,
		ignoreMissing: ignoreMissing,
		only:          onlyContext,
		line:          includeLine,
	}

	return includeNode, nil
}

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

func (p *Parser) parseDo(parser *Parser) (Node, error) {
	// Placeholder for do parsing
	return nil, fmt.Errorf("do not implemented yet")
}

func (p *Parser) parseMacro(parser *Parser) (Node, error) {
	// Get the line number of the macro token
	macroLine := parser.tokens[parser.tokenIndex-2].Line

	// Get the macro name
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
		return nil, fmt.Errorf("expected macro name after macro keyword at line %d", macroLine)
	}

	macroName := parser.tokens[parser.tokenIndex].Value
	parser.tokenIndex++

	// Expect opening parenthesis for parameters
	if parser.tokenIndex >= len(parser.tokens) ||
		parser.tokens[parser.tokenIndex].Type != TOKEN_PUNCTUATION ||
		parser.tokens[parser.tokenIndex].Value != "(" {
		return nil, fmt.Errorf("expected '(' after macro name at line %d", macroLine)
	}
	parser.tokenIndex++

	// Parse parameters
	var params []string
	defaults := make(map[string]Node)

	// If we don't have a closing parenthesis immediately, we have parameters
	if parser.tokenIndex < len(parser.tokens) &&
		(parser.tokens[parser.tokenIndex].Type != TOKEN_PUNCTUATION ||
			parser.tokens[parser.tokenIndex].Value != ")") {

		for {
			// Get parameter name
			if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected parameter name at line %d", macroLine)
			}

			paramName := parser.tokens[parser.tokenIndex].Value
			params = append(params, paramName)
			parser.tokenIndex++

			// Check for default value
			if parser.tokenIndex < len(parser.tokens) &&
				parser.tokens[parser.tokenIndex].Type == TOKEN_OPERATOR &&
				parser.tokens[parser.tokenIndex].Value == "=" {
				parser.tokenIndex++ // Skip =

				// Parse default value expression
				defaultExpr, err := parser.parseExpression()
				if err != nil {
					return nil, err
				}

				defaults[paramName] = defaultExpr
			}

			// Check if we have more parameters
			if parser.tokenIndex < len(parser.tokens) &&
				parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION &&
				parser.tokens[parser.tokenIndex].Value == "," {
				parser.tokenIndex++ // Skip comma
				continue
			}

			break
		}
	}

	// Expect closing parenthesis
	if parser.tokenIndex >= len(parser.tokens) ||
		parser.tokens[parser.tokenIndex].Type != TOKEN_PUNCTUATION ||
		parser.tokens[parser.tokenIndex].Value != ")" {
		return nil, fmt.Errorf("expected ')' after macro parameters at line %d", macroLine)
	}
	parser.tokenIndex++

	// Expect block end
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after macro declaration at line %d", macroLine)
	}
	parser.tokenIndex++

	// Parse the macro body
	bodyNodes, err := parser.parseOuterTemplate()
	if err != nil {
		return nil, err
	}

	// Expect endmacro tag
	if parser.tokenIndex+1 >= len(parser.tokens) ||
		parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START ||
		parser.tokens[parser.tokenIndex+1].Type != TOKEN_NAME ||
		parser.tokens[parser.tokenIndex+1].Value != "endmacro" {
		return nil, fmt.Errorf("missing endmacro tag for macro '%s' at line %d",
			macroName, macroLine)
	}

	// Skip {% endmacro %}
	parser.tokenIndex += 2 // Skip {% endmacro

	// Expect block end
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after endmacro at line %d", parser.tokens[parser.tokenIndex].Line)
	}
	parser.tokenIndex++

	// Create the macro node
	return NewMacroNode(macroName, params, defaults, bodyNodes, macroLine), nil
}

func (p *Parser) parseImport(parser *Parser) (Node, error) {
	// Get the line number of the import token
	importLine := parser.tokens[parser.tokenIndex-2].Line

	// Get the template to import
	templateExpr, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect 'as' keyword
	if parser.tokenIndex >= len(parser.tokens) ||
		parser.tokens[parser.tokenIndex].Type != TOKEN_NAME ||
		parser.tokens[parser.tokenIndex].Value != "as" {
		return nil, fmt.Errorf("expected 'as' after template path at line %d", importLine)
	}
	parser.tokenIndex++

	// Get the alias name
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
		return nil, fmt.Errorf("expected identifier after 'as' at line %d", importLine)
	}

	alias := parser.tokens[parser.tokenIndex].Value
	parser.tokenIndex++

	// Expect block end
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after import statement at line %d", importLine)
	}
	parser.tokenIndex++

	// Create import node
	return NewImportNode(templateExpr, alias, importLine), nil
}

func (p *Parser) parseFrom(parser *Parser) (Node, error) {
	// Get the line number of the from token
	fromLine := parser.tokens[parser.tokenIndex-2].Line

	// Get the template to import from
	templateExpr, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect 'import' keyword
	if parser.tokenIndex >= len(parser.tokens) ||
		parser.tokens[parser.tokenIndex].Type != TOKEN_NAME ||
		parser.tokens[parser.tokenIndex].Value != "import" {
		return nil, fmt.Errorf("expected 'import' after template path at line %d", fromLine)
	}
	parser.tokenIndex++

	// Parse the imported items
	var macros []string
	aliases := make(map[string]string)

	// We need at least one macro to import
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
		return nil, fmt.Errorf("expected at least one identifier after 'import' at line %d", fromLine)
	}

	for parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {
		// Get macro name
		macroName := parser.tokens[parser.tokenIndex].Value
		parser.tokenIndex++

		// Check for 'as' keyword for aliasing
		if parser.tokenIndex < len(parser.tokens) &&
			parser.tokens[parser.tokenIndex].Type == TOKEN_NAME &&
			parser.tokens[parser.tokenIndex].Value == "as" {
			parser.tokenIndex++ // Skip 'as'

			// Get alias name
			if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected identifier after 'as' at line %d", fromLine)
			}

			aliasName := parser.tokens[parser.tokenIndex].Value
			aliases[macroName] = aliasName
			parser.tokenIndex++
		} else {
			// No alias, just add to macros list
			macros = append(macros, macroName)
		}

		// Check for comma to separate items
		if parser.tokenIndex < len(parser.tokens) &&
			parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION &&
			parser.tokens[parser.tokenIndex].Value == "," {
			parser.tokenIndex++ // Skip comma

			// Expect another identifier after comma
			if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected identifier after ',' at line %d", fromLine)
			}
		} else {
			// End of imports
			break
		}
	}

	// Expect block end
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after from import statement at line %d", fromLine)
	}
	parser.tokenIndex++

	// Create from import node
	return NewFromImportNode(templateExpr, macros, aliases, fromLine), nil
}

// parseEndTag handles closing tags like endif, endfor, endblock, etc.
// These tags should only be encountered inside their respective block parsing methods,
// so if we reach here directly, it's an error.
func (p *Parser) parseEndTag(parser *Parser) (Node, error) {
	// Get the line number and tag name
	tagLine := parser.tokens[parser.tokenIndex-2].Line
	tagName := parser.tokens[parser.tokenIndex-1].Value

	return nil, fmt.Errorf("unexpected '%s' tag at line %d", tagName, tagLine)
}

// parseSpaceless parses a spaceless block
func (p *Parser) parseSpaceless(parser *Parser) (Node, error) {
	// Get the line number of the spaceless token
	spacelessLine := parser.tokens[parser.tokenIndex-2].Line

	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
		return nil, fmt.Errorf("expected block end token after spaceless at line %d", spacelessLine)
	}
	parser.tokenIndex++

	// Parse the spaceless body
	spacelessBody, err := parser.parseOuterTemplate()
	if err != nil {
		return nil, err
	}

	// Expect endspaceless tag
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START {
		return nil, fmt.Errorf("expected endspaceless tag at line %d", spacelessLine)
	}
	parser.tokenIndex++

	// Expect the endspaceless token
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME ||
		parser.tokens[parser.tokenIndex].Value != "endspaceless" {
		return nil, fmt.Errorf("expected endspaceless token at line %d", parser.tokens[parser.tokenIndex].Line)
	}
	parser.tokenIndex++

	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
		return nil, fmt.Errorf("expected block end token after endspaceless at line %d", parser.tokens[parser.tokenIndex].Line)
	}
	parser.tokenIndex++

	// Create and return the spaceless node
	return NewSpacelessNode(spacelessBody, spacelessLine), nil
}
