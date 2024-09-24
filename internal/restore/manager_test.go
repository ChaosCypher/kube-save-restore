package restore

import (
	"testing"
)

func TestCountResources(t *testing.T) {
	rm := &Manager{}
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	expected := 3

	result := rm.countResources(files)
	if result != expected {
		t.Errorf("countResources() = %d; want %d", result, expected)
	}
}
