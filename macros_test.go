package twig

import (
	"strings"
	"testing"
)

// TestMacrosWithDefaults tests macro functionality with default parameters
func TestMacrosWithDefaults(t *testing.T) {
	engine := New()

	// Create a template with macros that include default values
	source := `
	{% macro input(name, value = '', type = 'text', size = 20) %}
	<input type="{{ type }}" name="{{ name }}" value="{{ value }}" size="{{ size }}">
	{% endmacro %}
	
	{% macro textarea(name, value = '', rows = 10, cols = 40) %}
	<textarea name="{{ name }}" rows="{{ rows }}" cols="{{ cols }}">{{ value }}</textarea>
	{% endmacro %}
	
	{% macro label(text, for = '') %}
	<label{% if for %} for="{{ for }}"{% endif %}>{{ text }}</label>
	{% endmacro %}
	
	{{ input('username', 'john') }}
	{{ input('password', '****', 'password') }}
	{{ textarea('description', 'This is a test') }}
	{{ label('Username', 'username') }}
	{{ label('Simple Label') }}
	`

	engine.RegisterString("test_macros_defaults", source)
	result, err := engine.Render("test_macros_defaults", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	// Check the output contains the expected HTML
	expectedHtml := []string{
		`<input type="text" name="username" value="john" size="20">`,
		`<input type="password" name="password" value="****" size="20">`,
		`<textarea name="description" rows="10" cols="40">This is a test</textarea>`,
		`<label for="username">Username</label>`,
		`<label>Simple Label</label>`,
	}

	for _, expected := range expectedHtml {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected %q in result, but got: %s", expected, result)
		}
	}
}

// TestMacrosWithEscaping tests macro functionality with escaped parameters
func TestMacrosWithEscaping(t *testing.T) {
	engine := New()

	// Create a template with macros that use the escape filter
	source := `
	{% macro input(name, value = '', type = 'text') %}
	<input type="{{ type }}" name="{{ name }}" value="{{ value|e }}">
	{% endmacro %}
	
	{{ input('test', '<script>alert("XSS")</script>') }}
	`

	engine.RegisterString("test_macros_escape", source)
	result, err := engine.Render("test_macros_escape", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	expected := `<input type="text" name="test" value="&lt;script&gt;alert(&#34;XSS&#34;)&lt;/script&gt;">`
	if !strings.Contains(result, expected) {
		t.Errorf("Expected escaped output %q in result, but got: %s", expected, result)
	}
}

// TestMacrosImport tests importing macros from another template
func TestMacrosImport(t *testing.T) {
	engine := New()

	// Macro library template
	macroLib := `
	{% macro input(name, value = '', type = 'text', size = 20) %}
	<input type="{{ type }}" name="{{ name }}" value="{{ value }}" size="{{ size }}">
	{% endmacro %}
	
	{% macro button(name, value) %}
	<button name="{{ name }}">{{ value }}</button>
	{% endmacro %}
	`

	// Main template that imports macros
	mainTemplate := `
	{% import "macro_lib.twig" as forms %}
	
	<form>
		{{ forms.input('username', 'john') }}
		{{ forms.button('submit', 'Submit Form') }}
	</form>
	`

	// Register both templates
	engine.RegisterString("macro_lib.twig", macroLib)
	engine.RegisterString("main.twig", mainTemplate)

	// Render the main template
	result, err := engine.Render("main.twig", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	// Check the output
	expectedHtml := []string{
		`<input type="text" name="username" value="john" size="20">`,
		`<button name="submit">Submit Form</button>`,
	}

	for _, expected := range expectedHtml {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected %q in result, but got: %s", expected, result)
		}
	}
}

// TestMacrosImportAs tests importing macros using the import as syntax
func TestMacrosImportAs(t *testing.T) {
	engine := New()

	// Macro library template
	macroLib := `
	{% macro input(name, value = '', type = 'text') %}
	<input type="{{ type }}" name="{{ name }}" value="{{ value }}">
	{% endmacro %}
	
	{% macro textarea(name, value = '') %}
	<textarea name="{{ name }}">{{ value }}</textarea>
	{% endmacro %}
	
	{% macro button(name, value) %}
	<button name="{{ name }}">{{ value }}</button>
	{% endmacro %}
	`

	// Main template that imports macros using import as syntax
	mainTemplate := `
	{% import "macro_lib.twig" as lib %}
	
	<form>
		{{ lib.input('username', 'john') }}
		{{ lib.button('submit', 'Submit Form') }}
	</form>
	`

	// Register both templates
	err := engine.RegisterString("macro_lib.twig", macroLib)
	if err != nil {
		t.Fatalf("Error registering macro_lib.twig: %v", err)
	}

	err = engine.RegisterString("import_as.twig", mainTemplate)
	if err != nil {
		t.Fatalf("Error registering import_as.twig: %v", err)
	}

	// Render the main template
	result, err := engine.Render("import_as.twig", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	// Check the output
	expectedHtml := []string{
		`<input type="text" name="username" value="john">`,
		`<button name="submit">Submit Form</button>`,
	}

	for _, expected := range expectedHtml {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected %q in result, but got: %s", expected, result)
		}
	}
}

// TestMacrosFromImport tests selective importing macros using the from import syntax
func TestMacrosFromImport(t *testing.T) {
	engine := New()

	// Macro library template
	macroLib := `
	{% macro input(name, value = '', type = 'text') %}
	<input type="{{ type }}" name="{{ name }}" value="{{ value }}">
	{% endmacro %}
	
	{% macro textarea(name, value = '') %}
	<textarea name="{{ name }}">{{ value }}</textarea>
	{% endmacro %}
	
	{% macro button(name, value) %}
	<button name="{{ name }}">{{ value }}</button>
	{% endmacro %}
	`

	// Main template that selectively imports macros using from import syntax
	mainTemplate := `{% from "macro_lib.twig" import input, button %}
	
	<form>
		{{ input('username', 'john') }}
		{{ button('submit', 'Submit Form') }}
	</form>
	`

	// Register both templates
	err := engine.RegisterString("macro_lib.twig", macroLib)
	if err != nil {
		t.Fatalf("Error registering macro_lib.twig: %v", err)
	}

	err = engine.RegisterString("from_import.twig", mainTemplate)
	if err != nil {
		t.Fatalf("Error registering from_import.twig: %v", err)
	}

	// Render the main template
	result, err := engine.Render("from_import.twig", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	// Check the output
	expectedHtml := []string{
		`<input type="text" name="username" value="john">`,
		`<button name="submit">Submit Form</button>`,
	}

	for _, expected := range expectedHtml {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected %q in result, but got: %s", expected, result)
		}
	}
}

// TestMacrosFromImportWithAliases tests importing macros with aliases using the from import syntax
func TestMacrosFromImportWithAliases(t *testing.T) {
	engine := New()

	// Macro library template
	macroLib := `
	{% macro input(name, value = '', type = 'text') %}
	<input type="{{ type }}" name="{{ name }}" value="{{ value }}">
	{% endmacro %}
	
	{% macro textarea(name, value = '') %}
	<textarea name="{{ name }}">{{ value }}</textarea>
	{% endmacro %}
	
	{% macro button(name, value) %}
	<button name="{{ name }}">{{ value }}</button>
	{% endmacro %}
	`

	// Main template that imports macros with aliases using from import syntax
	mainTemplate := `{% from "macro_lib.twig" import input as field, button as btn %}
	
	<form>
		{{ field('username', 'john') }}
		{{ btn('submit', 'Submit Form') }}
	</form>
	`

	// Register both templates
	err := engine.RegisterString("macro_lib.twig", macroLib)
	if err != nil {
		t.Fatalf("Error registering macro_lib.twig: %v", err)
	}

	err = engine.RegisterString("from_import_aliases.twig", mainTemplate)
	if err != nil {
		t.Fatalf("Error registering from_import_aliases.twig: %v", err)
	}

	// Render the main template
	result, err := engine.Render("from_import_aliases.twig", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	// Check the output
	expectedHtml := []string{
		`<input type="text" name="username" value="john">`,
		`<button name="submit">Submit Form</button>`,
	}

	for _, expected := range expectedHtml {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected %q in result, but got: %s", expected, result)
		}
	}
}

// TestMixedImportApproaches tests using both import and from import syntax in the same template
func TestMixedImportApproaches(t *testing.T) {
	engine := New()

	// First macro library template
	formsMacroLib := `
	{% macro input(name, value = '') %}
	<input name="{{ name }}" value="{{ value }}">
	{% endmacro %}
	
	{% macro bold(text) %}
	<b>{{ text }}</b>
	{% endmacro %}
	`

	// Second macro library template
	layoutMacroLib := `
	{% macro header(text) %}
	<h1>{{ text }}</h1>
	{% endmacro %}
	
	{% macro box(content) %}
	<div class="box">{{ content }}</div>
	{% endmacro %}
	`

	// Main template that uses both import approaches
	mainTemplate := `{% import "forms_macros.twig" as forms %}
{% from "layout_macros.twig" import header %}
	
<div>
	{{ header('Hello') }}
	{{ forms.input('username', 'john') }}
	{{ forms.bold('Welcome') }}
</div>
	`

	// Register templates
	err := engine.RegisterString("forms_macros.twig", formsMacroLib)
	if err != nil {
		t.Fatalf("Error registering forms_macros.twig: %v", err)
	}

	err = engine.RegisterString("layout_macros.twig", layoutMacroLib)
	if err != nil {
		t.Fatalf("Error registering layout_macros.twig: %v", err)
	}

	err = engine.RegisterString("mixed_imports.twig", mainTemplate)
	if err != nil {
		t.Fatalf("Error registering mixed_imports.twig: %v", err)
	}

	// Render the main template
	result, err := engine.Render("mixed_imports.twig", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	// Check the output
	expectedElements := []string{
		`<h1>Hello</h1>`,
		`<input name="username" value="john">`,
		`<b>Welcome</b>`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected %q in result, but got: %s", expected, result)
		}
	}
}

// TestMacrosWithContext tests macros with context variables
func TestMacrosWithContext(t *testing.T) {
	engine := New()

	// Create a template with macros that access context variables
	source := `
	{% macro greeting(name) %}
	Hello {{ name }}{% if company %} from {{ company }}{% endif %}!
	{% endmacro %}
	
	{{ greeting('John') }}
	`

	// Set up context
	context := map[string]interface{}{
		"company": "Acme Inc",
	}

	engine.RegisterString("test_macros_context", source)
	result, err := engine.Render("test_macros_context", context)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	expected := `Hello John from Acme Inc!`
	if !strings.Contains(result, expected) {
		t.Errorf("Expected %q in result, but got: %s", expected, result)
	}
}

// TestMacrosWithComplexExpression tests macros with more complex expressions
func TestMacrosWithComplexExpression(t *testing.T) {
	engine := New()

	// Create a template with macros that have complex expressions
	source := `
	{% macro conditional_class(condition, class1, class2) %}
	<div class="{{ condition ? class1 : class2 }}">Content</div>
	{% endmacro %}
	
	{{ conditional_class(isActive, 'active', 'inactive') }}
	{{ conditional_class(isAdmin, 'admin-panel', 'user-panel') }}
	`

	// Set up context
	context := map[string]interface{}{
		"isActive": true,
		"isAdmin":  false,
	}

	engine.RegisterString("test_macros_complex", source)
	result, err := engine.Render("test_macros_complex", context)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	expectedHtml := []string{
		`<div class="active">Content</div>`,
		`<div class="user-panel">Content</div>`,
	}

	for _, expected := range expectedHtml {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected %q in result, but got: %s", expected, result)
		}
	}
}

// TestNestedMacros tests nested macro calls
func TestNestedMacros(t *testing.T) {
	engine := New()

	// Create a template with nested macro calls
	source := `
	{% macro field(name, value) %}
	<div class="field">
		{{ label(name) }}
		{{ input(name, value) }}
	</div>
	{% endmacro %}
	
	{% macro label(text) %}
	<label>{{ text }}</label>
	{% endmacro %}
	
	{% macro input(name, value) %}
	<input name="{{ name }}" value="{{ value }}">
	{% endmacro %}
	
	{{ field('username', 'john') }}
	`

	engine.RegisterString("test_nested_macros", source)
	result, err := engine.Render("test_nested_macros", nil)
	if err != nil {
		t.Fatalf("Error parsing/rendering template: %v", err)
	}

	// Check for the presence of the required elements rather than exact formatting
	expectedElements := []string{
		`<div class="field">`,
		`<label>username</label>`,
		`<input name="username" value="john">`,
		`</div>`,
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected element %q not found in result: %s", element, result)
		}
	}
}
