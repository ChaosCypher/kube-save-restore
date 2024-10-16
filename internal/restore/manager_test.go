package restore

import (
	"testing"
)

// TestCountResources tests the countResources method of the Manager struct.
func TestCountResources(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		expectedCount int
	}{
		{
			name:          "Empty file list",
			files:         []string{},
			expectedCount: 0,
		},
		{
			name:          "Single file",
			files:         []string{"file1.json"},
			expectedCount: 1,
		},
		{
			name:          "Multiple files",
			files:         []string{"file1.json", "file2.json", "file3.json"},
			expectedCount: 3,
		},
	}

	rm := &Manager{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := rm.countResources(tt.files)
			if count != tt.expectedCount {
				t.Errorf("countResources() = %d; want %d", count, tt.expectedCount)
			}
		})
	}
}
