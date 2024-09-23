package backup

import (
	"context"
	"fmt"

	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
	"github.com/chaoscypher/k8s-backup-restore/internal/workerpool"
)

const maxConcurrency = 10

type Manager struct {
	client    KubernetesClient
	backupDir string
	dryRun    bool
	logger    logger.LoggerInterface
}

// NewManager creates a new Manager instance.
func NewManager(client KubernetesClient, backupDir string, dryRun bool, logger logger.LoggerInterface) *Manager {
	return &Manager{
		client:    client,
		backupDir: backupDir,
		dryRun:    dryRun,
		logger:    logger,
	}
}

// PerformBackup initiates the backup process for all namespaces.
func (bm *Manager) PerformBackup(ctx context.Context) error {
	bm.logger.Info("Starting backup operation")

	namespaces, err := bm.client.ListNamespaces(ctx)
	if err != nil {
		return fmt.Errorf("error listing namespaces: %v", err)
	}

	totalResources := bm.countResources(ctx, namespaces)

	if bm.dryRun {
		bm.logger.Info("Dry run mode: No files will be written")
	}

	wp := workerpool.NewWorkerPool(maxConcurrency, 1000)
	bm.enqueueTasks(namespaces, wp)

	errors := wp.Run(ctx)
	if len(errors) > 0 {
		for _, err := range errors {
			bm.logger.Errorf("Error during backup: %v", err)
		}
	}

	bm.logCompletionMessage(totalResources)
	return nil
}
