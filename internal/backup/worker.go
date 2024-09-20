package backup

import (
	"context"
	"sync"
)

type backupTask struct {
	resourceType string
	namespace    string
}

func (bm *Manager) worker(ctx context.Context, wg *sync.WaitGroup, tasks <-chan backupTask) {
	defer wg.Done()
	for task := range tasks {
		if err := bm.backupResource(ctx, task.resourceType, task.namespace); err != nil {
			bm.logger.Errorf("Error backing up resource: %v", err)
		}
	}
}

func (bm *Manager) enqueueTasks(namespaces []string, tasks chan<- backupTask) {
	for _, ns := range namespaces {
		tasks <- backupTask{resourceType: "deployments", namespace: ns}
		tasks <- backupTask{resourceType: "services", namespace: ns}
		tasks <- backupTask{resourceType: "configmaps", namespace: ns}
		tasks <- backupTask{resourceType: "secrets", namespace: ns}
	}
}

func (bm *Manager) logCompletionMessage(totalResources int) {
	if bm.dryRun {
		bm.logger.Infof("Dry run completed. %d resources would be backed up to: %s", totalResources, bm.backupDir)
	} else {
		bm.logger.Infof("Backup completed. %d resources saved to: %s", totalResources, bm.backupDir)
	}
}
