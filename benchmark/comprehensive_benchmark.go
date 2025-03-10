package main

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime"
	"testing"
	"time"

	"github.com/semihalev/twig"
)

// SimpleContext for template data
type SimpleContext struct {
	Title    string
	Name     string
	Age      int
	Friends  []string
	Messages []string
	User     User
	Nested   map[string]interface{}
}

// User struct for template
type User struct {
	Name     string
	Email    string
	IsActive bool
}

// getSimpleContext returns sample data
func getSimpleContext() SimpleContext {
	return SimpleContext{
		Title:    "Template Benchmark",
		Name:     "John Doe",
		Age:      30,
		Friends:  []string{"Alice", "Bob", "Charlie", "Dave", "Eve"},
		Messages: []string{"Hello", "World", "This", "Is", "A", "Test"},
		User: User{
			Name:     "John Doe",
			Email:    "john@example.com",
			IsActive: true,
		},
		Nested: map[string]interface{}{
			"Level1": map[string]interface{}{
				"Level2": map[string]interface{}{
					"Level3": "Nested Value",
				},
			},
		},
	}
}

// Templates
const (
	// Simple templates
	SimpleTemplateText = `<h1>{{ title }}</h1>
<p>Hello {{ name }}!</p>`

	// Medium complexity templates
	MediumTemplateText = `<h1>{{ title }}</h1>
<div class="profile">
  <h2>{{ name }}</h2>
  <p>Age: {{ age }}</p>
  <p>Email: {{ user.Email }}</p>
  {% if user.IsActive %}
    <p>Status: <span class="active">Active</span></p>
  {% else %}
    <p>Status: <span class="inactive">Inactive</span></p>
  {% endif %}
</div>`

	// Complex templates
	ComplexTemplateText = `<!DOCTYPE html>
<html>
<head>
  <title>{{ title }}</title>
</head>
<body>
  <header>
    <h1>{{ title }}</h1>
  </header>
  
  <main>
    <section class="profile">
      <h2>{{ name }}'s Profile</h2>
      <div class="details">
        <p>Age: <strong>{{ age }}</strong></p>
        <p>Email: <a href="mailto:{{ user.Email }}">{{ user.Email }}</a></p>
        <p>Status: 
          {% if user.IsActive %}
            <span class="badge badge-success">Active</span>
          {% else %}
            <span class="badge badge-danger">Inactive</span>
          {% endif %}
        </p>
      </div>
    </section>
    
    <section class="friends">
      <h3>Friends ({{ friends|length }})</h3>
      <ul class="friends-list">
        {% for friend in friends %}
          <li class="friend-item">{{ friend }}</li>
        {% endfor %}
      </ul>
    </section>
    
    <section class="messages">
      <h3>Messages</h3>
      <div class="message-container">
        {% for message in messages %}
          <div class="message {% if loop.first %}message-first{% endif %} {% if loop.last %}message-last{% endif %}">
            <p>{{ loop.index }}. {{ message }}</p>
            {% if not loop.last %}
              <hr class="message-divider">
            {% endif %}
          </div>
        {% endfor %}
      </div>
    </section>
    
    <section class="nested">
      <h3>Nested Data</h3>
      <p>{{ nested.Level1.Level2.Level3 }}</p>
    </section>
  </main>
  
  <footer>
    <p>&copy; {{ "now"|date("Y") }} Template Benchmark</p>
  </footer>
</body>
</html>`
)

// Equivalent templates for Go's html/template
const (
	// Simple Go template
	SimpleGoTemplateText = `<h1>{{ .Title }}</h1>
<p>Hello {{ .Name }}!</p>`

	// Medium complexity Go template
	MediumGoTemplateText = `<h1>{{ .Title }}</h1>
<div class="profile">
  <h2>{{ .Name }}</h2>
  <p>Age: {{ .Age }}</p>
  <p>Email: {{ .User.Email }}</p>
  {{if .User.IsActive }}
    <p>Status: <span class="active">Active</span></p>
  {{else}}
    <p>Status: <span class="inactive">Inactive</span></p>
  {{end}}
</div>`

	// Complex Go template
	ComplexGoTemplateText = `<!DOCTYPE html>
<html>
<head>
  <title>{{ .Title }}</title>
</head>
<body>
  <header>
    <h1>{{ .Title }}</h1>
  </header>
  
  <main>
    <section class="profile">
      <h2>{{ .Name }}'s Profile</h2>
      <div class="details">
        <p>Age: <strong>{{ .Age }}</strong></p>
        <p>Email: <a href="mailto:{{ .User.Email }}">{{ .User.Email }}</a></p>
        <p>Status: 
          {{if .User.IsActive }}
            <span class="badge badge-success">Active</span>
          {{else}}
            <span class="badge badge-danger">Inactive</span>
          {{end}}
        </p>
      </div>
    </section>
    
    <section class="friends">
      <h3>Friends ({{ len .Friends }})</h3>
      <ul class="friends-list">
        {{range .Friends}}
          <li class="friend-item">{{ . }}</li>
        {{end}}
      </ul>
    </section>
    
    <section class="messages">
      <h3>Messages</h3>
      <div class="message-container">
        {{range $i, $message := .Messages}}
          <div class="message {{if eq $i 0}}message-first{{end}} {{if eq $i (sub (len $.Messages) 1)}}message-last{{end}}">
            <p>{{ add $i 1 }}. {{ $message }}</p>
            {{if ne $i (sub (len $.Messages) 1)}}
              <hr class="message-divider">
            {{end}}
          </div>
        {{end}}
      </div>
    </section>
    
    <section class="nested">
      <h3>Nested Data</h3>
      <p>{{ index (index (index .Nested "Level1") "Level2") "Level3" }}</p>
    </section>
  </main>
  
  <footer>
    <p>&copy; {{ currentYear }} Template Benchmark</p>
  </footer>
</body>
</html>`
)

// BenchmarkTwig runs benchmarks for the Twig engine
func BenchmarkTwig(b *testing.B) {
	templates := map[string]string{
		"simple":  SimpleTemplateText,
		"medium":  MediumTemplateText,
		"complex": ComplexTemplateText,
	}

	// Create Twig engine
	engine := twig.New()

	// Register templates
	for name, content := range templates {
		err := engine.RegisterString(name, content)
		if err != nil {
			b.Fatalf("Failed to register template %s: %v", name, err)
		}
	}

	// Get context data
	context := getSimpleContext()
	contextMap := map[string]interface{}{
		"title":    context.Title,
		"name":     context.Name,
		"age":      context.Age,
		"friends":  context.Friends,
		"messages": context.Messages,
		"user":     context.User,
		"nested":   context.Nested,
	}

	// Run benchmarks for each template
	for name := range templates {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, err := engine.Render(name, contextMap)
				if err != nil {
					b.Fatalf("Error rendering template: %v", err)
				}
				if len(result) == 0 {
					b.Fatalf("Empty result")
				}
			}
		})
	}
}

// BenchmarkGoTemplate runs benchmarks for Go's html/template
func BenchmarkGoTemplate(b *testing.B) {
	templates := map[string]string{
		"simple":  SimpleGoTemplateText,
		"medium":  MediumGoTemplateText,
		"complex": ComplexGoTemplateText,
	}

	// Create function map
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"currentYear": func() int {
			return time.Now().Year()
		},
	}

	// Parse templates
	tmpl := template.New("benchmark").Funcs(funcMap)
	for name, content := range templates {
		_, err := tmpl.New(name).Parse(content)
		if err != nil {
			b.Fatalf("Failed to parse template %s: %v", name, err)
		}
	}

	// Get context data
	context := getSimpleContext()

	// Run benchmarks for each template
	for name := range templates {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			var buf bytes.Buffer
			for i := 0; i < b.N; i++ {
				buf.Reset()
				err := tmpl.ExecuteTemplate(&buf, name, context)
				if err != nil {
					b.Fatalf("Error rendering template: %v", err)
				}
				if buf.Len() == 0 {
					b.Fatalf("Empty result")
				}
			}
		})
	}
}

// Additional benchmarks can be added for other engines:
// - BenchmarkPongo2
// - BenchmarkStick
// etc.

// runBenchmarkSuite runs all benchmarks and reports results
func runBenchmarkSuite() {
	fmt.Println("===================================================")
	fmt.Println("           Template Engine Benchmark")
	fmt.Println("===================================================")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("CPU: %d cores\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("===================================================")
	fmt.Println()

	benchmarks := []testing.InternalBenchmark{
		{Name: "BenchmarkTwig/simple", F: func(b *testing.B) { BenchmarkTwig(b) }},
		{Name: "BenchmarkGoTemplate/simple", F: func(b *testing.B) { BenchmarkGoTemplate(b) }},
	}

	// Execute benchmarks
	fmt.Println("Running benchmarks...")
	fmt.Println()

	for _, bm := range benchmarks {
		result := testing.Benchmark(bm.F)
		fmt.Printf("%s:\n", bm.Name)
		fmt.Printf("  Operations: %d\n", result.N)
		fmt.Printf("  Time per op: %s\n", result.T/time.Duration(result.N))
		fmt.Printf("  Bytes per op: %d\n", result.AllocedBytesPerOp())
		fmt.Printf("  Allocs per op: %d\n", result.AllocsPerOp())
		fmt.Println()
	}

	fmt.Println("===================================================")
	fmt.Println("Benchmark complete!")
	fmt.Println("===================================================")
}

func main() {
	runBenchmarkSuite()
}
