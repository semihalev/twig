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
