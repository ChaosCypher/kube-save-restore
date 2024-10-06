package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/chaoscypher/kube-save-restore/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MockKubernetesClient is a mock implementation of the KubernetesClient interface.
type MockKubernetesClient struct {
	mock.Mock
}

// ListNamespaces mocks the ListNamespaces method of the KubernetesClient interface.
func (m *MockKubernetesClient) ListNamespaces(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// ListDeployments mocks the ListDeployments method of the KubernetesClient interface.
func (m *MockKubernetesClient) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*appsv1.DeploymentList), args.Error(1)
}

// ListServices mocks the ListServices method of the KubernetesClient interface.
func (m *MockKubernetesClient) ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.ServiceList), args.Error(1)
}

// ListConfigMaps mocks the ListConfigMaps method of the KubernetesClient interface.
func (m *MockKubernetesClient) ListConfigMaps(ctx context.Context, namespace string) (*corev1.ConfigMapList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.ConfigMapList), args.Error(1)
}

// ListSecrets mocks the ListSecrets method of the KubernetesClient interface.
func (m *MockKubernetesClient) ListSecrets(ctx context.Context, namespace string) (*corev1.SecretList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.SecretList), args.Error(1)
}

// ListStatefulSets mocks the ListStatefulSets method of the KubernetesClient interface.
func (m *MockKubernetesClient) ListStatefulSets(ctx context.Context, namespace string) (*appsv1.StatefulSetList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*appsv1.StatefulSetList), args.Error(1)
}

// ListHorizontalPodAutoscalers mocks the ListHorizontalPodAutoscalers method of the KubernetesClient interface.
func (m *MockKubernetesClient) ListHorizontalPodAutoscalers(ctx context.Context, namespace string) (*autoscalingv2.HorizontalPodAutoscalerList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*autoscalingv2.HorizontalPodAutoscalerList), args.Error(1)
}

// ListCronJobs mocks the ListCronJobs method of the KubernetesClient interface.
func (m *MockKubernetesClient) ListCronJobs(ctx context.Context, namespace string) (*batchv1.CronJobList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*batchv1.CronJobList), args.Error(1)
}

// setupMockClient creates and configures a MockKubernetesClient with default expectations.
func setupMockClient() *MockKubernetesClient {
	mockClient := new(MockKubernetesClient)
	mockClient.On("ListNamespaces", mock.Anything).Return([]string{"default"}, nil)
	mockClient.On("ListDeployments", mock.Anything, "default").Return(&appsv1.DeploymentList{}, nil)
	mockClient.On("ListServices", mock.Anything, "default").Return(&corev1.ServiceList{}, nil)
	mockClient.On("ListConfigMaps", mock.Anything, "default").Return(&corev1.ConfigMapList{}, nil)
	mockClient.On("ListSecrets", mock.Anything, "default").Return(&corev1.SecretList{}, nil)
	mockClient.On("ListStatefulSets", mock.Anything, "default").Return(&appsv1.StatefulSetList{}, nil)
	mockClient.On("ListHorizontalPodAutoscalers", mock.Anything, "default").Return(&autoscalingv2.HorizontalPodAutoscalerList{}, nil)
	mockClient.On("ListCronJobs", mock.Anything, "default").Return(&batchv1.CronJobList{}, nil)
	return mockClient
}

// setupManager creates a new Manager instance with the given parameters.
func setupManager(mockClient *MockKubernetesClient, backupDir string, dryRun bool) *Manager {
	mockLogger := logger.NewLogger(os.Stdout, logger.DEBUG)
	return NewManager(mockClient, backupDir, dryRun, mockLogger)
}

// TestPerformBackup tests the PerformBackup method of the Manager.
func TestPerformBackup(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockKubernetesClient)
		expectedError  bool
		expectedLogs   []string
		dryRun         bool
		namespaces     []string
		resourceCounts map[string]int
	}{
		{
			name: "Successful backup",
			setupMock: func(m *MockKubernetesClient) {
				m.On("ListNamespaces", mock.Anything).Return([]string{"default", "kube-system"}, nil)
				setupMockResourceLists(m, []string{"default", "kube-system"})
			},
			expectedError: false,
			expectedLogs: []string{
				"Starting backup operation",
				"Found 2 namespaces",
				"Backup completed successfully",
			},
			dryRun:     false,
			namespaces: []string{"default", "kube-system"},
			resourceCounts: map[string]int{
				"deployments":  2,
				"services":     2,
				"configmaps":   2,
				"secrets":      2,
				"hpas":         2,
				"statefulsets": 2,
				"cronjobs":     2,
			},
		},
		{
			name: "Dry run backup",
			setupMock: func(m *MockKubernetesClient) {
				m.On("ListNamespaces", mock.Anything).Return([]string{"default"}, nil)
				setupMockResourceLists(m, []string{"default"})
			},
			expectedError: false,
			expectedLogs: []string{
				"Starting backup operation",
				"Dry run mode: No files will be written",
				"Backup completed successfully",
			},
			dryRun:     true,
			namespaces: []string{"default"},
			resourceCounts: map[string]int{
				"deployments":  1,
				"services":     1,
				"configmaps":   1,
				"secrets":      1,
				"hpas":         1,
				"statefulsets": 1,
				"cronjobs":     1,
			},
		},
		{
			name: "Error listing namespaces",
			setupMock: func(m *MockKubernetesClient) {
				m.On("ListNamespaces", mock.Anything).Return([]string{}, fmt.Errorf("namespace list error"))
			},
			expectedError: true,
			expectedLogs: []string{
				"Starting backup operation",
				"error listing namespaces: namespace list error",
			},
			dryRun:         false,
			namespaces:     []string{},
			resourceCounts: map[string]int{},
		},
		{
			name: "Error backing up resource",
			setupMock: func(m *MockKubernetesClient) {
				m.On("ListNamespaces", mock.Anything).Return([]string{"default"}, nil)
				m.On("ListDeployments", mock.Anything, "default").Return(&appsv1.DeploymentList{}, fmt.Errorf("deployment list error"))
				setupMockResourceLists(m, []string{"default"})
			},
			expectedError: false,
			expectedLogs: []string{
				"Starting backup operation",
				"Error during backup:",
				"Completed backup with 1 errors",
			},
			dryRun:     false,
			namespaces: []string{"default"},
			resourceCounts: map[string]int{
				"services":     1,
				"configmaps":   1,
				"secrets":      1,
				"hpas":         1,
				"statefulsets": 1,
				"cronjobs":     1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary backup directory
			backupDir, err := os.MkdirTemp("", "k8s-backup-test")
			if err != nil {
				t.Fatalf("Failed to create temporary backup directory: %v", err)
			}
			defer os.RemoveAll(backupDir)

			mockClient := new(MockKubernetesClient)
			tt.setupMock(mockClient)

			var logBuffer bytes.Buffer
			mockLogger := logger.NewLogger(&logBuffer, logger.DEBUG)
			manager := NewManager(mockClient, backupDir, tt.dryRun, mockLogger)

			err = manager.PerformBackup(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			logOutput := logBuffer.String()
			for _, expectedLog := range tt.expectedLogs {
				assert.Contains(t, logOutput, expectedLog)
			}

			if !tt.dryRun {
				// Verify that the backup directory is created
				info, err := os.Stat(backupDir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			}

			// Verify resource counts
			for resourceType, expectedCount := range tt.resourceCounts {
				assert.Equal(t, expectedCount, manager.resourceCounts[resourceType], "Incorrect count for %s", resourceType)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func setupMockResourceLists(m *MockKubernetesClient, namespaces []string) {
	for _, ns := range namespaces {
		m.On("ListDeployments", mock.Anything, ns).Return(&appsv1.DeploymentList{Items: []appsv1.Deployment{{}}}, nil)
		m.On("ListServices", mock.Anything, ns).Return(&corev1.ServiceList{Items: []corev1.Service{{}}}, nil)
		m.On("ListConfigMaps", mock.Anything, ns).Return(&corev1.ConfigMapList{Items: []corev1.ConfigMap{{}}}, nil)
		m.On("ListSecrets", mock.Anything, ns).Return(&corev1.SecretList{Items: []corev1.Secret{{}}}, nil)
		m.On("ListStatefulSets", mock.Anything, ns).Return(&appsv1.StatefulSetList{Items: []appsv1.StatefulSet{{}}}, nil)
		m.On("ListHorizontalPodAutoscalers", mock.Anything, ns).Return(&autoscalingv2.HorizontalPodAutoscalerList{Items: []autoscalingv2.HorizontalPodAutoscaler{{}}}, nil)
		m.On("ListCronJobs", mock.Anything, ns).Return(&batchv1.CronJobList{Items: []batchv1.CronJob{{}}}, nil)
	}
}

// TestPerformBackupDryRun tests the PerformBackup method of the Manager in dry-run mode.
func TestPerformBackupDryRun(t *testing.T) {
	backupDir := filepath.Join(os.TempDir(), "k8s-backup-test")
	mockClient := setupMockClient()
	manager := setupManager(mockClient, backupDir, true)

	// Perform the backup in dry-run mode
	err := manager.PerformBackup(context.Background())
	assert.NoError(t, err)

	// Verify that the backup directory is not created in dry-run mode
	_, err = os.Stat(backupDir)
	assert.True(t, os.IsNotExist(err))

	// Verify that all expected methods were called on the mock client
	mockClient.AssertExpectations(t)
}

// TestBackupResource tests the backupResource method
func TestBackupResource(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-resource-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name         string
		resourceType string
		setupMock    func(*MockKubernetesClient)
		expectError  bool
	}{
		{
			name:         "Successful deployment backup",
			resourceType: "deployments",
			setupMock: func(m *MockKubernetesClient) {
				m.On("ListDeployments", mock.Anything, mock.Anything).Return(&appsv1.DeploymentList{
					Items: []appsv1.Deployment{{
						ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
					}},
				}, nil)
			},
			expectError: false,
		},
		{
			name:         "Successful service backup",
			resourceType: "services",
			setupMock: func(m *MockKubernetesClient) {
				m.On("ListServices", mock.Anything, mock.Anything).Return(&corev1.ServiceList{
					Items: []corev1.Service{{
						ObjectMeta: metav1.ObjectMeta{Name: "test-service"},
					}},
				}, nil)
			},
			expectError: false,
		},
		{
			name:         "Error listing deployments",
			resourceType: "deployments",
			setupMock: func(m *MockKubernetesClient) {
				m.On("ListDeployments", mock.Anything, mock.Anything).Return(&appsv1.DeploymentList{}, fmt.Errorf("error listing deployments"))
			},
			expectError: true,
		},
		{
			name:         "Unknown resource type",
			resourceType: "unknown",
			setupMock:    func(m *MockKubernetesClient) {},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockKubernetesClient)
			tt.setupMock(mockClient)

			mockLogger := logger.NewLogger(os.Stdout, logger.DEBUG)
			manager := NewManager(mockClient, tempDir, false, mockLogger)

			ctx := context.Background()
			namespace := "default"

			err := manager.backupResource(ctx, tt.resourceType, namespace)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if !manager.dryRun {
					_, err := os.Stat(filepath.Join(tempDir, namespace, tt.resourceType))
					assert.NoError(t, err)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// TestBackupDeployments tests the backupDeployments method
func TestBackupDeployments(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-deployments-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockClient := new(MockKubernetesClient)
	mockLogger := logger.NewLogger(os.Stdout, logger.DEBUG)
	manager := NewManager(mockClient, tempDir, false, mockLogger)

	ctx := context.Background()
	namespace := "default"

	mockClient.On("ListDeployments", ctx, namespace).Return(&appsv1.DeploymentList{
		Items: []appsv1.Deployment{{
			ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
		}},
	}, nil)

	err = manager.backupDeployments(ctx, namespace)
	assert.NoError(t, err)

	// Check if the file was created
	_, err = os.Stat(filepath.Join(tempDir, namespace, "deployments", "test-deployment.json"))
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

// TestSaveResource tests the saveResource method
func TestSaveResource(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "save-resource-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockLogger := logger.NewLogger(os.Stdout, logger.DEBUG)
	manager := NewManager(nil, tempDir, false, mockLogger)

	tests := []struct {
		name     string
		resource interface{}
		kind     string
		filename string
	}{
		{
			name: "Save deployment",
			resource: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
			},
			kind:     "Deployment",
			filename: "test-deployment.json",
		},
		{
			name: "Save service",
			resource: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "test-service"},
			},
			kind:     "Service",
			filename: "test-service.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(tempDir, tt.filename)
			err := manager.saveResource(tt.resource, tt.kind, filename)
			assert.NoError(t, err)

			// Check if the file was created
			_, err = os.Stat(filename)
			assert.NoError(t, err)

			// Read the file content
			data, err := os.ReadFile(filename)
			assert.NoError(t, err)

			var savedResource struct {
				Kind     string      `json:"kind"`
				Resource interface{} `json:"resource"`
			}
			err = json.Unmarshal(data, &savedResource)
			assert.NoError(t, err)
			assert.Equal(t, tt.kind, savedResource.Kind)
		})
	}
}

// TestIncrementResourceCount tests the incrementResourceCount method
func TestIncrementResourceCount(t *testing.T) {
	manager := NewManager(nil, "", false, nil)

	tests := []struct {
		name         string
		resourceType string
		increment    int
		expected     int
	}{
		{
			name:         "Increment deployments",
			resourceType: "deployments",
			increment:    3,
			expected:     3,
		},
		{
			name:         "Increment services",
			resourceType: "services",
			increment:    2,
			expected:     2,
		},
		{
			name:         "Increment multiple times",
			resourceType: "configmaps",
			increment:    5,
			expected:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.increment; i++ {
				manager.incrementResourceCount(tt.resourceType)
			}
			assert.Equal(t, tt.expected, manager.resourceCounts[tt.resourceType])
		})
	}
}

// TestLogCompletionMessage tests the logCompletionMessage method
func TestLogCompletionMessage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "log-completion-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var logBuffer bytes.Buffer
	mockLogger := logger.NewLogger(&logBuffer, logger.DEBUG)

	tests := []struct {
		name           string
		dryRun         bool
		resourceCounts map[string]int
		expectedLogs   []string
	}{
		{
			name:   "Normal backup completion",
			dryRun: false,
			resourceCounts: map[string]int{
				"deployments": 2,
				"services":    3,
			},
			expectedLogs: []string{
				"Backup completed. 5 resources saved to:",
				"Backed up 2 deployments",
				"Backed up 3 services",
			},
		},
		{
			name:   "Dry run completion",
			dryRun: true,
			resourceCounts: map[string]int{
				"configmaps": 1,
				"secrets":    2,
			},
			expectedLogs: []string{
				"Dry run completed. 3 resources would be backed up to:",
				"Backed up 1 configmaps",
				"Backed up 2 secrets",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer.Reset()
			manager := NewManager(nil, tempDir, tt.dryRun, mockLogger)
			manager.resourceCounts = tt.resourceCounts

			manager.logCompletionMessage()

			logOutput := logBuffer.String()
			for _, expectedLog := range tt.expectedLogs {
				assert.Contains(t, logOutput, expectedLog)
			}
		})
	}
}

// TestNewManager tests the NewManager function
func TestNewManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-newmanager-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockClient := new(MockKubernetesClient)
	mockLogger := logger.NewLogger(os.Stdout, logger.DEBUG)
	dryRun := false

	manager := NewManager(mockClient, tempDir, dryRun, mockLogger)

	assert.NotNil(t, manager)
	assert.Equal(t, mockClient, manager.client)
	assert.Equal(t, tempDir, manager.backupDir)
	assert.Equal(t, dryRun, manager.dryRun)
	assert.Equal(t, mockLogger, manager.logger)
	assert.NotNil(t, manager.resourceCounts)
}

func TestIncrementResourceCountConcurrency(t *testing.T) {
	manager := NewManager(nil, "", false, nil)
	resourceType := "deployments"
	numGoroutines := 100
	incrementsPerGoroutine := 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				manager.incrementResourceCount(resourceType)
			}
		}()
	}

	wg.Wait()

	expectedCount := numGoroutines * incrementsPerGoroutine
	assert.Equal(t, expectedCount, manager.resourceCounts[resourceType])
}
