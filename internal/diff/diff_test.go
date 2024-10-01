package diff

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockLogger() *logger.Logger {
	mockLogger := logger.NewLogger(os.Stdout, logger.DEBUG)
	return mockLogger
}

// TestNewComparer tests the NewComparer function.
func TestNewComparer(t *testing.T) {
	resourceMaps := make(map[string]map[string]*ResourceWrapper)
	mockLog := setupMockLogger()

	comparer := NewComparer(resourceMaps, mockLog)

	assert.NotNil(t, comparer)
	assert.Equal(t, resourceMaps, comparer.resourceMaps)
	assert.Equal(t, mockLog, comparer.logger)
}

// TestComparer_Compare_AllResourcesPresent tests the Compare method when all resources are present in all backups.
func TestComparer_Compare_AllResourcesPresent(t *testing.T) {
	resourceMaps := map[string]map[string]*ResourceWrapper{
		"backup1": {
			"resource1": &ResourceWrapper{ /* initialize as needed */ },
			"resource2": &ResourceWrapper{ /* initialize as needed */ },
		},
		"backup2": {
			"resource1": &ResourceWrapper{ /* initialize as needed */ },
			"resource2": &ResourceWrapper{ /* initialize as needed */ },
		},
	}

	mockLog := setupMockLogger()
	comparer := NewComparer(resourceMaps, mockLog)

	report := comparer.Compare()

	assert.NotNil(t, report)
	assert.Equal(t, 2, len(report.Resources))
	assert.Empty(t, report.Missing)
}

// TestComparer_Compare_MissingResources tests the Compare method when some resources are missing from backups.
func TestComparer_Compare_MissingResources(t *testing.T) {
	resourceMaps := map[string]map[string]*ResourceWrapper{
		"backup1": {
			"resource1": &ResourceWrapper{ /* initialize as needed */ },
		},
		"backup2": {
			"resource1": &ResourceWrapper{ /* initialize as needed */ },
			"resource2": &ResourceWrapper{ /* initialize as needed */ },
		},
	}

	mockLog := setupMockLogger()
	comparer := NewComparer(resourceMaps, mockLog)

	report := comparer.Compare()

	assert.NotNil(t, report)
	assert.Equal(t, 2, len(report.Resources))
	assert.Len(t, report.Missing, 1)
	assert.Contains(t, report.Missing, "resource2")
	assert.Equal(t, []string{"backup1"}, report.Missing["resource2"])
}

// TestComparer_identifyMissing tests the identifyMissing helper method.
func TestComparer_identifyMissing(t *testing.T) {
	resourceMaps := map[string]map[string]*ResourceWrapper{
		"backup1": {
			"resource1": &ResourceWrapper{},
		},
		"backup2": {},
		"backup3": {
			"resource1": &ResourceWrapper{},
		},
	}

	mockLog := setupMockLogger()
	comparer := NewComparer(resourceMaps, mockLog)

	backups := map[string]*ResourceWrapper{
		"backup1": { /* initialize as needed */ },
		"backup3": { /* initialize as needed */ },
	}

	missing := comparer.identifyMissing(backups)

	assert.Len(t, missing, 1)
	assert.Contains(t, missing, "backup2")
}

// TestComparer_backupDirs tests the backupDirs helper method.
func TestComparer_backupDirs(t *testing.T) {
	resourceMaps := map[string]map[string]*ResourceWrapper{
		"backupA": {},
		"backupB": {},
		"backupC": {},
	}

	mockLog := setupMockLogger()
	comparer := NewComparer(resourceMaps, mockLog)

	dirs := comparer.backupDirs()

	assert.Len(t, dirs, 3)
	assert.Contains(t, dirs, "backupA")
	assert.Contains(t, dirs, "backupB")
	assert.Contains(t, dirs, "backupC")
}

// TestNewDiffManager tests the NewDiffManager function.
func TestNewDiffManager(t *testing.T) {
	backupDirs := []string{"backup1", "backup2"}
	mockLog := setupMockLogger()

	diffManager := NewDiffManager(backupDirs, mockLog)

	assert.NotNil(t, diffManager)
	assert.Equal(t, backupDirs, diffManager.backupDirs)
	assert.Equal(t, mockLog, diffManager.logger)
}

// TestDiffManager_PerformDiff tests the PerformDiff method.
func TestDiffManager_PerformDiff(t *testing.T) {
	// Create a temporary base directory
	tempBaseDir, err := os.MkdirTemp("", "backup-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)

	// Create temporary directories for testing
	tempDir1 := filepath.Join(tempBaseDir, "backup1")
	tempDir2 := filepath.Join(tempBaseDir, "backup2")
	require.NoError(t, os.Mkdir(tempDir1, 0755))
	require.NoError(t, os.Mkdir(tempDir2, 0755))

	// Create test JSON files
	createTestJSONFile(t, tempDir1, "resource1.json", `{"metadata": {"kind": "Pod", "namespace": "default", "name": "pod1"}}`)
	createTestJSONFile(t, tempDir2, "resource1.json", `{"metadata": {"kind": "Pod", "namespace": "default", "name": "pod1"}}`)
	createTestJSONFile(t, tempDir2, "resource2.json", `{"metadata": {"kind": "Service", "namespace": "default", "name": "svc1"}}`)

	backupDirs := []string{tempDir1, tempDir2}
	mockLog := setupMockLogger()

	diffManager := NewDiffManager(backupDirs, mockLog)

	report, err := diffManager.PerformDiff(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 2, len(report.Resources))
	assert.Len(t, report.Missing, 1)
	assert.Contains(t, report.Missing, "Service/default/svc1")
	assert.Equal(t, []string{tempDir1}, report.Missing["Service/default/svc1"])
}

// Helper function to create test JSON files
func createTestJSONFile(t *testing.T, dir, filename, content string) {
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	assert.NoError(t, err)
}

// TestLoadBackupDirectory tests the loadBackupDirectory function.
func TestLoadBackupDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-load-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	createTestJSONFile(t, tempDir, "resource1.json", `{"metadata": {"kind": "Pod", "namespace": "default", "name": "pod1"}}`)
	createTestJSONFile(t, tempDir, "resource2.json", `{"metadata": {"kind": "Service", "namespace": "default", "name": "svc1"}}`)

	resources, err := LoadBackupDirectory(tempDir)

	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Contains(t, resources, "Pod/default/pod1")
	assert.Contains(t, resources, "Service/default/svc1")
}

// TestGenerateResourceKey tests the generateResourceKey function.
func TestGenerateResourceKey(t *testing.T) {
	key := GenerateResourceKey("Pod", "default", "pod1")
	assert.Equal(t, "Pod/default/pod1", key)
}

func TestNewDiffReport(t *testing.T) {
	report := NewDiffReport()

	assert.NotNil(t, report)
	assert.Empty(t, report.Resources)
	assert.Empty(t, report.Missing)
}

func TestDiffReport_Print(t *testing.T) {
	report := NewDiffReport()
	report.Resources = map[string]map[string]*ResourceWrapper{
		"Pod/default/pod1": {
			"backup1": &ResourceWrapper{},
			"backup2": &ResourceWrapper{},
		},
		"Service/default/svc1": {
			"backup1": &ResourceWrapper{},
		},
	}
	report.Missing = map[string][]string{
		"Service/default/svc1": {"backup2"},
	}

	// Create a temporary file to capture the output
	tmpfile, err := os.CreateTemp("", "report-output")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Redirect stdout to the temporary file
	old := os.Stdout
	os.Stdout = tmpfile
	defer func() { os.Stdout = old }()

	report.Print()

	// Close the file to ensure all data is written
	tmpfile.Close()

	// Read the content of the temporary file
	content, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)

	// Assert the expected output
	expectedOutput := `Resource: Pod/default/pod1
  Backup: backup1 - Exists
  Backup: backup2 - Exists

Resource: Service/default/svc1
  Backup: backup1 - Exists
  Backup: backup2 - Missing

`
	assert.Equal(t, expectedOutput, string(content))
}

func TestDiffReport_AddResource(t *testing.T) {
	report := NewDiffReport()
	
	report.AddResource("Pod/default/pod1", "backup1", &ResourceWrapper{})
	report.AddResource("Pod/default/pod1", "backup2", &ResourceWrapper{})
	report.AddResource("Service/default/svc1", "backup1", &ResourceWrapper{})

	assert.Len(t, report.Resources, 2)
	assert.Len(t, report.Resources["Pod/default/pod1"], 2)
	assert.Len(t, report.Resources["Service/default/svc1"], 1)
}

func TestDiffReport_AddMissing(t *testing.T) {
	report := NewDiffReport()
	
	report.AddMissing("Pod/default/pod1", "backup2")
	report.AddMissing("Service/default/svc1", "backup2")
	report.AddMissing("Service/default/svc1", "backup3")

	assert.Len(t, report.Missing, 2)
	assert.Len(t, report.Missing["Pod/default/pod1"], 1)
	assert.Len(t, report.Missing["Service/default/svc1"], 2)
}

func TestLoadResourceFromFile(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		content  string
		expected *ResourceWrapper
		wantErr  bool
	}{
		{
			name: "Valid resource file",
			content: `{
				"metadata": {
					"name": "test-pod",
					"namespace": "default",
					"kind": "Pod"
				},
				"resource": {
					"apiVersion": "v1",
					"spec": {
						"containers": [
							{
								"name": "test-container",
								"image": "nginx:latest"
							}
						]
					}
				}
			}`,
			expected: &ResourceWrapper{
				Metadata: ResourceMetadata{
					Name:      "test-pod",
					Namespace: "default",
					Kind:      "Pod",
				},
				Resource: map[string]interface{}{
					"apiVersion": "v1",
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "test-container",
								"image": "nginx:latest",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "Invalid JSON file",
			content:  `{"invalid": "json"`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Empty file",
			content:  "",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpfile, err := os.CreateTemp("", "test-resource-*.json")
			assert.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			// Write the test content to the file
			_, err = tmpfile.Write([]byte(tt.content))
			assert.NoError(t, err)
			tmpfile.Close()

			// Test LoadResourceFromFile
			got, err := LoadResourceFromFile(tmpfile.Name())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestLoadResourceFromFile_FileNotFound(t *testing.T) {
	_, err := LoadResourceFromFile("non_existent_file.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading file")
}

func TestResourceWrapper_Serialization(t *testing.T) {
	wrapper := &ResourceWrapper{
		Metadata: ResourceMetadata{
			Name:      "test-service",
			Namespace: "default",
			Kind:      "Service",
		},
		Resource: map[string]interface{}{
			"apiVersion": "v1",
			"spec": map[string]interface{}{
				"ports": []interface{}{
					map[string]interface{}{
						"port":     80,
						"protocol": "TCP",
					},
				},
			},
		},
	}

	data, err := json.Marshal(wrapper)
	assert.NoError(t, err)

	var newWrapper ResourceWrapper
	err = json.Unmarshal(data, &newWrapper)
	assert.NoError(t, err)

	assert.Equal(t, wrapper, &newWrapper)
}

func TestResourceMetadata_Serialization(t *testing.T) {
	metadata := ResourceMetadata{
		Name:      "test-configmap",
		Namespace: "kube-system",
		Kind:      "ConfigMap",
	}

	data, err := json.Marshal(metadata)
	assert.NoError(t, err)

	var newMetadata ResourceMetadata
	err = json.Unmarshal(data, &newMetadata)
	assert.NoError(t, err)

	assert.Equal(t, metadata, newMetadata)
}
