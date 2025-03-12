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

	// Use the optimized tokenizer for maximum performance and minimal allocations
	// This will treat everything outside twig tags as TEXT tokens
	var err error

	// Use zero allocation tokenizer for optimal performance
	tokenizer := GetTokenizer(p.source, 0)

	// Use optimized version for larger templates
	if len(p.source) > 4096 {
		// Use the optimized tag detection for large templates
		p.tokens, err = tokenizer.TokenizeOptimized()
	} else {
		// Use regular tokenization for smaller templates
		p.tokens, err = tokenizer.TokenizeHtmlPreserving()
	}

	// Apply whitespace control to handle whitespace trimming directives
	if err == nil {
		tokenizer.ApplyWhitespaceControl()
	}

	// Return the tokenizer to the pool
	ReleaseTokenizer(tokenizer)

	if err != nil {
		return nil, fmt.Errorf("tokenization error: %w", err)
	}

	// Template tokenization complete
	// Whitespace control has already been applied by the tokenizer

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
		"apply":     p.parseApply,

		// Special closing tags - they will be handled in their corresponding open tag parsers
		"endif":        p.parseEndTag,
		"endfor":       p.parseEndTag,
		"endmacro":     p.parseEndTag,
		"endblock":     p.parseEndTag,
		"endspaceless": p.parseEndTag,
		"endapply":     p.parseEndTag,

		"else":        p.parseEndTag,
		"elseif":      p.parseEndTag,
		"endverbatim": p.parseEndTag,
	}
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
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
				blockName == "endspaceless" || blockName == "endapply" || blockName == "endverbatim" {
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

// HtmlPreservingTokenize is for testing - uses the optimized tokenizer
func (p *Parser) HtmlPreservingTokenize() ([]Token, error) {
	tokenizer := GetTokenizer(p.source, 0)
	defer ReleaseTokenizer(tokenizer)
	return tokenizer.TokenizeHtmlPreserving()
}

// SetSource sets the source for parsing - used for testing
func (p *Parser) SetSource(source string) {
	p.source = source
}
