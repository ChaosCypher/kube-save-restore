package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultNamespace = "default"
	maxConcurrency   = 10
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

// setupMockClient creates and configures a MockKubernetesClient with default expectations.
func setupMockClient() *MockKubernetesClient {
	mockClient := new(MockKubernetesClient)
	mockClient.On("ListNamespaces", mock.Anything).Return([]string{defaultNamespace}, nil)
	mockClient.On("ListDeployments", mock.Anything, defaultNamespace).Return(&appsv1.DeploymentList{}, nil)
	mockClient.On("ListServices", mock.Anything, defaultNamespace).Return(&corev1.ServiceList{}, nil)
	mockClient.On("ListConfigMaps", mock.Anything, defaultNamespace).Return(&corev1.ConfigMapList{}, nil)
	mockClient.On("ListSecrets", mock.Anything, defaultNamespace).Return(&corev1.SecretList{}, nil)
	return mockClient
}

// setupManager creates a new Manager instance with the given parameters.
func setupManager(mockClient *MockKubernetesClient, backupDir string, dryRun bool) *Manager {
	mockLogger := logger.NewLogger(os.Stdout, logger.DEBUG)
	return NewManager(mockClient, backupDir, dryRun, mockLogger, maxConcurrency)
}

// TestPerformBackup tests the PerformBackup method of the Manager.
func TestPerformBackup(t *testing.T) {
	t.Helper()

	backupDir, err := os.MkdirTemp("", "k8s-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temporary backup directory: %v", err)
	}
	defer os.RemoveAll(backupDir)

	mockClient := setupMockClient()
	manager := setupManager(mockClient, backupDir, false)

	// Perform the backup
	err = manager.PerformBackup(context.Background())
	assert.NoError(t, err)

	// Verify that the backup directory is created
	info, err := os.Stat(backupDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify that all expected methods were called on the mock client
	mockClient.AssertExpectations(t)
}

// TestPerformBackupDryRun tests the PerformBackup method of the Manager in dry-run mode.
func TestPerformBackupDryRun(t *testing.T) {
	t.Helper()

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
