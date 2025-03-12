package twig

import (
	"sync"
)

// This file implements optimized token handling functions to reduce allocations
// during the tokenization process.

// PooledToken represents a token from the token pool
// We use a separate struct to avoid accidentally returning the same instance
type PooledToken struct {
	token *Token // Reference to the token from the pool
}

// PooledTokenSlice is a slice of tokens with a reference to the original pooled slice
type PooledTokenSlice struct {
	tokens  []Token     // The token slice
	poolRef *[]Token    // Reference to the original slice from the pool
	used    bool        // Whether this slice has been used
	tmpPool sync.Pool   // Pool for temporary token objects
	scratch []*Token    // Scratch space for temporary tokens
}

// GetPooledTokenSlice gets a token slice from the pool with the given capacity hint
func GetPooledTokenSlice(capacityHint int) *PooledTokenSlice {
	slice := &PooledTokenSlice{
		tmpPool: sync.Pool{
			New: func() interface{} {
				return &Token{}
			},
		},
		scratch: make([]*Token, 0, 16), // Pre-allocate scratch space
		used:    false,
	}

	// Get a token slice from the pool
	pooledSlice := GetTokenSlice(capacityHint)
	slice.tokens = pooledSlice
	slice.poolRef = &pooledSlice

	return slice
}

// AppendToken adds a token to the slice using pooled tokens
func (s *PooledTokenSlice) AppendToken(tokenType int, value string, line int) {
	if s.used {
		// This slice has already been finalized, can't append anymore
		return
	}

	// Get a token from the pool
	token := s.tmpPool.Get().(*Token)
	token.Type = tokenType
	token.Value = value
	token.Line = line

	// Keep a reference to this token so we can clean it up later
	s.scratch = append(s.scratch, token)

	// Add a copy of the token to the slice
	s.tokens = append(s.tokens, *token)
}

// Finalize returns the token slice and cleans up temporary tokens
func (s *PooledTokenSlice) Finalize() []Token {
	if s.used {
		// Already finalized
		return s.tokens
	}

	// Mark as used so we don't accidentally use it again
	s.used = true

	// Clean up temporary tokens
	for _, token := range s.scratch {
		token.Value = ""
		s.tmpPool.Put(token)
	}

	// Clear scratch slice but keep capacity
	s.scratch = s.scratch[:0]

	return s.tokens
}

// Release returns the token slice to the pool
func (s *PooledTokenSlice) Release() {
	if s.poolRef != nil {
		ReleaseTokenSlice(*s.poolRef)
		s.poolRef = nil
	}

	// Clean up any remaining temporary tokens
	for _, token := range s.scratch {
		token.Value = ""
		s.tmpPool.Put(token)
	}

	// Clear references
	s.scratch = nil
	s.tokens = nil
	s.used = true
}

// getPooledToken gets a token from the pool (for internal use)
func getPooledToken() *Token {
	return TokenPool.Get().(*Token)
}

// releasePooledToken returns a token to the pool (for internal use)
func releasePooledToken(token *Token) {
	if token == nil {
		return
	}
	token.Value = ""
	TokenPool.Put(token)
}

// TOKEN SLICES - additional optimization for token slice reuse

// TokenNodePool provides a pool for pre-sized token node arrays
var TokenNodePool = sync.Pool{
	New: func() interface{} {
		// Default capacity that covers most cases
		slice := make([]Node, 0, 32)
		return &slice
	},
}

// GetTokenNodeSlice gets a slice of Node from the pool
func GetTokenNodeSlice(capacityHint int) *[]Node {
	slice := TokenNodePool.Get().(*[]Node)
	
	// If the capacity is too small, allocate a new slice
	if cap(*slice) < capacityHint {
		*slice = make([]Node, 0, capacityHint)
	} else {
		// Otherwise, clear the slice but keep capacity
		*slice = (*slice)[:0]
	}
	
	return slice
}

// ReleaseTokenNodeSlice returns a slice of Node to the pool
func ReleaseTokenNodeSlice(slice *[]Node) {
	if slice == nil {
		return
	}
	
	// Only pool reasonably sized slices
	if cap(*slice) > 1000 || cap(*slice) < 32 {
		return
	}
	
	// Clear references to help GC
	for i := range *slice {
		(*slice)[i] = nil
	}
	
	// Clear slice but keep capacity
	*slice = (*slice)[:0]
	TokenNodePool.Put(slice)
}