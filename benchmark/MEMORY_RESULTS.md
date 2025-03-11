
## Memory Benchmark Results (as of 2025-03-11)

Environment:
- Go version: go1.24.1
- CPU: 8 cores
- GOMAXPROCS: 8

| Engine      | Time (µs/op) | Memory Usage (KB/op) |
|-------------|--------------|----------------------|
| Twig        | 7.00         | 1.25                 |
| Go Template | 10.84         | 1.35                 |

Twig is 0.65x faster than Go's template engine.
Twig uses 0.92x less memory than Go's template engine.
