package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"runtime"
	"time"

	"github.com/semihalev/twig"
)

const (
	// Complex template for memory testing (Twig)
	ComplexTwigTemplate = `
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
</div>`

	// Complex template for memory testing (Go)
	ComplexGoTemplate = `
<div>
  <h1>{{ .Title }}</h1>
  <p>Welcome {{ .Name }}!</p>
  
  <h2>Items</h2>
  <ul>
    {{range $index, $item := .Items}}
      <li class="{{if isOdd $index}}odd{{else}}even{{end}}">
        <span>{{ $item.Name }}</span>
        <span>${{ $item.Price }}</span>
        {{if $item.Available}}
          <span class="available">In Stock</span>
        {{else}}
          <span class="unavailable">Out of Stock</span>
        {{end}}
      </li>
    {{end}}
  </ul>
  
  <div class="footer">
    <p>Total items: {{ len .Items }}</p>
  </div>
</div>`
)

// Item struct for template data
type Item struct {
	Name      string
	Price     float64
	Available bool
}

// Context for Go templates
type GoContext struct {
	Name  string
	Title string
	Items []Item
}

// printMemoryUsage prints the memory usage in GB and MB
func printMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MB, TotalAlloc = %v MB, HeapAlloc = %v MB\n",
		bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.HeapAlloc))
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

// getMemoryStats gets heap allocation metrics
func getMemoryStats() (float64, float64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return bToMb(m.HeapAlloc), bToMb(m.TotalAlloc)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("Template Engine Memory Benchmark")
	fmt.Println("==================================================")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("CPU: %d cores\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("==================================================")
	fmt.Println()

	// Define test data
	items := []Item{
		{Name: "Phone", Price: 599.99, Available: true},
		{Name: "Laptop", Price: 1299.99, Available: true},
		{Name: "Tablet", Price: 399.99, Available: false},
		{Name: "Watch", Price: 199.99, Available: true},
		{Name: "Headphones", Price: 99.99, Available: false},
	}

	twigContext := map[string]interface{}{
		"name":  "John",
		"title": "Product List",
		"items": items,
	}

	goContext := GoContext{
		Name:  "John",
		Title: "Product List",
		Items: items,
	}

	// Go template function map
	funcMap := template.FuncMap{
		"isOdd": func(i int) bool {
			return i%2 == 1
		},
	}

	// Create and parse templates
	twigEngine := twig.New()
	err := twigEngine.RegisterString("complex", ComplexTwigTemplate)
	if err != nil {
		fmt.Printf("Error registering Twig template: %v\n", err)
		return
	}

	goComplexTmpl := template.Must(template.New("complex").Funcs(funcMap).Parse(ComplexGoTemplate))

	// Warm up to stabilize memory usage
	fmt.Println("Warming up templates...")
	for i := 0; i < 100; i++ {
		twigEngine.Render("complex", twigContext)
		var buf bytes.Buffer
		goComplexTmpl.Execute(&buf, goContext)
	}

	// Force garbage collection before test
	runtime.GC()
	time.Sleep(time.Millisecond * 100)

	// Benchmark Twig memory usage
	fmt.Println("\nTwig Memory Usage Test:")
	fmt.Println("--------------------------------------------------")

	iterations := 10000
	var twigResults []string

	// Record initial memory state
	heapBefore, totalBefore := getMemoryStats()

	// Run twig template rendering
	startTime := time.Now()
	for i := 0; i < iterations; i++ {
		result, err := twigEngine.Render("complex", twigContext)
		if err != nil {
			fmt.Printf("Error rendering Twig template: %v\n", err)
			return
		}
		twigResults = append(twigResults, result)
	}
	twigTime := time.Since(startTime)

	// Record memory after test
	heapAfter, totalAfter := getMemoryStats()
	heapDiffTwig := heapAfter - heapBefore
	totalDiffTwig := totalAfter - totalBefore

	// Memory per operation
	memPerOpTwig := heapDiffTwig * 1024 / float64(iterations) // KB per operation

	fmt.Printf("  %d iterations in %v (%.2f µs/op)\n",
		iterations, twigTime, float64(twigTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("  Heap memory increased by: %.2f MB (%.2f KB per operation)\n", heapDiffTwig, memPerOpTwig)
	fmt.Printf("  Total allocations increased by: %.2f MB\n", totalDiffTwig)

	// Clear results and force garbage collection
	twigResults = nil
	runtime.GC()
	time.Sleep(time.Millisecond * 100)

	// Benchmark Go template memory usage
	fmt.Println("\nGo Template Memory Usage Test:")
	fmt.Println("--------------------------------------------------")

	var goResults []string

	// Record initial memory state
	heapBefore, totalBefore = getMemoryStats()

	// Run go template rendering
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		var buf bytes.Buffer
		err := goComplexTmpl.Execute(&buf, goContext)
		if err != nil {
			fmt.Printf("Error rendering Go template: %v\n", err)
			return
		}
		goResults = append(goResults, buf.String())
	}
	goTime := time.Since(startTime)

	// Record memory after test
	heapAfter, totalAfter = getMemoryStats()
	heapDiffGo := heapAfter - heapBefore
	totalDiffGo := totalAfter - totalBefore

	// Memory per operation
	memPerOpGo := heapDiffGo * 1024 / float64(iterations) // KB per operation

	fmt.Printf("  %d iterations in %v (%.2f µs/op)\n",
		iterations, goTime, float64(goTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("  Heap memory increased by: %.2f MB (%.2f KB per operation)\n", heapDiffGo, memPerOpGo)
	fmt.Printf("  Total allocations increased by: %.2f MB\n", totalDiffGo)

	// Summary
	fmt.Println("\n==================================================")
	fmt.Println("Memory Benchmark Results Summary")
	fmt.Println("==================================================")
	fmt.Printf("Twig:         %.2f µs/op, %.2f KB/op\n",
		float64(twigTime.Nanoseconds())/float64(iterations)/1000.0, memPerOpTwig)
	fmt.Printf("Go Template:  %.2f µs/op, %.2f KB/op\n",
		float64(goTime.Nanoseconds())/float64(iterations)/1000.0, memPerOpGo)

	fmt.Println("\nRelative Performance:")
	fmt.Printf("Speed: Twig is %.2fx %s than Go Template\n",
		abs(float64(twigTime.Nanoseconds())/float64(goTime.Nanoseconds())),
		ifThenElse(twigTime.Nanoseconds() < goTime.Nanoseconds(), "faster", "slower"))
	fmt.Printf("Memory: Twig uses %.2fx %s memory than Go Template\n",
		abs(memPerOpTwig/memPerOpGo),
		ifThenElse(memPerOpTwig < memPerOpGo, "less", "more"))
	fmt.Println("==================================================")

	// Write results to file
	goVersion := runtime.Version()
	resultData := fmt.Sprintf(`
## Memory Benchmark Results (as of %s)

Environment:
- Go version: %s
- CPU: %d cores
- GOMAXPROCS: %d

| Engine      | Time (µs/op) | Memory Usage (KB/op) |
|-------------|--------------|----------------------|
| Twig        | %.2f         | %.2f                 |
| Go Template | %.2f         | %.2f                 |

Twig is %.2fx %s than Go's template engine.
Twig uses %.2fx %s memory than Go's template engine.
`,
		time.Now().Format("2006-01-02"),
		goVersion,
		runtime.NumCPU(),
		runtime.GOMAXPROCS(0),
		float64(twigTime.Nanoseconds())/float64(iterations)/1000.0, memPerOpTwig,
		float64(goTime.Nanoseconds())/float64(iterations)/1000.0, memPerOpGo,
		abs(float64(twigTime.Nanoseconds())/float64(goTime.Nanoseconds())),
		ifThenElse(twigTime.Nanoseconds() < goTime.Nanoseconds(), "faster", "slower"),
		abs(memPerOpTwig/memPerOpGo),
		ifThenElse(memPerOpTwig < memPerOpGo, "less", "more"))

	os.WriteFile("MEMORY_RESULTS.md", []byte(resultData), 0644)
	fmt.Println("Memory benchmark results written to MEMORY_RESULTS.md")
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func ifThenElse(condition bool, a, b string) string {
	if condition {
		return a
	}
	return b
}