package backup

import (
	"context"
	"fmt"

	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
	"github.com/chaoscypher/k8s-backup-restore/internal/workerpool"
)

const maxConcurrency = 10

type Manager struct {
	client     KubernetesClient
	backupDir  string
	dryRun     bool
	logger     logger.LoggerInterface
	workerPool *workerpool.WorkerPool
}

// NewManager creates a new Manager instance.
func NewManager(client KubernetesClient, backupDir string, dryRun bool, logger logger.LoggerInterface) *Manager {
	return &Manager{
		client:     client,
		backupDir:  backupDir,
		dryRun:     dryRun,
		logger:     logger,
		workerPool: workerpool.NewWorkerPool(maxConcurrency),
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

	for _, ns := range namespaces {
		bm.enqueueTasks(ns)
	}

	bm.workerPool.Close()
	errors := bm.workerPool.Run(ctx)

	if len(errors) > 0 {
		bm.logger.Errorf("Encountered %d errors during backup", len(errors))
		for _, err := range errors {
			bm.logger.Error(err)
		}
	}

	bm.logCompletionMessage(totalResources)
	return nil
}
