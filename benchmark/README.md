# Twig Template Engine Benchmarks

This directory contains benchmark tests for comparing the Twig template engine with other popular Go template engines.

## Available Benchmarks

1. **template_benchmark.go**: Basic benchmark comparing Twig with Go's standard html/template.
2. **comprehensive_benchmark.go**: More detailed benchmark with various template complexities.
3. **full_comparison.go**: Complete benchmark suite including other popular template engines.

## Running the Benchmarks

To run the benchmarks:

```bash
# Basic benchmark
go run template_benchmark.go

# Comprehensive benchmark
go run comprehensive_benchmark.go 

# Complex compares benchmark
go run complex_comparison.go

# For full benchmarks
go run full_comparison.go
```

## Benchmark Methodology

These benchmarks measure:

1. **Rendering speed**: Operations per second and time per operation
2. **Memory usage**: Bytes allocated per operation
3. **Allocation count**: Number of memory allocations per operation

We test with various template complexities:
- Simple templates: Basic variable substitution
- Medium templates: Conditionals and object properties
- Complex templates: Loops, nested conditionals, and filters

## Sample Results

Results will vary based on your system, but here's a sample of what to expect:

```
===================================================
           Template Engine Benchmark
===================================================
Go version: go1.24.1
CPU: 8 cores
GOMAXPROCS: 8
Date: 2025-03-10 15:04:05
===================================================

Running benchmarks...

BenchmarkTwig/simple:
  Operations: 500000
  Time per op: 2.5µs
  Bytes per op: 128
  Allocs per op: 2

BenchmarkGoTemplate/simple:
  Operations: 1000000
  Time per op: 1.2µs
  Bytes per op: 64
  Allocs per op: 1

...and so on for other engines and template complexities
```

## Notes

- The standard Go template engine is generally fastest for simple templates
- Twig provides better syntax and features for complex templates
- Memory usage tends to increase with template complexity
- The benchmarks include compilation time, which affects initial rendering
- For production use, template caching significantly improves performance