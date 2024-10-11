package backup

import (
	"context"
	"fmt"
	"sync"
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

// maxConcurrency defines the maximum number of concurrent backup operations
const maxConcurrency = 10

// resourceTypes defines the Kubernetes resource types to be backed up
var resourceTypes = []string{"deployments", "services", "configmaps", "secrets", "hpas", "statefulsets", "cronjobs", "pvcs"}

// Manager handles the backup process for Kubernetes resources
type Manager struct {
	client    KubernetesClient
	backupDir string
	dryRun    bool
	logger    Logger
	tasks     chan func() error
	errors    []error
	errMu     sync.Mutex
	wg        sync.WaitGroup
}

// NewManager creates a new Manager instance
func NewManager(client KubernetesClient, backupDir string, dryRun bool, logger Logger) *Manager {
	return &Manager{
		client:    client,
		backupDir: backupDir,
		dryRun:    dryRun,
		logger:    logger,
		tasks:     make(chan func() error, 1000),
	}
}

// PerformBackup initiates the backup process for all namespaces
func (bm *Manager) PerformBackup(ctx context.Context) error {
	bm.logger.Info("Starting backup operation")

	// List all namespaces
	namespaces, err := bm.client.ListNamespaces(ctx)
	if err != nil {
		return fmt.Errorf("error listing namespaces: %v", err)
	}

	// Count total resources to be backed up
	totalResources := bm.countResources(ctx)

	if bm.dryRun {
		bm.logger.Info("Dry run mode: No files will be written")
	}

	// Start worker goroutines
	for i := 0; i < maxConcurrency; i++ {
		go bm.worker()
	}

	// Enqueue backup tasks
	bm.enqueueTasks(namespaces)

	// Wait for all tasks to complete
	bm.wg.Wait()
	close(bm.tasks)

	// Check for errors
	if len(bm.errors) > 0 {
		for _, err := range bm.errors {
			bm.logger.Errorf("Error during backup: %v", err)
		}
	}

	bm.logCompletionMessage(totalResources)
	return nil
}

// worker processes tasks from the tasks channel
func (bm *Manager) worker() {
	for task := range bm.tasks {
		if err := task(); err != nil {
			bm.errMu.Lock()
			bm.errors = append(bm.errors, err)
			bm.errMu.Unlock()
		}
		bm.wg.Done()
	}
}

// enqueueTasks adds backup tasks for each resource type in each namespace
func (bm *Manager) enqueueTasks(namespaces []string) {
	for _, ns := range namespaces {
		for _, resourceType := range resourceTypes {
			bm.wg.Add(1)
			task := func() error {
				return bm.backupResource(context.Background(), resourceType, ns)
			}
			bm.tasks <- task
		}
	}
}

// logCompletionMessage logs a message indicating the completion of the backup process
func (bm *Manager) logCompletionMessage(totalResources int) {
	if bm.dryRun {
		bm.logger.Infof("Dry run completed. %d resources would be backed up to: %s", totalResources, bm.backupDir)
	} else {
		bm.logger.Infof("Backup completed. %d resources saved to: %s", totalResources, bm.backupDir)
	}
}
