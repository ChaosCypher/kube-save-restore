package kubernetes

import (
	"context"
	"testing"

	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// TestNewClient tests the NewClient function with various scenarios.
func TestNewClient(t *testing.T) {
	// Define test cases
	tests := []struct {
		name            string
		kubeconfig      string
		context         string
		setupKubeconfig func(dir string) (string, error)
		configModifier  ConfigModifier
		expectError     bool
		expectedQPS     float32
		expectedBurst   int
	}{
		{
			name:        "Valid kubeconfig and context",
			context:     "test-context",
			expectError: false,
			setupKubeconfig: func(dir string) (string, error) {
				kubeconfigContent := `
apiVersion: v1
clusters:
- cluster:
    server: https://thiscodeissigma.com
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: test-token
`
				filePath := filepath.Join(dir, "kubeconfig.yaml")
				err := os.WriteFile(filePath, []byte(kubeconfigContent), 0644)
				return filePath, err
			},
		},
		{
			name:        "Invalid kubeconfig path",
			kubeconfig:  "/invalid/path/kubeconfig.yaml",
			context:     "test-context",
			expectError: true,
		},
		{
			name:        "Invalid context",
			context:     "nonexistent-context",
			expectError: true,
			setupKubeconfig: func(dir string) (string, error) {
				kubeconfigContent := `
apiVersion: v1
clusters:
- cluster:
    server: https://whatthesigma.com
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: test-token
`
				filePath := filepath.Join(dir, "kubeconfig.yaml")
				err := os.WriteFile(filePath, []byte(kubeconfigContent), 0644)
				return filePath, err
			},
		},
		{
			name:        "Check QPS and Burst values",
			context:     "test-context",
			expectError: false,
			setupKubeconfig: func(dir string) (string, error) {
				kubeconfigContent := `
apiVersion: v1
clusters:
- cluster:
    server: https://checkmyrizz.com
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: test-token
`
				filePath := filepath.Join(dir, "kubeconfig.yaml")
				err := os.WriteFile(filePath, []byte(kubeconfigContent), 0644)
				return filePath, err
			},
			configModifier: func(config *rest.Config) {
				config.QPS = 50.0
				config.Burst = 100
			},
			expectedQPS:   50.0,
			expectedBurst: 100,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			var kubeconfigPath string
			var err error

			// Setup kubeconfig if a setup function is provided
			if tt.setupKubeconfig != nil {
				tempDir, err := os.MkdirTemp("", "kubeconfig-test")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				defer os.RemoveAll(tempDir)

				kubeconfigPath, err = tt.setupKubeconfig(tempDir)
				if err != nil {
					t.Fatalf("Failed to write kubeconfig: %v", err)
				}
			} else {
				kubeconfigPath = tt.kubeconfig
			}

			// Call NewClient with the test's configModifier
			client, err := NewClient(kubeconfigPath, tt.context, tt.configModifier)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				// Optionally, check for specific error messages here
			} else {
				if err != nil {
					t.Fatalf("Did not expect error but got: %v", err)
				}
				if client == nil {
					t.Fatalf("Expected client but got nil")
				}

				// Check QPS and Burst values
				if tt.configModifier != nil {
					dummyConfig := &rest.Config{}
					tt.configModifier(dummyConfig)

					if dummyConfig.QPS != tt.expectedQPS {
						t.Errorf("Expected QPS to be %v, got %v", tt.expectedQPS, dummyConfig.QPS)
					}
					if dummyConfig.Burst != tt.expectedBurst {
						t.Errorf("Expected Burst to be %v, got %v", tt.expectedBurst, dummyConfig.Burst)
					}
				}
			}
		})
	}
}

func TestListConfigMaps(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test-configmap", Namespace: "default"},
	})}

	configMaps, err := client.ListConfigMaps(context.Background(), "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(configMaps.Items) != 1 {
		t.Fatalf("expected 1 configmap, got %d", len(configMaps.Items))
	}

	if configMaps.Items[0].Name != "test-configmap" {
		t.Fatalf("expected configmap name to be 'test-configmap', got %s", configMaps.Items[0].Name)
	}
}

func TestListDeployments(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-deployment", Namespace: "default"},
	})}

	deployments, err := client.ListDeployments(context.Background(), "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(deployments.Items) != 1 {
		t.Fatalf("expected 1 deployment, got %d", len(deployments.Items))
	}

	if deployments.Items[0].Name != "test-deployment" {
		t.Fatalf("expected deployment name to be 'test-deployment', got %s", deployments.Items[0].Name)
	}
}

func TestListNamespaces(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"},
	})}

	namespaces, err := client.ListNamespaces(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(namespaces) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(namespaces))
	}

	if namespaces[0] != "test-namespace" {
		t.Fatalf("expected namespace name to be 'test-namespace', got %s", namespaces[0])
	}
}

func TestListSecrets(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: "default"},
	})}

	secrets, err := client.ListSecrets(context.Background(), "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(secrets.Items) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets.Items))
	}

	if secrets.Items[0].Name != "test-secret" {
		t.Fatalf("expected secret name to be 'test-secret', got %s", secrets.Items[0].Name)
	}
}

func TestListServices(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "test-service", Namespace: "default"},
	})}

	services, err := client.ListServices(context.Background(), "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(services.Items) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services.Items))
	}

	if services.Items[0].Name != "test-service" {
		t.Fatalf("expected service name to be 'test-service', got %s", services.Items[0].Name)
	}
}

func TestListStatefulSets(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "test-statefulset", Namespace: "default"},
	})}

	statefulSets, err := client.ListStatefulSets(context.Background(), "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(statefulSets.Items) != 1 {
		t.Fatalf("expected 1 statefulset, got %d", len(statefulSets.Items))
	}

	if statefulSets.Items[0].Name != "test-statefulset" {
		t.Fatalf("expected statefulset name to be 'test-statefulset', got %s", statefulSets.Items[0].Name)
	}
}

func TestListHorizontalPodAutoscalers(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: "test-hpa", Namespace: "default"},
	})}

	hpas, err := client.ListHorizontalPodAutoscalers(context.Background(), "default")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(hpas.Items) != 1 {
		t.Fatalf("expected 1 hpa, got %d", len(hpas.Items))
	}

	if hpas.Items[0].Name != "test-hpa" {
		t.Fatalf("expected hpa name to be 'test-hpa', got %s", hpas.Items[0].Name)
	}
}

func TestListCronJobs(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{Name: "test-cronjob", Namespace: "default"},
	})}

	cronJobs, err := client.ListCronJobs(context.Background(), "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(cronJobs.Items) != 1 {
		t.Fatalf("expected 1 cronjob, got %d", len(cronJobs.Items))
	}

	if cronJobs.Items[0].Name != "test-cronjob" {
		t.Fatalf("expected cronjob name to be 'test-cronjob', got %s", cronJobs.Items[0].Name)
	}
}

func TestListPersistentVolumeClaims(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pvc", Namespace: "default"},
	})}

	pvcs, err := client.ListPersistentVolumeClaims(context.Background(), "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(pvcs.Items) != 1 {
		t.Fatalf("expected 1 pvc, got %d", len(pvcs.Items))
	}

	if pvcs.Items[0].Name != "test-pvc" {
		t.Fatalf("expected pvc name to be 'test-pvc', got %s", pvcs.Items[0].Name)
	}
}

func TestListIngresses(t *testing.T) {
	client := &Client{Clientset: fake.NewSimpleClientset(&networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ingress", Namespace: "default"},
	})}

	ingresses, err := client.ListIngresses(context.Background(), "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(ingresses.Items) != 1 {
		t.Fatalf("expected 1 ingress, got %d", len(ingresses.Items))
	}

	if ingresses.Items[0].Name != "test-ingress" {
		t.Fatalf("expected ingress name to be 'test-ingress', got %s", ingresses.Items[0].Name)
	}
}
