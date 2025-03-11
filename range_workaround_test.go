package twig

import (
	"testing"
)

// TestRangeNegativeStepWorkaround demonstrates the workaround for using
// negative step values with the range function
func TestRangeNegativeStepWorkaround(t *testing.T) {
	// No need for engine here, we're testing the function directly

	// Test directly calling the range function with a negative step
	// This simulates what happens when the engine calls the function
	extension := &CoreExtension{}
	rangeFunc := extension.GetFunctions()["range"]

	// Call range(5, 1, -1)
	result, err := rangeFunc(5, 1, -1)
	if err != nil {
		t.Fatalf("Error calling range function directly: %v", err)
	}

	// Check the result
	slice, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", result)
	}

	// Log what the range function returned
	t.Logf("Range function with negative step returned: %v", slice)

	// Expect values from 5 down to 1 (inclusive)
	expected := []int{5, 4, 3, 2, 1}
	if len(slice) != len(expected) {
		t.Fatalf("Expected %d elements, got %d", len(expected), len(slice))
	}

	// Check each element
	for i, val := range slice {
		if val != expected[i] {
			t.Errorf("Expected slice[%d]=%d, got %v", i, expected[i], val)
		}
	}

	// Test the function with our parser improvements
	engine := New()

	// Test with direct negative step in template (should now work with parser improvements)
	err = engine.RegisterString("range_direct_neg", `{% for i in range(5, 1, -1) %}{{ i }}{% endfor %}`)
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	templateResult, err := engine.Render("range_direct_neg", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	expectedString := "54321"
	if templateResult != expectedString {
		t.Fatalf("Expected: %q, Got: %q", expectedString, templateResult)
	}

	// Test with direct negative step and more complex expressions
	err = engine.RegisterString("range_complex_neg", `{% for i in range(10, 0, (-2)) %}{{ i }}{% endfor %}`)
	if err != nil {
		t.Fatalf("Error registering template: %v", err)
	}

	templateResult, err = engine.Render("range_complex_neg", nil)
	if err != nil {
		t.Fatalf("Error rendering template: %v", err)
	}

	expectedString = "1086420"
	if templateResult != expectedString {
		t.Fatalf("Expected: %q, Got: %q", expectedString, templateResult)
	}

	t.Logf("The range function itself correctly handles negative steps")
	t.Logf("Our parser improvements now allow direct negative literals in templates")
}