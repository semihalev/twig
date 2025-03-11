# Twig Template Engine Findings & Solutions

## Performance and Memory Optimizations (Added in Latest Updates)

### Memory Management
- Fixed memory leak in RenderContext.Release() by properly clearing references to large objects
- Improved sync.Pool usage for efficient object reuse
- Enhanced attribute caching with proper thread safety

### Performance Optimizations
- Optimized string operations by minimizing unnecessary string conversions
- Improved collection functions with faster algorithms (contains, etc.)
- Removed excessive debug output that impacted performance
- Optimized PrintNode.Render to properly handle numeric types

### Known Issues
1. Range function support: 
   - Fixed! Direct usage of range function in for loops now works correctly
   - Now supports `{% for i in range(1, 3) %}` syntax
   - Also supports step parameter: `{% for i in range(1, 10, 2) %}`
   - Negative step values now work via variables: `{% set step = -1 %}{% for i in range(5, 1, step) %}`
   - Note: Direct negative literals like `{% for i in range(5, 1, -1) %}` require tokenizer improvements

2. Advanced Filters:
   - Some advanced filter combinations don't work correctly
   - Complex expressions with filters need more parser work
   
3. Sort behavior:
   - The sort filter handles mixed types with natural sort, not lexicographic sort

4. Short-Circuit Evaluation Issues:
   - The `and` operator lacks proper short-circuit evaluation when checking for existence combined with other operations
   - Expressions like `{% if foo is defined and foo > 5 %}` fail with "unsupported binary operator" errors when foo is undefined
   - The issue is in `evaluateBinaryOp` in render.go where both sides of `and` are evaluated before applying the operator
   - Need to implement proper short-circuit evaluation so that right-side expressions aren't evaluated when left side is false

### Fixed Issues

1. 'not defined' syntax support:
   - Added special handling in parseBinaryExpression for "not defined" pattern
   - Parser now correctly converts `{% if var not defined %}` to a TestNode wrapped in a UnaryNode
   - Both standard form `{% if var is not defined %}` and alternative form `{% if var not defined %}` now work

## Problem Summary

The Twig template engine in Go has critical issues with HTML rendering, particularly:

1. **Unquoted HTML Attributes**: Many HTML attributes are rendered without proper quotes, causing browser JavaScript errors
2. **Variable Replacement in Attributes**: Template variables within HTML attributes like `href` and `src` aren't being processed correctly
3. **HTML Structure Preservation**: The engine alters the original HTML structure with improper whitespace handling
4. **JavaScript Context**: Issues with variable values in script tags causing JavaScript syntax errors

## Root Causes

After investigating the codebase, we identified several issues in the core rendering logic:

1. In `node.go`:
   - The `TextNode.Render` method's HTML attribute handling is insufficient
   - The regex patterns for attribute handling in `preserveAttributes` aren't robust enough
   - The whitespace handling is causing HTML structure changes

2. In function handling:
   - Variables in JavaScript context aren't properly quoted
   - Template variables in HTML attributes aren't consistently processed

## Attempted Solutions

### 1. Improved HTML Attribute Handling

We modified the `TextNode.Render` method to better handle HTML attributes, specifically:
- Added special handling for `href` and `src` attributes to ensure proper quoting
- Improved handling of boolean attributes that need empty values
- Added detection for script and style tags to preserve their content exactly

### 2. Better JavaScript Context Handling

In the `PrintNode.Render` method, we enhanced the JavaScript variable handling:
- Added context detection for script tags
- Improved handling of JS literals vs. string values
- Fixed issues with variable substitution in script tags

### 3. Attribute Preservation

We made several changes to preserve HTML attributes correctly:
- Added better regex patterns to detect and fix unquoted attributes
- Fixed issues with attributes running together
- Addressed specific problematic patterns (like lang and dir attributes running together)

### 4. Temporary Post-processing Solution

For immediate compatibility, we created a post-processing function that:
- Fixes attribute formatting for critical HTML elements
- Adds proper quotes to attribute values
- Fixes cases where multiple attributes run together
- Handles special cases like boolean attributes

## Next Steps

1. **Complete Engine Fixes**:
   - Finish updating the `TextNode.Render` method to properly handle all HTML attributes
   - Ensure whitespace is preserved correctly without breaking HTML structure
   - Fix the `node.go` file to properly implement all needed changes

2. **Improve Variable Processing**:
   - Enhance how variables are processed within different contexts (HTML, JS, CSS)
   - Fix issues with function call output in attribute values

3. **Testing Strategy**:
   - Create a comprehensive test suite with various HTML templates
   - Test browser rendering with actual templates
   - Verify JavaScript execution works correctly with generated HTML

4. **Documentation**:
   - Document the changes and proper configuration settings
   - Add notes about whitespace preservation and HTML attribute handling

## Implementation Details

The key files that need modifications:

1. `/Users/semih/go/src/github.com/semihalev/twig/node.go`:
   - `TextNode.Render`: Improve attribute handling without changing HTML structure
   - `PrintNode.Render`: Better context-aware variable processing

2. Engine configuration:
   - Set `PreserveWhitespace` to true to maintain HTML structure
   - Enable `PreserveAttributes` for proper attribute handling

## Immediate Workaround

Until the engine fixes are completed, we created a post-processing function that fixes the most critical HTML rendering issues. This function:

1. Uses regex to identify and fix unquoted attributes
2. Handles special cases for common attributes
3. Fixes instances where multiple attributes run together
4. Ensures JavaScript in script tags is properly formatted

This temporary solution can be used while the core engine is being fixed.

## Conclusions

The HTML rendering issues in the Twig template engine stem from inadequate handling of HTML attributes and whitespace. The fixes required are focused on improving how attributes are quoted and how variables are processed in different contexts. By implementing these changes, we should achieve HTML output that renders correctly in browsers without JavaScript errors.

# Operator Expression Parsing Issues - Findings & Solutions

## Problem Summary

In addition to the HTML rendering issues, we've identified several problems with operator expressions in the template engine:

1. **Exponentiation Operator (^)**: The ^ operator was not being correctly parsed in templates
2. **String Concatenation (~)**: Multiple concatenation operators were not parsed correctly
3. **Operator Precedence**: Expressions like "2 + 3 * 4" were evaluated incorrectly
4. **Multiple Operator Support**: Expressions with more than one operator weren't properly handled

## Root Causes

After investigation, we found several issues in the parser:

1. **Tokenization Issues**: The tokenizer didn't properly identify all operators in all contexts
2. **Single-pass Expression Parsing**: The parser only supported a single binary operator in each expression
3. **No Precedence Rules**: All operators were treated with equal precedence
4. **Missing Multiple Operator Logic**: The parser couldn't handle sequences of binary operators

## Implemented Solutions

### 1. Tokenizer Enhancements

We fixed the tokenizer to correctly identify all operators:
- Added missing operators to string checks in the tokenizer
- Improved token type assignment for operators
- Made the tokenizeExpression function handle complex expressions correctly

### 2. Multiple Operator Support

We modified the parseExpression method to handle sequences of operators:
```go
// Loop to handle multiple binary operators in sequence
for p.tokenIndex < len(p.tokens) &&
    (p.tokens[p.tokenIndex].Type == TOKEN_OPERATOR || ...) {
    expr, err = p.parseBinaryExpression(expr)
    if err != nil {
        return nil, err
    }
}
```

### 3. Operator Precedence System

We implemented a robust operator precedence system:
- Defined precedence levels for all operators
- Modified the parser to respect precedence when building expression trees
- Added special handling for higher-precedence operators

### 4. Comprehensive Operator Tests

We created tests for all operator types and combinations:
- Basic arithmetic: +, -, *, /
- Exponentiation: ^
- Modulo: %
- String concatenation: ~
- Operator precedence: 2 + 3 * 4
- Parenthesized expressions: (2 + 3) * 4

## Implementation Details

The key files modified:

1. **parser.go**:
   - Added precedence levels for operators
   - Updated the parseExpression method to handle multiple operators
   - Enhanced parseBinaryExpression to consider operator precedence

2. **html_preserving_tokenizer.go**:
   - Updated tokenizeExpression to better handle operators
   - Fixed token type assignment for operators

3. **render.go**:
   - Verified operator evaluation logic was correct

## Conclusions

The operator expression issues in the template engine stemmed from an incomplete parser implementation. By enhancing the tokenizer and parser with proper operator precedence and multiple operator support, we've made the engine capable of handling complex expressions correctly, bringing it in line with the expectations of Twig template users.