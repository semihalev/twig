package twig

import (
	"unsafe"
)

// TagType represents the type of tag found
type TagType int

const (
	TAG_NONE TagType = iota
	TAG_VAR
	TAG_VAR_TRIM
	TAG_BLOCK
	TAG_BLOCK_TRIM
	TAG_COMMENT
)

// TagLocation represents the location of a tag in a template
type TagLocation struct {
	Type     TagType // Type of tag
	Position int     // Position in source
	Length   int     // Length of opening tag
}

// FindNextTag finds the next twig tag in a template string using
// optimized detection methods to reduce allocations and string operations.
func FindNextTag(source string, startPos int) TagLocation {
	// Quick check for empty source or position at end
	if len(source) == 0 || startPos >= len(source) || startPos < 0 {
		return TagLocation{TAG_NONE, -1, 0}
	}

	// Define the remaining source to search
	remainingSource := source[startPos:]
	remainingLen := len(remainingSource)

	// Fast paths for common cases
	if remainingLen < 2 {
		return TagLocation{TAG_NONE, -1, 0}
	}

	// Direct byte comparison for opening characters
	// This avoids string allocations and uses pointer arithmetic
	srcPtr := unsafe.Pointer(unsafe.StringData(remainingSource))
	
	// Quick check for potential tag start with { character
	for i := 0; i < remainingLen-1; i++ {
		if *(*byte)(unsafe.Add(srcPtr, i)) != '{' {
			continue
		}

		// We found a '{', check next character
		secondChar := *(*byte)(unsafe.Add(srcPtr, i+1))
		
		// Check for start of blocks
		tagPosition := startPos + i

		// Check for possible tag patterns
		switch secondChar {
		case '{': // Potential variable tag {{
			if i+2 < remainingLen && *(*byte)(unsafe.Add(srcPtr, i+2)) == '-' {
				return TagLocation{TAG_VAR_TRIM, tagPosition, 3}
			}
			return TagLocation{TAG_VAR, tagPosition, 2}
		
		case '%': // Potential block tag {%
			if i+2 < remainingLen && *(*byte)(unsafe.Add(srcPtr, i+2)) == '-' {
				return TagLocation{TAG_BLOCK_TRIM, tagPosition, 3}
			}
			return TagLocation{TAG_BLOCK, tagPosition, 2}
		
		case '#': // Comment tag {#
			return TagLocation{TAG_COMMENT, tagPosition, 2}
		}
	}

	// No tags found
	return TagLocation{TAG_NONE, -1, 0}
}

// FindTagEnd finds the end of a tag based on the type
func FindTagEnd(source string, startPos int, tagType TagType) int {
	if startPos >= len(source) {
		return -1
	}

	switch tagType {
	case TAG_VAR, TAG_VAR_TRIM:
		// Find "}}" sequence
		for i := startPos; i < len(source)-1; i++ {
			if source[i] == '}' && source[i+1] == '}' {
				return i
			}
		}
	case TAG_BLOCK, TAG_BLOCK_TRIM:
		// Find "%}" sequence
		for i := startPos; i < len(source)-1; i++ {
			if source[i] == '%' && source[i+1] == '}' {
				return i
			}
		}
	case TAG_COMMENT:
		// Find "#}" sequence
		for i := startPos; i < len(source)-1; i++ {
			if source[i] == '#' && source[i+1] == '}' {
				return i
			}
		}
	}
	
	return -1
}

// scanForTagEnd scans for a tag end sequence (e.g. }}, %}, #})
// using direct byte comparisons to avoid allocations
func scanForTagEnd(source string, startPos int, middle byte, end byte) int {
	srcLen := len(source)
	for i := startPos; i < srcLen-1; i++ {
		if source[i] == middle && i+1 < srcLen && source[i+1] == end {
			return i
		}
	}
	return -1
}

// indexOf finds the index of a substring in a string
// This is a simplified version of strings.Index for our specific use case
func indexOf(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	// For two-char sequences, use direct comparison
	if len(substr) == 2 {
		for i := 0; i < len(s)-1; i++ {
			if s[i] == substr[0] && s[i+1] == substr[1] {
				return i
			}
		}
		return -1
	}
	// For longer sequences, use a simple scan
	for i := 0; i <= len(s)-len(substr); i++ {
		matched := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				matched = false
				break
			}
		}
		if matched {
			return i
		}
	}
	return -1
}