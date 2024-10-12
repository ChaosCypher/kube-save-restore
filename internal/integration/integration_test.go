//go:build integration

package integration

import (
	"context"
	"fmt"
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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TestRunBackup tests the backup functionality of the application
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

// TestRunRestore tests the restore functionality of the application
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

// Helper function to get the kubeconfig path for testing
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

// setupTestEnvironment sets up the test environment
func setupTestEnvironment(t *testing.T) (*config.Config, logger.LoggerInterface, *kubernetes.Client) {
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "kube-save-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "minikube",
		DryRun:     false,
	}

	log := logger.SetupLogger(testConfig)

	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context, kubernetes.DefaultConfigModifier)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	return testConfig, log, kubeClient
}

// createTestNamespace creates a test namespace
func createTestNamespace(ctx context.Context, kubeClient *kubernetes.Client, name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err := kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	return err
}

// performBackupAndRestore performs a backup and restore
func performBackupAndRestore(testConfig *config.Config, kubeClient *kubernetes.Client, log logger.LoggerInterface) error {
	backupManager := backup.NewManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, log)
	err := backupManager.PerformBackup(context.Background())
	if err != nil {
		return fmt.Errorf("Backup failed: %v", err)
	}

	testConfig.Mode = "restore"
	testConfig.RestoreDir = testConfig.BackupDir
	restoreManager := restore.NewManager(kubeClient, log)
	err = restoreManager.PerformRestore(testConfig.RestoreDir, testConfig.DryRun)
	if err != nil {
		return fmt.Errorf("Restore failed: %v", err)
	}

	return nil
}

// TestBackupAndRestoreDeployment tests the backup and restore of a deployment
func TestBackupAndRestoreDeployment(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-deployment-namespace"
	deploymentName := "test-deployment"

	testConfig, log, kubeClient := setupTestEnvironment(t)

	err := createTestNamespace(ctx, kubeClient, testNamespace)
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

	err = performBackupAndRestore(testConfig, kubeClient, log)
	if err != nil {
		t.Fatal(err)
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

// TestBackupAndRestoreHorizontalPodAutoscaler tests the backup and restore of a HorizontalPodAutoscaler
func TestBackupAndRestoreHorizontalPodAutoscaler(t *testing.T) {
	ctx := context.Background()
	deploymentName := "test-hpa-deployment"
	testNamespace := "test-hpa-namespace"
	hpaName := "test-hpa"

	testConfig, log, kubeClient := setupTestEnvironment(t)

	err := createTestNamespace(ctx, kubeClient, testNamespace)
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

	err = performBackupAndRestore(testConfig, kubeClient, log)
	if err != nil {
		t.Fatal(err)
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

// TestBackupAndRestoreSecret tests the backup and restore of a Secret
func TestBackupAndRestoreSecret(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-secret-namespace"
	secretName := "test-secret"

	testConfig, log, kubeClient := setupTestEnvironment(t)

	err := createTestNamespace(ctx, kubeClient, testNamespace)
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

	err = performBackupAndRestore(testConfig, kubeClient, log)
	if err != nil {
		t.Fatal(err)
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

// TestBackupAndRestoreConfigMap tests the backup and restore of a ConfigMap
func TestBackupAndRestoreConfigMap(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-configmap-namespace"
	configMapName := "test-configmap"

	testConfig, log, kubeClient := setupTestEnvironment(t)

	err := createTestNamespace(ctx, kubeClient, testNamespace)
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

	err = performBackupAndRestore(testConfig, kubeClient, log)
	if err != nil {
		t.Fatal(err)
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

// TestBackupAndRestoreService tests the backup and restore of a Service
func TestBackupAndRestoreService(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-service-namespace"
	serviceName := "test-service"

	testConfig, log, kubeClient := setupTestEnvironment(t)

	err := createTestNamespace(ctx, kubeClient, testNamespace)
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

	err = performBackupAndRestore(testConfig, kubeClient, log)
	if err != nil {
		t.Fatal(err)
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

// TestBackupAndRestoreStatefulSet tests the backup and restore of a StatefulSet
func TestBackupAndRestoreStatefulSet(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-statefulset-namespace"
	statefulsetName := "test-statefulset"

	testConfig, log, kubeClient := setupTestEnvironment(t)

	err := createTestNamespace(ctx, kubeClient, testNamespace)
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

	err = performBackupAndRestore(testConfig, kubeClient, log)
	if err != nil {
		t.Fatal(err)
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

// TestBackupAndRestoreCronJob tests the backup and restore of a CronJob
func TestBackupAndRestoreCronJob(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-cronjob-namespace"
	cronJobName := "test-cronjob"

	testConfig, log, kubeClient := setupTestEnvironment(t)

	err := createTestNamespace(ctx, kubeClient, testNamespace)
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

	err = performBackupAndRestore(testConfig, kubeClient, log)
	if err != nil {
		t.Fatal(err)
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

// TestBackupAndRestorePersistentVolumeClaim tests the backup and restore of a PersistentVolumeClaim
func TestBackupAndRestorePersistentVolumeClaim(t *testing.T) {
	ctx := context.Background()
	testNamespace := "test-pvc-namespace"
	pvcName := "test-pvc"

	testConfig, log, kubeClient := setupTestEnvironment(t)

	err := createTestNamespace(ctx, kubeClient, testNamespace)
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

	err = performBackupAndRestore(testConfig, kubeClient, log)
	if err != nil {
		t.Fatal(err)
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

// int32Ptr returns a pointer to an int32
func int32Ptr(i int32) *int32 { return &i }
