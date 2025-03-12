package twig

import (
	"sync"
)

const (
	// Common HTML/Twig strings to pre-cache
	maxCacheableLength = 64 // Only cache strings shorter than this to avoid memory bloat
	
	// Common HTML tags
	stringDiv    = "div"
	stringSpan   = "span"
	stringP      = "p"
	stringA      = "a"
	stringImg    = "img"
	stringHref   = "href"
	stringClass  = "class"
	stringId     = "id"
	stringStyle  = "style"
	
	// Common Twig syntax
	stringIf     = "if"
	stringFor    = "for"
	stringEnd    = "end"
	stringEndif  = "endif"
	stringEndfor = "endfor"
	stringElse   = "else"
	stringBlock  = "block"
	stringSet    = "set"
	stringInclude = "include"
	stringExtends = "extends"
	stringMacro  = "macro"
	
	// Common operators
	stringEquals = "=="
	stringNotEquals = "!="
	stringAnd    = "and"
	stringOr     = "or"
	stringNot    = "not"
	stringIn     = "in"
	stringIs     = "is"
)

// GlobalStringCache provides a centralized cache for string interning
type GlobalStringCache struct {
	sync.RWMutex
	strings map[string]string
}

var (
	// Singleton instance of the global string cache
	globalCache = newGlobalStringCache()
)

// newGlobalStringCache creates a new global string cache with pre-populated common strings
func newGlobalStringCache() *GlobalStringCache {
	cache := &GlobalStringCache{
		strings: make(map[string]string, 64), // Pre-allocate capacity
	}
	
	// Pre-populate with common strings
	commonStrings := []string{
		stringDiv, stringSpan, stringP, stringA, stringImg,
		stringHref, stringClass, stringId, stringStyle,
		stringIf, stringFor, stringEnd, stringEndif, stringEndfor,
		stringElse, stringBlock, stringSet, stringInclude, stringExtends,
		stringMacro, stringEquals, stringNotEquals, stringAnd, 
		stringOr, stringNot, stringIn, stringIs,
		// Add empty string as well
		"",
	}
	
	for _, s := range commonStrings {
		cache.strings[s] = s
	}
	
	return cache
}

// Intern returns an interned version of the input string
// For strings that are already in the cache, the cached version is returned
// Otherwise, the input string is added to the cache and returned
func Intern(s string) string {
	// Fast path for very common strings to avoid lock contention
	switch s {
	case stringDiv, stringSpan, stringP, stringA, stringImg, 
	     stringIf, stringFor, stringEnd, stringEndif, stringEndfor, 
	     stringElse, "":
		return s
	}
	
	// Don't intern strings that are too long
	if len(s) > maxCacheableLength {
		return s
	}
	
	// Use read lock for lookup first (less contention)
	globalCache.RLock()
	cached, exists := globalCache.strings[s]
	globalCache.RUnlock()
	
	if exists {
		return cached
	}
	
	// Not found with read lock, acquire write lock to add
	globalCache.Lock()
	defer globalCache.Unlock()
	
	// Check again after acquiring write lock (double-checked locking)
	if cached, exists := globalCache.strings[s]; exists {
		return cached
	}
	
	// Add to cache and return
	globalCache.strings[s] = s
	return s
}

// InternSlice interns all strings in a slice
func InternSlice(slice []string) []string {
	for i, s := range slice {
		slice[i] = Intern(s)
	}
	return slice
}