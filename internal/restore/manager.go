package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chaoscypher/k8s-backup-restore/internal/kubernetes"
	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
	"github.com/chaoscypher/k8s-backup-restore/internal/workerpool"
)

const maxConcurrency = 10

// Manager handles the restore operations.
type Manager struct {
	k8sClient *kubernetes.Client
	logger    logger.LoggerInterface
}

// NewManager creates a new restore Manager.
func NewManager(k8sClient *kubernetes.Client, logger logger.LoggerInterface) *Manager {
	return &Manager{
		k8sClient: k8sClient,
		logger:    logger,
	}
}

// PerformRestore performs the restore operation by reading resource files from the specified directory
// and applying them to the Kubernetes cluster. If dryRun is true, no changes will be made.
func (m *Manager) PerformRestore(restoreDir string, dryRun bool) error {
	m.logger.Info("Starting restore operation")

	files, err := getResourceFiles(restoreDir)
	if err != nil {
		return fmt.Errorf("error getting resource files: %v", err)
	}

	totalResources := len(files)

	if dryRun {
		m.logger.Info("Dry run mode: No resources will be created or modified")
	}

	// Initialize the worker pool
	wp := workerpool.NewWorkerPool(maxConcurrency, totalResources)
	m.enqueueTasks(files, wp, dryRun)

	// Run the worker pool and collect any errors
	errors := wp.Run(context.Background())
	if len(errors) > 0 {
		for _, err := range errors {
			m.logger.Errorf("Error during restore: %v", err)
		}
	}

	m.logCompletionMessage(totalResources, dryRun, restoreDir)
	return nil
}

// enqueueTasks adds restore tasks for each resource file to the worker pool.
func (m *Manager) enqueueTasks(files []string, wp *workerpool.WorkerPool, dryRun bool) {
	for _, file := range files {
		resourceFile := file // capture range variable
		task := func(ctx context.Context) error {
			return m.RestoreResource(resourceFile, dryRun)
		}
		if err := wp.AddTask(task); err != nil {
			m.logger.Errorf("Failed to add task for file %s: %v", resourceFile, err)
		}
	}
	wp.Close()
}

// logCompletionMessage logs a summary message upon completion of the restore operation.
func (m *Manager) logCompletionMessage(totalResources int, dryRun bool, restoreDir string) {
	if dryRun {
		m.logger.Infof("Dry run completed. %d resources would be restored from: %s", totalResources, restoreDir)
	} else {
		m.logger.Infof("Restore completed. %d resources restored from: %s", totalResources, restoreDir)
	}
}

// RestoreResource restores a single resource from the specified file. If dryRun is true, no changes will be made.
func (m *Manager) RestoreResource(filename string, dryRun bool) error {
	m.logger.Debugf("Restoring resource from file: %s", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filename, err)
	}

	var rawResource map[string]interface{}
	if err := json.Unmarshal(data, &rawResource); err != nil {
		return fmt.Errorf("error unmarshaling resource: %v", err)
	}

	resource, kind := adjustResourceStructure(rawResource)
	if err := validateResource(resource); err != nil {
		return fmt.Errorf("invalid resource structure: %v", err)
	}

	if dryRun {
		m.logger.Infof("Dry run: would restore %s/%s", kind, filename)
		return nil
	}

	name, namespace := getResourceIdentifiers(resource)
	m.logger.Infof("Restoring %s/%s in namespace %s", kind, name, namespace)
	return applyResource(m.k8sClient, resource, kind, namespace)
}
