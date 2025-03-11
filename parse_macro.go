package twig

import (
	"fmt"
	"strconv"
	"strings"
)

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
