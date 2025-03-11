# Memory Allocation Analysis for Zero-Allocation Rendering

This guide explains how to use the memory profiling tools provided to identify and optimize memory allocation hotspots in the Twig template engine.

## Running the Memory Analysis

To run a comprehensive memory analysis and generate reports:

```bash
./scripts/analyze_memory.sh
```

This script will:
1. Run benchmarks with memory allocation tracking
2. Generate heap allocation profiles
3. Analyze the profiles to identify allocation hotspots
4. Create a comprehensive report in the `reports/` directory

## Interpreting the Results

The analysis generates several files:

- `reports/memory_optimization_report.md` - Main report with allocation analysis and recommendations
- `reports/benchmark_results.txt` - Raw benchmark results with memory statistics
- `reports/top_allocations.txt` - Top memory allocation sources from pprof
- `reports/heap.prof` - Heap allocation profile that can be analyzed with `go tool pprof`
- `reports/simple_profile.txt`, `reports/medium_profile.txt`, `reports/complex_profile.txt` - Profile results by template complexity

## Using Individual Profiling Tools

### Running Benchmarks with Memory Stats

```bash
go test -run=^$ -bench=. -benchmem ./memory_profile_test.go
```

This command runs all benchmarks and reports allocations per operation.

### Generating a Heap Profile

```bash
go test -run=^$ -bench=BenchmarkRenderComplexTemplate -benchmem -memprofile=heap.prof ./memory_profile_test.go
```

### Analyzing the Heap Profile

```bash
go tool pprof -alloc_space heap.prof
```

Common pprof commands:
- `top` - Show top allocation sources
- `list FunctionName` - Show line-by-line allocations in a function
- `web` - Open a web visualization of the profile

### Using the Profiling Tool

For targeted profiling of specific template complexity:

```bash
go run cmd/profile/main.go -complexity=3 -iterations=1000 -memprofile=complex.prof
```

Options:
- `-complexity` - Template complexity level (1=simple, 2=medium, 3=complex)
- `-iterations` - Number of template renders to perform
- `-memprofile` - Output file for memory profile
- `-cpuprofile` - Output file for CPU profile (optional)

## Zero-Allocation Implementation Strategy

Based on the profile results, implement optimizations in this order:

1. **Object Pooling** - Implement pools for all temporary objects
2. **String Operations** - Optimize string handling to avoid allocations
3. **Context Management** - Improve context creation, cloning, and cleanup
4. **Expression Evaluation** - Minimize allocations in expression execution
5. **Buffer Management** - Reuse output buffers with proper pooling

## Testing Your Optimizations

After each optimization, run the memory benchmarks again to verify the reduction in allocations:

```bash
go test -run=^$ -bench=BenchmarkRender -benchmem ./memory_profile_test.go
```

The goal is to see zero (or minimal) allocations per operation in the `allocs/op` column.