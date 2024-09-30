package restore

import (
	"testing"
)

// TestCountResources tests the countResources method of the Manager struct.
func TestCountResources(t *testing.T) {
	// Create a new Manager instance
	rm := &Manager{}

	// Define a slice of file names
	files := []string{"file1.txt", "file2.txt", "file3.txt"}

	// Define the expected result
	expected := 3

	// Call the countResources method and store the result
	result := rm.countResources(files)

	// Check if the result matches the expected value
	if result != expected {
		t.Errorf("countResources() = %d; want %d", result, expected)
	}
}
