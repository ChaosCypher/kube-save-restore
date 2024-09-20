//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chaoscypher/k8s-backup-restore/internal/backup"
	"github.com/chaoscypher/k8s-backup-restore/internal/config"
	"github.com/chaoscypher/k8s-backup-restore/internal/kubernetes"
	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
	"github.com/chaoscypher/k8s-backup-restore/internal/restore"
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
			kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context)
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
			kubeClient, err := kubernetes.NewClient(testConfig.KubeConfig, testConfig.Context)
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
