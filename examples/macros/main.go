package main

import (
	"fmt"
	"os"

	"github.com/semihalev/twig"
)

func main() {
	// Create a new Twig engine
	engine := twig.New()

	// Create template with macros
	macrosTemplate := `
{# Define macros in a separate template #}
{% macro input(name, value = '', type = 'text', size = 20) %}
  <input type="{{ type }}" name="{{ name }}" value="{{ value|e }}" size="{{ size }}">
{% endmacro %}

{% macro textarea(name, value = '', rows = 10, cols = 40) %}
  <textarea name="{{ name }}" rows="{{ rows }}" cols="{{ cols }}">{{ value|e }}</textarea>
{% endmacro %}

{% macro label(text, for = '') %}
  <label{% if for %} for="{{ for }}"{% endif %}>{{ text }}</label>
{% endmacro %}
`

	// Create a template that imports and uses macros
	mainTemplate := `
{% import "macros.twig" as forms %}

<form>
  <div class="form-row">
    {{ forms.label('Username', 'username') }}
    {{ forms.input('username', user.username) }}
  </div>
  <div class="form-row">
    {{ forms.label('Bio', 'bio') }}
    {{ forms.textarea('bio', user.bio, 5, 60) }}
  </div>
  <div class="form-row">
    {{ forms.input('submit', 'Submit', 'submit') }}
  </div>
</form>
`

	// Register templates
	err := engine.RegisterString("macros.twig", macrosTemplate)
	if err != nil {
		fmt.Printf("Error registering macros template: %v\n", err)
		return
	}

	err = engine.RegisterString("main.twig", mainTemplate)
	if err != nil {
		fmt.Printf("Error registering main template: %v\n", err)
		return
	}

	// Create context with user data
	context := map[string]interface{}{
		"user": map[string]interface{}{
			"username": "johndoe",
			"bio":      "I'm a passionate developer and open source contributor.",
		},
	}

	// Render the template
	fmt.Println("Rendering template with macros:")
	err = engine.RenderTo(os.Stdout, "main.twig", context)
	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return
	}
}
