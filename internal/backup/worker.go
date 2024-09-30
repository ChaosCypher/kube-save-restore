package backup

import (
	"context"

	"github.com/chaoscypher/k8s-backup-restore/internal/workerpool"
)

func (bm *Manager) enqueueTasks(namespaces []string, wp *workerpool.WorkerPool) {
	for _, ns := range namespaces {
		for _, resourceType := range []string{"deployments", "services", "configmaps", "secrets", "hpas", "statefulsets"} {
			task := func(ctx context.Context) error {
				return bm.backupResource(ctx, resourceType, ns)
			}
			if err := wp.AddTask(task); err != nil {
				bm.logger.Errorf("Failed to add task: %v", err)
			}
		}
	}
	wp.Close()
}

func (bm *Manager) logCompletionMessage(totalResources int) {
	if bm.dryRun {
		bm.logger.Infof("Dry run completed. %d resources would be backed up to: %s", totalResources, bm.backupDir)
	} else {
		bm.logger.Infof("Backup completed. %d resources saved to: %s", totalResources, bm.backupDir)
	}
}
