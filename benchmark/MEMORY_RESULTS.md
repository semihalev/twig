
## Memory Benchmark Results (as of 2025-03-11)

Environment:
- Go version: go1.24.1
- CPU: 8 cores
- GOMAXPROCS: 8

| Engine      | Time (µs/op) | Memory Usage (KB/op) |
|-------------|--------------|----------------------|
| Twig        | 3.24         | 1.23                 |
| Go Template | 8.71         | 1.26                 |

Twig is 0.37x faster than Go's template engine.
Twig uses 0.98x less memory than Go's template engine.
