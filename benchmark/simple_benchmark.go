package main

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime"
	"time"

	"github.com/semihalev/twig"
)

// Sample templates for benchmarking
const (
	// Simple template with variable substitution
	SimpleTemplate = `Hello {{ name }}!`

	// Simple Go template with variable substitution
	SimpleGoTemplate = `Hello {{ .Name }}!`

	// Template with condition
	ConditionTemplate = `{% if age > 18 %}Adult{% else %}Minor{% endif %}`

	// Go template with condition
	ConditionGoTemplate = `{{if gt .Age 18}}Adult{{else}}Minor{{end}}`

	// Template with loop
	LoopTemplate = `{% for user in users %}{{ user.Name }}, {% endfor %}`

	// Go template with loop
	LoopGoTemplate = `{{range .Users}}{{ .Name }}, {{end}}`
)

// User struct for benchmarks
type User struct {
	Name  string
	Email string
	Age   int
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("Template Engine Benchmark")
	fmt.Println("==================================================")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("CPU: %d cores\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("==================================================")
	fmt.Println()

	// Sample data for templates
	users := []User{
		{Name: "Alice", Email: "alice@example.com", Age: 28},
		{Name: "Bob", Email: "bob@example.com", Age: 35},
		{Name: "Charlie", Email: "charlie@example.com", Age: 17},
	}

	// Twig Engine benchmarks
	fmt.Println("Twig Engine:")
	twigEngine := twig.New()

	// Register templates
	err := twigEngine.RegisterString("simple", SimpleTemplate)
	if err != nil {
		fmt.Printf("Error registering simple template: %v\n", err)
		return
	}

	err = twigEngine.RegisterString("condition", ConditionTemplate)
	if err != nil {
		fmt.Printf("Error registering condition template: %v\n", err)
		return
	}

	err = twigEngine.RegisterString("loop", LoopTemplate)
	if err != nil {
		fmt.Printf("Error registering loop template: %v\n", err)
		return
	}

	// Simple template benchmark for Twig
	fmt.Println("\n  Simple template:")
	startTime := time.Now()
	iterations := 100000
	for i := 0; i < iterations; i++ {
		_, err := twigEngine.Render("simple", map[string]interface{}{
			"name": "World",
		})
		if err != nil {
			fmt.Printf("Error rendering simple template: %v\n", err)
			return
		}
	}
	twigSimpleTime := time.Since(startTime)
	fmt.Printf("    %d iterations in %v (%.2f µs/op)\n", iterations, twigSimpleTime, float64(twigSimpleTime.Nanoseconds())/float64(iterations)/1000.0)

	// Condition template benchmark for Twig
	fmt.Println("\n  Condition template:")
	startTime = time.Now()
	iterations = 100000
	for i := 0; i < iterations; i++ {
		_, err := twigEngine.Render("condition", map[string]interface{}{
			"age": 25,
		})
		if err != nil {
			fmt.Printf("Error rendering condition template: %v\n", err)
			return
		}
	}
	twigConditionTime := time.Since(startTime)
	fmt.Printf("    %d iterations in %v (%.2f µs/op)\n", iterations, twigConditionTime, float64(twigConditionTime.Nanoseconds())/float64(iterations)/1000.0)

	// Loop template benchmark for Twig
	fmt.Println("\n  Loop template:")
	startTime = time.Now()
	iterations = 100000
	for i := 0; i < iterations; i++ {
		_, err := twigEngine.Render("loop", map[string]interface{}{
			"users": users,
		})
		if err != nil {
			fmt.Printf("Error rendering loop template: %v\n", err)
			return
		}
	}
	twigLoopTime := time.Since(startTime)
	fmt.Printf("    %d iterations in %v (%.2f µs/op)\n", iterations, twigLoopTime, float64(twigLoopTime.Nanoseconds())/float64(iterations)/1000.0)

	// Go Template benchmarks
	fmt.Println("\nGo Template:")

	// Parse templates
	simpleGoTmpl, err := template.New("simple").Parse(SimpleGoTemplate)
	if err != nil {
		fmt.Printf("Error parsing simple Go template: %v\n", err)
		return
	}

	conditionGoTmpl, err := template.New("condition").Parse(ConditionGoTemplate)
	if err != nil {
		fmt.Printf("Error parsing condition Go template: %v\n", err)
		return
	}

	loopGoTmpl, err := template.New("loop").Parse(LoopGoTemplate)
	if err != nil {
		fmt.Printf("Error parsing loop Go template: %v\n", err)
		return
	}

	// Simple template benchmark for Go
	fmt.Println("\n  Simple template:")
	startTime = time.Now()
	iterations = 100000
	var buf bytes.Buffer
	for i := 0; i < iterations; i++ {
		buf.Reset()
		err := simpleGoTmpl.Execute(&buf, struct{ Name string }{"World"})
		if err != nil {
			fmt.Printf("Error rendering simple Go template: %v\n", err)
			return
		}
	}
	goSimpleTime := time.Since(startTime)
	fmt.Printf("    %d iterations in %v (%.2f µs/op)\n", iterations, goSimpleTime, float64(goSimpleTime.Nanoseconds())/float64(iterations)/1000.0)

	// Condition template benchmark for Go
	fmt.Println("\n  Condition template:")
	startTime = time.Now()
	iterations = 100000
	for i := 0; i < iterations; i++ {
		buf.Reset()
		err := conditionGoTmpl.Execute(&buf, struct{ Age int }{25})
		if err != nil {
			fmt.Printf("Error rendering condition Go template: %v\n", err)
			return
		}
	}
	goConditionTime := time.Since(startTime)
	fmt.Printf("    %d iterations in %v (%.2f µs/op)\n", iterations, goConditionTime, float64(goConditionTime.Nanoseconds())/float64(iterations)/1000.0)

	// Loop template benchmark for Go
	fmt.Println("\n  Loop template:")
	startTime = time.Now()
	iterations = 100000
	for i := 0; i < iterations; i++ {
		buf.Reset()
		err := loopGoTmpl.Execute(&buf, struct{ Users []User }{users})
		if err != nil {
			fmt.Printf("Error rendering loop Go template: %v\n", err)
			return
		}
	}
	goLoopTime := time.Since(startTime)
	fmt.Printf("    %d iterations in %v (%.2f µs/op)\n", iterations, goLoopTime, float64(goLoopTime.Nanoseconds())/float64(iterations)/1000.0)

	// Comparison results
	fmt.Println("\nComparison (Go vs Twig):")
	fmt.Printf("  Simple templates: %.2fx\n", float64(twigSimpleTime.Nanoseconds())/float64(goSimpleTime.Nanoseconds()))
	fmt.Printf("  Condition templates: %.2fx\n", float64(twigConditionTime.Nanoseconds())/float64(goConditionTime.Nanoseconds()))
	fmt.Printf("  Loop templates: %.2fx\n", float64(twigLoopTime.Nanoseconds())/float64(goLoopTime.Nanoseconds()))
	fmt.Println("\n(Values greater than 1.0 mean Twig is slower than Go by that factor)")
	
	fmt.Println("\n==================================================")
	fmt.Println("Benchmark complete!")
	fmt.Println("==================================================")
}