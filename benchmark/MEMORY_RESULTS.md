
## Memory Benchmark Results (as of 2025-03-11)

Environment:
- Go version: go1.24.1
- CPU: 8 cores
- GOMAXPROCS: 8

| Engine      | Time (µs/op) | Memory Usage (KB/op) |
|-------------|--------------|----------------------|
| Twig        | 6.69         | 1.28                 |
| Go Template | 11.38         | 1.43                 |

Twig is 0.59x faster than Go's template engine.
Twig uses 0.89x less memory than Go's template engine.
