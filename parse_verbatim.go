package twig

import (
	"fmt"
	"strings"
)

// parseVerbatim parses a verbatim tag and its content
func (p *Parser) parseVerbatim(parser *Parser) (Node, error) {
	// Get the line number of the verbatim token
	verbatimLine := parser.tokens[parser.tokenIndex-2].Line

	// Expect the block end token
	if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
		return nil, fmt.Errorf("expected block end after verbatim tag at line %d", verbatimLine)
	}
	parser.tokenIndex++

	// Collect all content until we find the endverbatim tag
	var contentBuilder strings.Builder

	for parser.tokenIndex < len(parser.tokens) {
		token := parser.tokens[parser.tokenIndex]

		// Look for the endverbatim tag
		if token.Type == TOKEN_BLOCK_START || token.Type == TOKEN_BLOCK_START_TRIM {
			// Check if this is the endverbatim tag
			if parser.tokenIndex+1 < len(parser.tokens) &&
				parser.tokens[parser.tokenIndex+1].Type == TOKEN_NAME &&
				parser.tokens[parser.tokenIndex+1].Value == "endverbatim" {

				// Skip the block start and endverbatim name
				parser.tokenIndex += 2 // Now at the endverbatim token

				// Expect the block end token
				if parser.tokenIndex >= len(parser.tokens) || !isBlockEndToken(parser.tokens[parser.tokenIndex].Type) {
					return nil, fmt.Errorf("expected block end after endverbatim at line %d", token.Line)
				}
				parser.tokenIndex++ // Skip the block end token

				// Create a verbatim node with the collected content
				return NewVerbatimNode(contentBuilder.String(), verbatimLine), nil
			}
		}

		// Add this token's content to our verbatim content
		if token.Type == TOKEN_TEXT {
			contentBuilder.WriteString(token.Value)
		} else if token.Type == TOKEN_VAR_START || token.Type == TOKEN_VAR_START_TRIM {
			// For variable tags, preserve them as literal text
			contentBuilder.WriteString("{{")

			// Skip variable start token and process until variable end
			parser.tokenIndex++

			// Process tokens until variable end
			for parser.tokenIndex < len(parser.tokens) {
				innerToken := parser.tokens[parser.tokenIndex]

				if innerToken.Type == TOKEN_VAR_END || innerToken.Type == TOKEN_VAR_END_TRIM {
					contentBuilder.WriteString("}}")
					break
				} else if innerToken.Type == TOKEN_NAME || innerToken.Type == TOKEN_STRING ||
					innerToken.Type == TOKEN_NUMBER || innerToken.Type == TOKEN_OPERATOR ||
					innerToken.Type == TOKEN_PUNCTUATION {
					contentBuilder.WriteString(innerToken.Value)
				}

				parser.tokenIndex++
			}
		} else if token.Type == TOKEN_BLOCK_START || token.Type == TOKEN_BLOCK_START_TRIM {
			// For block tags, preserve them as literal text
			contentBuilder.WriteString("{%")

			// Skip block start token and process until block end
			parser.tokenIndex++

			// Process tokens until block end
			for parser.tokenIndex < len(parser.tokens) {
				innerToken := parser.tokens[parser.tokenIndex]

				if innerToken.Type == TOKEN_BLOCK_END || innerToken.Type == TOKEN_BLOCK_END_TRIM {
					contentBuilder.WriteString("%}")
					break
				} else if innerToken.Type == TOKEN_NAME || innerToken.Type == TOKEN_STRING ||
					innerToken.Type == TOKEN_NUMBER || innerToken.Type == TOKEN_OPERATOR ||
					innerToken.Type == TOKEN_PUNCTUATION {
					// If this is the first TOKEN_NAME in a block, add a space after it
					if innerToken.Type == TOKEN_NAME && parser.tokenIndex > 0 &&
						(parser.tokens[parser.tokenIndex-1].Type == TOKEN_BLOCK_START ||
							parser.tokens[parser.tokenIndex-1].Type == TOKEN_BLOCK_START_TRIM) {
						contentBuilder.WriteString(innerToken.Value + " ")
					} else {
						contentBuilder.WriteString(innerToken.Value)
					}
				}

				parser.tokenIndex++
			}
		} else if token.Type == TOKEN_COMMENT_START {
			// For comment tags, preserve them as literal text
			contentBuilder.WriteString("{#")

			// Skip comment start token and process until comment end
			parser.tokenIndex++

			// Process tokens until comment end
			for parser.tokenIndex < len(parser.tokens) {
				innerToken := parser.tokens[parser.tokenIndex]

				if innerToken.Type == TOKEN_COMMENT_END {
					contentBuilder.WriteString("#}")
					break
				} else if innerToken.Type == TOKEN_TEXT {
					contentBuilder.WriteString(innerToken.Value)
				}

				parser.tokenIndex++
			}
		}

		parser.tokenIndex++

		// Check for end of tokens
		if parser.tokenIndex >= len(parser.tokens) {
			return nil, fmt.Errorf("unexpected end of template, unclosed verbatim tag at line %d", verbatimLine)
		}
	}

	// If we get here, we never found the endverbatim tag
	return nil, fmt.Errorf("unclosed verbatim tag at line %d", verbatimLine)
}
