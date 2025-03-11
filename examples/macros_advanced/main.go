package main

import (
	"fmt"
	"os"

	"github.com/semihalev/twig"
)

func main() {
	// Create a new Twig engine
	engine := twig.New()

	// Template with _self reference for macros
	selfRefTemplate := `
{% macro input(name, value = '') %}
    <input type="text" name="{{ name }}" value="{{ value }}">
{% endmacro %}

{% macro form(action, method = 'post') %}
    <form action="{{ action }}" method="{{ method }}">
        {# Use _self to reference macros in the same template #}
        {{ _self.input('username', 'john') }}
        <button type="submit">Submit</button>
    </form>
{% endmacro %}

<div class="container">
    <h3>Example of using _self to reference macros in the same template:</h3>
    {{ _self.form('/submit') }}
</div>
`

	// Template demonstrating macro scope
	scopeTemplate := `
{% set name = 'Global' %}

{% macro greet(name = 'Default') %}
    {# This only sees the 'name' parameter, not the global 'name' #}
    <div class="greeting">Hello, {{ name }}!</div>
{% endmacro %}

<div class="container">
    <h3>Example of macro variable scope:</h3>
    <p>Using default parameter: {{ greet() }}</p>
    <p>Using provided parameter: {{ greet('Local') }}</p>
    <p>Global variable outside macro: {{ name }}</p>
</div>
`

	// Template showing global context access vs macro scope
	contextTemplate := `
{% set site_name = 'Twig Examples' %}
{% set current_user = 'admin' %}

{% macro show_user_info(user) %}
    <div class="user-info">
        <h4>User Information:</h4>
        <ul>
            <li>Name: {{ user.name }}</li>
            <li>Email: {{ user.email }}</li>
            <li>Role: {{ user.role }}</li>
        </ul>
    </div>
{% endmacro %}

{% macro greet(username) %}
    <div class="greeting">
        <p>Hello, {{ username }}!</p>
        <p>Welcome to {{ site_name }}</p> {# This won't work - site_name is not in scope #}
    </div>
{% endmacro %}

<div class="container">
    <h3>Example of macro scope vs global scope:</h3>
    
    <div class="global-scope">
        <h4>These variables are available in the global scope:</h4>
        <ul>
            <li>site_name: {{ site_name }}</li>
            <li>current_user: {{ current_user }}</li>
            <li>user.name: {{ user.name }}</li>
        </ul>
    </div>
    
    {# Passing specific values to the macro #}
    {{ show_user_info(user) }}
    
    {# This will show that site_name is not accessible in macro scope #}
    {{ greet(user.name) }}
</div>
`

	// Template with macro library organization
	libraryTemplate := `
{# This template simulates a component library with multiple macros #}

{% macro panel(title, content, type = 'default') %}
    <div class="panel panel-{{ type }}">
        <div class="panel-heading">{{ title }}</div>
        <div class="panel-body">{{ content }}</div>
    </div>
{% endmacro %}

{% macro alert(message, type = 'info') %}
    <div class="alert alert-{{ type }}">
        {{ message }}
    </div>
{% endmacro %}

{% macro badge(text, color = 'blue') %}
    <span class="badge" style="background-color: {{ color }}">{{ text }}</span>
{% endmacro %}
`

	// Template that uses the library
	useLibraryTemplate := `
{% import "library.twig" as ui %}

<div class="container">
    <h3>Example of organizing macros in libraries:</h3>
    
    {{ ui.panel('User Profile', 'Welcome back, ' ~ user.name, 'primary') }}
    
    {{ ui.alert('Your account has been verified!', 'success') }}
    
    <p>Status: {{ ui.badge('Active', 'green') }}</p>
</div>
`

	// Template demonstrating selective import (workaround for "from ... import")
	fromImportTemplate := `
{# Current implementation doesn't fully support "from X import Y" syntax directly #}
{# This example just shows how to use the namespace with "import" #}
{% import "library.twig" as lib %}

<div class="container">
    <h3>Example of selective macro usage:</h3>
    
    {# Use specific macros from the library, but not all of them #}
    {{ lib.panel('Quick Stats', 'Visitors today: 1,234', 'info') }}
    
    {{ lib.alert('New message received!', 'warning') }}
    
    {# We don't use the badge macro to simulate selective import #}
    
    <p class="note">Note: Full "from ... import" selective import syntax is on the roadmap</p>
</div>
`

	// Template with performance optimization techniques
	optimizationTemplate := `
<div class="container">
    <h3>Macro Performance Optimization Techniques:</h3>

    {# 1. Use imported macros for better performance #}
    <div class="technique">
        <h4>Technique #1: Use imported macros when possible</h4>
        <p>Imported macros perform better than direct usage.</p>
        <pre>&#123;% import "forms.twig" as forms %&#125;
&#123;&#123; forms.input('username') &#125;&#125;</pre>
    </div>

    {# 2. Group related macros in separate files #}
    <div class="technique">
        <h4>Technique #2: Group related macros in separate files</h4>
        <p>This helps with organization and caching efficiency.</p>
        <pre>forms.twig - Input and form macros
layout.twig - Layout components
widgets.twig - UI widgets</pre>
    </div>

    {# 3. Import once, use multiple times #}
    <div class="technique">
        <h4>Technique #3: Import once, use multiple times</h4>
        <p>Import at the top of your template to reuse throughout.</p>
        <pre>&#123;% import "macros.twig" as m %&#125;
&#123;&#123; m.widget1() &#125;&#125;
&#123;&#123; m.widget2() &#125;&#125;
&#123;&#123; m.widget3() &#125;&#125;</pre>
    </div>

    {# 4. Simple macros with focused responsibility #}
    <div class="technique">
        <h4>Technique #4: Keep macros simple and focused</h4>
        <p>Smaller, focused macros are easier to optimize and cache.</p>
        <pre>&#123;% macro avatar(user) %&#125;
  &lt;img src="&#123;&#123; user.avatar &#125;&#125;" alt="&#123;&#123; user.name &#125;&#125;"&gt;
&#123;% endmacro %&#125;

&#123;% macro userinfo(user) %&#125;
  &lt;div&gt;&#123;&#123; avatar(user) &#125;&#125; &#123;&#123; user.name &#125;&#125;&lt;/div&gt;
&#123;% endmacro %&#125;</pre>
    </div>
</div>
`

	// Register templates
	templates := map[string]string{
		"self_ref.twig":     selfRefTemplate,
		"scope.twig":        scopeTemplate,
		"context.twig":      contextTemplate,
		"library.twig":      libraryTemplate,
		"use_library.twig":  useLibraryTemplate,
		"from_import.twig":  fromImportTemplate,
		"optimization.twig": optimizationTemplate,
	}

	for name, content := range templates {
		err := engine.RegisterString(name, content)
		if err != nil {
			fmt.Printf("Error registering template %s: %v\n", name, err)
			return
		}
	}

	// Create context for templates
	context := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"role":  "Administrator",
		},
		"items": []string{"Item 1", "Item 2", "Item 3"},
		"settings": map[string]interface{}{
			"theme":  "dark",
			"public": true,
		},
	}

	// Render each template
	fmt.Println("===== ADVANCED MACRO EXAMPLES =====")

	for _, name := range []string{"self_ref.twig", "scope.twig", "context.twig", "use_library.twig", "from_import.twig", "optimization.twig"} {
		fmt.Printf("\n----- %s -----\n\n", name)

		err := engine.RenderTo(os.Stdout, name, context)
		if err != nil {
			fmt.Printf("Error rendering template %s: %v\n", name, err)
			continue
		}

		fmt.Println()
	}
}
