//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chaoscypher/kube-save-restore/internal/backup"
	"github.com/chaoscypher/kube-save-restore/internal/config"
	"github.com/chaoscypher/kube-save-restore/internal/kubernetes"
	"github.com/chaoscypher/kube-save-restore/internal/logger"
	"github.com/chaoscypher/kube-save-restore/internal/restore"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TestRunBackup tests the backup functionality of the application.
func TestRunBackup(t *testing.T) {
	testCases := []struct {
		name   string
		dryRun bool
	}{
		{
			name:   "Dry Run Backup",
			dryRun: true,
		},
		{
			name:   "Actual Backup",
			dryRun: false,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Setup test configuration
			testConfig := &config.Config{
				Mode:       "backup",
				BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
				KubeConfig: getTestKubeconfig(t),
				Context:    "minikube",
				DryRun:     tc.dryRun,
			}

			// Setup logger
			logger := logger.SetupLogger(testConfig)

			// Create Kubernetes client
			kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
			if err != nil {
				t.Fatalf("Failed to create Kubernetes client: %v", err)
			}

			// Execute backup
			err = backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, logger).PerformBackup(context.Background())
			if err != nil {
				t.Fatalf("Backup failed: %v", err)
			}

			if !tc.dryRun {
				// Verify backup directory exists and is not empty
				info, err := os.Stat(testConfig.BackupDir)
				if err != nil {
					t.Fatalf("Backup directory does not exist: %v", err)
				}
				if !info.IsDir() {
					t.Fatalf("Backup path is not a directory")
				}

				dirEntries, err := os.ReadDir(testConfig.BackupDir)
				if err != nil {
					t.Fatalf("Failed to read backup directory: %v", err)
				}
				if len(dirEntries) == 0 {
					t.Fatalf("Backup directory is empty")
				}
			} else {
				// For dry run, ensure that backup directory is not created or empty
				info, err := os.Stat(testConfig.BackupDir)
				if err == nil {
					if info.IsDir() {
						dirEntries, err := os.ReadDir(testConfig.BackupDir)
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
		})
	}
}

// TestRunRestore tests the restore functionality of the application.
func TestRunRestore(t *testing.T) {
	testCases := []struct {
		name   string
		dryRun bool
	}{
		{
			name:   "Dry Run Restore",
			dryRun: true,
		},
		{
			name:   "Actual Restore",
			dryRun: false,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Setup test configuration
			testConfig := &config.Config{
				Mode:       "restore",
				RestoreDir: filepath.Join(os.TempDir(), "kube-save-restore-test"),
				KubeConfig: getTestKubeconfig(t),
				Context:    "minikube",
				DryRun:     tc.dryRun,
			}

			// Setup logger
			logger := logger.SetupLogger(testConfig)

			// Create Kubernetes client
			kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
			if err != nil {
				t.Fatalf("Failed to create Kubernetes client: %v", err)
			}

			// Execute restore
			restoreManager := restore.NewManager(kubeClient, logger)
			err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
			if err != nil {
				t.Fatalf("Restore failed: %v", err)
			}

			if !tc.dryRun {
				// Verify restore directory exists and contains expected files
				info, err := os.Stat(testConfig.RestoreDir)
				if err != nil {
					t.Fatalf("Restore directory does not exist: %v", err)
				}
				if !info.IsDir() {
					t.Fatalf("Restore path is not a directory")
				}

				dirEntries, err := os.ReadDir(testConfig.RestoreDir)
				if err != nil {
					t.Fatalf("Failed to read restore directory: %v", err)
				}
				if len(dirEntries) == 0 {
					t.Fatalf("Restore directory is empty")
				}
			}
		})
	}
}

// Helper function to get the kubeconfig path for testing.
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

func TestBackupAndRestoreDeployment(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-deployment-namespace"
	deploymentName := "test-deployment"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: testNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "busybox",
							Image: "busybox:latest",
						},
					},
				},
			},
		},
	}

	_, err = kubeClient.Clientset.AppsV1().Deployments(testNamespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create deployment: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the deployment
	deployment, err = kubeClient.Clientset.AppsV1().Deployments(testNamespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}
	deployment.Spec.Replicas = int32Ptr(2)
	_, err = kubeClient.Clientset.AppsV1().Deployments(testNamespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the deployment's settings were restored
	deployment, err = kubeClient.Clientset.AppsV1().Deployments(testNamespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}
	if *deployment.Spec.Replicas != 1 {
		t.Fatalf("Expected replicas to be 1, got %d", *deployment.Spec.Replicas)
	}
}

func TestBackupAndRestoreHorizontalPodAutoscaler(t *testing.T) {
	ctx := context.Background()
	deploymentName := "test-hpa-deployment"
	testNamespace := "test-hpa-namespace"
	hpaName := "test-hpa"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: testNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "busybox",
							Image: "busybox:latest",
						},
					},
				},
			},
		},
	}
	_, err = kubeClient.Clientset.AppsV1().Deployments(testNamespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create deployment: %v", err)
	}

	// Create a HorizontalPodAutoscaler
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hpaName,
			Namespace: testNamespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       deploymentName,
			},
			MinReplicas: int32Ptr(1),
			MaxReplicas: 10,
		},
	}
	_, err = kubeClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(testNamespace).Create(ctx, hpa, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create HorizontalPodAutoscaler: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the HPA
	hpa, err = kubeClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(testNamespace).Get(ctx, hpaName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get HorizontalPodAutoscaler: %v", err)
	}
	hpa.Spec.MinReplicas = int32Ptr(2)
	_, err = kubeClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(testNamespace).Update(ctx, hpa, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update HorizontalPodAutoscaler: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the HPA's settings were restored
	hpa, err = kubeClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(testNamespace).Get(ctx, hpaName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get HorizontalPodAutoscaler: %v", err)
	}
	if *hpa.Spec.MinReplicas != 1 {
		t.Fatalf("Expected min replicas to be 1, got %d", *hpa.Spec.MinReplicas)
	}
}

func TestBackupAndRestoreSecret(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-secret-namespace"
	secretName := "test-secret"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("admin"),
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Secrets(testNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create Secret: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the Secret
	secret, err = kubeClient.Clientset.CoreV1().Secrets(testNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Secret: %v", err)
	}
	secret.Data["password"] = []byte("newpassword")
	_, err = kubeClient.Clientset.CoreV1().Secrets(testNamespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update Secret: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the Secret's settings were restored
	secret, err = kubeClient.Clientset.CoreV1().Secrets(testNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Secret: %v", err)
	}
	if _, ok := secret.Data["password"]; !ok {
		t.Fatalf("Expected password to be restored, but it was not found")
	}
}

func TestBackupAndRestoreConfigMap(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-configmap-namespace"
	configMapName := "test-configmap"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: testNamespace,
		},
		Data: map[string]string{
			"config": "value",
		},
	}
	_, err = kubeClient.Clientset.CoreV1().ConfigMaps(testNamespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create ConfigMap: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the ConfigMap
	configMap, err = kubeClient.Clientset.CoreV1().ConfigMaps(testNamespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get ConfigMap: %v", err)
	}
	configMap.Data["config"] = "newvalue"
	_, err = kubeClient.Clientset.CoreV1().ConfigMaps(testNamespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update ConfigMap: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the ConfigMap's settings were restored
	configMap, err = kubeClient.Clientset.CoreV1().ConfigMaps(testNamespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get ConfigMap: %v", err)
	}
	if _, ok := configMap.Data["config"]; !ok {
		t.Fatalf("Expected config to be restored, but it was not found")
	}
}

func TestBackupAndRestoreService(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-service-namespace"
	serviceName := "test-service"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: testNamespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "test"},
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Services(testNamespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create Service: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the Service
	service, err = kubeClient.Clientset.CoreV1().Services(testNamespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Service: %v", err)
	}
	service.Spec.Type = corev1.ServiceTypeLoadBalancer
	_, err = kubeClient.Clientset.CoreV1().Services(testNamespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update Service: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the Service's settings were restored
	service, err = kubeClient.Clientset.CoreV1().Services(testNamespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Service: %v", err)
	}
	if service.Spec.Type != corev1.ServiceTypeClusterIP {
		t.Fatalf("Expected service type to be ClusterIP, got %v", service.Spec.Type)
	}
}

func TestBackupAndRestoreStatefulSet(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-statefulset-namespace"
	statefulsetName := "test-statefulset"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a statefulset
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      statefulsetName,
			Namespace: testNamespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "busybox",
							Image: "busybox:latest",
						},
					},
				},
			},
		},
	}

	_, err = kubeClient.Clientset.AppsV1().StatefulSets(testNamespace).Create(ctx, statefulset, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create statefulset: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the statefulset
	statefulset, err = kubeClient.Clientset.AppsV1().StatefulSets(testNamespace).Get(ctx, statefulsetName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get statefulset: %v", err)
	}
	statefulset.Spec.Replicas = int32Ptr(2)
	_, err = kubeClient.Clientset.AppsV1().StatefulSets(testNamespace).Update(ctx, statefulset, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update statefulset: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the statefulset's settings were restored
	statefulset, err = kubeClient.Clientset.AppsV1().StatefulSets(testNamespace).Get(ctx, statefulsetName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get statefulset: %v", err)
	}
	if *statefulset.Spec.Replicas != 1 {
		t.Fatalf("Expected replicas to be 1, got %d", *statefulset.Spec.Replicas)
	}
}

func TestBackupAndRestoreDaemonSet(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-daemonset-namespace"
	daemonsetName := "test-daemonset"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a daemonset
	daemonset := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonsetName,
			Namespace: testNamespace,
			Labels: map[string]string{
				"app": "test",
			},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "busybox",
							Image:   "busybox:latest",
							Command: []string{"sleep", "3600"},
						},
					},
				},
			},
		},
	}

	_, err = kubeClient.Clientset.AppsV1().DaemonSets(testNamespace).Create(ctx, daemonset, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create daemonset: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the daemonset
	daemonset, err = kubeClient.Clientset.AppsV1().DaemonSets(testNamespace).Get(ctx, daemonsetName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get daemonset: %v", err)
	}
	daemonset.Labels["environment"] = "production"
	daemonset.Spec.Template.Spec.Containers[0].Image = "busybox:1.35"
	_, err = kubeClient.Clientset.AppsV1().DaemonSets(testNamespace).Update(ctx, daemonset, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update daemonset: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the daemonset's settings were restored
	daemonset, err = kubeClient.Clientset.AppsV1().DaemonSets(testNamespace).Get(ctx, daemonsetName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get daemonset: %v", err)
	}
	if _, exists := daemonset.Labels["environment"]; exists {
		t.Fatalf("Expected environment label to be removed, but it still exists")
	}
	if daemonset.Spec.Template.Spec.Containers[0].Image != "busybox:latest" {
		t.Fatalf("Expected image to be busybox:latest, got %s", daemonset.Spec.Template.Spec.Containers[0].Image)
	}

	// Verify backup directory structure
	daemonsetBackupDir := filepath.Join(testConfig.BackupDir, testNamespace, "daemonsets")
	info, err := os.Stat(daemonsetBackupDir)
	if err != nil {
		t.Fatalf("DaemonSet backup directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("DaemonSet backup path is not a directory")
	}

	// Verify the daemonset backup file exists
	daemonsetBackupFile := filepath.Join(daemonsetBackupDir, daemonsetName+".json")
	if _, err := os.Stat(daemonsetBackupFile); err != nil {
		t.Fatalf("DaemonSet backup file does not exist: %v", err)
	}
}

func TestBackupAndRestoreCronJob(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-cronjob-namespace"
	cronJobName := "test-cronjob"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a CronJob
	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronJobName,
			Namespace: testNamespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "0 0 * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "busybox",
									Image: "busybox:latest",
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}

	_, err = kubeClient.Clientset.BatchV1().CronJobs(testNamespace).Create(ctx, cronJob, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create cron job: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the CronJob
	cronJob, err = kubeClient.Clientset.BatchV1().CronJobs(testNamespace).Get(ctx, cronJobName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get cron job: %v", err)
	}
	cronJob.Spec.Schedule = "0 1 * * *"
	_, err = kubeClient.Clientset.BatchV1().CronJobs(testNamespace).Update(ctx, cronJob, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update cron job: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the CronJob's settings were restored
	cronJob, err = kubeClient.Clientset.BatchV1().CronJobs(testNamespace).Get(ctx, cronJobName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get cron job: %v", err)
	}
	if cronJob.Spec.Schedule != "0 0 * * *" {
		t.Fatalf("Expected schedule to be 0 0 * * *, got %s", cronJob.Spec.Schedule)
	}
}

func TestBackupAndRestoreJob(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-job-namespace"
	jobName := "test-job"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: testNamespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "busybox",
							Image:   "busybox:latest",
							Command: []string{"echo", "hello world"},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	_, err = kubeClient.Clientset.BatchV1().Jobs(testNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the Job
	job, err = kubeClient.Clientset.BatchV1().Jobs(testNamespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}
	if job.Labels == nil {
		job.Labels = make(map[string]string)
	}
	job.Labels["test-label"] = "updated"
	_, err = kubeClient.Clientset.BatchV1().Jobs(testNamespace).Update(ctx, job, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update job: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the Job's settings were restored to original
	job, err = kubeClient.Clientset.BatchV1().Jobs(testNamespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}
	if _, exists := job.Labels["test-label"]; exists {
		t.Fatalf("Expected test-label to be removed, but it still exists")
	}
}

func TestBackupAndRestorePersistentVolumeClaim(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-pvc-namespace"
	pvcName := "test-pvc"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a PersistentVolumeClaim
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: testNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}
	_, err = kubeClient.Clientset.CoreV1().PersistentVolumeClaims(testNamespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create persistent volume claim: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the PersistentVolumeClaim
	pvc, err = kubeClient.Clientset.CoreV1().PersistentVolumeClaims(testNamespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get persistent volume claim: %v", err)
	}
	if pvc.Labels == nil {
		pvc.Labels = make(map[string]string)
	}
	pvc.Labels["test-label"] = "updated"
	_, err = kubeClient.Clientset.CoreV1().PersistentVolumeClaims(testNamespace).Update(ctx, pvc, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update persistent volume claim: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the PersistentVolumeClaim's settings were restored to original
	pvc, err = kubeClient.Clientset.CoreV1().PersistentVolumeClaims(testNamespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get persistent volume claim: %v", err)
	}
	if _, exists := pvc.Labels["test-label"]; exists {
		t.Fatalf("Expected test-label to be removed, but it still exists")
	}
}

func TestBackupAndRestoreServiceAccount(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-serviceaccount-namespace"
	serviceAccountName := "test-serviceaccount"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a ServiceAccount
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: testNamespace,
			Labels: map[string]string{
				"app": "test",
			},
		},
		AutomountServiceAccountToken: boolPtr(true),
	}
	_, err = kubeClient.Clientset.CoreV1().ServiceAccounts(testNamespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create ServiceAccount: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the ServiceAccount
	serviceAccount, err = kubeClient.Clientset.CoreV1().ServiceAccounts(testNamespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get ServiceAccount: %v", err)
	}
	serviceAccount.Labels["environment"] = "production"
	serviceAccount.AutomountServiceAccountToken = boolPtr(false)
	_, err = kubeClient.Clientset.CoreV1().ServiceAccounts(testNamespace).Update(ctx, serviceAccount, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update ServiceAccount: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the ServiceAccount's settings were restored
	serviceAccount, err = kubeClient.Clientset.CoreV1().ServiceAccounts(testNamespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get ServiceAccount: %v", err)
	}
	if _, exists := serviceAccount.Labels["environment"]; exists {
		t.Fatalf("Expected environment label to be removed, but it still exists")
	}
	if serviceAccount.AutomountServiceAccountToken == nil || *serviceAccount.AutomountServiceAccountToken != true {
		t.Fatalf("Expected AutomountServiceAccountToken to be true, got %v", serviceAccount.AutomountServiceAccountToken)
	}

	// Verify backup directory structure
	serviceAccountBackupDir := filepath.Join(testConfig.BackupDir, testNamespace, "serviceaccounts")
	info, err := os.Stat(serviceAccountBackupDir)
	if err != nil {
		t.Fatalf("ServiceAccount backup directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("ServiceAccount backup path is not a directory")
	}

	// Verify the serviceaccount backup file exists
	serviceAccountBackupFile := filepath.Join(serviceAccountBackupDir, serviceAccountName+".json")
	if _, err := os.Stat(serviceAccountBackupFile); err != nil {
		t.Fatalf("ServiceAccount backup file does not exist: %v", err)
	}
}

func TestBackupAndRestoreIngress(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-ingress-namespace"
	ingressName := "test-ingress"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create an Ingress
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: testNamespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "test.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/app",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "test-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = kubeClient.Clientset.NetworkingV1().Ingresses(testNamespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create Ingress: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the Ingress
	ingress, err = kubeClient.Clientset.NetworkingV1().Ingresses(testNamespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Ingress: %v", err)
	}
	ingress.Spec.Rules[0].Host = "updated.example.com"
	_, err = kubeClient.Clientset.NetworkingV1().Ingresses(testNamespace).Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update Ingress: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the Ingress's settings were restored
	ingress, err = kubeClient.Clientset.NetworkingV1().Ingresses(testNamespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Ingress: %v", err)
	}
	if ingress.Spec.Rules[0].Host != "test.example.com" {
		t.Fatalf("Expected host to be test.example.com, got %s", ingress.Spec.Rules[0].Host)
	}

	// Verify backup directory structure
	ingressBackupDir := filepath.Join(testConfig.BackupDir, testNamespace, "ingresses")
	info, err := os.Stat(ingressBackupDir)
	if err != nil {
		t.Fatalf("Ingress backup directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("Ingress backup path is not a directory")
	}

	// Verify the ingress backup file exists
	ingressBackupFile := filepath.Join(ingressBackupDir, ingressName+".json")
	if _, err := os.Stat(ingressBackupFile); err != nil {
		t.Fatalf("Ingress backup file does not exist: %v", err)
	}
}

func TestBackupAndRestoreRole(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-role-namespace"
	roleName := "test-role"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a Role
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: testNamespace,
			Labels: map[string]string{
				"app": "test",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	_, err = kubeClient.Clientset.RbacV1().Roles(testNamespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create Role: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the Role
	role, err = kubeClient.Clientset.RbacV1().Roles(testNamespace).Get(ctx, roleName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Role: %v", err)
	}
	role.Rules = append(role.Rules, rbacv1.PolicyRule{
		APIGroups: []string{""},
		Resources: []string{"services"},
		Verbs:     []string{"get", "list"},
	})
	_, err = kubeClient.Clientset.RbacV1().Roles(testNamespace).Update(ctx, role, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update Role: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the Role's settings were restored
	role, err = kubeClient.Clientset.RbacV1().Roles(testNamespace).Get(ctx, roleName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Role: %v", err)
	}
	if len(role.Rules) != 1 {
		t.Fatalf("Expected 1 policy rule, got %d", len(role.Rules))
	}
	if len(role.Rules[0].Resources) != 1 || role.Rules[0].Resources[0] != "pods" {
		t.Fatalf("Expected resource 'pods', got %v", role.Rules[0].Resources)
	}

	// Verify backup directory structure
	roleBackupDir := filepath.Join(testConfig.BackupDir, testNamespace, "roles")
	info, err := os.Stat(roleBackupDir)
	if err != nil {
		t.Fatalf("Role backup directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("Role backup path is not a directory")
	}

	// Verify the role backup file exists
	roleBackupFile := filepath.Join(roleBackupDir, roleName+".json")
	if _, err := os.Stat(roleBackupFile); err != nil {
		t.Fatalf("Role backup file does not exist: %v", err)
	}
}

func TestBackupAndRestoreRoleBinding(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-rolebinding-namespace"
	roleBindingName := "test-rolebinding"

	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	// Setup logger
	log := logger.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create the test namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// Create a RoleBinding
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: testNamespace,
			Labels: map[string]string{
				"app": "test",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: testNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "view",
		},
	}
	_, err = kubeClient.Clientset.RbacV1().RoleBindings(testNamespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create RoleBinding: %v", err)
	}

	// Perform backup
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err = backupManager.PerformBackup(ctx)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Edit the RoleBinding
	roleBinding, err = kubeClient.Clientset.RbacV1().RoleBindings(testNamespace).Get(ctx, roleBindingName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get RoleBinding: %v", err)
	}
	roleBinding.Labels["environment"] = "production"
	roleBinding.Subjects = append(roleBinding.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      "extra-sa",
		Namespace: testNamespace,
	})
	_, err = kubeClient.Clientset.RbacV1().RoleBindings(testNamespace).Update(ctx, roleBinding, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update RoleBinding: %v", err)
	}

	// Perform restore
	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Confirm the RoleBinding's settings were restored
	roleBinding, err = kubeClient.Clientset.RbacV1().RoleBindings(testNamespace).Get(ctx, roleBindingName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get RoleBinding: %v", err)
	}
	if _, exists := roleBinding.Labels["environment"]; exists {
		t.Fatalf("Expected environment label to be removed, but it still exists")
	}
	if len(roleBinding.Subjects) != 1 || roleBinding.Subjects[0].Name != "default" {
		t.Fatalf("Expected Subjects to be restored to original single subject 'default', got %v", roleBinding.Subjects)
	}

	// Verify backup directory structure
	roleBindingBackupDir := filepath.Join(testConfig.BackupDir, testNamespace, "rolebindings")
	info, err := os.Stat(roleBindingBackupDir)
	if err != nil {
		t.Fatalf("RoleBinding backup directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("RoleBinding backup path is not a directory")
	}

	// Verify the rolebinding backup file exists
	roleBindingBackupFile := filepath.Join(roleBindingBackupDir, roleBindingName+".json")
	if _, err := os.Stat(roleBindingBackupFile); err != nil {
		t.Fatalf("RoleBinding backup file does not exist: %v", err)
	}
}

func int32Ptr(i int32) *int32 { return &i }

func boolPtr(b bool) *bool { return &b }
