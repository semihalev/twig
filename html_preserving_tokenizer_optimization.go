package twig

import (
	"fmt"
	"strings"
)

// optimizedHtmlPreservingTokenize is an optimized version of htmlPreservingTokenize
// that reduces memory allocations by reusing token objects and slices
func (p *Parser) optimizedHtmlPreservingTokenize() ([]Token, error) {
	// Pre-allocate tokens with estimated capacity based on source length
	estimatedTokenCount := len(p.source) / 20 // Rough estimate: one token per 20 chars
	tokenSlice := GetPooledTokenSlice(estimatedTokenCount)
	
	// Ensure the token slice is released even if an error occurs
	defer tokenSlice.Release()
	
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
				tokenSlice.AppendToken(TOKEN_TEXT, preText, line)
				line += countNewlines(preText)
			}

			// Add the tag itself as literal text (without the backslash)
			tokenSlice.AppendToken(TOKEN_TEXT, matchedPos.pattern, line)

			// Move past the tag
			currentPosition = nextTagPos + matchedPos.length
			continue
		}

		if nextTagPos == -1 {
			// No more tags found, add the rest as TEXT
			content := p.source[currentPosition:]
			if len(content) > 0 {
				line += countNewlines(content)
				tokenSlice.AppendToken(TOKEN_TEXT, content, line)
			}
			break
		}

		// Add the text before the tag (HTML content)
		if nextTagPos > currentPosition {
			content := p.source[currentPosition:nextTagPos]
			line += countNewlines(content)
			tokenSlice.AppendToken(TOKEN_TEXT, content, line)
		}

		// Add the tag start token
		tokenSlice.AppendToken(tagType, "", line)

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
		line += countNewlines(tagContent) // Update line count

		// Process the content between the tags based on tag type
		if tagType == TOKEN_COMMENT_START {
			// For comments, just store the content as a TEXT token
			if len(tagContent) > 0 {
				tokenSlice.AppendToken(TOKEN_TEXT, tagContent, line)
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
					tokenSlice.AppendToken(TOKEN_NAME, blockName, line)

					// Different handling based on block type
					if blockName == "if" || blockName == "elseif" {
						// For if/elseif blocks, tokenize the condition
						if len(parts) > 1 {
							condition := strings.TrimSpace(parts[1])
							// Tokenize the condition properly
							p.optimizedTokenizeExpression(condition, tokenSlice, line)
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
										tokenSlice.AppendToken(TOKEN_NAME, keyVar, line)
										tokenSlice.AppendToken(TOKEN_PUNCTUATION, ",", line)
										tokenSlice.AppendToken(TOKEN_NAME, valueVar, line)
									}
								} else {
									// Single iterator variable
									tokenSlice.AppendToken(TOKEN_NAME, iterators, line)
								}

								// Add "in" keyword
								tokenSlice.AppendToken(TOKEN_NAME, "in", line)

								// Check if collection is a function call (contains ( and ))
								if strings.Contains(collection, "(") && strings.Contains(collection, ")") {
									// Tokenize the collection as a complex expression
									p.optimizedTokenizeExpression(collection, tokenSlice, line)
								} else {
									// Add collection as a simple variable
									tokenSlice.AppendToken(TOKEN_NAME, collection, line)
								}
							} else {
								// Fallback if "in" keyword not found
								tokenSlice.AppendToken(TOKEN_NAME, forExpr, line)
							}
						}
					} else if blockName == "do" {
						// Special handling for do tag with assignments and expressions
						if len(parts) > 1 {
							doExpr := strings.TrimSpace(parts[1])

							// Check if it's an assignment (contains =)
							assignPos := strings.Index(doExpr, "=")
							if assignPos > 0 && !strings.Contains(doExpr[:assignPos], "==") {
								// It's an assignment
								varName := strings.TrimSpace(doExpr[:assignPos])
								valueExpr := strings.TrimSpace(doExpr[assignPos+1:])

								// Add the variable name
								tokenSlice.AppendToken(TOKEN_NAME, varName, line)

								// Add the equals sign
								tokenSlice.AppendToken(TOKEN_OPERATOR, "=", line)

								// Tokenize the expression on the right side
								p.optimizedTokenizeExpression(valueExpr, tokenSlice, line)
							} else {
								// It's just an expression, tokenize it
								p.optimizedTokenizeExpression(doExpr, tokenSlice, line)
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
									tokenSlice.AppendToken(TOKEN_STRING, templateName, line)
								} else {
									// Unquoted name, add as name token
									tokenSlice.AppendToken(TOKEN_NAME, templatePart, line)
								}

								// Add "with" keyword
								tokenSlice.AppendToken(TOKEN_NAME, "with", line)

								// Add opening brace for the parameters
								tokenSlice.AppendToken(TOKEN_PUNCTUATION, "{", line)

								// For parameters that might include nested objects, we need a different approach
								// Tokenize the parameter string, preserving nested structures
								optimizedTokenizeComplexObject(paramsPart, tokenSlice, line)

								// Add closing brace
								tokenSlice.AppendToken(TOKEN_PUNCTUATION, "}", line)
							} else {
								// No 'with' keyword, just a template name
								if (strings.HasPrefix(includeExpr, "\"") && strings.HasSuffix(includeExpr, "\"")) ||
									(strings.HasPrefix(includeExpr, "'") && strings.HasSuffix(includeExpr, "'")) {
									// Extract template name without quotes
									templateName := includeExpr[1 : len(includeExpr)-1]
									// Add as a string token
									tokenSlice.AppendToken(TOKEN_STRING, templateName, line)
								} else {
									// Not quoted, add as name token
									tokenSlice.AppendToken(TOKEN_NAME, includeExpr, line)
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
								tokenSlice.AppendToken(TOKEN_STRING, templateName, line)
							} else {
								// Not quoted, tokenize as a normal expression
								p.optimizedTokenizeExpression(extendsExpr, tokenSlice, line)
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
								tokenSlice.AppendToken(TOKEN_NAME, varName, line)

								// Add the assignment operator
								tokenSlice.AppendToken(TOKEN_OPERATOR, "=", line)

								// Tokenize the value expression
								p.optimizedTokenizeExpression(value, tokenSlice, line)
							} else {
								// Handle case without assignment (e.g., {% set var %})
								tokenSlice.AppendToken(TOKEN_NAME, setExpr, line)
							}
						}
					} else {
						// For other block types, just add parameters as NAME tokens
						if len(parts) > 1 {
							tokenSlice.AppendToken(TOKEN_NAME, parts[1], line)
						}
					}
				}
			} else {
				// For variable tags, tokenize the expression
				if len(tagContent) > 0 {
					// If it's a simple variable name, add it directly
					if !strings.ContainsAny(tagContent, ".|[](){}\"',+-*/=!<>%&^~") {
						tokenSlice.AppendToken(TOKEN_NAME, tagContent, line)
					} else {
						// For complex expressions, tokenize properly
						p.optimizedTokenizeExpression(tagContent, tokenSlice, line)
					}
				}
			}
		}

		// Add the end tag token
		tokenSlice.AppendToken(endTagType, "", line)

		// Move past the end tag
		currentPosition = currentPosition + endPos + endTagLength
	}

	// Add EOF token
	tokenSlice.AppendToken(TOKEN_EOF, "", line)
	
	// Finalize and return the token slice
	return tokenSlice.Finalize(), nil
}

// optimizedTokenizeExpression handles tokenizing expressions inside Twig tags with reduced allocations
func (p *Parser) optimizedTokenizeExpression(expr string, tokens *PooledTokenSlice, line int) {
	var inString bool
	var stringDelimiter byte
	var stringStart int // Position where string content starts

	for i := 0; i < len(expr); i++ {
		c := expr[i]

		// Handle string literals with quotes
		if (c == '"' || c == '\'') && (i == 0 || expr[i-1] != '\\') {
			if inString && c == stringDelimiter {
				// End of string
				inString = false
				// Add the string token
				tokens.AppendToken(TOKEN_STRING, expr[stringStart:i], line)
			} else if !inString {
				// Start of string
				inString = true
				stringDelimiter = c
				// Remember the start position (for string content)
				stringStart = i + 1
			} else {
				// Quote inside a string with different delimiter
				// Skip
			}
			continue
		}

		// If we're inside a string, just skip this character
		if inString {
			continue
		}

		// Handle operators (including two-character operators)
		if isOperator(c) {
			// Check for two-character operators
			if i+1 < len(expr) {
				nextChar := expr[i+1]
				
				// Direct comparison for common two-char operators
				if (c == '=' && nextChar == '=') ||
					(c == '!' && nextChar == '=') ||
					(c == '>' && nextChar == '=') ||
					(c == '<' && nextChar == '=') ||
					(c == '&' && nextChar == '&') ||
					(c == '|' && nextChar == '|') ||
					(c == '?' && nextChar == '?') {
					
					// Add the two-character operator token
					tokens.AppendToken(TOKEN_OPERATOR, string([]byte{c, nextChar}), line)
					i++ // Skip the next character
					continue
				}
			}
			
			// Add single-character operator
			tokens.AppendToken(TOKEN_OPERATOR, string([]byte{c}), line)
			continue
		}

		// Handle punctuation
		if isPunctuation(c) {
			tokens.AppendToken(TOKEN_PUNCTUATION, string([]byte{c}), line)
			continue
		}

		// Handle whitespace - skip it
		if isWhitespace(c) {
			continue
		}

		// Handle identifiers and keywords
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			// Start of an identifier
			start := i
			
			// Find the end of the identifier
			for i++; i < len(expr) && ((expr[i] >= 'a' && expr[i] <= 'z') || 
				(expr[i] >= 'A' && expr[i] <= 'Z') || 
				(expr[i] >= '0' && expr[i] <= '9') || 
				expr[i] == '_'); i++ {
			}
			
			// Extract the identifier
			identifier := expr[start:i]
			i-- // Adjust for the loop increment
			
			// Add the token based on the identifier
			if identifier == "true" || identifier == "false" || identifier == "null" {
				tokens.AppendToken(TOKEN_NAME, identifier, line)
			} else {
				tokens.AppendToken(TOKEN_NAME, identifier, line)
			}
			
			continue
		}

		// Handle numbers
		if isDigit(c) || (c == '-' && i+1 < len(expr) && isDigit(expr[i+1])) {
			start := i
			
			// Skip the negative sign if present
			if c == '-' {
				i++
			}
			
			// Find the end of the number
			for i++; i < len(expr) && isDigit(expr[i]); i++ {
			}
			
			// Check for decimal point
			if i < len(expr) && expr[i] == '.' {
				i++
				// Find the end of the decimal part
				for ; i < len(expr) && isDigit(expr[i]); i++ {
				}
			}
			
			// Extract the number
			number := expr[start:i]
			i-- // Adjust for the loop increment
			
			// Add the number token
			tokens.AppendToken(TOKEN_NUMBER, number, line)
			continue
		}
	}
}

// optimizedTokenizeComplexObject parses and tokenizes a complex object with reduced allocations
func optimizedTokenizeComplexObject(objStr string, tokens *PooledTokenSlice, line int) {
	// First strip outer braces if present
	objStr = strings.TrimSpace(objStr)
	if strings.HasPrefix(objStr, "{") && strings.HasSuffix(objStr, "}") {
		objStr = strings.TrimSpace(objStr[1 : len(objStr)-1])
	}

	// Tokenize the object contents
	optimizedTokenizeObjectContents(objStr, tokens, line)
}

// optimizedTokenizeObjectContents parses key-value pairs with reduced allocations
func optimizedTokenizeObjectContents(content string, tokens *PooledTokenSlice, line int) {
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
				// Extract the key and value
				keyStr := strings.TrimSpace(content[start:colonPos])
				valueStr := strings.TrimSpace(content[colonPos+1 : i])

				// Process the key
				if (len(keyStr) >= 2 && keyStr[0] == '\'' && keyStr[len(keyStr)-1] == '\'') ||
					(len(keyStr) >= 2 && keyStr[0] == '"' && keyStr[len(keyStr)-1] == '"') {
					// Quoted key - add as a string token
					tokens.AppendToken(TOKEN_STRING, keyStr[1:len(keyStr)-1], line)
				} else {
					// Unquoted key
					tokens.AppendToken(TOKEN_NAME, keyStr, line)
				}

				// Add colon separator
				tokens.AppendToken(TOKEN_PUNCTUATION, ":", line)

				// Process the value based on type
				if len(valueStr) >= 2 && valueStr[0] == '{' && valueStr[len(valueStr)-1] == '}' {
					// Nested object
					tokens.AppendToken(TOKEN_PUNCTUATION, "{", line)
					optimizedTokenizeObjectContents(valueStr[1:len(valueStr)-1], tokens, line)
					tokens.AppendToken(TOKEN_PUNCTUATION, "}", line)
				} else if len(valueStr) >= 2 && valueStr[0] == '[' && valueStr[len(valueStr)-1] == ']' {
					// Array
					tokens.AppendToken(TOKEN_PUNCTUATION, "[", line)
					optimizedTokenizeArrayElements(valueStr[1:len(valueStr)-1], tokens, line)
					tokens.AppendToken(TOKEN_PUNCTUATION, "]", line)
				} else if (len(valueStr) >= 2 && valueStr[0] == '\'' && valueStr[len(valueStr)-1] == '\'') ||
					(len(valueStr) >= 2 && valueStr[0] == '"' && valueStr[len(valueStr)-1] == '"') {
					// String literal
					tokens.AppendToken(TOKEN_STRING, valueStr[1:len(valueStr)-1], line)
				} else if isNumericValue(valueStr) {
					// Numeric value
					tokens.AppendToken(TOKEN_NUMBER, valueStr, line)
				} else if valueStr == "true" || valueStr == "false" {
					// Boolean literal
					tokens.AppendToken(TOKEN_NAME, valueStr, line)
				} else if valueStr == "null" || valueStr == "nil" {
					// Null/nil literal
					tokens.AppendToken(TOKEN_NAME, valueStr, line)
				} else {
					// Variable or other value
					tokens.AppendToken(TOKEN_NAME, valueStr, line)
				}

				// Add comma if needed
				if isComma && i < len(content)-1 {
					tokens.AppendToken(TOKEN_PUNCTUATION, ",", line)
				}

				// Reset state for next key-value pair
				start = i + 1
				colonPos = -1
			}
			continue
		}

		// Handle quotes and nested structures
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

// optimizedTokenizeArrayElements parses and tokenizes array elements with reduced allocations
func optimizedTokenizeArrayElements(arrStr string, tokens *PooledTokenSlice, line int) {
	// State tracking
	inSingleQuote := false
	inDoubleQuote := false
	inObject := 0
	inArray := 0
	
	// Track the start position of each element
	elemStart := 0
	
	for i := 0; i <= len(arrStr); i++ {
		// At the end of the string or at a comma at the top level
		atEnd := i == len(arrStr)
		isComma := !atEnd && arrStr[i] == ','
		
		// Process element when we reach a comma or the end
		if (isComma || atEnd) && inObject == 0 && inArray == 0 && !inSingleQuote && !inDoubleQuote {
			// Extract the element
			if i > elemStart {
				element := strings.TrimSpace(arrStr[elemStart:i])
				
				// Process the element based on its type
				if len(element) >= 2 {
					if element[0] == '{' && element[len(element)-1] == '}' {
						// Nested object
						tokens.AppendToken(TOKEN_PUNCTUATION, "{", line)
						optimizedTokenizeObjectContents(element[1:len(element)-1], tokens, line)
						tokens.AppendToken(TOKEN_PUNCTUATION, "}", line)
					} else if element[0] == '[' && element[len(element)-1] == ']' {
						// Nested array
						tokens.AppendToken(TOKEN_PUNCTUATION, "[", line)
						optimizedTokenizeArrayElements(element[1:len(element)-1], tokens, line)
						tokens.AppendToken(TOKEN_PUNCTUATION, "]", line)
					} else if (element[0] == '\'' && element[len(element)-1] == '\'') ||
						(element[0] == '"' && element[len(element)-1] == '"') {
						// String literal
						tokens.AppendToken(TOKEN_STRING, element[1:len(element)-1], line)
					} else if isNumericValue(element) {
						// Numeric value
						tokens.AppendToken(TOKEN_NUMBER, element, line)
					} else if element == "true" || element == "false" {
						// Boolean literal
						tokens.AppendToken(TOKEN_NAME, element, line)
					} else if element == "null" || element == "nil" {
						// Null/nil literal
						tokens.AppendToken(TOKEN_NAME, element, line)
					} else {
						// Variable or other value
						tokens.AppendToken(TOKEN_NAME, element, line)
					}
				}
			}
			
			// Add comma if needed
			if isComma && i < len(arrStr)-1 {
				tokens.AppendToken(TOKEN_PUNCTUATION, ",", line)
			}
			
			// Move to next element
			elemStart = i + 1
			continue
		}
		
		// Handle quotes and nested structures
		if !atEnd {
			c := arrStr[i]
			
			// Handle quote characters
			if c == '\'' && (i == 0 || arrStr[i-1] != '\\') {
				inSingleQuote = !inSingleQuote
			} else if c == '"' && (i == 0 || arrStr[i-1] != '\\') {
				inDoubleQuote = !inDoubleQuote
			}
			
			// Skip everything inside quotes
			if inSingleQuote || inDoubleQuote {
				continue
			}
			
			// Handle nesting
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
	}
}