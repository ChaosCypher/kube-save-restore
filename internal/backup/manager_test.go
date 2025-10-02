package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chaoscypher/kube-save-restore/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// MockKubernetesClient is a mock implementation of the KubernetesClient interface
type MockKubernetesClient struct {
	mock.Mock
}

// ListNamespaces mocks the ListNamespaces method of the KubernetesClient interface
func (m *MockKubernetesClient) ListNamespaces(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// GetNamespaces mocks the GetNamespaces method of the KubernetesClient interface
func (m *MockKubernetesClient) GetNamespaces(ctx context.Context) (*corev1.NamespaceList, error) {
	args := m.Called(ctx)
	return args.Get(0).(*corev1.NamespaceList), args.Error(1)
}

// ListDeployments mocks the ListDeployments method of the KubernetesClient interface
func (m *MockKubernetesClient) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*appsv1.DeploymentList), args.Error(1)
}

// ListServices mocks the ListServices method of the KubernetesClient interface
func (m *MockKubernetesClient) ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.ServiceList), args.Error(1)
}

// ListConfigMaps mocks the ListConfigMaps method of the KubernetesClient interface
func (m *MockKubernetesClient) ListConfigMaps(ctx context.Context, namespace string) (*corev1.ConfigMapList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.ConfigMapList), args.Error(1)
}

// ListSecrets mocks the ListSecrets method of the KubernetesClient interface
func (m *MockKubernetesClient) ListSecrets(ctx context.Context, namespace string) (*corev1.SecretList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.SecretList), args.Error(1)
}

// ListServiceAccounts mocks the ListServiceAccounts method of the KubernetesClient interface
func (m *MockKubernetesClient) ListServiceAccounts(ctx context.Context, namespace string) (*corev1.ServiceAccountList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.ServiceAccountList), args.Error(1)
}

// ListStatefulSets mocks the ListStatefulSets method of the KubernetesClient interface
func (m *MockKubernetesClient) ListStatefulSets(ctx context.Context, namespace string) (*appsv1.StatefulSetList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*appsv1.StatefulSetList), args.Error(1)
}

// ListHorizontalPodAutoscalers mocks the ListHorizontalPodAutoscalers method of the KubernetesClient interface
func (m *MockKubernetesClient) ListHorizontalPodAutoscalers(ctx context.Context, namespace string) (*autoscalingv2.HorizontalPodAutoscalerList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*autoscalingv2.HorizontalPodAutoscalerList), args.Error(1)
}

// ListCronJobs mocks the ListCronJobs method of the KubernetesClient interface.
func (m *MockKubernetesClient) ListCronJobs(ctx context.Context, namespace string) (*batchv1.CronJobList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*batchv1.CronJobList), args.Error(1)
}

// ListJobs mocks the ListJobs method of the KubernetesClient interface
func (m *MockKubernetesClient) ListJobs(ctx context.Context, namespace string) (*batchv1.JobList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*batchv1.JobList), args.Error(1)
}

// ListPersistentVolumeClaims mocks the ListPersistentVolumeClaims method of the KubernetesClient interface
func (m *MockKubernetesClient) ListPersistentVolumeClaims(ctx context.Context, namespace string) (*corev1.PersistentVolumeClaimList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev1.PersistentVolumeClaimList), args.Error(1)
}

// ListDaemonSets mocks the ListDaemonSets method of the KubernetesClient interface
func (m *MockKubernetesClient) ListDaemonSets(ctx context.Context, namespace string) (*appsv1.DaemonSetList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*appsv1.DaemonSetList), args.Error(1)
}

// ListIngresses mocks the ListIngresses method of the KubernetesClient interface
func (m *MockKubernetesClient) ListIngresses(ctx context.Context, namespace string) (*networkingv1.IngressList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*networkingv1.IngressList), args.Error(1)
}

// ListNetworkPolicies mocks the ListNetworkPolicies method of the KubernetesClient interface
func (m *MockKubernetesClient) ListNetworkPolicies(ctx context.Context, namespace string) (*networkingv1.NetworkPolicyList, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*networkingv1.NetworkPolicyList), args.Error(1)
}

// setupMockClient creates and configures a MockKubernetesClient with default expectations
func setupMockClient() *MockKubernetesClient {
	mockClient := new(MockKubernetesClient)
	mockClient.On("ListNamespaces", mock.Anything).Return([]string{"default", "kube-system"}, nil)
	mockClient.On("GetNamespaces", mock.Anything).Return(&corev1.NamespaceList{Items: make([]corev1.Namespace, 2)}, nil)

	// Set up expectations for the default namespace
	mockClient.On("ListDeployments", mock.Anything, "default").Return(&appsv1.DeploymentList{Items: make([]appsv1.Deployment, 1)}, nil)
	mockClient.On("ListServices", mock.Anything, "default").Return(&corev1.ServiceList{Items: make([]corev1.Service, 1)}, nil)
	mockClient.On("ListConfigMaps", mock.Anything, "default").Return(&corev1.ConfigMapList{Items: make([]corev1.ConfigMap, 1)}, nil)
	mockClient.On("ListSecrets", mock.Anything, "default").Return(&corev1.SecretList{Items: make([]corev1.Secret, 1)}, nil)
	mockClient.On("ListServiceAccounts", mock.Anything, "default").Return(&corev1.ServiceAccountList{Items: make([]corev1.ServiceAccount, 1)}, nil)
	mockClient.On("ListStatefulSets", mock.Anything, "default").Return(&appsv1.StatefulSetList{Items: make([]appsv1.StatefulSet, 1)}, nil)
	mockClient.On("ListHorizontalPodAutoscalers", mock.Anything, "default").Return(&autoscalingv2.HorizontalPodAutoscalerList{Items: make([]autoscalingv2.HorizontalPodAutoscaler, 1)}, nil)
	mockClient.On("ListDaemonSets", mock.Anything, "default").Return(&appsv1.DaemonSetList{Items: make([]appsv1.DaemonSet, 1)}, nil)
	mockClient.On("ListCronJobs", mock.Anything, "default").Return(&batchv1.CronJobList{Items: make([]batchv1.CronJob, 1)}, nil)
	mockClient.On("ListJobs", mock.Anything, "default").Return(&batchv1.JobList{Items: make([]batchv1.Job, 1)}, nil)
	mockClient.On("ListPersistentVolumeClaims", mock.Anything, "default").Return(&corev1.PersistentVolumeClaimList{Items: make([]corev1.PersistentVolumeClaim, 1)}, nil)
	mockClient.On("ListIngresses", mock.Anything, "default").Return(&networkingv1.IngressList{Items: make([]networkingv1.Ingress, 1)}, nil)
	mockClient.On("ListNetworkPolicies", mock.Anything, "default").Return(&networkingv1.NetworkPolicyList{Items: make([]networkingv1.NetworkPolicy, 1)}, nil)
	// Set up expectations for the kube-system namespace
	mockClient.On("ListDeployments", mock.Anything, "kube-system").Return(&appsv1.DeploymentList{Items: make([]appsv1.Deployment, 2)}, nil)
	mockClient.On("ListServices", mock.Anything, "kube-system").Return(&corev1.ServiceList{Items: make([]corev1.Service, 3)}, nil)
	mockClient.On("ListConfigMaps", mock.Anything, "kube-system").Return(&corev1.ConfigMapList{Items: make([]corev1.ConfigMap, 4)}, nil)
	mockClient.On("ListSecrets", mock.Anything, "kube-system").Return(&corev1.SecretList{Items: make([]corev1.Secret, 5)}, nil)
	mockClient.On("ListServiceAccounts", mock.Anything, "kube-system").Return(&corev1.ServiceAccountList{Items: make([]corev1.ServiceAccount, 3)}, nil)
	mockClient.On("ListStatefulSets", mock.Anything, "kube-system").Return(&appsv1.StatefulSetList{Items: make([]appsv1.StatefulSet, 1)}, nil)
	mockClient.On("ListHorizontalPodAutoscalers", mock.Anything, "kube-system").Return(&autoscalingv2.HorizontalPodAutoscalerList{Items: make([]autoscalingv2.HorizontalPodAutoscaler, 2)}, nil)
	mockClient.On("ListDaemonSets", mock.Anything, "kube-system").Return(&appsv1.DaemonSetList{Items: make([]appsv1.DaemonSet, 2)}, nil)
	mockClient.On("ListCronJobs", mock.Anything, "kube-system").Return(&batchv1.CronJobList{Items: make([]batchv1.CronJob, 1)}, nil)
	mockClient.On("ListJobs", mock.Anything, "kube-system").Return(&batchv1.JobList{Items: make([]batchv1.Job, 1)}, nil)
	mockClient.On("ListPersistentVolumeClaims", mock.Anything, "kube-system").Return(&corev1.PersistentVolumeClaimList{Items: make([]corev1.PersistentVolumeClaim, 2)}, nil)
	mockClient.On("ListIngresses", mock.Anything, "kube-system").Return(&networkingv1.IngressList{Items: make([]networkingv1.Ingress, 2)}, nil)
	mockClient.On("ListNetworkPolicies", mock.Anything, "kube-system").Return(&networkingv1.NetworkPolicyList{Items: make([]networkingv1.NetworkPolicy, 2)}, nil)
	return mockClient
}

// setupManager creates a new Manager instance with the given parameters
func setupManager(mockClient *MockKubernetesClient, backupDir string, dryRun bool) *Manager {
	mockLogger := logger.NewLogger(os.Stdout, logger.DEBUG)
	return NewManager(mockClient, backupDir, dryRun, mockLogger)
}

// TestPerformBackup tests the PerformBackup method of the Manager
func TestPerformBackup(t *testing.T) {
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

// TestPerformBackupDryRun tests the PerformBackup method of the Manager in dry-run mode
func TestPerformBackupDryRun(t *testing.T) {
	backupDir := filepath.Join(os.TempDir(), "k8s-backup-test")

	defer os.RemoveAll(backupDir)

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

// TestCountResources tests that the correct number of resources are counted correctly
func TestCountResources(t *testing.T) {
	backupDir := filepath.Join(os.TempDir(), "k8s-backup-test")
	defer os.RemoveAll(backupDir)

	mockClient := setupMockClient()
	manager := setupManager(mockClient, backupDir, false)

	count := manager.countResources(context.Background())

	// The total should be the sum of all resources in both namespaces
	// and cluster-wide resources
	// default namespace: 13 (1 of each resource type)
	// kube-system namespace: 30
	// cluster-wide: 2
	expectedCount := 45
	assert.Equal(t, expectedCount, count)

	mockClient.AssertExpectations(t)
}
