package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/semihalev/twig"
)

const (
	// Macro definition template
	MacroTemplate = `
{% macro input(name, value = '', type = 'text') %}
    <input type="{{ type }}" name="{{ name }}" value="{{ value|e }}" />
{% endmacro %}

{% macro form(action, method = "POST") %}
    <form action="{{ action }}" method="{{ method }}">
        {{ caller() }}
    </form>
{% endmacro %}

{{ input('username') }}
{{ input('password', '', 'password') }}
{{ input('submit', 'Login', 'submit') }}
`

	// Import macro template
	ImportTemplate = `
{% macro input(name, value = '', type = 'text') %}
    <input type="{{ type }}" name="{{ name }}" value="{{ value|e }}" />
{% endmacro %}

{% macro form(action, method = "POST") %}
    <form action="{{ action }}" method="{{ method }}">
        {{ caller() }}
    </form>
{% endmacro %}
`

	// Template that imports macros
	UseMacrosTemplate = `
{% import "macros.twig" as forms %}

{{ forms.input('username') }}
{{ forms.input('password', '', 'password') }}
{{ forms.input('submit', 'Login', 'submit') }}
`

	// Nested macro calls
	NestedMacrosTemplate = `
{% macro input(name, value = '', type = 'text') %}
    <input type="{{ type }}" name="{{ name }}" value="{{ value|e }}" />
{% endmacro %}

{% macro field(name, value = '', type = 'text', label = '') %}
    <div class="field">
        {% if label %}
            <label for="{{ name }}">{{ label }}</label>
        {% endif %}
        {{ input(name, value, type) }}
    </div>
{% endmacro %}

{{ field('username', '', 'text', 'Username') }}
{{ field('password', '', 'password', 'Password') }}
{{ field('submit', 'Login', 'submit') }}
`
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("Twig Macro Performance Benchmark")
	fmt.Println("==================================================")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("CPU: %d cores\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("==================================================")
	fmt.Println()

	// Create Twig engine
	twigEngine := twig.New()

	// Register templates
	err := twigEngine.RegisterString("macros", MacroTemplate)
	if err != nil {
		fmt.Printf("Error registering macro template: %v\n", err)
		return
	}

	err = twigEngine.RegisterString("macros.twig", ImportTemplate)
	if err != nil {
		fmt.Printf("Error registering import template: %v\n", err)
		return
	}

	err = twigEngine.RegisterString("use_macros", UseMacrosTemplate)
	if err != nil {
		fmt.Printf("Error registering use macros template: %v\n", err)
		return
	}

	err = twigEngine.RegisterString("nested_macros", NestedMacrosTemplate)
	if err != nil {
		fmt.Printf("Error registering nested macros template: %v\n", err)
		return
	}

	// Benchmark direct macro usage
	fmt.Println("1. Direct Macro Usage")
	fmt.Println("--------------------------------------------------")
	iterations := 10000

	// Warm up
	for i := 0; i < 5; i++ {
		twigEngine.Render("macros", nil)
	}

	startTime := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := twigEngine.Render("macros", nil)
		if err != nil {
			fmt.Printf("Error rendering macro template: %v\n", err)
			return
		}
	}
	directMacroTime := time.Since(startTime)
	fmt.Printf("  %d iterations in %v (%.2f µs/op)\n", 
		iterations, directMacroTime, float64(directMacroTime.Nanoseconds())/float64(iterations)/1000.0)

	// Benchmark imported macro usage
	fmt.Println("\n2. Imported Macro Usage")
	fmt.Println("--------------------------------------------------")

	// Warm up
	for i := 0; i < 5; i++ {
		twigEngine.Render("use_macros", nil)
	}

	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := twigEngine.Render("use_macros", nil)
		if err != nil {
			fmt.Printf("Error rendering imported macro template: %v\n", err)
			return
		}
	}
	importedMacroTime := time.Since(startTime)
	fmt.Printf("  %d iterations in %v (%.2f µs/op)\n", 
		iterations, importedMacroTime, float64(importedMacroTime.Nanoseconds())/float64(iterations)/1000.0)

	// Benchmark nested macro calls
	fmt.Println("\n3. Nested Macro Calls")
	fmt.Println("--------------------------------------------------")

	// Warm up
	for i := 0; i < 5; i++ {
		twigEngine.Render("nested_macros", nil)
	}

	startTime = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := twigEngine.Render("nested_macros", nil)
		if err != nil {
			fmt.Printf("Error rendering nested macro template: %v\n", err)
			return
		}
	}
	nestedMacroTime := time.Since(startTime)
	fmt.Printf("  %d iterations in %v (%.2f µs/op)\n", 
		iterations, nestedMacroTime, float64(nestedMacroTime.Nanoseconds())/float64(iterations)/1000.0)

	// Summary
	fmt.Println("\n==================================================")
	fmt.Println("Macro Benchmark Results Summary")
	fmt.Println("==================================================")
	fmt.Printf("Direct macro usage:    %.2f µs/op\n", float64(directMacroTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("Imported macro usage:  %.2f µs/op\n", float64(importedMacroTime.Nanoseconds())/float64(iterations)/1000.0)
	fmt.Printf("Nested macro calls:    %.2f µs/op\n", float64(nestedMacroTime.Nanoseconds())/float64(iterations)/1000.0)
	
	fmt.Println("\nRelative Performance:")
	fmt.Printf("Imported vs Direct:    %.2fx\n", float64(importedMacroTime.Nanoseconds())/float64(directMacroTime.Nanoseconds()))
	fmt.Printf("Nested vs Direct:      %.2fx\n", float64(nestedMacroTime.Nanoseconds())/float64(directMacroTime.Nanoseconds()))
	fmt.Println("==================================================")
}