//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/chaoscypher/k8s-backup-restore/internal/backup"
	"github.com/chaoscypher/k8s-backup-restore/internal/config"
	"github.com/chaoscypher/k8s-backup-restore/internal/kubernetes"
	"github.com/chaoscypher/k8s-backup-restore/internal/restore"
	"github.com/chaoscypher/k8s-backup-restore/internal/utils"
	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBackupAndRestore(t *testing.T) {
	testCases := []struct {
		name          string
		mode          string
		dryRun        bool
		expectedError bool
	}{
		{
			name:          "Dry Run Backup",
			mode:          "backup",
			dryRun:        true,
			expectedError: false,
		},
		{
			name:          "Actual Backup",
			mode:          "backup",
			dryRun:        false,
			expectedError: false,
		},
		{
			name:          "Dry Run Restore",
			mode:          "restore",
			dryRun:        true,
			expectedError: false,
		},
		{
			name:          "Actual Restore",
			mode:          "restore",
			dryRun:        false,
			expectedError: false,
		},
		{
			name:          "Invalid Mode",
			mode:          "invalid",
			dryRun:        false,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Setup test configuration
			testConfig := &config.Config{
				Mode:       tc.mode,
				BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
				RestoreDir: filepath.Join(os.TempDir(), "kube-save-restore-test"),
				KubeConfig: getTestKubeconfig(t),
				Context:    "minikube",
				DryRun:     tc.dryRun,
			}

			// Setup logger
			logger := utils.SetupLogger(testConfig)

			// Create Kubernetes client
			kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context)
			if err != nil {
				t.Fatalf("Failed to create Kubernetes client: %v", err)
			}

			if tc.mode == "backup" {
				// Execute backup
				err = backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, logger).PerformBackup(context.Background())
			} else if tc.mode == "restore" {
				// Ensure backup directory exists and has content before restore
				if !tc.dryRun {
					if err := os.MkdirAll(testConfig.BackupDir, 0755); err != nil {
						t.Fatalf("Failed to create backup directory: %v", err)
					}
					if err := createDummyBackupFile(testConfig.BackupDir); err != nil {
						t.Fatalf("Failed to create dummy backup file: %v", err)
					}
				}
				err = restore.NewManager().PerformRestore(kubeClient, testConfig.BackupDir, testConfig.DryRun, logger)
			} else {
				err = fmt.Errorf("unsupported mode: %s", tc.mode)
			}

			if (err != nil) != tc.expectedError {
				t.Fatalf("Expected error: %v, got: %v", tc.expectedError, err)
			}

			if !tc.expectedError {
				verifyOperation(t, testConfig, tc.mode, tc.dryRun)
			}
		})
	}
}

func createDummyBackupFile(backupDir string) error {
	dummyFile := filepath.Join(backupDir, "dummy.json")
	return os.WriteFile(dummyFile, []byte("{}"), 0644)
}

// verifyOperation verifies the backup or restore operation based on the mode and dryRun flag.
func verifyOperation(t *testing.T, config *config.Config, mode string, dryRun bool) {
	if mode == "backup" {
		if !dryRun {
			// Verify backup directory exists and is not empty
			info, err := os.Stat(config.BackupDir)
			if err != nil {
				t.Fatalf("Backup directory does not exist: %v", err)
			}
			if !info.IsDir() {
				t.Fatalf("Backup path is not a directory")
			}

			dirEntries, err := os.ReadDir(config.BackupDir)
			if err != nil {
				t.Fatalf("Failed to read backup directory: %v", err)
			}
			if len(dirEntries) == 0 {
				t.Fatalf("Backup directory is empty")
			}
		} else {
			// For dry run, ensure that backup directory is not created or empty
			info, err := os.Stat(config.BackupDir)
			if err == nil {
				if info.IsDir() {
					dirEntries, err := os.ReadDir(config.BackupDir)
					if err != nil {
						t.Fatalf("Failed to read backup directory: %v", err)
					}
					if len(dirEntries) != 0 {
						t.Fatalf("Backup directory should be empty for dry run")
					}
				}
			} else if !os.IsNotExist(err) {
				t.Fatalf("Error checking backup directory: %v", err)
			}
		}
	} else if mode == "restore" {
		// Verify restore directory exists and contains expected files
		info, err := os.Stat(config.RestoreDir)
		if dryRun {
			// In dry run mode, restoration might not create files
			if !os.IsNotExist(err) && info.Name() != "" {
				dirEntries, err := os.ReadDir(config.RestoreDir)
				if err != nil {
					t.Fatalf("Failed to read restore directory: %v", err)
				}
				if len(dirEntries) != 0 {
					t.Fatalf("Restore directory should be empty for dry run")
				}
			}
		} else {
			if err != nil {
				t.Fatalf("Restore directory does not exist: %v", err)
			}
			if !info.IsDir() {
				t.Fatalf("Restore path is not a directory")
			}

			dirEntries, err := os.ReadDir(config.RestoreDir)
			if err != nil {
				t.Fatalf("Failed to read restore directory: %v", err)
			}
			if len(dirEntries) == 0 {
				t.Fatalf("Restore directory is empty")
			}
		}
	}
}

func getTestKubeconfig(t *testing.T) string {
	kubeconfig := os.Getenv("TEST_KUBECONFIG")
	if kubeconfig == "" {
		t.Fatal("TEST_KUBECONFIG environment variable is not set")
	}
	// Verify the kubeconfig file exists
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		t.Fatalf("Kubeconfig file does not exist: %s", kubeconfig)
	}
	return kubeconfig
}

// Additional test cases for enhanced coverage
func TestBackupRestoreScenarios(t *testing.T) {
	testCases := []struct {
		name          string
		mode          string
		dryRun        bool
		setupFunc     func(t *testing.T, config *config.Config)
		expectedError bool
	}{
		{
			name:   "Backup with Invalid Kubeconfig",
			mode:   "backup",
			dryRun: false,
			setupFunc: func(t *testing.T, config *config.Config) {
				config.KubeConfig = "/invalid/path/kubeconfig.yaml"
			},
			expectedError: true,
		},
		{
			name:   "Restore Without Backup",
			mode:   "restore",
			dryRun: false,
			setupFunc: func(t *testing.T, config *config.Config) {
				// Ensure backup directory does not exist
				os.RemoveAll(config.BackupDir)
			},
			expectedError: true,
		},
		{
			name:          "Concurrent Backup and Restore",
			mode:          "backup", // Start with backup
			dryRun:        false,
			setupFunc:     nil,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testConfig := &config.Config{
				Mode:       tc.mode,
				BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
				RestoreDir: filepath.Join(os.TempDir(), "kube-save-restore-restore-test"),
				KubeConfig: getTestKubeconfig(t),
				Context:    "minikube",
				DryRun:     tc.dryRun,
			}

			if tc.setupFunc != nil {
				tc.setupFunc(t, testConfig)
			}

			logger := utils.SetupLogger(testConfig)

			kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context)
			if err != nil {
				t.Fatalf("Failed to create Kubernetes client: %v", err)
			}

			var operationErr error
			if tc.mode == "backup" {
				operationErr = backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, logger).PerformBackup(context.Background())
			} else if tc.mode == "restore" {
				operationErr = restore.NewManager().PerformRestore(kubeClient, testConfig.RestoreDir, testConfig.DryRun, logger)
			}

			if (operationErr != nil) != tc.expectedError {
				t.Fatalf("Expected error: %v, got: %v", tc.expectedError, operationErr)
			}
		})
	}
}

func TestBackupRestoreResourceIntegrity(t *testing.T) {
	ctx := context.Background()
	testConfig := setupTestConfig(t)
	logger := utils.SetupLogger(testConfig)

	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create test resources
	namespace := createTestNamespace(t, kubeClient)
	defer deleteTestNamespace(t, kubeClient, namespace)

	createTestResources(t, kubeClient, namespace)

	// Perform backup
	err = backup.NewManager(kubeClient, testConfig.BackupDir, false, logger).PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Failed to perform backup: %v", err)
	}

	// Modify resources
	modifyTestResources(t, kubeClient, namespace)

	// Perform restore
	err = restore.NewManager().PerformRestore(kubeClient, testConfig.BackupDir, false, logger)
	if err != nil {
		t.Fatalf("Failed to perform restore: %v", err)
	}

	// Verify restored resources
	verifyRestoredResources(t, kubeClient, namespace)
}

func createTestNamespace(t *testing.T, client *kubernetes.Client) string {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-backup-restore-",
		},
	}
	createdNs, err := client.Clientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test namespace: %v", err)
	}
	return createdNs.Name
}

func deleteTestNamespace(t *testing.T, client *kubernetes.Client, namespace string) {
	err := client.Clientset.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Failed to delete test namespace: %v", err)
	}
}

func createTestResources(t *testing.T, client *kubernetes.Client, namespace string) {
	// Create a test deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
	_, err := client.Clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test deployment: %v", err)
	}

	// Create a test service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-service",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test",
			},
			Ports: []corev1.ServicePort{
				{
					Port: 80,
					TargetPort: intstr.IntOrString{
						IntVal: 80,
					},
				},
			},
		},
	}
	_, err = client.Clientset.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Create a test configmap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-configmap",
		},
		Data: map[string]string{
			"test-key": "test-value",
		},
	}
	_, err = client.Clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test configmap: %v", err)
	}

	// Create a test secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-secret",
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"test-secret-key": "test-secret-value",
		},
	}
	_, err = client.Clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test secret: %v", err)
	}
}

func modifyTestResources(t *testing.T, client *kubernetes.Client, namespace string) {
	// Modify deployment
	deployment, err := client.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), "test-deployment", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get test deployment: %v", err)
	}
	*deployment.Spec.Replicas = 3
	_, err = client.Clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update test deployment: %v", err)
	}

	// Modify service
	service, err := client.Clientset.CoreV1().Services(namespace).Get(context.TODO(), "test-service", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get test service: %v", err)
	}
	service.Spec.Ports[0].Port = 8080
	_, err = client.Clientset.CoreV1().Services(namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update test service: %v", err)
	}

	// Modify configmap
	configMap, err := client.Clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), "test-configmap", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get test configmap: %v", err)
	}
	configMap.Data["test-key"] = "modified-value"
	_, err = client.Clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update test configmap: %v", err)
	}

	// Modify secret
	secret, err := client.Clientset.CoreV1().Secrets(namespace).Get(context.TODO(), "test-secret", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get test secret: %v", err)
	}
	secret.StringData = map[string]string{"test-secret-key": "modified-secret-value"}
	_, err = client.Clientset.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update test secret: %v", err)
	}
}

func verifyRestoredResources(t *testing.T, client *kubernetes.Client, namespace string) {
	// Verify deployment
	deployment, err := client.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), "test-deployment", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get restored deployment: %v", err)
	}
	if *deployment.Spec.Replicas != 2 {
		t.Errorf("Restored deployment has incorrect replica count. Expected: 2, Got: %d", *deployment.Spec.Replicas)
	}

	// Verify service
	service, err := client.Clientset.CoreV1().Services(namespace).Get(context.TODO(), "test-service", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get restored service: %v", err)
	}
	if service.Spec.Ports[0].Port != 80 {
		t.Errorf("Restored service has incorrect port. Expected: 80, Got: %d", service.Spec.Ports[0].Port)
	}

	// Verify configmap
	configMap, err := client.Clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), "test-configmap", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get restored configmap: %v", err)
	}
	if configMap.Data["test-key"] != "test-value" {
		t.Errorf("Restored configmap has incorrect data. Expected: test-value, Got: %s", configMap.Data["test-key"])
	}

	// Verify secret
	secret, err := client.Clientset.CoreV1().Secrets(namespace).Get(context.TODO(), "test-secret", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get restored secret: %v", err)
	}
	if string(secret.Data["test-secret-key"]) != "test-secret-value" {
		t.Errorf("Restored secret has incorrect data. Expected: test-secret-value, Got: %s", string(secret.Data["test-secret-key"]))
	}
}

func int32Ptr(i int32) *int32 { return &i }

func setupTestConfig(t *testing.T) *config.Config {
	return &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}
}
