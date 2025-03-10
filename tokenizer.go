package twig

import (
	"strings"
)

// Special processor for HTML attributes with embedded Twig variables
func processHtmlAttributesWithTwigVars(source string) []Token {
	var tokens []Token
	line := 1
	
	// Break up the HTML tag with embedded variables
	// For example: <input type="{{ type }}" name="{{ name }}">
	
	// First, find all the attribute pairs
	for len(source) > 0 {
		// Count newlines for line tracking
		if idx := strings.IndexByte(source, '\n'); idx >= 0 {
			line += strings.Count(source[:idx], "\n")
		}
		
		// Look for the attribute pattern: attr="{{ var }}"
		attrNameEnd := strings.Index(source, "=\"{{")
		if attrNameEnd < 0 {
			// No more embedded variables, add remaining as TEXT token
			if len(source) > 0 {
				tokens = append(tokens, Token{Type: TOKEN_TEXT, Value: source, Line: line})
			}
			break
		}
		
		// Get the attribute name
		attrName := source[:attrNameEnd]
		
		// Find the end of the embedded variable
		varEnd := strings.Index(source[attrNameEnd+3:], "}}")
		if varEnd < 0 {
			// Error - unclosed variable, just add as text
			tokens = append(tokens, Token{Type: TOKEN_TEXT, Value: source, Line: line})
			break
		}
		varEnd += attrNameEnd + 3
		
		// Extract the variable name (inside {{ }})
		varName := strings.TrimSpace(source[attrNameEnd+3:varEnd])
		
		// Add tokens: text for the attribute name, then VAR_START, var name, VAR_END
		if attrNameEnd > 0 {
			tokens = append(tokens, Token{Type: TOKEN_TEXT, Value: attrName + "=\"", Line: line})
		}
		
		// Add variable tokens
		tokens = append(tokens, Token{Type: TOKEN_VAR_START, Line: line})
		tokens = append(tokens, Token{Type: TOKEN_NAME, Value: varName, Line: line})
		tokens = append(tokens, Token{Type: TOKEN_VAR_END, Line: line})
		
		// Find the closing quote and add the rest as text
		quoteEnd := strings.Index(source[varEnd+2:], "\"")
		if quoteEnd < 0 {
			// Error - unclosed quote, just add the rest as text
			tokens = append(tokens, Token{Type: TOKEN_TEXT, Value: "\"" + source[varEnd+2:], Line: line})
			break
		}
		quoteEnd += varEnd + 2
		
		// Add the closing quote
		tokens = append(tokens, Token{Type: TOKEN_TEXT, Value: "\"", Line: line})
		
		// Move past this attribute
		source = source[quoteEnd+1:]
	}
	
	return tokens
}