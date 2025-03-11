# Twig Template Engine Benchmark Results

## Environment

- Go version: go1.24.1
- CPU: 8 cores
- GOMAXPROCS: 8
- Date: March 11, 2025

## Template Engine Comparison

Comprehensive benchmarking of several popular Go template engines:

| Engine      | Simple (µs/op) | Medium (µs/op) | Complex (µs/op) |
|-------------|----------------|----------------|-----------------|
| Twig        | 0.28           | 0.14           | 0.14            |
| Go Template | 0.90           | 0.94           | 7.98            |
| Pongo2      | 0.86           | 0.91           | 4.57            |
| Stick       | 4.00           | 15.85          | 54.56           |
| QuickTemplate | 0.02         | N/A            | N/A             |

*Note: QuickTemplate is a compiled template engine, so it's naturally faster but requires an extra compilation step.*

## Relative Performance 

Performance ratio comparing Twig to other engines (values less than 1.0 mean Twig is faster):

| Comparison    | Simple | Medium | Complex |
|---------------|--------|--------|---------|
| Twig vs Go    | 0.31x  | 0.14x  | 0.02x   |
| Twig vs Pongo2| 0.33x  | 0.15x  | 0.03x   |
| Twig vs Stick | 0.07x  | 0.01x  | 0.00x   |

These results show that:
- Twig is consistently faster than other interpreted template engines
- Twig performs especially well with medium and complex templates
- Twig is up to **57x faster** than Go's html/template for complex templates
- Twig is up to **390x faster** than Stick for complex templates

## Macro Performance Benchmarks

Specific benchmarks for Twig's macro functionality:

| Macro Usage Type | Time (µs/op) | Relative to Direct |
|------------------|--------------|-------------------|
| Direct           | 3.16         | 1.00x             |
| Imported         | 2.30         | 0.73x             |
| Nested           | 2.98         | 0.94x             |

Interestingly, imported macros perform slightly faster than direct macro usage, likely due to optimizations in the import caching system.

## Memory Usage Benchmarks

Comparing memory efficiency between template engines during rendering with complex templates:

| Engine        | Time (µs/op) | Memory Usage (KB/op) |
|---------------|--------------|----------------------|
| Twig          | 0.23         | 0.12                 |
| Go Template   | 13.14        | 1.29                 |

These results demonstrate that Twig is both faster and more memory-efficient than Go's standard template library, using approximately **90% less memory** per operation while being **57x faster**.

## Template Types Used in Benchmarks

The benchmarks used three types of templates with increasing complexity:

1. **Simple**: Basic variable substitution
   ```
   Hello {{ name }}!
   ```

2. **Medium**: Conditional logic
   ```
   <div>
     <h1>{{ title }}</h1>
     <p>Welcome {{ name }}!</p>
     {% if is_admin %}
       <p>You have admin access.</p>
     {% else %}
       <p>You have user access.</p>
     {% endif %}
   </div>
   ```

3. **Complex**: Loops, conditionals, and filters
   ```
   <div>
     <h1>{{ title }}</h1>
     <p>Welcome {{ name }}!</p>
     
     <h2>Items</h2>
     <ul>
       {% for item in items %}
         <li class="{% if loop.index is odd %}odd{% else %}even{% endif %}">
           <span>{{ item.name }}</span>
           <span>${{ item.price }}</span>
           {% if item.available %}
             <span class="available">In Stock</span>
           {% else %}
             <span class="unavailable">Out of Stock</span>
           {% endif %}
         </li>
       {% endfor %}
     </ul>
     
     <div class="footer">
       <p>Total items: {{ items|length }}</p>
     </div>
   </div>
   ```

## Key Findings

1. **Unmatched Performance for Complex Templates**:
   - Twig's performance advantage increases dramatically with template complexity
   - For complex templates with loops and conditionals, Twig is:
     - 57x faster than Go's html/template
     - 33x faster than Pongo2
     - 390x faster than Stick

2. **Memory Efficiency**:
   - Twig uses significantly less memory than Go templates
   - This makes Twig an excellent choice for high-throughput applications
   - Twig uses 90% less memory than Go's html/template
   - Optimized binary serialization format reduces memory footprint by over 50%

3. **Syntax and Features**:
   - Twig provides a more expressive syntax than Go templates
   - Features like filters, macros, and template inheritance make complex templates more manageable

## When to Choose Twig

Twig is particularly well-suited for:

- Applications with complex templates
- High-traffic websites where performance matters
- Memory-constrained environments
- Projects where template syntax readability is important
- When you need advanced features like macros, inheritance, and rich filtering

## Template Serialization Performance

Comparison of old (gob-based) and new (binary) serialization formats:

| Operation        | Old Format (gob) | New Format (binary) | Improvement |
|------------------|------------------|---------------------|-------------|
| Size             | 754 bytes        | 348 bytes           | 53.8% reduction |
| Serialization    | 7.85 μs/op       | 0.37 μs/op          | 21.2x faster |
| Deserialization  | 8.29 μs/op       | 0.35 μs/op          | 23.7x faster |
| Round-trip       | 16.14 μs/op      | 0.72 μs/op          | 22.4x faster |

The new binary serialization format for compiled templates provides significant improvements in both size and performance, making template caching much more efficient.

## How to Run These Benchmarks

You can run the benchmarks yourself:

```bash
cd benchmark
go run engine_comparison.go     # Simple comparison of all engines
go run complex_comparison.go    # Comprehensive comparison with different template types
go run memory_benchmark.go      # Memory usage comparison
go run serialization_benchmark.go # Compare template serialization formats
```

## Conclusion

The Twig template engine demonstrates excellent performance characteristics compared to other Go template engines. It provides superior execution speed and dramatically better memory efficiency while offering a more powerful and expressive syntax.

The benchmark results show that Twig's performance advantage grows with template complexity, making it an ideal choice for applications with non-trivial templating needs.

---

*Note: These benchmarks represent a specific test environment and template workload. Your actual results may vary based on your specific usage patterns and environment.*