package backup

import (
	"context"
)

func (bm *Manager) enqueueTasks(namespace string) {
	resourceTypes := []string{"deployments", "services", "configmaps", "secrets"}
	for _, resourceType := range resourceTypes {
		bm.workerPool.AddTask(func(ctx context.Context) error {
			return bm.backupResource(ctx, resourceType, namespace)
		})
	}
}

func (bm *Manager) logCompletionMessage(totalResources int) {
	if bm.dryRun {
		bm.logger.Infof("Dry run completed. %d resources would be backed up to: %s", totalResources, bm.backupDir)
	} else {
		bm.logger.Infof("Backup completed. %d resources saved to: %s", totalResources, bm.backupDir)
	}
}
