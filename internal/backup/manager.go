package backup

import (
	"context"
	"fmt"

	"github.com/chaoscypher/kube-save-restore/internal/workerpool"
)

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Close()
}

// maxConcurrency defines the maximum number of concurrent backup operations.
const maxConcurrency = 10

// Manager handles the backup process for Kubernetes resources.
type Manager struct {
	client    KubernetesClient
	backupDir string
	dryRun    bool
	logger    Logger
}

// NewManager creates a new Manager instance.
func NewManager(client KubernetesClient, backupDir string, dryRun bool, logger Logger) *Manager {
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
	wp := workerpool.NewWorkerPool(maxConcurrency, 1000)
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
