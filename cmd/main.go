package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"context"
	"k8s-backup-restore/internal/backup"
	"k8s-backup-restore/internal/config"
	"k8s-backup-restore/internal/kubernetes"
	"k8s-backup-restore/internal/restore"
	"k8s-backup-restore/internal/utils"
)

// main is the entry point of the application.
func main() {
	config := config.ParseFlags()
	logger := utils.SetupLogger(config)

	if err := run(config, logger); err != nil {
		logger.Error("Error:", err)
		os.Exit(1)
	}
}

// run executes the main logic based on the provided configuration and logger.
func run(config *config.Config, logger *utils.Logger) error {
	kubeconfigPath := getKubeconfigPath(config.KubeConfig, logger)

	k8sClient, err := kubernetes.NewClient(kubeconfigPath, config.Context)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	switch config.Mode {
	case "backup":
		return handleBackup(config, k8sClient, logger)
	case "restore":
		return handleRestore(config, k8sClient, logger)
	default:
		return fmt.Errorf("invalid mode: %s. Use 'backup' or 'restore'", config.Mode)
	}
}

// getKubeconfigPath returns the path to the kubeconfig file.
// If the kubeconfig path is not provided, it defaults to the user's home directory.
func getKubeconfigPath(kubeconfig string, logger *utils.Logger) string {
	if kubeconfig != "" {
		return kubeconfig
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Error getting user home directory: %v", err)
		os.Exit(1)
	}
	return filepath.Join(homeDir, ".kube", "config")
}

// handleBackup performs the backup operation using the provided configuration and Kubernetes client.
func handleBackup(config *config.Config, k8sClient *kubernetes.Client, logger *utils.Logger) error {
	if config.BackupDir == "" {
		config.BackupDir = filepath.Join(".", fmt.Sprintf("k8s-backup-%s", time.Now().Format("20060102-150405")))
	}
	backupManager := backup.NewBackupManager(k8sClient, config.BackupDir, config.DryRun, logger)
	return backupManager.PerformBackup(context.Background())
}

// handleRestore performs the restore operation using the provided configuration and Kubernetes client.
func handleRestore(config *config.Config, k8sClient *kubernetes.Client, logger *utils.Logger) error {
	if config.RestoreDir == "" {
		return fmt.Errorf("--restore-dir flag is required for restore mode")
	}
	restoreManager := restore.NewRestoreManager()
	return restoreManager.PerformRestore(k8sClient, config.RestoreDir, config.DryRun, logger)
}
