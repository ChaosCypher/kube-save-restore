package backup

import (
	"context"
	"fmt"

	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
	"github.com/chaoscypher/k8s-backup-restore/internal/workerpool"
)

// Manager handles the backup process for Kubernetes resources.
type Manager struct {
	client         KubernetesClient
	backupDir      string
	dryRun         bool
	logger         logger.LoggerInterface
	maxConcurrency int
}

// NewManager creates a new Manager instance.
func NewManager(client KubernetesClient, backupDir string, dryRun bool, logger logger.LoggerInterface, maxConcurrency int) *Manager {
	return &Manager{
		client:         client,
		backupDir:      backupDir,
		dryRun:         dryRun,
		logger:         logger,
		maxConcurrency: maxConcurrency,
	}
}

// PerformBackup initiates the backup process for all namespaces.
func (bm *Manager) PerformBackup(ctx context.Context) error {
	bm.logger.Info("Starting backup operation")

	bm.logger.Debug("Fetching namespaces")
	// List all namespaces
	namespaces, err := bm.client.ListNamespaces(ctx)
	if err != nil {
		return fmt.Errorf("error listing namespaces: %v", err)
	}

	// Count total resources to be backed up
	totalResources := bm.countResources(ctx, namespaces)

	if bm.dryRun {
		bm.logger.Info("Dry run mode: No files will be written")
	}

	// Create a worker pool for concurrent backup operations
	wp := workerpool.NewWorkerPool(bm.maxConcurrency, 1000)
	bm.enqueueTasks(namespaces, wp)

	// Run the worker pool and collect any errors
	errors := wp.Run(ctx)
	if len(errors) > 0 {
		for _, err := range errors {
			bm.logger.Errorf("Error during backup: %v", err)
		}
	}

	bm.logCompletionMessage(totalResources)
	return nil
}
