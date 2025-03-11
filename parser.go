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

	// Use the HTML-preserving tokenizer to preserve HTML content exactly
	// This will treat everything outside twig tags as TEXT tokens
	var err error
	p.tokens, err = p.htmlPreservingTokenize()
	if err != nil {
		return nil, fmt.Errorf("tokenization error: %w", err)
	}

	// Template tokenization complete

	// Apply whitespace control processing to the tokens to handle
	// the whitespace trimming between template elements
	p.tokens = processWhitespaceControl(p.tokens)

	// Parse tokens into nodes
	nodes, err := p.parseOuterTemplate()
	if err != nil {
		// Clean up token slice on error
		ReleaseTokenSlice(p.tokens)
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	// Clean up token slice after successful parsing
	ReleaseTokenSlice(p.tokens)

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
		"verbatim":  p.parseVerbatim,

		// Special closing tags - they will be handled in their corresponding open tag parsers
		"endif":        p.parseEndTag,
		"endfor":       p.parseEndTag,
		"endmacro":     p.parseEndTag,
		"endblock":     p.parseEndTag,
		"endspaceless": p.parseEndTag,
		"else":         p.parseEndTag,
		"elseif":       p.parseEndTag,
		"endverbatim":  p.parseEndTag,
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
			startLine := p.line
			p.cursor++

			var sb strings.Builder

			for p.cursor < len(p.source) && p.current() != quote {
				// Handle escape sequences properly
				if p.current() == '\\' && p.cursor+1 < len(p.source) {
					p.cursor++ // Skip the backslash
					// Just collect the escaped character
					sb.WriteByte(p.current())
					p.cursor++
					continue
				}

				// Just add the character
				sb.WriteByte(p.current())

				if p.current() == '\n' {
					p.line++
				}
				p.cursor++
			}

			if p.cursor >= len(p.source) {
				return nil, fmt.Errorf("unterminated string at line %d", startLine)
			}

			tokens = append(tokens, Token{Type: TOKEN_STRING, Value: sb.String(), Line: startLine})
			p.cursor++ // Skip closing quote
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

		// Handle plain text - this is the entire HTML content
		// We should collect all text up to the next twig tag start
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
			// Get the text segment as a single token, preserving ALL characters
			// This is critical for HTML parsing - we do not want to tokenize HTML!
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
	return strings.ContainsRune("+-*/=<>!&~^%", rune(c))
}

func isPunctuation(c byte) bool {
	return strings.ContainsRune("()[]{},.:|?", rune(c))
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
			case '{':
				// Special case for escaping Twig variable/block syntax
				result.WriteByte('{')
			case '}':
				// Special case for escaping Twig variable/block syntax
				result.WriteByte('}')
			default:
				result.WriteByte(s[i])
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
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

	// Check for array access with square brackets
	for p.tokenIndex < len(p.tokens) &&
		p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
		p.tokens[p.tokenIndex].Value == "[" {

		// Get the line number for error reporting
		line := p.tokens[p.tokenIndex].Line

		// Skip the opening bracket
		p.tokenIndex++

		// Parse the index expression
		indexExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// Expect closing bracket
		if p.tokenIndex >= len(p.tokens) ||
			p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
			p.tokens[p.tokenIndex].Value != "]" {
			return nil, fmt.Errorf("expected closing bracket after array index at line %d", line)
		}
		p.tokenIndex++ // Skip closing bracket

		// Create a GetItemNode
		expr = NewGetItemNode(expr, indexExpr, line)
	}

	// Now check for filter operator (|)
	// Process all filters in a loop to handle consecutive filters properly
	for p.tokenIndex < len(p.tokens) &&
		p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
		p.tokens[p.tokenIndex].Value == "|" {

		expr, err = p.parseFilters(expr)
		if err != nil {
			return nil, err
		}
	}

	// Check for binary operators (and, or, ==, !=, <, >, etc.)
	// Loop to handle multiple binary operators in sequence, such as 'hello' ~ ' ' ~ 'world'
	for p.tokenIndex < len(p.tokens) &&
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

	// Handle unary operators like 'not' and unary minus/plus
	if (token.Type == TOKEN_NAME && token.Value == "not") ||
		(token.Type == TOKEN_OPERATOR && (token.Value == "-" || token.Value == "+")) {
		// Skip the operator token
		operator := token.Value
		p.tokenIndex++

		// Get the line number for the unary node
		line := token.Line

		// Parse the operand
		operand, err := p.parseSimpleExpression()
		if err != nil {
			return nil, err
		}

		// Create a unary node
		return NewUnaryNode(operator, operand, line), nil
	}

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

		// Check for attribute access (obj.attr) or method calls (obj.method())
		for p.tokenIndex < len(p.tokens) &&
			p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
			p.tokens[p.tokenIndex].Value == "." {

			p.tokenIndex++

			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != TOKEN_NAME {
				return nil, fmt.Errorf("expected attribute name at line %d", varLine)
			}

			attrName := p.tokens[p.tokenIndex].Value
			attrNode := NewLiteralNode(attrName, p.tokens[p.tokenIndex].Line)
			p.tokenIndex++

			// Check if this is a method call like (module.method())
			if p.tokenIndex < len(p.tokens) &&
				p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
				p.tokens[p.tokenIndex].Value == "(" {

				if IsDebugEnabled() && debugger.level >= DebugVerbose {
					LogVerbose("Detected module.method call: %s.%s(...)", varName, attrName)
				}

				// This is a method call with the method stored in attrName
				// We'll use the moduleExpr field in FunctionNode to store the module expression

				// Parse the arguments
				p.tokenIndex++ // Skip opening parenthesis

				// Parse arguments
				var args []Node

				// If there are arguments (not empty parentheses)
				if p.tokenIndex < len(p.tokens) &&
					!(p.tokenIndex < len(p.tokens) &&
						p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
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
					return nil, fmt.Errorf("expected closing parenthesis after method arguments at line %d", varLine)
				}
				p.tokenIndex++ // Skip closing parenthesis

				// Create a function call with the module expression and method name
				result = &FunctionNode{
					ExpressionNode: ExpressionNode{
						exprType: ExprFunction,
						line:     varLine,
					},
					name: attrName,
					args: args,
					// Special handling - We'll store the module in the FunctionNode
					moduleExpr: result,
				}
			} else {
				// Regular attribute access (not a method call)
				result = NewGetAttrNode(result, attrNode, varLine)
			}
		}

		return result, nil

	case TOKEN_PUNCTUATION:
		// Handle array literals [1, 2, 3]
		if token.Value == "[" {
			return p.parseArrayExpression()
		}

		// Handle hash/map literals {'key': value}
		if token.Value == "{" {
			return p.parseMapExpression()
		}

		// Handle parenthesized expressions
		if token.Value == "(" {
			p.tokenIndex++ // Skip "("

			// Check for unary operator immediately after opening parenthesis
			if p.tokenIndex < len(p.tokens) &&
				p.tokens[p.tokenIndex].Type == TOKEN_OPERATOR &&
				(p.tokens[p.tokenIndex].Value == "-" || p.tokens[p.tokenIndex].Value == "+") {

				// Handle unary operation inside parentheses
				unaryToken := p.tokens[p.tokenIndex]
				operator := unaryToken.Value
				line := unaryToken.Line
				p.tokenIndex++ // Skip the operator

				// Parse the operand
				operand, err := p.parseExpression()
				if err != nil {
					return nil, err
				}

				// Create a unary node
				expr := NewUnaryNode(operator, operand, line)

				// Expect closing parenthesis
				if p.tokenIndex >= len(p.tokens) ||
					p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
					p.tokens[p.tokenIndex].Value != ")" {
					return nil, fmt.Errorf("expected closing parenthesis at line %d", token.Line)
				}
				p.tokenIndex++ // Skip ")"

				return expr, nil
			}

			// Regular parenthesized expression
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

// parseMapExpression parses a hash/map literal expression, like {'key': value}
func (p *Parser) parseMapExpression() (Node, error) {
	// Save the line number for error reporting
	line := p.tokens[p.tokenIndex].Line

	// Skip the opening brace
	p.tokenIndex++

	// Parse the map key-value pairs
	items := make(map[Node]Node)

	// Check if there are any items
	if p.tokenIndex < len(p.tokens) &&
		!(p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
			p.tokens[p.tokenIndex].Value == "}") {

		for {
			// Parse key expression
			keyExpr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			// Expect colon separator
			if p.tokenIndex >= len(p.tokens) ||
				p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
				p.tokens[p.tokenIndex].Value != ":" {
				return nil, fmt.Errorf("expected ':' after map key at line %d", line)
			}
			p.tokenIndex++ // Skip colon

			// Parse value expression
			valueExpr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			// Add key-value pair to map
			items[keyExpr] = valueExpr

			// Check for comma separator between items
			if p.tokenIndex < len(p.tokens) &&
				p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
				p.tokens[p.tokenIndex].Value == "," {
				p.tokenIndex++ // Skip comma
				continue
			}

			// No comma, so must be end of map
			break
		}
	}

	// Expect closing brace
	if p.tokenIndex >= len(p.tokens) ||
		p.tokens[p.tokenIndex].Type != TOKEN_PUNCTUATION ||
		p.tokens[p.tokenIndex].Value != "}" {
		return nil, fmt.Errorf("expected closing brace after map items at line %d", line)
	}
	p.tokenIndex++ // Skip closing brace

	// Create hash node
	return &HashNode{
		ExpressionNode: ExpressionNode{
			exprType: ExprHash,
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

// Operator precedence levels (higher number = higher precedence)
const (
	PREC_LOWEST  = 0
	PREC_OR      = 1 // or, ||
	PREC_AND     = 2 // and, &&
	PREC_COMPARE = 3 // ==, !=, <, >, <=, >=, in, not in, matches, starts with, ends with
	PREC_SUM     = 4 // +, -
	PREC_PRODUCT = 5 // *, /, %
	PREC_POWER   = 6 // ^
	PREC_PREFIX  = 7 // not, !, +, - (unary)
)

// Get operator precedence
func getOperatorPrecedence(operator string) int {
	switch operator {
	case "or", "||":
		return PREC_OR
	case "and", "&&":
		return PREC_AND
	case "==", "!=", "<", ">", "<=", ">=", "in", "not in", "matches", "starts with", "ends with", "is", "is not":
		return PREC_COMPARE
	case "+", "-", "~":
		return PREC_SUM
	case "*", "/", "%":
		return PREC_PRODUCT
	case "^":
		return PREC_POWER
	default:
		return PREC_LOWEST
	}
}

// Parse binary expressions (a + b, a and b, a in b, etc.)
func (p *Parser) parseBinaryExpression(left Node) (Node, error) {
	token := p.tokens[p.tokenIndex]
	operator := token.Value
	line := token.Line

	// Special handling for "not defined" pattern
	// This is the common pattern used in Twig: {% if variable not defined %}
	if operator == "not" && p.tokenIndex+1 < len(p.tokens) &&
		p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
		p.tokens[p.tokenIndex+1].Value == "defined" {

		// Next token should be "defined"
		p.tokenIndex += 2 // Skip both "not" and "defined"

		// Create a TestNode with "defined" test
		testNode := &TestNode{
			ExpressionNode: ExpressionNode{
				exprType: ExprTest,
				line:     line,
			},
			node: left,
			test: "defined",
			args: []Node{},
		}

		// Then wrap it in a unary "not" node
		return &UnaryNode{
			ExpressionNode: ExpressionNode{
				exprType: ExprUnary,
				line:     line,
			},
			operator: "not",
			node:     testNode,
		}, nil
	}

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

	// Get precedence of current operator
	precedence := getOperatorPrecedence(operator)

	// Parse the right side expression
	right, err := p.parseSimpleExpression()
	if err != nil {
		return nil, err
	}

	// Create the current binary node
	binaryNode := NewBinaryNode(operator, left, right, line)

	// Check for another binary operator
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

		// Get the next operator and its precedence
		nextOperator := p.tokens[p.tokenIndex].Value
		if p.tokens[p.tokenIndex].Type == TOKEN_NAME {
			// Handle multi-word operators
			if nextOperator == "not" && p.tokenIndex+1 < len(p.tokens) &&
				p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
				p.tokens[p.tokenIndex+1].Value == "in" {
				nextOperator = "not in"
			} else if nextOperator == "is" && p.tokenIndex+1 < len(p.tokens) &&
				p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
				p.tokens[p.tokenIndex+1].Value == "not" {
				nextOperator = "is not"
			} else if nextOperator == "starts" && p.tokenIndex+1 < len(p.tokens) &&
				p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
				p.tokens[p.tokenIndex+1].Value == "with" {
				nextOperator = "starts with"
			} else if nextOperator == "ends" && p.tokenIndex+1 < len(p.tokens) &&
				p.tokens[p.tokenIndex+1].Type == TOKEN_NAME &&
				p.tokens[p.tokenIndex+1].Value == "with" {
				nextOperator = "ends with"
			}
		}

		nextPrecedence := getOperatorPrecedence(nextOperator)

		// If the next operator has higher precedence, we need to parse it first
		if nextPrecedence > precedence {
			// Replace the right side with a binary expression
			newRight, err := p.parseBinaryExpression(right)
			if err != nil {
				return nil, err
			}

			// Update the binary node with the new right side
			binaryNode = NewBinaryNode(operator, left, newRight, line)
		}
	}

	// Check for ternary operator after parsing the binary expression
	if p.tokenIndex < len(p.tokens) &&
		p.tokens[p.tokenIndex].Type == TOKEN_PUNCTUATION &&
		p.tokens[p.tokenIndex].Value == "?" {
		// This is a conditional expression, use the binary node as the condition
		return p.parseConditionalExpression(binaryNode)
	}

	return binaryNode, nil
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

	// Parse the if body (statements between if and endif/else/elseif)
	ifBody, err := parser.parseOuterTemplate()
	if err != nil {
		return nil, err
	}

	// Initialize conditions and bodies arrays with the initial if condition and body
	conditions := []Node{condition}
	bodies := [][]Node{ifBody}
	var elseBody []Node

	// Keep track of whether we've seen an else block
	var hasElseBlock bool

	// Process subsequent tags (elseif, else, endif)
	for {
		// We expect a block start token for elseif, else, or endif
		if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START {
			return nil, fmt.Errorf("unexpected end of template, expected endif at line %d", ifLine)
		}
		parser.tokenIndex++

		// We expect a name token (elseif, else, or endif)
		if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
			return nil, fmt.Errorf("expected block name at line %d", parser.tokens[parser.tokenIndex-1].Line)
		}

		// Get the tag name
		blockName := parser.tokens[parser.tokenIndex].Value
		blockLine := parser.tokens[parser.tokenIndex].Line
		parser.tokenIndex++

		// Process based on the tag type
		if blockName == "elseif" {
			// Check if we've already seen an else block - elseif can't come after else
			if hasElseBlock {
				return nil, fmt.Errorf("unexpected elseif after else at line %d", blockLine)
			}

			// Handle elseif condition
			elseifCondition, err := parser.parseExpression()
			if err != nil {
				return nil, err
			}

			// Expect block end token
			if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
				return nil, fmt.Errorf("expected block end after elseif condition at line %d", blockLine)
			}
			parser.tokenIndex++

			// Parse the elseif body
			elseifBody, err := parser.parseOuterTemplate()
			if err != nil {
				return nil, err
			}

			// Add condition and body to our arrays
			conditions = append(conditions, elseifCondition)
			bodies = append(bodies, elseifBody)

			// Continue checking for more elseif/else/endif tags
		} else if blockName == "else" {
			// Check if we've already seen an else block - can't have multiple else blocks
			if hasElseBlock {
				return nil, fmt.Errorf("multiple else blocks found at line %d", blockLine)
			}

			// Mark that we've seen an else block
			hasElseBlock = true

			// Expect block end token
			if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
				return nil, fmt.Errorf("expected block end after else tag at line %d", blockLine)
			}
			parser.tokenIndex++

			// Parse the else body
			elseBody, err = parser.parseOuterTemplate()
			if err != nil {
				return nil, err
			}

			// After else, we need to find endif next (handled in next iteration)
		} else if blockName == "endif" {
			// Expect block end token
			if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
				return nil, fmt.Errorf("expected block end after endif at line %d", blockLine)
			}
			parser.tokenIndex++

			// We found the endif, we're done
			break
		} else {
			return nil, fmt.Errorf("expected elseif, else, or endif, got %s at line %d", blockName, blockLine)
		}
	}

	// Create and return the if node
	ifNode := &IfNode{
		conditions: conditions,
		bodies:     bodies,
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

	// Check for filter operator (|) - needed for cases where filter detection might be missed
	if IsDebugEnabled() {
		LogDebug("For loop sequence expression type: %T", sequence)
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

			// Check for opening brace
			if parser.tokenIndex < len(parser.tokens) &&
				parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION &&
				parser.tokens[parser.tokenIndex].Value == "{" {
				parser.tokenIndex++ // Skip opening brace

				// Parse key-value pairs
				for {
					// If we see a closing brace, we're done
					if parser.tokenIndex < len(parser.tokens) &&
						parser.tokens[parser.tokenIndex].Type == TOKEN_PUNCTUATION &&
						parser.tokens[parser.tokenIndex].Value == "}" {
						parser.tokenIndex++ // Skip closing brace
						break
					}

					// Get the variable name - can be string literal or name token
					var varName string
					if parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_STRING {
						// It's a quoted string key
						varName = parser.tokens[parser.tokenIndex].Value
						parser.tokenIndex++
					} else if parser.tokenIndex < len(parser.tokens) && parser.tokens[parser.tokenIndex].Type == TOKEN_NAME {
						// It's an unquoted key
						varName = parser.tokens[parser.tokenIndex].Value
						parser.tokenIndex++
					} else {
						return nil, fmt.Errorf("expected variable name or string at line %d", includeLine)
					}

					// Expect colon or equals
					if parser.tokenIndex >= len(parser.tokens) ||
						((parser.tokens[parser.tokenIndex].Type != TOKEN_PUNCTUATION &&
							parser.tokens[parser.tokenIndex].Value != ":") &&
							(parser.tokens[parser.tokenIndex].Type != TOKEN_OPERATOR &&
								parser.tokens[parser.tokenIndex].Value != "=")) {
						return nil, fmt.Errorf("expected ':' or '=' after variable name at line %d", includeLine)
					}
					parser.tokenIndex++ // Skip : or =

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
					}

					// If we see whitespace, skip it
					for parser.tokenIndex < len(parser.tokens) &&
						parser.tokens[parser.tokenIndex].Type == TOKEN_TEXT &&
						strings.TrimSpace(parser.tokens[parser.tokenIndex].Value) == "" {
						parser.tokenIndex++
					}
				}
			} else {
				// If there's no opening brace, expect name-value pairs in the old format
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
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
		return nil, fmt.Errorf("expected block end token after include at line %d, found token type %d with value '%s'",
			includeLine,
			parser.tokens[parser.tokenIndex].Type,
			parser.tokens[parser.tokenIndex].Value)
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

func (p *Parser) parseMacro(parser *Parser) (Node, error) {
	// Use debug logging if enabled
	if IsDebugEnabled() && debugger.level >= DebugVerbose {
		tokenIndex := parser.tokenIndex - 2
		LogVerbose("Parsing macro, tokens available:")
		for i := 0; i < 10 && tokenIndex+i < len(parser.tokens); i++ {
			token := parser.tokens[tokenIndex+i]
			LogVerbose("  Token %d: Type=%d, Value=%q, Line=%d", i, token.Type, token.Value, token.Line)
		}
	}

	// Get the line number of the macro token
	macroLine := parser.tokens[parser.tokenIndex-2].Line

	// Get the macro name
	if parser.tokenIndex >= len(parser.tokens) || parser.tokens[parser.tokenIndex].Type != TOKEN_NAME {
		return nil, fmt.Errorf("expected macro name after macro keyword at line %d", macroLine)
	}

	// Special handling for incorrectly tokenized macro declarations
	macroNameRaw := parser.tokens[parser.tokenIndex].Value
	if IsDebugEnabled() && debugger.level >= DebugVerbose {
		LogVerbose("Raw macro name: %s", macroNameRaw)
	}

	// Check if the name contains parentheses (incorrectly tokenized)
	if strings.Contains(macroNameRaw, "(") {
		// Extract the actual name before the parenthesis
		parts := strings.SplitN(macroNameRaw, "(", 2)
		if len(parts) == 2 {
			macroName := parts[0]
			paramStr := "(" + parts[1]
			if IsDebugEnabled() && debugger.level >= DebugVerbose {
				LogVerbose("Fixed macro name: %s", macroName)
				LogVerbose("Parameter string: %s", paramStr)
			}

			// Parse parameters
			var params []string
			defaults := make(map[string]Node)

			// Simple parameter parsing - split by comma
			paramList := strings.TrimRight(paramStr[1:], ")")
			if paramList != "" {
				paramItems := strings.Split(paramList, ",")

				for _, param := range paramItems {
					param = strings.TrimSpace(param)

					// Check for default value
					if strings.Contains(param, "=") {
						parts := strings.SplitN(param, "=", 2)
						paramName := strings.TrimSpace(parts[0])
						defaultValue := strings.TrimSpace(parts[1])

						params = append(params, paramName)

						// Handle quoted strings in default values
						if (strings.HasPrefix(defaultValue, "'") && strings.HasSuffix(defaultValue, "'")) ||
							(strings.HasPrefix(defaultValue, "\"") && strings.HasSuffix(defaultValue, "\"")) {
							// Remove quotes
							strValue := defaultValue[1 : len(defaultValue)-1]
							defaults[paramName] = NewLiteralNode(strValue, macroLine)
						} else if defaultValue == "true" {
							defaults[paramName] = NewLiteralNode(true, macroLine)
						} else if defaultValue == "false" {
							defaults[paramName] = NewLiteralNode(false, macroLine)
						} else if i, err := strconv.Atoi(defaultValue); err == nil {
							defaults[paramName] = NewLiteralNode(i, macroLine)
						} else {
							// Fallback to string
							defaults[paramName] = NewLiteralNode(defaultValue, macroLine)
						}
					} else {
						params = append(params, param)
					}
				}
			}

			// Skip to the end of the token
			parser.tokenIndex++

			// Expect block end
			if parser.tokenIndex >= len(parser.tokens) ||
				(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
					parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
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
				(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START &&
					parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START_TRIM) ||
				parser.tokens[parser.tokenIndex+1].Type != TOKEN_NAME ||
				parser.tokens[parser.tokenIndex+1].Value != "endmacro" {
				return nil, fmt.Errorf("missing endmacro tag for macro '%s' at line %d",
					macroName, macroLine)
			}

			// Skip {% endmacro %}
			parser.tokenIndex += 2 // Skip {% endmacro

			// Expect block end
			if parser.tokenIndex >= len(parser.tokens) ||
				(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
					parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
				return nil, fmt.Errorf("expected block end token after endmacro at line %d", parser.tokens[parser.tokenIndex].Line)
			}
			parser.tokenIndex++

			// Create the macro node
			if IsDebugEnabled() && debugger.level >= DebugVerbose {
				LogVerbose("Creating MacroNode with %d parameters and %d defaults", len(params), len(defaults))
			}
			return NewMacroNode(macroName, params, defaults, bodyNodes, macroLine), nil
		}
	}

	// Regular parsing path
	macroName := parser.tokens[parser.tokenIndex].Value
	if IsDebugEnabled() && debugger.level >= DebugVerbose {
		LogVerbose("Macro name: %s", macroName)
	}
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
			fmt.Println("DEBUG: Parameter name:", paramName)
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
					fmt.Println("DEBUG: Error parsing default value:", err)
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
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
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
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_START_TRIM) ||
		parser.tokens[parser.tokenIndex+1].Type != TOKEN_NAME ||
		parser.tokens[parser.tokenIndex+1].Value != "endmacro" {
		return nil, fmt.Errorf("missing endmacro tag for macro '%s' at line %d",
			macroName, macroLine)
	}

	// Skip {% endmacro %}
	parser.tokenIndex += 2 // Skip {% endmacro

	// Expect block end
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
		return nil, fmt.Errorf("expected block end token after endmacro at line %d", parser.tokens[parser.tokenIndex].Line)
	}
	parser.tokenIndex++

	// Create the macro node
	if IsDebugEnabled() && debugger.level >= DebugVerbose {
		LogVerbose("Creating MacroNode with %d parameters and %d defaults", len(params), len(defaults))
	}
	return NewMacroNode(macroName, params, defaults, bodyNodes, macroLine), nil
}

func (p *Parser) parseImport(parser *Parser) (Node, error) {
	// Use debug logging if enabled
	if IsDebugEnabled() && debugger.level >= DebugVerbose {
		tokenIndex := parser.tokenIndex - 2
		LogVerbose("Parsing import, tokens available:")
		for i := 0; i < 10 && tokenIndex+i < len(parser.tokens); i++ {
			token := parser.tokens[tokenIndex+i]
			LogVerbose("  Token %d: Type=%d, Value=%q, Line=%d", i, token.Type, token.Value, token.Line)
		}
	}

	// Get the line number of the import token
	importLine := parser.tokens[parser.tokenIndex-2].Line

	// Check for incorrectly tokenized import syntax
	if parser.tokenIndex < len(parser.tokens) &&
		parser.tokens[parser.tokenIndex].Type == TOKEN_NAME &&
		strings.Contains(parser.tokens[parser.tokenIndex].Value, " as ") {

		// Special handling for combined syntax like "path.twig as alias"
		parts := strings.SplitN(parser.tokens[parser.tokenIndex].Value, " as ", 2)
		if len(parts) == 2 {
			templatePath := strings.TrimSpace(parts[0])
			alias := strings.TrimSpace(parts[1])

			if IsDebugEnabled() && debugger.level >= DebugVerbose {
				LogVerbose("Found combined import syntax: template=%q, alias=%q", templatePath, alias)
			}

			// Create an expression node for the template path
			var templateExpr Node
			if strings.HasPrefix(templatePath, "\"") && strings.HasSuffix(templatePath, "\"") {
				// It's already a quoted string
				templateExpr = NewLiteralNode(templatePath[1:len(templatePath)-1], importLine)
			} else {
				// Create a string literal node
				templateExpr = NewLiteralNode(templatePath, importLine)
			}

			// Skip to end of token
			parser.tokenIndex++

			// Expect block end
			if parser.tokenIndex >= len(parser.tokens) ||
				(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
					parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
				return nil, fmt.Errorf("expected block end token after import statement at line %d", importLine)
			}
			parser.tokenIndex++

			// Create import node
			return NewImportNode(templateExpr, alias, importLine), nil
		}
	}

	// Standard parsing path
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
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
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
	if parser.tokenIndex >= len(parser.tokens) ||
		(parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END &&
			parser.tokens[parser.tokenIndex].Type != TOKEN_BLOCK_END_TRIM) {
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

// HtmlPreservingTokenize is an exported version of htmlPreservingTokenize for testing
func (p *Parser) HtmlPreservingTokenize() ([]Token, error) {
	return p.htmlPreservingTokenize()
}

// SetSource sets the source for parsing - used for testing
func (p *Parser) SetSource(source string) {
	p.source = source
}

// parseVerbatim parses a verbatim tag and its content
func (p *Parser) parseVerbatim(parser *Parser) (Node, error) {
	// Get the line number of the verbatim token
	verbatimLine := parser.tokens[parser.tokenIndex-2].Line

	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
		return nil, fmt.Errorf("expected block end after verbatim tag at line %d", verbatimLine)
	}
	parser.tokenIndex++

	// Collect all content until we find the endverbatim tag
	var contentBuilder strings.Builder

	for parser.tokenIndex < len(parser.tokens) {
		token := parser.tokens[parser.tokenIndex]

		// Look for the endverbatim tag
		if token.Type == TOKEN_BLOCK_START || token.Type == TOKEN_BLOCK_START_TRIM {
			// Check if this is the endverbatim tag
			if parser.tokenIndex+1 < len(parser.tokens) &&
				parser.tokens[parser.tokenIndex+1].Type == TOKEN_NAME &&
				parser.tokens[parser.tokenIndex+1].Value == "endverbatim" {

				// Skip the block start and endverbatim name
				parser.tokenIndex += 2 // Now at the endverbatim token

				// Expect the block end token
				if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
					return nil, fmt.Errorf("expected block end after endverbatim at line %d", token.Line)
				}
				parser.tokenIndex++ // Skip the block end token

				// Create a verbatim node with the collected content
				return NewVerbatimNode(contentBuilder.String(), verbatimLine), nil
			}
		}

		// Add this token's content to our verbatim content
		if token.Type == TOKEN_TEXT {
			contentBuilder.WriteString(token.Value)
		} else if token.Type == TOKEN_VAR_START || token.Type == TOKEN_VAR_START_TRIM {
			// For variable tags, preserve them as literal text
			contentBuilder.WriteString("{{")

			// Skip variable start token and process until variable end
			parser.tokenIndex++

			// Process tokens until variable end
			for parser.tokenIndex < len(parser.tokens) {
				innerToken := parser.tokens[parser.tokenIndex]

				if innerToken.Type == TOKEN_VAR_END || innerToken.Type == TOKEN_VAR_END_TRIM {
					contentBuilder.WriteString("}}")
					break
				} else if innerToken.Type == TOKEN_NAME || innerToken.Type == TOKEN_STRING ||
					innerToken.Type == TOKEN_NUMBER || innerToken.Type == TOKEN_OPERATOR ||
					innerToken.Type == TOKEN_PUNCTUATION {
					contentBuilder.WriteString(innerToken.Value)
				}

				parser.tokenIndex++
			}
		} else if token.Type == TOKEN_BLOCK_START || token.Type == TOKEN_BLOCK_START_TRIM {
			// For block tags, preserve them as literal text
			contentBuilder.WriteString("{%")

			// Skip block start token and process until block end
			parser.tokenIndex++

			// Process tokens until block end
			for parser.tokenIndex < len(parser.tokens) {
				innerToken := parser.tokens[parser.tokenIndex]

				if innerToken.Type == TOKEN_BLOCK_END || innerToken.Type == TOKEN_BLOCK_END_TRIM {
					contentBuilder.WriteString("%}")
					break
				} else if innerToken.Type == TOKEN_NAME || innerToken.Type == TOKEN_STRING ||
					innerToken.Type == TOKEN_NUMBER || innerToken.Type == TOKEN_OPERATOR ||
					innerToken.Type == TOKEN_PUNCTUATION {
					// If this is the first TOKEN_NAME in a block, add a space after it
					if innerToken.Type == TOKEN_NAME && parser.tokenIndex > 0 &&
						(parser.tokens[parser.tokenIndex-1].Type == TOKEN_BLOCK_START ||
							parser.tokens[parser.tokenIndex-1].Type == TOKEN_BLOCK_START_TRIM) {
						contentBuilder.WriteString(innerToken.Value + " ")
					} else {
						contentBuilder.WriteString(innerToken.Value)
					}
				}

				parser.tokenIndex++
			}
		} else if token.Type == TOKEN_COMMENT_START {
			// For comment tags, preserve them as literal text
			contentBuilder.WriteString("{#")

			// Skip comment start token and process until comment end
			parser.tokenIndex++

			// Process tokens until comment end
			for parser.tokenIndex < len(parser.tokens) {
				innerToken := parser.tokens[parser.tokenIndex]

				if innerToken.Type == TOKEN_COMMENT_END {
					contentBuilder.WriteString("#}")
					break
				} else if innerToken.Type == TOKEN_TEXT {
					contentBuilder.WriteString(innerToken.Value)
				}

				parser.tokenIndex++
			}
		}

		parser.tokenIndex++

		// Check for end of tokens
		if parser.tokenIndex >= len(parser.tokens) {
			return nil, fmt.Errorf("unexpected end of template, unclosed verbatim tag at line %d", verbatimLine)
		}
	}

	// If we get here, we never found the endverbatim tag
	return nil, fmt.Errorf("unclosed verbatim tag at line %d", verbatimLine)
}
