package backup

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
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

// resourceTypes defines the Kubernetes resource types to be backed up
var resourceTypes = []string{"deployments", "services", "configmaps", "secrets", "hpas", "statefulsets", "cronjobs", "pvcs"}

// Manager handles the backup process for Kubernetes resources
type Manager struct {
	client    KubernetesClient
	backupDir string
	dryRun    bool
	logger    Logger
}

// NewManager creates a new Manager instance
func NewManager(client KubernetesClient, backupDir string, dryRun bool, logger Logger) *Manager {
	return &Manager{
		client:    client,
		backupDir: backupDir,
		dryRun:    dryRun,
		logger:    logger,
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

	g, ctx := errgroup.WithContext(ctx)

	// First, backup namespaces themselves
	g.Go(func() error {
		return bm.backupNamespaces(ctx)
	})

	// Enqueue backup tasks using errgroup
	for _, ns := range namespaces {
		for _, resourceType := range resourceTypes {
			resourceType := resourceType // capture range variable
			ns := ns
			g.Go(func() error {
				return bm.backupResource(ctx, resourceType, ns)
			})
		}
	}

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		bm.logger.Errorf("Error during backup: %v", err)
	}

	bm.logCompletionMessage(totalResources)
	return nil
}

// logCompletionMessage logs a message indicating the completion of the backup process
func (bm *Manager) logCompletionMessage(totalResources int) {
	if bm.dryRun {
		bm.logger.Infof("Dry run completed. %d resources would be backed up to: %s", totalResources, bm.backupDir)
	} else {
		bm.logger.Infof("Backup completed. %d resources saved to: %s", totalResources, bm.backupDir)
	}
}
