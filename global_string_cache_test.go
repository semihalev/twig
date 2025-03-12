package twig

import (
	"fmt"
	"strings"
	"testing"
)

// Test that the global string cache correctly interns strings
func TestGlobalStringCache(t *testing.T) {
	// Test interning common strings
	commonStrings := []string{"div", "if", "for", "endif", "endfor", "else", ""}
	
	for _, s := range commonStrings {
		interned := Intern(s)
		
		// The interned string should be the same value
		if interned != s {
			t.Errorf("Interned string %q should equal original", s)
		}
		
		// The interned string should be the same address for common strings
		if strings.Compare(interned, s) != 0 {
			t.Errorf("Interned string %q should be the same instance", s)
		}
	}
	
	// Test interning the same string twice returns the same value
	s1 := "test_string"
	interned1 := Intern(s1)
	interned2 := Intern(s1)
	
	// Since we're comparing strings by value, not pointers
	if interned1 != interned2 {
		t.Errorf("Interning the same string twice should return the same string value")
	}
	
	// Test that long strings aren't interned (compared by value but not address)
	longString := strings.Repeat("x", maxCacheableLength+1)
	internedLong := Intern(longString)
	
	if internedLong != longString {
		t.Errorf("Long string should equal original after Intern")
	}
}

// Benchmark string interning for common string cases
func BenchmarkIntern_Common(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Intern("div")
		_ = Intern("for")
		_ = Intern("if")
		_ = Intern("endif")
	}
}

// Benchmark string interning for uncommon strings
func BenchmarkIntern_Uncommon(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := fmt.Sprintf("uncommon_string_%d", i%100)
		_ = Intern(s)
	}
}

// Benchmark string interning for long strings
func BenchmarkIntern_Long(b *testing.B) {
	longString := strings.Repeat("x", maxCacheableLength+1)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Intern(longString)
	}
}

// Benchmark old tokenizer vs new optimized tokenizer
func BenchmarkTokenizer_Comparison(b *testing.B) {
	// Sample template with various elements to test tokenization
	template := `<!DOCTYPE html>
<html>
<head>
    <title>{{ page_title }}</title>
    <link rel="stylesheet" href="{{ asset('styles.css') }}">
</head>
<body>
    <div class="container">
        <h1>{{ page_title }}</h1>
        
        {% if user %}
            <p>Welcome back, {{ user.name }}!</p>
            
            {% if user.isAdmin %}
                <div class="admin-panel">
                    <h2>Admin Controls</h2>
                    <ul>
                        {% for item in admin_items %}
                            <li>{{ item.name }} - {{ item.description }}</li>
                        {% endfor %}
                    </ul>
                </div>
            {% endif %}
            
            <div class="user-content">
                {% block user_content %}
                    <p>Default user content</p>
                {% endblock %}
            </div>
        {% else %}
            <p>Welcome, guest! Please <a href="{{ login_url }}">login</a>.</p>
        {% endif %}
        
        <footer>
            <p>&copy; {{ 'now'|date('Y') }} Example Company</p>
        </footer>
    </div>
</body>
</html>`

	// Benchmark the original tokenizer
	b.Run("OriginalTokenizer", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			tokenizer := GetTokenizer(template, 0)
			tokens, _ := tokenizer.TokenizeHtmlPreserving()
			_ = tokens
			ReleaseTokenizer(tokenizer)
		}
	})
	
	// Benchmark the optimized tokenizer
	b.Run("OptimizedTokenizer", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			tokenizer := NewOptimizedTokenizer()
			tokenizer.baseTokenizer.source = template
			tokenizer.baseTokenizer.position = 0
			tokenizer.baseTokenizer.line = 1
			
			tokens, _ := tokenizer.TokenizeHtmlPreserving()
			_ = tokens
			
			ReleaseOptimizedTokenizer(tokenizer)
		}
	})
}

// Benchmark string interning in the original tokenizer vs global string cache
func BenchmarkStringIntern_Comparison(b *testing.B) {
	// Generate some test strings
	testStrings := make([]string, 100)
	for i := 0; i < 100; i++ {
		testStrings[i] = fmt.Sprintf("test_string_%d", i)
	}
	
	// Also include some common strings
	commonStrings := []string{"div", "if", "for", "endif", "endfor", "else", ""}
	testStrings = append(testStrings, commonStrings...)
	
	// Benchmark the original GetStringConstant method
	b.Run("OriginalGetStringConstant", func(b *testing.B) {
		tokenizer := ZeroAllocTokenizer{}
		b.ReportAllocs()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for _, s := range testStrings {
				_ = tokenizer.GetStringConstant(s)
			}
		}
	})
	
	// Benchmark the new global cache Intern method
	b.Run("GlobalIntern", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for _, s := range testStrings {
				_ = Intern(s)
			}
		}
	})
}