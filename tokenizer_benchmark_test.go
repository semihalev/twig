package twig

import (
	"testing"
)

func BenchmarkHtmlPreservingTokenize(b *testing.B) {
	// A sample template with HTML and Twig tags
	source := `<!DOCTYPE html>
<html>
<head>
    <title>{{ title }}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{ asset_url('styles.css') }}">
</head>
<body>
    <header>
        <h1>{{ page.title }}</h1>
        <nav>
            <ul>
                {% for item in menu %}
                    <li><a href="{{ item.url }}">{{ item.label }}</a></li>
                {% endfor %}
            </ul>
        </nav>
    </header>
    
    <main>
        {% if content %}
            <article>
                {{ content|raw }}
            </article>
        {% else %}
            <p>No content available.</p>
        {% endif %}
        
        {% block sidebar %}
            <aside>
                {% include "sidebar.twig" with {items: sidebar_items} %}
            </aside>
        {% endblock %}
    </main>
    
    <footer>
        <p>&copy; {{ "now"|date("Y") }} {{ site_name }}</p>
    </footer>
</body>
</html>`

	parser := &Parser{source: source}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.htmlPreservingTokenize()
	}
}

func BenchmarkOptimizedHtmlPreservingTokenize(b *testing.B) {
	// Sample template
	source := `<!DOCTYPE html>
<html>
<head>
    <title>{{ title }}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{ asset_url('styles.css') }}">
</head>
<body>
    <header>
        <h1>{{ page.title }}</h1>
        <nav>
            <ul>
                {% for item in menu %}
                    <li><a href="{{ item.url }}">{{ item.label }}</a></li>
                {% endfor %}
            </ul>
        </nav>
    </header>
    
    <main>
        {% if content %}
            <article>
                {{ content|raw }}
            </article>
        {% else %}
            <p>No content available.</p>
        {% endif %}
        
        {% block sidebar %}
            <aside>
                {% include "sidebar.twig" with {items: sidebar_items} %}
            </aside>
        {% endblock %}
    </main>
    
    <footer>
        <p>&copy; {{ "now"|date("Y") }} {{ site_name }}</p>
    </footer>
</body>
</html>`

	parser := &Parser{source: source}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.optimizedHtmlPreservingTokenize()
	}
}

func BenchmarkImprovedHtmlPreservingTokenize(b *testing.B) {
	// Sample template (same as above)
	source := `<!DOCTYPE html>
<html>
<head>
    <title>{{ title }}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{ asset_url('styles.css') }}">
</head>
<body>
    <header>
        <h1>{{ page.title }}</h1>
        <nav>
            <ul>
                {% for item in menu %}
                    <li><a href="{{ item.url }}">{{ item.label }}</a></li>
                {% endfor %}
            </ul>
        </nav>
    </header>
    
    <main>
        {% if content %}
            <article>
                {{ content|raw }}
            </article>
        {% else %}
            <p>No content available.</p>
        {% endif %}
        
        {% block sidebar %}
            <aside>
                {% include "sidebar.twig" with {items: sidebar_items} %}
            </aside>
        {% endblock %}
    </main>
    
    <footer>
        <p>&copy; {{ "now"|date("Y") }} {{ site_name }}</p>
    </footer>
</body>
</html>`

	parser := &Parser{source: source}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.improvedHtmlPreservingTokenize()
	}
}

func BenchmarkZeroAllocHtmlTokenize(b *testing.B) {
	// Same sample template used in other benchmarks
	source := `<!DOCTYPE html>
<html>
<head>
    <title>{{ title }}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{ asset_url('styles.css') }}">
</head>
<body>
    <header>
        <h1>{{ page.title }}</h1>
        <nav>
            <ul>
                {% for item in menu %}
                    <li><a href="{{ item.url }}">{{ item.label }}</a></li>
                {% endfor %}
            </ul>
        </nav>
    </header>
    
    <main>
        {% if content %}
            <article>
                {{ content|raw }}
            </article>
        {% else %}
            <p>No content available.</p>
        {% endif %}
        
        {% block sidebar %}
            <aside>
                {% include "sidebar.twig" with {items: sidebar_items} %}
            </aside>
        {% endblock %}
    </main>
    
    <footer>
        <p>&copy; {{ "now"|date("Y") }} {{ site_name }}</p>
    </footer>
</body>
</html>`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenizer := GetTokenizer(source, 0)
		_, _ = tokenizer.TokenizeHtmlPreserving()
		ReleaseTokenizer(tokenizer)
	}
}

func BenchmarkTokenizeExpression(b *testing.B) {
	source := `user.name ~ " is " ~ user.age ~ " years old and lives in " ~ user.address.city`
	parser := &Parser{source: source}
	tokens := make([]Token, 0, 30)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokens = tokens[:0]
		parser.tokenizeExpression(source, &tokens, 1)
	}
}

func BenchmarkOptimizedTokenizeExpression(b *testing.B) {
	source := `user.name ~ " is " ~ user.age ~ " years old and lives in " ~ user.address.city`
	parser := &Parser{source: source}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenSlice := GetPooledTokenSlice(30)
		parser.optimizedTokenizeExpression(source, tokenSlice, 1)
		tokenSlice.Release()
	}
}

func BenchmarkImprovedTokenizeExpression(b *testing.B) {
	source := `user.name ~ " is " ~ user.age ~ " years old and lives in " ~ user.address.city`
	parser := &Parser{source: source}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenSlice := GetImprovedTokenSlice(30)
		parser.optimizedTokenizeExpressionImproved(source, tokenSlice, 1)
		tokenSlice.Release()
	}
}

func BenchmarkZeroAllocTokenize(b *testing.B) {
	source := `user.name ~ " is " ~ user.age ~ " years old and lives in " ~ user.address.city`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenizer := GetTokenizer(source, 30)
		tokenizer.TokenizeExpression(source)
		ReleaseTokenizer(tokenizer)
	}
}

func BenchmarkComplexTokenize(b *testing.B) {
	// A more complex example with nested structures
	source := `{% for user in users %}
    {% if user.active %}
        <div class="user {{ user.role }}">
            <h2>{{ user.name|title }}</h2>
            <p>{{ user.bio|striptags|truncate(100) }}</p>
            
            {% if user.permissions is defined and 'admin' in user.permissions %}
                <span class="admin-badge">Admin</span>
            {% endif %}
            
            <ul class="contact-info">
                {% for method, value in user.contacts %}
                    <li class="{{ method }}">{{ value }}</li>
                {% endfor %}
            </ul>
            
            {% set stats = user.getStatistics() %}
            <div class="stats">
                <span>Posts: {{ stats.posts }}</span>
                <span>Comments: {{ stats.comments }}</span>
                <span>Last active: {{ stats.lastActive|date("d M Y") }}</span>
            </div>
        </div>
    {% else %}
        <!-- User {{ user.name }} is inactive -->
    {% endif %}
{% endfor %}`

	parser := &Parser{source: source}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.optimizedHtmlPreservingTokenize()
	}
}

func BenchmarkTokenizeComplexObject(b *testing.B) {
	// A complex object with nested structures
	source := `{
		name: "John Doe",
		age: 30,
		address: {
			street: "123 Main St",
			city: "New York",
			country: "USA"
		},
		preferences: {
			theme: "dark",
			notifications: true,
			privacy: {
				showEmail: false,
				showPhone: true
			}
		},
		contacts: ["john@example.com", "+1234567890"],
		scores: [95, 87, 92, 78],
		metadata: {
			created: "2023-01-15",
			modified: "2023-06-22",
			tags: ["user", "premium", "verified"]
		}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenSlice := GetPooledTokenSlice(100)
		optimizedTokenizeComplexObject(source, tokenSlice, 1)
		tokenSlice.Release()
	}
}