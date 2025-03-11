package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Printf("Running serialization benchmarks at %s\n", time.Now().Format(time.RFC1123))
	
	// Run benchmarks
	benchmarkSerialization()
	
	fmt.Println("\nBenchmarks completed successfully.")
}