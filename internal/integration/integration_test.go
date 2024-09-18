package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"k8s-backup-restore/internal/backup"
	"k8s-backup-restore/internal/config"
	"k8s-backup-restore/internal/kubernetes"
	"k8s-backup-restore/internal/restore"
	"k8s-backup-restore/internal/utils"
)

// TestRunBackup tests the backup functionality of the application.
func TestRunBackup(t *testing.T) {
	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "backup",
		BackupDir:  filepath.Join(os.TempDir(), "k8s-backup-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "test-context",
		DryRun:     false,
	}

	// Ensure backup directory is clean
	defer os.RemoveAll(testConfig.BackupDir)

	// Setup logger
	logger := utils.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Execute backup
	err = backup.NewBackupManager(kubeClient, testConfig.BackupDir, testConfig.DryRun, logger).PerformBackup(context.Background())
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

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
}

// TestRunRestore tests the restore functionality of the application.
func TestRunRestore(t *testing.T) {
	// Setup test configuration
	testConfig := &config.Config{
		Mode:       "restore",
		RestoreDir: filepath.Join(os.TempDir(), "k8s-restore-test"),
		KubeConfig: getTestKubeconfig(t),
		Context:    "test-context",
		DryRun:     false,
	}

	// Ensure restore directory exists with necessary backup files
	err := os.MkdirAll(testConfig.RestoreDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create restore directory: %v", err)
	}
	// Simulate backup files (this should be replaced with actual backup data)
	testBackupFile := filepath.Join(testConfig.RestoreDir, "backup.json")
	err = os.WriteFile(testBackupFile, []byte(`{"dummy": "data"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy backup file: %v", err)
	}
	defer os.RemoveAll(testConfig.RestoreDir)

	// Setup logger
	logger := utils.SetupLogger(testConfig)

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Execute restore
	err = restore.NewRestoreManager().PerformRestore(kubeClient, testConfig.RestoreDir, testConfig.DryRun, logger)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
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
