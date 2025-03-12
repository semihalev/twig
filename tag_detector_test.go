package twig

import (
	"testing"
)

func TestFindNextTag(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		startPos int
		expected TagLocation
	}{
		{
			name:     "Empty source",
			source:   "",
			startPos: 0,
			expected: TagLocation{TAG_NONE, -1, 0},
		},
		{
			name:     "No tags",
			source:   "Hello world",
			startPos: 0,
			expected: TagLocation{TAG_NONE, -1, 0},
		},
		{
			name:     "Variable tag",
			source:   "Hello {{ name }}",
			startPos: 0,
			expected: TagLocation{TAG_VAR, 6, 2},
		},
		{
			name:     "Variable trim tag",
			source:   "Hello {{- name }}",
			startPos: 0,
			expected: TagLocation{TAG_VAR_TRIM, 6, 3},
		},
		{
			name:     "Block tag",
			source:   "Hello {% if name %}",
			startPos: 0,
			expected: TagLocation{TAG_BLOCK, 6, 2},
		},
		{
			name:     "Block trim tag",
			source:   "Hello {%- if name %}",
			startPos: 0,
			expected: TagLocation{TAG_BLOCK_TRIM, 6, 3},
		},
		{
			name:     "Comment tag",
			source:   "Hello {# comment #}",
			startPos: 0,
			expected: TagLocation{TAG_COMMENT, 6, 2},
		},
		{
			name:     "Start position after tag",
			source:   "{{ var1 }} text {{ var2 }}",
			startPos: 10,
			expected: TagLocation{TAG_VAR, 16, 2},
		},
		{
			name:     "Escaped tag",
			source:   "Hello \\{{ name }}",
			startPos: 0,
			expected: TagLocation{TAG_VAR, 7, 2}, // We find the tag, escaping handled separately
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FindNextTag(test.source, test.startPos)
			if result.Type != test.expected.Type ||
				result.Position != test.expected.Position ||
				result.Length != test.expected.Length {
				t.Errorf("Expected %+v, got %+v", test.expected, result)
			}
		})
	}
}

func TestFindTagEnd(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		startPos int
		tagType  TagType
		expected int
	}{
		{
			name:     "Variable tag end",
			source:   "{{ name }}",
			startPos: 2,
			tagType:  TAG_VAR,
			expected: 8,
		},
		{
			name:     "Block tag end",
			source:   "{% if name %}",
			startPos: 2,
			tagType:  TAG_BLOCK,
			expected: 11,
		},
		{
			name:     "Comment tag end",
			source:   "{# comment #}",
			startPos: 2,
			tagType:  TAG_COMMENT,
			expected: 11,
		},
		{
			name:     "No end found",
			source:   "{{ name",
			startPos: 2,
			tagType:  TAG_VAR,
			expected: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FindTagEnd(test.source, test.startPos, test.tagType)
			if result != test.expected {
				t.Errorf("Expected %d, got %d", test.expected, result)
			}
		})
	}
}

func BenchmarkFindNextTag(b *testing.B) {
	// Sample templates of different sizes
	templates := []struct {
		name   string
		source string
	}{
		{
			name:   "Small",
			source: "<div>{{ simple_var }}</div>",
		},
		{
			name:   "Medium",
			source: `<!DOCTYPE html>
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
</html>`,
		},
		{
			name:   "Large",
			source: `<!DOCTYPE html>
<html>
<head>
    <title>{{ page_title }}</title>
    <meta name="description" content="{{ meta_description }}">
    <link rel="stylesheet" href="/css/style.css">
    <script src="/js/main.js"></script>
</head>
<body>
    <header>
        <h1>{{ page_title }}</h1>
        <nav>
            {% for item in menu_items %}
                <a href="{{ item.url }}" {% if item.active %}class="active"{% endif %}>
                    {{ item.label }}
                </a>
            {% endfor %}
        </nav>
    </header>
    
    <main>
        {% if user %}
            <div class="user-greeting">
                Hello, {{ user.name }}!
                {% if user.messages|length > 0 %}
                    <div class="messages">
                        <h3>You have {{ user.messages|length }} new messages</h3>
                        <ul>
                            {% for message in user.messages %}
                                <li>{{ message.text }} - {{ message.date|date("M d, Y") }}</li>
                            {% endfor %}
                        </ul>
                    </div>
                {% endif %}
            </div>
        {% else %}
            <div class="login-prompt">Please <a href="/login">log in</a> to see your messages.</div>
        {% endif %}
        
        <div class="content">
            {{ content|raw }}
        </div>
    </main>
    
    <footer>
        &copy; {{ "now"|date("Y") }} {{ site_name }} - All rights reserved.
    </footer>
</body>
</html>`,
		},
	}

	for _, template := range templates {
		b.Run(template.name, func(b *testing.B) {
			source := template.source
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				pos := 0
				for {
					tag := FindNextTag(source, pos)
					if tag.Type == TAG_NONE {
						break
					}
					end := FindTagEnd(source, tag.Position+tag.Length, tag.Type)
					if end == -1 {
						break
					}
					pos = end + 2 // Move past the closing tag
				}
			}
		})
	}
}