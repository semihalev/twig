# Findings on Macro String Parsing Issues

## Issue Summary
The macro implementation is having trouble parsing string literals in macro parameters, especially when special characters or escape sequences are used.

## Specific Problems Identified
1. String tokenization in the tokenizer function may not be properly handling escape sequences
2. There's an issue with VAR_START/VAR_END token handling when they appear within string literals
3. The test files themselves have issues with Go's own string escaping when trying to test escape sequences

## Recommended Fixes
1. Modify the tokenizer to correctly handle escape sequences in string literals (line ~161-185)
2. Improve the parseExpression function to handle nested expressions correctly (line ~430-450)
3. Add string processing function to handle escape sequences properly
4. Fix the HTML attribute value handling in the macro test to properly parse template variables

## Implementation Plan
The key fix is in how string literals are tokenized and then processed in the expression parser.
The tokenize function needs to properly handle escape sequences at tokenization time.
The parseExpression function needs to be updated to properly evaluate string literals with escape sequences.

