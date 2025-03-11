package twig

// Processes whitespace control modifiers in the template
// Applies whitespace trimming to adjacent text tokens based on the token types
// This is called after tokenization to handle the whitespace around trimming tokens
func processWhitespaceControl(tokens []Token) []Token {
	if len(tokens) == 0 {
		return tokens
	}

	var result []Token = make([]Token, len(tokens))
	copy(result, tokens)

	// Process each token to apply whitespace trimming
	for i := 0; i < len(result); i++ {
		token := result[i]

		// Handle opening tags that trim whitespace before them
		if token.Type == TOKEN_VAR_START_TRIM || token.Type == TOKEN_BLOCK_START_TRIM {
			// If there's a text token before this, trim its trailing whitespace
			if i > 0 && result[i-1].Type == TOKEN_TEXT {
				result[i-1].Value = trimTrailingWhitespace(result[i-1].Value)
			}
		}

		// Handle closing tags that trim whitespace after them
		if token.Type == TOKEN_VAR_END_TRIM || token.Type == TOKEN_BLOCK_END_TRIM {
			// If there's a text token after this, trim its leading whitespace
			if i+1 < len(result) && result[i+1].Type == TOKEN_TEXT {
				result[i+1].Value = trimLeadingWhitespace(result[i+1].Value)
			}
		}
	}

	return result
}
