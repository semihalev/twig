# Twig v1.0.3 Release Notes

## New Features

### Security Enhancement
- **Sandbox Mode**: Added comprehensive sandbox security system for templates
  - Implemented a `SecurityPolicy` interface to restrict allowed functions, filters, and tags
  - Added `DefaultSecurityPolicy` with safe defaults for common operations
  - Added `sandboxed` option to include tag for secure template inclusion
  - Added engine-level sandbox control methods: `EnableSandbox()` and `DisableSandbox()`

### Template Inheritance Enhancement
- **Parent Function**: Implemented `parent()` function for accessing parent block content
  - Allows child templates to extend rather than completely replace block content
  - Preserves proper inheritance chain for multi-level inheritance
  - Enhanced block context tracking for proper parent reference resolution

### Whitespace Control
- **Dash Modifiers**: Added support for whitespace control with dash modifiers (`-`)
  - Added support for `{{- ... }}` and `{{ ... -}}` to trim whitespace before/after output
  - Added support for `{%- ... %}` and `{% ... -%}` to trim whitespace before/after block tags
  - Preserves template readability while controlling output formatting

## Improvements
- **Code Organization**: Enhanced internal APIs for better maintainability
- **Documentation**: Updated README and wiki with new features
- **Code Quality**: Improved error handling and memory management
- **Performance**: Optimized context handling for sandboxed includes

## Comprehensive Testing
- Added tests for all new functionality:
  - Sandbox security policy tests
  - Sandbox context tests
  - Parent function in template inheritance
  - Whitespace control with dash modifiers

## Wiki Documentation
- Added wiki pages for new features:
  - Sandbox mode and security policy
  - Parent function and advanced template inheritance
  - Whitespace control with examples

---

# Twig v1.0.2 Release Notes

## New Features

### Added New Tags
- **Apply Tag**: Implement `{% apply filter %}...{% endapply %}` tag that applies filters to blocks of content
- **Verbatim Tag**: Added `{% verbatim %}...{% endverbatim %}` tag to output Twig syntax without processing it
- **Do Tag**: Implement `{% do %}` tag for performing expressions without outputting results

### Added New Filter
- **Spaceless Filter**: Added `spaceless` filter that removes whitespace between HTML tags

## Improvements
- **Path Resolution**: Fixed template path resolution for relative paths in templates
  - Properly resolves paths starting with "./" or "../" relative to the current template's directory
  - Enables templates in subdirectories to properly include/extend templates using relative paths
- **Code Organization**: Split parser functions into separate files for better maintainability
- **Documentation**: Updated README with new tags and filter documentation
- **Code Quality**: Cleaned up formatting and removed debug code

## Comprehensive Testing
- Added tests for all new functionality:
  - Verbatim tag tests
  - Apply tag tests
  - Spaceless filter tests
  - From tag tests
  - Relative path resolution tests

## Wiki Documentation
- Added comprehensive wiki pages for better documentation organization
