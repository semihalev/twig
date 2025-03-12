package twig

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

// ZeroAllocTokenizer is an allocation-free tokenizer
// It uses a pre-allocated token buffer for all token operations
type ZeroAllocTokenizer struct {
	tokenBuffer []Token      // Pre-allocated buffer of tokens
	source      string       // Source string being tokenized
	position    int          // Current position in source
	line        int          // Current line
	result      []Token      // Slice of actually used tokens
	tempStrings []string     // String constants that we can reuse
}

// This array contains commonly used strings in tokenization to avoid allocations
var commonStrings = []string{
	// Common twig words and operators
	"if", "else", "elseif", "endif", "for", "endfor", "in",
	"block", "endblock", "extends", "include", "with", "set",
	"macro", "endmacro", "import", "from", "as", "do",
	
	// Common operators
	"+", "-", "*", "/", "=", "==", "!=", ">", "<", ">=", "<=",
	"and", "or", "not", "~", "%", "?", ":", "??",
	
	// Common punctuation
	"(", ")", "[", "]", "{", "}", ".", ",", "|", ";",
	
	// Common literals
	"true", "false", "null",
	
	// Empty string
	"",
}

// TokenizerPooled holds a set of resources for zero-allocation tokenization
type TokenizerPooled struct {
	tokenizer ZeroAllocTokenizer
	used      bool
}

// TokenizerPool is a pool of tokenizer resources
var tokenizerPool = sync.Pool{
	New: func() interface{} {
		// Create a pre-allocated tokenizer with reasonable defaults
		return &TokenizerPooled{
			tokenizer: ZeroAllocTokenizer{
				tokenBuffer: make([]Token, 0, 256),  // Buffer for tokens
				tempStrings: append([]string{}, commonStrings...),
				result:      nil,
			},
			used: false,
		}
	},
}

// GetTokenizer gets a tokenizer from the pool
func GetTokenizer(source string, capacityHint int) *ZeroAllocTokenizer {
	pooled := tokenizerPool.Get().(*TokenizerPooled)
	
	// Reset the tokenizer
	tokenizer := &pooled.tokenizer
	tokenizer.source = source
	tokenizer.position = 0
	tokenizer.line = 1
	
	// Ensure token buffer has enough capacity
	neededCapacity := capacityHint
	if neededCapacity <= 0 {
		// Estimate capacity based on source length
		neededCapacity = len(source) / 10
		if neededCapacity < 32 {
			neededCapacity = 32
		}
	}
	
	// Resize token buffer if needed
	if cap(tokenizer.tokenBuffer) < neededCapacity {
		tokenizer.tokenBuffer = make([]Token, 0, neededCapacity)
	} else {
		tokenizer.tokenBuffer = tokenizer.tokenBuffer[:0]
	}
	
	// Reset result
	tokenizer.result = nil
	
	// Mark as used
	pooled.used = true
	
	return tokenizer
}

// ReleaseTokenizer returns a tokenizer to the pool
func ReleaseTokenizer(tokenizer *ZeroAllocTokenizer) {
	// Get the parent pooled struct
	pooled := (*TokenizerPooled)(unsafe.Pointer(
		uintptr(unsafe.Pointer(tokenizer)) - unsafe.Offsetof(TokenizerPooled{}.tokenizer)))
	
	// Only return to pool if it's used
	if pooled.used {
		// Mark as not used and clear references that might prevent GC
		pooled.used = false
		tokenizer.source = ""
		tokenizer.result = nil
		
		// Return to pool
		tokenizerPool.Put(pooled)
	}
}

// AddToken adds a token to the buffer
func (t *ZeroAllocTokenizer) AddToken(tokenType int, value string, line int) {
	// Create a token
	var token Token
	token.Type = tokenType
	token.Value = value
	token.Line = line
	
	// Add to buffer
	t.tokenBuffer = append(t.tokenBuffer, token)
}

// GetStringConstant checks if a string exists in our constants and returns
// the canonical version to avoid allocation
func (t *ZeroAllocTokenizer) GetStringConstant(s string) string {
	// First check common strings
	for _, constant := range t.tempStrings {
		if constant == s {
			return constant
		}
	}
	
	// Add to temp strings if it's a short string that might be reused
	if len(s) <= 20 {
		t.tempStrings = append(t.tempStrings, s)
	}
	
	return s
}

// TokenizeExpression tokenizes an expression string with zero allocations
func (t *ZeroAllocTokenizer) TokenizeExpression(expr string) []Token {
	// Save current position and set new source context
	savedSource := t.source
	savedPosition := t.position
	savedLine := t.line
	
	t.source = expr
	t.position = 0
	startTokenCount := len(t.tokenBuffer)
	
	var inString bool
	var stringDelimiter byte
	var stringStart int
	
	for t.position < len(t.source) {
		c := t.source[t.position]
		
		// Handle string literals
		if (c == '"' || c == '\'') && (t.position == 0 || t.source[t.position-1] != '\\') {
			if inString && c == stringDelimiter {
				// End of string, add the string token
				value := t.source[stringStart:t.position] 
				t.AddToken(TOKEN_STRING, value, t.line)
				inString = false
			} else if !inString {
				// Start of string
				inString = true
				stringDelimiter = c
				stringStart = t.position + 1
			}
			t.position++
			continue
		}
		
		// Skip chars inside strings
		if inString {
			t.position++
			continue
		}
		
		// Handle operators (includes multi-char operators like ==, !=, etc.)
		if isOperator(c) {
			op := string(c)
			t.position++
			
			// Check for two-character operators
			if t.position < len(t.source) {
				nextChar := t.source[t.position]
				twoCharOp := string([]byte{c, nextChar})
				
				// Check common two-char operators
				if (c == '=' && nextChar == '=') ||
					(c == '!' && nextChar == '=') ||
					(c == '>' && nextChar == '=') ||
					(c == '<' && nextChar == '=') ||
					(c == '&' && nextChar == '&') ||
					(c == '|' && nextChar == '|') ||
					(c == '?' && nextChar == '?') {
					
					op = twoCharOp
					t.position++
				}
			}
			
			// Use constant version of the operator string if possible
			op = t.GetStringConstant(op)
			t.AddToken(TOKEN_OPERATOR, op, t.line)
			continue
		}
		
		// Handle punctuation
		if isPunctuation(c) {
			// Use constant version of punctuation
			punct := t.GetStringConstant(string(c))
			t.AddToken(TOKEN_PUNCTUATION, punct, t.line)
			t.position++
			continue
		}
		
		// Skip whitespace
		if isWhitespace(c) {
			t.position++
			if c == '\n' {
				t.line++
			}
			continue
		}
		
		// Handle identifiers, literals, etc.
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			// Start of an identifier
			start := t.position
			
			// Find the end
			t.position++
			for t.position < len(t.source) && 
				((t.source[t.position] >= 'a' && t.source[t.position] <= 'z') || 
				 (t.source[t.position] >= 'A' && t.source[t.position] <= 'Z') || 
				 (t.source[t.position] >= '0' && t.source[t.position] <= '9') || 
				 t.source[t.position] == '_') {
				t.position++
			}
			
			// Extract the identifier
			identifier := t.source[start:t.position]
			
			// Try to use a canonical string
			identifier = t.GetStringConstant(identifier)
			
			// Keywords/literals get special token types
			if identifier == "true" || identifier == "false" || identifier == "null" {
				t.AddToken(TOKEN_NAME, identifier, t.line)
			} else {
				t.AddToken(TOKEN_NAME, identifier, t.line)
			}
			
			continue
		}
		
		// Handle numbers
		if (c >= '0' && c <= '9') || (c == '-' && t.position+1 < len(t.source) && t.source[t.position+1] >= '0' && t.source[t.position+1] <= '9') {
			start := t.position
			
			// Skip the negative sign if present
			if c == '-' {
				t.position++
			}
			
			// Consume digits
			for t.position < len(t.source) && t.source[t.position] >= '0' && t.source[t.position] <= '9' {
				t.position++
			}
			
			// Handle decimal point
			if t.position < len(t.source) && t.source[t.position] == '.' {
				t.position++
				
				// Consume fractional digits
				for t.position < len(t.source) && t.source[t.position] >= '0' && t.source[t.position] <= '9' {
					t.position++
				}
			}
			
			// Add the number token
			t.AddToken(TOKEN_NUMBER, t.source[start:t.position], t.line)
			continue
		}
		
		// Unrecognized character
		t.position++
	}
	
	// Create slice of tokens
	tokens := t.tokenBuffer[startTokenCount:]
	
	// Restore original context
	t.source = savedSource
	t.position = savedPosition
	t.line = savedLine
	
	return tokens
}

// TokenizeHtmlPreserving performs full tokenization of a template with HTML preservation
func (t *ZeroAllocTokenizer) TokenizeHtmlPreserving() ([]Token, error) {
	// Reset position and line
	t.position = 0
	t.line = 1
	
	// Clear token buffer
	t.tokenBuffer = t.tokenBuffer[:0]
	
	tagPatterns := [5]string{"{{-", "{{", "{%-", "{%", "{#"}
	tagTypes := [5]int{TOKEN_VAR_START_TRIM, TOKEN_VAR_START, TOKEN_BLOCK_START_TRIM, TOKEN_BLOCK_START, TOKEN_COMMENT_START}
	tagLengths := [5]int{3, 2, 3, 2, 2}
	
	for t.position < len(t.source) {
		// Find the next tag
		nextTagPos := -1
		tagType := -1
		tagLength := 0
		
		// Check for all possible tag patterns
		// This loop avoids allocations by manually checking prefixes
		remainingSource := t.source[t.position:]
		for i := 0; i < 5; i++ {
			pattern := tagPatterns[i]
			if len(remainingSource) >= len(pattern) && 
			   remainingSource[:len(pattern)] == pattern {
				// Tag found at current position
				nextTagPos = t.position
				tagType = tagTypes[i]
				tagLength = tagLengths[i]
				break
			}
			
			// If not found at current position, find it in the remainder
			patternPos := strings.Index(remainingSource, pattern)
			if patternPos != -1 {
				pos := t.position + patternPos
				if nextTagPos == -1 || pos < nextTagPos {
					nextTagPos = pos
					tagType = tagTypes[i]
					tagLength = tagLengths[i]
				}
			}
		}
		
		// Check if the tag is escaped
		if nextTagPos != -1 && nextTagPos > 0 && t.source[nextTagPos-1] == '\\' {
			// Add text up to the backslash
			if nextTagPos-1 > t.position {
				preText := t.source[t.position:nextTagPos-1]
				t.AddToken(TOKEN_TEXT, preText, t.line)
				t.line += countNewlines(preText)
			}
			
			// Add the tag as literal text (without the backslash)
			// Find which pattern was matched
			for i := 0; i < 5; i++ {
				if tagType == tagTypes[i] {
					t.AddToken(TOKEN_TEXT, tagPatterns[i], t.line)
					break
				}
			}
			
			// Move past this tag
			t.position = nextTagPos + tagLength
			continue
		}
		
		// No more tags found - add the rest as TEXT
		if nextTagPos == -1 {
			if t.position < len(t.source) {
				remainingText := t.source[t.position:]
				t.AddToken(TOKEN_TEXT, remainingText, t.line)
				t.line += countNewlines(remainingText)
			}
			break
		}
		
		// Add text before the tag
		if nextTagPos > t.position {
			textContent := t.source[t.position:nextTagPos]
			t.AddToken(TOKEN_TEXT, textContent, t.line)
			t.line += countNewlines(textContent)
		}
		
		// Add the tag start token
		t.AddToken(tagType, "", t.line)
		
		// Move past opening tag
		t.position = nextTagPos + tagLength
		
		// Find matching end tag
		var endTag string
		var endTagType int
		var endTagLength int
		
		if tagType == TOKEN_VAR_START || tagType == TOKEN_VAR_START_TRIM {
			// Look for "}}" or "-}}"
			endPos1 := strings.Index(t.source[t.position:], "}}")
			endPos2 := strings.Index(t.source[t.position:], "-}}")
			
			if endPos1 != -1 && (endPos2 == -1 || endPos1 < endPos2) {
				endTag = "}}"
				endTagType = TOKEN_VAR_END
				endTagLength = 2
			} else if endPos2 != -1 {
				endTag = "-}}"
				endTagType = TOKEN_VAR_END_TRIM
				endTagLength = 3
			} else {
				return nil, fmt.Errorf("unclosed variable tag at line %d", t.line)
			}
		} else if tagType == TOKEN_BLOCK_START || tagType == TOKEN_BLOCK_START_TRIM {
			// Look for "%}" or "-%}"
			endPos1 := strings.Index(t.source[t.position:], "%}")
			endPos2 := strings.Index(t.source[t.position:], "-%}")
			
			if endPos1 != -1 && (endPos2 == -1 || endPos1 < endPos2) {
				endTag = "%}"
				endTagType = TOKEN_BLOCK_END
				endTagLength = 2
			} else if endPos2 != -1 {
				endTag = "-%}"
				endTagType = TOKEN_BLOCK_END_TRIM
				endTagLength = 3
			} else {
				return nil, fmt.Errorf("unclosed block tag at line %d", t.line)
			}
		} else if tagType == TOKEN_COMMENT_START {
			// Look for "#}"
			endPos := strings.Index(t.source[t.position:], "#}")
			if endPos == -1 {
				return nil, fmt.Errorf("unclosed comment at line %d", t.line)
			}
			endTag = "#}"
			endTagType = TOKEN_COMMENT_END
			endTagLength = 2
		}
		
		// Find position of the end tag
		endPos := strings.Index(t.source[t.position:], endTag)
		if endPos == -1 {
			return nil, fmt.Errorf("unclosed tag at line %d", t.line)
		}
		
		// Get content between tags
		tagContent := t.source[t.position:t.position+endPos]
		t.line += countNewlines(tagContent)
		
		// Process tag content based on type
		if tagType == TOKEN_COMMENT_START {
			// Store comments as TEXT tokens
			if len(tagContent) > 0 {
				t.AddToken(TOKEN_TEXT, tagContent, t.line)
			}
		} else {
			// For variable and block tags, tokenize the content
			tagContent = strings.TrimSpace(tagContent)
			
			if tagType == TOKEN_BLOCK_START || tagType == TOKEN_BLOCK_START_TRIM {
				// Process block tags with specialized tokenization
				t.processBlockTag(tagContent)
			} else {
				// Process variable tags with optimized tokenization
				if len(tagContent) > 0 {
					if !strings.ContainsAny(tagContent, ".|[](){}\"',+-*/=!<>%&^~") {
						// Simple variable name
						identifier := t.GetStringConstant(tagContent)
						t.AddToken(TOKEN_NAME, identifier, t.line)
					} else {
						// Complex expression
						t.TokenizeExpression(tagContent)
					}
				}
			}
		}
		
		// Add the end tag token
		t.AddToken(endTagType, "", t.line)
		
		// Move past the end tag
		t.position = t.position + endPos + endTagLength
	}
	
	// Add EOF token
	t.AddToken(TOKEN_EOF, "", t.line)
	
	// Save the token buffer to result
	t.result = t.tokenBuffer
	return t.result, nil
}

// processBlockTag handles specialized block tag tokenization
func (t *ZeroAllocTokenizer) processBlockTag(content string) {
	// Extract the tag name
	spacePos := strings.IndexByte(content, ' ')
	var blockName string
	var blockContent string
	
	if spacePos == -1 {
		// No space found, the whole content is the tag name
		blockName = content
		blockContent = ""
	} else {
		blockName = content[:spacePos]
		blockContent = strings.TrimSpace(content[spacePos+1:])
	}
	
	// Use canonical string for block name
	blockName = t.GetStringConstant(blockName)
	t.AddToken(TOKEN_NAME, blockName, t.line)
	
	// If there's no content, we're done
	if blockContent == "" {
		return
	}
	
	// Process based on block type
	switch blockName {
	case "if", "elseif":
		// For conditional blocks, tokenize expression
		t.TokenizeExpression(blockContent)
		
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
					// Process iterator variables
					keyVar := t.GetStringConstant(strings.TrimSpace(iterParts[0]))
					valueVar := t.GetStringConstant(strings.TrimSpace(iterParts[1]))
					
					t.AddToken(TOKEN_NAME, keyVar, t.line)
					t.AddToken(TOKEN_PUNCTUATION, ",", t.line)
					t.AddToken(TOKEN_NAME, valueVar, t.line)
				}
			} else {
				// Single iterator
				iterator := t.GetStringConstant(iterators)
				t.AddToken(TOKEN_NAME, iterator, t.line)
			}
			
			// Add 'in' keyword
			t.AddToken(TOKEN_NAME, "in", t.line)
			
			// Process collection expression
			t.TokenizeExpression(collection)
		} else {
			// Fallback for malformed for loops
			t.AddToken(TOKEN_NAME, blockContent, t.line)
		}
		
	case "set":
		// Handle variable assignment
		assignPos := strings.Index(blockContent, "=")
		if assignPos != -1 {
			varName := strings.TrimSpace(blockContent[:assignPos])
			value := strings.TrimSpace(blockContent[assignPos+1:])
			
			// Add the variable name token
			varName = t.GetStringConstant(varName)
			t.AddToken(TOKEN_NAME, varName, t.line)
			
			// Add the assignment operator
			t.AddToken(TOKEN_OPERATOR, "=", t.line)
			
			// Tokenize the value expression
			t.TokenizeExpression(value)
		} else {
			// Simple set without assignment
			blockContent = t.GetStringConstant(blockContent)
			t.AddToken(TOKEN_NAME, blockContent, t.line)
		}
		
	case "do":
		// Handle variable assignment similar to set tag
		assignPos := strings.Index(blockContent, "=")
		if assignPos != -1 {
			varName := strings.TrimSpace(blockContent[:assignPos])
			value := strings.TrimSpace(blockContent[assignPos+1:])
			
			// Check if varName is valid (should be a variable name)
			// In Twig, variable names must start with a letter or underscore
			if len(varName) > 0 && (isCharAlpha(varName[0]) || varName[0] == '_') {
				// Add the variable name token
				varName = t.GetStringConstant(varName)
				t.AddToken(TOKEN_NAME, varName, t.line)
				
				// Add the assignment operator
				t.AddToken(TOKEN_OPERATOR, "=", t.line)
				
				// Tokenize the value expression
				if len(value) > 0 {
					t.TokenizeExpression(value)
				} else {
					// Empty value after =, which is invalid
					// Add an error token to trigger proper parser error
					t.AddToken(TOKEN_EOF, "ERROR_MISSING_VALUE", t.line)
				}
			} else {
				// Invalid variable name (like a number or operator)
				// Just tokenize as expressions to produce an error in the parser
				t.TokenizeExpression(varName)
				t.AddToken(TOKEN_OPERATOR, "=", t.line)
				t.TokenizeExpression(value)
			}
		} else {
			// No assignment, just an expression to evaluate
			t.TokenizeExpression(blockContent)
		}
		
	case "include":
		// Handle include with template path and optional context
		withPos := strings.Index(strings.ToLower(blockContent), " with ")
		if withPos != -1 {
			templatePath := strings.TrimSpace(blockContent[:withPos])
			contextExpr := strings.TrimSpace(blockContent[withPos+6:])
			
			// Process template path
			t.tokenizeTemplatePath(templatePath)
			
			// Add 'with' keyword
			t.AddToken(TOKEN_NAME, "with", t.line)
			
			// Process context expression as object
			if strings.HasPrefix(contextExpr, "{") && strings.HasSuffix(contextExpr, "}") {
				// Context is an object literal
				t.AddToken(TOKEN_PUNCTUATION, "{", t.line)
				objectContent := contextExpr[1:len(contextExpr)-1]
				t.tokenizeObjectContents(objectContent)
				t.AddToken(TOKEN_PUNCTUATION, "}", t.line)
			} else {
				// Context is a variable or expression
				t.TokenizeExpression(contextExpr)
			}
		} else {
			// Just a template path
			t.tokenizeTemplatePath(blockContent)
		}
		
	case "extends":
		// Handle extends tag (similar to include template path)
		t.tokenizeTemplatePath(blockContent)
		
	case "from":
		// Handle from tag which has a special format:
		// {% from "template.twig" import macro1, macro2 as alias %}
		importPos := strings.Index(strings.ToLower(blockContent), " import ")
		if importPos != -1 {
			// Extract template path and macros list
			templatePath := strings.TrimSpace(blockContent[:importPos])
			macrosStr := strings.TrimSpace(blockContent[importPos+8:]) // 8 = len(" import ")
			
			// Process template path
			t.tokenizeTemplatePath(templatePath)
			
			// Add 'import' keyword
			t.AddToken(TOKEN_NAME, "import", t.line)
			
			// Process macro imports
			macros := strings.Split(macrosStr, ",")
			for i, macro := range macros {
				macro = strings.TrimSpace(macro)
				
				// Check for "as" alias
				asPos := strings.Index(strings.ToLower(macro), " as ")
				if asPos != -1 {
					// Extract macro name and alias
					macroName := strings.TrimSpace(macro[:asPos])
					alias := strings.TrimSpace(macro[asPos+4:])
					
					// Add macro name
					macroName = t.GetStringConstant(macroName)
					t.AddToken(TOKEN_NAME, macroName, t.line)
					
					// Add 'as' keyword
					t.AddToken(TOKEN_NAME, "as", t.line)
					
					// Add alias
					alias = t.GetStringConstant(alias)
					t.AddToken(TOKEN_NAME, alias, t.line)
				} else {
					// Just the macro name
					macro = t.GetStringConstant(macro)
					t.AddToken(TOKEN_NAME, macro, t.line)
				}
				
				// Add comma if not the last macro
				if i < len(macros)-1 {
					t.AddToken(TOKEN_PUNCTUATION, ",", t.line)
				}
			}
		} else {
			// Malformed from tag, just tokenize as expression
			t.TokenizeExpression(blockContent)
		}
		
	case "import":
		// Handle import tag which allows importing entire templates
		// {% import "template.twig" as alias %}
		asPos := strings.Index(strings.ToLower(blockContent), " as ")
		if asPos != -1 {
			// Extract template path and alias
			templatePath := strings.TrimSpace(blockContent[:asPos])
			alias := strings.TrimSpace(blockContent[asPos+4:])
			
			// Process template path
			t.tokenizeTemplatePath(templatePath)
			
			// Add 'as' keyword
			t.AddToken(TOKEN_NAME, "as", t.line)
			
			// Add alias
			alias = t.GetStringConstant(alias)
			t.AddToken(TOKEN_NAME, alias, t.line)
		} else {
			// Simple import without alias
			t.TokenizeExpression(blockContent)
		}
		
	default:
		// Other block types - tokenize as expression
		t.TokenizeExpression(blockContent)
	}
}

// Helper methods for specialized tag tokenization

// tokenizeTemplatePath handles template paths in extends/include tags
func (t *ZeroAllocTokenizer) tokenizeTemplatePath(path string) {
	path = strings.TrimSpace(path)
	
	// If it's a quoted string
	if (strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"")) ||
	   (strings.HasPrefix(path, "'") && strings.HasSuffix(path, "'")) {
		// Extract content without quotes
		content := path[1:len(path)-1]
		t.AddToken(TOKEN_STRING, content, t.line)
	} else {
		// Otherwise tokenize as expression
		t.TokenizeExpression(path)
	}
}

// isCharAlpha checks if a byte is an alphabetic character
func isCharAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// tokenizeObjectContents handles object literal contents 
func (t *ZeroAllocTokenizer) tokenizeObjectContents(content string) {
	// Track state for nested structures
	inString := false
	stringDelim := byte(0)
	inObject := 0
	inArray := 0
	
	start := 0
	colonPos := -1
	
	for i := 0; i <= len(content); i++ {
		// At end of string or at a comma at the top level
		atEnd := i == len(content)
		isComma := !atEnd && content[i] == ','
		
		// Process key-value pair when we find a comma or reach the end
		if (isComma || atEnd) && inObject == 0 && inArray == 0 && !inString {
			if colonPos != -1 {
				// We have a key-value pair
				keyStr := strings.TrimSpace(content[start:colonPos])
				valueStr := strings.TrimSpace(content[colonPos+1:i])
				
				// Process key
				if (len(keyStr) >= 2 && keyStr[0] == '"' && keyStr[len(keyStr)-1] == '"') ||
					(len(keyStr) >= 2 && keyStr[0] == '\'' && keyStr[len(keyStr)-1] == '\'') {
					// Quoted key
					t.AddToken(TOKEN_STRING, keyStr[1:len(keyStr)-1], t.line)
				} else {
					// Unquoted key
					keyStr = t.GetStringConstant(keyStr)
					t.AddToken(TOKEN_NAME, keyStr, t.line)
				}
				
				// Add colon
				t.AddToken(TOKEN_PUNCTUATION, ":", t.line)
				
				// Process value
				t.TokenizeExpression(valueStr)
				
				// Add comma if needed
				if isComma && i < len(content)-1 {
					t.AddToken(TOKEN_PUNCTUATION, ",", t.line)
				}
				
				// Reset for next pair
				start = i + 1
				colonPos = -1
			}
			continue
		}
		
		// Skip end of string case
		if atEnd {
			continue
		}
		
		// Current character
		c := content[i]
		
		// Handle string literals
		if (c == '"' || c == '\'') && (i == 0 || content[i-1] != '\\') {
			if inString && c == stringDelim {
				inString = false
			} else if !inString {
				inString = true
				stringDelim = c
			}
			continue
		}
		
		// Skip processing inside strings
		if inString {
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
		
		// Track colon position for key-value separator
		if c == ':' && inObject == 0 && inArray == 0 && colonPos == -1 {
			colonPos = i
		}
	}
}

// ApplyWhitespaceControl applies whitespace control to the tokenized result
func (t *ZeroAllocTokenizer) ApplyWhitespaceControl() {
	tokens := t.result
	
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		
		// Handle opening tags that trim whitespace before them
		if token.Type == TOKEN_VAR_START_TRIM || token.Type == TOKEN_BLOCK_START_TRIM {
			// If there's a text token before this, trim its trailing whitespace
			if i > 0 && tokens[i-1].Type == TOKEN_TEXT {
				tokens[i-1].Value = trimTrailingWhitespace(tokens[i-1].Value)
			}
		}
		
		// Handle closing tags that trim whitespace after them
		if token.Type == TOKEN_VAR_END_TRIM || token.Type == TOKEN_BLOCK_END_TRIM {
			// If there's a text token after this, trim its leading whitespace
			if i+1 < len(tokens) && tokens[i+1].Type == TOKEN_TEXT {
				tokens[i+1].Value = trimLeadingWhitespace(tokens[i+1].Value)
			}
		}
	}
}