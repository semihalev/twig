package twig

import (
	"bytes"
	"testing"
)

// BenchmarkNodePooling tests the impact of node object pooling on memory allocations
func BenchmarkNodePooling(b *testing.B) {
	// Create test templates that use different node types heavily
	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "TextNodes",
			template: "This is a template with lots of text. It has multiple paragraphs.\nAnd newlines.\nAnd more text.",
		},
		{
			name:     "PrintNodes",
			template: "{{ a }} {{ b }} {{ c }} {{ d }} {{ e }} {{ f }} {{ g }} {{ h }} {{ i }} {{ j }}",
		},
		{
			name:     "IfNodes",
			template: "{% if a %}A{% endif %}{% if b %}B{% endif %}{% if c %}C{% endif %}{% if d %}D{% endif %}{% if e %}E{% endif %}",
		},
		{
			name:     "ForNodes",
			template: "{% for i in items %}{{ i }}{% endfor %}{% for j in range(1, 10) %}{{ j }}{% endfor %}",
		},
		{
			name:     "MixedNodes",
			template: "Text {{ var }} {% if cond %}Conditional Content {{ value }}{% endif %}{% for item in items %}{{ item.name }}{% endfor %}",
		},
		{
			name:     "ComplexTemplate",
			template: `
				<div class="container">
					<h1>{{ page.title }}</h1>
					<div class="content">
						{% if user.authenticated %}
							<p>Welcome back, {{ user.name }}!</p>
							{% if user.admin %}
								<div class="admin-panel">
									<h2>Admin Tools</h2>
									<ul>
										{% for tool in admin_tools %}
											<li><a href="{{ tool.url }}">{{ tool.name }}</a></li>
										{% endfor %}
									</ul>
								</div>
							{% endif %}
						{% else %}
							<p>Welcome, guest! Please <a href="/login">login</a>.</p>
						{% endif %}
						
						<h2>Recent Items</h2>
						<ul class="items">
							{% for item in items %}
								<li class="item {% if item.featured %}featured{% endif %}">
									<h3>{{ item.title }}</h3>
									<p>{{ item.description }}</p>
									{% if item.tags|length > 0 %}
										<div class="tags">
											{% for tag in item.tags %}
												<span class="tag">{{ tag }}</span>
											{% endfor %}
										</div>
									{% endif %}
								</li>
							{% endfor %}
						</ul>
					</div>
				</div>
			`,
		},
	}

	// Create a test context with all variables needed
	testContext := map[string]interface{}{
		"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", 
		"f": "F", "g": "G", "h": "H", "i": "I", "j": "J",
		"cond": true, "value": "Value", "var": "Variable",
		"items": []map[string]interface{}{
			{"name": "Item 1", "title": "Title 1", "description": "Description 1", "featured": true, "tags": []string{"tag1", "tag2"}},
			{"name": "Item 2", "title": "Title 2", "description": "Description 2", "featured": false, "tags": []string{"tag2", "tag3"}},
		},
		"page": map[string]interface{}{"title": "Page Title"},
		"user": map[string]interface{}{
			"authenticated": true,
			"name": "John Doe",
			"admin": true,
		},
		"admin_tools": []map[string]interface{}{
			{"name": "Dashboard", "url": "/admin/dashboard"},
			{"name": "Users", "url": "/admin/users"},
			{"name": "Settings", "url": "/admin/settings"},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			engine := New()
			
			// Pre-compile the template to isolate rendering performance
			tmpl, err := engine.ParseTemplate(tt.template)
			if err != nil {
				b.Fatalf("Failed to parse template: %v", err)
			}

			// Create a reusable buffer to avoid allocations in the benchmark loop
			buf := new(bytes.Buffer)
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				buf.Reset()
				
				// Render directly to buffer to avoid string conversions
				err = tmpl.RenderTo(buf, testContext)
				if err != nil {
					b.Fatalf("Failed to render template: %v", err)
				}
			}
		})
	}
}

// BenchmarkTokenPooling tests the impact of token pooling during parsing
func BenchmarkTokenPooling(b *testing.B) {
	// Create test templates with varying numbers of tokens
	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "SimpleTokens",
			template: "{{ var }} {{ var2 }} {{ var3 }}",
		},
		{
			name:     "MediumTokens",
			template: "{% if cond %}Text{{ var }}{% else %}OtherText{{ var2 }}{% endif %}{% for item in items %}{{ item }}{% endfor %}",
		},
		{
			name:     "ManyTokens",
			template: `
				{% for i in range(1, 10) %}
					{% if i > 5 %}
						{{ i }} is greater than 5
					{% else %}
						{{ i }} is less than or equal to 5
						{% if i == 5 %}
							{{ i }} is exactly 5
						{% endif %}
					{% endif %}
				{% endfor %}
			`,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			engine := New()
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				tmpl, err := engine.ParseTemplate(tt.template)
				if err != nil {
					b.Fatalf("Failed to parse template: %v", err)
				}
				
				// Ensure the template is valid
				if tmpl == nil {
					b.Fatalf("Template is nil")
				}
			}
		})
	}
}

// BenchmarkFullTemplateLifecycle tests the full lifecycle of template parsing and rendering
func BenchmarkFullTemplateLifecycle(b *testing.B) {
	// Complex template that heavily exercises the node pool
	complexTemplate := `
		<div class="container">
			<h1>{{ page.title }}</h1>
			{% if user.authenticated %}
				<div class="user-panel">
					<p>Welcome, {{ user.name }}</p>
					<ul class="menu">
						{% for item in menu_items %}
							<li class="{% if item.active %}active{% endif %}">
								<a href="{{ item.url }}">{{ item.label }}</a>
								{% if item.sub_items|length > 0 %}
									<ul class="submenu">
										{% for sub_item in item.sub_items %}
											<li><a href="{{ sub_item.url }}">{{ sub_item.label }}</a></li>
										{% endfor %}
									</ul>
								{% endif %}
							</li>
						{% endfor %}
					</ul>
				</div>
			{% else %}
				<div class="guest-panel">
					<p>Please login to continue.</p>
					<a href="/login" class="button">Login</a>
				</div>
			{% endif %}
		</div>
	`

	// Create a test context with all variables needed
	testContext := map[string]interface{}{
		"page": map[string]interface{}{
			"title": "Dashboard",
		},
		"user": map[string]interface{}{
			"authenticated": true,
			"name": "John Doe",
		},
		"menu_items": []map[string]interface{}{
			{
				"label": "Home",
				"url": "/",
				"active": true,
				"sub_items": []map[string]interface{}{},
			},
			{
				"label": "Products",
				"url": "/products",
				"active": false,
				"sub_items": []map[string]interface{}{
					{
						"label": "Category 1",
						"url": "/products/cat1",
					},
					{
						"label": "Category 2",
						"url": "/products/cat2",
					},
				},
			},
			{
				"label": "About",
				"url": "/about",
				"active": false,
				"sub_items": []map[string]interface{}{},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		engine := New()
		
		// Parse the template (tests token pooling)
		tmpl, err := engine.ParseTemplate(complexTemplate)
		if err != nil {
			b.Fatalf("Failed to parse template: %v", err)
		}
		
		// Render the template (tests node pooling)
		_, err = tmpl.Render(testContext)
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}