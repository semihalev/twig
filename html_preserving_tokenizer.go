package twig

import (
	"fmt"
	"strconv"
	"strings"
)

// createToken is a helper to create a token from the pool
func createToken(tokenType int, value string, line int) Token {
	// We return by value to avoid heap allocations for small/temporary tokens
	// For tokens that get stored in slices, we'd use the pointer version
	token := TokenPool.Get().(*Token)
	token.Type = tokenType
	token.Value = value
	token.Line = line
	
	// Create a copy to return by value
	result := *token
	
	// Return the original to the pool
	TokenPool.Put(token)
	
	return result
}

// htmlPreservingTokenize is a replacement for the standard tokenize function
// that properly preserves HTML content while correctly handling Twig syntax.
// It treats everything outside of Twig tags as raw TEXT tokens.
func (p *Parser) htmlPreservingTokenize() ([]Token, error) {
	// Pre-allocate tokens with estimated capacity based on source length
	// This avoids frequent slice reallocations
	estimatedTokenCount := len(p.source) / 20 // Rough estimate: one token per 20 chars
	tokens := GetTokenSlice(estimatedTokenCount)
	var currentPosition int
	line := 1

	for currentPosition < len(p.source) {
		// Find the next twig tag start
		nextTagPos := -1
		tagType := -1
		var matchedPos struct {
			pos     int
			pattern string
			ttype   int
			length  int
		}

		// Use a single substring for all pattern searches to reduce allocations
		remainingSource := p.source[currentPosition:]

		// Check for all possible tag starts, including whitespace control variants
		positions := []struct {
			pos     int
			pattern string
			ttype   int
			length  int
		}{
			{strings.Index(remainingSource, "{{-"), "{{-", TOKEN_VAR_START_TRIM, 3},
			{strings.Index(remainingSource, "{{"), "{{", TOKEN_VAR_START, 2},
			{strings.Index(remainingSource, "{%-"), "{%-", TOKEN_BLOCK_START_TRIM, 3},
			{strings.Index(remainingSource, "{%"), "{%", TOKEN_BLOCK_START, 2},
			{strings.Index(remainingSource, "{#"), "{#", TOKEN_COMMENT_START, 2},
		}

		// Find the closest tag
		for _, pos := range positions {
			if pos.pos != -1 {
				adjustedPos := currentPosition + pos.pos
				if nextTagPos == -1 || adjustedPos < nextTagPos {
					nextTagPos = adjustedPos
					tagType = pos.ttype
					matchedPos = pos
				}
			}
		}

		// Check if the tag is escaped with a backslash
		if nextTagPos != -1 && nextTagPos > 0 && p.source[nextTagPos-1] == '\\' {
			// This tag is escaped with a backslash, treat it as literal text
			// Add text up to the backslash (if any)
			if nextTagPos-1 > currentPosition {
				preText := p.source[currentPosition : nextTagPos-1]
				tokens = append(tokens, createToken(TOKEN_TEXT, preText, line))
				line += strings.Count(preText, "\n")
			}

			// Add the tag itself as literal text (without the backslash)
			tokens = append(tokens, createToken(TOKEN_TEXT, matchedPos.pattern, line))

			// Move past the tag
			currentPosition = nextTagPos + matchedPos.length
			continue
		}

		if nextTagPos == -1 {
			// No more tags found, add the rest as TEXT
			content := p.source[currentPosition:]
			if len(content) > 0 {
				line += strings.Count(content, "\n")
				tokens = append(tokens, createToken(TOKEN_TEXT, content, line))
			}
			break
		}

		// Add the text before the tag (HTML content)
		if nextTagPos > currentPosition {
			content := p.source[currentPosition:nextTagPos]
			line += strings.Count(content, "\n")
			tokens = append(tokens, createToken(TOKEN_TEXT, content, line))
		}

		// Add the tag start token
		tokens = append(tokens, createToken(tagType, "", line))

		// Determine tag length and move past the opening
		tagLength := 2 // Default for "{{", "{%", "{#"
		if tagType == TOKEN_VAR_START_TRIM || tagType == TOKEN_BLOCK_START_TRIM {
			tagLength = 3 // For "{{-" or "{%-"
		}
		currentPosition = nextTagPos + tagLength

		// Find the matching end tag
		var endTag string
		var endTagType int
		var endTagLength int

		if tagType == TOKEN_VAR_START || tagType == TOKEN_VAR_START_TRIM {
			// For variable tags, look for "}}" or "-}}"
			endPos1 := strings.Index(p.source[currentPosition:], "}}")
			endPos2 := strings.Index(p.source[currentPosition:], "-}}")

			if endPos1 != -1 && (endPos2 == -1 || endPos1 < endPos2) {
				endTag = "}}"
				endTagType = TOKEN_VAR_END
				endTagLength = 2
			} else if endPos2 != -1 {
				endTag = "-}}"
				endTagType = TOKEN_VAR_END_TRIM
				endTagLength = 3
			} else {
				return nil, fmt.Errorf("unclosed variable tag starting at line %d", line)
			}
		} else if tagType == TOKEN_BLOCK_START || tagType == TOKEN_BLOCK_START_TRIM {
			// For block tags, look for "%}" or "-%}"
			endPos1 := strings.Index(p.source[currentPosition:], "%}")
			endPos2 := strings.Index(p.source[currentPosition:], "-%}")

			if endPos1 != -1 && (endPos2 == -1 || endPos1 < endPos2) {
				endTag = "%}"
				endTagType = TOKEN_BLOCK_END
				endTagLength = 2
			} else if endPos2 != -1 {
				endTag = "-%}"
				endTagType = TOKEN_BLOCK_END_TRIM
				endTagLength = 3
			} else {
				return nil, fmt.Errorf("unclosed block tag starting at line %d", line)
			}
		} else if tagType == TOKEN_COMMENT_START {
			// For comment tags, look for "#}"
			endPos := strings.Index(p.source[currentPosition:], "#}")
			if endPos == -1 {
				return nil, fmt.Errorf("unclosed comment starting at line %d", line)
			}
			endTag = "#}"
			endTagType = TOKEN_COMMENT_END
			endTagLength = 2
		}

		// Find the position of the end tag
		endPos := strings.Index(p.source[currentPosition:], endTag)
		if endPos == -1 {
			return nil, fmt.Errorf("unclosed tag starting at line %d", line)
		}

		// Get the content between the tags
		tagContent := p.source[currentPosition : currentPosition+endPos]
		line += strings.Count(tagContent, "\n") // Update line count

		// Process the content between the tags based on tag type
		if tagType == TOKEN_COMMENT_START {
			// For comments, just store the content as a TEXT token
			if len(tagContent) > 0 {
				tokens = append(tokens, createToken(TOKEN_TEXT, tagContent, line))
			}
		} else {
			// For variable and block tags, tokenize the content properly
			// Trim whitespace from the tag content
			tagContent = strings.TrimSpace(tagContent)

			if tagType == TOKEN_BLOCK_START || tagType == TOKEN_BLOCK_START_TRIM {
				// Process block tags like if, for, etc.
				// First, extract the tag name
				parts := strings.SplitN(tagContent, " ", 2)
				if len(parts) > 0 {
					blockName := parts[0]
					tokens = append(tokens, createToken(TOKEN_NAME, blockName, line))

					// Different handling based on block type
					if blockName == "if" || blockName == "elseif" {
						// For if/elseif blocks, tokenize the condition
						if len(parts) > 1 {
							condition := strings.TrimSpace(parts[1])
							// Tokenize the condition properly
							// Special tokenization for important expressions
							p.tokenizeExpression(condition, &tokens, line)
						}
					} else if blockName == "for" {
						// For for loops, tokenize iterator variables and collection
						if len(parts) > 1 {
							forExpr := strings.TrimSpace(parts[1])
							// Check for proper "in" keyword
							inPos := strings.Index(strings.ToLower(forExpr), " in ")
							if inPos != -1 {
								// Extract iterators and collection
								iterators := strings.TrimSpace(forExpr[:inPos])
								collection := strings.TrimSpace(forExpr[inPos+4:])

								// Handle key, value iterators (e.g., "key, value in collection")
								if strings.Contains(iterators, ",") {
									iterParts := strings.SplitN(iterators, ",", 2)
									if len(iterParts) == 2 {
										keyVar := strings.TrimSpace(iterParts[0])
										valueVar := strings.TrimSpace(iterParts[1])

										// Add tokens for key and value variables
										tokens = append(tokens, createToken(TOKEN_NAME, keyVar, line))
										tokens = append(tokens, createToken(TOKEN_PUNCTUATION, ",", line))
										tokens = append(tokens, createToken(TOKEN_NAME, valueVar, line))
									}
								} else {
									// Single iterator variable
									tokens = append(tokens, createToken(TOKEN_NAME, iterators, line))
								}

								// Add "in" keyword
								tokens = append(tokens, createToken(TOKEN_NAME, "in", line))

								// Check if collection is a function call (contains ( and ))
								if strings.Contains(collection, "(") && strings.Contains(collection, ")") {
									// Tokenize the collection as a complex expression
									p.tokenizeExpression(collection, &tokens, line)
								} else {
									// Add collection as a simple variable
									tokens = append(tokens, createToken(TOKEN_NAME, collection, line))
								}
							} else {
								// Fallback if "in" keyword not found
								tokens = append(tokens, createToken(TOKEN_NAME, forExpr, line))
							}
						}
					} else if blockName == "include" {
						// Special handling for include tag with quoted template names
						if len(parts) > 1 {
							includeExpr := strings.TrimSpace(parts[1])

							// First check if we have a 'with' keyword which separates template name from params
							withPos := strings.Index(strings.ToLower(includeExpr), " with ")

							if withPos > 0 {
								// Split the include expression into template name and parameters
								templatePart := strings.TrimSpace(includeExpr[:withPos])
								paramsPart := strings.TrimSpace(includeExpr[withPos+6:]) // +6 to skip " with "

								// Handle quoted template names
								if (strings.HasPrefix(templatePart, "\"") && strings.HasSuffix(templatePart, "\"")) ||
									(strings.HasPrefix(templatePart, "'") && strings.HasSuffix(templatePart, "'")) {
									// Extract the template name without quotes
									templateName := templatePart[1 : len(templatePart)-1]
									// Add as a string token
									tokens = append(tokens, createToken(TOKEN_STRING, templateName, line))
								} else {
									// Unquoted name, add as name token
									tokens = append(tokens, createToken(TOKEN_NAME, templatePart, line))
								}

								// Add "with" keyword
								tokens = append(tokens, createToken(TOKEN_NAME, "with", line))

								// Add opening brace for the parameters
								tokens = append(tokens, createToken(TOKEN_PUNCTUATION, "{", line))

								// For parameters that might include nested objects, we need a different approach
								// Tokenize the parameter string, preserving nested structures
								tokenizeComplexObject(paramsPart, &tokens, line)

								// Add closing brace
								tokens = append(tokens, createToken(TOKEN_PUNCTUATION, "}", line))
							} else {
								// No 'with' keyword, just a template name
								if (strings.HasPrefix(includeExpr, "\"") && strings.HasSuffix(includeExpr, "\"")) ||
									(strings.HasPrefix(includeExpr, "'") && strings.HasSuffix(includeExpr, "'")) {
									// Extract template name without quotes
									templateName := includeExpr[1 : len(includeExpr)-1]
									// Add as a string token
									tokens = append(tokens, createToken(TOKEN_STRING, templateName, line))
								} else {
									// Not quoted, add as name token
									tokens = append(tokens, createToken(TOKEN_NAME, includeExpr, line))
								}
							}
						}
					} else if blockName == "extends" {
						// Special handling for extends tag with quoted template names
						if len(parts) > 1 {
							extendsExpr := strings.TrimSpace(parts[1])

							// Handle quoted template names
							if (strings.HasPrefix(extendsExpr, "\"") && strings.HasSuffix(extendsExpr, "\"")) ||
								(strings.HasPrefix(extendsExpr, "'") && strings.HasSuffix(extendsExpr, "'")) {
								// Extract the template name without quotes
								templateName := extendsExpr[1 : len(extendsExpr)-1]
								// Add as a string token
								tokens = append(tokens, createToken(TOKEN_STRING, templateName, line))
							} else {
								// Not quoted, tokenize as a normal expression
								p.tokenizeExpression(extendsExpr, &tokens, line)
							}
						}
					} else if blockName == "set" {
						// Special handling for set tag to properly tokenize variable assignments
						if len(parts) > 1 {
							setExpr := strings.TrimSpace(parts[1])

							// Check for the assignment operator
							assignPos := strings.Index(setExpr, "=")

							if assignPos != -1 {
								// Split into variable name and value
								varName := strings.TrimSpace(setExpr[:assignPos])
								value := strings.TrimSpace(setExpr[assignPos+1:])

								// Add the variable name token
								tokens = append(tokens, createToken(TOKEN_NAME, varName, line))

								// Add the assignment operator
								tokens = append(tokens, createToken(TOKEN_OPERATOR, "=", line))

								// Tokenize the value expression
								p.tokenizeExpression(value, &tokens, line)
							} else {
								// Handle case without assignment (e.g., {% set var %})
								tokens = append(tokens, createToken(TOKEN_NAME, setExpr, line))
							}
						}
					} else {
						// For other block types, just add parameters as NAME tokens
						if len(parts) > 1 {
							tokens = append(tokens, createToken(TOKEN_NAME, parts[1], line))
						}
					}
				}
			} else {
				// For variable tags, tokenize the expression
				if len(tagContent) > 0 {
					// If it's a simple variable name, add it directly
					if !strings.ContainsAny(tagContent, ".|[](){}\"',+-*/=!<>%&^~") {
						tokens = append(tokens, createToken(TOKEN_NAME, tagContent, line))
					} else {
						// For complex expressions, tokenize properly
						p.tokenizeExpression(tagContent, &tokens, line)
					}
				}
			}
		}

		// Add the end tag token
		tokens = append(tokens, createToken(endTagType, "", line))

		// Move past the end tag
		currentPosition = currentPosition + endPos + endTagLength
	}

	// Add EOF token
	tokens = append(tokens, createToken(TOKEN_EOF, "", line))
	return tokens, nil
}

// Param represents a key-value parameter from include tag
type Param struct {
	key      string
	value    string
	rawValue string // Original value before quote stripping
}

// extractParams parses a parameter string like {'key': 'value', 'key2': 'value2'}
// and returns a slice of key-value parameters
// tokenizeComplexObject parses and tokenizes a complex object with nested structures
// for use with include parameters
func tokenizeComplexObject(objStr string, tokens *[]Token, line int) {
	// First strip outer braces if present
	objStr = strings.TrimSpace(objStr)
	if strings.HasPrefix(objStr, "{") && strings.HasSuffix(objStr, "}") {
		objStr = strings.TrimSpace(objStr[1 : len(objStr)-1])
	}

	// Tokenize the object contents
	tokenizeObjectContents(objStr, tokens, line)
}

// tokenizeObjectContents parses key-value pairs within an object
func tokenizeObjectContents(content string, tokens *[]Token, line int) {
	// Estimate the number of key-value pairs to pre-allocate token space
	commaCount := 0
	for i := 0; i < len(content); i++ {
		if content[i] == ',' {
			commaCount++
		}
	}

	// Pre-grow the tokens slice: each key-value pair creates about 4 tokens on average
	estimatedTokenCount := len(*tokens) + (commaCount+1)*4
	if cap(*tokens) < estimatedTokenCount {
		newTokens := make([]Token, len(*tokens), estimatedTokenCount)
		copy(newTokens, *tokens)
		*tokens = newTokens
	}

	// State tracking
	inSingleQuote := false
	inDoubleQuote := false
	inObject := 0 // Nesting level for objects
	inArray := 0  // Nesting level for arrays

	start := 0
	colonPos := -1

	for i := 0; i <= len(content); i++ {
		// At the end of the string or at a comma at the top level
		atEnd := i == len(content)
		isComma := !atEnd && content[i] == ','

		if (isComma || atEnd) && inObject == 0 && inArray == 0 && !inSingleQuote && !inDoubleQuote {
			// We've found the end of a key-value pair
			if colonPos != -1 {
				// Extract the key and value - reuse same slice memory
				keyStr := content[start:colonPos]
				keyStr = strings.TrimSpace(keyStr)
				valueStr := content[colonPos+1 : i]
				valueStr = strings.TrimSpace(valueStr)

				// Check key characteristics once to avoid multiple prefix/suffix checks
				keyHasSingleQuotes := len(keyStr) >= 2 && keyStr[0] == '\'' && keyStr[len(keyStr)-1] == '\''
				keyHasDoubleQuotes := len(keyStr) >= 2 && keyStr[0] == '"' && keyStr[len(keyStr)-1] == '"'

				// Process the key
				if keyHasSingleQuotes || keyHasDoubleQuotes {
					// Quoted key - add as a string token
					*tokens = append(*tokens, Token{
						Type:  TOKEN_STRING,
						Value: keyStr[1 : len(keyStr)-1], // Remove quotes
						Line:  line,
					})
				} else {
					// Unquoted key
					*tokens = append(*tokens, createToken(TOKEN_NAME, keyStr, line))
				}

				// Add colon separator - using static string to avoid allocation
				*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, ":", line))

				// Check value characteristics once
				valueHasSingleQuotes := len(valueStr) >= 2 && valueStr[0] == '\'' && valueStr[len(valueStr)-1] == '\''
				valueHasDoubleQuotes := len(valueStr) >= 2 && valueStr[0] == '"' && valueStr[len(valueStr)-1] == '"'
				valueStartsWithBrace := len(valueStr) >= 2 && valueStr[0] == '{'
				valueEndsWithBrace := len(valueStr) >= 1 && valueStr[len(valueStr)-1] == '}'
				valueStartsWithBracket := len(valueStr) >= 2 && valueStr[0] == '['
				valueEndsWithBracket := len(valueStr) >= 1 && valueStr[len(valueStr)-1] == ']'

				// Process the value - more complex values need special handling
				if valueStartsWithBrace && valueEndsWithBrace {
					// Nested object
					*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, "{", line))
					// Pass substring directly to avoid another allocation
					tokenizeObjectContents(valueStr[1:len(valueStr)-1], tokens, line)
					*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, "}", line))
				} else if valueStartsWithBracket && valueEndsWithBracket {
					// Array
					*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, "[", line))
					tokenizeArrayElements(valueStr[1:len(valueStr)-1], tokens, line)
					*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, "]", line))
				} else if valueHasSingleQuotes || valueHasDoubleQuotes {
					// String literal - reuse memory by slicing
					*tokens = append(*tokens, Token{
						Type:  TOKEN_STRING,
						Value: valueStr[1 : len(valueStr)-1], // Remove quotes
						Line:  line,
					})
				} else if isNumericValue(valueStr) {
					// Numeric value
					*tokens = append(*tokens, createToken(TOKEN_NUMBER, valueStr, line))
				} else if valueStr == "true" || valueStr == "false" {
					// Boolean literal
					*tokens = append(*tokens, createToken(TOKEN_NAME, valueStr, line))
				} else if valueStr == "null" || valueStr == "nil" {
					// Null/nil literal
					*tokens = append(*tokens, createToken(TOKEN_NAME, valueStr, line))
				} else {
					// Variable or other value
					*tokens = append(*tokens, createToken(TOKEN_NAME, valueStr, line))
				}

				// Add comma if needed - using static string to avoid allocation
				if isComma && i < len(content)-1 {
					*tokens = append(*tokens, Token{Type: TOKEN_PUNCTUATION, Value: ",", Line: line})
				}

				// Reset state for next key-value pair
				start = i + 1
				colonPos = -1
			}
			continue
		}

		// Handle quotes - they affect how we interpret other characters
		if i < len(content) {
			c := content[i]

			// Handle quote characters
			if c == '\'' && (i == 0 || content[i-1] != '\\') {
				inSingleQuote = !inSingleQuote
			} else if c == '"' && (i == 0 || content[i-1] != '\\') {
				inDoubleQuote = !inDoubleQuote
			}

			// Skip everything inside quotes
			if inSingleQuote || inDoubleQuote {
				continue
			}

			// Handle object and array nesting
			if c == '{' {
				inObject++
			} else if c == '}' {
				inObject--
			} else if c == '[' {
				inArray++
			} else if c == ']' {
				inArray--
			}

			// Find the colon separator if we're not in a nested structure
			if c == ':' && inObject == 0 && inArray == 0 && colonPos == -1 {
				colonPos = i
			}
		}
	}
}

// tokenizeArrayElements parses and tokenizes array elements
func tokenizeArrayElements(arrStr string, tokens *[]Token, line int) {
	// Split the array elements, respecting nested structures
	elements := splitArrayElements(arrStr)

	// Pre-grow the tokens slice to minimize reallocations
	// Estimate: each element adds at least 1 token plus commas between elements
	estimatedTokenCount := len(*tokens) + len(elements)*2
	if cap(*tokens) < estimatedTokenCount {
		newTokens := make([]Token, len(*tokens), estimatedTokenCount)
		copy(newTokens, *tokens)
		*tokens = newTokens
	}

	// Add each element with appropriate tokens
	for i, element := range elements {
		element = strings.TrimSpace(element)
		if element == "" {
			continue
		}

		// Process the element based on its type
		firstChar := byte(0)
		lastChar := byte(0)
		if len(element) > 0 {
			firstChar = element[0]
			lastChar = element[len(element)-1]
		}

		if firstChar == '{' && lastChar == '}' {
			// Nested object
			*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, "{", line))
			// Pass substring directly to avoid another allocation
			tokenizeObjectContents(element[1:len(element)-1], tokens, line)
			*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, "}", line))
		} else if firstChar == '[' && lastChar == ']' {
			// Nested array
			*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, "[", line))
			tokenizeArrayElements(element[1:len(element)-1], tokens, line)
			*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, "]", line))
		} else if (firstChar == '\'' && lastChar == '\'') ||
			(firstChar == '"' && lastChar == '"') {
			// String literal - reuse memory by slicing rather than copying
			*tokens = append(*tokens, Token{
				Type:  TOKEN_STRING,
				Value: element[1 : len(element)-1], // Remove quotes
				Line:  line,
			})
		} else if isNumericValue(element) {
			// Numeric value
			*tokens = append(*tokens, createToken(TOKEN_NUMBER, element, line))
		} else if element == "true" || element == "false" {
			// Boolean literal
			*tokens = append(*tokens, createToken(TOKEN_NAME, element, line))
		} else if element == "null" || element == "nil" {
			// Null/nil literal
			*tokens = append(*tokens, createToken(TOKEN_NAME, element, line))
		} else {
			// Variable or other value
			*tokens = append(*tokens, createToken(TOKEN_NAME, element, line))
		}

		// Add comma between elements (except last)
		if i < len(elements)-1 {
			*tokens = append(*tokens, Token{Type: TOKEN_PUNCTUATION, Value: ",", Line: line})
		}
	}
}

// splitArrayElements splits array elements respecting nested structures
func splitArrayElements(arrStr string) []string {
	// Estimate the number of elements to pre-allocate the slice
	commaCount := 0
	for i := 0; i < len(arrStr); i++ {
		if arrStr[i] == ',' {
			commaCount++
		}
	}
	// Allocate with capacity for estimated number of elements
	elements := make([]string, 0, commaCount+1)

	// Pre-allocate the string builder with a reasonable capacity
	// to avoid frequent reallocations
	var current strings.Builder
	current.Grow(len(arrStr) / (commaCount + 1)) // Average element size

	inSingleQuote := false
	inDoubleQuote := false
	inObject := 0
	inArray := 0

	for i := 0; i < len(arrStr); i++ {
		c := arrStr[i]

		// Handle quotes
		if c == '\'' && (i == 0 || arrStr[i-1] != '\\') {
			inSingleQuote = !inSingleQuote
		} else if c == '"' && (i == 0 || arrStr[i-1] != '\\') {
			inDoubleQuote = !inDoubleQuote
		}

		// Handle nesting
		if !inSingleQuote && !inDoubleQuote {
			if c == '{' {
				inObject++
			} else if c == '}' {
				inObject--
			} else if c == '[' {
				inArray++
			} else if c == ']' {
				inArray--
			}
		}

		// Handle element separator
		if c == ',' && !inSingleQuote && !inDoubleQuote && inObject == 0 && inArray == 0 {
			elements = append(elements, current.String())
			current.Reset()
			continue
		}

		// Add the character to the current element
		current.WriteByte(c)
	}

	// Add the last element if any
	if current.Len() > 0 {
		elements = append(elements, current.String())
	}

	return elements
}

// isNumericValue checks if a string is a valid number (integer or float)
func isNumericValue(s string) bool {
	// Check for empty string
	if s == "" {
		return false
	}

	// Check if it's a valid integer or float (including negative numbers)
	_, err1 := strconv.Atoi(s)
	if err1 == nil {
		return true
	}

	_, err2 := strconv.ParseFloat(s, 64)
	return err2 == nil
}

// onlyContainsDigitsOrDot checks if a string only contains digits, at most one dot, and optionally a leading minus sign
func onlyContainsDigitsOrDot(s string) bool {
	if s == "" {
		return false
	}

	dotCount := 0
	for i, c := range s {
		// Allow a minus sign, but only at the beginning of the string
		if c == '-' && i == 0 {
			continue
		}

		if c == '.' {
			dotCount++
			if dotCount > 1 {
				return false
			}
		} else if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// findMatchingParen finds the matching closing parenthesis for an opening one
func findMatchingParen(s string, openPos int) int {
	if openPos >= len(s) || s[openPos] != '(' {
		return -1
	}

	nestLevel := 1
	for i := openPos + 1; i < len(s); i++ {
		if s[i] == '(' {
			nestLevel++
		} else if s[i] == ')' {
			nestLevel--
			if nestLevel == 0 {
				return i
			}
		}
	}

	return -1 // No matching closing parenthesis found
}

// extractParams (deprecated) - kept for reference
func extractParams(paramStr string) []Param {
	var params []Param

	// Strip outer braces if present
	paramStr = strings.TrimSpace(paramStr)
	if strings.HasPrefix(paramStr, "{") && strings.HasSuffix(paramStr, "}") {
		paramStr = strings.TrimSpace(paramStr[1 : len(paramStr)-1])
	}

	// Simple parsing approach - split by commas, but respect quotes
	var parts []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(paramStr); i++ {
		c := paramStr[i]

		// Handle quote characters
		if c == '\'' && (i == 0 || paramStr[i-1] != '\\') {
			inSingleQuote = !inSingleQuote
			current.WriteByte(c)
			continue
		}
		if c == '"' && (i == 0 || paramStr[i-1] != '\\') {
			inDoubleQuote = !inDoubleQuote
			current.WriteByte(c)
			continue
		}

		// Handle commas - only split on commas outside of quotes
		if c == ',' && !inSingleQuote && !inDoubleQuote {
			// Add the current part
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}

	// Add the last part if any
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	// Process each part as a key:value pair
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Find the key-value separator
		colonPos := -1
		inQuote := false
		quoteChar := byte(0)

		for i := 0; i < len(part); i++ {
			c := part[i]

			// Handle quotes
			if (c == '"' || c == '\'') && (i == 0 || part[i-1] != '\\') {
				if !inQuote {
					inQuote = true
					quoteChar = c
				} else if c == quoteChar {
					inQuote = false
				}
				continue
			}

			// Find colon outside of quotes
			if c == ':' && !inQuote && colonPos == -1 {
				colonPos = i
				break
			}
		}

		// If no colon found, skip this part
		if colonPos == -1 {
			continue
		}

		// Extract key and value
		keyStr := strings.TrimSpace(part[:colonPos])
		valueStr := strings.TrimSpace(part[colonPos+1:])

		// Store the raw value before removing quotes
		rawValueStr := valueStr

		// Remove quotes if present
		if (strings.HasPrefix(keyStr, "'") && strings.HasSuffix(keyStr, "'")) ||
			(strings.HasPrefix(keyStr, "\"") && strings.HasSuffix(keyStr, "\"")) {
			keyStr = keyStr[1 : len(keyStr)-1]
		}

		if (strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'")) ||
			(strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"")) {
			valueStr = valueStr[1 : len(valueStr)-1]
		}

		// Add the parameter
		params = append(params, Param{key: keyStr, value: valueStr, rawValue: rawValueStr})
	}

	return params
}

// Helper method to tokenize a string and add tokens directly to the given token slice
func (p *Parser) tokenizeAndAppend(source string, tokens *[]Token, line int) {
	// Use a simplified approach - just add the expression as a name token
	// The parser will handle the include parameters later
	*tokens = append(*tokens, createToken(TOKEN_NAME, source, line))
}

// These helper functions (isDigit, isOperator, isPunctuation, isWhitespace) are defined in parser.go

// tokenizeExpression handles tokenizing expressions inside Twig tags
func (p *Parser) tokenizeExpression(expr string, tokens *[]Token, line int) {
	// This is a tokenizer for expressions that handles Twig syntax properly

	// First, pre-grow the tokens slice to minimize reallocations
	// Estimate: one token per 5 characters in the expression (rough average)
	estimatedTokenCount := len(*tokens) + len(expr)/5 + 1
	if cap(*tokens) < estimatedTokenCount {
		newTokens := make([]Token, len(*tokens), estimatedTokenCount)
		copy(newTokens, *tokens)
		*tokens = newTokens
	}

	// Pre-allocate the string builder with a reasonable capacity
	var currentToken strings.Builder
	currentToken.Grow(16) // Reasonable size for most identifiers/numbers

	var inString bool
	var stringDelimiter byte

	for i := 0; i < len(expr); i++ {
		c := expr[i]

		// Handle string literals with quotes
		if (c == '"' || c == '\'') && (i == 0 || expr[i-1] != '\\') {
			if inString && c == stringDelimiter {
				// End of string
				inString = false
				*tokens = append(*tokens, createToken(TOKEN_STRING, currentToken.String(), line))
				currentToken.Reset()
			} else if !inString {
				// Start of string
				inString = true
				stringDelimiter = c
				// First add any accumulated token
				if currentToken.Len() > 0 {
					// Reuse the tokenValue to avoid extra allocations
					tokenValue := currentToken.String()

					// Check if the token is a number
					if onlyContainsDigitsOrDot(tokenValue) {
						*tokens = append(*tokens, createToken(TOKEN_NUMBER, tokenValue, line))
					} else {
						*tokens = append(*tokens, createToken(TOKEN_NAME, tokenValue, line))
					}
					currentToken.Reset()
				}
			} else {
				// Quote inside a string with different delimiter
				currentToken.WriteByte(c)
			}
			continue
		}

		// If we're inside a string, just accumulate characters
		if inString {
			currentToken.WriteByte(c)
			continue
		}

		// Handle colons (key-value separator in objects) specially
		if c == ':' {
			// First, add any accumulated token
			if currentToken.Len() > 0 {
				tokenValue := currentToken.String()
				// Check if the token is a number
				if onlyContainsDigitsOrDot(tokenValue) {
					*tokens = append(*tokens, createToken(TOKEN_NUMBER, tokenValue, line))
				} else {
					*tokens = append(*tokens, createToken(TOKEN_NAME, tokenValue, line))
				}
				currentToken.Reset()
			}

			// Add the colon token
			*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, ":", line))
			continue
		}

		// Check for operators (including two-character operators)
		if isOperator(c) {
			// First, add any accumulated token
			if currentToken.Len() > 0 {
				tokenValue := currentToken.String()
				// Check if the token is a number
				if onlyContainsDigitsOrDot(tokenValue) {
					*tokens = append(*tokens, createToken(TOKEN_NUMBER, tokenValue, line))
				} else {
					*tokens = append(*tokens, createToken(TOKEN_NAME, tokenValue, line))
				}
				currentToken.Reset()
			}

			// Check for two-character operators
			if i+1 < len(expr) {
				nextChar := expr[i+1]
				isTwoChar := false
				var twoChar string

				// Avoiding string concatenation and using direct comparison
				if (c == '=' && nextChar == '=') ||
					(c == '!' && nextChar == '=') ||
					(c == '>' && nextChar == '=') ||
					(c == '<' && nextChar == '=') ||
					(c == '&' && nextChar == '&') ||
					(c == '|' && nextChar == '|') ||
					(c == '?' && nextChar == '?') {

					// Only allocate the string when we need it
					twoChar = string([]byte{c, nextChar})
					isTwoChar = true
				}

				if isTwoChar {
					*tokens = append(*tokens, createToken(TOKEN_OPERATOR, twoChar, line))
					i++ // Skip the next character
					continue
				}
			}

			// Add the single-character operator token
			*tokens = append(*tokens, createToken(TOKEN_OPERATOR, string([]byte{c}), line))
			continue
		}

		// Check for punctuation (braces, brackets, parens, commas)
		if isPunctuation(c) {
			// First, add any accumulated token
			if currentToken.Len() > 0 {
				tokenValue := currentToken.String()
				// Check if the token is a number
				if onlyContainsDigitsOrDot(tokenValue) {
					*tokens = append(*tokens, createToken(TOKEN_NUMBER, tokenValue, line))
				} else {
					*tokens = append(*tokens, createToken(TOKEN_NAME, tokenValue, line))
				}
				currentToken.Reset()
			}

			// Add the punctuation token
			*tokens = append(*tokens, createToken(TOKEN_PUNCTUATION, string([]byte{c}), line))
			continue
		}

		// Check for whitespace
		if isWhitespace(c) {
			// Add any accumulated token
			if currentToken.Len() > 0 {
				tokenValue := currentToken.String()
				// Check if the token is a number
				if onlyContainsDigitsOrDot(tokenValue) {
					*tokens = append(*tokens, createToken(TOKEN_NUMBER, tokenValue, line))
				} else {
					*tokens = append(*tokens, createToken(TOKEN_NAME, tokenValue, line))
				}
				currentToken.Reset()
			}
			continue
		}

		// Handle negative numbers and regular numbers
		if currentToken.Len() == 0 &&
			((c == '-' && i+1 < len(expr) && isDigit(expr[i+1])) || // Negative number
				isDigit(c) ||
				(c == '.' && i+1 < len(expr) && isDigit(expr[i+1]))) { // Regular number or decimal

			// Pre-allocate for numeric tokens
			currentToken.Grow(10) // Reasonable for most numbers

			// Handle negative sign if present
			if c == '-' {
				currentToken.WriteByte(c)
				i++
				c = expr[i] // Move to next character (should be a digit)
			}

			// Start of a number
			currentToken.WriteByte(c)

			// Consume all digits and at most one decimal point
			hasDot := c == '.'
			i++
			for i < len(expr) {
				if isDigit(expr[i]) {
					currentToken.WriteByte(expr[i])
				} else if expr[i] == '.' && !hasDot {
					hasDot = true
					currentToken.WriteByte(expr[i])
				} else {
					break
				}
				i++
			}
			i-- // Adjust for the outer loop increment

			*tokens = append(*tokens, createToken(TOKEN_NUMBER, currentToken.String(), line))
			currentToken.Reset()
			continue
		}

		// Accumulate the character into the current token
		currentToken.WriteByte(c)
	}

	// Add any final token
	if currentToken.Len() > 0 {
		tokenValue := currentToken.String()

		// Check if the final token is a special keyword
		// Use direct comparison instead of multiple string checks
		if tokenValue == "true" || tokenValue == "false" ||
			tokenValue == "null" || tokenValue == "nil" {
			// These are literals, not names
			*tokens = append(*tokens, createToken(TOKEN_NAME, tokenValue, line))
		} else if onlyContainsDigitsOrDot(tokenValue) {
			// It's a number
			*tokens = append(*tokens, createToken(TOKEN_NUMBER, tokenValue, line))
		} else {
			// Regular name token
			*tokens = append(*tokens, createToken(TOKEN_NAME, tokenValue, line))
		}
	}
}
