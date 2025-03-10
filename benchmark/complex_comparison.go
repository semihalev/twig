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
)

// Templates for different engines
const (
	// Simple templates
	TwigSimpleTemplate = `Hello {{ name }}!`
	GoSimpleTemplate = `Hello {{ .Name }}!`
	PongoSimpleTemplate = `Hello {{ name }}!`
	StickSimpleTemplate = `Hello {{ name }}!`

	// Medium templates with conditions
	TwigMediumTemplate = `
<div>
  <h1>{{ title }}</h1>
  <p>Welcome {{ name }}!</p>
  {% if is_admin %}
    <p>You have admin access.</p>
  {% else %}
    <p>You have user access.</p>
  {% endif %}
</div>`

	GoMediumTemplate = `
<div>
  <h1>{{ .Title }}</h1>
  <p>Welcome {{ .Name }}!</p>
  {{if .IsAdmin}}
    <p>You have admin access.</p>
  {{else}}
    <p>You have user access.</p>
  {{end}}
</div>`

	PongoMediumTemplate = `
<div>
  <h1>{{ title }}</h1>
  <p>Welcome {{ name }}!</p>
  {% if is_admin %}
    <p>You have admin access.</p>
  {% else %}
    <p>You have user access.</p>
  {% endif %}
</div>`

	StickMediumTemplate = `
<div>
  <h1>{{ title }}</h1>
  <p>Welcome {{ name }}!</p>
  {% if is_admin %}
    <p>You have admin access.</p>
  {% else %}
    <p>You have user access.</p>
  {% endif %}
</div>`

	// Complex templates with loops
	TwigComplexTemplate = `
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

	GoComplexTemplate = `
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

	PongoComplexTemplate = `
<div>
  <h1>{{ title }}</h1>
  <p>Welcome {{ name }}!</p>
  
  <h2>Items</h2>
  <ul>
    {% for item in items %}
      <li class="{% cycle 'odd' 'even' %}">
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

	StickComplexTemplate = `
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
)

// Item struct for template data
type Item struct {
	Name      string
	Price     float64
	Available bool
}

// Context for Go templates
type GoContext struct {
	Name    string
	Title   string
	IsAdmin bool
	Items   []Item
}

func main() {
	fmt.Println("=========================================================")
	fmt.Println("Template Engine Comprehensive Comparison")
	fmt.Println("=========================================================")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("CPU: %d cores\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("=========================================================")
	fmt.Println()

	// Define common data
	name := "John"
	title := "Product List"
	isAdmin := true
	items := []Item{
		{Name: "Phone", Price: 599.99, Available: true},
		{Name: "Laptop", Price: 1299.99, Available: true},
		{Name: "Tablet", Price: 399.99, Available: false},
		{Name: "Watch", Price: 199.99, Available: true},
		{Name: "Headphones", Price: 99.99, Available: false},
	}

	// Convert items for different template engines
	goItems := items
	pongoItems := make([]map[string]interface{}, len(items))
	stickItems := make([]map[string]stick.Value, len(items))
	
	for i, item := range items {
		pongoItems[i] = map[string]interface{}{
			"name":      item.Name,
			"price":     item.Price,
			"available": item.Available,
		}
		
		stickItems[i] = map[string]stick.Value{
			"name":      item.Name,
			"price":     item.Price,
			"available": item.Available,
		}
	}

	// Iterations for benchmarks
	iterations := 10000

	// Initialize template engines
	twigEngine := twig.New()
	twigEngine.RegisterString("simple", TwigSimpleTemplate)
	twigEngine.RegisterString("medium", TwigMediumTemplate)
	twigEngine.RegisterString("complex", TwigComplexTemplate)

	// Go template functions
	funcMap := template.FuncMap{
		"isOdd": func(i int) bool {
			return i%2 == 1
		},
	}
	
	goSimpleTmpl := template.Must(template.New("simple").Parse(GoSimpleTemplate))
	goMediumTmpl := template.Must(template.New("medium").Funcs(funcMap).Parse(GoMediumTemplate))
	goComplexTmpl := template.Must(template.New("complex").Funcs(funcMap).Parse(GoComplexTemplate))

	// Pongo2 templates
	pongoSimpleTmpl := pongo2.Must(pongo2.FromString(PongoSimpleTemplate))
	pongoMediumTmpl := pongo2.Must(pongo2.FromString(PongoMediumTemplate))
	pongoComplexTmpl := pongo2.Must(pongo2.FromString(PongoComplexTemplate))

	// Stick templates
	stickEnv := stick.New(nil)

	// Data contexts
	twigSimpleContext := map[string]interface{}{
		"name": name,
	}
	
	twigMediumContext := map[string]interface{}{
		"name":     name,
		"title":    title,
		"is_admin": isAdmin,
	}
	
	twigComplexContext := map[string]interface{}{
		"name":     name,
		"title":    title,
		"is_admin": isAdmin,
		"items":    items,
	}

	goSimpleContext := GoContext{
		Name: name,
	}
	
	goMediumContext := GoContext{
		Name:    name,
		Title:   title,
		IsAdmin: isAdmin,
	}
	
	goComplexContext := GoContext{
		Name:    name,
		Title:   title,
		IsAdmin: isAdmin,
		Items:   goItems,
	}

	pongoSimpleContext := pongo2.Context{
		"name": name,
	}
	
	pongoMediumContext := pongo2.Context{
		"name":     name,
		"title":    title,
		"is_admin": isAdmin,
	}
	
	pongoComplexContext := pongo2.Context{
		"name":     name,
		"title":    title,
		"is_admin": isAdmin,
		"items":    pongoItems,
	}

	stickSimpleContext := map[string]stick.Value{
		"name": name,
	}
	
	stickMediumContext := map[string]stick.Value{
		"name":     name,
		"title":    title,
		"is_admin": isAdmin,
	}
	
	stickComplexContext := map[string]stick.Value{
		"name":     name,
		"title":    title,
		"is_admin": isAdmin,
		"items":    stickItems,
	}

	// Run benchmarks for each template
	fmt.Println("SIMPLE TEMPLATE BENCHMARKS")
	fmt.Println("----------------------------------------------------------")
	
	// Twig Simple
	// Warm up
	for i := 0; i < 5; i++ {
		twigEngine.Render("simple", twigSimpleContext)
	}
	
	startTime := time.Now()
	for i := 0; i < iterations; i++ {
		twigEngine.Render("simple", twigSimpleContext)
	}
	twigSimpleTime := time.Since(startTime)
	fmt.Printf("Twig Simple:    %v for %d iterations (%.2f µs/op)\n", 
		twigSimpleTime, iterations, float64(twigSimpleTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Go Simple
	var buf bytes.Buffer
	for i := 0; i < 5; i++ {
		buf.Reset()
		goSimpleTmpl.Execute(&buf, goSimpleContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		buf.Reset()
		goSimpleTmpl.Execute(&buf, goSimpleContext)
	}
	goSimpleTime := time.Since(startTime)
	fmt.Printf("Go Simple:      %v for %d iterations (%.2f µs/op)\n", 
		goSimpleTime, iterations, float64(goSimpleTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Pongo2 Simple
	for i := 0; i < 5; i++ {
		pongoSimpleTmpl.Execute(pongoSimpleContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		pongoSimpleTmpl.Execute(pongoSimpleContext)
	}
	pongoSimpleTime := time.Since(startTime)
	fmt.Printf("Pongo2 Simple:  %v for %d iterations (%.2f µs/op)\n", 
		pongoSimpleTime, iterations, float64(pongoSimpleTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Stick Simple
	var stickBuf bytes.Buffer
	for i := 0; i < 5; i++ {
		stickBuf.Reset()
		stickEnv.Execute(StickSimpleTemplate, &stickBuf, stickSimpleContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		stickBuf.Reset()
		stickEnv.Execute(StickSimpleTemplate, &stickBuf, stickSimpleContext)
	}
	stickSimpleTime := time.Since(startTime)
	fmt.Printf("Stick Simple:   %v for %d iterations (%.2f µs/op)\n\n", 
		stickSimpleTime, iterations, float64(stickSimpleTime.Nanoseconds())/float64(iterations)/1000.0)
	
	fmt.Println("MEDIUM TEMPLATE BENCHMARKS (conditions)")
	fmt.Println("----------------------------------------------------------")
	
	// Twig Medium
	for i := 0; i < 5; i++ {
		twigEngine.Render("medium", twigMediumContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		twigEngine.Render("medium", twigMediumContext)
	}
	twigMediumTime := time.Since(startTime)
	fmt.Printf("Twig Medium:    %v for %d iterations (%.2f µs/op)\n", 
		twigMediumTime, iterations, float64(twigMediumTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Go Medium
	for i := 0; i < 5; i++ {
		buf.Reset()
		goMediumTmpl.Execute(&buf, goMediumContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		buf.Reset()
		goMediumTmpl.Execute(&buf, goMediumContext)
	}
	goMediumTime := time.Since(startTime)
	fmt.Printf("Go Medium:      %v for %d iterations (%.2f µs/op)\n", 
		goMediumTime, iterations, float64(goMediumTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Pongo2 Medium
	for i := 0; i < 5; i++ {
		pongoMediumTmpl.Execute(pongoMediumContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		pongoMediumTmpl.Execute(pongoMediumContext)
	}
	pongoMediumTime := time.Since(startTime)
	fmt.Printf("Pongo2 Medium:  %v for %d iterations (%.2f µs/op)\n", 
		pongoMediumTime, iterations, float64(pongoMediumTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Stick Medium
	for i := 0; i < 5; i++ {
		stickBuf.Reset()
		stickEnv.Execute(StickMediumTemplate, &stickBuf, stickMediumContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		stickBuf.Reset()
		stickEnv.Execute(StickMediumTemplate, &stickBuf, stickMediumContext)
	}
	stickMediumTime := time.Since(startTime)
	fmt.Printf("Stick Medium:   %v for %d iterations (%.2f µs/op)\n\n", 
		stickMediumTime, iterations, float64(stickMediumTime.Nanoseconds())/float64(iterations)/1000.0)
	
	fmt.Println("COMPLEX TEMPLATE BENCHMARKS (loops and conditionals)")
	fmt.Println("----------------------------------------------------------")
	
	// Twig Complex
	for i := 0; i < 5; i++ {
		twigEngine.Render("complex", twigComplexContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		twigEngine.Render("complex", twigComplexContext)
	}
	twigComplexTime := time.Since(startTime)
	fmt.Printf("Twig Complex:   %v for %d iterations (%.2f µs/op)\n", 
		twigComplexTime, iterations, float64(twigComplexTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Go Complex
	for i := 0; i < 5; i++ {
		buf.Reset()
		goComplexTmpl.Execute(&buf, goComplexContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		buf.Reset()
		goComplexTmpl.Execute(&buf, goComplexContext)
	}
	goComplexTime := time.Since(startTime)
	fmt.Printf("Go Complex:     %v for %d iterations (%.2f µs/op)\n", 
		goComplexTime, iterations, float64(goComplexTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Pongo2 Complex
	for i := 0; i < 5; i++ {
		pongoComplexTmpl.Execute(pongoComplexContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		pongoComplexTmpl.Execute(pongoComplexContext)
	}
	pongoComplexTime := time.Since(startTime)
	fmt.Printf("Pongo2 Complex: %v for %d iterations (%.2f µs/op)\n", 
		pongoComplexTime, iterations, float64(pongoComplexTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Stick Complex
	for i := 0; i < 5; i++ {
		stickBuf.Reset()
		stickEnv.Execute(StickComplexTemplate, &stickBuf, stickComplexContext)
	}
	
	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		stickBuf.Reset()
		stickEnv.Execute(StickComplexTemplate, &stickBuf, stickComplexContext)
	}
	stickComplexTime := time.Since(startTime)
	fmt.Printf("Stick Complex:  %v for %d iterations (%.2f µs/op)\n\n", 
		stickComplexTime, iterations, float64(stickComplexTime.Nanoseconds())/float64(iterations)/1000.0)
	
	// Summary
	fmt.Println("=========================================================")
	fmt.Println("BENCHMARK RESULTS SUMMARY (µs per operation)")
	fmt.Println("=========================================================")
	fmt.Printf("Engine     | Simple  | Medium  | Complex\n")
	fmt.Printf("-----------|---------|---------|----------\n")
	fmt.Printf("Twig       | %.2f    | %.2f    | %.2f\n", 
		float64(twigSimpleTime.Nanoseconds())/float64(iterations)/1000.0,
		float64(twigMediumTime.Nanoseconds())/float64(iterations)/1000.0,
		float64(twigComplexTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("Go Template| %.2f    | %.2f    | %.2f\n", 
		float64(goSimpleTime.Nanoseconds())/float64(iterations)/1000.0,
		float64(goMediumTime.Nanoseconds())/float64(iterations)/1000.0,
		float64(goComplexTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("Pongo2     | %.2f    | %.2f    | %.2f\n", 
		float64(pongoSimpleTime.Nanoseconds())/float64(iterations)/1000.0,
		float64(pongoMediumTime.Nanoseconds())/float64(iterations)/1000.0,
		float64(pongoComplexTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("Stick      | %.2f    | %.2f    | %.2f\n\n", 
		float64(stickSimpleTime.Nanoseconds())/float64(iterations)/1000.0,
		float64(stickMediumTime.Nanoseconds())/float64(iterations)/1000.0,
		float64(stickComplexTime.Nanoseconds())/float64(iterations)/1000.0)
	
	fmt.Println("RELATIVE PERFORMANCE TO TWIG (smaller is better for Twig)")
	fmt.Println("=========================================================")
	fmt.Printf("Comparison | Simple  | Medium  | Complex\n")
	fmt.Printf("-----------|---------|---------|----------\n")
	fmt.Printf("Twig vs Go | %.2fx    | %.2fx    | %.2fx\n", 
		float64(twigSimpleTime.Nanoseconds())/float64(goSimpleTime.Nanoseconds()),
		float64(twigMediumTime.Nanoseconds())/float64(goMediumTime.Nanoseconds()),
		float64(twigComplexTime.Nanoseconds())/float64(goComplexTime.Nanoseconds()))
	fmt.Printf("Twig vs Pongo2 | %.2fx    | %.2fx    | %.2fx\n", 
		float64(twigSimpleTime.Nanoseconds())/float64(pongoSimpleTime.Nanoseconds()),
		float64(twigMediumTime.Nanoseconds())/float64(pongoMediumTime.Nanoseconds()),
		float64(twigComplexTime.Nanoseconds())/float64(pongoComplexTime.Nanoseconds()))
	fmt.Printf("Twig vs Stick | %.2fx    | %.2fx    | %.2fx\n", 
		float64(twigSimpleTime.Nanoseconds())/float64(stickSimpleTime.Nanoseconds()),
		float64(twigMediumTime.Nanoseconds())/float64(stickMediumTime.Nanoseconds()),
		float64(twigComplexTime.Nanoseconds())/float64(stickComplexTime.Nanoseconds()))
	fmt.Println("=========================================================")
}