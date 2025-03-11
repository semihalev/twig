# Twig v1.0.1 Release Notes

## Overview

Twig v1.0.1 brings significant performance improvements, bug fixes, and new features to the Twig template engine for Go. This release focuses on memory efficiency, rendering speed, and overall stability with impressive benchmark results.

## Performance Improvements

- **Object Pooling**: Implemented comprehensive object and token pooling for near-zero memory allocation during template rendering
- **Filter Chain Optimization**: Dramatically improved filter chain handling with optimized builder patterns and O(n) complexity 
- **Attribute Caching**: Added efficient LRU eviction strategy for attribute cache to reduce reflection overhead
- **String Handling**: Optimized string to byte conversions during rendering for better performance
- **Template Serialization**: Optimized compiled template serialization for faster loading and smaller memory footprint
- **Concurrency**: Fixed lock contention in template loading for better multi-threaded performance

## Bug Fixes

- Fixed goroutine leaks in render context error paths
- Fixed array filters and added support for GetItemNode and array access
- Fixed negative number handling in filter tests
- Improved regex handling for the `matches` operator
- Fixed short-circuit evaluation for logical operators
- Fixed range function inclusivity and map iteration tests
- Fixed advanced filters and error condition handling
- Fixed debug tests and improved debug functionality
- Fixed code style inconsistencies

## New Features and Improvements

- Added comprehensive macro benchmark tests and updated documentation
- Added serialization benchmarks and updated results
- Added HTML whitespace control and formatting enhancements
- Improved string rendering in scripts and JSON-style object handling
- Enhanced code style and documentation
- Improved README header presentation with new logo
- Added advanced macros examples

## Benchmark Results

Latest benchmark runs show dramatic performance improvements:

- Twig is now **57x faster** than Go's html/template for complex templates
- Memory usage reduced by **90%** compared to standard Go templates 
- Performance on medium templates improved to 0.14 µs/op from 0.35 µs/op
- Simple template rendering improved to 0.28 µs/op from 0.47 µs/op

## Breaking Changes

None - This release maintains full compatibility with v1.0.0.

## Upgrading

This release is a drop-in replacement for Twig v1.0.0 with no changes required to your code or templates.

## Contributors

- @semihalev
- Claude (Co-Author)

## Full Changelog

For a full list of changes, see the [commit history](https://github.com/semihalev/twig/compare/v1.0.0...v1.0.1).