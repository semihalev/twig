// Code generated by parsegen; DO NOT EDIT.
package twig

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser is responsible for parsing tokens into an abstract syntax tree
type Parser struct {
	tokens     []Token
	tokenIndex int
	blockHandlers map[string]BlockHandlerFunc
}

// BlockHandlerFunc is a function that parses a specific block tag
type BlockHandlerFunc func(*Parser) (Node, error)

// NewParser creates a new parser for the given tokens
func NewParser() *Parser {
	p := &Parser{
		tokens:     nil,
		tokenIndex: 0,
		blockHandlers: make(map[string]BlockHandlerFunc),
	}
	p.initBlockHandlers()
	return p
}

// Parse parses a template source into a node tree
func (p *Parser) Parse(source string) (Node, error) {
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexer error: %w", err)
	}
	
	p.tokens = tokens
	p.tokenIndex = 0
	
	return p.parseOuterTemplate()
}

// initBlockHandlers initializes the block handlers map
func (p *Parser) initBlockHandlers() {
	p.blockHandlers = map[string]BlockHandlerFunc{
		"if":      p.parseIf,
		"for":     p.parseFor,
		"set":     p.parseSet,
		"block":   p.parseBlock,
		"extends": p.parseExtends,
		"include": p.parseInclude,
		"do":      p.parseDo,
		"macro":   p.parseMacro,
		"import":  p.parseImport,
		"from":    p.parseFrom,
		
		// Special closing tags - they will be handled in their corresponding open tag parsers
		"endif":   p.parseEndTag,
		"endfor":  p.parseEndTag,
		"endblock": p.parseEndTag,
		"else":    p.parseEndTag,
		"elseif":  p.parseEndTag,
	}
}

// parseOuterTemplate parses the outer template structure
func (p *Parser) parseOuterTemplate() (Node, error) {
	nodes := []Node{}
	
	for p.tokenIndex < len(p.tokens) {
		token := p.tokens[p.tokenIndex]
		
		switch token.Type {
		case T_EOF:
			// End of file, exit loop
			p.tokenIndex++
			break
			
		case T_TEXT:
			// Text node
			nodes = append(nodes, NewTextNode(token.Value, token.Line))
			p.tokenIndex++
			
		case T_VAR_START:
			// Variable output
			p.tokenIndex++ // Skip {{
			
			// Parse expression
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			
			// Expect }}
			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_VAR_END {
				return nil, fmt.Errorf("unexpected token, expected }} at line %d", 
					token.Line)
			}
			
			// Create print node
			nodes = append(nodes, NewPrintNode(expr, token.Line))
			p.tokenIndex++ // Skip }}
			
		case T_BLOCK_START:
			// Block tag
			p.tokenIndex++ // Skip {%
			
			// Get block name
			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_IDENT {
				return nil, fmt.Errorf("unexpected token after {%%, expected block name at line %d", 
					token.Line)
			}
			
			blockName := p.tokens[p.tokenIndex].Value
			p.tokenIndex++
			
			// Check if this is a control ending tag (endif, endfor, endblock, etc.)
			if blockName == "endif" || blockName == "endfor" || blockName == "endblock" || 
			   blockName == "else" || blockName == "elseif" {
				// We should return to the parent parser that's handling the parent block
				// First move back two steps to the start of the block tag
				p.tokenIndex -= 2
				return NewRootNode(nodes, 1), nil
			}
			
			// Check if we have a handler for this block type
			handler, ok := p.blockHandlers[blockName]
			if !ok {
				return nil, fmt.Errorf("unknown block tag '%s' at line %d", 
					blockName, token.Line)
			}
			
			// Call the handler
			node, err := handler(p)
			if err != nil {
				return nil, err
			}
			
			// Add the result to the nodes list
			if node != nil {
				nodes = append(nodes, node)
			}
			
		default:
			// Unexpected token
			return nil, fmt.Errorf("unexpected token '%s' at line %d", 
				token.Value, token.Line)
		}
	}
	
	return NewRootNode(nodes, 1), nil
}

// parseEndTag parses an end tag (endif, endfor, etc.)
func (p *Parser) parseEndTag(parser *Parser) (Node, error) {
	// End tags should be handled by the corresponding start tag parser
	return nil, fmt.Errorf("unexpected end tag at line %d", p.tokens[p.tokenIndex-1].Line)
}

// parseExpression parses an expression
func (p *Parser) parseExpression() (Node, error) {
	// For now, implement a simple expression parser
	// This will be expanded later to handle complex expressions
	
	if p.tokenIndex >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input while parsing expression")
	}
	
	token := p.tokens[p.tokenIndex]
	
	switch token.Type {
	case T_IDENT:
		// Variable reference
		p.tokenIndex++
		return NewVariableNode(token.Value, token.Line), nil
		
	case T_STRING:
		// String literal
		p.tokenIndex++
		return NewLiteralNode(token.Value, token.Line), nil
		
	case T_NUMBER:
		// Number literal
		p.tokenIndex++
		
		// Parse number
		if strings.Contains(token.Value, ".") {
			// Float
			val, _ := strconv.ParseFloat(token.Value, 64)
			return NewLiteralNode(val, token.Line), nil
		} else {
			// Integer
			val, _ := strconv.ParseInt(token.Value, 10, 64)
			return NewLiteralNode(int(val), token.Line), nil
		}
		
	default:
		return nil, fmt.Errorf("unexpected token '%s' in expression at line %d", 
			token.Value, token.Line)
	}
}

// parseIf parses an if statement
func (p *Parser) parseIf(*Parser) (Node, error) {
	// Get condition
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	// Expect %}
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_BLOCK_END {
		return nil, fmt.Errorf("unexpected token, expected %%} at line %d", 
			p.tokens[p.tokenIndex-1].Line)
	}
	p.tokenIndex++ // Skip %}
	
	// Parse if body
	ifBody, err := p.parseOuterTemplate()
	if err != nil {
		return nil, err
	}
	
	// Get body nodes
	var bodyNodes []Node
	if rootNode, ok := ifBody.(*RootNode); ok {
		bodyNodes = rootNode.Children()
	} else {
		bodyNodes = []Node{ifBody}
	}
	
	// Check for else
	var elseNodes []Node
	
	// Check if we're at an else tag
	if p.tokenIndex+1 < len(p.tokens) && 
	   p.tokens[p.tokenIndex].Type == T_BLOCK_START && 
	   p.tokens[p.tokenIndex+1].Type == T_IDENT &&
	   p.tokens[p.tokenIndex+1].Value == "else" {
		
		// Skip {% else %}
		p.tokenIndex += 2 // Skip {% else
		
		// Expect %}
		if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_BLOCK_END {
			return nil, fmt.Errorf("unexpected token, expected %%} after else at line %d", 
				p.tokens[p.tokenIndex-1].Line)
		}
		p.tokenIndex++ // Skip %}
		
		// Parse else body
		elseBody, err := p.parseOuterTemplate()
		if err != nil {
			return nil, err
		}
		
		// Get else nodes
		if rootNode, ok := elseBody.(*RootNode); ok {
			elseNodes = rootNode.Children()
		} else {
			elseNodes = []Node{elseBody}
		}
	}
	
	// Check for endif
	if p.tokenIndex+1 >= len(p.tokens) || 
	   p.tokens[p.tokenIndex].Type != T_BLOCK_START || 
	   p.tokens[p.tokenIndex+1].Type != T_IDENT ||
	   p.tokens[p.tokenIndex+1].Value != "endif" {
		return nil, fmt.Errorf("missing endif tag at line %d", 
			p.tokens[p.tokenIndex].Line)
	}
	
	// Skip {% endif %}
	p.tokenIndex += 2 // Skip {% endif
	
	// Expect %}
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_BLOCK_END {
		return nil, fmt.Errorf("unexpected token, expected %%} after endif at line %d", 
			p.tokens[p.tokenIndex-1].Line)
	}
	p.tokenIndex++ // Skip %}
	
	// Create if node
	return NewIfNode(condition, bodyNodes, elseNodes, p.tokens[p.tokenIndex-1].Line), nil
}

// parseFor parses a for loop
func (p *Parser) parseFor(*Parser) (Node, error) {
	// TODO: Implement for loop parsing
	return nil, fmt.Errorf("for loop not implemented yet")
}

// parseSet parses a set statement
func (p *Parser) parseSet(*Parser) (Node, error) {
	// Get the line number of the set token
	setLine := p.tokens[p.tokenIndex-2].Line
	
	// Get the variable name
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_IDENT {
		return nil, fmt.Errorf("expected variable name after set at line %d", setLine)
	}
	
	varName := p.tokens[p.tokenIndex].Value
	p.tokenIndex++
	
	// Expect '='
	if p.tokenIndex >= len(p.tokens) || 
	   p.tokens[p.tokenIndex].Type != T_OPERATOR || 
	   p.tokens[p.tokenIndex].Value != "=" {
		return nil, fmt.Errorf("expected '=' after variable name at line %d", setLine)
	}
	p.tokenIndex++
	
	// Parse the value expression
	valueExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	// For expressions like 5 + 10, we need to parse both sides and make a binary node
	// Check if there's an operator after the first token
	if p.tokenIndex < len(p.tokens) && 
	   p.tokens[p.tokenIndex].Type == T_OPERATOR && 
	   p.tokens[p.tokenIndex].Value != "=" {
		
		// Get the operator
		operator := p.tokens[p.tokenIndex].Value
		p.tokenIndex++
		
		// Parse the right side
		rightExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		
		// Create a binary node
		valueExpr = NewBinaryNode(operator, valueExpr, rightExpr, setLine)
	}
	
	// Expect the block end token
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after set expression at line %d", setLine)
	}
	p.tokenIndex++
	
	// Create the set node
	setNode := NewSetNode(varName, valueExpr, setLine)
	
	return setNode, nil
}

// parseBlock parses a block definition
func (p *Parser) parseBlock(*Parser) (Node, error) {
	// TODO: Implement block parsing
	return nil, fmt.Errorf("block tag not implemented yet")
}

// parseExtends parses an extends statement
func (p *Parser) parseExtends(*Parser) (Node, error) {
	// TODO: Implement extends parsing
	return nil, fmt.Errorf("extends tag not implemented yet")
}

// parseInclude parses an include statement
func (p *Parser) parseInclude(*Parser) (Node, error) {
	// TODO: Implement include parsing
	return nil, fmt.Errorf("include tag not implemented yet")
}

// parseDo parses a do statement
func (p *Parser) parseDo(*Parser) (Node, error) {
	// TODO: Implement do parsing
	return nil, fmt.Errorf("do tag not implemented yet")
}

// parseMacro parses a macro definition
func (p *Parser) parseMacro(*Parser) (Node, error) {
	// Get the line number of the macro token
	macroLine := p.tokens[p.tokenIndex-2].Line
	
	// Get the macro name
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_IDENT {
		return nil, fmt.Errorf("expected macro name after macro keyword at line %d", macroLine)
	}
	
	macroName := p.tokens[p.tokenIndex].Value
	p.tokenIndex++
	
	// Expect opening parenthesis for parameters
	if p.tokenIndex >= len(p.tokens) || 
	   p.tokens[p.tokenIndex].Type != T_PUNCTUATION || 
	   p.tokens[p.tokenIndex].Value != "(" {
		return nil, fmt.Errorf("expected '(' after macro name at line %d", macroLine)
	}
	p.tokenIndex++
	
	// Parse parameters
	var params []string
	defaults := make(map[string]Node)
	
	// If we don't have a closing parenthesis immediately, we have parameters
	if p.tokenIndex < len(p.tokens) && 
	   (p.tokens[p.tokenIndex].Type != T_PUNCTUATION || 
	    p.tokens[p.tokenIndex].Value != ")") {
		
		for {
			// Get parameter name
			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_IDENT {
				return nil, fmt.Errorf("expected parameter name at line %d", macroLine)
			}
			
			paramName := p.tokens[p.tokenIndex].Value
			params = append(params, paramName)
			p.tokenIndex++
			
			// Check for default value
			if p.tokenIndex < len(p.tokens) && 
			   p.tokens[p.tokenIndex].Type == T_OPERATOR && 
			   p.tokens[p.tokenIndex].Value == "=" {
				p.tokenIndex++ // Skip =
				
				// Parse default value expression
				defaultExpr, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				
				defaults[paramName] = defaultExpr
			}
			
			// Check if we have more parameters
			if p.tokenIndex < len(p.tokens) && 
			   p.tokens[p.tokenIndex].Type == T_PUNCTUATION && 
			   p.tokens[p.tokenIndex].Value == "," {
				p.tokenIndex++ // Skip comma
				continue
			}
			
			break
		}
	}
	
	// Expect closing parenthesis
	if p.tokenIndex >= len(p.tokens) || 
	   p.tokens[p.tokenIndex].Type != T_PUNCTUATION || 
	   p.tokens[p.tokenIndex].Value != ")" {
		return nil, fmt.Errorf("expected ')' after macro parameters at line %d", macroLine)
	}
	p.tokenIndex++
	
	// Expect %}
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after macro declaration at line %d", macroLine)
	}
	p.tokenIndex++
	
	// Parse the macro body
	bodyNode, err := p.parseOuterTemplate()
	if err != nil {
		return nil, err
	}
	
	// Extract body nodes
	var bodyNodes []Node
	if rootNode, ok := bodyNode.(*RootNode); ok {
		bodyNodes = rootNode.Children()
	} else {
		bodyNodes = []Node{bodyNode}
	}
	
	// Expect endmacro tag
	if p.tokenIndex+1 >= len(p.tokens) || 
	   p.tokens[p.tokenIndex].Type != T_BLOCK_START || 
	   p.tokens[p.tokenIndex+1].Type != T_IDENT || 
	   p.tokens[p.tokenIndex+1].Value != "endmacro" {
		return nil, fmt.Errorf("missing endmacro tag for macro '%s' at line %d", 
			macroName, macroLine)
	}
	
	// Skip {% endmacro %}
	p.tokenIndex += 2 // Skip {% endmacro
	
	// Expect %}
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after endmacro at line %d", p.tokens[p.tokenIndex].Line)
	}
	p.tokenIndex++
	
	// Create the macro node
	return NewMacroNode(macroName, params, defaults, bodyNodes, macroLine), nil
}

// parseImport parses an import statement
func (p *Parser) parseImport(*Parser) (Node, error) {
	// Get the line number of the import token
	importLine := p.tokens[p.tokenIndex-2].Line
	
	// Get the template to import
	templateExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	// Expect 'as' keyword
	if p.tokenIndex >= len(p.tokens) || 
	   p.tokens[p.tokenIndex].Type != T_IDENT || 
	   p.tokens[p.tokenIndex].Value != "as" {
		return nil, fmt.Errorf("expected 'as' after template path at line %d", importLine)
	}
	p.tokenIndex++
	
	// Get the alias name
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_IDENT {
		return nil, fmt.Errorf("expected identifier after 'as' at line %d", importLine)
	}
	
	alias := p.tokens[p.tokenIndex].Value
	p.tokenIndex++
	
	// Expect %}
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after import statement at line %d", importLine)
	}
	p.tokenIndex++
	
	// Create import node
	return NewImportNode(templateExpr, alias, importLine), nil
}

// parseFrom parses a from statement
func (p *Parser) parseFrom(*Parser) (Node, error) {
	// Get the line number of the from token
	fromLine := p.tokens[p.tokenIndex-2].Line
	
	// Get the template to import from
	templateExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	// Expect 'import' keyword
	if p.tokenIndex >= len(p.tokens) || 
	   p.tokens[p.tokenIndex].Type != T_IDENT || 
	   p.tokens[p.tokenIndex].Value != "import" {
		return nil, fmt.Errorf("expected 'import' after template path at line %d", fromLine)
	}
	p.tokenIndex++
	
	// Parse the imported items
	var macros []string
	aliases := make(map[string]string)
	
	// We need at least one macro to import
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_IDENT {
		return nil, fmt.Errorf("expected at least one identifier after 'import' at line %d", fromLine)
	}
	
	for p.tokenIndex < len(p.tokens) && p.tokens[p.tokenIndex].Type == T_IDENT {
		// Get macro name
		macroName := p.tokens[p.tokenIndex].Value
		p.tokenIndex++
		
		// Check for 'as' keyword for aliasing
		if p.tokenIndex < len(p.tokens) && 
		   p.tokens[p.tokenIndex].Type == T_IDENT && 
		   p.tokens[p.tokenIndex].Value == "as" {
			p.tokenIndex++ // Skip 'as'
			
			// Get alias name
			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_IDENT {
				return nil, fmt.Errorf("expected identifier after 'as' at line %d", fromLine)
			}
			
			aliasName := p.tokens[p.tokenIndex].Value
			aliases[macroName] = aliasName
			p.tokenIndex++
		} else {
			// No alias, just add to macros list
			macros = append(macros, macroName)
		}
		
		// Check for comma to separate items
		if p.tokenIndex < len(p.tokens) && 
		   p.tokens[p.tokenIndex].Type == T_PUNCTUATION && 
		   p.tokens[p.tokenIndex].Value == "," {
			p.tokenIndex++ // Skip comma
			
			// Expect another identifier after comma
			if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_IDENT {
				return nil, fmt.Errorf("expected identifier after ',' at line %d", fromLine)
			}
		} else {
			// End of imports
			break
		}
	}
	
	// Expect %}
	if p.tokenIndex >= len(p.tokens) || p.tokens[p.tokenIndex].Type != T_BLOCK_END {
		return nil, fmt.Errorf("expected block end token after from import statement at line %d", fromLine)
	}
	p.tokenIndex++
	
	// Create from import node
	return NewFromImportNode(templateExpr, macros, aliases, fromLine), nil
}