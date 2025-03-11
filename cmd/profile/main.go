// Package main provides a simple utility to run memory profiling on the Twig template engine
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/semihalev/twig"
)

func main() {
	// Command-line flags
	cpuProfile := flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile := flag.String("memprofile", "", "write memory profile to file")
	complexityLevel := flag.Int("complexity", 2, "template complexity level (1-3)")
	iterations := flag.Int("iterations", 1000, "number of template renders to perform")
	flag.Parse()

	// CPU profiling if requested
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Run the appropriate benchmark based on complexity level
	var totalTime time.Duration
	var totalAllocBytes uint64
	var numGC uint32

	// Record memory stats before
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	switch *complexityLevel {
	case 1:
		totalTime = runSimpleTemplateBenchmark(*iterations)
	case 2:
		totalTime = runMediumTemplateBenchmark(*iterations)
	case 3:
		totalTime = runComplexTemplateBenchmark(*iterations)
	default:
		log.Fatalf("Invalid complexity level: %d (must be 1-3)", *complexityLevel)
	}

	// Record memory stats after
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Calculate allocation statistics
	totalAllocBytes = memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc
	numGC = memStatsAfter.NumGC - memStatsBefore.NumGC

	// Report results
	fmt.Printf("=== Twig Memory Profiling Results ===\n")
	fmt.Printf("Complexity Level: %d\n", *complexityLevel)
	fmt.Printf("Total Iterations: %d\n", *iterations)
	fmt.Printf("Total Time: %v\n", totalTime)
	fmt.Printf("Time per Iteration: %v\n", totalTime/time.Duration(*iterations))
	fmt.Printf("Total Memory Allocated: %d bytes\n", totalAllocBytes)
	fmt.Printf("Memory per Iteration: %d bytes\n", totalAllocBytes/uint64(*iterations))
	fmt.Printf("Number of GCs: %d\n", numGC)

	// Memory profiling if requested
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // Get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

// runSimpleTemplateBenchmark runs a benchmark with a simple template
func runSimpleTemplateBenchmark(iterations int) time.Duration {
	engine := twig.New()
	err := engine.RegisterString("simple", "Hello, {{ name }}!")
	if err != nil {
		log.Fatalf("Error registering template: %v", err)
	}

	context := map[string]interface{}{
		"name": "World",
	}

	startTime := time.Now()
	for i := 0; i < iterations; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("simple")
		err := template.RenderTo(&buf, context)
		if err != nil {
			log.Fatalf("Error rendering template: %v", err)
		}
	}
	return time.Since(startTime)
}

// runMediumTemplateBenchmark runs a benchmark with a medium complexity template
func runMediumTemplateBenchmark(iterations int) time.Duration {
	engine := twig.New()
	templateContent := `
<div class="profile">
  <h1>{{ user.name }}</h1>
  <p>Age: {{ user.age }}</p>
  {% if user.bio %}
    <div class="bio">{{ user.bio|capitalize }}</div>
  {% else %}
    <div class="bio">No bio available</div>
  {% endif %}
  <ul class="skills">
    {% for skill in user.skills %}
      <li>{{ skill.name }} ({{ skill.level }})</li>
    {% endfor %}
  </ul>
</div>
`
	err := engine.RegisterString("medium", templateContent)
	if err != nil {
		log.Fatalf("Error registering template: %v", err)
	}

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John Doe",
			"age":  30,
			"bio":  "web developer and open source enthusiast",
			"skills": []map[string]interface{}{
				{"name": "Go", "level": "Advanced"},
				{"name": "JavaScript", "level": "Intermediate"},
				{"name": "CSS", "level": "Beginner"},
				{"name": "HTML", "level": "Advanced"},
			},
		},
	}

	startTime := time.Now()
	for i := 0; i < iterations; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("medium")
		err := template.RenderTo(&buf, context)
		if err != nil {
			log.Fatalf("Error rendering template: %v", err)
		}
	}
	return time.Since(startTime)
}

// runComplexTemplateBenchmark runs a benchmark with a complex template
func runComplexTemplateBenchmark(iterations int) time.Duration {
	engine := twig.New()
	
	// Base template
	baseTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
    {% block styles %}
        <style>
            body { font-family: Arial, sans-serif; }
        </style>
    {% endblock %}
</head>
<body>
    <header>{% block header %}Default Header{% endblock %}</header>
    <main>{% block content %}Default Content{% endblock %}</main>
    <footer>{% block footer %}© {{ "now"|date("Y") }} Sample Site{% endblock %}</footer>
</body>
</html>
`
	err := engine.RegisterString("base", baseTemplate)
	if err != nil {
		log.Fatalf("Error registering template: %v", err)
	}
	
	// Macro template
	macroTemplate := `
{% macro renderProduct(product) %}
    <div class="product">
        <h3>{{ product.name }}</h3>
        <p>{{ product.description|capitalize }}</p>
        <div class="price">{{ product.price|format("$%.2f") }}</div>
        {% if product.tags %}
            <div class="tags">
                {% for tag in product.tags %}
                    <span class="tag">{{ tag }}</span>
                {% endfor %}
            </div>
        {% endif %}
    </div>
{% endmacro %}
`
	err = engine.RegisterString("macros", macroTemplate)
	if err != nil {
		log.Fatalf("Error registering template: %v", err)
	}
	
	// Page template
	pageTemplate := `
{% extends "base" %}
{% import "macros" as components %}

{% block title %}{{ page.title }} - {{ parent() }}{% endblock %}

{% block styles %}
    {{ parent() }}
    <style>
        .product { border: 1px solid #ddd; padding: 10px; margin-bottom: 10px; }
        .price { font-weight: bold; color: #c00; }
        .tag { background: #eee; padding: 2px 5px; margin-right: 5px; }
    </style>
{% endblock %}

{% block header %}
    <h1>{{ page.title }}</h1>
    <nav>
        {% for item in navigation %}
            <a href="{{ item.url }}">{{ item.text }}</a>
            {% if not loop.last %} | {% endif %}
        {% endfor %}
    </nav>
{% endblock %}

{% block content %}
    <div class="products">
        <h2>Product List ({{ products|length }} items)</h2>
        
        {% for product in products %}
            {{ components.renderProduct(product) }}
        {% endfor %}
        
        {% set total = 0 %}
        {% for product in products %}
            {% set total = total + product.price %}
        {% endfor %}
        
        <div class="summary">
            <p>Total products: {{ products|length }}</p>
            <p>Average price: {{ (total / products|length)|format("$%.2f") }}</p>
            <p>Price range: {{ products|map(p => p.price)|sort|first|format("$%.2f") }} - {{ products|map(p => p.price)|sort|last|format("$%.2f") }}</p>
        </div>
    </div>
{% endblock %}
`
	err = engine.RegisterString("page", pageTemplate)
	if err != nil {
		log.Fatalf("Error registering template: %v", err)
	}

	// Create a complex context with various data types
	products := make([]map[string]interface{}, 20)
	for i := 0; i < 20; i++ {
		products[i] = map[string]interface{}{
			"id":          i + 1,
			"name":        "Product " + strconv.Itoa(i+1),
			"description": "This is product " + strconv.Itoa(i+1) + " with detailed information.",
			"price":       15.0 + float64(i)*1.5,
			"tags":        []string{"tag1", "tag2", "tag3"}[0:1+(i%3)],
		}
	}
	
	context := map[string]interface{}{
		"page": map[string]interface{}{
			"title": "Product Catalog",
		},
		"navigation": []map[string]interface{}{
			{"url": "/", "text": "Home"},
			{"url": "/products", "text": "Products"},
			{"url": "/about", "text": "About"},
			{"url": "/contact", "text": "Contact"},
		},
		"products": products,
	}

	startTime := time.Now()
	for i := 0; i < iterations; i++ {
		var buf bytes.Buffer
		template, _ := engine.Load("page")
		err := template.RenderTo(&buf, context)
		if err != nil {
			log.Fatalf("Error rendering template: %v", err)
		}
	}
	return time.Since(startTime)
}