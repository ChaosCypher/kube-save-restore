package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"context"

	"github.com/chaoscypher/kube-save-restore/internal/backup"
	"github.com/chaoscypher/kube-save-restore/internal/config"
	"github.com/chaoscypher/kube-save-restore/internal/kubernetes"
	"github.com/chaoscypher/kube-save-restore/internal/logger"
	"github.com/chaoscypher/kube-save-restore/internal/restore"
)

// main is the entry point of the application.
func main() {
	config := config.ParseFlags()
	logger := logger.SetupLogger(config)

	if err := run(config, logger); err != nil {
		logger.Error("Error:", err)
		os.Exit(1)
	}
}

// run executes the main logic based on the provided configuration and logger.
func run(config *config.Config, logger logger.LoggerInterface) error {
	kubeconfigPath := getKubeconfigPath(config.KubeConfig, logger)

	k8sClient, err := kubernetes.NewClient(kubeconfigPath, config.Context, kubernetes.DefaultConfigModifier)
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
func getKubeconfigPath(kubeconfig string, logger logger.LoggerInterface) string {
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
func handleBackup(config *config.Config, k8sClient *kubernetes.Client, logger logger.LoggerInterface) error {
	if config.BackupDir == "" {
		config.BackupDir = filepath.Join(".", fmt.Sprintf("k8s-backup-%s", time.Now().Format("20060102-150405")))
	}
	backupManager := backup.NewManager(k8sClient, config.BackupDir, config.DryRun, logger)
	return backupManager.PerformBackup(context.Background())
}

// handleRestore performs the restore operation using the provided configuration and Kubernetes client.
func handleRestore(config *config.Config, k8sClient *kubernetes.Client, logger logger.LoggerInterface) error {
	if config.RestoreDir == "" {
		return fmt.Errorf("--restore-dir flag is required for restore mode")
	}
	restoreManager := restore.NewManager(k8sClient, logger)
	return restoreManager.PerformRestore(config.RestoreDir, config.DryRun)
}
