// test_helpers.go

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// createMockBackup creates a temporary directory with mock backup files for testing.
// It returns the path to the temporary directory and a cleanup function.
func createMockBackup(t *testing.T) (string, string, func()) {
	// Create a temporary directory for source
	sourceDir, err := os.MkdirTemp("", "test-backup-source")
	if err != nil {
		t.Fatalf("Failed to create source temp directory: %v", err)
	}

	// Create a temporary directory for target
	targetDir, err := os.MkdirTemp("", "test-backup-target")
	if err != nil {
		t.Fatalf("Failed to create target temp directory: %v", err)
	}

	// Create mock backup directory structure for source
	mockSourceBackupDir := filepath.Join(sourceDir, "all")
	err = os.MkdirAll(mockSourceBackupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create mock source backup directory: %v", err)
	}

	// Create mock backup directory structure for target
	mockTargetBackupDir := filepath.Join(targetDir, "all")
	err = os.MkdirAll(mockTargetBackupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create mock target backup directory: %v", err)
	}

	// Create a mock backup file in source
	createMockFile(t, mockSourceBackupDir, "mock-resource-source.yaml")

	// Create a mock backup file in target
	createMockFile(t, mockTargetBackupDir, "mock-resource-target.yaml")

	// Return the temp directory paths and a cleanup function
	return sourceDir, targetDir, func() {
		os.RemoveAll(sourceDir)
		os.RemoveAll(targetDir)
	}
}

func createMockFile(t *testing.T, dir, filename string) {
	mockContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key: value
`
	err := os.WriteFile(filepath.Join(dir, filename), []byte(mockContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock backup file: %v", err)
	}
}
