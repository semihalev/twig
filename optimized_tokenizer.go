package twig

import (
	"strings"
	"sync"
)

// OptimizedTokenizer implements a tokenizer that uses the global string cache
// for zero-allocation string interning
type OptimizedTokenizer struct {
	// Use the underlying tokenizer methods but intern strings
	baseTokenizer ZeroAllocTokenizer
	// Local cache of whether a string is a tag name
	tagCache map[string]bool
}

// optimizedTokenizerPool is a sync.Pool for OptimizedTokenizer objects
var optimizedTokenizerPool = sync.Pool{
	New: func() interface{} {
		return &OptimizedTokenizer{
			tagCache: make(map[string]bool, 32), // Pre-allocate with reasonable capacity
		}
	},
}

// NewOptimizedTokenizer gets an OptimizedTokenizer from the pool
func NewOptimizedTokenizer() *OptimizedTokenizer {
	return optimizedTokenizerPool.Get().(*OptimizedTokenizer)
}

// ReleaseOptimizedTokenizer returns an OptimizedTokenizer to the pool
func ReleaseOptimizedTokenizer(t *OptimizedTokenizer) {
	// Clear map but preserve capacity
	for k := range t.tagCache {
		delete(t.tagCache, k)
	}
	
	// Return to pool
	optimizedTokenizerPool.Put(t)
}

// TokenizeHtmlPreserving tokenizes HTML, preserving its structure
func (t *OptimizedTokenizer) TokenizeHtmlPreserving() ([]Token, error) {
	// Use the base tokenizer for complex operations
	tokens, err := t.baseTokenizer.TokenizeHtmlPreserving()
	if err != nil {
		return nil, err
	}
	
	// Optimize token strings by interning
	for i := range tokens {
		// Intern the value field of each token
		if tokens[i].Value != "" {
			tokens[i].Value = Intern(tokens[i].Value)
		}
		
		// For tag names, intern them as well
		if tokens[i].Type == TOKEN_BLOCK_START || tokens[i].Type == TOKEN_BLOCK_START_TRIM ||
		   tokens[i].Type == TOKEN_VAR_START || tokens[i].Type == TOKEN_VAR_START_TRIM {
			// Skip processing as these tokens don't have values
			continue
		}
		
		// Process tag names - these will be TOKEN_NAME after a block start token
		if i > 0 && tokens[i].Type == TOKEN_NAME && 
		   (tokens[i-1].Type == TOKEN_BLOCK_START || tokens[i-1].Type == TOKEN_BLOCK_START_TRIM) {
			// Intern the tag name
			tokens[i].Value = Intern(tokens[i].Value)
			
			// Cache whether this is a tag
			t.tagCache[tokens[i].Value] = true
		}
	}
	
	return tokens, nil
}

// TokenizeExpression tokenizes a Twig expression
func (t *OptimizedTokenizer) TokenizeExpression(expression string) []Token {
	// Use the base tokenizer for complex operations
	tokens := t.baseTokenizer.TokenizeExpression(expression)
	
	// Optimize token strings by interning
	for i := range tokens {
		if tokens[i].Value != "" {
			tokens[i].Value = Intern(tokens[i].Value)
		}
	}
	
	return tokens
}

// ApplyWhitespaceControl applies whitespace control for trimming tokens
func (t *OptimizedTokenizer) ApplyWhitespaceControl() {
	t.baseTokenizer.ApplyWhitespaceControl()
}

// Helper to extract tag name from a token value
func extractTagName(value string) string {
	value = strings.TrimSpace(value)
	space := strings.IndexByte(value, ' ')
	if space >= 0 {
		return value[:space]
	}
	return value
}

// IsTag checks if a string is a known tag name (cached)
func (t *OptimizedTokenizer) IsTag(name string) bool {
	// Fast path for common tags
	switch name {
	case stringIf, stringFor, stringEnd, stringEndif, stringEndfor, 
	     stringElse, stringBlock, stringSet, stringInclude, stringExtends:
		return true
	}
	
	// Check the local cache
	if isTag, exists := t.tagCache[name]; exists {
		return isTag
	}
	
	// Fall back to the base tokenizer's logic
	return false
}