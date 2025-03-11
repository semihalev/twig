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

	t.Logf("The range function itself correctly handles negative steps")
	t.Logf("While direct template syntax fails due to tokenizer limitations")
	t.Logf("The solution is to use a variable for the negative step:")
	t.Logf("{%% set step = -1 %%}{%% for i in range(5, 1, step) %%}{{ i }}{%% endfor %%}")
}
