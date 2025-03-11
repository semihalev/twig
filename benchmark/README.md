# Twig Template Engine Benchmarks

This directory contains benchmark tests for comparing the Twig template engine with other popular Go template engines.

## Available Benchmarks

1. **simple_benchmark.go**: Basic benchmark comparing Twig with Go's standard html/template.
2. **complex_comparison.go**: Comprehensive benchmark with various template complexities comparing multiple engines.
3. **engine_comparison.go**: Simple comparison of multiple template engines.
4. **macro_benchmark.go**: Special benchmark for testing macro functionality performance.
5. **memory_benchmark.go**: Detailed memory usage comparison between Twig and Go's template.

## Running the Benchmarks

To run the benchmarks:

```bash
# Basic benchmark
go run simple_benchmark.go

# Comprehensive comparison of multiple engines
go run complex_comparison.go 

# Simple engine comparison
go run engine_comparison.go

# Macro performance benchmark
go run macro_benchmark.go

# Memory usage benchmark
go run memory_benchmark.go
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
- Macro templates: Testing macro definition, calling, and importing

## Key Findings

1. Twig consistently outperforms other interpreted template engines
2. Twig's performance advantage increases with template complexity
3. Twig is significantly more memory-efficient than other engines
4. Imported macros perform slightly better than direct macros
5. Twig is up to 2.5x faster than Go's templates for complex templates

For detailed results, see the [BENCHMARK_RESULTS.md](./BENCHMARK_RESULTS.md) file.

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