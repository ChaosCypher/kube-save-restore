package backup

import (
	"context"
	"fmt"
	"sync"

	"github.com/chaoscypher/kube-save-restore/internal/logger"
)

// Manager handles the backup process for Kubernetes resources.
type Manager struct {
	client         KubernetesClient
	backupDir      string
	dryRun         bool
	logger         logger.LoggerInterface
	resourceCounts map[string]int
	countMutex     sync.Mutex
}

// NewManager creates a new Manager instance.
func NewManager(client KubernetesClient, backupDir string, dryRun bool, logger logger.LoggerInterface) *Manager {
	return &Manager{
		client:         client,
		backupDir:      backupDir,
		dryRun:         dryRun,
		logger:         logger,
		resourceCounts: make(map[string]int),
	}
}

// PerformBackup initiates the backup process for all namespaces.
func (bm *Manager) PerformBackup(ctx context.Context) error {
	bm.logger.Info("Starting backup operation")

	// List all namespaces
	namespaces, err := bm.client.ListNamespaces(ctx)
	if err != nil {
		bm.logger.Errorf("error listing namespaces: %v", err)
		return fmt.Errorf("error listing namespaces: %v", err)
	}

	bm.logger.Debugf("Found %d namespaces", len(namespaces))

	if bm.dryRun {
		bm.logger.Info("Dry run mode: No files will be written")
	}

	resourceTypes := []string{"deployments", "services", "configmaps", "secrets", "hpas", "statefulsets", "cronjobs"}

	errChan := make(chan error, len(namespaces)*len(resourceTypes))
	var wg sync.WaitGroup

	for _, ns := range namespaces {
		for _, resourceType := range resourceTypes {
			wg.Add(1)
			go func(ns, rt string) {
				defer wg.Done()
				if err := bm.backupResource(ctx, rt, ns); err != nil {
					errChan <- fmt.Errorf("error backing up %s in namespace %s: %v", rt, ns, err)
				}
			}(ns, resourceType)
		}
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		for _, err := range errors {
			bm.logger.Error("Error during backup:", err)
		}
		bm.logger.Warnf("Completed backup with %d errors", len(errors))
	} else {
		bm.logger.Info("Backup completed successfully")
	}

	bm.logCompletionMessage()
	return nil
}

func (bm *Manager) incrementResourceCount(resourceType string) {
	bm.countMutex.Lock()
	defer bm.countMutex.Unlock()
	bm.resourceCounts[resourceType]++
}

func (bm *Manager) logCompletionMessage() {
	totalResources := 0
	for resourceType, count := range bm.resourceCounts {
		totalResources += count
		bm.logger.Debugf("Backed up %d %s", count, resourceType)
	}

	if bm.dryRun {
		bm.logger.Infof("Dry run completed. %d resources would be backed up to: %s", totalResources, bm.backupDir)
	} else {
		bm.logger.Infof("Backup completed. %d resources saved to: %s", totalResources, bm.backupDir)
	}
}
