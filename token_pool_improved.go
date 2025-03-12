package twig

import (
	"fmt"
	"strings"
	"sync"
)

// ImprovedTokenSlice is a more efficient implementation of a token slice pool 
// that truly minimizes allocations during tokenization
type ImprovedTokenSlice struct {
	tokens   []Token  // The actual token slice
	capacity int      // Capacity hint for the token slice
	used     bool     // Whether this slice has been used
}

// global pool for ImprovedTokenSlice objects
var improvedTokenSlicePool = sync.Pool{
	New: func() interface{} {
		// Start with a reasonably sized token slice
		tokens := make([]Token, 0, 64)
		return &ImprovedTokenSlice{
			tokens:   tokens,
			capacity: 64,
			used:     false,
		}
	},
}

// Global token object pool
var tokenObjectPool = sync.Pool{
	New: func() interface{} {
		return &Token{}
	},
}

// GetImprovedTokenSlice gets a token slice from the pool
func GetImprovedTokenSlice(capacityHint int) *ImprovedTokenSlice {
	slice := improvedTokenSlicePool.Get().(*ImprovedTokenSlice)
	
	// Reset the slice but keep capacity
	if cap(slice.tokens) < capacityHint {
		// Need to allocate a larger slice
		slice.tokens = make([]Token, 0, capacityHint)
		slice.capacity = capacityHint
	} else {
		// Reuse existing slice
		slice.tokens = slice.tokens[:0]
	}
	
	slice.used = false
	return slice
}

// AppendToken adds a token to the slice
func (s *ImprovedTokenSlice) AppendToken(tokenType int, value string, line int) {
	if s.used {
		return // Already finalized
	}
	
	// Create a token and add it to the slice
	token := Token{
		Type:  tokenType,
		Value: value,
		Line:  line,
	}
	
	s.tokens = append(s.tokens, token)
}

// Finalize returns the token slice
func (s *ImprovedTokenSlice) Finalize() []Token {
	if s.used {
		return s.tokens
	}
	
	s.used = true
	return s.tokens
}

// Release returns the token slice to the pool
func (s *ImprovedTokenSlice) Release() {
	if s.used && cap(s.tokens) <= 1024 { // Don't pool very large slices
		// Only return reasonably sized slices to the pool
		improvedTokenSlicePool.Put(s)
	}
}

// optimizedTokenizeExpressionImproved is a minimal allocation version of tokenizeExpression
func (p *Parser) optimizedTokenizeExpressionImproved(expr string, tokens *ImprovedTokenSlice, line int) {
	var inString bool
	var stringDelimiter byte
	var stringStart int
	
	// Preallocate a buffer for building tokens
	buffer := make([]byte, 0, 64)
	
	for i := 0; i < len(expr); i++ {
		c := expr[i]
		
		// Handle string literals
		if (c == '"' || c == '\'') && (i == 0 || expr[i-1] != '\\') {
			if inString && c == stringDelimiter {
				// End of string, add the string token
				tokens.AppendToken(TOKEN_STRING, expr[stringStart:i], line)
				inString = false
			} else if !inString {
				// Start of string
				inString = true
				stringDelimiter = c
				stringStart = i + 1
			}
			continue
		}
		
		// Skip chars inside strings
		if inString {
			continue
		}
		
		// Handle operators
		if isCharOperator(c) {
			// Check for two-character operators
			if i+1 < len(expr) {
				nextChar := expr[i+1]
				
				if (c == '=' && nextChar == '=') ||
					(c == '!' && nextChar == '=') ||
					(c == '>' && nextChar == '=') ||
					(c == '<' && nextChar == '=') ||
					(c == '&' && nextChar == '&') ||
					(c == '|' && nextChar == '|') ||
					(c == '?' && nextChar == '?') {
					
					// Two-char operator
					buffer = buffer[:0]
					buffer = append(buffer, c, nextChar)
					tokens.AppendToken(TOKEN_OPERATOR, string(buffer), line)
					i++
					continue
				}
			}
			
			// Single-char operator
			tokens.AppendToken(TOKEN_OPERATOR, string([]byte{c}), line)
			continue
		}
		
		// Handle punctuation
		if isCharPunctuation(c) {
			tokens.AppendToken(TOKEN_PUNCTUATION, string([]byte{c}), line)
			continue
		}
		
		// Skip whitespace
		if isCharWhitespace(c) {
			continue
		}
		
		// Handle identifiers, literals, etc.
		if isCharAlpha(c) || c == '_' {
			// Start of an identifier
			start := i
			
			// Find the end
			for i++; i < len(expr) && (isCharAlpha(expr[i]) || isCharDigit(expr[i]) || expr[i] == '_'); i++ {
			}
			
			// Extract the identifier
			identifier := expr[start:i]
			i-- // Adjust for loop increment
			
			// Add token
			tokens.AppendToken(TOKEN_NAME, identifier, line)
			continue
		}
		
		// Handle numbers
		if isCharDigit(c) || (c == '-' && i+1 < len(expr) && isCharDigit(expr[i+1])) {
			start := i
			
			// Skip negative sign if present
			if c == '-' {
				i++
			}
			
			// Find end of number
			for i++; i < len(expr) && isCharDigit(expr[i]); i++ {
			}
			
			// Check for decimal point
			if i < len(expr) && expr[i] == '.' {
				i++
				for ; i < len(expr) && isCharDigit(expr[i]); i++ {
				}
			}
			
			// Extract the number
			number := expr[start:i]
			i-- // Adjust for loop increment
			
			tokens.AppendToken(TOKEN_NUMBER, number, line)
			continue
		}
	}
}

// Helper functions to reduce allocations for character checks - inlined to avoid naming conflicts

// isCharAlpha checks if a character is alphabetic
func isCharAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isCharDigit checks if a character is a digit
func isCharDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// isCharOperator checks if a character is an operator
func isCharOperator(c byte) bool {
	return c == '=' || c == '+' || c == '-' || c == '*' || c == '/' || 
		   c == '%' || c == '&' || c == '|' || c == '^' || c == '~' || 
		   c == '<' || c == '>' || c == '!' || c == '?'
}

// isCharPunctuation checks if a character is punctuation
func isCharPunctuation(c byte) bool {
	return c == '(' || c == ')' || c == '[' || c == ']' || c == '{' || c == '}' || 
		   c == '.' || c == ',' || c == ':' || c == ';'
}

// isCharWhitespace checks if a character is whitespace
func isCharWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// improvedHtmlPreservingTokenize is a zero-allocation version of the HTML preserving tokenizer
func (p *Parser) improvedHtmlPreservingTokenize() ([]Token, error) {
	// Estimate token count based on source length
	estimatedTokens := len(p.source) / 20 // Rough estimate
	tokens := GetImprovedTokenSlice(estimatedTokens)
	defer tokens.Release()
	
	var currentPosition int
	line := 1
	
	// Reusable buffers to avoid allocations
	tagPatterns := [5]string{"{{-", "{{", "{%-", "{%", "{#"}
	tagTypes := [5]int{TOKEN_VAR_START_TRIM, TOKEN_VAR_START, TOKEN_BLOCK_START_TRIM, TOKEN_BLOCK_START, TOKEN_COMMENT_START}
	tagLengths := [5]int{3, 2, 3, 2, 2}
	
	for currentPosition < len(p.source) {
		// Find the next tag
		nextTagPos := -1
		tagType := -1
		tagLength := 0
		
		// Check for all possible tag patterns
		for i := 0; i < 5; i++ {
			pos := strings.Index(p.source[currentPosition:], tagPatterns[i])
			if pos != -1 {
				// Adjust position relative to current position
				pos += currentPosition
				
				// If this is the first tag found or it's closer than previous ones
				if nextTagPos == -1 || pos < nextTagPos {
					nextTagPos = pos
					tagType = tagTypes[i]
					tagLength = tagLengths[i]
				}
			}
		}
		
		// Check if the tag is escaped
		if nextTagPos != -1 && nextTagPos > 0 && p.source[nextTagPos-1] == '\\' {
			// Add text up to the backslash
			if nextTagPos-1 > currentPosition {
				preText := p.source[currentPosition:nextTagPos-1]
				tokens.AppendToken(TOKEN_TEXT, preText, line)
				line += countNewlines(preText)
			}
			
			// Add the tag as literal text (without the backslash)
			// Find which pattern was matched
			for i := 0; i < 5; i++ {
				if tagType == tagTypes[i] {
					tokens.AppendToken(TOKEN_TEXT, tagPatterns[i], line)
					break
				}
			}
			
			// Move past this tag
			currentPosition = nextTagPos + tagLength
			continue
		}
		
		// No more tags found - add the rest as TEXT
		if nextTagPos == -1 {
			remainingText := p.source[currentPosition:]
			if len(remainingText) > 0 {
				tokens.AppendToken(TOKEN_TEXT, remainingText, line)
				line += countNewlines(remainingText)
			}
			break
		}
		
		// Add text before the tag
		if nextTagPos > currentPosition {
			textContent := p.source[currentPosition:nextTagPos]
			tokens.AppendToken(TOKEN_TEXT, textContent, line)
			line += countNewlines(textContent)
		}
		
		// Add the tag start token
		tokens.AppendToken(tagType, "", line)
		
		// Move past opening tag
		currentPosition = nextTagPos + tagLength
		
		// Find matching end tag
		var endTag string
		var endTagType int
		var endTagLength int
		
		if tagType == TOKEN_VAR_START || tagType == TOKEN_VAR_START_TRIM {
			// Look for "}}" or "-}}"
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
				return nil, fmt.Errorf("unclosed variable tag at line %d", line)
			}
		} else if tagType == TOKEN_BLOCK_START || tagType == TOKEN_BLOCK_START_TRIM {
			// Look for "%}" or "-%}"
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
				return nil, fmt.Errorf("unclosed block tag at line %d", line)
			}
		} else if tagType == TOKEN_COMMENT_START {
			// Look for "#}"
			endPos := strings.Index(p.source[currentPosition:], "#}")
			if endPos == -1 {
				return nil, fmt.Errorf("unclosed comment at line %d", line)
			}
			endTag = "#}"
			endTagType = TOKEN_COMMENT_END
			endTagLength = 2
		}
		
		// Find position of the end tag
		endPos := strings.Index(p.source[currentPosition:], endTag)
		if endPos == -1 {
			return nil, fmt.Errorf("unclosed tag at line %d", line)
		}
		
		// Get content between tags
		tagContent := p.source[currentPosition:currentPosition+endPos]
		line += countNewlines(tagContent)
		
		// Process tag content based on type
		if tagType == TOKEN_COMMENT_START {
			// Store comments as TEXT tokens
			if len(tagContent) > 0 {
				tokens.AppendToken(TOKEN_TEXT, tagContent, line)
			}
		} else {
			// For variable and block tags, tokenize the content
			tagContent = strings.TrimSpace(tagContent)
			
			if tagType == TOKEN_BLOCK_START || tagType == TOKEN_BLOCK_START_TRIM {
				// Process block tags with optimized tokenization
				processBlockTag(tagContent, tokens, line, p)
			} else {
				// Process variable tags with optimized tokenization
				if len(tagContent) > 0 {
					if !strings.ContainsAny(tagContent, ".|[](){}\"',+-*/=!<>%&^~") {
						// Simple variable name
						tokens.AppendToken(TOKEN_NAME, tagContent, line)
					} else {
						// Complex expression
						expressionTokens := GetImprovedTokenSlice(len(tagContent) / 4)
						p.optimizedTokenizeExpressionImproved(tagContent, expressionTokens, line)
						
						// Copy tokens
						for _, token := range expressionTokens.tokens {
							tokens.AppendToken(token.Type, token.Value, token.Line)
						}
						
						expressionTokens.Release()
					}
				}
			}
		}
		
		// Add the end tag token
		tokens.AppendToken(endTagType, "", line)
		
		// Move past the end tag
		currentPosition = currentPosition + endPos + endTagLength
	}
	
	// Add EOF token
	tokens.AppendToken(TOKEN_EOF, "", line)
	
	return tokens.Finalize(), nil
}

// Helper function to process block tags
func processBlockTag(content string, tokens *ImprovedTokenSlice, line int, p *Parser) {
	// Extract the tag name
	parts := strings.SplitN(content, " ", 2)
	if len(parts) > 0 {
		blockName := parts[0]
		tokens.AppendToken(TOKEN_NAME, blockName, line)
		
		// Process rest of the block content
		if len(parts) > 1 {
			blockContent := strings.TrimSpace(parts[1])
			
			switch blockName {
			case "if", "elseif":
				// For conditional blocks, tokenize expression
				exprTokens := GetImprovedTokenSlice(len(blockContent) / 4)
				p.optimizedTokenizeExpressionImproved(blockContent, exprTokens, line)
				
				// Copy tokens
				for _, token := range exprTokens.tokens {
					tokens.AppendToken(token.Type, token.Value, token.Line)
				}
				
				exprTokens.Release()
				
			case "for":
				// Process for loop with iterator(s) and collection
				inPos := strings.Index(strings.ToLower(blockContent), " in ")
				if inPos != -1 {
					iterators := strings.TrimSpace(blockContent[:inPos])
					collection := strings.TrimSpace(blockContent[inPos+4:])
					
					// Handle key, value iterator syntax
					if strings.Contains(iterators, ",") {
						iterParts := strings.SplitN(iterators, ",", 2)
						if len(iterParts) == 2 {
							tokens.AppendToken(TOKEN_NAME, strings.TrimSpace(iterParts[0]), line)
							tokens.AppendToken(TOKEN_PUNCTUATION, ",", line)
							tokens.AppendToken(TOKEN_NAME, strings.TrimSpace(iterParts[1]), line)
						}
					} else {
						// Single iterator
						tokens.AppendToken(TOKEN_NAME, iterators, line)
					}
					
					// Add 'in' keyword
					tokens.AppendToken(TOKEN_NAME, "in", line)
					
					// Process collection expression
					collectionTokens := GetImprovedTokenSlice(len(collection) / 4)
					p.optimizedTokenizeExpressionImproved(collection, collectionTokens, line)
					
					// Copy tokens
					for _, token := range collectionTokens.tokens {
						tokens.AppendToken(token.Type, token.Value, token.Line)
					}
					
					collectionTokens.Release()
				} else {
					// Fallback for malformed for loops
					tokens.AppendToken(TOKEN_NAME, blockContent, line)
				}
				
			case "set":
				// Handle variable assignment
				assignPos := strings.Index(blockContent, "=")
				if assignPos != -1 {
					varName := strings.TrimSpace(blockContent[:assignPos])
					value := strings.TrimSpace(blockContent[assignPos+1:])
					
					tokens.AppendToken(TOKEN_NAME, varName, line)
					tokens.AppendToken(TOKEN_OPERATOR, "=", line)
					
					// Tokenize value expression
					valueTokens := GetImprovedTokenSlice(len(value) / 4)
					p.optimizedTokenizeExpressionImproved(value, valueTokens, line)
					
					// Copy tokens
					for _, token := range valueTokens.tokens {
						tokens.AppendToken(token.Type, token.Value, token.Line)
					}
					
					valueTokens.Release()
				} else {
					// Simple set without assignment
					tokens.AppendToken(TOKEN_NAME, blockContent, line)
				}
				
			default:
				// Other block types
				tokens.AppendToken(TOKEN_NAME, blockContent, line)
			}
		}
	}
}