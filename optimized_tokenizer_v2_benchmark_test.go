package twig

import (
	"strconv"
	"strings"
	"testing"
)

// Test templates of different sizes
var (
	smallTemplate = `<div>{{ simple_var }}</div>`
	
	mediumTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>{{ page_title }}</title>
</head>
<body>
    <h1>{{ page_title }}</h1>
    <div class="content">
        {% if show_message %}
            <p>{{ message }}</p>
        {% endif %}
    </div>
</body>
</html>`
)

// Create a large template by repeating patterns
func generateLargeTemplate() string {
	var sb strings.Builder
	
	sb.WriteString(`<!DOCTYPE html><html><head><title>{{ page_title }}</title></head><body>`)
	
	// Add many tags to make it large
	for i := 0; i < 500; i++ {
		sb.WriteString(`<div>{{ var_`)
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(` }}</div>{% if condition_`)
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(` %}<p>{{ message_`)
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(` }}</p>{% endif %}`)
	}
	
	sb.WriteString(`</body></html>`)
	return sb.String()
}

var largeTemplate = generateLargeTemplate()

// BenchmarkTokenizerV2Comparison compares all tokenizer implementations
func BenchmarkTokenizerV2Comparison(b *testing.B) {
	templates := []struct {
		name     string
		template string
	}{
		{"Small", smallTemplate},
		{"Medium", mediumTemplate},
		{"Large", largeTemplate},
	}
	
	for _, tmpl := range templates {
		// Original tokenizer
		b.Run(tmpl.name+"_Original", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				tokenizer := GetTokenizer(tmpl.template, 0)
				tokens, _ := tokenizer.TokenizeHtmlPreserving()
				_ = tokens
				ReleaseTokenizer(tokenizer)
			}
		})
		
		// Optimized tokenizer (V1)
		b.Run(tmpl.name+"_OptimizedV1", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				tokenizer := NewOptimizedTokenizer()
				tokenizer.baseTokenizer.source = tmpl.template
				tokenizer.baseTokenizer.position = 0
				tokenizer.baseTokenizer.line = 1
				
				tokens, _ := tokenizer.TokenizeHtmlPreserving()
				_ = tokens
				
				ReleaseOptimizedTokenizer(tokenizer)
			}
		})
		
		// Optimized tokenizer V2
		b.Run(tmpl.name+"_OptimizedV2", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				// Use hybrid approach based on size
				if len(tmpl.template) > 4096 {
					tokenizer := GetOptimizedTokenizerV2()
					tokenizer.Initialize(tmpl.template)
					
					tokens, _ := tokenizer.TokenizeHtmlPreserving()
					_ = tokens
					
					ReleaseOptimizedTokenizerV2(tokenizer)
				} else {
					tokenizer := GetTokenizer(tmpl.template, 0)
					tokens, _ := tokenizer.TokenizeHtmlPreserving()
					_ = tokens
					ReleaseTokenizer(tokenizer)
				}
			}
		})
	}
}

// BenchmarkTagDetector benchmarks the tag detector
func BenchmarkTagDetector(b *testing.B) {
	templates := []struct {
		name     string
		template string
	}{
		{"Small", smallTemplate},
		{"Medium", mediumTemplate},
		{"Large", largeTemplate},
	}
	
	for _, tmpl := range templates {
		b.Run(tmpl.name, func(b *testing.B) {
			source := tmpl.template
			b.ReportAllocs()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				pos := 0
				for {
					tag := FindNextTag(source, pos)
					if tag.Type == TAG_NONE {
						break
					}
					pos = tag.Position + tag.Length
				}
			}
		})
	}
}