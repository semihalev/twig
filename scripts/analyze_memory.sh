#!/bin/bash

# Create output directory
mkdir -p reports

echo "=== Running Memory Allocation Analysis for Twig Template Engine ==="
echo ""

# Run standard benchmarks with memory statistics
echo "Step 1: Running benchmarks with memory allocation reporting..."
go test -run=^$ -bench=BenchmarkRender -benchmem ./memory_profile_test.go | tee reports/benchmark_results.txt

# Generate allocation profile
echo ""
echo "Step 2: Generating heap allocation profile..."
go test -run=^$ -bench=BenchmarkRenderComplexTemplate -benchmem -memprofile=reports/heap.prof ./memory_profile_test.go

# Run heap profile analysis and save top allocations
echo ""
echo "Step 3: Analyzing allocation profile..."
go tool pprof -text -alloc_space reports/heap.prof > reports/top_allocations.txt

echo ""
echo "Step 4: Generating detailed memory profile report..."
# Run with different template complexities
echo "   - Profiling simple templates..."
go run cmd/profile/main.go -complexity=1 -iterations=1000 -memprofile=reports/simple.prof > reports/simple_profile.txt

echo "   - Profiling medium templates..."
go run cmd/profile/main.go -complexity=2 -iterations=1000 -memprofile=reports/medium.prof > reports/medium_profile.txt

echo "   - Profiling complex templates..."
go run cmd/profile/main.go -complexity=3 -iterations=1000 -memprofile=reports/complex.prof > reports/complex_profile.txt

# Generate flamegraph (requires go-torch if available)
if command -v go-torch &> /dev/null
then
    echo ""
    echo "Step 5: Generating flamegraph visualization..."
    go-torch -alloc_space reports/heap.prof -file reports/allocations_flamegraph.svg
fi

# Compile the comprehensive report
echo ""
echo "Step 6: Compiling final report..."

cat > reports/memory_optimization_report.md << 'EOF'
# Twig Template Engine Memory Optimization Report

## Summary

This report analyzes memory allocation patterns in the Twig template engine to identify areas for implementing a zero-allocation rendering path.

## Benchmark Results

```
EOF

cat reports/benchmark_results.txt >> reports/memory_optimization_report.md

cat >> reports/memory_optimization_report.md << 'EOF'
```

## Top Allocation Sources

The following are the top functions allocating memory during template rendering:

```
EOF

head -20 reports/top_allocations.txt >> reports/memory_optimization_report.md

cat >> reports/memory_optimization_report.md << 'EOF'
```

## Memory Profile by Template Complexity

### Simple Templates

EOF

grep -A 10 "Memory" reports/simple_profile.txt >> reports/memory_optimization_report.md

cat >> reports/memory_optimization_report.md << 'EOF'

### Medium Templates

EOF

grep -A 10 "Memory" reports/medium_profile.txt >> reports/memory_optimization_report.md

cat >> reports/memory_optimization_report.md << 'EOF'

### Complex Templates

EOF

grep -A 10 "Memory" reports/complex_profile.txt >> reports/memory_optimization_report.md

cat >> reports/memory_optimization_report.md << 'EOF'

## Key Allocation Hotspots

Based on the profiling data, these areas should be prioritized for optimization:

1. **String Operations** - String concatenation, substring operations, and conversions
2. **Context Creation** - Creating and copying RenderContext objects
3. **Map Allocations** - Temporary maps created during rendering
4. **Slice Allocations** - Dynamic arrays for node collections
5. **Expression Evaluation** - Temporary objects created during expression processing
6. **Buffer Management** - Output buffer allocations
7. **Function/Filter Calls** - Parameter passing and result handling

## Optimization Strategies

### String Operations

- Replace string concatenation with direct writes to io.Writer
- Use pooled byte buffers instead of creating new strings
- Implement specialized ToString methods to avoid allocations for common types

### Context Handling

- Implement pooling for RenderContext objects
- Create a linked-context mechanism instead of copying values
- Preallocate and reuse context maps

### Map and Slice Allocations

- Preallocate maps and slices with known capacities
- Reuse map and slice objects from pools
- Avoid unnecessary copying of collections

### Expression Evaluation

- Pool expression evaluation result objects
- Optimize common expression patterns with specialized handlers
- Reduce intermediate allocations in expression trees

### Implementation Plan

1. Start with the highest allocation areas first
2. Implement object pooling for all major components
3. Create specialized non-allocating paths for common operations
4. Revise string handling to minimize allocations
5. Optimize hot spots in critical rendering code paths

## Next Steps

1. Implement object pools for all identified allocation sources
2. Create benchmarks to validate each optimization
3. Develop specialized string handling utilities
4. Optimize context handling and cloning
5. Enhance expression evaluation to minimize allocations

EOF

echo ""
echo "Report generation complete. See reports/memory_optimization_report.md for results."
echo ""