
## Memory Benchmark Results (as of 2025-03-12)

Environment:
- Go version: go1.24.1
- CPU: 8 cores
- GOMAXPROCS: 8

| Engine      | Time (µs/op) | Memory Usage (KB/op) |
|-------------|--------------|----------------------|
| Twig        | 0.22         | 0.12                 |
| Go Template | 9.14         | 1.47                 |

Twig is 0.02x faster than Go's template engine.
Twig uses 0.08x less memory than Go's template engine.
