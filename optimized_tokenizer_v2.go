package twig

import (
	"fmt"
	"sync"
)

// OptimizedTokenizerV2 is an improved tokenizer that uses the tag detector
// for faster tokenization with fewer allocations
type OptimizedTokenizerV2 struct {
	source          string
	position        int
	line            int
	tokenBuffer     []Token
	result          []Token
	tagCache        map[string]bool
	stringConstants map[string]string
}

// tokenBufferPool is a pool of token buffers to reduce allocations
var tokenBufferPool = sync.Pool{
	New: func() interface{} {
		buffer := make([]Token, 0, 256)
		return &buffer
	},
}

// OptimizedTokenizerV2Pool is a pool of tokenizers
var OptimizedTokenizerV2Pool = sync.Pool{
	New: func() interface{} {
		return &OptimizedTokenizerV2{
			tokenBuffer:     make([]Token, 0, 256),
			tagCache:        make(map[string]bool, 32),
			stringConstants: make(map[string]string, 64),
		}
	},
}

// GetOptimizedTokenizerV2 gets a tokenizer from the pool
func GetOptimizedTokenizerV2() *OptimizedTokenizerV2 {
	return OptimizedTokenizerV2Pool.Get().(*OptimizedTokenizerV2)
}

// ReleaseOptimizedTokenizerV2 returns a tokenizer to the pool
func ReleaseOptimizedTokenizerV2(t *OptimizedTokenizerV2) {
	// Clear maps but preserve capacity
	for k := range t.tagCache {
		delete(t.tagCache, k)
	}
	for k := range t.stringConstants {
		delete(t.stringConstants, k)
	}
	
	// Reset fields
	t.source = ""
	t.position = 0
	t.line = 0
	t.tokenBuffer = t.tokenBuffer[:0]
	t.result = nil
	
	// Return to pool
	OptimizedTokenizerV2Pool.Put(t)
}

// Initialize prepares the tokenizer for a new source
func (t *OptimizedTokenizerV2) Initialize(source string) {
	t.source = source
	t.position = 0
	t.line = 1
	t.tokenBuffer = t.tokenBuffer[:0]
	t.result = nil
	
	// Pre-load common string constants
	for _, s := range commonStrings {
		t.stringConstants[s] = s
	}
}

// TokenizeHtmlPreserving tokenizes HTML, preserving its structure
func (t *OptimizedTokenizerV2) TokenizeHtmlPreserving() ([]Token, error) {
	// Get a token buffer from the pool
	t.tokenBuffer = t.tokenBuffer[:0]
	
	// Track position
	pos := 0
	sourceLen := len(t.source)
	
	for pos < sourceLen {
		// Find the next tag
		tag := FindNextTag(t.source, pos)
		
		// No more tags found
		if tag.Type == TAG_NONE {
			// Add remaining text as a TOKEN_TEXT
			if pos < sourceLen {
				remainingText := t.source[pos:]
				t.addToken(TOKEN_TEXT, remainingText)
				t.line += countNewlines(remainingText)
			}
			break
		}
		
		// Check if the tag is escaped
		if tag.Position > 0 && t.source[tag.Position-1] == '\\' {
			// Add text up to the backslash
			if tag.Position-1 > pos {
				preText := t.source[pos:tag.Position-1]
				t.addToken(TOKEN_TEXT, preText)
				t.line += countNewlines(preText)
			}
			
			// Add the tag as literal text (without the backslash)
			tagText := t.source[tag.Position:tag.Position+tag.Length]
			t.addToken(TOKEN_TEXT, tagText)
			
			// Move past this tag
			pos = tag.Position + tag.Length
			continue
		}
		
		// Add text before the tag
		if tag.Position > pos {
			textContent := t.source[pos:tag.Position]
			t.addToken(TOKEN_TEXT, textContent)
			t.line += countNewlines(textContent)
		}
		
		// Add the tag start token
		tokenType := tagTypeToTokenType(tag.Type)
		t.addToken(tokenType, "")
		
		// Move past opening tag
		pos = tag.Position + tag.Length
		
		// Find matching end tag
		endPos := FindTagEnd(t.source, pos, tag.Type)
		if endPos == -1 {
			// Tag not closed
			tagName := ""
			switch tag.Type {
			case TAG_VAR, TAG_VAR_TRIM:
				tagName = "variable"
			case TAG_BLOCK, TAG_BLOCK_TRIM:
				tagName = "block"
			case TAG_COMMENT:
				tagName = "comment"
			}
			return nil, fmt.Errorf("unclosed %s tag at line %d", tagName, t.line)
		}
		
		// Get content between tags
		tagContent := t.source[pos:endPos]
		t.line += countNewlines(tagContent)
		
		// Process tag content based on type
		if tag.Type == TAG_COMMENT {
			// Store comments as TEXT tokens
			if len(tagContent) > 0 {
				t.addToken(TOKEN_TEXT, tagContent)
			}
		} else {
			// For variable and block tags, tokenize the content
			tagContent = t.internString(trimSpace(tagContent))
			
			if tag.Type == TAG_BLOCK || tag.Type == TAG_BLOCK_TRIM {
				// Process block tags with specialized tokenization
				t.tokenizeBlockTag(tagContent)
			} else {
				// Process variable tags
				if len(tagContent) > 0 {
					// Simple variable name (no operators or complex expressions)
					if !containsAny(tagContent, ".|[](){}\"',+-*/=!<>%&^~") {
						t.addToken(TOKEN_NAME, t.internString(tagContent))
					} else {
						// Complex expression - tokenize it
						t.tokenizeExpression(tagContent)
					}
				}
			}
		}
		
		// Add the end tag token
		endTokenType := tagTypeToEndTokenType(tag.Type)
		t.addToken(endTokenType, "")
		
		// Move past the end tag
		pos = endPos + 2
	}
	
	// Add EOF token
	t.addToken(TOKEN_EOF, "")
	
	// Save the token buffer to result
	t.result = t.tokenBuffer
	return t.result, nil
}

// Helper functions

// internString returns an interned version of the string
// to reduce string allocations
func (t *OptimizedTokenizerV2) internString(s string) string {
	// First check local cache
	if cached, exists := t.stringConstants[s]; exists {
		return cached
	}
	
	// Then try global cache
	interned := Intern(s)
	
	// Store in local cache if under size limit
	if len(s) <= 64 {
		t.stringConstants[interned] = interned
	}
	
	return interned
}

// addToken adds a token to the buffer
func (t *OptimizedTokenizerV2) addToken(tokenType int, value string) {
	var token Token
	token.Type = tokenType
	token.Value = value
	token.Line = t.line
	
	t.tokenBuffer = append(t.tokenBuffer, token)
}

// tokenizeBlockTag handles specialized block tag tokenization
func (t *OptimizedTokenizerV2) tokenizeBlockTag(content string) {
	// Extract the tag name
	var blockName, blockContent string
	
	// Find space to separate tag name from content
	spacePos := indexByte(content, ' ')
	if spacePos == -1 {
		// No space found, the whole content is the tag name
		blockName = content
		blockContent = ""
	} else {
		blockName = content[:spacePos]
		blockContent = trimSpace(content[spacePos+1:])
	}
	
	// Use interned string for block name
	blockName = t.internString(blockName)
	
	// Store in tag cache for future reference
	t.tagCache[blockName] = true
	
	// For simplicity and compatibility, let's use the original tokenizer for block content
	
	// Create a temporary tokenizer
	tokenizer := GetTokenizer("", 0)
	defer ReleaseTokenizer(tokenizer)
	
	// Add the block name
	t.addToken(TOKEN_NAME, blockName)
	
	// If there's no content, we're done
	if blockContent == "" {
		return
	}
	
	// Handle simple cases directly
	if !containsAny(blockContent, ".|[](){}\"',+-*/=!<>%&^~") {
		t.addToken(TOKEN_NAME, t.internString(blockContent))
		return
	}
	
	// For complex block content, use the original tokenizer
	tokens := tokenizer.TokenizeExpression(blockContent)
	
	// Copy tokens to our buffer
	for _, token := range tokens {
		// Intern string values to reduce allocations
		if token.Value != "" {
			token.Value = t.internString(token.Value)
		}
		t.tokenBuffer = append(t.tokenBuffer, token)
	}
}

// tokenizeForTag handles the special processing of for loop tags
func (t *OptimizedTokenizerV2) tokenizeForTag(content string) {
	// For simplicity in Phase 2, let's tokenize the entire for loop
	// using the original tokenizer to ensure compatibility
	
	// Tokenize "for <iterator> in <collection>"
	t.tokenizeExpression("for " + content) 
}

// tokenizeExpression tokenizes a Twig expression
func (t *OptimizedTokenizerV2) tokenizeExpression(expr string) {
	// For simplicity in Phase 2, let's fall back to the original tokenizer
	// for expressions which are more complex
	// This ensures we don't break any functionality while optimizing the common path
	
	// Create a temporary tokenizer
	tokenizer := GetTokenizer("", 0)
	defer ReleaseTokenizer(tokenizer)
	
	// Tokenize the expression
	tokens := tokenizer.TokenizeExpression(expr)
	
	// Copy tokens to our buffer
	for _, token := range tokens {
		// Intern string values to reduce allocations
		if token.Value != "" {
			token.Value = t.internString(token.Value)
		}
		t.tokenBuffer = append(t.tokenBuffer, token)
	}
}

// ApplyWhitespaceControl applies whitespace control to the tokenized result
func (t *OptimizedTokenizerV2) ApplyWhitespaceControl() {
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

// Utility functions

// tagTypeToTokenType converts a TagType to a token type
func tagTypeToTokenType(tagType TagType) int {
	switch tagType {
	case TAG_VAR:
		return TOKEN_VAR_START
	case TAG_VAR_TRIM:
		return TOKEN_VAR_START_TRIM
	case TAG_BLOCK:
		return TOKEN_BLOCK_START
	case TAG_BLOCK_TRIM:
		return TOKEN_BLOCK_START_TRIM
	case TAG_COMMENT:
		return TOKEN_COMMENT_START
	default:
		return TOKEN_TEXT
	}
}

// tagTypeToEndTokenType converts a TagType to a token end type
func tagTypeToEndTokenType(tagType TagType) int {
	switch tagType {
	case TAG_VAR:
		return TOKEN_VAR_END
	case TAG_VAR_TRIM:
		return TOKEN_VAR_END_TRIM
	case TAG_BLOCK:
		return TOKEN_BLOCK_END
	case TAG_BLOCK_TRIM:
		return TOKEN_BLOCK_END_TRIM
	case TAG_COMMENT:
		return TOKEN_COMMENT_END
	default:
		return TOKEN_TEXT
	}
}

// Optimized string utility functions to reduce allocations

// trimSpace is an allocation-free version of strings.TrimSpace for simple cases
func trimSpace(s string) string {
	// Empty string check
	if s == "" {
		return s
	}
	
	// Find first non-space character
	start := 0
	for start < len(s) && isWhitespace(s[start]) {
		start++
	}
	
	// Find last non-space character
	end := len(s) - 1
	for end >= start && isWhitespace(s[end]) {
		end--
	}
	
	// Return the trimmed string
	if start > 0 || end < len(s)-1 {
		return s[start:end+1]
	}
	
	return s
}

// lower converts a string to lowercase for case-insensitive comparisons
func lower(s string) string {
	// Quick check for empty string
	if s == "" {
		return s
	}
	
	// Check if the string needs modification
	hasUpper := false
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			hasUpper = true
			break
		}
	}
	
	// If no uppercase, return original
	if !hasUpper {
		return s
	}
	
	// Create a lowercase copy
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		} else {
			b[i] = c
		}
	}
	
	return string(b)
}

// indexByte finds the first occurrence of a byte in a string
func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// hasSubstring checks if a string contains a substring
func hasSubstring(s, substr string) bool {
	return indexOf(s, substr) >= 0
}

// containsAny checks if a string contains any of the given characters
func containsAny(s, chars string) bool {
	for i := 0; i < len(s); i++ {
		for j := 0; j < len(chars); j++ {
			if s[i] == chars[j] {
				return true
			}
		}
	}
	return false
}

// split splits a string into substrings by separator with a maximum count
func split(s, sep string, count int) []string {
	// Check for empty string or separator
	if s == "" {
		return []string{s}
	}
	if sep == "" {
		return []string{s}
	}
	
	// Find the separators
	var parts []string
	start := 0
	for i := 0; i <= len(s)-len(sep) && (count == 0 || len(parts) < count-1); i++ {
		found := true
		for j := 0; j < len(sep); j++ {
			if s[i+j] != sep[j] {
				found = false
				break
			}
		}
		
		if found {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i = start - 1 // Adjust i for the next iteration
		}
	}
	
	// Add the final part
	parts = append(parts, s[start:])
	
	return parts
}