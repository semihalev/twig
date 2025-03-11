# HTML Rendering Fix for Twig Template Engine

## Issue Summary

The original Twig template engine implementation had several issues with HTML rendering:

1. Whitespace trimming and HTML minification were breaking HTML tag structure
2. Character escaping in HTML attributes, JavaScript, and CSS was inconsistent
3. The engine was modifying HTML structure outside template blocks
4. Code added to fix these problems was overly complex

## Core Engine Fix

The solution applies a simple principle: **A template engine should not manipulate HTML at all.**

### Implementation

1. **Removed all HTML-specific settings and code:**
   - Removed `preserveWhitespace`, `prettyOutputHTML`, and `preserveAttributes` settings
   - Eliminated HTML detection and formatting logic completely

2. **Simplified TextNode.Render to one line:**
   ```go
   func (n *TextNode) Render(w io.Writer, ctx *RenderContext) error {
       // Simply write the content as-is without any modification
       _, err := w.Write([]byte(n.content))
       return err
   }
   ```

3. **Simplified PrintNode.Render to output values exactly as provided:**
   ```go
   func (n *PrintNode) Render(w io.Writer, ctx *RenderContext) error {
       // Get the value
       result, err := ctx.EvaluateExpression(n.expression)
       if err != nil {
           return err
       }

       // Handle callable results (for macros)
       if callable, ok := result.(func(io.Writer) error); ok {
           return callable(w)
       }

       // Convert to string and output directly
       str := ctx.ToString(result)
       _, err = w.Write([]byte(str))
       return err
   }
   ```

4. **Simplified SpacelessNode.Render to just render content without HTML manipulation:**
   ```go
   func (n *SpacelessNode) Render(w io.Writer, ctx *RenderContext) error {
       // Just render the content directly - no HTML manipulation
       for _, node := range n.body {
           if err := node.Render(w, ctx); err != nil {
               return err
           }
       }
       return nil
   }
   ```

## Template Writing Best Practices

With these changes, template authors must handle HTML formatting themselves:

1. **Always quote HTML attributes properly:**
   ```html
   <input name="{{ name }}" value="{{ value }}">
   ```

2. **Add quotes around JavaScript string values:**
   ```html
   <script>
     var apiUrl = '{{ url }}';     <!-- Single quotes around variable -->
     var count = {{ count }};      <!-- No quotes for numbers -->
   </script>
   ```

3. **Use whitespace control modifiers when needed:**
   ```html
   <div>
     {{- content -}}
   </div>
   ```

## Benefits

This simplified approach offers several advantages:

1. **Simpler code** - Removed complex HTML parsing and formatting logic
2. **More predictable output** - Templates render exactly as written
3. **Improved performance** - No extra HTML processing overhead
4. **Better separation of concerns** - Template authors handle markup, engine handles variables
5. **Fewer bugs** - Less code means fewer places for bugs to hide

## Example

```html
<!DOCTYPE html>
<html>
<head>
    <title>{{ title }}</title>
    <style>
        .item { color: {{ textColor }}; }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{ heading }}</h1>
        {% for item in items %}
            <div class="item">{{ item.name }}</div>
        {% endfor %}
    </div>
    <script>
        var apiUrl = '{{ apiEndpoint }}';
        var maxItems = {{ maxItems }};
    </script>
</body>
</html>
```

# Operator Expression Fixes

## Fixed Issues

1. **Exponentiation Operator (^)**: Implemented proper handling of the exponentiation operator (^) in Twig templates.
   - Added the ^ character to the isOperator check in the tokenizer
   - Ensured that the operator is correctly parsed and processed in expressions
   - Added the exponentiation operation in evaluateBinaryOp using math.Pow()
   - Added test cases for various exponentiation scenarios

2. **String Concatenation (~)**: Fixed issues with the string concatenation operator in Twig templates.
   - Added proper handling in the tokenizer and parser
   - Updated the parseExpression method to handle multiple binary operators in sequence

3. **Short-Circuit Evaluation**: Fixed the logical operators to implement proper short-circuit evaluation.
   - Modified `BinaryNode` evaluation to only evaluate the right side when needed
   - Fixed issue with expressions like `{% if foo is defined and foo > 5 %}` that failed when variables were undefined
   - Implementation in `EvaluateExpression` method:
   ```go
   case *BinaryNode:
       left, err := ctx.EvaluateExpression(n.left)
       if err != nil {
           return nil, err
       }

       // Implement short-circuit evaluation for logical operators
       if n.operator == "and" || n.operator == "&&" {
           // For "and" operator, if left side is false, return false without evaluating right side
           if !ctx.toBool(left) {
               return false, nil
           }
       } else if n.operator == "or" || n.operator == "||" {
           // For "or" operator, if left side is true, return true without evaluating right side
           if ctx.toBool(left) {
               return true, nil
           }
       }

       // For other operators or if short-circuit condition not met, evaluate right side
       right, err := ctx.EvaluateExpression(n.right)
       if err != nil {
           return nil, err
       }

       return ctx.evaluateBinaryOp(n.operator, left, right)
   }
   ```

# Function Support in For Loops

## Problem Description

The Twig template engine previously couldn't properly handle function calls directly in for loop sequences:

```twig
{% for i in range(1, 3) %}
  {{ i }}
{% endfor %}
```

Instead, users had to use a workaround by first storing the function result in a variable:

```twig
{% set numbers = range(1, 3) %}
{% for i in numbers %}
  {{ i }}
{% endfor %}
```

This made templates more verbose and less intuitive compared to standard Twig syntax.

### Negative Step Values

There's a limitation with direct use of negative step values in range functions:

```twig
{% for i in range(5, 1, -1) %}  {# This doesn't work correctly #}
  {{ i }}
{% endfor %}
```

This is because the tokenizer treats the minus sign and the number as separate tokens. To work around this, use variables:

```twig
{% set step = -1 %}
{% for i in range(5, 1, step) %}  {# This works correctly #}
  {{ i }}
{% endfor %}
```

## Solution Approach

We implemented a comprehensive fix with the following components:

1. **Tokenizer Enhancement**:
   - Modified the HTML-preserving tokenizer to properly identify function calls in for loop sequences
   - Added special detection for parentheses and arguments in function calls
   - Used the tokenizeExpression method to properly break down function calls into their component tokens
   
2. **Parser Improvement**:
   - Enhanced the ForNode implementation to handle FunctionNode as a sequence
   - Added direct function call support within the ForNode.Render method
   - Implemented special handling for the range function to call it directly with evaluated arguments
   
3. **Code Structure Changes**:
   - Added a new ForNode.renderForLoop helper method to separate sequence evaluation from iteration logic
   - Enhanced argument evaluation to ensure proper parameter passing to functions
   - Added validation and error handling for function calls

## Results

After implementing these changes:

1. Templates can now use functions directly in for loops:
   ```twig
   {% for i in range(1, 3) %}
     {{ i }}
   {% endfor %}
   ```
   
2. Multiple function signatures are supported:
   ```twig
   {% for i in range(1, 10, 2) %}  {# With step parameter #}
     {{ i }}
   {% endfor %}
   ```
   
3. Loop variables work correctly with function results:
   ```twig
   {% for i in range(1, 3) %}
     {{ loop.index }}: {{ i }}
   {% endfor %}
   ```
   
This enhancement brings the template engine closer to full Twig compatibility, making templates more concise and intuitive for users familiar with the Twig syntax.

3. **Operator Precedence**: Implemented proper operator precedence rules for arithmetic operations.
   - Added precedence levels for different operators (e.g., * and / have higher precedence than + and -)
   - Modified the parser to correctly build the expression tree based on operator precedence
   - Added test cases to verify precedence rules are followed

4. **Modulo Operator (%)**: Added support for the modulo operator in templates.
   - Implemented evaluateBinaryOp case for the % operator using math.Mod
   - Added test cases for modulo operations

## Implementation Details

1. **Operator Tokenization**:
   - Added ^ and ~ to the list of recognized operators in ContainsAny checks
   - Modified tokenizeExpression to properly handle operators in expressions

2. **Multiple Operators in Expressions**:
   - Updated the parseExpression method to use a loop for binary operators:
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

3. **Operator Precedence System**:
   - Defined precedence levels for operators:
   ```go
   const (
       PREC_LOWEST  = 0
       PREC_OR      = 1  // or, ||
       PREC_AND     = 2  // and, &&
       PREC_COMPARE = 3  // ==, !=, <, >, <=, >=, in, not in, etc.
       PREC_SUM     = 4  // +, -, ~
       PREC_PRODUCT = 5  // *, /, %
       PREC_POWER   = 6  // ^
       PREC_PREFIX  = 7  // not, !, +, - (unary)
   )
   ```
   - Modified the parser to build the expression tree according to precedence

4. **Operator Evaluation**:
   - Added implementation for the exponentiation and modulo operators:
   ```go
   case "^":
       // Exponentiation operator
       if lNum, lok := ctx.toNumber(left); lok {
           if rNum, rok := ctx.toNumber(right); rok {
               return math.Pow(lNum, rNum), nil
           }
       }
   
   case "%":
       // Modulo operator
       if lNum, lok := ctx.toNumber(left); lok {
           if rNum, rok := ctx.toNumber(right); rok {
               if rNum == 0 {
                   return nil, errors.New("modulo by zero")
               }
               return math.Mod(lNum, rNum), nil
           }
       }
   ```

## Testing

Created comprehensive test cases:
- Basic arithmetic operators: +, -, *, /
- Exponentiation operator: ^
- Modulo operator: %
- String concatenation operator: ~
- Operator precedence: e.g., 2 + 3 * 4 = 14 (not 20)
- Parenthesized expressions: (2 + 3) * 4 = 20

## Example Templates

```twig
{# Exponentiation #}
{{ 2 ^ 3 }}                        {# Outputs: 8 #}

{# Modulo #}
{{ 10 % 3 }}                       {# Outputs: 1 #}

{# String concatenation #}
{{ 'hello' ~ ' ' ~ 'world' }}      {# Outputs: hello world #}

{# Operator precedence #}
{{ 2 + 3 * 4 }}                    {# Outputs: 14 (not 20) #}
{{ (2 + 3) * 4 }}                  {# Outputs: 20 #}
```