package main

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/semihalev/twig"
	"github.com/tyler-sommer/stick"
	qt "github.com/valyala/quicktemplate"
)

// Simple template for each engine
const (
	SimpleText      = `Hello {{ name }}!`
	SimpleGoText    = `Hello {{ .Name }}!`
	SimplePongoText = `Hello {{ name }}!`
	SimpleStickText = `Hello {{ name }}!`
)

// Context for template data
type Context struct {
	Name string
}

// Compare simple template rendering across all engines
func main() {
	// Benchmark configuration

	fmt.Println("==================================================")
	fmt.Println("Template Engine Comparison - Simple Rendering")
	fmt.Println("==================================================")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("CPU: %d cores\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("==================================================")
	fmt.Println()

	// Define iterations
	iterations := 100000

	//--------------------------------------------------
	// Twig Benchmark
	//--------------------------------------------------
	fmt.Println("Twig Template Engine:")
	twigEngine := twig.New()

	err := twigEngine.RegisterString("simple", SimpleText)
	if err != nil {
		fmt.Printf("Error registering template: %v\n", err)
		return
	}

	// Warm up
	for i := 0; i < 5; i++ {
		twigEngine.Render("simple", map[string]interface{}{
			"name": "World",
		})
	}

	startTime := time.Now()

	// Run benchmark
	for i := 0; i < iterations; i++ {
		_, err := twigEngine.Render("simple", map[string]interface{}{
			"name": "World",
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	twigTime := time.Since(startTime)
	fmt.Printf("  Time: %v for %d iterations (%.2f µs/op)\n",
		twigTime, iterations, float64(twigTime.Nanoseconds())/float64(iterations)/1000.0)

	//--------------------------------------------------
	// Go Template Benchmark
	//--------------------------------------------------
	fmt.Println("\nGo html/template:")
	goTmpl, err := template.New("simple").Parse(SimpleGoText)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	// Warm up
	var buf bytes.Buffer
	for i := 0; i < 5; i++ {
		buf.Reset()
		goTmpl.Execute(&buf, Context{Name: "World"})
	}

	startTime = time.Now()

	// Run benchmark
	for i := 0; i < iterations; i++ {
		buf.Reset()
		err := goTmpl.Execute(&buf, Context{Name: "World"})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	goTime := time.Since(startTime)
	fmt.Printf("  Time: %v for %d iterations (%.2f µs/op)\n",
		goTime, iterations, float64(goTime.Nanoseconds())/float64(iterations)/1000.0)

	//--------------------------------------------------
	// Pongo2 Benchmark
	//--------------------------------------------------
	fmt.Println("\nPongo2 Template Engine:")
	pongoTmpl, err := pongo2.FromString(SimplePongoText)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	// Warm up
	for i := 0; i < 5; i++ {
		pongoTmpl.Execute(pongo2.Context{"name": "World"})
	}

	startTime = time.Now()

	// Run benchmark
	for i := 0; i < iterations; i++ {
		_, err := pongoTmpl.Execute(pongo2.Context{"name": "World"})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	pongoTime := time.Since(startTime)
	fmt.Printf("  Time: %v for %d iterations (%.2f µs/op)\n",
		pongoTime, iterations, float64(pongoTime.Nanoseconds())/float64(iterations)/1000.0)

	//--------------------------------------------------
	// Stick Benchmark
	//--------------------------------------------------
	fmt.Println("\nStick Template Engine:")
	stickEnv := stick.New(nil)

	// Warm up
	var stickBuf bytes.Buffer
	for i := 0; i < 5; i++ {
		stickBuf.Reset()
		stickEnv.Execute(SimpleStickText, &stickBuf, map[string]stick.Value{"name": "World"})
	}

	startTime = time.Now()

	// Run benchmark
	for i := 0; i < iterations; i++ {
		stickBuf.Reset()
		err := stickEnv.Execute(SimpleStickText, &stickBuf, map[string]stick.Value{"name": "World"})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	stickTime := time.Since(startTime)
	fmt.Printf("  Time: %v for %d iterations (%.2f µs/op)\n",
		stickTime, iterations, float64(stickTime.Nanoseconds())/float64(iterations)/1000.0)

	//--------------------------------------------------
	// QuickTemplate Benchmark (simplified)
	//--------------------------------------------------
	fmt.Println("\nQuickTemplate (direct writer):")

	// Warm up
	var qtBuf bytes.Buffer
	for i := 0; i < 5; i++ {
		qtBuf.Reset()
		w := qt.AcquireWriter(&qtBuf)
		w.N().S("Hello ")
		w.N().S("World")
		w.N().S("!")
		qt.ReleaseWriter(w)
	}

	startTime = time.Now()

	// Run benchmark
	for i := 0; i < iterations; i++ {
		qtBuf.Reset()
		w := qt.AcquireWriter(&qtBuf)
		w.N().S("Hello ")
		w.N().S("World")
		w.N().S("!")
		qt.ReleaseWriter(w)
	}

	qtTime := time.Since(startTime)
	fmt.Printf("  Time: %v for %d iterations (%.2f µs/op)\n",
		qtTime, iterations, float64(qtTime.Nanoseconds())/float64(iterations)/1000.0)

	//--------------------------------------------------
	// Results Summary
	//--------------------------------------------------
	fmt.Println("\n==================================================")
	fmt.Println("Benchmark Results Summary")
	fmt.Println("==================================================")
	fmt.Printf("Twig:         %.2f µs/op\n", float64(twigTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("Go Template:  %.2f µs/op\n", float64(goTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("Pongo2:       %.2f µs/op\n", float64(pongoTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("Stick:        %.2f µs/op\n", float64(stickTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("QuickTemplate: %.2f µs/op\n", float64(qtTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Println("\nRelative Performance (lower is better):")
	fmt.Printf("Twig vs Go:   %.2fx\n", float64(twigTime.Nanoseconds())/float64(goTime.Nanoseconds()))
	fmt.Printf("Twig vs Pongo2: %.2fx\n", float64(twigTime.Nanoseconds())/float64(pongoTime.Nanoseconds()))
	fmt.Printf("Twig vs Stick: %.2fx\n", float64(twigTime.Nanoseconds())/float64(stickTime.Nanoseconds()))
	fmt.Printf("Twig vs QuickTemplate: %.2fx\n", float64(twigTime.Nanoseconds())/float64(qtTime.Nanoseconds()))
	fmt.Println("==================================================")
}
